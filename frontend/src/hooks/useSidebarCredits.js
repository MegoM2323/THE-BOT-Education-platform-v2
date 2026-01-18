import { useCredits } from './useCredits.js';

/**
 * Hook для баланса кредитов в sidebar с автообновлением
 * @deprecated Use useCredits({ refetchInterval: 10000 }) directly instead
 */
export const useSidebarCredits = (options = {}) => {
  const { interval = 10000, enabled = true } = options;

  // Use unified useCredits hook with polling
  const credits = useCredits({
    refetchInterval: enabled ? interval : false,
  });

  return {
    balance: credits.balance ?? 0,
    loading: credits.loading,
    error: credits.error,
    refetch: credits.refetch,
    isPolling: enabled,
    // Deprecated methods - kept for backward compatibility with console warnings
    startPolling: () => {
      console.warn('useSidebarCredits.startPolling() is deprecated. Use useCredits({ refetchInterval }) instead');
    },
    stopPolling: () => {
      console.warn('useSidebarCredits.stopPolling() is deprecated. Use useCredits({ refetchInterval: false }) instead');
    },
    lastFetchTime: Date.now(),
  };
};

export default useSidebarCredits;
