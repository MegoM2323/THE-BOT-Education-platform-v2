import { useState, useEffect, useMemo, useRef, useCallback } from 'react';
import { getStudentsAll } from '../../api/users.js';
import './StudentFilterSearch.css';

export const StudentFilterSearch = ({
  selectedStudents,
  onStudentsSelect,
  disabled = false,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const blurTimeoutRef = useRef(null);
  const abortControllerRef = useRef(null);

  useEffect(() => {
    abortControllerRef.current = new AbortController();
    const loadStudents = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await getStudentsAll({ signal: abortControllerRef.current.signal });
        setStudents(data || []);
      } catch (err) {
        if (err.name !== 'AbortError') {
          console.error('Failed to load students:', err);
          setError(err.message || 'Ошибка при загрузке студентов');
          setStudents([]);
        }
      } finally {
        setLoading(false);
      }
    };

    loadStudents();

    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  const filteredStudents = useMemo(
    () => students.filter((student) =>
      (student.full_name || '').toLowerCase().includes(searchTerm.toLowerCase())
    ),
    [students, searchTerm]
  );

  const selectedStudentNames = useMemo(
    () => {
      if (!selectedStudents || selectedStudents.length === 0) {
        return [];
      }
      return selectedStudents
        .map(id => students.find(s => String(s.id) === id))
        .filter(Boolean)
        .map(s => s.full_name || '');
    },
    [selectedStudents, students]
  );

  const handleStudentToggle = useCallback((student) => {
    const studentIdStr = String(student.id);
    if (selectedStudents?.includes(studentIdStr)) {
      onStudentsSelect((selectedStudents || []).filter(id => id !== studentIdStr));
    } else {
      onStudentsSelect([...(selectedStudents || []), studentIdStr]);
    }
  }, [selectedStudents, onStudentsSelect]);

  const handleChipRemove = useCallback((studentId) => {
    onStudentsSelect((selectedStudents || []).filter(id => id !== studentId));
  }, [selectedStudents, onStudentsSelect]);

  const handleClearAll = useCallback(() => {
    onStudentsSelect([]);
  }, [onStudentsSelect]);

  const handleInputChange = (e) => {
    setSearchTerm(e.target.value);
    setShowDropdown(true);
    setHighlightedIndex(-1);
  };

  const handleInputFocus = () => {
    setShowDropdown(true);
  };

  const handleInputBlur = useCallback(() => {
    blurTimeoutRef.current = setTimeout(() => setShowDropdown(false), 200);
  }, []);

  const handleKeyDown = useCallback((e) => {
    if (!showDropdown && e.key !== 'ArrowDown') {
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setShowDropdown(true);
        setHighlightedIndex((prevIndex) =>
          prevIndex < filteredStudents.length - 1 ? prevIndex + 1 : prevIndex
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightedIndex((prevIndex) => (prevIndex > 0 ? prevIndex - 1 : -1));
        break;
      case ' ':
        e.preventDefault();
        if (highlightedIndex >= 0 && filteredStudents[highlightedIndex]) {
          handleStudentToggle(filteredStudents[highlightedIndex]);
        }
        break;
      case 'Escape':
        e.preventDefault();
        setShowDropdown(false);
        setHighlightedIndex(-1);
        break;
      default:
        break;
    }
  }, [showDropdown, highlightedIndex, filteredStudents, handleStudentToggle]);

  useEffect(() => {
    return () => {
      if (blurTimeoutRef.current) {
        clearTimeout(blurTimeoutRef.current);
      }
    };
  }, []);

  return (
    <div className="student-filter-search-container">
      {selectedStudents && selectedStudents.length > 0 && (
        <div className="student-filter-search-chips-container">
          {selectedStudentNames.map((name, index) => (
            <div key={selectedStudents[index]} className="student-filter-search-chip">
              <span>{name}</span>
              <button
                type="button"
                className="student-filter-search-chip-remove"
                onClick={() => handleChipRemove(selectedStudents[index])}
                disabled={disabled}
                aria-label={`Удалить ${name}`}
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      )}

      <div className="student-filter-search-input-group">
        <input
          type="text"
          placeholder="Поиск студента..."
          value={searchTerm}
          onChange={handleInputChange}
          onFocus={handleInputFocus}
          onBlur={handleInputBlur}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          className="filter-select student-filter-search-input"
          aria-label="Поиск студента"
          aria-expanded={showDropdown}
          aria-autocomplete="list"
          aria-controls="student-filter-dropdown"
          aria-busy={loading}
        />
      </div>

      {showDropdown && (
        <div
          className="student-filter-search-dropdown"
          id="student-filter-dropdown"
          role="listbox"
          aria-label="Список студентов"
        >
          {loading ? (
            <div className="student-filter-search-item student-filter-search-loading">
              Загрузка...
            </div>
          ) : error ? (
            <div className="student-filter-search-item student-filter-search-error">
              <div>{error}</div>
            </div>
          ) : filteredStudents.length === 0 ? (
            <div className="student-filter-search-item student-filter-search-no-results">
              Студентов не найдено
            </div>
          ) : (
            <>
              <div role="group" aria-label="Студенты">
                {filteredStudents.map((student, index) => {
                  const isSelected = selectedStudents?.includes(String(student.id));
                  return (
                    <label
                      key={student.id}
                      className={`student-filter-search-item ${
                        isSelected ? 'checked' : ''
                      } ${highlightedIndex === index ? 'highlighted' : ''}`}
                      onKeyDown={(e) => {
                        if (e.key === ' ' || e.key === 'Enter') {
                          e.preventDefault();
                          handleStudentToggle(student);
                        }
                      }}
                      tabIndex={highlightedIndex === index ? 0 : -1}
                    >
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => handleStudentToggle(student)}
                        onClick={(e) => e.stopPropagation()}
                        onKeyDown={(e) => e.stopPropagation()}
                        role="checkbox"
                        aria-checked={isSelected}
                        disabled={disabled}
                      />
                      <span title={student.full_name}>
                        {student.full_name || 'Без имени'}
                      </span>
                    </label>
                  );
                })}
              </div>
              {selectedStudents && selectedStudents.length > 0 && (
                <button
                  onClick={handleClearAll}
                  disabled={disabled}
                  className="student-filter-search-clear-all-btn"
                  type="button"
                  aria-label="Очистить все выборы"
                >
                  Очистить все
                </button>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default StudentFilterSearch;
