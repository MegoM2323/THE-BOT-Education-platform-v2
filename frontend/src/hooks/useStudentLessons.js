import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import * as lessonsAPI from '../api/lessons.js';
import { useNotification } from './useNotification.js';

// Константа для пустого объекта фильтров - предотвращает пересоздание
const EMPTY_FILTERS = {};

/**
 * Hook for loading lessons visible to student
 * @param {string} weekStartDate - Start date of week (YYYY-MM-DD) - используется только для кэша, не для фильтрации
 * @param {Object} filters - Additional filters (teacher_id, day, time)
 */
export const useStudentLessons = (weekStartDate, filters) => {
  const notification = useNotification();

  // Используем стабильную ссылку на пустой объект если фильтры не переданы
  const filtersToUse = filters || EMPTY_FILTERS;

  // Мемоизировать фильтры чтобы избежать бесконечного цикла
  const filtersJson = JSON.stringify(filtersToUse);
  const memoizedFilters = useMemo(
    () => filtersToUse,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [filtersJson]
  );

  const queryKey = useMemo(
    () => ['studentLessons', weekStartDate, memoizedFilters],
    [weekStartDate, memoizedFilters]
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
        // НЕ передаем start_date в API запрос - студенты должны видеть ВСЕ занятия
        // включая прошедшие (completed lessons). Фильтрация по дате происходит
        // на уровне UI компонента (календарь, список)
        const response = await lessonsAPI.getLessons({
          ...memoizedFilters,
        });

        // Normalize response
        const lessonsArray = Array.isArray(response) ? response : (response?.lessons || []);

        // DEBUG: Логирование для проверки консистентности расписаний (T006)
        console.log(`[useStudentLessons] Loaded ${lessonsArray.length} lessons for student view`);

        // Возвращаем ВСЕ занятия (прошедшие и будущие)
        // Студент должен видеть свою полную историю забронированных занятий
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

