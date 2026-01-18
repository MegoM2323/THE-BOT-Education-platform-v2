/**
 * Utility hook for creating cancellable fetch requests
 * Prevents memory leaks by cancelling pending requests on component unmount
 *
 * @example
 * const fetchData = useCancellableRequest();
 *
 * useEffect(() => {
 *   fetchData(apiClient.get('/endpoint'))
 *     .then(data => setData(data))
 *     .catch(err => {
 *       if (err.name !== 'AbortError') {
 *         console.error('Error:', err);
 *       }
 *     });
 * }, [fetchData]);
 */

import { useRef, useCallback, useEffect } from 'react';
import { logger } from './logger.js';

/**
 * Hook для создания отменяемых запросов с автоматической очисткой при размонтировании
 *
 * IMPORTANT: Большинство hooks в этом проекте используют @tanstack/react-query,
 * который уже автоматически отменяет запросы при размонтировании компонента.
 *
 * Используйте этот hook только для:
 * - Запросов вне React Query (например, в AuthContext)
 * - Ручных fetch вызовов в useEffect
 * - Custom hooks без React Query
 *
 * @returns {Function} fetchWithCancel - функция для выполнения отменяемого запроса
 */
export const useCancellableRequest = () => {
  const abortControllerRef = useRef(null);
  const isMountedRef = useRef(true);

  useEffect(() => {
    isMountedRef.current = true;

    // Cleanup: отменяем все pending запросы при размонтировании
    return () => {
      isMountedRef.current = false;
      if (abortControllerRef.current) {
        logger.debug('Cancelling pending request on unmount');
        abortControllerRef.current.abort();
        abortControllerRef.current = null;
      }
    };
  }, []);

  /**
   * Выполняет запрос с возможностью отмены
   * @param {Function} requestFn - функция запроса (например, () => apiClient.get('/endpoint'))
   * @param {Object} options - дополнительные опции
   * @returns {Promise} - промис с результатом запроса
   */
  const fetchWithCancel = useCallback((requestFn, options = {}) => {
    // Отменяем предыдущий запрос если есть
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Создаём новый AbortController
    abortControllerRef.current = new AbortController();
    const signal = abortControllerRef.current.signal;

    // Выполняем запрос с signal
    return requestFn({ signal, ...options })
      .then(data => {
        // Проверяем что компонент ещё смонтирован
        if (!isMountedRef.current) {
          logger.debug('Component unmounted, ignoring response');
          return Promise.reject(new Error('Component unmounted'));
        }
        return data;
      })
      .catch(error => {
        // Не логируем AbortError как ошибку - это нормальное поведение
        if (error.name === 'AbortError') {
          logger.debug('Request cancelled:', error.message);
        }
        throw error;
      });
  }, []);

  return fetchWithCancel;
};

/**
 * Создаёт AbortController для ручного управления отменой запросов
 * Используется когда нужен более тонкий контроль над отменой
 *
 * @example
 * const controller = useAbortController();
 *
 * useEffect(() => {
 *   apiClient.get('/endpoint', { signal: controller.signal })
 *     .then(setData)
 *     .catch(err => {
 *       if (err.name !== 'AbortError') console.error(err);
 *     });
 *
 *   return () => controller.abort();
 * }, []);
 */
export const useAbortController = () => {
  const controllerRef = useRef(null);

  if (!controllerRef.current) {
    controllerRef.current = new AbortController();
  }

  useEffect(() => {
    const controller = controllerRef.current;

    // Cleanup: отменяем запрос при размонтировании
    return () => {
      if (controller) {
        logger.debug('Aborting request on unmount');
        controller.abort();
      }
    };
  }, []);

  return controllerRef.current;
};

export default useCancellableRequest;
