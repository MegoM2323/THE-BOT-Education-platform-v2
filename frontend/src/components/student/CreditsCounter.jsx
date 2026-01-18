import { useCredits } from '../../hooks/useCredits.js';
import Spinner from '../common/Spinner.jsx';
import './CreditsCounter.css';

export const CreditsCounter = ({ inline = false }) => {
  // useCredits now has improved capabilities:
  // - staleTime: 30000ms (increased from 10000ms)
  // - refetchOnWindowFocus: true (auto-refresh when returning)
  // - retry: 3 with exponential backoff (auto-retry on error)
  const { balance, loading, error, fetchCredits } = useCredits();

  const displayBalance = balance ?? 0;

  // Retry handler - user can manually refresh on error
  const handleRetry = () => {
    fetchCredits();
  };

  if (inline) {
    return (
      <div className="credits-counter-inline" data-testid="credits-counter">
        <span className="credits-counter-inline-icon">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
          </svg>
        </span>
        {loading && balance === null ? (
          <Spinner size="sm" />
        ) : error ? (
          <button
            className="credits-counter-error-btn"
            onClick={handleRetry}
            title={error}
            aria-label="Повторить загрузку баланса"
          >
            Ошибка
          </button>
        ) : (
          <span className="credits-counter-inline-value">{displayBalance} кредитов</span>
        )}
      </div>
    );
  }

  return (
    <div className="credits-counter" data-testid="credits-counter">
      <div className="credits-label">Ваш баланс</div>
      <div className="credits-value" data-testid="credits-balance">
        {loading && balance === null ? (
          <Spinner />
        ) : error ? (
          <div className="error-container">
            <span className="error-text" title={error}>
              Ошибка загрузки
            </span>
            <button
              className="error-retry-btn"
              onClick={handleRetry}
              aria-label="Повторить загрузку баланса"
            >
              Повторить
            </button>
          </div>
        ) : (
          displayBalance
        )}
      </div>
      <div className="credits-unit">кредитов</div>
    </div>
  );
};

export default CreditsCounter;
