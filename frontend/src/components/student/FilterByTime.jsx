import './Filters.css';

const TIME_SLOTS = [
  { value: 'morning', label: 'Утро (6:00 - 12:00)' },
  { value: 'afternoon', label: 'День (12:00 - 18:00)' },
  { value: 'evening', label: 'Вечер (18:00 - 00:00)' },
];

/**
 * FilterByTime - Dropdown filter for selecting time of day
 */
export const FilterByTime = ({ selected, onSelect }) => {
  const handleChange = (e) => {
    const value = e.target.value;
    onSelect(value === '' ? null : value);
  };

  return (
    <div className="filter-group" data-testid="filter-by-time">
      <label className="filter-label" htmlFor="time-filter">
        Время
      </label>
      <select
        id="time-filter"
        className="filter-select"
        value={selected || ''}
        onChange={handleChange}
      >
        <option value="">Всё время</option>
        {TIME_SLOTS.map(slot => (
          <option key={slot.value} value={slot.value}>
            {slot.label}
          </option>
        ))}
      </select>
    </div>
  );
};

export default FilterByTime;
