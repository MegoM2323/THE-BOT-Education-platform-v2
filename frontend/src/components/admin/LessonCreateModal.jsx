import { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import ColorPicker from '../common/ColorPicker.jsx';
import StudentCheckboxList from './StudentCheckboxList.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import { useAuth } from '../../hooks/useAuth.js';
import { ROLES } from '../../utils/constants.js';
import * as lessonAPI from '../../api/lessons.js';
import * as userAPI from '../../api/users.js';
import * as creditAPI from '../../api/credits.js';
import { invalidateLessonData } from '../../utils/queryInvalidation.js';
import './LessonCreateModal.css';

/**
 * Modal для создания нового занятия
 * Унифицированный дизайн с LessonEditModal
 */
export const LessonCreateModal = ({
  isOpen,
  onClose,
  selectedDate,
  onLessonCreated
}) => {
  const { showNotification } = useNotification();
  const { user } = useAuth();
  const queryClient = useQueryClient();

  // Проверка роли: методист может назначать только себя
  const isMethodologist = user?.role === ROLES.METHODOLOGIST;

  // Состояния для формы
  const [teachers, setTeachers] = useState([]);
  const [formData, setFormData] = useState({
    teacher_id: '',
    lesson_date: '',
    start_time: '09:00',
    end_time: '11:00',
    max_students: 4,
    credits_cost: 1,
    subject: '',
    link: '',
    color: '#2563eb'
  });
  const [errors, setErrors] = useState({});

  // Состояния для студентов
  const [allStudents, setAllStudents] = useState([]);
  const [studentCredits, setStudentCredits] = useState({});
  const [selectedStudentIds, setSelectedStudentIds] = useState([]);

  // Состояния для повторяющихся занятий
  const [isRecurring, setIsRecurring] = useState(false);
  const [recurringWeeks, setRecurringWeeks] = useState(4);

  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);

  /**
   * Сбросить форму к начальным значениям
   */
  const resetForm = () => {
    setFormData({
      teacher_id: '',
      lesson_date: '',
      start_time: '09:00',
      end_time: '11:00',
      max_students: 4,
      credits_cost: 1,
      subject: '',
      link: '',
      color: '#2563eb'
    });
    setErrors({});
    setSelectedStudentIds([]);
    setIsRecurring(false);
    setRecurringWeeks(4);
  };

  // Загрузка данных при открытии modal
  useEffect(() => {
    if (isOpen) {
      loadTeachers();
      loadStudentsAndCredits();
      initializeFormData();
    } else {
      resetForm();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, selectedDate]);

  /**
   * Загрузить список преподавателей
   * Методист может назначать только себя
   */
  const loadTeachers = async () => {
    try {
      // Для методиста показываем только его самого
      if (isMethodologist && user?.id) {
        const selfTeacher = {
          id: user.id,
          full_name: user.full_name || user.name || user.email
        };
        setTeachers([selfTeacher]);
        setFormData(prev => ({ ...prev, teacher_id: user.id }));
        return;
      }

      // Для админа загружаем всех преподавателей
      const teachersResponse = await userAPI.getAssignableTeachersAll();
      const teachersList = Array.isArray(teachersResponse) ? teachersResponse : [];
      setTeachers(teachersList);

      // Установить первого преподавателя по умолчанию
      if (teachersList.length > 0) {
        setFormData(prev => ({ ...prev, teacher_id: teachersList[0].id }));
      }
    } catch (error) {
      console.error('Ошибка загрузки преподавателей:', error);
      showNotification('Ошибка загрузки списка преподавателей', 'error');
    }
  };

  /**
   * Загрузить список студентов и их кредитов
   */
  const loadStudentsAndCredits = async () => {
    try {
      setLoading(true);

      // Загружаем студентов и кредиты параллельно
      const [studentsResponse, creditsResponse] = await Promise.all([
        userAPI.getStudentsAll(),
        creditAPI.getAllCredits()
      ]);

      const studentsList = Array.isArray(studentsResponse) ? studentsResponse : [];
      setAllStudents(studentsList);

      // Создать карту кредитов
      const creditsMap = {};
      if (creditsResponse && creditsResponse.balances && Array.isArray(creditsResponse.balances)) {
        creditsResponse.balances.forEach(({ user_id, balance }) => {
          creditsMap[user_id] = balance || 0;
        });
      }
      setStudentCredits(creditsMap);
    } catch (error) {
      console.error('Ошибка загрузки студентов:', error);
      showNotification('Ошибка загрузки списка студентов', 'error');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Инициализировать данные формы с выбранной датой
   */
  const initializeFormData = () => {
    if (selectedDate) {
      const date = new Date(selectedDate);
      // Форматировать дату в YYYY-MM-DD (локальная дата)
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const dateString = `${year}-${month}-${day}`;

      setFormData(prev => ({
        ...prev,
        lesson_date: dateString,
        start_time: '09:00',
        end_time: '11:00'
      }));
    }
  };

  /**
   * Обработать изменение времени начала
   */
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
    setErrors(prev => ({ ...prev, start_time: '', end_time: '' }));
  };

  /**
   * Обработать изменение поля формы
   */
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    let parsedValue = value;

    if (name === 'max_students') {
      const parsed = parseInt(value, 10);
      parsedValue = isNaN(parsed) || parsed < 1 ? 1 : Math.min(parsed, 20);
    }

    if (name === 'credits_cost') {
      const parsed = parseInt(value, 10);
      parsedValue = isNaN(parsed) || parsed < 0 ? 0 : parsed;
    }

    setFormData(prev => ({
      ...prev,
      [name]: parsedValue
    }));
    setErrors(prev => ({ ...prev, [name]: '' }));
  };

  /**
   * Валидировать форму
   */
  const validateForm = () => {
    const newErrors = {};

    if (!formData.teacher_id) {
      newErrors.teacher_id = 'Пожалуйста, выберите преподавателя';
    }

    if (!formData.lesson_date) {
      newErrors.lesson_date = 'Пожалуйста, выберите дату';
    }

    if (!formData.start_time) {
      newErrors.start_time = 'Время начала обязательно';
    }

    if (!formData.end_time) {
      newErrors.end_time = 'Время окончания обязательно';
    }

    // Сравнить время только если оба заполнены
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

  /**
   * Проверка превышения лимита студентов
   */
  const isStudentLimitExceeded = selectedStudentIds.length > formData.max_students;

  /**
   * Обработчик выбора/снятия студента
   * @param {string} studentId - ID студента
   * @param {boolean} shouldEnroll - true = добавить, false = удалить
   */
  const handleStudentToggle = (studentId, shouldEnroll) => {
    // Получить кредиты студента
    const credits = studentCredits[studentId] ?? 0;

    if (shouldEnroll) {
      // Валидировать кредиты ПЕРЕД добавлением
      const requiredCredits = formData.credits_cost || 1;
      if (credits < requiredCredits) {
        showNotification(
          `Студент имеет недостаточно кредитов (${credits}, требуется ${requiredCredits})`,
          'error'
        );
        return;
      }

      // Проверка лимита студентов
      if (selectedStudentIds.length >= formData.max_students) {
        showNotification(
          `Превышен лимит студентов (максимум: ${formData.max_students})`,
          'warning'
        );
        return;
      }

      // Добавить студента
      setSelectedStudentIds(prev => [...prev, studentId]);
    } else {
      // Удалить студента
      setSelectedStudentIds(prev => prev.filter(id => id !== studentId));
    }
  };

  /**
   * Создать новое занятие
   */
  const handleCreateLesson = async () => {
    if (!validateForm()) {
      return;
    }

    try {
      setCreating(true);

      // Создать полный ISO datetime из даты и времени
      const startDateTime = new Date(`${formData.lesson_date}T${formData.start_time}:00`);
      const startDateTimeISO = startDateTime.toISOString();

      const endDateTime = new Date(`${formData.lesson_date}T${formData.end_time}:00`);
      const endDateTimeISO = endDateTime.toISOString();

      // Подготовить данные для отправки
      const requestData = {
        teacher_id: formData.teacher_id,
        start_time: startDateTimeISO,
        end_time: endDateTimeISO,
        max_students: formData.max_students,
        credits_cost: formData.credits_cost || 1,
        color: formData.color || '#2563eb'
      };

      // Добавить subject если указан
      if (formData.subject && formData.subject.trim()) {
        requestData.subject = formData.subject.trim().slice(0, 200);
      }

      // Добавить link если указан
      if (formData.link && formData.link.trim()) {
        requestData.link = formData.link.trim();
      }

      // Добавить параметры повторяющихся занятий
      if (isRecurring) {
        requestData.is_recurring = true;
        requestData.recurring_weeks = recurringWeeks;
      }

      // Добавить выбранных студентов
      if (selectedStudentIds.length > 0) {
        requestData.student_ids = selectedStudentIds;
      }

      const result = await lessonAPI.createLesson(requestData);

      invalidateLessonData(queryClient);

      const successMessage = isRecurring && result?.count
        ? `Создано ${result.count} повторяющихся занятий`
        : 'Занятие успешно создано';
      showNotification(successMessage, 'success');
      resetForm();
      onLessonCreated?.();
      onClose();
    } catch (error) {
      console.error('Ошибка создания занятия:', error);
      const errorMsg = error.response?.data?.message
        || error.response?.data?.error
        || error.message
        || 'Не удалось создать занятие';
      showNotification(errorMsg, 'error');
    } finally {
      setCreating(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Создание занятия"
      footer={
        <div className="lesson-create-footer">
          <Button
            variant="secondary"
            onClick={onClose}
            disabled={creating}
          >
            Отмена
          </Button>
          <Button
            variant="primary"
            onClick={handleCreateLesson}
            loading={creating}
            disabled={isStudentLimitExceeded}
          >
            Создать занятие
          </Button>
        </div>
      }
    >
      {loading ? (
        <div className="lesson-create-loading">
          <Spinner />
        </div>
      ) : (
        <div className="lesson-create-content">
          {/* Секция основной информации */}
          <section className="lesson-create-section">
            <h4 className="section-title">Основное</h4>
            <form data-testid="lesson-create-form">
              <div className="form-row">
                <div className="form-group">
                  <label className="form-label">Преподаватель *</label>
                  <select
                    className={`form-select ${errors.teacher_id ? 'error' : ''}`}
                    name="teacher_id"
                    value={formData.teacher_id}
                    onChange={handleInputChange}
                    disabled={creating || teachers.length === 0 || isMethodologist}
                    title={isMethodologist ? 'Вы можете назначать только себя' : ''}
                  >
                    <option value="">Выберите преподавателя</option>
                    {teachers.map(teacher => (
                      <option key={teacher.id} value={teacher.id}>
                        {teacher.full_name}
                      </option>
                    ))}
                  </select>
                  {errors.teacher_id && <span className="form-error">{errors.teacher_id}</span>}
                  {isMethodologist && (
                    <small className="form-hint">Вы можете назначать только себя</small>
                  )}
                </div>

                <div className="form-group">
                  <label className="form-label">Дата занятия *</label>
                  <input
                    type="date"
                    className={`form-input ${errors.lesson_date ? 'error' : ''}`}
                    value={formData.lesson_date}
                    onChange={(e) => {
                      setFormData({ ...formData, lesson_date: e.target.value });
                      setErrors(prev => ({ ...prev, lesson_date: '' }));
                    }}
                    disabled={creating}
                  />
                  {errors.lesson_date && <span className="form-error">{errors.lesson_date}</span>}
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
                    disabled={creating}
                  />
                  {errors.start_time && <span className="form-error">{errors.start_time}</span>}
                </div>

                <div className="form-group">
                  <label className="form-label">Время окончания *</label>
                  <input
                    type="time"
                    className={`form-input ${errors.end_time ? 'error' : ''}`}
                    value={formData.end_time}
                    onChange={(e) => {
                      setFormData({ ...formData, end_time: e.target.value });
                      setErrors(prev => ({ ...prev, end_time: '' }));
                    }}
                    disabled={creating}
                  />
                  {errors.end_time && <span className="form-error">{errors.end_time}</span>}
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label className="form-label recurring-label">
                    <input
                      type="checkbox"
                      checked={isRecurring}
                      onChange={(e) => setIsRecurring(e.target.checked)}
                      disabled={creating}
                    />
                    <span>Повторять еженедельно</span>
                  </label>
                </div>

                {isRecurring && (
                  <div className="form-group">
                    <label className="form-label">Количество недель</label>
                    <select
                      className="form-select"
                      value={recurringWeeks}
                      onChange={(e) => setRecurringWeeks(parseInt(e.target.value))}
                      disabled={creating}
                    >
                      <option value={4}>4 недели</option>
                      <option value={8}>8 недель</option>
                      <option value={12}>12 недель</option>
                    </select>
                  </div>
                )}
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label className="form-label">Максимум студентов *</label>
                  <input
                    type="number"
                    className={`form-input ${errors.max_students ? 'error' : ''}`}
                    name="max_students"
                    min="1"
                    max="20"
                    value={formData.max_students}
                    onChange={handleInputChange}
                    disabled={creating}
                  />
                  {errors.max_students && <span className="form-error">{errors.max_students}</span>}
                </div>

                <div className="form-group">
                  <label className="form-label">Стоимость в кредитах *</label>
                  <input
                    type="number"
                    className={`form-input ${errors.credits_cost ? 'error' : ''}`}
                    name="credits_cost" min="0"
                    value={formData.credits_cost}
                    onChange={handleInputChange}
                    disabled={creating}
                  />
                  {errors.credits_cost && <span className="form-error">{errors.credits_cost}</span>}
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">Тема занятия</label>
                <input
                  type="text"
                  className="form-input"
                  name="subject"
                  maxLength="200"
                  placeholder="Например: Математика 10 класс"
                  value={formData.subject}
                  onChange={handleInputChange}
                  disabled={creating}
                />
                <small className="form-hint">Максимум 200 символов</small>
              </div>

              <div className="form-group">
                <label className="form-label">Ссылка</label>
                <input
                  type="text"
                  className="form-input"
                  name="link"
                  placeholder="Например: https://meet.google.com/..."
                  value={formData.link}
                  onChange={handleInputChange}
                  disabled={creating}
                />
                <small className="form-hint">Ссылка на видеоконференцию или материалы</small>
              </div>

              <div className="form-group">
                <ColorPicker
                  value={formData.color}
                  onChange={(color) => setFormData(prev => ({ ...prev, color }))}
                  disabled={creating}
                />
              </div>
            </form>
          </section>

          {/* Управление студентами */}
          <section className="lesson-create-section">
            <StudentCheckboxList
              allStudents={allStudents.map(s => ({
                id: s.id,
                name: s.full_name || s.name,
                full_name: s.full_name || s.name,
                email: s.email,
                credits: studentCredits[s.id] || 0
              }))}
              enrolledStudentIds={selectedStudentIds}
              onToggle={handleStudentToggle}
            />
            {isStudentLimitExceeded && (
              <div className="form-error" style={{ marginTop: '8px' }}>
                Выбрано студентов ({selectedStudentIds.length}) больше, чем максимум ({formData.max_students})
              </div>
            )}
            <small className="form-hint">
              Выбрано: {selectedStudentIds.length} / {formData.max_students}
            </small>
          </section>
        </div>
      )}
    </Modal>
  );
};

export default LessonCreateModal;
