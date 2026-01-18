import Modal from "./Modal.jsx";
import Button from './Button.jsx';
import './ConfirmModal.css';

export const ConfirmModal = ({
  isOpen,
  onClose,
  onConfirm,
  title = 'Подтверждение',
  message,
  confirmText = 'Да',
  cancelText = 'Отмена',
  confirmVariant = 'primary',
  loading = false,
}) => {
  const handleConfirm = async () => {
    try {
      await onConfirm();
      onClose();
    } catch (error) {
      // Error handling is managed by parent component (AllLessons)
      // Modal stays open to allow user to retry or cancel
      throw error;
    }
  };

  const footer = (
    <div className="confirm-modal-actions">
      <Button variant="secondary" onClick={onClose} disabled={loading}>
        {cancelText}
      </Button>
      <Button variant={confirmVariant} onClick={handleConfirm} loading={loading}>
        {confirmText}
      </Button>
    </div>
  );

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title} footer={footer} loading={loading}>
      <div className="confirm-modal-message">{message}</div>
    </Modal>
  );
};

export default ConfirmModal;
