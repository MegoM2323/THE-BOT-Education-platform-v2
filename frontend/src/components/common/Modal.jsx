import { useEffect } from "react";
import "./Modal.css";

export const Modal = ({
  isOpen,
  onClose,
  title,
  children,
  footer,
  loading = false,
  closeOnBackdropClick = false,
}) => {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "unset";
    }

    return () => {
      document.body.style.overflow = "unset";
    };
  }, [isOpen]);

  // ESC key disabled - modal closes only via X button

  if (!isOpen) return null;

  const handleBackdropClick = (e) => {
    if (e.target === e.currentTarget && closeOnBackdropClick) {
      onClose();
    }
    e.stopPropagation();
  };

  return (
    <div
      className="modal-backdrop modal"
      onClick={handleBackdropClick}
      data-testid="modal"
      role="dialog"
      aria-modal="true"
      style={{ cursor: closeOnBackdropClick ? "pointer" : "default" }}
    >
      <div className="modal-content" data-testid="modal-content">
        <div className="modal-header">
          <h3 className="modal-title" data-testid="modal-title">
            {title}
          </h3>
          <button
            className="modal-close"
            onClick={onClose}
            aria-label="Закрыть"
            data-testid="modal-close"
            disabled={loading}
          >
            &times;
          </button>
        </div>
        <div className="modal-body" data-testid="modal-body">
          {children}
        </div>
        {footer && <div className="modal-footer">{footer}</div>}
      </div>
    </div>
  );
};

export default Modal;
