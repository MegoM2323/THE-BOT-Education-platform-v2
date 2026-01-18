import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import Spinner from '../common/Spinner.jsx';
import * as creditsAPI from '../../api/credits.js';
import './AdminCalendarCreditsDisplay.css';

/**
 * Component for displaying student credits in admin calendar navigation area
 * Shows credits for the currently selected student when admin is editing lessons
 * Uses React Query for automatic cache synchronization with other credit components
 *
 * Props:
 * - selectedStudentId: string (optional) - ID of the student to display credits for
 * - onCreditsLoaded: function (optional) - Callback when credits are loaded
 */
export const AdminCalendarCreditsDisplay = ({ selectedStudentId, onCreditsLoaded }) => {
  // Используем React Query для автоматической синхронизации с другими компонентами кредитов
  const {
    data: creditsData,
    isLoading: loading,
    error: queryError,
  } = useQuery({
    queryKey: ['credits', 'user', selectedStudentId],
    queryFn: async () => {
      const result = await creditsAPI.getUserCredits(selectedStudentId);
      return result;
    },
    enabled: !!selectedStudentId, // Запрос выполняется только если есть studentId
    staleTime: 30000, // 30 секунд
    refetchOnWindowFocus: true,
  });

  const balance = creditsData?.balance ?? null;
  const error = queryError?.message || null;

  // Уведомляем родителя когда кредиты загружены
  useEffect(() => {
    if (onCreditsLoaded && balance !== null) {
      onCreditsLoaded(balance);
    }
  }, [balance, onCreditsLoaded]);

  // Don't render if no student is selected
  if (!selectedStudentId) {
    return null;
  }

  const displayBalance = balance ?? 0;

  return (
    <div className="admin-calendar-credits-display" data-testid="admin-calendar-credits">
      <div className="admin-calendar-credits-content">
        {/* Credits Icon */}
        <svg
          className="admin-calendar-credits-icon"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          aria-hidden="true"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
        </svg>

        {/* Label */}
        <span className="admin-calendar-credits-label">Кредиты:</span>

        {/* Value or Loading/Error state */}
        {loading && balance === null ? (
          <Spinner size="sm" />
        ) : error ? (
          <span className="admin-calendar-credits-error" title={error}>
            Ошибка
          </span>
        ) : (
          <span className="admin-calendar-credits-value" data-testid="admin-calendar-credits-balance">
            {displayBalance}
          </span>
        )}
      </div>
    </div>
  );
};

export default AdminCalendarCreditsDisplay;
