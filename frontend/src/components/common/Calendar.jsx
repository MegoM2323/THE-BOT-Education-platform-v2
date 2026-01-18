import { useState, useEffect, useCallback, useMemo, memo } from 'react';
import PropTypes from 'prop-types';
import Button from './Button.jsx';
import Spinner from './Spinner.jsx';
import { formatTime } from '../../utils/dateFormat.js';
import './Calendar.css';

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
 * Memoized LessonCard component to prevent unnecessary re-renders
 * Only re-renders when lesson data or handlers change
 */
const LessonCard = memo(({
  lesson,
  role,
  myBookings,
  getLessonClassName,
  renderLessonContent,
  getDefaultLessonColor,
  getTextColor,
  renderDefaultLessonContent,
  onLessonClick,
  onLessonRightClick,
}) => {
  const [isHovered, setIsHovered] = useState(false);

  const lessonClassName = getLessonClassName
    ? getLessonClassName(lesson)
    : getDefaultLessonColor(lesson);

  const getFallbackColor = (className) => {
    const colorMap = {
      'lesson-available': '#006600',    // dark green
      'lesson-almost-full': '#006666',  // dark cyan
      'lesson-full': '#666600',         // dark brown
      'lesson-booked': '#000066',       // dark blue
      'lesson-past': '#CCCCCC',         // gray
    };
    return colorMap[className] || '#2563eb'; // default blue
  };

  const baseColor = lesson.color || getFallbackColor(lessonClassName);
  const textColor = getTextColor();

  const isOwned = role === 'student'
    ? myBookings.some(b => b.lesson_id === lesson.id && b.status === 'active')
    : false;

  const bgAlpha = isHovered ? 0.35 : 0.15;

  // Для отфильтрованных занятий не применяем inline styles - CSS полностью контролирует внешний вид
  const filteredStyle = lesson.isFiltered ? {} : {
    borderLeftColor: baseColor,
    backgroundColor: hexToRgba(baseColor, bgAlpha),
    color: textColor,
    borderColor: baseColor
  };

  return (
    <div
      className={`calendar-lesson ${lessonClassName} ${
        lesson.applied_from_template ? 'lesson-from-template' : ''
      } ${lesson.isFiltered ? 'calendar-lesson-filtered' : ''} ${
        isOwned ? 'calendar-lesson-owned' : ''
      }`}
      style={filteredStyle}
      onClick={(e) => {
        // Не обрабатываем клики на отфильтрованные занятия
        if (lesson.isFiltered) {
          e.preventDefault();
          e.stopPropagation();
          return;
        }
        onLessonClick(lesson, e);
      }}
      onContextMenu={(e) => {
        // Не обрабатываем клики на отфильтрованные занятия
        if (lesson.isFiltered) {
          e.preventDefault();
          e.stopPropagation();
          return;
        }
        onLessonRightClick(lesson, e);
      }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      data-testid="calendar-lesson"
      title={
        lesson.applied_from_template
          ? 'Из шаблона. Нажмите для просмотра, ПКМ для редактирования'
          : 'Нажмите для просмотра'
      }
    >
      {renderLessonContent
        ? renderLessonContent(lesson)
        : renderDefaultLessonContent(lesson)
      }
    </div>
  );
}, (prevProps, nextProps) => {
  // Custom comparison: only re-render if lesson data changes
  // ВАЖНО: НЕ мемоизируем при изменении даты, чтобы избежать артефактов
  return (
    prevProps.lesson.id === nextProps.lesson.id &&
    prevProps.lesson.updated_at === nextProps.lesson.updated_at &&
    prevProps.lesson.current_students === nextProps.lesson.current_students &&
    prevProps.lesson.color === nextProps.lesson.color &&
    prevProps.lesson.start_time === nextProps.lesson.start_time &&
    prevProps.lesson.isFiltered === nextProps.lesson.isFiltered &&
    prevProps.lesson.applied_from_template === nextProps.lesson.applied_from_template &&
    prevProps.lesson.homework_text === nextProps.lesson.homework_text &&
    prevProps.lesson.homework_count === nextProps.lesson.homework_count &&
    prevProps.lesson.broadcasts_count === nextProps.lesson.broadcasts_count &&
    prevProps.myBookings?.length === nextProps.myBookings?.length
  );
});

LessonCard.displayName = 'LessonCard';

LessonCard.propTypes = {
  lesson: PropTypes.object.isRequired,
  role: PropTypes.string.isRequired,
  myBookings: PropTypes.array,
  getLessonClassName: PropTypes.func,
  renderLessonContent: PropTypes.func,
  getDefaultLessonColor: PropTypes.func.isRequired,
  getTextColor: PropTypes.func.isRequired,
  renderDefaultLessonContent: PropTypes.func.isRequired,
  onLessonClick: PropTypes.func.isRequired,
  onLessonRightClick: PropTypes.func.isRequired,
};

/**
 * Reusable Calendar component for both Admin and Student views
 *
 * @param {Object} props
 * @param {Array} props.lessons - Array of lesson objects
 * @param {string} props.role - User role: 'admin', 'teacher', 'student'
 * @param {Function} props.onLessonClick - Callback when lesson is clicked
 * @param {Function} props.onLessonRightClick - Callback when lesson is right-clicked (optional)
 * @param {Date} props.currentWeek - Current week start date (Monday)
 * @param {Function} props.onWeekChange - Callback to change week
 * @param {boolean} props.isLoading - Loading state
 * @param {string|null} props.error - Error message
 * @param {string} props.view - View mode: 'week' or 'month' (default: 'week')
 * @param {Function} props.onViewChange - Callback to change view mode (optional)
 * @param {ReactNode} props.headerActions - Additional header actions (optional)
 * @param {ReactNode} props.weekActions - Actions for week header (optional)
 * @param {Function} props.getLessonClassName - Custom function to get lesson CSS class
 * @param {Function} props.renderLessonContent - Custom function to render lesson content
 * @param {Array} props.myBookings - Student's bookings (for student view)
 * @param {boolean} props.hideNavigation - Hide navigation controls (optional)
 */
export const Calendar = ({
  lessons = [],
  role,
  onLessonClick,
  onLessonRightClick,
  currentWeek,
  onWeekChange,
  isLoading = false,
  error = null,
  view = 'week',
  onViewChange,
  headerActions,
  headerActionsTop,
  weekActions,
  navActions,
  getLessonClassName,
  renderLessonContent,
  myBookings = [],
  hideNavigation = false,
}) => {
  const [currentDate, setCurrentDate] = useState(currentWeek || new Date());
  const [isTransitioning, setIsTransitioning] = useState(false);

  // Sync internal state with prop
  useEffect(() => {
    if (currentWeek) {
      setCurrentDate(currentWeek);
    }
  }, [currentWeek]);

  const getDaysInView = useCallback(() => {
    if (view === 'week') {
      const startOfWeek = new Date(currentDate);
      const day = startOfWeek.getDay();
      const daysSinceMonday = (day + 6) % 7;  // ISO week: 0=Mon, 1=Tue, ..., 6=Sun
      startOfWeek.setDate(currentDate.getDate() - daysSinceMonday);
      startOfWeek.setHours(0, 0, 0, 0);
      const days = [];
      for (let i = 0; i < 7; i++) {
        const currentDay = new Date(startOfWeek);
        currentDay.setDate(startOfWeek.getDate() + i);
        days.push(currentDay);
      }
      return days;
    } else {
      // Month view - полная календарная сетка 6x7
      const year = currentDate.getFullYear();
      const month = currentDate.getMonth();
      const firstDay = new Date(year, month, 1);
      const lastDay = new Date(year, month + 1, 0);
      const days = [];

      // Add placeholder days from previous month
      const firstDayOfWeek = firstDay.getDay() === 0 ? 6 : firstDay.getDay() - 1;
      for (let i = firstDayOfWeek; i > 0; i--) {
        const day = new Date(year, month, 1 - i);
        days.push(day);
      }

      // Add current month days
      for (let i = 1; i <= lastDay.getDate(); i++) {
        days.push(new Date(year, month, i));
      }

      // Add placeholder days from next month to complete the grid (must be 42 days = 6 weeks)
      const daysNeeded = 42 - days.length;
      for (let i = 1; i <= daysNeeded; i++) {
        const day = new Date(year, month + 1, i);
        days.push(day);
      }

      return days;
    }
  }, [currentDate, view]);

  const getLessonsForDay = useCallback((day) => {
    // Используем локальное время вместо UTC для корректного соответствия дат
    // Преобразуем день в формат YYYY-MM-DD используя локальное время
    const year = day.getFullYear();
    const month = String(day.getMonth() + 1).padStart(2, '0');
    const date = String(day.getDate()).padStart(2, '0');
    const dayStr = `${year}-${month}-${date}`;

    return lessons.filter((lesson) => {
      const lessonTime = new Date(lesson.start_time);
      const lessonYear = lessonTime.getFullYear();
      const lessonMonth = String(lessonTime.getMonth() + 1).padStart(2, '0');
      const lessonDate = String(lessonTime.getDate()).padStart(2, '0');
      const lessonDateStr = `${lessonYear}-${lessonMonth}-${lessonDate}`;

      return lessonDateStr === dayStr;
    }).sort((a, b) => new Date(a.start_time) - new Date(b.start_time));
  }, [lessons]);

  /**
   * Определяет цвет текста (светлый/тёмный) на основе яркости фона
   * @param {string} bgColor - Цвет фона в формате #RRGGBB
   * @returns {string} - '#FFFFFF' для тёмного фона, '#000000' для светлого
   */
  const getTextColor = useCallback(() => {
    // T020: Всегда возвращаем чёрный текст
    return '#000000';
  }, []);

  const getDefaultLessonColor = useCallback((lesson) => {
    if (role === 'student') {
      // Student view: color by booking status
      const isBooked = myBookings.some(
        b => b.lesson_id === lesson.id && b.status === 'active'
      );
      const isPast = new Date(lesson.start_time) <= new Date();
      const isFull = (lesson.current_students || 0) >= lesson.max_students;

      if (isBooked) return 'lesson-booked';
      if (isPast) return 'lesson-past';
      if (isFull) {
        // Individual lessons (max_students=1) remain clickable even when full
        if (lesson.max_students === 1) {
          return 'lesson-full-individual';
        }
        return 'lesson-full';
      }
      return 'lesson-available';
    } else {
      // Admin/Teacher view: color by availability
      const spotsLeft = lesson.max_students - (lesson.current_students || 0);
      if (spotsLeft === 0) {
        // Individual lessons (max_students=1) remain clickable even when full
        if (lesson.max_students === 1) {
          return 'lesson-full-individual';
        }
        return 'lesson-full';
      }
      if (spotsLeft <= 2) return 'lesson-almost-full';
      return 'lesson-available';
    }
  }, [role, myBookings]);

  const goToPrevious = useCallback(() => {
    setIsTransitioning(true);
    const newDate = new Date(currentDate);
    if (view === 'week') {
      newDate.setDate(currentDate.getDate() - 7);
    } else {
      newDate.setMonth(currentDate.getMonth() - 1);
    }
    setCurrentDate(newDate);
    if (onWeekChange) {
      onWeekChange(newDate);
    }
    // Сбросить transition state после небольшой задержки
    setTimeout(() => setIsTransitioning(false), 150);
  }, [currentDate, view, onWeekChange]);

  const goToNext = useCallback(() => {
    setIsTransitioning(true);
    const newDate = new Date(currentDate);
    if (view === 'week') {
      newDate.setDate(currentDate.getDate() + 7);
    } else {
      newDate.setMonth(currentDate.getMonth() + 1);
    }
    setCurrentDate(newDate);
    if (onWeekChange) {
      onWeekChange(newDate);
    }
    // Сбросить transition state после небольшой задержки
    setTimeout(() => setIsTransitioning(false), 150);
  }, [currentDate, view, onWeekChange]);

  const goToToday = useCallback(() => {
    setIsTransitioning(true);
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    // IMPORTANT: Calculate Monday of current week, not raw today
    const weekStart = new Date(today);
    const day = weekStart.getDay();
    const daysSinceMonday = (day + 6) % 7;  // ISO week: 0=Mon, 1=Tue, ..., 6=Sun
    weekStart.setDate(today.getDate() - daysSinceMonday);

    setCurrentDate(weekStart);
    if (onWeekChange) {
      onWeekChange(weekStart);                // Send Monday of current week
    }
    setTimeout(() => setIsTransitioning(false), 150);
  }, [onWeekChange]);

  const getViewTitle = useCallback(() => {
    if (view === 'week') {
      const days = getDaysInView();
      const firstDay = days[0];
      const lastDay = days[days.length - 1];
      return `${firstDay.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} - ${lastDay.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })}`;
    } else {
      return currentDate.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
    }
  }, [currentDate, view, getDaysInView]);

  const handleLessonClick = useCallback((lesson, e) => {
    e.preventDefault();
    // Не обрабатываем клики на отфильтрованные занятия
    if (lesson.isFiltered) {
      return;
    }
    if (onLessonClick) {
      onLessonClick(lesson);
    }
  }, [onLessonClick]);

  const handleLessonRightClick = useCallback((lesson, e) => {
    e.preventDefault();
    // Не обрабатываем клики на отфильтрованные занятия
    if (lesson.isFiltered) {
      return;
    }
    if (onLessonRightClick) {
      onLessonRightClick(lesson);
    }
  }, [onLessonRightClick]);

  const renderDefaultLessonContent = useCallback((lesson) => {
    const spotsLeft = lesson.max_students - (lesson.current_students || 0);

    return (
      <>
        {/* T024: Новый порядок - Тема → Время → Преподаватель → Свободно X из Y */}
        {lesson.subject && (
          <div className="calendar-lesson-subject" title={lesson.subject}>
            {lesson.subject}
          </div>
        )}
        <div className="calendar-lesson-time">
          {formatTime(lesson.start_time)}
        </div>
        <div className="calendar-lesson-teacher">
          {lesson.teacher_name}
        </div>
        {/* T024: Добавлен элемент "Свободно X из Y мест" */}
        <div className="calendar-lesson-spots">
          Свободно {spotsLeft} из {lesson.max_students} мест
        </div>
        {(lesson.credits_cost ?? 1) > 0 && (
          <div className="calendar-lesson-credits">
            {lesson.credits_cost ?? 1} кр.
          </div>
        )}
      </>
    );
  }, []);

  // Memoize expensive calculations (MUST be before conditional returns)
  const days = useMemo(() => getDaysInView(), [getDaysInView]);
  const today = useMemo(() => {
    // Используем локальное время вместо UTC для корректного определения "сегодня"
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const date = String(now.getDate()).padStart(2, '0');
    return `${year}-${month}-${date}`;
  }, []);
  const viewTitle = useMemo(() => getViewTitle(), [getViewTitle]);

  if (isLoading && lessons.length === 0) {
    return (
      <div className="calendar-loading" data-testid="calendar-loading">
        <Spinner />
      </div>
    );
  }

  if (error && lessons.length === 0) {
    const errorMessage = error instanceof Error ? error.message : (error?.message || String(error));
    return (
      <div className="calendar-error" data-testid="calendar-error">
        <p>Ошибка загрузки календаря: {errorMessage}</p>
      </div>
    );
  }

  return (
    <div className="calendar" data-testid="calendar">
      {/* ВЕРХНЯЯ СТРОКА HEADER - только для admin с headerActionsTop */}
      {headerActionsTop && (
        <div className="calendar-header-wrapper">
          <div className="calendar-header-top">
            {headerActionsTop}
          </div>

          {/* НИЖНЯЯ СТРОКА HEADER - навигация + диапазон дат + кнопка "Применить шаблон" */}
          {/* T5: Скрываем только если teacher передал свою навигацию в headerActionsTop */}
          {!(role === 'teacher' && headerActionsTop) && (
            <div className="calendar-header-bottom">
              <div className="calendar-header-bottom-left">
                <div className="calendar-nav">
                  <Button size="md" onClick={goToPrevious} data-testid="prev-week">
                    ←
                  </Button>
                  <Button size="md" variant="secondary" onClick={goToToday} data-testid="today-btn">
                    Сегодня
                  </Button>
                  <Button size="md" onClick={goToNext} data-testid="next-week">
                    →
                  </Button>
                  {navActions && navActions}
                </div>
                <div className="calendar-date">
                  {viewTitle}
                </div>
              </div>

              {weekActions && (
                <div className="calendar-week-actions">
                  {weekActions}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* СТАРЫЙ ФОРМАТ HEADER - для других ролей (student/teacher) */}
      {!headerActionsTop && !hideNavigation && (
        <div className="calendar-header">
          <div className="calendar-nav">
            <Button size="md" onClick={goToPrevious} data-testid="prev-week">
              ←
            </Button>
            <Button size="md" variant="secondary" onClick={goToToday} data-testid="today-btn">
              Сегодня
            </Button>
            <Button size="md" onClick={goToNext} data-testid="next-week">
              →
            </Button>
          </div>

          {onViewChange && (
            <div className="calendar-view-toggle">
              <Button
                size="md"
                variant={view === 'week' ? 'primary' : 'secondary'}
                className={view === 'week' ? 'btn-active' : ''}
                onClick={() => onViewChange('week')}
                data-testid="view-week-btn"
              >
                Неделя
              </Button>
              <Button
                size="md"
                variant={view === 'month' ? 'primary' : 'secondary'}
                className={view === 'month' ? 'btn-active' : ''}
                onClick={() => onViewChange('month')}
                data-testid="view-month-btn"
              >
                Месяц
              </Button>
            </div>
          )}

          {headerActions && (
            <div className="calendar-header-actions">
              {headerActions}
            </div>
          )}
        </div>
      )}

      {/* calendar-week-header удалён - диапазон дат теперь во второй строке header для админа */}
      {!headerActionsTop && (
        <div className="calendar-week-header">
          <div className="calendar-view-title">{viewTitle}</div>
          {weekActions && (
            <div className="calendar-week-actions">
              {weekActions}
            </div>
          )}
        </div>
      )}

      <div
        className={`calendar-grid ${view === 'month' ? 'calendar-grid-month' : ''} ${isTransitioning ? 'calendar-grid-transitioning' : ''}`}
        key={`${currentDate.getFullYear()}-${currentDate.getMonth()}-${view}`}
      >
        {[
          { full: 'Понедельник', short: 'Пн' },
          { full: 'Вторник', short: 'Вт' },
          { full: 'Среда', short: 'Ср' },
          { full: 'Четверг', short: 'Чт' },
          { full: 'Пятница', short: 'Пт' },
          { full: 'Суббота', short: 'Сб' },
          { full: 'Воскресенье', short: 'Вс' },
        ].map((day) => (
          <div key={day.short} className="calendar-day-header">
            <span className="day-name-full">{day.full}</span>
            <span className="day-name-short">{day.short}</span>
          </div>
        ))}
        {days.map((day, index) => {
          // Используем локальное время вместо UTC для корректного отображения
          const dayYear = day.getFullYear();
          const dayMonth = String(day.getMonth() + 1).padStart(2, '0');
          const dayDate = String(day.getDate()).padStart(2, '0');
          const dayStr = `${dayYear}-${dayMonth}-${dayDate}`;

          const dayLessons = getLessonsForDay(day);
          const isToday = dayStr === today;
          const isCurrentMonth = day.getMonth() === currentDate.getMonth();

          return (
            <div
              key={`${dayStr}-${index}`}
              className={`calendar-day ${isToday ? 'calendar-day-today' : ''} ${
                !isCurrentMonth && view === 'month' ? 'calendar-day-other-month' : ''
              }`}
              data-testid={`calendar-day-${index}`}
            >
              <div className="calendar-day-number">
                {day.getDate()}
              </div>
              <div className="calendar-day-lessons">
                {dayLessons.map((lesson) => (
                  <LessonCard
                    key={`${lesson.id}-${dayStr}`}
                    lesson={lesson}
                    role={role}
                    myBookings={myBookings}
                    getLessonClassName={getLessonClassName}
                    renderLessonContent={renderLessonContent}
                    getDefaultLessonColor={getDefaultLessonColor}
                    getTextColor={getTextColor}
                    renderDefaultLessonContent={renderDefaultLessonContent}
                    onLessonClick={handleLessonClick}
                    onLessonRightClick={handleLessonRightClick}
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>

      {/* T006: Calendar legend removed - can be restored if needed
      {role === 'admin' && (
        <div className="calendar-legend">
          <div className="legend-item">
            <span className="legend-color lesson-available"></span>
            <span>Есть места</span>
          </div>
          <div className="legend-item">
            <span className="legend-color lesson-almost-full"></span>
            <span>Мало мест</span>
          </div>
          <div className="legend-item">
            <span className="legend-color lesson-full"></span>
            <span>Занято</span>
          </div>
        </div>
      )}

      {role === 'student' && (
        <div className="calendar-legend">
          <div className="legend-item">
            <span className="legend-color lesson-available"></span>
            <span>Доступно</span>
          </div>
          <div className="legend-item">
            <span className="legend-color lesson-booked"></span>
            <span>Забронировано</span>
          </div>
          <div className="legend-item">
            <span className="legend-color lesson-full"></span>
            <span>Занято</span>
          </div>
          <div className="legend-item">
            <span className="legend-color lesson-past"></span>
            <span>Прошедшие</span>
          </div>
        </div>
      )}
      */}
    </div>
  );
};

Calendar.propTypes = {
  lessons: PropTypes.arrayOf(PropTypes.object).isRequired,
  role: PropTypes.oneOf(['admin', 'teacher', 'student']).isRequired,
  onLessonClick: PropTypes.func.isRequired,
  onLessonRightClick: PropTypes.func,
  currentWeek: PropTypes.instanceOf(Date),
  onWeekChange: PropTypes.func,
  isLoading: PropTypes.bool,
  error: PropTypes.string,
  view: PropTypes.oneOf(['week', 'month']),
  onViewChange: PropTypes.func,
  headerActions: PropTypes.node,
  headerActionsTop: PropTypes.node,
  weekActions: PropTypes.node,
  navActions: PropTypes.node,
  getLessonClassName: PropTypes.func,
  renderLessonContent: PropTypes.func,
  myBookings: PropTypes.arrayOf(PropTypes.object),
  hideNavigation: PropTypes.bool,
};

export default Calendar;
