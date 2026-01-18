import { useState, useMemo } from 'react';
import SharedCalendar from '../common/Calendar.jsx';
import Button from '../common/Button.jsx';
import TeacherLessonModal from './TeacherLessonModal.jsx';
import { useTeacherSchedule } from '../../hooks/useTeacherSchedule.js';
import { formatDateRange } from '../../utils/dateFormat.js';
import './TeacherCalendar.css';

/**
 * Получить понедельник текущей недели
 */
const getMonday = (date) => {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  const day = d.getDay();
  const daysSinceMonday = (day + 6) % 7;
  d.setDate(d.getDate() - daysSinceMonday);
  return d;
};

/**
 * Добавить дни к дате
 */
const addDays = (date, days) => {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
};

/**
 * Вычесть дни из даты
 */
const subDays = (date, days) => {
  const result = new Date(date);
  result.setDate(result.getDate() - days);
  return result;
};

/**
 * Календарь преподавателя
 * Показывает только занятия текущего преподавателя
 * При клике открывается read-only модал с homework и broadcasts
 */
export const TeacherCalendar = () => {
  const [calendarDate, setCalendarDate] = useState(new Date());
  const [selectedLesson, setSelectedLesson] = useState(null);
  const [showModal, setShowModal] = useState(false);
  const [viewMode, setViewMode] = useState('week');
  const [subjectFilter] = useState('');
  const [showIndividual, setShowIndividual] = useState(true);
  const [showGroup, setShowGroup] = useState(true);
  const [hideFull, setHideFull] = useState(false);

  // Расчёт дат для запроса в зависимости от viewMode
  const { startDate, endDate, weekStart, weekEnd } = useMemo(() => {
    const ws = getMonday(calendarDate);
    const we = addDays(ws, 6);

    if (viewMode === 'week') {
      return { startDate: ws, endDate: we, weekStart: ws, weekEnd: we };
    }
    // Для месячного view: покрываем 6 недель (42 дня)
    const firstDayOfMonth = new Date(calendarDate.getFullYear(), calendarDate.getMonth(), 1);
    const dayOfWeek = firstDayOfMonth.getDay();
    const daysToMonday = (dayOfWeek + 6) % 7;
    const monthViewStart = subDays(firstDayOfMonth, daysToMonday);
    const monthViewEnd = addDays(monthViewStart, 41);
    return { startDate: monthViewStart, endDate: monthViewEnd, weekStart: ws, weekEnd: we };
  }, [viewMode, calendarDate]);

  // Загрузка занятий преподавателя
  const { data: lessons = [], isLoading, error } = useTeacherSchedule(
    startDate.toISOString(),
    endDate.toISOString()
  );

  /**
   * Применение фильтров - НЕ скрываем занятия, а помечаем их как isFiltered
   * Отфильтрованные занятия отображаются серыми (opacity + grayscale)
   */
  const filteredLessons = useMemo(() => {
    return lessons.map(lesson => {
      let isFiltered = false;

      // Фильтр по предмету
      if (subjectFilter && lesson.subject !== subjectFilter) {
        isFiltered = true;
      }

      // Фильтр по типу занятия (индивидуальное/групповое)
      const isIndividual = lesson.max_students === 1;
      if (!showIndividual && isIndividual) {
        isFiltered = true;
      }
      if (!showGroup && !isIndividual) {
        isFiltered = true;
      }

      // Фильтр "Скрыть полные"
      if (hideFull && lesson.current_students >= lesson.max_students) {
        isFiltered = true;
      }

      return { ...lesson, isFiltered };
    });
  }, [lessons, subjectFilter, showIndividual, showGroup, hideFull]);

  /**
   * Обработчик клика на занятие
   */
  const handleLessonClick = (lesson) => {
    setSelectedLesson(lesson);
    setShowModal(true);
  };

  /**
   * Закрыть модал
   */
  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedLesson(null);
  };

  /**
   * Навигация: сегодня
   */
  const handleToday = () => {
    setCalendarDate(new Date());
  };

  /**
   * Кастомный рендер для занятий преподавателя
   * Показывает: тема, время, количество учеников (без имен)
   */
  const renderLessonContent = (lesson) => {
    const currentStudents = lesson.current_students || 0;
    const maxStudents = lesson.max_students;
    const studentsText = `${currentStudents} из ${maxStudents}`;

    return (
      <>
        {lesson.subject && (
          <div className="teacher-lesson-subject" title={lesson.subject}>
            {lesson.subject.length > 30 ? `${lesson.subject.slice(0, 30)}...` : lesson.subject}
          </div>
        )}
        <div className="teacher-lesson-time">
          {new Date(lesson.start_time).toLocaleTimeString('ru-RU', {
            hour: '2-digit',
            minute: '2-digit'
          })}
        </div>
        <div className="teacher-lesson-students" title={`Записано: ${studentsText}`}>
          {studentsText}
        </div>
      </>
    );
  };

  /**
   * Кастомный className для занятий
   */
  const getLessonClassName = (lesson) => {
    const isPast = new Date(lesson.start_time) < new Date();
    if (isPast) return 'teacher-lesson-past';

    const spotsLeft = lesson.max_students - (lesson.current_students || 0);
    if (spotsLeft === 0) return 'teacher-lesson-full';
    if (spotsLeft <= 2) return 'teacher-lesson-almost-full';
    return 'teacher-lesson-available';
  };

  // Получить диапазон дат для отображения
  const dateRangeText = formatDateRange(weekStart, weekEnd);

  /**
   * Навигация: предыдущий период (неделя или месяц)
   */
  const handlePrev = () => {
    if (viewMode === 'week') {
      setCalendarDate(prev => subDays(prev, 7));
    } else {
      setCalendarDate(prev => {
        const newDate = new Date(prev);
        newDate.setMonth(newDate.getMonth() - 1);
        return newDate;
      });
    }
  };

  /**
   * Навигация: следующий период (неделя или месяц)
   */
  const handleNext = () => {
    if (viewMode === 'week') {
      setCalendarDate(prev => addDays(prev, 7));
    } else {
      setCalendarDate(prev => {
        const newDate = new Date(prev);
        newDate.setMonth(newDate.getMonth() + 1);
        return newDate;
      });
    }
  };

  /**
   * Получить текст для отображения периода
   */
  const getViewTitle = () => {
    if (viewMode === 'week') {
      return dateRangeText;
    } else {
      return calendarDate.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
    }
  };

  // Комбинированный header с навигацией и фильтрами
  const teacherHeaderActions = (
    <div className="teacher-calendar-header-wrapper">
      <div className="teacher-calendar-left-section">
        <div className="teacher-view-toggle">
          <Button
            size="md"
            variant={viewMode === 'week' ? 'primary' : 'secondary'}
            className={viewMode === 'week' ? 'btn-active' : ''}
            onClick={() => setViewMode('week')}
            data-testid="view-week-btn"
          >
            Неделя
          </Button>
          <Button
            size="md"
            variant={viewMode === 'month' ? 'primary' : 'secondary'}
            className={viewMode === 'month' ? 'btn-active' : ''}
            onClick={() => setViewMode('month')}
            data-testid="view-month-btn"
          >
            Месяц
          </Button>
        </div>
        <div className="teacher-calendar-nav-wrapper">
          <div className="teacher-calendar-nav-buttons">
            <Button
              size="md"
              onClick={handlePrev}
              data-testid="prev-week"
              aria-label={viewMode === 'week' ? 'Предыдущая неделя' : 'Предыдущий месяц'}
            >
              ←
            </Button>
            <Button
              size="md"
              variant="secondary"
              onClick={handleToday}
              data-testid="today-btn"
            >
              Сегодня
            </Button>
            <Button
              size="md"
              onClick={handleNext}
              data-testid="next-week"
              aria-label={viewMode === 'week' ? 'Следующая неделя' : 'Следующий месяц'}
            >
              →
            </Button>
          </div>
          <div className="teacher-calendar-date-range">
            <span className="teacher-date-range-text">{getViewTitle()}</span>
          </div>
        </div>
      </div>
      <div className="teacher-filters-inline">
        <label className="teacher-filter-checkbox">
          <input
            type="checkbox"
            checked={showIndividual}
            onChange={(e) => setShowIndividual(e.target.checked)}
            data-testid="filter-show-individual"
          />
          <span>Индивидуальные</span>
        </label>

        <label className="teacher-filter-checkbox">
          <input
            type="checkbox"
            checked={showGroup}
            onChange={(e) => setShowGroup(e.target.checked)}
            data-testid="filter-show-group"
          />
          <span>Групповые</span>
        </label>

        <label className="teacher-filter-checkbox">
          <input
            type="checkbox"
            checked={hideFull}
            onChange={(e) => setHideFull(e.target.checked)}
            data-testid="filter-hide-full"
          />
          <span>Скрыть полные</span>
        </label>
      </div>
    </div>
  );

  return (
    <div className="teacher-calendar-container" data-testid="teacher-calendar">
      <div className="teacher-calendar-header">
        <h2 className="teacher-calendar-title">График занятий</h2>
      </div>

      <SharedCalendar
        lessons={filteredLessons}
        role="teacher"
        onLessonClick={handleLessonClick}
        currentWeek={calendarDate}
        onWeekChange={setCalendarDate}
        isLoading={isLoading}
        error={error ? (error?.message || String(error)) : null}
        view={viewMode}
        onViewChange={setViewMode}
        headerActionsTop={teacherHeaderActions}
        renderLessonContent={renderLessonContent}
        getLessonClassName={getLessonClassName}
      />

      {/* Модал для просмотра занятия */}
      {selectedLesson && (
        <TeacherLessonModal
          isOpen={showModal}
          onClose={handleCloseModal}
          lesson={selectedLesson}
        />
      )}
    </div>
  );
};

export default TeacherCalendar;
