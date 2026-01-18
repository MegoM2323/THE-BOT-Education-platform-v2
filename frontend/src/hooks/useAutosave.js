import { useEffect, useRef, useCallback, useState } from 'react';

/**
 * Hook для автосохранения с debounce и retry
 *
 * @param {*} data - Данные для сохранения
 * @param {Function} saveFunction - Асинхронная функция сохранения
 * @param {number} delay - Задержка debounce в миллисекундах (по умолчанию 500)
 * @param {number} maxRetries - Максимальное количество попыток (по умолчанию 3)
 * @returns {Object} { isSaving, lastSaved, saveNow, error }
 */
export const useAutosave = (data, saveFunction, delay = 500, maxRetries = 3) => {
  const [isSaving, setIsSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState(null);
  const [error, setError] = useState(null);
  const timeoutRef = useRef(null);
  const previousDataRef = useRef(data);
  const saveInProgressRef = useRef(false);
  const retryCountRef = useRef(0);
  const isMountedRef = useRef(true);
  const retryTimeoutRef = useRef(null);

  /**
   * Retry с экспоненциальной задержкой (с возможностью отмены)
   */
  const sleep = (ms) => new Promise((resolve) => {
    retryTimeoutRef.current = setTimeout(resolve, ms);
  });

  /**
   * Сохранить данные с retry логикой
   */
  const saveWithRetry = useCallback(async (currentData, retryCount = 0) => {
    // Проверить, что компонент все еще смонтирован
    if (!isMountedRef.current) {
      console.log('[Autosave] Component unmounted, cancelling save');
      return false;
    }

    try {
      await saveFunction(currentData);

      // Проверить снова перед setState
      if (!isMountedRef.current) {
        return false;
      }

      setLastSaved(new Date());
      setError(null);
      retryCountRef.current = 0;
      previousDataRef.current = currentData;
      return true;
    } catch (err) {
      console.error(`Autosave attempt ${retryCount + 1} failed:`, err);

      // Проверить, что компонент все еще смонтирован
      if (!isMountedRef.current) {
        console.log('[Autosave] Component unmounted during retry, cancelling');
        return false;
      }

      // Если это 400 ошибка, не ретраим (плохие данные)
      const status = err.response?.status;
      if (status === 400) {
        console.error('Bad request - data validation error, not retrying');
        if (isMountedRef.current) {
          setError(err);
          retryCountRef.current = 0;
        }
        throw err;
      }

      // Если достигли максимума попыток
      if (retryCount >= maxRetries - 1) {
        console.error('Max retries reached');
        if (isMountedRef.current) {
          setError(err);
          retryCountRef.current = 0;
        }
        throw err;
      }

      // Экспоненциальная задержка: 1s, 2s, 4s
      const delayMs = Math.pow(2, retryCount) * 1000;
      console.log(`Retrying in ${delayMs}ms...`);

      try {
        await sleep(delayMs);
      } catch (sleepErr) {
        // Sleep был отменён
        console.log('[Autosave] Sleep cancelled, stopping retry');
        return false;
      }

      // Проверить снова перед рекурсией
      if (!isMountedRef.current) {
        console.log('[Autosave] Component unmounted after sleep, cancelling retry');
        return false;
      }

      // Рекурсивный retry
      return saveWithRetry(currentData, retryCount + 1);
    }
  }, [saveFunction, maxRetries]);

  /**
   * Сохранить данные немедленно (без debounce)
   */
  const saveNow = useCallback(async () => {
    if (!isMountedRef.current) {
      console.log('[Autosave] Component unmounted, skipping save');
      return;
    }

    if (saveInProgressRef.current) {
      // Если уже идет сохранение, пропустить
      console.log('[Autosave] Save already in progress, skipping');
      return;
    }

    try {
      saveInProgressRef.current = true;
      if (isMountedRef.current) {
        setIsSaving(true);
      }
      await saveWithRetry(data, retryCountRef.current);
    } catch (error) {
      console.error('Autosave failed after retries:', error);
      throw error;
    } finally {
      if (isMountedRef.current) {
        setIsSaving(false);
      }
      saveInProgressRef.current = false;
    }
  }, [data, saveWithRetry]);

  /**
   * Автосохранение с debounce при изменении данных
   */
  useEffect(() => {
    // Проверить, изменились ли данные
    if (JSON.stringify(data) === JSON.stringify(previousDataRef.current)) {
      return;
    }

    // Очистить предыдущий таймер
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    // Установить новый таймер для автосохранения
    timeoutRef.current = setTimeout(() => {
      if (isMountedRef.current) {
        saveNow();
      }
    }, delay);

    // Cleanup
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    };
  }, [data, delay]); // Убрали saveNow из зависимостей - это устраняет бесконечный цикл

  /**
   * Очистка при размонтировании
   */
  useEffect(() => {
    // Устанавливаем флаг при монтировании
    isMountedRef.current = true;

    return () => {
      // Сбрасываем флаг при размонтировании
      isMountedRef.current = false;

      // Очищаем все таймеры
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }

      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
        retryTimeoutRef.current = null;
      }

      // Сбрасываем флаг сохранения
      saveInProgressRef.current = false;

      console.log('[Autosave] Component unmounted, all timers cleared');
    };
  }, []);

  return {
    isSaving,
    lastSaved,
    error,
    saveNow, // Для принудительного сохранения при закрытии
  };
};

export default useAutosave;
