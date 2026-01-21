import { useState, useEffect, useRef } from 'react';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import ColorPicker from '../common/ColorPicker';
import { useNotification } from '../../hooks/useNotification.js';
import { useAuth } from '../../hooks/useAuth.js';
import { ROLES } from '../../utils/constants.js';
import * as usersAPI from '../../api/users.js';
import apiClient from '../../api/client.js';
import { allSettledWithLabels, withFallback } from '../../utils/promiseHelpers.js';
import './TemplateLessonEditorModal.css';

const DAYS_OF_WEEK = [
  { value: 1, label: 'Понедельник' },
  { value: 2, label: 'Вторник' },
  { value: 3, label: 'Среда' },
  { value: 4, label: 'Четверг' },
  { value: 5, label: 'Пятница' },
  { value: 6, label: 'Суббота' },
  { value: 0, label: 'Воскресенье' },
];

export const TemplateLessonEditorModal = ({
  isOpen,
  mode = 'create',
  templateId,
  lessonId = null,
  prefilledData = null,
  onClose,
  onSave,
  onError = null,
}) => {
  const notification = useNotification();
  const [teachers, setTeachers] = useState([]);
  const [students, setStudents] = useState([]);
  const [loadingData, setLoadingData] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [savingColor, setSavingColor] = useState(false);
  const colorSaveTimeoutRef = useRef(null);

  const [formData, setFormData] = useState({
    day_of_week: 1,
    start_time: '',
    teacher_id: '',
    max_students: 1,
    student_ids: [],
    color: '#2563eb',
    credits_cost: 1,
  });

  const [errors, setErrors] = useState({});

  // Cleanup color save timeout on unmount
  useEffect(() => {
    return () => {
      if (colorSaveTimeoutRef.current) {
        clearTimeout(colorSaveTimeoutRef.current);
      }
    };
  }, []);

  // Load teachers and students on mount

  useEffect(() => {
    if (isOpen) {
      loadInitialData();
    }
  }, [isOpen]);

  // Pre-fill form data in edit mode
  useEffect(() => {
    if (isOpen && mode === 'edit' && prefilledData) {
      // Извлекаем student_ids из разных форматов данных:
      // - assigned_students (массив ID) — из prefilledData компонента
      // - student_ids (массив ID) — из API ответа
      // - students (массив объектов с student_id) — из backend ответа
      let studentIds = [];
      if (prefilledData.assigned_students && Array.isArray(prefilledData.assigned_students)) {
        studentIds = prefilledData.assigned_students;
      } else if (prefilledData.student_ids && Array.isArray(prefilledData.student_ids)) {
        studentIds = prefilledData.student_ids;
      } else if (prefilledData.students && Array.isArray(prefilledData.students)) {
        studentIds = prefilledData.students.map(s => s.student_id || s.id);
      }

      setFormData({
        day_of_week: prefilledData.day_of_week || 1,
        start_time: prefilledData.start_time || '',
        teacher_id: prefilledData.teacher_id || '',
        max_students: prefilledData.max_students || 1,
        student_ids: studentIds,
        color: prefilledData.color || '#2563eb',
        credits_cost: prefilledData.credits_cost || 1,
      });
    } else if (isOpen && mode === 'create') {
      // Reset form for create mode
      setFormData({
        day_of_week: 1,
        start_time: '',
        teacher_id: '',
        max_students: 1,
        student_ids: [],
        color: '#2563eb',
        credits_cost: 1,
      });
      setSearchQuery('');
      setErrors({});
    }
  }, [isOpen, mode, prefilledData]);

  const loadInitialData = async () => {
    try {
      setLoadingData(true);

      // Use Promise.allSettled to handle partial failures
      const { results, failures } = await allSettledWithLabels(
        [
          usersAPI.getTeachersAll(),
          usersAPI.getStudentsAll(),
        ],
        ['Преподаватели', 'Студенты']
      );

      // Extract data with fallbacks
      const teachersData = withFallback(results[0], []);
      const studentsData = withFallback(results[1], []);

      setTeachers(teachersData);
      setStudents(studentsData);

      // Show specific errors for failed items
      if (failures.length > 0) {
        failures.forEach(({ label, error }) => {
          console.error(`Не удалось загрузить: ${label}`, error);
        });
        notification.error(
          `Не удалось загрузить: ${failures.map(f => f.label).join(', ')}`
        );
        if (onError) {
          onError(failures[0].error);
        }
      }
    } catch (error) {
      console.error('Unexpected error in loadInitialData:', error);
      notification.error('Не удалось загрузить преподавателей и студентов');
      if (onError) {
        onError(error);
      }
    } finally {
      setLoadingData(false);
    }
  };

  const handleStudentToggle = (studentId) => {
    setFormData((prev) => ({
      ...prev,
      student_ids: prev.student_ids.includes(studentId)
        ? prev.student_ids.filter((id) => id !== studentId)
        : [...prev.student_ids, studentId],
    }));
  };

  const handleRemoveStudent = (studentId) => {
    setFormData((prev) => ({
      ...prev,
      student_ids: prev.student_ids.filter((id) => id !== studentId),
    }));
  };

  const validateForm = () => {
    const newErrors = {};

    if (!formData.start_time) {
      newErrors.start_time = 'Время начала обязательно';
    } else {
      // Validate HH:MM format
      const timeRegex = /^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/;
      if (!timeRegex.test(formData.start_time)) {
        newErrors.start_time = 'Неверный формат времени (ЧЧ:ММ)';
      }
    }

    if (!formData.teacher_id) {
      newErrors.teacher_id = 'Преподаватель обязателен';
    }

    if (formData.max_students < 1) {
      newErrors.max_students = 'Минимум 1 студент';
    }

    if (formData.credits_cost === undefined || formData.credits_cost === null || formData.credits_cost < 0) {
      newErrors.credits_cost = 'Стоимость в кредитах должна быть не менее 0';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Auto-save color with debounce (500ms)
  const handleColorChange = (color) => {
    setFormData({ ...formData, color });

    // Only auto-save if in edit mode and lesson exists
    if (mode === 'edit' && lessonId) {
      // Clear previous timeout
      if (colorSaveTimeoutRef.current) {
        clearTimeout(colorSaveTimeoutRef.current);
      }

      // Set new timeout for debounced save
      setSavingColor(true);
      colorSaveTimeoutRef.current = setTimeout(async () => {
        try {
          await apiClient.patch(
            `/templates/${templateId}/lessons/${lessonId}`,
            { color }
          );
          notification.success('Цвет занятия сохранён');
        } catch (error) {
          notification.error('Не удалось сохранить цвет');
          console.error('Color save error:', error);
        } finally {
          setSavingColor(false);
        }
      }, 500);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setSubmitting(true);

    try {
      const lessonData = {
        day_of_week: parseInt(formData.day_of_week),
        start_time: formData.start_time,
        teacher_id: formData.teacher_id,
        max_students: parseInt(formData.max_students),
        student_ids: formData.student_ids,
        color: formData.color,
        credits_cost: parseInt(formData.credits_cost),
      };

      let savedLesson;

      if (mode === 'edit' && lessonId) {
        // Update existing template lesson
        savedLesson = await apiClient.patch(
          `/templates/${templateId}/lessons/${lessonId}`,
          lessonData
        );
        notification.success('Шаблон занятия обновлен');
      } else {
        // Create new template lesson
        savedLesson = await apiClient.post(`/templates/${templateId}/lessons`, lessonData);
        notification.success('Шаблон занятия создан');
      }

      if (onSave) {
        onSave(savedLesson);
      }

      onClose();
    } catch (error) {
      const errorMessage = error.message || 'Не удалось сохранить шаблон занятия';
      notification.error(errorMessage);
      if (onError) {
        onError(error);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const filteredStudents = students.filter(
    (student) =>
      student.full_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      student.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const selectedStudentObjects = students.filter((s) => formData.student_ids.includes(s.id));

  const modalTitle = mode === 'edit' ? 'Редактировать шаблон занятия' : 'Добавить шаблон занятия';

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={modalTitle} loading={submitting}>
      <div className="template-lesson-editor-modal">
        {loadingData ? (
          <div className="template-lesson-editor-loading">
            <Spinner />
          </div>
        ) : (
          <form onSubmit={handleSubmit} data-testid="template-lesson-editor-form">
            <div className="form-row">
              <div className="form-group">
                <label className="form-label">
                  День недели <span className="required">*</span>
                </label>
                <select
                  className={`form-select ${errors.day_of_week ? 'error' : ''}`}
                  value={formData.day_of_week}
                  onChange={(e) =>
                    setFormData({ ...formData, day_of_week: parseInt(e.target.value) })
                  }
                  required
                >
                  {DAYS_OF_WEEK.map((day) => (
                    <option key={day.value} value={day.value}>
                      {day.label}
                    </option>
                  ))}
                </select>
                {errors.day_of_week && <span className="form-error">{errors.day_of_week}</span>}
              </div>

              <div className="form-group">
                <label className="form-label">
                  Время начала <span className="required">*</span>
                </label>
                <input
                  type="time"
                  className={`form-input ${errors.start_time ? 'error' : ''}`}
                  value={formData.start_time}
                  onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
                  required
                />
                {errors.start_time && <span className="form-error">{errors.start_time}</span>}
              </div>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">
                  Преподаватель <span className="required">*</span>
                </label>
                <select
                  className={`form-select ${errors.teacher_id ? 'error' : ''}`}
                  value={formData.teacher_id}
                  onChange={(e) => setFormData({ ...formData, teacher_id: e.target.value })}
                  required
                >
                  <option value="">Выберите преподавателя</option>
                  {teachers.map((teacher) => (
                    <option key={teacher.id} value={teacher.id}>
                      {teacher.full_name}
                    </option>
                  ))}
                </select>
                {errors.teacher_id && <span className="form-error">{errors.teacher_id}</span>}
              </div>

              <div className="form-group">
                <label className="form-label">
                  Максимум студентов <span className="required">*</span>
                </label>
                <input
                  type="number"
                  className={`form-input ${errors.max_students ? 'error' : ''}`}
                  value={formData.max_students}
                  onChange={(e) =>
                    setFormData({ ...formData, max_students: parseInt(e.target.value) })
                  }
                  min="1"
                  max="20"
                  required
                />
                {errors.max_students && <span className="form-error">{errors.max_students}</span>}
              </div>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">
                  Стоимость в кредитах <span className="required">*</span>
                </label>
                <input
                  type="number"
                  name="credits_cost" min="0"
                  className={`form-input ${errors.credits_cost ? 'error' : ''}`}
                  value={formData.credits_cost}
                  onChange={(e) =>
                    setFormData({ ...formData, credits_cost: parseInt(e.target.value) })
                  }
                  required
                />
                {errors.credits_cost && <span className="form-error">{errors.credits_cost}</span>}
              </div>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">
                  Цвет занятия
                  {savingColor && <span className="saving-indicator"> (сохраняется...)</span>}
                </label>
                <ColorPicker
                  value={formData.color}
                  onChange={handleColorChange}
                  disabled={submitting || (mode === 'create')}
                />
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">
                Назначенные студенты ({formData.student_ids.length})
              </label>

              {/* Selected students display */}
              {selectedStudentObjects.length > 0 && (
                <div className="selected-students">
                  {selectedStudentObjects.map((student) => (
                    <div key={student.id} className="selected-student-chip">
                      <span>{student.full_name}</span>
                      <button
                        type="button"
                        className="remove-student-btn"
                        onClick={() => handleRemoveStudent(student.id)}
                        aria-label={`Remove ${student.full_name}`}
                      >
                        &times;
                      </button>
                    </div>
                  ))}
                </div>
              )}

              {/* Student search */}
              <input
                type="text"
                className="student-search-input"
                placeholder="Поиск студентов..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />

              {/* Student list with checkboxes */}
              <div className="student-list">
                {filteredStudents.length === 0 ? (
                  <div className="student-list-empty">
                    {searchQuery ? 'Студенты не найдены' : 'Нет доступных студентов'}
                  </div>
                ) : (
                  filteredStudents.map((student) => {
                    const isSelected = formData.student_ids.includes(student.id);
                    return (
                      <div
                        key={student.id}
                        className={`student-item ${isSelected ? 'selected' : ''}`}
                        onClick={() => handleStudentToggle(student.id)}
                      >
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleStudentToggle(student.id)}
                          onClick={(e) => e.stopPropagation()}
                        />
                        <div className="student-info">
                          <div className="student-name">{student.full_name}</div>
                          <div className="student-email">{student.email}</div>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            <div className="form-actions">
              <Button type="button" variant="secondary" onClick={onClose} disabled={submitting}>
                Отмена
              </Button>
              <Button type="submit" loading={submitting}>
                {mode === 'edit' ? 'Сохранить изменения' : 'Добавить занятие'}
              </Button>
            </div>
          </form>
        )}
      </div>
    </Modal>
  );
};

export default TemplateLessonEditorModal;
