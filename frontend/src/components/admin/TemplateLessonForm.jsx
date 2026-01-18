import { useState, useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import Button from "../common/Button.jsx";
import Input from '../common/Input.jsx';
import Spinner from '../common/Spinner.jsx';
import ColorPicker from '../common/ColorPicker.jsx';
import ConfirmModal from '../common/ConfirmModal.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import * as usersAPI from '../../api/users.js';
import apiClient from '../../api/client.js';
import TemplateStudentAssignmentModal from './TemplateStudentAssignmentModal.jsx';
import './TemplateLessonForm.css';

const DAYS_OF_WEEK = [
  { value: 0, label: 'Понедельник' },
  { value: 1, label: 'Вторник' },
  { value: 2, label: 'Среда' },
  { value: 3, label: 'Четверг' },
  { value: 4, label: 'Пятница' },
  { value: 5, label: 'Суббота' },
  { value: 6, label: 'Воскресенье' },
];

  const normalizeStudentIds = (ids) => {
  if (!Array.isArray(ids)) return new Set();
  return new Set(
    ids
      .map(id => {
        if (typeof id === 'object' && (id.id || id.student_id)) {
          return String(id.id || id.student_id).trim();
        }
        return String(id).trim();
      })
      .filter(Boolean)
  );
};

const studentIdsChanged = (oldIds, newIds) => {
  const oldSet = normalizeStudentIds(oldIds);
  const newSet = normalizeStudentIds(newIds);

  if (oldSet.size !== newSet.size) return true;

  return ![...oldSet].every(id => newSet.has(id));
};

export const TemplateLessonForm = ({ templateId, onSave, editingLesson = null, onCancel, forceDay = null, onClose }) => {
  const notification = useNotification();
  const queryClient = useQueryClient();
  const [teachers, setTeachers] = useState([]);
  const [loadingTeachers, setLoadingTeachers] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [showStudentModal, setShowStudentModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const [formData, setFormData] = useState({
    teacher_id: '',
    day_of_week: forceDay !== null ? forceDay : 0,
    start_time: '10:00',
    end_time: '12:00',
    max_students: 4,
    credits_cost: 1,
    subject: '',
    color: '#2563eb',
    student_ids: [],
  });

  const [errors, setErrors] = useState({});
  const handleCancelRef = useRef(null);

  useEffect(() => {
    fetchTeachers();
  }, []);

  // Сохраняем ссылку на handleCancel для вызова извне
  // Обновляем ref при каждом изменении handleCancel
  useEffect(() => {
    handleCancelRef.current = handleCancel;
  });

  // Передаем функцию handleClose через onClose callback
  useEffect(() => {
    if (onClose) {
      // onClose должен быть функцией, которая принимает handleClose
      if (typeof onClose === 'function') {
        onClose(async () => {
          if (handleCancelRef.current) {
            await handleCancelRef.current();
          }
        });
      }
    }
    // Cleanup при размонтировании
    return () => {
      if (window.__templateLessonFormHandleClose) {
        delete window.__templateLessonFormHandleClose;
      }
    };
  }, [onClose]);

  useEffect(() => {
    if (editingLesson) {
      console.log('[TemplateLessonForm] Loading editingLesson:', {
        lesson_id: editingLesson.id,
        credits_cost: editingLesson.credits_cost,
        credits_cost_type: typeof editingLesson.credits_cost,
        raw_object: editingLesson
      });

      // Извлекаем student_ids из разных форматов данных:
      // - student_ids (массив ID) — из API ответа
      // - students (массив объектов с student_id) — из backend ответа
      let studentIds = [];
      if (editingLesson.student_ids && Array.isArray(editingLesson.student_ids)) {
        studentIds = editingLesson.student_ids;
      } else if (editingLesson.students && Array.isArray(editingLesson.students)) {
        studentIds = editingLesson.students.map(s => s.student_id || s.id);
      }

      // Форматируем время из формата HH:MM:SS или HH:MM в HH:MM
      const formatTime = (time) => {
        if (!time) return '10:00';
        // Если формат уже HH:MM, возвращаем как есть
        if (time.length === 5) return time;
        // Если HH:MM:SS, отрезаем секунды
        if (time.length >= 5) return time.substring(0, 5);
        return time;
      };

      const newFormData = {
        teacher_id: editingLesson.teacher_id || '',
        day_of_week: editingLesson.day_of_week !== undefined ? parseInt(editingLesson.day_of_week) : 0,
        start_time: formatTime(editingLesson.start_time),
        end_time: formatTime(editingLesson.end_time),
        max_students: editingLesson.max_students || 4,
        credits_cost: editingLesson.credits_cost ?? 1,
        subject: editingLesson.subject || '',
        color: editingLesson.color || '#2563eb',
        student_ids: studentIds,
      };

      console.log('[TemplateLessonForm] Populated formData:', {
        credits_cost: newFormData.credits_cost,
        credits_cost_type: typeof newFormData.credits_cost,
        full_data: newFormData
      });

      setFormData(newFormData);
    }
  }, [editingLesson]);

  const fetchTeachers = async () => {
    try {
      setLoadingTeachers(true);
      const data = await usersAPI.getTeachersAll();
      setTeachers(data);
    } catch (error) {
      notification.error('Не удалось загрузить список преподавателей');
    } finally {
      setLoadingTeachers(false);
    }
  };

  const handleStartTimeChange = (value) => {
    setFormData((prev) => {
      const newData = { ...prev, start_time: value };

      // Auto-calculate end time as start + 2 hours
      if (value) {
        const [hours, minutes] = value.split(':').map(Number);
        const endHours = (hours + 2) % 24;
        const endTime = `${String(endHours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
        newData.end_time = endTime;
      }

      return newData;
    });
  };

  const validateForm = () => {
    const newErrors = {};

    if (!formData.teacher_id) {
      newErrors.teacher_id = 'Пожалуйста, выберите преподавателя';
    }

    if (!formData.start_time) {
      newErrors.start_time = 'Время начала обязательно';
    }

    if (!formData.end_time) {
      newErrors.end_time = 'Время окончания обязательно';
    }

    if (formData.start_time && formData.end_time) {
      const [startH, startM] = formData.start_time.split(':').map(Number);
      const [endH, endM] = formData.end_time.split(':').map(Number);
      const startMinutes = startH * 60 + startM;
      const endMinutes = endH * 60 + endM;

      if (endMinutes <= startMinutes) {
        newErrors.end_time = 'Время окончания должно быть после времени начала';
      }
    }

    if (formData.max_students < 1 || formData.max_students > 20) {
      newErrors.max_students = 'Максимум студентов должен быть от 1 до 20';
    }

    if (formData.credits_cost === undefined || formData.credits_cost === null || formData.credits_cost < 0) {
      newErrors.credits_cost = 'Стоимость в кредитах должна быть не менее 0';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const saveLesson = async () => {
    if (!validateForm()) {
      return false;
    }

    setSubmitting(true);

    try {
      const lessonData = {
        teacher_id: formData.teacher_id,
        day_of_week: parseInt(formData.day_of_week),
        start_time: formData.start_time,
        end_time: formData.end_time,
        max_students: parseInt(formData.max_students),
        credits_cost: parseInt(formData.credits_cost),
        subject: formData.subject.trim().slice(0, 200) || '',
        color: formData.color || '#2563eb',
        student_ids: formData.student_ids,
      };

      const operation = editingLesson ? 'UPDATE' : 'CREATE';
      console.log(`[TemplateLessonForm] ${operation} lesson - sending API request:`, {
        operation,
        lesson_id: editingLesson?.id || 'new',
        template_id: templateId,
        credits_cost: lessonData.credits_cost,
        credits_cost_type: typeof lessonData.credits_cost,
        payload: lessonData
      });

      let response;
      if (editingLesson) {
        // Update existing template lesson
        response = await apiClient.put(`/templates/${templateId}/lessons/${editingLesson.id}`, lessonData);
        console.log('[TemplateLessonForm] UPDATE response:', {
          status: response.status,
          credits_cost: response.data?.credits_cost,
          full_response: response.data
        });
        notification.success('Занятие шаблона обновлено');
      } else {
        // Create new template lesson
        response = await apiClient.post(`/templates/${templateId}/lessons`, lessonData);
        console.log('[TemplateLessonForm] CREATE response:', {
          status: response.status,
          credits_cost: response.data?.credits_cost,
          full_response: response.data
        });
        notification.success('Занятие шаблона добавлено');
      }

      // Инвалидация кэша React Query для немедленного обновления UI
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      queryClient.invalidateQueries({ queryKey: ['templates', templateId] });

      // Reset form
      setFormData({
        teacher_id: '',
        day_of_week: 0,
        start_time: '10:00',
        end_time: '12:00',
        max_students: 4,
        credits_cost: 1,
        subject: '',
        color: '#2563eb',
        student_ids: [],
      });
      setErrors({});

      if (onSave) {
        onSave();
      }
      return true;
    } catch (error) {
      console.error('[TemplateLessonForm] Save error:', {
        error_message: error.message,
        error_response: error.response?.data,
        full_error: error
      });
      notification.error(error.message || 'Не удалось сохранить занятие шаблона');
      return false;
    } finally {
      setSubmitting(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    await saveLesson();
  };

  const handleCancel = async () => {
    // Сохраняем перед закрытием если были изменения
    if (editingLesson) {
      // Извлекаем student_ids из разных форматов данных
      let originalStudentIds = [];
      if (editingLesson.student_ids && Array.isArray(editingLesson.student_ids)) {
        originalStudentIds = editingLesson.student_ids;
      } else if (editingLesson.students && Array.isArray(editingLesson.students)) {
        originalStudentIds = editingLesson.students.map(s => s.student_id || s.id);
      }

      // Форматируем время для сравнения
      const formatTime = (time) => {
        if (!time) return '10:00';
        if (time.length === 5) return time;
        if (time.length >= 5) return time.substring(0, 5);
        return time;
      };

      const originalData = {
        teacher_id: String(editingLesson.teacher_id || ''),
        day_of_week: editingLesson.day_of_week !== undefined ? editingLesson.day_of_week : 0,
        start_time: formatTime(editingLesson.start_time),
        end_time: formatTime(editingLesson.end_time),
        max_students: editingLesson.max_students || 4,
        credits_cost: editingLesson.credits_cost ?? 1,
        subject: (editingLesson.subject || '').trim(),
        color: editingLesson.color || '#2563eb',
        student_ids: originalStudentIds,
      };

      // Нормализуем текущие данные для сравнения
      const currentData = {
        teacher_id: String(formData.teacher_id || ''),
        day_of_week: parseInt(formData.day_of_week) !== undefined ? parseInt(formData.day_of_week) : 0,
        start_time: formatTime(formData.start_time),
        end_time: formatTime(formData.end_time),
        max_students: formData.max_students,
        credits_cost: parseInt(formData.credits_cost) || 1,
        subject: (formData.subject || '').trim(),
        color: formData.color || '#2563eb',
        student_ids: formData.student_ids || [],
      };

      // Проверяем изменения
      const hasChanges =
        currentData.teacher_id !== originalData.teacher_id ||
        currentData.day_of_week !== originalData.day_of_week ||
        currentData.start_time !== originalData.start_time ||
        currentData.end_time !== originalData.end_time ||
        currentData.max_students !== originalData.max_students ||
        currentData.credits_cost !== originalData.credits_cost ||
        currentData.subject !== originalData.subject ||
        currentData.color !== originalData.color ||
        studentIdsChanged(originalData.student_ids, currentData.student_ids);

      if (hasChanges) {
        console.log('[TemplateLessonForm] Changes detected - comparing data:', {
          original_credits_cost: originalData.credits_cost,
          current_credits_cost: currentData.credits_cost,
          credits_cost_changed: currentData.credits_cost !== originalData.credits_cost,
          changed_fields: {
            teacher_id: currentData.teacher_id !== originalData.teacher_id,
            day_of_week: currentData.day_of_week !== originalData.day_of_week,
            start_time: currentData.start_time !== originalData.start_time,
            end_time: currentData.end_time !== originalData.end_time,
            max_students: currentData.max_students !== originalData.max_students,
            credits_cost: currentData.credits_cost !== originalData.credits_cost,
            subject: currentData.subject !== originalData.subject,
            color: currentData.color !== originalData.color,
            student_ids: studentIdsChanged(originalData.student_ids, currentData.student_ids)
          }
        });
        // Есть изменения - сохраняем
        setSubmitting(true);
        const saved = await saveLesson();
        setSubmitting(false);
        if (!saved) return; // Если сохранение не удалось, не закрываем
      }
    }

    if (onCancel) {
      onCancel();
    }
  };

  const handleStudentsAssigned = (studentIds) => {
    setFormData((prev) => ({
      ...prev,
      student_ids: studentIds,
    }));
  };

  const handleDeleteLesson = () => {
    setShowDeleteConfirm(true);
  };

  const confirmDeleteLesson = async () => {
    setDeleting(true);

    try {
      await apiClient.delete(`/templates/${templateId}/lessons/${editingLesson.id}`);
      notification.success('Занятие шаблона удалено');

      // Инвалидация кэша React Query для немедленного обновления UI
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      queryClient.invalidateQueries({ queryKey: ['templates', templateId] });

      // Reset form
      setFormData({
        teacher_id: '',
        day_of_week: 0,
        start_time: '10:00',
        end_time: '12:00',
        max_students: 4,
        credits_cost: 1,
        subject: '',
        color: '#2563eb',
        student_ids: [],
      });
      setErrors({});

      if (onSave) {
        onSave();
      }
    } catch (error) {
      notification.error(error.message || 'Не удалось удалить занятие шаблона');
    } finally {
      setDeleting(false);
    }
  };

  if (loadingTeachers) {
    return (
      <div className="template-lesson-form-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="template-lesson-form">
      <h3 className="template-lesson-form-title">
        {editingLesson ? 'Редактировать занятие шаблона' : 'Добавить занятие в шаблон'}
      </h3>

      <form onSubmit={handleSubmit} data-testid="template-lesson-form" data-template-lesson-form>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">День недели *</label>
            <select
              className={`form-select ${errors.day_of_week ? 'error' : ''}`}
              value={parseInt(formData.day_of_week)}
              onChange={(e) => setFormData({ ...formData, day_of_week: parseInt(e.target.value) })}
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
            <label className="form-label">Преподаватель *</label>
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
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Время начала *</label>
            <input
              type="time"
              className={`form-input ${errors.start_time ? 'error' : ''}`}
              value={formData.start_time}
              onChange={(e) => handleStartTimeChange(e.target.value)}
              required
            />
            {errors.start_time && <span className="form-error">{errors.start_time}</span>}
          </div>

          <div className="form-group">
            <label className="form-label">Время окончания *</label>
            <input
              type="time"
              className={`form-input ${errors.end_time ? 'error' : ''}`}
              value={formData.end_time}
              onChange={(e) => setFormData({ ...formData, end_time: e.target.value })}
              required
            />
            {errors.end_time && <span className="form-error">{errors.end_time}</span>}
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Максимум студентов *</label>
            <input
              type="number"
              className={`form-input ${errors.max_students ? 'error' : ''}`}
              value={formData.max_students}
              onChange={(e) => setFormData({ ...formData, max_students: parseInt(e.target.value) })}
              min="1"
              max="20"
              required
            />
            {errors.max_students && <span className="form-error">{errors.max_students}</span>}
          </div>

          <div className="form-group">
            <label className="form-label">Стоимость в кредитах *</label>
            <input
              type="number"
              className={`form-input ${errors.credits_cost ? 'error' : ''}`}
              value={formData.credits_cost}
              onChange={(e) => {
                const newValue = parseInt(e.target.value);
                console.log('[TemplateLessonForm] credits_cost input changed:', {
                  raw_input: e.target.value,
                  parsed_value: newValue,
                  type: typeof newValue
                });
                setFormData({ ...formData, credits_cost: newValue });
              }}
              min="0"
              required
            />
            {errors.credits_cost && <span className="form-error">{errors.credits_cost}</span>}
          </div>
        </div>

        <div className="form-group">
          <label className="form-label">Тема занятия</label>
          <input
            type="text"
            className="form-input"
            value={formData.subject}
            onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
            placeholder="Например: Математика 10 класс"
            maxLength="200"
          />
          <small className="form-hint">Максимум 200 символов</small>
        </div>

        <div className="form-group">
          <ColorPicker
            value={formData.color}
            onChange={(color) => setFormData({ ...formData, color })}
            disabled={submitting}
          />
        </div>

        <div className="form-group">
          <label className="form-label">Назначенные студенты ({formData.student_ids.length})</label>
          <Button
            type="button"
            variant="outline"
            size="small"
            onClick={() => setShowStudentModal(true)}
          >
            Управление студентами
          </Button>
        </div>

        <div className="template-lesson-form-actions">
          {editingLesson && (
            <Button
              type="button"
              variant="danger"
              onClick={handleDeleteLesson}
              disabled={deleting || submitting}
            >
              Удалить
            </Button>
          )}
          {!editingLesson && (
            <Button type="submit" loading={submitting} disabled={deleting}>
              Добавить занятие
            </Button>
          )}
        </div>
      </form>

      <TemplateStudentAssignmentModal
        isOpen={showStudentModal}
        onClose={() => setShowStudentModal(false)}
        selectedStudents={formData.student_ids}
        onStudentsAssigned={handleStudentsAssigned}
        maxStudents={formData.max_students}
      />

      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={confirmDeleteLesson}
        title="Удаление занятия шаблона"
        message="Вы уверены, что хотите удалить это занятие?"
        confirmText="Удалить"
        confirmVariant="danger"
        loading={deleting}
      />
    </div>
  );
};

export default TemplateLessonForm;
