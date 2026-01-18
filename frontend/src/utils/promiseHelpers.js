/**
 * Promise utilities for handling partial failures gracefully
 */

/**
 * Like Promise.all but allows partial failures
 * Returns { results, failures } where:
 * - results: array of { status: 'fulfilled'|'rejected', value: any, index: number }
 * - failures: array of { error: Error, index: number, label: string }
 *
 * @param {Promise[]} promises - Array of promises to execute
 * @param {string[]} labels - Optional labels for each promise (for error reporting)
 * @returns {Promise<{results: Array, failures: Array, hasFailures: boolean}>}
 */
export const allSettledWithLabels = async (promises, labels = []) => {
  const results = await Promise.allSettled(promises);

  const failures = [];
  const enrichedResults = results.map((result, index) => {
    if (result.status === 'rejected') {
      failures.push({
        error: result.reason,
        index,
        label: labels[index] || `Promise ${index}`,
      });
    }
    return {
      ...result,
      index,
      label: labels[index] || `Promise ${index}`,
    };
  });

  return {
    results: enrichedResults,
    failures,
    hasFailures: failures.length > 0,
  };
};

/**
 * Extract successful values from allSettled results
 * @param {Array} results - Results from Promise.allSettled
 * @returns {Array} - Array of successful values
 */
export const getSuccessfulValues = (results) => {
  return results
    .filter(result => result.status === 'fulfilled')
    .map(result => result.value);
};

/**
 * Extract failed reasons from allSettled results
 * @param {Array} results - Results from Promise.allSettled
 * @returns {Array} - Array of errors
 */
export const getFailedReasons = (results) => {
  return results
    .filter(result => result.status === 'rejected')
    .map(result => result.reason);
};

/**
 * Helper to retry a single failed promise
 * @param {Function} promiseFn - Function that returns a promise
 * @param {number} maxRetries - Maximum number of retries
 * @param {number} delay - Delay between retries in ms
 * @returns {Promise}
 */
export const retryPromise = async (promiseFn, maxRetries = 3, delay = 1000) => {
  let lastError;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await promiseFn();
    } catch (error) {
      lastError = error;
      if (attempt < maxRetries) {
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }
  }

  throw lastError;
};

/**
 * Create a fallback value for failed promises
 * @param {any} result - Result from Promise.allSettled
 * @param {any} fallbackValue - Value to use if promise failed
 * @returns {any}
 */
export const withFallback = (result, fallbackValue) => {
  return result.status === 'fulfilled' ? result.value : fallbackValue;
};
