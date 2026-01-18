import { useMemo } from 'react';
import './Filters.css';

/**
 * FilterByAvailability - Checkbox filter for showing only lessons with available spots
 */
export const FilterByAvailability = ({ lessons, selected, onSelect }) => {
  // Calculate count of lessons with available spots
  const availableCount = useMemo(() => {
    if (!lessons || !Array.isArray(lessons)) return 0;

    return lessons.filter(lesson =>
      lesson.current_students < lesson.max_students
    ).length;
  }, [lessons]);

  const handleChange = (e) => {
    onSelect(e.target.checked);
  };

  return (
    <div className="filter-group filter-checkbox-group" data-testid="filter-by-availability">
      <label className="filter-checkbox-label" htmlFor="availability-filter">
        <input
          id="availability-filter"
          type="checkbox"
          className="filter-checkbox"
          checked={selected}
          onChange={handleChange}
        />
        <span className="filter-checkbox-text">
          Есть места
          {availableCount > 0 && (
            <span className="filter-count"> ({availableCount})</span>
          )}
        </span>
      </label>
    </div>
  );
};

export default FilterByAvailability;
