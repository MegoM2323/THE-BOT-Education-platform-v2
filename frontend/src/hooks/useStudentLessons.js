import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import * as lessonsAPI from '../api/lessons.js';
import { useNotification } from './useNotification.js';

/**
 * Hook for loading lessons visible to student
 * Загружает ТОЛЬКО те занятия, на которые студент записан (через активные bookings)
 * @returns {Object} { lessons, isLoading, error, refetch }
 */
export const useStudentLessons = () => {
  const notification = useNotification();

  const queryKey = useMemo(
    () => ['studentLessons'],
    []
  );

  const {
    data: lessons = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey,
    queryFn: async () => {
      try {
        // FIX: Используем getMyLessons для студентов - возвращает ТОЛЬКО те занятия,
        // на которые студент записан (через активные bookings).
        // Раньше использовали getLessons, который показывал ВСЕ групповые занятия,
        // что было некорректно.
        const response = await lessonsAPI.getMyLessons();

        // Normalize response
        const lessonsArray = Array.isArray(response) ? response : (response?.lessons || []);

        // DEBUG: Логирование для проверки консистентности расписаний (T006)
        console.log(`[useStudentLessons] Loaded ${lessonsArray.length} lessons for student view`);

        // Возвращаем ВСЕ занятия (прошедшие и будущие)
        // Студент видит свою полную историю забронированных занятий
        return lessonsArray;
      } catch (err) {
        // Не показываем generic ошибку при 401 (обрабатывается глобально) или отмене запроса
        if (err.status !== 401 && err.name !== 'AbortError') {
          console.error('useStudentLessons error:', err);
          // Показываем ошибку только при реальных проблемах с сервером
          const status = err.status || err.response?.status || 0;
          if (status >= 500 || status === 0) {
            // status === 0 означает сетевую ошибку или другую проблему
            const errorMessage = err.message || 'Ошибка сервера. Попробуйте позже.';
            notification.error(errorMessage);
          }
        }
        // Создаем объект ошибки с правильным сообщением для отображения
        const errorToThrow = err instanceof Error 
          ? err 
          : new Error(err?.message || 'Failed to retrieve lessons');
        if (err.status) {
          errorToThrow.status = err.status;
        }
        throw errorToThrow;
      }
    },
    staleTime: 60000, // 1 minute
    refetchOnWindowFocus: true,
    retry: (failureCount, error) => {
      // Не повторяем при 401/403 - это проблема авторизации
      if (error?.status === 401 || error?.status === 403) {
        return false;
      }
      return failureCount < 2;
    },
  });

  return {
    lessons,
    isLoading,
    error,
    refetch,
  };
};

export default useStudentLessons;

