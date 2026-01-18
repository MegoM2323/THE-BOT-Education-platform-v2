import { useCallback, useMemo } from 'react';
import { logger } from '../utils/logger.js';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as lessonsAPI from '../api/lessons.js';
import { useNotification } from './useNotification.js';
import { invalidateCreditBalance } from './useCredits.js';

// Константа для пустого объекта фильтров - предотвращает пересоздание
const EMPTY_FILTERS = {};

/**
 * Кастомный хук для управления занятиями
 */
export const useLessons = (filters) => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Используем стабильную ссылку на пустой объект если фильтры не переданы
  const filtersToUse = filters || EMPTY_FILTERS;

  // Мемоизировать фильтры чтобы избежать бесконечного цикла
  const filtersJson = JSON.stringify(filtersToUse);
  const memoizedFilters = useMemo(
    () => filtersToUse,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [filtersJson]
  );

  // Уникальный ключ для кеша, основанный на фильтрах
  const queryKey = useMemo(() => ['lessons', memoizedFilters], [memoizedFilters]);

  // Функция для нормализации данных
  const normalizeData = (data) => {
    const lessonsArray = Array.isArray(data) ? data : (data?.lessons || []);
    if (!Array.isArray(lessonsArray)) {
      logger.warn('normalizeData: Invalid data structure received', data);
      return [];
    }
    return lessonsArray;
  };

  // useQuery для получения занятий с фильтрами
  const {
    data: lessons = [],
    isLoading: loading,
    error: queryError,
    refetch: refetchQuery,
  } = useQuery({
    queryKey,
    queryFn: async () => {
      logger.debug('useQuery: Fetching lessons with filters:', memoizedFilters);
      try {
        const data = await lessonsAPI.getLessons(memoizedFilters);
        logger.debug('useQuery: Raw data received:', data);
        const normalized = normalizeData(data);
        logger.debug('useQuery: Successfully loaded', normalized.length, 'lessons');

        // DEBUG: Логирование для проверки консистентности расписаний (T006)
        console.log(`[useLessons] Loaded ${normalized.length} lessons for admin/teacher view with filters:`, memoizedFilters);

        return normalized;
      } catch (err) {
        console.error('useQuery error:', err.message);
        notification.error('Ошибка загрузки занятий');
        throw err;
      }
    },
  });

  // Отдельный query для "мои занятия" - загружается с правильным эндпоинтом для студентов
  const {
    data: myLessons = [],
    isLoading: myLessonsLoading,
    error: myLessonsError,
    refetch: refetchMyLessons,
  } = useQuery({
    queryKey: ['myLessons'],
    queryFn: async () => {
      logger.debug('useQuery: Fetching my lessons from /lessons/my endpoint');
      try {
        const data = await lessonsAPI.getMyLessons();
        logger.debug('useQuery: Raw data received from /lessons/my:', data);
        const normalized = normalizeData(data);
        logger.debug('useQuery: Successfully loaded', normalized.length, 'lessons');
        return normalized;
      } catch (err) {
        // Не показываем ошибку для 401 (неавторизован) - это нормальное состояние
        // Также не показываем generic ошибку - пустой список уже отображается корректно
        if (err.status !== 401) {
          console.error('useQuery error:', err.message);
        }
        throw err;
      }
    },
    enabled: true, // Автоматически загружаем при монтировании компонента
    staleTime: 30000, // 30 секунд
    retry: (failureCount, error) => {
      // Не повторяем запрос при 401 (неавторизован) или 403 (запрещено)
      if (error?.status === 401 || error?.status === 403) {
        return false;
      }
      return failureCount < 2;
    },
  });

  // Mutation для создания занятия
  const createLessonMutation = useMutation({
    mutationFn: async (lessonData) => {
      logger.debug('createLesson: Starting with data:', lessonData);
      const newLesson = await lessonsAPI.createLesson(lessonData);
      logger.debug('createLesson: Created successfully:', newLesson);
      return newLesson;
    },
    onSuccess: () => {
      notification.success('Занятие успешно создано');
      // Инвалидируем все связанные кеши для синхронизации всех представлений
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false });
    },
    onError: (err) => {
      console.error('createLesson error:', err.message);
      notification.error(err.message || 'Ошибка создания занятия');
    },
  });

  // Mutation для обновления занятия
  const updateLessonMutation = useMutation({
    mutationFn: async ({ lessonId, updates }) => {
      logger.debug('updateLesson: Starting with id:', lessonId, 'updates:', updates);
      const updated = await lessonsAPI.updateLesson(lessonId, updates);
      logger.debug('updateLesson: Updated successfully:', updated);
      return updated;
    },
    onSuccess: () => {
      notification.success('Занятие обновлено');
      // Инвалидируем все связанные кеши для синхронизации всех представлений
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false });
    },
    onError: (err) => {
      console.error('updateLesson error:', err.message);
      notification.error(err.message || 'Ошибка обновления занятия');
    },
  });

  // Mutation для удаления занятия
  const deleteLessonMutation = useMutation({
    mutationFn: async (lessonId) => {
      logger.debug('deleteLesson: Starting with id:', lessonId);
      await lessonsAPI.deleteLesson(lessonId);
      logger.debug('deleteLesson: Deleted successfully');
    },
    onSuccess: () => {
      notification.success('Занятие удалено');
      // Инвалидируем все связанные кеши при удалении (кредиты возвращаются студентам)
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['cancelledLessons'] });
      queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false });
      invalidateCreditBalance(queryClient);
    },
    onError: (err) => {
      console.error('deleteLesson error:', err.message);
      notification.error(err.message || 'Ошибка удаления занятия');
    },
  });

  // Обертки для удобного использования
  const createLesson = useCallback(
    (lessonData) => createLessonMutation.mutateAsync(lessonData),
    [createLessonMutation]
  );

  const updateLesson = useCallback(
    (lessonId, updates) => updateLessonMutation.mutateAsync({ lessonId, updates }),
    [updateLessonMutation]
  );

  const deleteLesson = useCallback(
    (lessonId) => deleteLessonMutation.mutateAsync(lessonId),
    [deleteLessonMutation]
  );

  const fetchMyLessons = useCallback(
    () => refetchMyLessons(),
    [refetchMyLessons]
  );

  const refresh = useCallback(() => {
    refetchQuery();
  }, [refetchQuery]);

  const error = queryError?.message || null;

  return {
    lessons,
    loading,
    error,
    hasError: !!error,
    myLessons,
    myLessonsLoading,
    myLessonsError,
    fetchLessons: useCallback(() => refetchQuery(), [refetchQuery]),
    fetchMyLessons,
    createLesson,
    updateLesson,
    deleteLesson,
    refresh,
    isCreating: createLessonMutation.isPending,
    isUpdating: updateLessonMutation.isPending,
    isDeleting: deleteLessonMutation.isPending,
  };
};

export default useLessons;
