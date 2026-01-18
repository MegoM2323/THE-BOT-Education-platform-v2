import './Filters.css';

const DAYS = [
  { value: 1, label: 'Пн' },
  { value: 2, label: 'Вт' },
  { value: 3, label: 'Ср' },
  { value: 4, label: 'Чт' },
  { value: 5, label: 'Пт' },
  { value: 6, label: 'Сб' },
  { value: 0, label: 'Вс' },
];

/**
 * FilterByDay - Dropdown filter for selecting day of week
 */
export const FilterByDay = ({ selected, onSelect }) => {
  const handleChange = (e) => {
    const value = e.target.value;
    onSelect(value === '' ? null : parseInt(value, 10));
  };

  return (
    <div className="filter-group" data-testid="filter-by-day">
      <label className="filter-label" htmlFor="day-filter">
        День
      </label>
      <select
        id="day-filter"
        className="filter-select"
        value={selected !== null ? selected : ''}
        onChange={handleChange}
      >
        <option value="">Все дни</option>
        {DAYS.map(day => (
          <option key={day.value} value={day.value}>
            {day.label}
          </option>
        ))}
      </select>
    </div>
  );
};

export default FilterByDay;
