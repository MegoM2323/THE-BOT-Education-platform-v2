import { useState, useEffect } from "react";
import SharedCalendar from "../common/Calendar.jsx";
import Button from "../common/Button.jsx";
import LessonEditModal from "./LessonEditModal.jsx";
import LessonCreateModal from "./LessonCreateModal.jsx";
import AdminCalendarCreditsDisplay from "./AdminCalendarCreditsDisplay.jsx";
import StudentFilterSearch from "./StudentFilterSearch.jsx";
import { SkeletonCalendar } from "../common/SkeletonLoader.jsx";
import SlowConnectionNotice from "../common/SlowConnectionNotice.jsx";
import { useLessons } from "../../hooks/useLessons.js";
import { useSlowConnection } from "../../hooks/useSlowConnection.js";
import { getTeachersAll } from "../../api/users.js";
import "./Calendar.css";

export const Calendar = () => {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [view, setView] = useState("week"); // 'week' or 'month'
  const [editingLesson, setEditingLesson] = useState(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedDateForCreate, setSelectedDateForCreate] = useState(null);
  const [dismissedSlowNotice, setDismissedSlowNotice] = useState(false);
  const [selectedStudent, setSelectedStudent] = useState(null); // For credits display in calendar nav

  // Фильтры календаря
  const [hideFull, setHideFull] = useState(false);
  const [selectedTeacher, setSelectedTeacher] = useState("");
  const [selectedFilterStudents, setSelectedFilterStudents] = useState([]);
  const [showIndividual, setShowIndividual] = useState(true);
  const [showGroup, setShowGroup] = useState(true);
  const [teachers, setTeachers] = useState([]);

  // Build filters object from filter states
  const filters = {};

  // Fetch lessons with filters (refetch when student filter changes)
  const { lessons, loading, error, fetchLessons } = useLessons(filters);

  // Detect slow connection
  const isSlow = useSlowConnection(loading, 3000);

  // Fetch lessons and teachers on mount only
  useEffect(() => {
    fetchLessons();
    loadTeachers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Загрузить список преподавателей
  const loadTeachers = async () => {
    try {
      const teachersList = await getTeachersAll();
      setTeachers(teachersList);
    } catch (err) {
      console.error("Failed to load teachers:", err);
    }
  };

  const handleLessonClick = (lesson) => {
    // Skip intermediate modal - directly open edit modal
    setEditingLesson(lesson);
  };

  const handleLessonRightClick = (lesson) => {
    setEditingLesson(lesson);
  };

  const handleWeekChange = (newDate) => {
    setCurrentDate(newDate);
  };

  const handleStudentSelected = (studentId) => {
    // Called from LessonEditModal when a student is selected
    setSelectedStudent(studentId);
  };

  // Mark lessons as filtered based on filter states
  const lessonsWithFilterState = lessons.map((lesson) => {
    let isFiltered = false;

    // Фильтрация по типу занятия (индивидуальное/групповое)
    if (showIndividual === false && lesson.max_students === 1) {
      isFiltered = true;
    }
    if (showGroup === false && lesson.max_students > 1) {
      isFiltered = true;
    }

    // Фильтрация по заполненности
    if (hideFull && lesson.current_students >= lesson.max_students) {
      isFiltered = true;
    }

    // Фильтрация по преподавателю (сравнение через String для избежания проблем int vs string)
    if (
      selectedTeacher &&
      String(lesson.teacher_id) !== String(selectedTeacher)
    ) {
      isFiltered = true;
    }

    // Фильтрация по студентам - работает как фильтр преподавателя
    if (selectedFilterStudents?.length > 0) {
      const hasBooking = lesson.bookings?.some(booking =>
        selectedFilterStudents.includes(String(booking.student_id))
      );
      if (!hasBooking) isFiltered = true;
    }

    return { ...lesson, isFiltered };
  });

  // Обработчик открытия модала создания занятия
  const handleOpenCreateModal = () => {
    setSelectedDateForCreate(currentDate);
    setShowCreateModal(true);
  };

  // Обработчик закрытия модала создания занятия
  const handleCloseCreateModal = () => {
    setShowCreateModal(false);
    setSelectedDateForCreate(null);
  };

  // Обработчик успешного создания занятия
  const handleLessonCreated = () => {
    fetchLessons();
    handleCloseCreateModal();
  };

  const createButton = (
    <Button onClick={handleOpenCreateModal} data-testid="create-lesson-btn">
      + Создать занятие
    </Button>
  );

  // Фильтры для передачи в SharedCalendar
  const calendarFilters = (
    <div className="calendar-filters-inline">
      <label className="filter-checkbox">
        <input
          type="checkbox"
          checked={showIndividual}
          onChange={(e) => setShowIndividual(e.target.checked)}
          data-testid="filter-show-individual"
        />
        <span>Показать индивидуальные</span>
      </label>

      <label className="filter-checkbox">
        <input
          type="checkbox"
          checked={showGroup}
          onChange={(e) => setShowGroup(e.target.checked)}
          data-testid="filter-show-group"
        />
        <span>Показать групповые</span>
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

      <select
        className="filter-select"
        value={selectedTeacher}
        onChange={(e) => setSelectedTeacher(e.target.value)}
        data-testid="filter-teacher"
      >
        <option value="">Все преподаватели</option>
        {teachers.map((teacher) => (
          <option key={teacher.id} value={teacher.id}>
            {teacher.full_name || teacher.email}
          </option>
        ))}
      </select>

      <StudentFilterSearch
        selectedStudents={selectedFilterStudents}
        onStudentsSelect={setSelectedFilterStudents}
        data-testid="filter-student"
      />
    </div>
  );

  // Верхняя строка: кнопки переключения вида (слева) + фильтры/выпадающий список/кнопка создания (справа)
  const headerActionsTop = (
    <>
      <div className="calendar-view-toggle">
        <Button
          size="md"
          variant={view === "week" ? "primary" : "secondary"}
          className={view === "week" ? "btn-active" : ""}
          onClick={() => setView("week")}
          data-testid="view-week-btn"
        >
          Неделя
        </Button>
        <Button
          size="md"
          variant={view === "month" ? "primary" : "secondary"}
          className={view === "month" ? "btn-active" : ""}
          onClick={() => setView("month")}
          data-testid="view-month-btn"
        >
          Месяц
        </Button>
      </div>
      <div className="calendar-header-top-actions">
        {calendarFilters}
        {createButton}
      </div>
    </>
  );

  // Show skeleton while initial loading
  if (loading && lessons.length === 0) {
    return (
      <div data-testid="admin-calendar">
        <div className="calendar-title-section">
          <h2 className="calendar-page-title">Календарь занятий</h2>
        </div>

        <div aria-busy="true" aria-live="polite">
          <SkeletonCalendar />
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
    <div data-testid="admin-calendar">
      <div className="calendar-title-section">
        <h2 className="calendar-page-title">Календарь занятий</h2>
      </div>

      {/* Calendar is always shown - handles empty state internally */}
      <SharedCalendar
        lessons={lessonsWithFilterState}
        role="admin"
        onLessonClick={handleLessonClick}
        onLessonRightClick={handleLessonRightClick}
        currentWeek={currentDate}
        onWeekChange={handleWeekChange}
        isLoading={loading}
        error={error ? error?.message || String(error) : null}
        view={view}
        onViewChange={setView}
        headerActionsTop={headerActionsTop}
        navActions={
          <AdminCalendarCreditsDisplay selectedStudentId={selectedStudent} />
        }
      />

      {editingLesson !== null && (
        <LessonEditModal
          isOpen={editingLesson !== null}
          onClose={() => {
            setEditingLesson(null);
            setSelectedStudent(null); // Clear selected student when closing modal
          }}
          lesson={editingLesson}
          onLessonUpdated={(updatedLessonData) => {
            // Handle result of update from modal
            setEditingLesson(null); // Close modal after any action
            setSelectedStudent(null); // Clear selected student when closing modal

            // If no data, this is probably deletion - reload all
            if (!updatedLessonData || !updatedLessonData.id) {
              fetchLessons();
            }
            // If there's data - useLessons hook already updated local state via updateLesson or deleteLesson
            // No need for additional reload - improves performance
          }}
          onStudentSelected={handleStudentSelected}
        />
      )}

      {/* Модал создания нового занятия */}
      <LessonCreateModal
        isOpen={showCreateModal}
        onClose={handleCloseCreateModal}
        selectedDate={selectedDateForCreate}
        onLessonCreated={handleLessonCreated}
      />
    </div>
  );
};

export default Calendar;
