import { useCallback, useMemo } from 'react';
import { logger } from '../utils/logger.js';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as usersAPI from '../api/users.js';
import { useNotification } from './useNotification.js';

// Константа для пустого объекта фильтров - предотвращает пересоздание
const EMPTY_FILTERS = {};

/**
 * Инвалидировать кеш студентов
 * @param {QueryClient} queryClient - React Query клиент
 */
export const invalidateStudentsCache = (queryClient) => {
  queryClient.invalidateQueries({ queryKey: ['students'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['users'], exact: false });
};

/**
 * Кастомный хук для управления студентами
 * Поддерживает фильтрацию, поиск и пагинацию
 *
 * @param {Object} filters - Опциональные фильтры
 * @param {string} filters.search - Поиск по имени или email
 * @param {boolean} filters.active - Фильтр по активности
 * @param {boolean} filters.enabled - Включить/выключить автоматическую загрузку (default: true)
 * @returns {Object} Состояние и методы для работы со студентами
 */
export const useStudents = (filters) => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Используем стабильную ссылку на пустой объект если фильтры не переданы
  const filtersToUse = filters || EMPTY_FILTERS;

  // Мемоизировать фильтры чтобы избежать бесконечного цикла
  const memoizedFilters = useMemo(
    () => ({
      ...filtersToUse,
      role: 'student', // Всегда фильтруем по роли student
    }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(filtersToUse)]
  );

  // Уникальный ключ для кеша, основанный на фильтрах
  const queryKey = useMemo(
    () => ['students', memoizedFilters],
    [memoizedFilters]
  );

  /**
   * Нормализация данных студентов
   * Backend может возвращать данные в разных форматах
   */
  const normalizeStudents = useCallback((data) => {
    // Обработка разных форматов ответа
    let studentsArray;

    if (Array.isArray(data)) {
      // Данные уже массив
      studentsArray = data;
    } else if (data?.users && Array.isArray(data.users)) {
      // Формат { users: [...], meta: {...} }
      studentsArray = data.users;
    } else if (data?.data?.users && Array.isArray(data.data.users)) {
      // Формат { data: { users: [...] } }
      studentsArray = data.data.users;
    } else {
      logger.warn('normalizeStudents: Unexpected data format', data);
      studentsArray = [];
    }

    // Валидация и нормализация каждого студента
    return studentsArray
      .filter((student) => {
        if (!student || typeof student !== 'object') {
          logger.warn('normalizeStudents: Invalid student object', student);
          return false;
        }
        if (!student.id) {
          logger.warn('normalizeStudents: Student missing id', student);
          return false;
        }
        return true;
      })
      .map((student) => ({
        id: student.id,
        email: student.email || '',
        full_name: student.full_name || student.name || 'Unknown Student',
        role: student.role || 'student',
        is_active: student.is_active !== false,
        telegram_username: student.telegram_username || null,
        created_at: student.created_at || null,
        updated_at: student.updated_at || null,
        // Дополнительные поля для совместимости
        credits: student.credits ?? student.credit_balance ?? 0,
      }));
  }, []);

  // useQuery для получения студентов с фильтрами
  const {
    data: students = [],
    isLoading: loading,
    error: queryError,
    refetch: refetchQuery,
    isFetching,
  } = useQuery({
    queryKey,
    queryFn: async () => {
      logger.debug('useStudents: Fetching students with filters:', memoizedFilters);
      try {
        const data = await usersAPI.getStudentsAll();
        logger.debug('useStudents: Raw data received:', data);
        const normalized = normalizeStudents(data);
        logger.debug('useStudents: Successfully loaded', normalized.length, 'students');
        return normalized;
      } catch (err) {
        console.error('useStudents query error:', err.message);
        // Не показываем ошибку если это 403 (недостаточно прав)
        if (err.status !== 403) {
          notification.error('Ошибка загрузки списка студентов');
        }
        throw err;
      }
    },
    enabled: filtersToUse.enabled !== false,
    staleTime: 30000, // 30 секунд
    cacheTime: 5 * 60 * 1000, // 5 минут
    retry: (failureCount, error) => {
      // Не повторять для 403/401 ошибок
      if (error?.status === 403 || error?.status === 401) {
        return false;
      }
      return failureCount < 2;
    },
  });

  // Локальная фильтрация по поиску (клиентская)
  const filteredStudents = useMemo(() => {
    if (!filtersToUse.search) {
      return students;
    }

    const searchLower = filtersToUse.search.toLowerCase().trim();
    if (!searchLower) {
      return students;
    }

    return students.filter((student) => {
      const name = (student.full_name || '').toLowerCase();
      const email = (student.email || '').toLowerCase();
      return name.includes(searchLower) || email.includes(searchLower);
    });
  }, [students, filtersToUse.search]);

  // Mutation для обновления студента (если нужно)
  const updateStudentMutation = useMutation({
    mutationFn: async ({ studentId, updates }) => {
      logger.debug('updateStudent: Starting with id:', studentId, 'updates:', updates);
      const updated = await usersAPI.updateUser(studentId, updates);
      logger.debug('updateStudent: Updated successfully:', updated);
      return updated;
    },
    onSuccess: () => {
      notification.success('Студент обновлен');
      invalidateStudentsCache(queryClient);
    },
    onError: (err) => {
      console.error('updateStudent error:', err.message);
      notification.error(err.message || 'Ошибка обновления студента');
    },
  });

  // Обертки для удобного использования
  const updateStudent = useCallback(
    (studentId, updates) => updateStudentMutation.mutateAsync({ studentId, updates }),
    [updateStudentMutation]
  );

  const refresh = useCallback(() => {
    refetchQuery();
  }, [refetchQuery]);

  // Получить студента по ID из кеша
  const getStudentById = useCallback(
    (studentId) => {
      return students.find((s) => s.id === studentId) || null;
    },
    [students]
  );

  const error = queryError?.message || null;

  return {
    // Данные
    students: filteredStudents,
    allStudents: students, // Все студенты без клиентской фильтрации
    loading,
    isFetching,
    error,
    hasError: !!error,

    // Методы
    fetchStudents: refresh,
    refresh,
    updateStudent,
    getStudentById,

    // Состояния мутаций
    isUpdating: updateStudentMutation.isPending,

    // Мета-данные
    totalCount: students.length,
    filteredCount: filteredStudents.length,
  };
};

export default useStudents;
