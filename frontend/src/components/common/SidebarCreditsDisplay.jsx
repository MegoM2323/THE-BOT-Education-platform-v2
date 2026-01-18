import React from 'react';
import { useSidebarCredits } from '../../hooks/useSidebarCredits.js';
import Spinner from './Spinner.jsx';
import './SidebarCreditsDisplay.css';

/**
 * Компонент для отображения баланса кредитов в sidebar
 * Автоматически обновляет баланс каждые 10 сек
 */
export const SidebarCreditsDisplay = ({ collapsed = false, interval = 10000 }) => {
  const { balance, loading, error, refetch } = useSidebarCredits({ interval });

  const displayBalance = balance ?? 0;

  const handleRefresh = (e) => {
    e.preventDefault();
    e.stopPropagation();
    refetch();
  };

  return (
    <div className="sidebar-credits-display" data-testid="sidebar-credits">
      <div className="sidebar-credits-header">
        <svg
          className="sidebar-credits-icon"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
        </svg>

        {!collapsed && (
          <span className="sidebar-credits-label">Кредиты</span>
        )}
      </div>

      <div className="sidebar-credits-content">
        {loading && balance === null ? (
          <Spinner size="sm" />
        ) : error ? (
          <span className="sidebar-credits-error" title={error}>
            Ошибка
          </span>
        ) : (
          <div className="sidebar-credits-value-group">
            <span className="sidebar-credits-value" data-testid="sidebar-credits-balance">
              {displayBalance}
            </span>
            {!collapsed && (
              <span className="sidebar-credits-unit">кр.</span>
            )}
          </div>
        )}

        {!collapsed && !error && (
          <button
            className="sidebar-credits-refresh"
            onClick={handleRefresh}
            title="Обновить баланс"
            aria-label="Обновить баланс кредитов"
            disabled={loading}
            data-testid="sidebar-credits-refresh"
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <polyline points="23 4 23 10 17 10" />
              <path d="M20.49 15a9 9 0 1 1-2-8.12" />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
};

export default SidebarCreditsDisplay;
