import { useState, useMemo, useEffect } from "react";
import { useStudentLessons } from "../hooks/useStudentLessons.js";
import { useMyBookings } from "../hooks/useMyBookings.js";
import { useStudentCredits } from "../hooks/useStudentCredits.js";
import { useSlowConnection } from "../hooks/useSlowConnection.js";
import { logger } from "../utils/logger.js";
import SharedCalendar from "../components/common/Calendar.jsx";
import LessonBookingModal from "../components/student/LessonBookingModal.jsx";
import LessonCard from "../components/student/LessonCard.jsx";
import Spinner from "../components/common/Spinner.jsx";
import Button from "../components/common/Button.jsx";
import CreditsCounter from "../components/student/CreditsCounter.jsx";
import {
  SkeletonCalendar,
  SkeletonList,
} from "../components/common/SkeletonLoader.jsx";
import SlowConnectionNotice from "../components/common/SlowConnectionNotice.jsx";
import { getUserTimezone, getTimezoneOffset } from "../utils/dateFormat.js";
import "./StudentSchedule.css";

/**
 * Get Monday of current week
 */
const getMonday = (date = new Date()) => {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  const day = d.getDay();
  const daysSinceMonday = (day + 6) % 7;  // ISO week: 0=Mon, 1=Tue, ..., 6=Sun
  d.setDate(d.getDate() - daysSinceMonday);
  return d;
};

/**
 * Format date to YYYY-MM-DD using local time (not UTC)
 */
const formatDate = (date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};

/**
 * StudentSchedule - Main student calendar page
 */
export const StudentSchedule = () => {
  // Автоматическое определение viewMode на основе размера экрана
  const [viewMode, setViewMode] = useState(() => {
    // При инициализации проверяем размер экрана
    if (typeof window !== 'undefined') {
      return window.innerWidth <= 768 ? "list" : "calendar";
    }
    return "calendar";
  });
  const [calendarView, setCalendarView] = useState("month"); // 'week' | 'month' - для календаря
  const [currentWeek, setCurrentWeek] = useState(getMonday());
  const [selectedLesson, setSelectedLesson] = useState(null);
  const [showBookingModal, setShowBookingModal] = useState(false);
  const [dismissedSlowNotice, setDismissedSlowNotice] = useState(false);

  // Автоматическое переключение на список при работе с телефона
  useEffect(() => {
    const handleResize = () => {
      const isMobile = window.innerWidth <= 768;
      if (isMobile) {
        setViewMode("list");
      } else {
        setViewMode("calendar");
      }
    };

    // Проверяем размер экрана при монтировании
    handleResize();

    // Добавляем обработчик изменения размера окна
    window.addEventListener("resize", handleResize);

    // Очистка при размонтировании
    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  // Filter state
  const [showIndividual, setShowIndividual] = useState(true);
  const [showGroup, setShowGroup] = useState(true);
  const [hideFull, setHideFull] = useState(false);
  const [myBookings, setMyBookings] = useState(false);
  const [selectedTeacher, setSelectedTeacher] = useState(null);

  const weekStartDate = formatDate(currentWeek);

  // Load data
  const {
    lessons,
    isLoading: lessonsLoading,
    error: lessonsError,
  } = useStudentLessons(weekStartDate);

  // Only get myBookings for checking if lessons are already booked
  const {
    myBookings: myBookingsData,
    isLoading: bookingsLoading,
    error: bookingsError,
    refetch: refetchBookings,
  } = useMyBookings();

  const { credits, isLoading: creditsLoading } = useStudentCredits();

  // Detect slow connection (show warning after 3 seconds)
  const isAnyLoading = lessonsLoading || bookingsLoading || creditsLoading;
  const isSlow = useSlowConnection(isAnyLoading, 3000);

  // Extract unique teachers from lessons
  const teachers = useMemo(() => {
    if (!Array.isArray(lessons)) return [];

    const teacherMap = new Map();
    lessons.forEach((lesson) => {
      if (lesson.teacher_id && lesson.teacher_name) {
        teacherMap.set(lesson.teacher_id, lesson.teacher_name);
      }
    });

    return Array.from(teacherMap.entries()).map(([id, name]) => ({ id, name }));
  }, [lessons]);

  // Apply filters to lessons
  const filteredLessons = useMemo(() => {
    if (!Array.isArray(lessons)) return [];

    // Для режима списка: вычисляем дату 2 недели назад
    const twoWeeksAgo = viewMode === "list" 
      ? new Date(Date.now() - 14 * 24 * 60 * 60 * 1000)
      : null;

    return lessons.map((lesson) => {
      let isFiltered = false;

      // Filter by date in list mode: show only lessons from last 2 weeks and future
      if (viewMode === "list" && twoWeeksAgo && lesson.start_time) {
        const lessonDate = new Date(lesson.start_time);
        if (lessonDate < twoWeeksAgo) {
          isFiltered = true;
        }
      }

      // Filter by lesson type (individual vs group)
      if (lesson.max_students === 1 && !showIndividual) {
        isFiltered = true;
      }
      if (lesson.max_students > 1 && !showGroup) {
        isFiltered = true;
      }

      // Filter by full lessons
      if (hideFull && lesson.current_students >= lesson.max_students) {
        isFiltered = true;
      }

      // Filter by teacher
      if (selectedTeacher && lesson.teacher_id !== selectedTeacher) {
        isFiltered = true;
      }

      // Filter by my bookings (show only lessons I've booked)
      if (myBookings) {
        const isBooked =
          Array.isArray(myBookingsData) && !bookingsError
            ? myBookingsData.some(
                (b) => b?.lesson_id === lesson?.id && b?.status === "active",
              )
            : false;
        if (!isBooked) {
          isFiltered = true;
        }
      }

      return { ...lesson, isFiltered };
    });
  }, [
    lessons,
    viewMode,
    showIndividual,
    showGroup,
    hideFull,
    selectedTeacher,
    myBookings,
    myBookingsData,
    bookingsError,
  ]);

  const handleWeekChange = (newWeek) => {
    setCurrentWeek(newWeek);
  };

  const handleLessonClick = (lesson) => {
    // Отладочное логирование для проверки данных урока
    console.log('[StudentSchedule] handleLessonClick - lesson from calendar:', JSON.stringify({
      id: lesson.id,
      max_students: lesson.max_students,
      current_students: lesson.current_students,
      subject: lesson.subject
    }, null, 2));
    console.log('[StudentSchedule] Full lesson object from calendar:', lesson);
    
    // Находим полный объект урока из исходного массива lessons
    // чтобы убедиться, что все поля присутствуют
    const fullLesson = lessons.find(l => l.id === lesson.id) || lesson;
    
    console.log('[StudentSchedule] Full lesson from lessons array:', JSON.stringify({
      id: fullLesson.id,
      max_students: fullLesson.max_students,
      current_students: fullLesson.current_students,
      subject: fullLesson.subject,
      found: !!lessons.find(l => l.id === lesson.id)
    }, null, 2));
    console.log('[StudentSchedule] Full lesson object from array:', fullLesson);
    
    // Убеждаемся, что передаем полный объект урока со всеми полями
    setSelectedLesson(fullLesson);
    setShowBookingModal(true);
  };

  const handleCloseModal = () => {
    setShowBookingModal(false);
    setSelectedLesson(null);
  };

  // Получаем информацию о часовом поясе пользователя
  const userTimezone = getUserTimezone();
  const timezoneOffset = getTimezoneOffset();

  // Log bookings error for debugging
  if (bookingsError) {
    logger.error(
      "StudentSchedule: Failed to load bookings:",
      bookingsError?.message || bookingsError,
    );
  }

  if (lessonsLoading || bookingsLoading || creditsLoading) {
    return (
      <div className="student-schedule" data-testid="student-schedule">
        {/* Row 1: Title (skeleton mode) */}
        <div className="student-schedule-header">
          <div className="student-schedule-title-wrapper">
            <h1 className="student-schedule-title">Расписание</h1>
            <span
              className="timezone-info"
              title={`Часовой пояс: ${getUserTimezone()}`}
            >
              Время: UTC{getTimezoneOffset()}
            </span>
          </div>
          <div className="student-schedule-credits">
            <CreditsCounter inline />
          </div>
        </div>

        {/* Show appropriate skeleton based on view mode */}
        <div aria-busy="true" aria-live="polite">
          {viewMode === "calendar" ? (
            <SkeletonCalendar />
          ) : (
            <SkeletonList items={6} />
          )}
        </div>

        {/* Show slow connection notice if loading takes too long */}
        {isSlow && !dismissedSlowNotice && (
          <SlowConnectionNotice
            onDismiss={() => setDismissedSlowNotice(true)}
          />
        )}
      </div>
    );
  }

  return (
    <div className="student-schedule" data-testid="student-schedule">
      {/* Bookings Error Banner */}
      {bookingsError && (
        <div
          className="error-banner"
          role="alert"
          data-testid="bookings-error-banner"
        >
          <span className="banner-icon">❌</span>
          <span className="banner-text">
            Ошибка загрузки ваших бронирований:{" "}
            {bookingsError?.message || "Неизвестная ошибка"}
          </span>
          <button
            className="retry-btn"
            onClick={() => refetchBookings()}
            data-testid="retry-bookings-btn"
          >
            Повторить попытку
          </button>
        </div>
      )}

      {/* Row 1: Title */}
      <div
        className="student-schedule-header"
        data-testid="student-schedule-header"
      >
        <div className="student-schedule-title-wrapper">
          <h1 className="student-schedule-title">Расписание</h1>
          <span
            className="timezone-info"
            title={`Часовой пояс: ${userTimezone}`}
          >
            Время: UTC{timezoneOffset}
          </span>
        </div>
        <div className="student-schedule-credits">
          <CreditsCounter inline />
        </div>
      </div>

      {/* Content Area - Calendar/List */}
      <main className="student-schedule-content">
        {/* Error message */}
        {lessonsError && (
          <div className="error-message" role="alert">
            <p>
              Ошибка загрузки занятий:{" "}
              {lessonsError?.message || String(lessonsError)}
            </p>
          </div>
        )}

        {/* Calendar - Show only in calendar mode */}
        {viewMode === "calendar" && (
          <div className="calendar-view" data-testid="calendar-view">
            <SharedCalendar
              lessons={filteredLessons}
              role="student"
              onLessonClick={handleLessonClick}
              currentWeek={currentWeek}
              onWeekChange={handleWeekChange}
              isLoading={lessonsLoading}
              error={lessonsError ? (lessonsError?.message || String(lessonsError)) : null}
              view={calendarView}
              onViewChange={setCalendarView}
              myBookings={
                bookingsError || !Array.isArray(myBookingsData)
                  ? []
                  : myBookingsData
              }
              headerActionsTop={
                <>
                  <div className="calendar-view-toggle">
                    <Button
                      size="md"
                      variant={calendarView === "week" ? "primary" : "secondary"}
                      className={calendarView === "week" ? "btn-active" : ""}
                      onClick={() => setCalendarView("week")}
                      data-testid="view-week-btn"
                    >
                      Неделя
                    </Button>
                    <Button
                      size="md"
                      variant={calendarView === "month" ? "primary" : "secondary"}
                      className={calendarView === "month" ? "btn-active" : ""}
                      onClick={() => setCalendarView("month")}
                      data-testid="view-month-btn"
                    >
                      Месяц
                    </Button>
                  </div>
                  <div className="calendar-header-top-actions">
                    {/* Filters - inline как у админа */}
                    <div className="calendar-filters-inline">
                      <label className="filter-checkbox">
                        <input
                          type="checkbox"
                          checked={showIndividual}
                          onChange={(e) => setShowIndividual(e.target.checked)}
                          data-testid="filter-show-individual"
                        />
                        <span>Индивид.</span>
                      </label>

                      <label className="filter-checkbox">
                        <input
                          type="checkbox"
                          checked={showGroup}
                          onChange={(e) => setShowGroup(e.target.checked)}
                          data-testid="filter-show-group"
                        />
                        <span>Группы</span>
                      </label>

                      <label className="filter-checkbox">
                        <input
                          type="checkbox"
                          checked={hideFull}
                          onChange={(e) => setHideFull(e.target.checked)}
                          data-testid="filter-hide-full"
                        />
                        <span>Скрыть полные</span>
                      </label>

                      <label className="filter-checkbox">
                        <input
                          type="checkbox"
                          checked={myBookings}
                          onChange={(e) => setMyBookings(e.target.checked)}
                          data-testid="filter-my-bookings"
                        />
                        <span>Мои занятия</span>
                      </label>

                      <select
                        className="filter-select"
                        value={selectedTeacher || ""}
                        onChange={(e) =>
                          setSelectedTeacher(e.target.value || null)
                        }
                        data-testid="filter-teacher"
                      >
                        <option value="">Все преподаватели</option>
                        {teachers.map((teacher) => (
                          <option key={teacher.id} value={teacher.id}>
                            {teacher.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                </>
              }
            />
          </div>
        )}

        {/* List View - Shows when in list mode */}
        {viewMode === "list" && (
          <div className="list-view" data-testid="list-view">
            {/* Filters bar for list view */}
            <div className="student-schedule-list-filters">
              <div className="calendar-filters-inline">
                <label className="filter-checkbox">
                  <input
                    type="checkbox"
                    checked={showIndividual}
                    onChange={(e) => setShowIndividual(e.target.checked)}
                    data-testid="filter-show-individual"
                  />
                  <span>Индивид.</span>
                </label>

                <label className="filter-checkbox">
                  <input
                    type="checkbox"
                    checked={showGroup}
                    onChange={(e) => setShowGroup(e.target.checked)}
                    data-testid="filter-show-group"
                  />
                  <span>Группы</span>
                </label>

                <label className="filter-checkbox">
                  <input
                    type="checkbox"
                    checked={hideFull}
                    onChange={(e) => setHideFull(e.target.checked)}
                    data-testid="filter-hide-full"
                  />
                  <span>Скрыть полные</span>
                </label>

                <label className="filter-checkbox">
                  <input
                    type="checkbox"
                    checked={myBookings}
                    onChange={(e) => setMyBookings(e.target.checked)}
                    data-testid="filter-my-bookings"
                  />
                  <span>Мои занятия</span>
                </label>

                <select
                  className="filter-select"
                  value={selectedTeacher || ""}
                  onChange={(e) => setSelectedTeacher(e.target.value || null)}
                  data-testid="filter-teacher"
                >
                  <option value="">Все преподаватели</option>
                  {teachers.map((teacher) => (
                    <option key={teacher.id} value={teacher.id}>
                      {teacher.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {lessonsLoading ? (
              <div className="lessons-loading">
                <Spinner />
              </div>
            ) : !Array.isArray(filteredLessons) ||
              filteredLessons.length === 0 ? (
              <div className="empty-state">
                <p>Занятия не найдены</p>
              </div>
            ) : (
              <div className="lessons-grid">
                {filteredLessons
                  .filter((lesson) => !lesson.isFiltered)
                  .map((lesson) => {
                    // Graceful degradation: if bookings failed, assume not booked
                    // Также проверяем что myBookingsData - массив
                    const isBooked =
                      bookingsError || !Array.isArray(myBookingsData)
                        ? false
                        : myBookingsData.some(
                            (b) =>
                              b?.lesson_id === lesson?.id &&
                              b?.status === "active",
                          );

                    return (
                      <LessonCard
                        key={lesson?.id || Math.random()}
                        lesson={lesson}
                        onClick={handleLessonClick}
                        isBooked={isBooked}
                        showActions={false}
                        isDisabled={lesson.isFiltered}
                      />
                    );
                  })}
              </div>
            )}
          </div>
        )}
      </main>

      {/* Booking Modal - shared by both views */}
      {selectedLesson && (
        <LessonBookingModal
          isOpen={showBookingModal}
          onClose={handleCloseModal}
          lesson={selectedLesson}
          myBookings={
            bookingsError || !Array.isArray(myBookingsData)
              ? []
              : myBookingsData
          }
          credits={credits ?? { balance: 0 }}
        />
      )}
    </div>
  );
};

export default StudentSchedule;
