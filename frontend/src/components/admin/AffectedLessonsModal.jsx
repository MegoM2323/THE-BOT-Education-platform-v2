/**
 * Affected Lessons Modal
 * Shows detailed list of lessons that will be affected by bulk edit
 */
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import { formatTime } from '../../utils/dateFormat.js';
import './AffectedLessonsModal.css';

export const AffectedLessonsModal = ({
  isOpen,
  onClose,
  onConfirm,
  lessons = [],
  modificationType,
  modificationSummary,
  isLoading = false,
}) => {
  const getModificationTypeLabel = (type) => {
    const labels = {
      add_student: 'Adding Student',
      remove_student: 'Removing Student',
      change_teacher: 'Changing Teacher',
      change_time: 'Changing Time',
      change_capacity: 'Changing Capacity',
    };
    return labels[type] || 'Modifying Lessons';
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Affected Lessons Preview"
      footer={
        <div className="affected-lessons-footer">
          <Button variant="secondary" onClick={onClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button variant="primary" onClick={onConfirm} disabled={isLoading}>
            Continue
          </Button>
        </div>
      }
    >
      <div className="affected-lessons-content">
        {isLoading ? (
          <div className="affected-lessons-loading">
            <Spinner />
            <p>Loading affected lessons...</p>
          </div>
        ) : (
          <>
            {/* Summary */}
            <div className="affected-summary">
              <h4 className="summary-title">
                {getModificationTypeLabel(modificationType)}
              </h4>
              <p className="summary-description">{modificationSummary}</p>
              <p className="summary-count">
                This will affect{' '}
                <strong>
                  {lessons.length} lesson{lessons.length !== 1 ? 's' : ''}
                </strong>
              </p>
            </div>

            {/* Lessons table */}
            {lessons.length > 0 ? (
              <div className="lessons-table-container">
                <table className="lessons-table">
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Time</th>
                      <th>Teacher</th>
                      <th>Current Students</th>
                      <th>Max Students</th>
                    </tr>
                  </thead>
                  <tbody>
                    {lessons.map((lesson) => {
                      const lessonDate = new Date(lesson.start_time);
                      return (
                        <tr key={lesson.id}>
                          <td>
                            <div className="date-cell">
                              <span className="date-weekday">
                                {lessonDate.toLocaleDateString('ru-RU', {
                                  weekday: 'short',
                                })}
                              </span>
                              <span className="date-full">
                                {lessonDate.toLocaleDateString('ru-RU', {
                                  month: 'short',
                                  day: 'numeric',
                                  year: 'numeric',
                                })}
                              </span>
                            </div>
                          </td>
                          <td>
                            {formatTime(lesson.start_time)} -{' '}
                            {formatTime(lesson.end_time)}
                          </td>
                          <td>{lesson.teacher_name || 'Unknown'}</td>
                          <td className="centered">{lesson.current_students}</td>
                          <td className="centered">{lesson.max_students}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="no-lessons">
                <p>No lessons will be affected</p>
              </div>
            )}

            {/* Preview note */}
            <div className="preview-note">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="note-icon"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clipRule="evenodd"
                />
              </svg>
              <p>
                Review the lessons above carefully. This modification will be applied
                to all of them.
              </p>
            </div>
          </>
        )}
      </div>
    </Modal>
  );
};

export default AffectedLessonsModal;
