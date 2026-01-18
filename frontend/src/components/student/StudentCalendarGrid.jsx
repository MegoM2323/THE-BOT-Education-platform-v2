import { useState, useMemo, useEffect } from "react";
import LessonBookingModal from "./LessonBookingModal.jsx";
import "./StudentCalendarGrid.css";

/**
 * Get array of 7 days starting from Monday
 */
const getWeekDays = (monday) => {
  const days = [];
  for (let i = 0; i < 7; i++) {
    const day = new Date(monday);
    day.setDate(monday.getDate() + i);
    days.push(day);
  }
  return days;
};

/**
 * Format date for display
 */
const formatDateHeader = (date) => {
  const days = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
  const dayName = days[date.getDay()];
  const dayNum = date.getDate();
  const month = date.getMonth() + 1;
  return `${dayName} ${month}/${dayNum}`;
};

/**
 * Format time for display
 */
const formatTime = (date) => {
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  return `${hours}:${minutes}`;
};

/**
 * Check if lesson can be booked
 */
const canBookLesson = (lesson, myBookings) => {
  const now = new Date();
  const lessonTime = new Date(lesson.start_time);
  const hoursUntilLesson = (lessonTime - now) / (1000 * 60 * 60);

  // Check if lesson is in past
  if (lessonTime <= now) return { canBook: false, reason: "Past lesson" };

  // Check if already booked
  const isBooked = myBookings.some(
    (b) => b.lesson_id === lesson.id && b.status === "active",
  );
  if (isBooked) return { canBook: false, reason: "Already booked" };

  // Check if full
  if (lesson.current_students >= lesson.max_students) {
    return { canBook: false, reason: "Full" };
  }

  return { canBook: true, reason: null };
};

/**
 * StudentCalendarGrid - Weekly calendar view for students
 */
export const StudentCalendarGrid = ({
  lessons,
  myBookings,
  currentWeek,
  credits,
}) => {
  const [selectedLesson, setSelectedLesson] = useState(null);
  const [showModal, setShowModal] = useState(false);

  const weekDays = useMemo(() => getWeekDays(currentWeek), [currentWeek]);

  // Group lessons by day
  const lessonsByDay = useMemo(() => {
    const grouped = {};

    weekDays.forEach((day, index) => {
      grouped[index] = [];
    });

    lessons.forEach((lesson) => {
      const lessonDate = new Date(lesson.start_time);
      const dayIndex = weekDays.findIndex(
        (d) =>
          d.getDate() === lessonDate.getDate() &&
          d.getMonth() === lessonDate.getMonth() &&
          d.getFullYear() === lessonDate.getFullYear(),
      );

      if (dayIndex !== -1) {
        grouped[dayIndex].push(lesson);
      }
    });

    // Sort lessons by time within each day
    Object.keys(grouped).forEach((dayIndex) => {
      grouped[dayIndex].sort((a, b) => {
        return new Date(a.start_time) - new Date(b.start_time);
      });
    });

    return grouped;
  }, [lessons, weekDays]);

  const handleLessonClick = (lesson) => {
    setSelectedLesson(lesson);
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedLesson(null);
  };

  const [windowWidth, setWindowWidth] = useState(
    typeof window !== "undefined" ? window.innerWidth : 1024,
  );

  useEffect(() => {
    const handleResize = () => setWindowWidth(window.innerWidth);
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const getGridClassName = () => {
    if (windowWidth < 480) return "calendar-grid calendar-grid-2";
    if (windowWidth < 768) return "calendar-grid calendar-grid-3";
    return "calendar-grid calendar-grid-7";
  };

  return (
    <div className="student-calendar-grid" data-testid="student-calendar-grid">
      <div className={`calendar-header ${getGridClassName()}`}>
        {weekDays.map((day, index) => (
          <div key={index} className="day-header">
            {formatDateHeader(day)}
          </div>
        ))}
      </div>

      <div className={`calendar-body ${getGridClassName()}`}>
        {weekDays.map((day, dayIndex) => {
          const isToday = new Date().toDateString() === day.toDateString();
          return (
          <div
            key={dayIndex}
            className={`day-column ${isToday ? 'today' : ''}`}
            data-testid={`day-column-${dayIndex}`}
          >
            {lessonsByDay[dayIndex]?.length > 0 ? (
              lessonsByDay[dayIndex].map((lesson) => {
                const bookingStatus = canBookLesson(lesson, myBookings);
                const isBooked = myBookings.some(
                  (b) => b.lesson_id === lesson.id && b.status === "active",
                );
                const isPast = new Date(lesson.start_time) <= new Date();
                const isFull = lesson.current_students >= lesson.max_students;

                let cardClass = "lesson-card";
                if (isBooked) cardClass += " booked";
                else if (isFull) cardClass += " full";
                else if (isPast) cardClass += " past";
                else if (bookingStatus.canBook) cardClass += " available";

                const startTime = new Date(lesson.start_time);
                const endTime = new Date(lesson.end_time);

                // Цвет из урока или дефолтный
                const lessonColor = lesson.color || "#3b82f6";

                // Конвертация HEX в RGBA для полупрозрачного фона
                const hexToRgba = (hex, alpha = 1) => {
                  const result =
                    /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
                  if (!result) return `rgba(59, 130, 246, ${alpha})`;
                  const r = parseInt(result[1], 16);
                  const g = parseInt(result[2], 16);
                  const b = parseInt(result[3], 16);
                  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
                };

                return (
                  <div
                    key={lesson.id}
                    className={cardClass}
                    onClick={() => handleLessonClick(lesson)}
                    data-testid="lesson-card"
                    style={{
                      borderLeftColor: lessonColor,
                      backgroundColor: hexToRgba(lessonColor, 0.08),
                      ...(isBooked && {
                        border: `1px solid ${lessonColor}`,
                        borderLeftWidth: "4px",
                      }),
                    }}
                  >
                    <div className="lesson-time">
                      {formatTime(startTime)} - {formatTime(endTime)}
                    </div>
                    <div
                      className="lesson-subject"
                      title={lesson.subject || "Lesson"}
                    >
                      {lesson.subject || "Lesson"}
                    </div>
                    <div className="lesson-teacher">
                      {lesson.teacher_name || "Teacher"}
                    </div>
                    <div className="lesson-capacity">
                      {lesson.current_students || 0} /{" "}
                      {lesson.max_students || 0}
                    </div>
                    <div className="lesson-status">
                      {!isBooked && isFull && (
                        <span className="badge full-badge">FULL</span>
                      )}
                      {!isBooked && isPast && (
                        <span className="badge past-badge">PAST</span>
                      )}
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="empty-day">
                <p>No lessons</p>
              </div>
            )}
          </div>
        );
        })}
      </div>

      {selectedLesson && (
        <LessonBookingModal
          isOpen={showModal}
          onClose={handleCloseModal}
          lesson={selectedLesson}
          myBookings={myBookings}
          credits={credits ?? { balance: 0 }}
        />
      )}
    </div>
  );
};

export default StudentCalendarGrid;
