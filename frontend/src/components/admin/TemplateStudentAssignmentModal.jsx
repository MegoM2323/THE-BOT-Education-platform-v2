import { useState, useEffect, useCallback } from 'react';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import * as usersAPI from '../../api/users.js';
import * as creditAPI from '../../api/credits.js';
import './TemplateStudentAssignmentModal.css';

export const TemplateStudentAssignmentModal = ({
  isOpen,
  onClose,
  selectedStudents = [],
  onStudentsAssigned,
  maxStudents = null // Новое: передаем максимальную вместимость занятия
}) => {
  const notification = useNotification();
  const [students, setStudents] = useState([]);
  const [studentCredits, setStudentCredits] = useState({});
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [localSelectedStudents, setLocalSelectedStudents] = useState([]);

  const fetchStudents = useCallback(async () => {
    try {
      setLoading(true);
      const [studentsData, creditsData] = await Promise.all([
        usersAPI.getStudentsAll(),
        creditAPI.getAllCredits()
      ]);
      setStudents(studentsData);

      // Создать карту кредитов
      const creditsMap = {};
      if (creditsData?.balances && Array.isArray(creditsData.balances)) {
        creditsData.balances.forEach(({ user_id, balance }) => {
          creditsMap[user_id] = balance || 0;
        });
      }
      setStudentCredits(creditsMap);
    } catch (error) {
      notification.error('Ошибка загрузки студентов');
    } finally {
      setLoading(false);
    }
  }, [notification]);

  useEffect(() => {
    if (isOpen) {
      fetchStudents();
      setLocalSelectedStudents([...selectedStudents]);
    }
  }, [isOpen, selectedStudents, fetchStudents]);

  const filteredStudents = students.filter((student) => {
    const name = (student.full_name || student.name || '').toLowerCase();
    const email = (student.email || '').toLowerCase();
    const query = searchQuery.toLowerCase();
    return name.includes(query) || email.includes(query);
  });

  const handleToggleStudent = (studentId) => {
    setLocalSelectedStudents((prev) => {
      if (prev.includes(studentId)) {
        return prev.filter((id) => id !== studentId);
      } else {
        // Проверка вместимости перед добавлением студента
        if (maxStudents !== null && prev.length >= maxStudents) {
          notification.error(`Невозможно добавить больше студентов. Максимальная вместимость занятия: ${maxStudents}`);
          return prev;
        }
        // Проверка кредитов перед добавлением студента
        const credits = studentCredits[studentId] || 0;
        if (credits < 1) {
          notification.error('У студента недостаточно кредитов для добавления на занятие');
          return prev;
        }
        return [...prev, studentId];
      }
    });
  };


  const handleConfirm = () => {
    onStudentsAssigned(localSelectedStudents);
    onClose();
  };

  const handleCancel = () => {
    setLocalSelectedStudents([...selectedStudents]);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleCancel}
      title="Assign Students to Template Lesson"
      loading={loading}
    >
      <div className="student-assignment-modal">
        {loading ? (
          <div className="student-assignment-loading">
            <Spinner />
          </div>
        ) : (
          <>
            <div className="student-assignment-header">
              <input
                type="text"
                className="student-search-input"
                placeholder="Search students by name or email..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                autoFocus
              />
            </div>

            <div className="student-count">
              {localSelectedStudents.length} of {students.length} students selected
              {maxStudents !== null && (
                <span className="capacity-indicator">
                  {' '}(Вместимость: {localSelectedStudents.length}/{maxStudents})
                  {localSelectedStudents.length >= maxStudents && (
                    <span className="capacity-warning"> - Занятие заполнено</span>
                  )}
                </span>
              )}
            </div>

            <div className="student-list">
              {filteredStudents.length === 0 ? (
                <div className="student-list-empty">
                  {searchQuery ? 'No students found matching your search' : 'No students available'}
                </div>
              ) : (
                filteredStudents.map((student) => {
                  const isSelected = localSelectedStudents.includes(student.id);
                  const credits = studentCredits[student.id] || 0;
                  const hasLowCredits = credits < 1;
                  // Проверка: достигнута ли максимальная вместимость и студент еще не выбран
                  const isDisabledByCapacity = maxStudents !== null &&
                                               localSelectedStudents.length >= maxStudents &&
                                               !isSelected;
                  // Студенты без кредитов не могут быть добавлены (кроме уже выбранных)
                  const isDisabled = (isDisabledByCapacity || (hasLowCredits && !isSelected));

                  return (
                    <div
                      key={student.id}
                      className={`student-item ${isSelected ? 'selected' : ''} ${isDisabled ? 'disabled' : ''} ${hasLowCredits ? 'low-credits' : ''}`}
                      onClick={() => !isDisabled && handleToggleStudent(student.id)}
                    >
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => handleToggleStudent(student.id)}
                        onClick={(e) => e.stopPropagation()}
                        disabled={isDisabled}
                      />
                      <div className="student-info">
                        <div className="student-name">{student.full_name || student.name || 'Unknown'}</div>
                        <div className="student-email">{student.email || ''}</div>
                      </div>
                      <div className="student-credits-container">
                        <span className={`student-credits ${hasLowCredits ? 'credits-warning' : ''}`}>
                          {credits} kr.
                        </span>
                        {hasLowCredits && !isSelected && (
                          <span className="credits-warning-icon" title="Insufficient credits">
                            !
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            <div className="student-assignment-footer">
              <Button variant="secondary" onClick={handleCancel}>
                Cancel
              </Button>
              <Button onClick={handleConfirm}>
                Confirm ({localSelectedStudents.length})
              </Button>
            </div>
          </>
        )}
      </div>
    </Modal>
  );
};

export default TemplateStudentAssignmentModal;
