import { useState, useEffect, useMemo, useCallback, memo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import Spinner from '../common/Spinner.jsx';
import ConfirmModal from '../common/ConfirmModal.jsx';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import TemplateLessonForm from './TemplateLessonForm.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import { useTemplate } from '../../hooks/useTemplates.js';
import apiClient from '../../api/client.js';
import { getContrastTextColor } from '../../utils/colorUtils.js';
import './TemplateCalendarView.css';
import '../common/Calendar.css';

const DAYS_OF_WEEK = [
  { value: 0, label: 'Понедельник', short: 'Пн' },
  { value: 1, label: 'Вторник', short: 'Вт' },
  { value: 2, label: 'Среда', short: 'Ср' },
  { value: 3, label: 'Четверг', short: 'Чт' },
  { value: 4, label: 'Пятница', short: 'Пт' },
  { value: 5, label: 'Суббота', short: 'Сб' },
  { value: 6, label: 'Воскресенье', short: 'Вс' },
];

/**
 * Конвертирует HEX цвет в RGBA формат
 * @param {string} hex - Цвет в формате #RRGGBB
 * @param {number} alpha - Прозрачность (0-1)
 * @returns {string} - Цвет в формате rgba(r, g, b, alpha)
 */
const hexToRgba = (hex, alpha = 1) => {
  if (!hex || !hex.startsWith('#')) return `rgba(102, 0, 0, ${alpha})`; // fallback to dark red

  const cleanHex = hex.replace('#', '');
  const r = parseInt(cleanHex.substr(0, 2), 16);
  const g = parseInt(cleanHex.substr(2, 2), 16);
  const b = parseInt(cleanHex.substr(4, 2), 16);

  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};

/**
 * Форматирует время HH:MM:SS в HH:MM
 * @param {string} time - Время в формате HH:MM:SS
 * @returns {string} - Время в формате HH:MM
 */
const formatTimeShort = (time) => {
  if (!time) return '';
  // Если формат уже HH:MM, возвращаем как есть
  if (time.length === 5) return time;
  // Если HH:MM:SS, отрезаем секунды
  return time.substring(0, 5);
};

/**
 * Memoized TemplateLessonCard component to prevent unnecessary re-renders
 * Only re-renders when lesson data changes
 * Компактный формат: 2 строки (как в календаре)
 */
const TemplateLessonCard = memo(({ lesson, onClick }) => {
  const baseColor = lesson.color || '#004231';

  // Вычисляем количество занятых мест из student_ids или students
  const currentStudents = lesson.student_ids?.length || lesson.students?.length || 0;
  const spotsLeft = lesson.max_students - currentStudents;

  // Определяем класс занятия по заполненности (как в календаре расписания)
  const getLessonClassName = () => {
    if (spotsLeft === 0) {
      if (lesson.max_students === 1) {
        return 'lesson-full-individual';
      }
      return 'lesson-full';
    }
    if (spotsLeft <= 2) return 'lesson-almost-full';
    return 'lesson-available';
  };

  const lessonClassName = getLessonClassName();
  const textColor = '#000000'; // Всегда чёрный текст (как в календаре)

  // Получаем credits_cost, по умолчанию 1 (как в TemplateLessonForm)
  // Старые занятия созданные до миграции 045 имеют NULL в БД, преобразуется в undefined
  const creditsCost = lesson.credits_cost ?? 1;

  console.log(`[TemplateLessonCard] Lesson: ${lesson.subject} at ${lesson.start_time}`, {
    raw_credits_cost: lesson.credits_cost,
    final_creditsCost: creditsCost,
    lesson_id: lesson.id,
    teacher_name: lesson.teacher_name,
  });

  return (
    <div
      className={`calendar-lesson ${lessonClassName} template-lesson-card ${lesson.max_students === 1 ? 'individual' : 'group'}`}
      data-testid="template-lesson-card"
      onClick={() => onClick(lesson)}
      style={{
        borderLeftColor: baseColor,
        backgroundColor: hexToRgba(baseColor, 0.15),
        color: textColor,
        borderColor: baseColor,
        cursor: 'pointer',
      }}
      title="Нажмите для просмотра"
    >
      {/* Предмет */}
      {lesson.subject && (
        <div className="calendar-lesson-subject" title={lesson.subject}>
          {lesson.subject}
        </div>
      )}
      
      {/* Время */}
      <div className="calendar-lesson-time">
        {formatTimeShort(lesson.start_time)}
      </div>
      
      {/* Преподаватель */}
      <div className="calendar-lesson-teacher">
        {lesson.teacher_name || 'Неизвестный учитель'}
      </div>
      
      {/* Свободно X из Y мест */}
      <div className="calendar-lesson-spots">
        Свободно {spotsLeft} из {lesson.max_students} мест
      </div>

      {/* Кредиты - показываем всегда с правильным значением (по умолчанию 1, как в форме) */}
      <div className="calendar-lesson-credits">
        {creditsCost} {creditsCost === 1 ? 'кредит' : creditsCost < 5 ? 'кредита' : 'кредитов'}
      </div>
    </div>
  );
}, (prevProps, nextProps) => {
  // Custom comparison: only re-render if lesson data changes
  const prevStudentCount = prevProps.lesson.student_ids?.length || prevProps.lesson.students?.length || 0;
  const nextStudentCount = nextProps.lesson.student_ids?.length || nextProps.lesson.students?.length || 0;
  return (
    prevProps.lesson.id === nextProps.lesson.id &&
    prevProps.lesson.start_time === nextProps.lesson.start_time &&
    prevProps.lesson.end_time === nextProps.lesson.end_time &&
    prevProps.lesson.teacher_name === nextProps.lesson.teacher_name &&
    prevProps.lesson.color === nextProps.lesson.color &&
    prevProps.lesson.subject === nextProps.lesson.subject &&
    prevProps.lesson.max_students === nextProps.lesson.max_students &&
    prevProps.lesson.credits_cost === nextProps.lesson.credits_cost &&
    prevStudentCount === nextStudentCount
  );
});

TemplateLessonCard.displayName = 'TemplateLessonCard';

export const TemplateCalendarView = ({ templateId }) => {
  const notification = useNotification();
  const queryClient = useQueryClient();
  const { template, loading, refetch } = useTemplate(templateId);
  const [deleteLessonId, setDeleteLessonId] = useState(null);
  const [editingLesson, setEditingLesson] = useState(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [addingLessonForDay, setAddingLessonForDay] = useState(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showGlobalAddModal, setShowGlobalAddModal] = useState(false);

  useEffect(() => {
    if (templateId) {
      refetch();
    }
  }, [templateId, refetch]);

  useEffect(() => {
    if (template) {
      console.log('[TemplateCalendarView] Template loaded:', {
        template_id: template.id,
        template_name: template.name,
        lessons_count: template.lessons?.length || 0,
        lessons: template.lessons?.map(lesson => ({
          id: lesson.id,
          subject: lesson.subject,
          day_of_week: lesson.day_of_week,
          start_time: lesson.start_time,
          teacher_name: lesson.teacher_name,
          credits_cost: lesson.credits_cost,
          credits_cost_type: typeof lesson.credits_cost,
          max_students: lesson.max_students,
          color: lesson.color,
        })) || []
      });
    }
  }, [template]);

  // Memoize getLessonsForDay to prevent recalculation on every render
  const getLessonsForDay = useCallback((dayOfWeek) => {
    if (!template || !template.lessons) return [];
    const filtered = template.lessons
      .filter((lesson) => lesson.day_of_week === dayOfWeek)
      .sort((a, b) => {
        const [aH, aM] = a.start_time.split(':').map(Number);
        const [bH, bM] = b.start_time.split(':').map(Number);
        return aH * 60 + aM - (bH * 60 + bM);
      });

    if (filtered.length > 0) {
      console.log(`[TemplateCalendarView] Lessons for day ${dayOfWeek}:`, {
        count: filtered.length,
        lessons: filtered.map(l => ({
          id: l.id,
          subject: l.subject,
          start_time: l.start_time,
          credits_cost: l.credits_cost,
        }))
      });
    }

    return filtered;
  }, [template]);

  // Memoize total lesson count
  const totalLessonCount = useMemo(() => {
    return template?.lessons?.length || 0;
  }, [template?.lessons?.length]);

  const handleDeleteLesson = useCallback(async () => {
    if (!deleteLessonId) return;

    try {
      await apiClient.delete(`/templates/${templateId}/lessons/${deleteLessonId}`);
      notification.success('Урок удален');
      setDeleteLessonId(null);

      // Инвалидация кэша для немедленного обновления
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      queryClient.invalidateQueries({ queryKey: ['templates', templateId] });
    } catch (error) {
      notification.error(error.message || 'Ошибка удаления урока');
    }
  }, [deleteLessonId, templateId, notification, queryClient]);

  const handleLessonClick = useCallback((lesson) => {
    setEditingLesson(lesson);
    setShowEditModal(true);
  }, []);


  const handleAddLesson = useCallback((dayOfWeek) => {
    setAddingLessonForDay(dayOfWeek);
    setShowAddModal(true);
  }, []);

  const handleAddLessonSaved = useCallback(() => {
    setShowAddModal(false);
    setAddingLessonForDay(null);
    // Кэш уже инвалидирован в TemplateLessonForm
  }, []);

  const handleEditLessonSaved = useCallback(() => {
    setShowEditModal(false);
    setEditingLesson(null);
    // Кэш уже инвалидирован в TemplateLessonForm
  }, []);

  if (!templateId) {
    return (
      <div className="template-calendar-empty">
        <p>Выберите шаблон для просмотра занятий</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="template-calendar-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="template-calendar-view">
      <div className="template-calendar-header">
        <div className="template-calendar-header-content">
          <div>
            <h3 className="template-calendar-title">Расписание на неделю</h3>
            <p className="template-calendar-subtitle">
              {totalLessonCount} занятие(й) в шаблоне
            </p>
          </div>
          <Button
            variant="primary"
            onClick={() => setShowGlobalAddModal(true)}
          >
            Добавить занятие
          </Button>
        </div>
      </div>

      <div className="template-calendar-grid">
        {DAYS_OF_WEEK.map((day) => (
          <div key={day.value} className="template-calendar-day">
            <div className="template-calendar-day-header">
              <div className="day-header-content">
                <div>
                  <span className="day-name-full">{day.label}</span>
                  <span className="day-name-short">{day.short}</span>
                </div>
              </div>
            </div>
            <div className="template-calendar-day-lessons">
              {getLessonsForDay(day.value).map((lesson) => (
                <TemplateLessonCard
                  key={lesson.id}
                  lesson={lesson}
                  onClick={handleLessonClick}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

<Modal
        isOpen={showGlobalAddModal}
        onClose={() => setShowGlobalAddModal(false)}
        title="Добавить занятие"
      >
        <TemplateLessonForm
          templateId={templateId}
          onSave={() => {
            setShowGlobalAddModal(false);
            // Кэш уже инвалидирован в TemplateLessonForm
          }}
          onCancel={() => setShowGlobalAddModal(false)}
        />
      </Modal>

      <Modal
        isOpen={showAddModal}
        onClose={() => {
          setShowAddModal(false);
          setAddingLessonForDay(null);
        }}
        title="Добавить урок"
      >
        <TemplateLessonForm
          templateId={templateId}
          forceDay={addingLessonForDay}
          onSave={handleAddLessonSaved}
          onCancel={() => {
            setShowAddModal(false);
            setAddingLessonForDay(null);
          }}
        />
      </Modal>

      <Modal
        isOpen={showEditModal}
        onClose={async () => {
          // При закрытии через × вызываем handleClose из формы
          // Это сохранит изменения перед закрытием
          try {
            if (window.__templateLessonFormHandleClose) {
              await window.__templateLessonFormHandleClose();
            }
          } catch (error) {
            console.error('Error saving lesson on close:', error);
          } finally {
            // Очищаем ссылку
            delete window.__templateLessonFormHandleClose;
            setShowEditModal(false);
            setEditingLesson(null);
          }
        }}
        title="Редактировать урок"
      >
        <TemplateLessonForm
          templateId={templateId}
          editingLesson={editingLesson}
          onSave={() => {
            delete window.__templateLessonFormHandleClose;
            setShowEditModal(false);
            setEditingLesson(null);
            handleEditLessonSaved();
          }}
          onCancel={async () => {
            // onCancel уже вызывает handleCancel, который сохраняет изменения
            delete window.__templateLessonFormHandleClose;
            setShowEditModal(false);
            setEditingLesson(null);
          }}
          onClose={(handleClose) => {
            // Сохраняем функцию handleClose для вызова при закрытии модального окна
            window.__templateLessonFormHandleClose = handleClose;
          }}
        />
      </Modal>

      <ConfirmModal
        isOpen={!!deleteLessonId}
        onClose={() => setDeleteLessonId(null)}
        onConfirm={handleDeleteLesson}
        title="Удалить урок"
        message="Вы уверены, что хотите удалить этот урок из шаблона?"
        confirmText="Удалить"
        confirmVariant="danger"
      />
    </div>
  );
};

export default TemplateCalendarView;
