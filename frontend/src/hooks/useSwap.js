import { useReducer, useCallback, useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import * as swapsAPI from '../api/swaps.js';
import { useNotification } from './useNotification.js';
import { invalidateCreditBalance } from './useCredits.js';

// Reducer для управления состоянием (избегаем race conditions)
const swapReducer = (state, action) => {
  switch (action.type) {
    case 'VALIDATE_START':
      return { ...state, validating: true };
    case 'VALIDATE_SUCCESS':
      return { ...state, validating: false, validationResult: action.payload };
    case 'VALIDATE_ERROR':
      return { ...state, validating: false };
    case 'SWAP_START':
      return { ...state, swapping: true };
    case 'SWAP_SUCCESS':
      return { ...state, swapping: false, validationResult: null };
    case 'SWAP_ERROR':
      return { ...state, swapping: false };
    case 'CLEAR_VALIDATION':
      return { ...state, validationResult: null };
    default:
      return state;
  }
};

const initialState = {
  validating: false,
  swapping: false,
  validationResult: null,
};

/**
 * Кастомный хук для управления обменами занятий
 * FIX T176: Используем useReducer для предотвращения race conditions при concurrent setState
 * FIX T173: Добавлена поддержка AbortController для отмены запросов при unmount
 */
export const useSwap = () => {
  const [state, dispatch] = useReducer(swapReducer, initialState);
  const notification = useNotification();
  const queryClient = useQueryClient();

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

  const validateSwap = useCallback(
    async (oldLessonId, newLessonId) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'VALIDATE_START' });
        const result = await swapsAPI.validateSwap(oldLessonId, newLessonId, { signal });
        dispatch({ type: 'VALIDATE_SUCCESS', payload: result });
        return result;
      } catch (err) {
        // Игнорируем AbortError - это нормальное поведение при unmount
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'VALIDATE_ERROR' });
        notification.error(err.message || 'Ошибка валидации обмена');
        throw err;
      } finally {
        cleanup();
      }
    },
    [notification, createRequest]
  );

  const performSwap = useCallback(
    async (oldLessonId, newLessonId) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'SWAP_START' });
        const result = await swapsAPI.performSwap(oldLessonId, newLessonId, { signal });
        dispatch({ type: 'SWAP_SUCCESS' });
        notification.success('Занятие успешно обменено');

        // Инвалидация кэша после успешного swap - занятия и бронирования изменились
        queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
        invalidateCreditBalance(queryClient);

        return result;
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'SWAP_ERROR' });
        notification.error(err.message || 'Ошибка обмена занятия');
        throw err;
      } finally {
        cleanup();
      }
    },
    [notification, createRequest, queryClient]
  );

  const clearValidation = useCallback(() => {
    dispatch({ type: 'CLEAR_VALIDATION' });
  }, []);

  return {
    validating: state.validating,
    swapping: state.swapping,
    validationResult: state.validationResult,
    validateSwap,
    performSwap,
    clearValidation,
  };
};

export default useSwap;
