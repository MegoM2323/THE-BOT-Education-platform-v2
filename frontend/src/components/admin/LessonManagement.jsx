import { useState, useEffect, useCallback } from 'react';
import Button from "../common/Button.jsx";
import Input from '../common/Input.jsx';
import Modal from "../common/Modal.jsx";
import ConfirmModal from '../common/ConfirmModal.jsx';
import Spinner from '../common/Spinner.jsx';
import { SkeletonTable } from '../common/SkeletonLoader.jsx';
import SlowConnectionNotice from '../common/SlowConnectionNotice.jsx';
import PartialDataNotice from '../common/PartialDataNotice.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import { useSlowConnection } from '../../hooks/useSlowConnection.js';
import * as lessonsAPI from '../../api/lessons.js';
import * as usersAPI from '../../api/users.js';
import { formatDate, formatTime } from '../../utils/dateFormat.js';
import { allSettledWithLabels, withFallback } from '../../utils/promiseHelpers.js';
import './LessonManagement.css';

export const LessonManagement = () => {
  const [lessons, setLessons] = useState([]);
  const [teachers, setTeachers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showLessonModal, setShowLessonModal] = useState(false);
  const [editingLesson, setEditingLesson] = useState(null);
  const [deleteLessonId, setDeleteLessonId] = useState(null);
  const [formData, setFormData] = useState({
    teacher_id: '',
    start_time: '',
    end_time: '',
    max_students: 6,
    credits_cost: 1,
  });
  const [submitting, setSubmitting] = useState(false);
  const [dataFailures, setDataFailures] = useState([]);
  const [retrying, setRetrying] = useState(false);
  const [dismissedSlowNotice, setDismissedSlowNotice] = useState(false);
  const notification = useNotification();

  // Detect slow connection
  const isSlow = useSlowConnection(loading, 3000);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setDataFailures([]);

      // Use Promise.allSettled to handle partial failures
      const { results, failures } = await allSettledWithLabels(
        [
          lessonsAPI.getLessons(),
          usersAPI.getTeachersAll(),
        ],
        ['Занятия', 'Преподаватели']
      );

      // Extract data with fallbacks
      const lessonsData = withFallback(results[0], []);
      const teachersData = withFallback(results[1], []);

      setLessons(lessonsData);
      setTeachers(teachersData);

      // Store failures for display
      if (failures.length > 0) {
        failures.forEach(({ label, error }) => {
          console.error(`Ошибка загрузки: ${label}`, error);
        });
        setDataFailures(failures);
      }
    } catch (error) {
      console.error('Unexpected error in fetchData:', error);
      notification.error('Ошибка загрузки данных');
    } finally {
      setLoading(false);
      setRetrying(false);
    }
  }, [notification]);

  const handleRetry = useCallback(async () => {
    setRetrying(true);
    await fetchData();
  }, [fetchData]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleOpenCreate = () => {
    setEditingLesson(null);
    const now = new Date();
    const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
    const startTime = new Date(tomorrow.setHours(10, 0, 0, 0)).toISOString().slice(0, 16);
    const endTime = new Date(tomorrow.setHours(11, 0, 0, 0)).toISOString().slice(0, 16);

    setFormData({
      teacher_id: teachers[0]?.id || '',
      start_time: startTime,
      end_time: endTime,
      max_students: 6,
      credits_cost: 1,
    });
    setShowLessonModal(true);
  };

  const handleOpenEdit = (lesson) => {
    setEditingLesson(lesson);
    setFormData({
      teacher_id: lesson.teacher_id,
      start_time: new Date(lesson.start_time).toISOString().slice(0, 16),
      end_time: new Date(lesson.end_time).toISOString().slice(0, 16),
      max_students: lesson.max_students,
      credits_cost: lesson.credits_cost || 1,
    });
    setShowLessonModal(true);
  };

  const handleCloseModal = () => {
    setShowLessonModal(false);
    setEditingLesson(null);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      // Валидация
      if (!formData.teacher_id || formData.teacher_id.trim() === '') {
        throw new Error('Пожалуйста, выберите преподавателя');
      }

      if (!formData.start_time || !formData.end_time) {
        throw new Error('Укажите время начала и окончания');
      }

      const startTime = new Date(formData.start_time);
      const endTime = new Date(formData.end_time);

      if (endTime <= startTime) {
        throw new Error('Время окончания должно быть после времени начала');
      }

      const maxStudents = parseInt(formData.max_students);
      if (isNaN(maxStudents) || maxStudents <= 0) {
        throw new Error('Количество студентов должно быть положительным числом');
      }

      const creditsCost = parseInt(formData.credits_cost);
      if (isNaN(creditsCost) || creditsCost < 0) {
        throw new Error('Стоимость в кредитах должна быть неотрицательным числом');
      }

      const lessonData = {
        teacher_id: formData.teacher_id.trim(), // UUID string
        max_students: maxStudents,
        credits_cost: creditsCost,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
      };

      if (editingLesson) {
        await lessonsAPI.updateLesson(editingLesson.id, lessonData);
        notification.success('Занятие обновлено');
      } else {
        await lessonsAPI.createLesson(lessonData);
        notification.success('Занятие создано');
      }
      handleCloseModal();
      fetchData();
    } catch (error) {
      notification.error(error.message || 'Ошибка сохранения занятия');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteLessonId) return;

    try {
      await lessonsAPI.deleteLesson(deleteLessonId);
      notification.success('Занятие удалено, кредиты возвращены ученикам');
      setDeleteLessonId(null);
      fetchData();
    } catch (error) {
      notification.error(error.message || 'Ошибка удаления занятия');
    }
  };

  if (loading) {
    return (
      <div className="lesson-management" data-testid="lesson-management">
        <div className="lesson-management-header">
          <h2 className="lesson-management-title">Управление занятиями</h2>
          <Button onClick={handleOpenCreate} disabled data-testid="create-lesson-button">Создать занятие</Button>
        </div>

        <div className="lesson-management-table-wrapper" aria-busy="true" aria-live="polite">
          <SkeletonTable rows={6} columns={5} />
        </div>

        {/* Show slow connection notice if loading takes too long */}
        {isSlow && !dismissedSlowNotice && (
          <SlowConnectionNotice onDismiss={() => setDismissedSlowNotice(true)} />
        )}
      </div>
    );
  }

  return (
    <div className="lesson-management" data-testid="lesson-management">
      <div className="lesson-management-header">
        <h2 className="lesson-management-title">Управление занятиями</h2>
        <Button onClick={handleOpenCreate} data-testid="create-lesson-button">Создать занятие</Button>
      </div>

      {/* Show partial data notice if some data failed to load */}
      <PartialDataNotice
        failures={dataFailures}
        onRetry={handleRetry}
        retrying={retrying}
      />

      <div className="lesson-management-table-wrapper">
        <table className="lesson-management-table">
          <thead>
            <tr>
              <th>Дата и время</th>
              <th>Преподаватель</th>
              <th>Занято/Макс</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {Array.isArray(lessons) && lessons.length > 0 ? (
              lessons.map((lesson) => (
                <tr key={lesson.id} data-testid="lesson-row">
                  <td>
                    <div className="lesson-datetime">
                      <div>{formatDate(lesson.start_time)}</div>
                      <div className="lesson-time">
                        {formatTime(lesson.start_time)} - {formatTime(lesson.end_time)}
                      </div>
                    </div>
                  </td>
                  <td>{lesson.teacher_name}</td>
                  <td>
                    <span className={`capacity-indicator ${
                      lesson.current_students >= lesson.max_students ? 'full' : ''
                    }`}>
                      {lesson.current_students || 0} / {lesson.max_students}
                    </span>
                  </td>
                  <td>
                    <div className="lesson-actions">
                      <Button
                        size="small"
                        variant="outline"
                        onClick={() => handleOpenEdit(lesson)}
                        data-testid="edit-lesson-button"
                      >
                        Редактировать
                      </Button>
                      <Button
                        size="small"
                        variant="danger"
                        onClick={() => setDeleteLessonId(lesson.id)}
                        data-testid="delete-lesson-button"
                      >
                        Удалить
                      </Button>
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="4" className="empty-message">
                  Занятия не найдены
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Modal
        isOpen={showLessonModal}
        onClose={handleCloseModal}
        title={editingLesson ? 'Редактировать занятие' : 'Создать занятие'}
      >
        <form onSubmit={handleSubmit} className="lesson-form" data-testid="lesson-form">
          <div className="form-group">
            <label className="form-label">Преподаватель</label>
            <select
              className="form-select"
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
          </div>
          <Input
            label="Начало занятия"
            type="datetime-local"
            value={formData.start_time}
            onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
            required
          />
          <Input
            label="Конец занятия"
            type="datetime-local"
            value={formData.end_time}
            onChange={(e) => setFormData({ ...formData, end_time: e.target.value })}
            required
          />
          <Input
            label="Максимум студентов"
            type="number"
            min="1"
            max="20"
            value={formData.max_students}
            onChange={(e) => setFormData({ ...formData, max_students: e.target.value })}
            required
          />
          <Input
            label="Стоимость в кредитах"
            type="number"
            min="1"
            value={formData.credits_cost}
            onChange={(e) => setFormData({ ...formData, credits_cost: e.target.value })}
            required
          />
          <div className="lesson-form-actions">
            <Button type="button" variant="secondary" onClick={handleCloseModal}>
              Отмена
            </Button>
            <Button type="submit" loading={submitting}>
              {editingLesson ? 'Сохранить' : 'Создать'}
            </Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        isOpen={!!deleteLessonId}
        onClose={() => setDeleteLessonId(null)}
        onConfirm={handleDelete}
        title="Удаление занятия"
        message="Вы действительно хотите удалить это занятие? Кредиты будут возвращены всем записанным ученикам."
        confirmText="Удалить"
        confirmVariant="danger"
      />
    </div>
  );
};

export default LessonManagement;
