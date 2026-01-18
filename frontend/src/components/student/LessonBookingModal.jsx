import { useState, useMemo, useRef, useEffect } from "react";
import { useBookLesson } from "../../hooks/useBookLesson.js";
import { useMyBookings } from "../../hooks/useMyBookings.js";
import { useCancelledLessons } from "../../hooks/useCancelledLessons.js";
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from "../common/Spinner.jsx";
import HomeworkSection from "../admin/HomeworkSection.jsx";
import { hexToRgba } from "../../utils/colorUtils.js";
import { formatTime } from "../../utils/dateFormat.js";
import "./LessonBookingModal.css";

/**
 * Format date and time for display in lesson details card
 */
const formatDateTimeRange = (startString, endString) => {
  const startDate = new Date(startString);
  const endDate = new Date(endString);
  
  const dateStr = startDate.toLocaleDateString("ru-RU", {
    month: "long",
    day: "numeric",
  });
  
  const startTime = startDate.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
  
  const endTime = endDate.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
  
  return `${dateStr}, ${startTime}-${endTime}`;
};

/**
 * Format credit count with proper Russian plural form
 * 1 кредит, 2-4 кредита, 5+ кредитов
 */
const formatCredits = (count) => {
  const lastDigit = count % 10;
  const lastTwoDigits = count % 100;

  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
    return `${count} кредитов`;
  }
  if (lastDigit === 1) {
    return `${count} кредит`;
  }
  if (lastDigit >= 2 && lastDigit <= 4) {
    return `${count} кредита`;
  }
  return `${count} кредитов`;
};

/**
 * Check if booking can be cancelled
 */
const canCancelBooking = (lesson, myBookings) => {
  const booking = myBookings.find(
    (b) => b.lesson_id === lesson.id && b.status === "active",
  );
  if (!booking) return { canCancel: false, reason: null, bookingId: null };

  const now = new Date();
  const lessonTime = new Date(lesson.start_time);
  const hoursUntilLesson = (lessonTime - now) / (1000 * 60 * 60);

  // Check if less than 24 hours
  if (hoursUntilLesson < 24) {
    return {
      canCancel: false,
      reason: "Отмена возможна только за 24 часа до начала занятия",
      bookingId: booking.id,
    };
  }

  return { canCancel: true, reason: null, bookingId: booking.id };
};

/**
 * Check if lesson can be booked and return reason
 */
const checkBookingEligibility = (lesson, myBookings, credits, isCancelled) => {
  const now = new Date();
  const lessonTime = new Date(lesson.start_time);
  const hoursUntilLesson = (lessonTime - now) / (1000 * 60 * 60);

  // Check if already booked
  const isBooked = myBookings.some(
    (b) => b.lesson_id === lesson.id && b.status === "active",
  );
  if (isBooked) {
    return { canBook: false, reason: "Вы уже записаны на это занятие" };
  }

  // Check if previously cancelled (re-booking blocked)
  if (isCancelled) {
    return {
      canBook: false,
      reason:
        "Вы отписались от этого занятия и больше не можете на него записаться",
      isPreviouslyCancelled: true,
    };
  }

  // Check if lesson is in past
  if (lessonTime <= now) {
    return { canBook: false, reason: "Это занятие уже началось или прошло" };
  }

  // Check if full
  if (lesson.current_students >= lesson.max_students) {
    return { canBook: false, reason: "Нет свободных мест" };
  }

  // Check credits - используем ?? для корректной обработки 0
  const creditsCost = lesson.credits_cost || 1;
  const creditBalance = credits?.balance ?? 0;
  if (creditBalance < creditsCost) {
    return {
      canBook: false,
      reason: `Недостаточно кредитов (требуется ${formatCredits(creditsCost)})`,
    };
  }

  return { canBook: true, reason: null };
};

/**
 * LessonBookingModal - Modal for viewing lesson details, booking, and cancellation
 */
export const LessonBookingModal = ({
  isOpen,
  onClose,
  lesson,
  myBookings,
  credits,
}) => {
  const [isBooking, setIsBooking] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [bookingSuccess, setBookingSuccess] = useState(false);
  const [cancellationSuccess, setCancellationSuccess] = useState(false);

  const bookLessonMutation = useBookLesson();
  const { cancelBooking } = useMyBookings();
  const { isLessonCancelled } = useCancelledLessons();

  // Refs для хранения IDs таймеров
  const bookingTimerRef = useRef(null);
  const cancellationTimerRef = useRef(null);

  // Check if lesson is already booked
  const isBooked = useMemo(
    () =>
      myBookings.some(
        (b) => b.lesson_id === lesson.id && b.status === "active",
      ),
    [lesson.id, myBookings],
  );

  // Check if lesson was previously cancelled by this student
  const isPreviouslyCancelled = useMemo(
    () => isLessonCancelled(lesson.id),
    [lesson.id, isLessonCancelled],
  );

  const eligibility = useMemo(
    () =>
      checkBookingEligibility(
        lesson,
        myBookings,
        credits,
        isPreviouslyCancelled,
      ),
    [lesson, myBookings, credits, isPreviouslyCancelled],
  );

  const cancellationEligibility = useMemo(
    () => canCancelBooking(lesson, myBookings),
    [lesson, myBookings],
  );

  const handleBook = async () => {
    if (!eligibility.canBook) return;

    setIsBooking(true);
    try {
      await bookLessonMutation.mutateAsync(lesson.id);
      setBookingSuccess(true);

      // Close modal after short delay
      bookingTimerRef.current = setTimeout(() => {
        onClose();
        setBookingSuccess(false);
        bookingTimerRef.current = null;
      }, 2000);
    } catch (error) {
      // Error notification handled by mutation
    } finally {
      setIsBooking(false);
    }
  };

  const handleCancel = async () => {
    if (
      !cancellationEligibility.canCancel ||
      !cancellationEligibility.bookingId
    )
      return;

    setIsCancelling(true);
    try {
      await cancelBooking(cancellationEligibility.bookingId);
      setCancellationSuccess(true);

      // Close modal after short delay
      cancellationTimerRef.current = setTimeout(() => {
        onClose();
        setCancellationSuccess(false);
        cancellationTimerRef.current = null;
      }, 2000);
    } catch (error) {
      // Error notification handled by mutation
    } finally {
      setIsCancelling(false);
    }
  };

  // Cleanup таймеров при unmount
  useEffect(() => {
    return () => {
      if (bookingTimerRef.current) {
        clearTimeout(bookingTimerRef.current);
      }
      if (cancellationTimerRef.current) {
        clearTimeout(cancellationTimerRef.current);
      }
    };
  }, []);

  const timeRange = formatDateTimeRange(lesson.start_time, lesson.end_time);
  const lessonColor = lesson.color || '#004231';
  const backgroundColor = hexToRgba(lessonColor, 0.15);
  
  // Вычисляем количество свободных мест - идентично календарю
  // Используем ту же логику, что и в Calendar.jsx (строка 362)
  // В календаре: const spotsLeft = lesson.max_students - (lesson.current_students || 0);
  // Защита от отсутствующих данных - используем те же значения, что и в календаре
  const maxStudents = lesson?.max_students ?? 0;
  const currentStudents = lesson?.current_students ?? 0;
  const spotsLeft = maxStudents - currentStudents;

  // Стили для lesson-details, идентичные calendar-lesson-owned
  const lessonDetailsStyle = {
    borderLeftColor: lessonColor,
    backgroundColor: backgroundColor,
    color: '#000000', // Всегда чёрный текст, как в calendar-lesson
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      closeOnBackdropClick={true}
      title="Детали занятия"
    >
      <div className="lesson-booking-modal" data-testid="lesson-booking-modal">
        {bookingSuccess ? (
          <div className="booking-success">
            <div className="success-icon">✓</div>
            <h3>Запись успешна!</h3>
            <p>Вы успешно записались на это занятие</p>
          </div>
        ) : cancellationSuccess ? (
          <div className="booking-success">
            <div className="success-icon">✓</div>
            <h3>Отмена успешна!</h3>
            <p>Ваше бронирование отменено, кредит возвращен</p>
          </div>
        ) : (
          <>
            <div 
              className={`lesson-details ${isBooked ? 'lesson-details-owned' : ''}`}
              style={lessonDetailsStyle}
            >
              {/* Структура идентична calendar-lesson */}
              {lesson.subject && (
                <div className="calendar-lesson-subject" title={lesson.subject}>
                  {lesson.subject}
                </div>
              )}
              <div className="calendar-lesson-time">
                {formatTime(lesson.start_time)}{lesson.end_time ? `-${formatTime(lesson.end_time)}` : ''}
              </div>
              <div className="calendar-lesson-teacher">
                {lesson.teacher_name || "Неизвестно"}
              </div>
              <div className="calendar-lesson-spots">
                Свободно {spotsLeft} из {maxStudents} мест
              </div>
              {lesson.credits_cost && lesson.credits_cost > 0 && (
                <div className="calendar-lesson-credits">
                  {lesson.credits_cost} {lesson.credits_cost === 1 ? 'кредит' : lesson.credits_cost < 5 ? 'кредита' : 'кредитов'}
                </div>
              )}
            </div>

            {/* Секция домашних заданий - показываем только для записанных студентов */}
            {isBooked && (
              <div className="homework-container">
                <h3 className="homework-title">Домашние задания</h3>
                <HomeworkSection lessonId={lesson.id} lesson={lesson} />
              </div>
            )}

            {/* Show booking UI if not booked */}
            {!isBooked && (
              <>
                {/* Show special blocked state if previously cancelled */}
                {eligibility.isPreviouslyCancelled ? (
                  <div className="previously-cancelled-section">
                    <div className="cancelled-badge">
                      <span className="badge-icon">🚫</span>
                      <span className="badge-text">Отписались</span>
                    </div>
                    <p className="cancelled-message">
                      Вы отписались от этого занятия и больше не можете на него
                      записаться
                    </p>
                  </div>
                ) : (
                  <>
                    <div className="credit-info">
                      <div className="credit-check">
                        <span className="label">Ваши кредиты:</span>
                        <span className="value">{credits?.balance ?? 0}</span>
                      </div>
                      <div className="credit-cost">
                        <span className="label">Стоимость занятия:</span>
                        <span className="value">{formatCredits(lesson.credits_cost || 1)}</span>
                      </div>
                      {eligibility.canBook && (
                        <div className="credit-remaining">
                          <span className="label">Останется после записи:</span>
                          <span className="value">
                            {formatCredits((credits?.balance ?? 0) - (lesson.credits_cost || 1))}
                          </span>
                        </div>
                      )}
                    </div>

                    {!eligibility.canBook && (
                      <div className="booking-error" role="alert">
                        <p>{eligibility.reason}</p>
                      </div>
                    )}
                  </>
                )}
              </>
            )}

            {/* Show cancellation UI if booked */}
            {isBooked && (
              <>
                {!cancellationEligibility.canCancel &&
                  cancellationEligibility.reason && (
                    <div className="booking-error" role="alert">
                      <p>{cancellationEligibility.reason}</p>
                    </div>
                  )}
              </>
            )}

            <div className="modal-actions">
              <Button
                onClick={onClose}
                variant="secondary"
                disabled={isBooking || isCancelling}
              >
                Закрыть
              </Button>

              {!isBooked ? (
                <Button
                  onClick={handleBook}
                  variant="primary"
                  disabled={!eligibility.canBook || isBooking}
                  data-testid="book-lesson-button"
                >
                  {isBooking ? (
                    <>
                      <Spinner size="small" />
                      <span style={{ marginLeft: "8px" }}>Запись...</span>
                    </>
                  ) : (
                    "Записаться"
                  )}
                </Button>
              ) : (
                <Button
                  onClick={handleCancel}
                  variant="danger"
                  disabled={!cancellationEligibility.canCancel || isCancelling}
                  data-testid="cancel-booking-button"
                >
                  {isCancelling ? (
                    <>
                      <Spinner size="small" />
                      <span style={{ marginLeft: "8px" }}>Отмена...</span>
                    </>
                  ) : (
                    "Отменить запись"
                  )}
                </Button>
              )}
            </div>
          </>
        )}
      </div>
    </Modal>
  );
};

export default LessonBookingModal;
