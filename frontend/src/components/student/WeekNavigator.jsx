import Button from "../common/Button.jsx";
import './WeekNavigator.css';

/**
 * Format week range for display
 */
const formatWeekRange = (monday) => {
  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);

  const startStr = monday.toLocaleDateString('ru-RU', {
    month: 'short',
    day: 'numeric',
  });

  const endStr = sunday.toLocaleDateString('ru-RU', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  return `${startStr} - ${endStr}`;
};

/**
 * Get Monday of a given date
 */
const getMonday = (date = new Date()) => {
  const d = new Date(date);
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  const monday = new Date(d.setDate(diff));
  monday.setHours(0, 0, 0, 0);
  return monday;
};

/**
 * WeekNavigator - Navigate between weeks
 */
export const WeekNavigator = ({ currentWeek, onWeekChange }) => {
  const handlePrevWeek = () => {
    const prevWeek = new Date(currentWeek);
    prevWeek.setDate(prevWeek.getDate() - 7);
    onWeekChange(prevWeek);
  };

  const handleNextWeek = () => {
    const nextWeek = new Date(currentWeek);
    nextWeek.setDate(nextWeek.getDate() + 7);
    onWeekChange(nextWeek);
  };

  const handleToday = () => {
    const today = getMonday(new Date());
    onWeekChange(today);
  };

  const isCurrentWeek = () => {
    const todayMonday = getMonday(new Date());
    return currentWeek.getTime() === todayMonday.getTime();
  };

  // Prevent going too far in past (optional - limit to 1 week ago)
  const canGoPrev = () => {
    const oneWeekAgo = new Date();
    oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
    const oneWeekAgoMonday = getMonday(oneWeekAgo);
    return currentWeek >= oneWeekAgoMonday;
  };

  // Prevent going too far in future (optional - limit to 8 weeks ahead)
  const canGoNext = () => {
    const eightWeeksAhead = new Date();
    eightWeeksAhead.setDate(eightWeeksAhead.getDate() + 56); // 8 weeks
    const eightWeeksAheadMonday = getMonday(eightWeeksAhead);
    return currentWeek < eightWeeksAheadMonday;
  };

  return (
    <div className="week-navigator" data-testid="week-navigator">
      <Button
        onClick={handlePrevWeek}
        variant="secondary"
        size="small"
        disabled={!canGoPrev()}
        data-testid="prev-week"
      >
        ← Назад
      </Button>

      <div className="week-range">
        <span className="week-text">{formatWeekRange(currentWeek)}</span>
      </div>

      <Button
        onClick={handleToday}
        variant="secondary"
        size="small"
        disabled={isCurrentWeek()}
        data-testid="today-button"
      >
        Сегодня
      </Button>

      <Button
        onClick={handleNextWeek}
        variant="secondary"
        size="small"
        disabled={!canGoNext()}
        data-testid="next-week"
      >
        Вперёд →
      </Button>
    </div>
  );
};

export default WeekNavigator;
