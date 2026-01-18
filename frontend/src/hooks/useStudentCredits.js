import { useCredits } from './useCredits.js';

/**
 * Hook для баланса и истории кредитов студента
 * @deprecated Use useCredits() directly instead
 */
export const useStudentCredits = () => {
  // Use unified useCredits hook
  const credits = useCredits();

  return {
    credits: { balance: credits.balance ?? 0 },
    isLoading: credits.loading,
    error: credits.error,
    refetchBalance: credits.fetchCredits,
    history: credits.history,
    historyLoading: credits.historyLoading,
    historyError: credits.historyError,
    refetchHistory: credits.fetchHistory,
  };
};

export default useStudentCredits;
