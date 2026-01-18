import { useState, useEffect } from 'react';

/**
 * Hook to detect slow network connections
 * Shows warning if data takes too long to load
 *
 * @param {boolean} isLoading - Current loading state
 * @param {number} threshold - Time in ms to consider connection slow (default: 3000ms)
 * @returns {boolean} - Whether connection appears to be slow
 */
export const useSlowConnection = (isLoading, threshold = 3000) => {
  const [isSlow, setIsSlow] = useState(false);

  useEffect(() => {
    if (!isLoading) {
      setIsSlow(false);
      return;
    }

    const timer = setTimeout(() => {
      if (isLoading) {
        setIsSlow(true);
      }
    }, threshold);

    return () => {
      clearTimeout(timer);
      setIsSlow(false);
    };
  }, [isLoading, threshold]);

  return isSlow;
};

export default useSlowConnection;
