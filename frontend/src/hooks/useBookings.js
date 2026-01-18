import { useCallback, useMemo } from 'react';
import { logger } from '../utils/logger.js';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as bookingsAPI from '../api/bookings.js';
import { useNotification } from './useNotification.js';
import { invalidateCreditBalance } from './useCredits.js';

/**
 * Анимационная константа для синхронизации с CSS анимациями
 * Соответствует: fadeOutSlide (200ms) + collapseHeight (200ms with 200ms delay) = 400ms total
 */
export const ANIMATION_DURATION = 400; // ms

/**
 * Определяет, является ли ошибка сетевой и требует retry
 */
const isNetworkError = (err) => {
  if (!err) return false;
  const errorMsg = err.message?.toLowerCase() || '';
  const errorCode = err.code || '';

  // Сетевые и временные ошибки
  const networkPatterns = [
    'network',
    'timeout',
    'econnrefused',
    'econnreset',
    'enotfound',
    'socket hang up'
  ];

  return networkPatterns.some(pattern =>
    errorMsg.includes(pattern) || errorCode.includes(pattern)
  );
};

/**
 * Утилита для retry логики при СЕТЕВЫХ ошибках (не бизнес-логики)
 */
const retryAsync = async (fn, maxRetries = 3, delayMs = 500) => {
  let lastError;
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;

      // Не retry бизнес-логики ошибок (например "not active")
      if (!isNetworkError(err)) {
        throw err;
      }

      if (i < maxRetries - 1) {
        await new Promise((resolve) => setTimeout(resolve, delayMs * (i + 1)));
      }
    }
  }
  throw lastError;
};

/**
 * Кастомный хук для управления записями на занятия с оптимистичными обновлениями
 */
export const useBookings = (filters = {}) => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Мемоизировать фильтры чтобы избежать бесконечного цикла
  // Используем JSON.stringify для стабильного сравнения объектов
  const filtersJson = JSON.stringify(filters);
  const memoizedFilters = useMemo(
    () => filters,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [filtersJson]
  );

  // Уникальный ключ для кеша, основанный на фильтрах
  const queryKey = useMemo(() => ['bookings', memoizedFilters], [memoizedFilters]);

  // Нормализация данных для валидации
  const normalizeBookings = (data) => {
    const bookingsArray = Array.isArray(data) ? data : (data?.bookings || []);

    if (!Array.isArray(bookingsArray)) {
      logger.warn('normalizeBookings: Invalid data structure received', data);
      return [];
    }

    // Убедимся что все бронирования имеют обязательные поля
    const validBookings = bookingsArray.filter((booking) => {
      if (!booking || typeof booking !== 'object') {
        logger.warn('normalizeBookings: Invalid booking structure', booking);
        return false;
      }

      if (!booking.id && !booking.booking_id) {
        logger.warn('normalizeBookings: Booking missing id field', booking);
        return false;
      }

      if (!booking.lesson_id) {
        logger.warn('normalizeBookings: Booking missing lesson_id field', booking);
        return false;
      }

      if (!booking.start_time && !booking.lesson?.start_time) {
        logger.warn('normalizeBookings: Booking missing start_time field', booking);
        return false;
      }

      return true;
    });

    if (validBookings.length !== bookingsArray.length) {
      logger.info(
        `normalizeBookings: Filtered ${bookingsArray.length - validBookings.length} invalid bookings`
      );
    }

    return validBookings;
  };

  // useQuery для получения записей с фильтрами
  const {
    data: bookings = [],
    isLoading: loading,
    error: queryError,
    refetch: refetchQuery,
  } = useQuery({
    queryKey,
    queryFn: async () => {
      logger.debug('useQuery: Fetching bookings with filters:', memoizedFilters);
      try {
        const data = await bookingsAPI.getBookings(memoizedFilters);
        logger.debug('useQuery: Raw data received:', data);
        const normalized = normalizeBookings(data);
        logger.debug('useQuery: Successfully loaded', normalized.length, 'bookings');
        return normalized;
      } catch (err) {
        console.error('useQuery error:', err.message);
        notification.error('Ошибка загрузки записей');
        throw err;
      }
    },
  });

  // Отдельный query для "мои записи"
  const {
    data: myBookings = [],
    isLoading: myBookingsLoading,
    error: myBookingsError,
    refetch: refetchMyBookings,
  } = useQuery({
    queryKey: ['myBookings'],
    queryFn: async () => {
      logger.debug('useQuery: Fetching my bookings');
      try {
        const data = await bookingsAPI.getMyBookings();
        logger.debug('useQuery: Raw data received:', data);
        const normalized = normalizeBookings(data);
        logger.debug('useQuery: Successfully loaded', normalized.length, 'bookings');
        return normalized;
      } catch (err) {
        console.error('useQuery error:', err.message);
        notification.error('Ошибка загрузки ваших записей');
        throw err;
      }
    },
    enabled: false, // По умолчанию не загружаем, загружаем по запросу
  });

  // Mutation для создания записи
  const createBookingMutation = useMutation({
    mutationFn: async (lessonId) => {
      logger.debug('createBooking: Starting with lessonId:', lessonId);
      const newBooking = await retryAsync(
        () => bookingsAPI.createBooking(lessonId),
        3,
        500
      );
      logger.debug('createBooking: Created successfully:', newBooking);
      return newBooking;
    },
    onMutate: async (lessonId) => {
      // Отменяем любые выполняющиеся запросы к lessons
      await queryClient.cancelQueries({ queryKey: ['lessons'] });

      // Сохраняем предыдущее состояние lessons для отката при ошибке
      const previousLessons = queryClient.getQueryData(['lessons']);

      // Оптимистично обновляем lessons - увеличиваем current_students на 1 (место заполняется)
      queryClient.setQueryData(['lessons'], (old) => {
        if (!Array.isArray(old)) return old;
        return old.map((lesson) => {
          if (lesson.id === lessonId) {
            return {
              ...lesson,
              current_students: Math.max(0, (lesson.current_students || 0) + 1)
            };
          }
          return lesson;
        });
      });

      return { previousLessons, lessonId };
    },
    onSuccess: () => {
      notification.success('Вы успешно записались на занятие');
      // Инвалидируем все связанные кеши для полного обновления всех представлений
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      invalidateCreditBalance(queryClient);
    },
    onError: (err, lessonId, context) => {
      console.error('createBooking error:', err.message);
      notification.error(err.message || 'Ошибка записи на занятие');

      // Откатываем оптимистичные обновления при ошибке
      if (context?.previousLessons) {
        queryClient.setQueryData(['lessons'], context.previousLessons);
      }
    },
  });

  // Mutation для отмены записи
  const cancelBookingMutation = useMutation({
    mutationFn: async (bookingId) => {
      logger.debug('cancelBooking: Starting with id:', bookingId);
      // Ждем анимации CSS перед удалением из DOM
      await new Promise((resolve) => setTimeout(resolve, ANIMATION_DURATION));

      try {
        // Пытаемся отменить бронирование с retry для сетевых ошибок
        const result = await retryAsync(
          () => bookingsAPI.cancelBooking(bookingId),
          3,
          500
        );
        logger.debug('cancelBooking: Cancelled successfully');
        return result;
      } catch (error) {
        // Если ошибка "not active" - это значит бронирование уже отменено (возможно другим запросом)
        // Обрабатываем это как успех, чтобы оптимистичное обновление не откатывалось
        if (error.message && error.message.toLowerCase().includes('not active')) {
          logger.debug('cancelBooking: Booking already cancelled, treating as success');
          // Возвращаем special флаг для onSuccess/onError обработчика
          return { success: true, alreadyCancelled: true };
        }
        // Для других ошибок пробрасываем дальше
        throw error;
      }
    },
    onMutate: async (bookingId) => {
      // Отменяем любые выполняющиеся запросы
      await queryClient.cancelQueries({ queryKey: ['lessons'] });
      await queryClient.cancelQueries({ queryKey: ['myBookings'] });

      // Сохраняем предыдущее состояние myBookings для отката при ошибке
      const previousMyBookings = queryClient.getQueryData(['myBookings']);
      const myBookings = previousMyBookings || [];

      // Находим бронирование и его lesson_id из myBookings
      const booking = Array.isArray(myBookings) && myBookings.find(b => b.id === bookingId);

      if (!booking) {
        logger.warn('cancelBooking: booking not found in myBookings', bookingId);
        return { previousLessons: null, previousMyBookings, bookingFound: false };
      }

      const lessonId = booking.lesson_id;

      // Оптимистично удаляем бронирование из myBookings
      queryClient.setQueryData(['myBookings'], (old) => {
        if (!Array.isArray(old)) return old;
        return old.filter(b => b.id !== bookingId);
      });

      // Сохраняем предыдущее состояние lessons для отката при ошибке
      const previousLessons = queryClient.getQueryData(['lessons']);

      // Оптимистично обновляем lessons - уменьшаем current_students на 1 (место освобождается)
      queryClient.setQueryData(['lessons'], (old) => {
        if (!Array.isArray(old)) return old;
        return old.map((lesson) => {
          if (lesson.id === lessonId) {
            return {
              ...lesson,
              current_students: Math.max(0, (lesson.current_students || 0) - 1)
            };
          }
          return lesson;
        });
      });

      return { previousLessons, previousMyBookings, lessonId, bookingFound: true };
    },
    onSuccess: (result) => {
      // Проверяем был ли это случай "уже отменено"
      if (result?.alreadyCancelled) {
        notification.info('Это занятие уже было отменено');
      } else {
        notification.success('Запись отменена');
      }
      // Инвалидируем все связанные кеши для полного обновления всех представлений
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['cancelledLessons'] });
      invalidateCreditBalance(queryClient);
    },
    onError: (err, bookingId, context) => {
      console.error('cancelBooking error:', err.message);

      // Откатываем оптимистичные обновления при ошибке
      if (context?.bookingFound) {
        if (context.previousLessons) {
          queryClient.setQueryData(['lessons'], context.previousLessons);
        }
        if (context.previousMyBookings) {
          queryClient.setQueryData(['myBookings'], context.previousMyBookings);
        }
      }

      notification.error(err.message || 'Ошибка отмены записи');
    },
  });

  // Обертки для удобного использования
  const createBooking = useCallback(
    (lessonId) => createBookingMutation.mutateAsync(lessonId),
    [createBookingMutation]
  );

  const cancelBooking = useCallback(
    (bookingId) => cancelBookingMutation.mutateAsync(bookingId),
    [cancelBookingMutation]
  );

  const fetchMyBookingsCallback = useCallback(
    () => refetchMyBookings(),
    [refetchMyBookings]
  );

  const refresh = useCallback(() => {
    refetchQuery();
  }, [refetchQuery]);

  const error = queryError?.message || null;
  // pendingOperations - пустой объект, используется для совместимости с компонентами
  const pendingOperations = {};

  return {
    bookings,
    loading,
    error,
    hasError: !!error,
    myBookings,
    myBookingsLoading,
    myBookingsError,
    pendingOperations,
    fetchBookings: useCallback(() => refetchQuery(), [refetchQuery]),
    fetchMyBookings: fetchMyBookingsCallback,
    createBooking,
    cancelBooking,
    refresh,
    isCreating: createBookingMutation.isPending,
    isCancelling: cancelBookingMutation.isPending,
  };
};

export default useBookings;
