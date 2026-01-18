/**
 * Bulk Edit Confirmation Dialog
 * Shows final confirmation before applying changes to all subsequent lessons
 */
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import AffectedLessonsPreview from './AffectedLessonsPreview.jsx';
import './BulkEditConfirmationDialog.css';

export const BulkEditConfirmationDialog = ({
  isOpen,
  onClose,
  onConfirm,
  modificationType,
  modificationDetails,
  affectedLessons = [],
  estimatedCount = 0,
  isLoading = false,
  error = null,
}) => {
  const getModificationTypeLabel = (type) => {
    const labels = {
      add_student: 'Add Student',
      remove_student: 'Remove Student',
      change_teacher: 'Change Teacher',
      change_time: 'Change Time',
      change_capacity: 'Change Capacity',
    };
    return labels[type] || 'Modify Lesson';
  };

  const getModificationDescription = () => {
    if (!modificationDetails) return '';

    switch (modificationType) {
      case 'add_student':
        return `Add ${modificationDetails.studentName} (ID: ${modificationDetails.studentId}) to all subsequent lessons`;
      case 'remove_student':
        return `Remove ${modificationDetails.studentName} (ID: ${modificationDetails.studentId}) from all subsequent lessons`;
      case 'change_teacher':
        return `Change teacher to ${modificationDetails.teacherName} (ID: ${modificationDetails.teacherId}) for all subsequent lessons`;
      case 'change_time':
        return `Change lesson time to ${modificationDetails.newStartTime} for all subsequent lessons`;
      case 'change_capacity':
        return `Change maximum students to ${modificationDetails.newMaxStudents} for all subsequent lessons`;
      default:
        return 'Apply modification to all subsequent lessons';
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Apply to All Subsequent Lessons"
      footer={
        <div className="bulk-edit-footer">
          <Button variant="secondary" onClick={onClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button variant="primary" onClick={onConfirm} loading={isLoading}>
            Yes, Apply to All
          </Button>
        </div>
      }
    >
      <div className="bulk-edit-confirmation-content">
        {isLoading ? (
          <div className="bulk-edit-loading">
            <Spinner />
            <p>Applying changes to all subsequent lessons...</p>
          </div>
        ) : (
          <>
            {/* Modification Type */}
            <div className="confirmation-section">
              <h4 className="section-label">Modification Type:</h4>
              <p className="modification-type">
                {getModificationTypeLabel(modificationType)}
              </p>
            </div>

            {/* Modification Details */}
            <div className="confirmation-section">
              <h4 className="section-label">Details:</h4>
              <p className="modification-details">{getModificationDescription()}</p>
            </div>

            {/* Affected Lessons Count */}
            <div className="confirmation-section">
              <h4 className="section-label">Affected Lessons:</h4>
              <p className="affected-count">
                {estimatedCount > 0
                  ? `Approximately ${estimatedCount} lesson${estimatedCount !== 1 ? 's' : ''} will be modified`
                  : 'Calculating affected lessons...'}
              </p>
            </div>

            {/* Preview of affected lessons (if available) */}
            {affectedLessons.length > 0 && (
              <AffectedLessonsPreview
                lessons={affectedLessons}
                totalCount={estimatedCount || affectedLessons.length}
              />
            )}

            {/* Warning */}
            <div className="confirmation-warning">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="warning-icon"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
              <div className="warning-content">
                <h5 className="warning-title">Warning</h5>
                <p className="warning-message">
                  This action cannot be undone for all lessons at once. Each lesson
                  will need to be modified individually to revert changes.
                </p>
              </div>
            </div>

            {/* Error display */}
            {error && (
              <div className="confirmation-error">
                <p>Error: {error instanceof Error ? error.message : (error?.message || String(error))}</p>
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  );
};

export default BulkEditConfirmationDialog;
