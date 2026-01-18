import { useMemo } from 'react';
import './Filters.css';

/**
 * ScheduleFiltersBar - Combined filter component for student schedule
 *
 * Props:
 * @param {boolean} showIndividual - Show individual lessons
 * @param {boolean} showGroup - Show group lessons
 * @param {boolean} hideFull - Hide full lessons
 * @param {boolean} myBookings - Show only my bookings
 * @param {string|number} selectedTeacher - Selected teacher ID
 * @param {Array} teachers - Array of teacher objects {id, name}
 * @param {Function} onFilterChange - Callback when filter changes (filterKey, value)
 * @param {string} viewMode - Current view mode: 'calendar' or 'list'
 * @param {Function} onViewModeChange - Callback when view mode changes
 */
export const ScheduleFiltersBar = ({
  showIndividual = true,
  showGroup = true,
  hideFull = false,
  myBookings = false,
  selectedTeacher = null,
  teachers = [],
  onFilterChange = () => {},
  viewMode = 'calendar',
  onViewModeChange = () => {}
}) => {
  // Ensure teachers is always an array
  const teacherList = useMemo(() => {
    if (!Array.isArray(teachers)) return [];
    return teachers;
  }, [teachers]);

  const handleCheckboxChange = (filterKey) => (e) => {
    const value = e.target.checked;
    onFilterChange(filterKey, value);
  };

  const handleTeacherChange = (e) => {
    const value = e.target.value;
    onFilterChange('selectedTeacher', value === '' ? null : value);
  };

  return (
    <div className="schedule-filters-bar" data-testid="schedule-filters-bar">
      {/* Checkbox Filters - Left Side */}
      <div className="filter-checkboxes-group">
        {/* Show Individual Lessons */}
        <label className="filter-checkbox-label" title="Показать индивидуальные занятия">
          <input
            type="checkbox"
            className="filter-checkbox"
            checked={showIndividual}
            onChange={handleCheckboxChange('showIndividual')}
            data-testid="filter-show-individual"
          />
          <span className="filter-checkbox-text">Индивид.</span>
        </label>

        {/* Show Group Lessons */}
        <label className="filter-checkbox-label" title="Показать групповые занятия">
          <input
            type="checkbox"
            className="filter-checkbox"
            checked={showGroup}
            onChange={handleCheckboxChange('showGroup')}
            data-testid="filter-show-group"
          />
          <span className="filter-checkbox-text">Группы</span>
        </label>

        {/* Hide Full Lessons */}
        <label className="filter-checkbox-label" title="Скрыть полные занятия">
          <input
            type="checkbox"
            className="filter-checkbox"
            checked={hideFull}
            onChange={handleCheckboxChange('hideFull')}
            data-testid="filter-hide-full"
          />
          <span className="filter-checkbox-text">Скрыть полные</span>
        </label>

      </div>

      {/* Teacher Dropdown */}
      <div className="filter-group filter-teacher-group">
        <select
          id="teacher-filter-bar"
          className="filter-select"
          value={selectedTeacher || ''}
          onChange={handleTeacherChange}
          data-testid="filter-teacher-select"
        >
          <option value="">Все преподаватели</option>
          {teacherList.map(teacher => (
            <option key={teacher.id} value={teacher.id}>
              {teacher.name}
            </option>
          ))}
        </select>
      </div>

      {/* View Mode Toggle (Calendar/List) */}
      <div className="filter-view-toggle">
        <button
          className={`filter-view-toggle-btn ${viewMode === 'calendar' ? 'active' : ''}`}
          onClick={() => onViewModeChange('calendar')}
          data-testid="filter-calendar-view-btn"
          title="Календарь"
        >
          📅 Календарь
        </button>
        <button
          className={`filter-view-toggle-btn ${viewMode === 'list' ? 'active' : ''}`}
          onClick={() => onViewModeChange('list')}
          data-testid="filter-list-view-btn"
          title="Список"
        >
          📋 Список
        </button>
      </div>
    </div>
  );
};

export default ScheduleFiltersBar;
