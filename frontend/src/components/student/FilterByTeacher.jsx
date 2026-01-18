import { useMemo } from 'react';
import './Filters.css';

/**
 * FilterByTeacher - Dropdown filter for selecting teacher
 */
export const FilterByTeacher = ({ lessons, selected, onSelect }) => {
  // Extract unique teachers from lessons
  const teachers = useMemo(() => {
    if (!lessons || !Array.isArray(lessons)) return [];

    const teacherMap = new Map();

    lessons.forEach(lesson => {
      if (lesson.teacher_id && lesson.teacher_name) {
        teacherMap.set(lesson.teacher_id, lesson.teacher_name);
      }
    });

    return Array.from(teacherMap.entries()).map(([id, name]) => ({ id, name }));
  }, [lessons]);

  const handleChange = (e) => {
    const value = e.target.value;
    onSelect(value === '' ? null : value);
  };

  return (
    <div className="filter-group" data-testid="filter-by-teacher">
      <label className="filter-label" htmlFor="teacher-filter">
        Преподаватель
      </label>
      <select
        id="teacher-filter"
        className="filter-select"
        value={selected || ''}
        onChange={handleChange}
      >
        <option value="">Все преподаватели</option>
        {teachers.map(teacher => (
          <option key={teacher.id} value={teacher.id}>
            {teacher.name}
          </option>
        ))}
      </select>
    </div>
  );
};

export default FilterByTeacher;
