import './SkeletonLoader.css';

/**
 * Notice shown when connection is slow
 * Automatically appears after threshold is exceeded
 */
export const SlowConnectionNotice = ({ onDismiss }) => {
  return (
    <div className="slow-connection-notice" role="alert" aria-live="polite">
      <span className="icon">⏳</span>
      <div className="message">
        <strong>Медленное соединение</strong>
        <div>Загрузка данных занимает больше времени, чем обычно...</div>
      </div>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="dismiss-btn"
          aria-label="Закрыть уведомление"
        >
          ×
        </button>
      )}
    </div>
  );
};

export default SlowConnectionNotice;
