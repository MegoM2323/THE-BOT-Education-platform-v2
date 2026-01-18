import { useReducer, useEffect, useCallback, useRef } from 'react';
import * as telegramAPI from '../api/telegram.js';
import { useNotification } from './useNotification.js';

// Reducer для управления состоянием (избегаем race conditions)
const telegramReducer = (state, action) => {
  switch (action.type) {
    case 'FETCH_START':
      return { ...state, loading: true, error: null };
    case 'FETCH_LINK_STATUS_SUCCESS':
      return { ...state, loading: false, linkStatus: action.payload };
    case 'GENERATE_TOKEN_SUCCESS':
      return {
        ...state,
        loading: false,
        token: action.payload.token,
        botUsername: action.payload.bot_username
      };
    case 'GENERATE_TOKEN_ERROR_WITH_STATUS':
      // Особый случай: ошибка + обновленный статус
      return {
        ...state,
        loading: false,
        error: action.payload.error,
        linkStatus: action.payload.linkStatus
      };
    case 'UNLINK_SUCCESS':
      return {
        ...state,
        loading: false,
        linkStatus: { linked: false },
        token: null,
        botUsername: null
      };
    case 'FETCH_ERROR':
      return { ...state, loading: false, error: action.payload };
    case 'CLEAR_ERROR':
      return { ...state, error: null };
    default:
      return state;
  }
};

const initialState = {
  linkStatus: null,
  token: null,
  botUsername: null,
  loading: true,
  error: null,
};

/**
 * Кастомный хук для управления привязкой Telegram
 * FIX T176: Используем useReducer для предотвращения race conditions при concurrent setState
 * FIX T173: Добавлена поддержка AbortController для отмены запросов при unmount
 */
export const useTelegram = () => {
  const [state, dispatch] = useReducer(telegramReducer, initialState);
  const notification = useNotification();

  // FIX T173: Ref для хранения активных AbortController
  const abortControllersRef = useRef([]);

  // Cleanup на unmount - отменяем все активные запросы
  useEffect(() => {
    const controllers = abortControllersRef.current;
    return () => {
      controllers.forEach(controller => {
        controller.abort();
      });
      controllers.length = 0;
    };
  }, []);

  // Утилита для создания AbortController с автоматической очисткой
  const createRequest = useCallback(() => {
    const controller = new AbortController();
    abortControllersRef.current.push(controller);

    const cleanup = () => {
      const index = abortControllersRef.current.indexOf(controller);
      if (index > -1) {
        abortControllersRef.current.splice(index, 1);
      }
    };

    return { signal: controller.signal, cleanup };
  }, []);

  /**
   * Загрузить статус привязки Telegram
   */
  const fetchLinkStatus = useCallback(async () => {
    const { signal, cleanup } = createRequest();
    try {
      dispatch({ type: 'FETCH_START' });
      const data = await telegramAPI.getMyTelegramLink({ signal });
      dispatch({ type: 'FETCH_LINK_STATUS_SUCCESS', payload: data });
    } catch (err) {
      // Игнорируем AbortError - это нормальное поведение при unmount
      if (err.name === 'AbortError') {
        return;
      }
      console.error('Telegram link status error:', err?.response?.status, err?.message);
      dispatch({ type: 'FETCH_ERROR', payload: err.message });
      // НЕ показываем notification здесь - это автоматическая загрузка,
      // а не юзер-экшн. Если Telegram не привязан - это нормально.
    } finally {
      cleanup();
    }
  }, [createRequest]);

  /**
   * Сгенерировать токен для привязки Telegram
   */
  const generateToken = useCallback(async () => {
    const { signal, cleanup } = createRequest();
    try {
      dispatch({ type: 'FETCH_START' });
      const data = await telegramAPI.generateLinkToken({ signal });
      dispatch({ type: 'GENERATE_TOKEN_SUCCESS', payload: data });
      notification.success('Токен создан. Откройте бота в Telegram для привязки');
      return data;
    } catch (err) {
      // Игнорируем AbortError
      if (err.name === 'AbortError') {
        return;
      }

      // Если получили ошибку привязки - обновляем статус для актуального состояния
      // Это может быть 409 "already linked" или другая ошибка
      const errorMessage = err.message;
      console.error('Telegram token generation error:', err?.response?.status, errorMessage);
      notification.error(errorMessage || 'Ошибка генерации токена');

      // Перезагружаем статус привязки для корректного отображения
      // (может быть привязан через другую вкладку/устройство)
      // Важно: сохраняем ошибку после перезагрузки статуса
      try {
        const updatedStatus = await telegramAPI.getMyTelegramLink({ signal });
        dispatch({
          type: 'GENERATE_TOKEN_ERROR_WITH_STATUS',
          payload: { error: errorMessage, linkStatus: updatedStatus }
        });
        // НЕ сбрасываем ошибку - пользователь должен видеть что произошло
      } catch (fetchError) {
        // Игнорируем AbortError при перезагрузке статуса
        if (fetchError.name === 'AbortError') {
          return;
        }
        // Если не удалось перезагрузить статус, просто устанавливаем ошибку
        console.error('Failed to refetch link status:', fetchError);
        dispatch({ type: 'FETCH_ERROR', payload: errorMessage });
      }

      throw err;
    } finally {
      cleanup();
    }
  }, [notification, createRequest]);

  /**
   * Отвязать Telegram аккаунт
   */
  const unlinkAccount = useCallback(async () => {
    const { signal, cleanup } = createRequest();
    try {
      dispatch({ type: 'FETCH_START' });
      await telegramAPI.unlinkTelegram({ signal });
      dispatch({ type: 'UNLINK_SUCCESS' });
      notification.success('Telegram успешно отвязан');
    } catch (err) {
      // Игнорируем AbortError
      if (err.name === 'AbortError') {
        return;
      }
      console.error('Telegram unlink error:', err?.response?.status, err?.message);
      dispatch({ type: 'FETCH_ERROR', payload: err.message });
      notification.error(err.message || 'Ошибка отвязки Telegram');
      throw err;
    } finally {
      cleanup();
    }
  }, [notification, createRequest]);

  // Загрузить статус привязки при монтировании
  useEffect(() => {
    fetchLinkStatus();
  }, [fetchLinkStatus]);

  // Auto-refresh linkStatus when user returns to tab (after Telegram linking)
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        fetchLinkStatus(); // Refetch when returning to active tab
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [fetchLinkStatus]);

  /**
   * Очистить ошибки вручную (для retry)
   */
  const clearError = useCallback(() => {
    dispatch({ type: 'CLEAR_ERROR' });
  }, []);

  return {
    linkStatus: state.linkStatus,
    token: state.token,
    botUsername: state.botUsername,
    loading: state.loading,
    error: state.error,
    generateToken,
    unlinkAccount,
    refetchLinkStatus: fetchLinkStatus,
    clearError,
  };
};

export default useTelegram;
