import { useState } from 'react';
import Button from "../common/Button.jsx";
import TemplateSelectionModal from './TemplateSelectionModal.jsx';
import './TemplateApplyButton.css';

/**
 * Button to apply template to a week
 * @param {Date} weekStartDate - Monday of the week
 * @param {Function} onApplied - Callback after template applied
 * @param {boolean} disabled - Disable button
 * @param {boolean} templateApplied - Whether template already applied to this week
 */
export const TemplateApplyButton = ({
  weekStartDate,
  onApplied,
  disabled = false,
  templateApplied = false,
}) => {
  const [showModal, setShowModal] = useState(false);

  // Check if week is in the past
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const weekStart = new Date(weekStartDate);
  weekStart.setHours(0, 0, 0, 0);
  const isPast = weekStart < today;

  const handleClick = () => {
    setShowModal(true);
  };

  const handleApplied = (result) => {
    setShowModal(false);
    if (onApplied) {
      onApplied(result);
    }
  };

  const isDisabled = disabled || isPast;

  return (
    <>
      <Button
        variant="primary"
        size="small"
        onClick={handleClick}
        disabled={isDisabled}
        className="template-apply-btn"
        title={
          isPast
            ? 'Невозможно применить шаблон к прошедшим неделям'
            : templateApplied
            ? 'Переприменить шаблон к этой неделе (заменит существующие занятия из шаблона)'
            : 'Применить шаблон к этой неделе'
        }
      >
        {templateApplied ? 'Переприменить шаблон' : 'Применить шаблон'}
      </Button>

      {showModal && (
        <TemplateSelectionModal
          isOpen={showModal}
          onClose={() => setShowModal(false)}
          weekStartDate={weekStartDate}
          onApplied={handleApplied}
        />
      )}
    </>
  );
};

export default TemplateApplyButton;
