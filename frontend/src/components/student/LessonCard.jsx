import Card from '../common/Card.jsx';
import Button from "../common/Button.jsx";
import { formatDateShort, formatTime } from '../../utils/dateFormat.js';
import './LessonCard.css';

export const LessonCard = ({
  lesson,
  onBook,
  onClick,
  isBooked = false,
  bookingId,
  onCancelClick,
  isRemoving = false,
  isOperating = false,
  isEntering = false,
  isCancelling = false,
  showActions = true,
  isDisabled = false,
}) => {
  if (!lesson) return null;

  const spotsLeft = (lesson.max_students ?? 0) - (lesson.current_students ?? 0);
  const isFull = spotsLeft <= 0;

  const lessonColor = lesson.color || '#004231';
  const cardStyle = {
    borderLeft: `4px solid ${lessonColor}`,
    background: isBooked ? `${lessonColor}15` : `${lessonColor}08`,
    ...(isBooked && {
      borderTop: `2px solid ${lessonColor}`,
      borderRight: `2px solid ${lessonColor}`,
      borderBottom: `2px solid ${lessonColor}`,
    }),
  };

  const canCancelBooking = () => {
    if (!lesson?.start_time) return false;
    const lessonStart = new Date(lesson.start_time);
    const now = new Date();
    const hoursUntilLesson = (lessonStart - now) / (1000 * 60 * 60);
    return hoursUntilLesson > 24;
  };

  const handleCancelClick = async () => {
    if (!canCancelBooking() || !onCancelClick || !bookingId) {
      return;
    }
    await onCancelClick(bookingId, lesson);
  };

  const handleBookClick = (e) => {
    e.stopPropagation();
    if (!isDisabled) {
      onBook(lesson.id);
    }
  };

  const handleCardClick = () => {
    if (onClick && !isDisabled) {
      onClick(lesson);
    }
  };

  const isClickable = Boolean(onClick) && !isDisabled;

  const formatDateTime = () => {
    if (!lesson?.start_time) return '';
    const date = formatDateShort(lesson.start_time);
    const time = lesson.end_time
      ? `${formatTime(lesson.start_time)}-${formatTime(lesson.end_time)}`
      : formatTime(lesson.start_time);
    return `${date}, ${time}`;
  };

  return (
    <Card
      className={`lesson-card lesson-card-compact ${isBooked ? 'lesson-card-booked' : ''} ${isRemoving ? 'lesson-card-removing' : ''} ${isOperating ? 'lesson-card-operating' : ''} ${isEntering ? 'lesson-card-entering' : ''} ${isDisabled ? 'lesson-card-filtered' : ''} ${isClickable ? 'lesson-card-clickable' : ''}`}
      style={cardStyle}
      onClick={handleCardClick}
      data-testid="lesson-card"
    >
      <div className="lesson-header-compact">
        <div className="lesson-main-info">
          <span className="lesson-subject-compact" title={lesson?.subject || ''}>
            {lesson?.subject || ''}
          </span>
          {lesson?.teacher_name && (
            <span className="lesson-teacher-compact">Учитель: {lesson.teacher_name}</span>
          )}
          <span className="lesson-datetime-compact">{formatDateTime()}</span>
        </div>
        <div className="lesson-right-section">
          {!isBooked && (
            <>
              <span className={`lesson-spots ${isFull ? 'full' : 'available'}`}>
                {isFull ? 'Нет мест' : `${spotsLeft} мест`}
              </span>
              {lesson?.credits_cost && lesson.credits_cost > 0 && (
                <span className="lesson-credits-cost">
                  {lesson.credits_cost} {lesson.credits_cost === 1 ? 'кредит' : lesson.credits_cost < 5 ? 'кредита' : 'кредитов'}
                </span>
              )}
            </>
          )}
          {showActions && lesson?.id && (
            <Button
              onClick={handleBookClick}
              disabled={isFull || isOperating || isDisabled}
              size="small"
              data-testid="book-button"
              title={isDisabled ? 'Урок скрыт фильтрами' : ''}
              className="lesson-book-btn"
            >
              {isOperating ? '...' : isBooked ? 'Открыть' : 'Записаться'}
            </Button>
          )}
        </div>
      </div>
    </Card>
  );
};

export default LessonCard;
