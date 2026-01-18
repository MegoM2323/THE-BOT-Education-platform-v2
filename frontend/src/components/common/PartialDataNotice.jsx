import { useState } from 'react';
import './PartialDataNotice.css';

/**
 * Component to show partial data load warnings with retry capability
 * @param {Object} props
 * @param {Array} props.failures - Array of failed items { label, error, index }
 * @param {Function} props.onRetry - Function to retry failed items
 * @param {boolean} props.retrying - Whether retry is in progress
 */
export const PartialDataNotice = ({ failures, onRetry, retrying = false }) => {
  const [expanded, setExpanded] = useState(false);

  if (!failures || failures.length === 0) {
    return null;
  }

  return (
    <div className="partial-data-notice" role="alert">
      <div className="partial-data-header">
        <div className="partial-data-icon">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
            <path
              d="M10 6V10M10 14H10.01M19 10C19 14.9706 14.9706 19 10 19C5.02944 19 1 14.9706 1 10C1 5.02944 5.02944 1 10 1C14.9706 1 19 5.02944 19 10Z"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
        <div className="partial-data-message">
          <strong>Частичная загрузка данных</strong>
          <span className="partial-data-summary">
            Не удалось загрузить: {failures.map(f => f.label).join(', ')}
          </span>
        </div>
        <button
          className="partial-data-toggle"
          onClick={() => setExpanded(!expanded)}
          aria-label={expanded ? 'Скрыть детали' : 'Показать детали'}
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            style={{ transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)' }}
          >
            <path
              d="M4 6L8 10L12 6"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>
      </div>

      {expanded && (
        <div className="partial-data-details">
          <ul className="partial-data-list">
            {failures.map(({ label, error }, idx) => (
              <li key={idx} className="partial-data-item">
                <strong>{label}:</strong>{' '}
                <span className="error-message">
                  {error?.message || error?.toString() || 'Неизвестная ошибка'}
                </span>
              </li>
            ))}
          </ul>
          {onRetry && (
            <button
              className="partial-data-retry-btn"
              onClick={onRetry}
              disabled={retrying}
            >
              {retrying ? (
                <>
                  <span className="retry-spinner"></span>
                  Повтор...
                </>
              ) : (
                'Повторить загрузку'
              )}
            </button>
          )}
        </div>
      )}
    </div>
  );
};

export default PartialDataNotice;
