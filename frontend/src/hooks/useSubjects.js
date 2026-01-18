import { useCallback, useMemo } from 'react';
import { logger } from '../utils/logger.js';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as subjectsAPI from '../api/subjects.js';
import { useNotification } from './useNotification.js';

/**
 * Custom hook for managing subjects
 */
export const useSubjects = () => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  /**
   * Hook to fetch all available subjects
   */
  const useAllSubjects = () => {
    const {
      data: subjects = [],
      isLoading: loading,
      error: queryError,
      refetch,
    } = useQuery({
      queryKey: ['subjects', 'all'],
      queryFn: async () => {
        logger.debug('useQuery: Fetching all subjects');
        try {
          const data = await subjectsAPI.fetchAllSubjects();
          logger.debug('useQuery: Received subjects:', data);

          // Normalize data - handle both array and object with data field
          const subjectsArray = Array.isArray(data) ? data : (data?.subjects || data?.data || []);

          if (!Array.isArray(subjectsArray)) {
            logger.warn('useAllSubjects: Invalid data structure', data);
            return [];
          }

          logger.debug('useQuery: Successfully loaded', subjectsArray.length, 'subjects');
          return subjectsArray;
        } catch (err) {
          console.error('useQuery error:', err.message);
          notification.error('Ошибка загрузки предметов');
          throw err;
        }
      },
      staleTime: 300000, // 5 minutes - subjects don't change often
    });

    const error = queryError?.message || null;

    return {
      subjects,
      loading,
      error,
      hasError: !!error,
      refetch,
    };
  };

  /**
   * Hook to fetch subjects for a specific teacher
   * @param {string} teacherId - Teacher's UUID
   */
  const useTeacherSubjects = (teacherId) => {
    const {
      data: subjects = [],
      isLoading: loading,
      error: queryError,
      refetch,
    } = useQuery({
      queryKey: ['subjects', 'teacher', teacherId],
      queryFn: async () => {
        if (!teacherId) {
          logger.warn('useTeacherSubjects: No teacherId provided');
          return [];
        }

        logger.debug('useQuery: Fetching subjects for teacher:', teacherId);
        try {
          const data = await subjectsAPI.fetchTeacherSubjects(teacherId);
          logger.debug('useQuery: Received teacher subjects:', data);

          // Normalize data
          const subjectsArray = Array.isArray(data) ? data : (data?.subjects || data?.data || []);

          if (!Array.isArray(subjectsArray)) {
            logger.warn('useTeacherSubjects: Invalid data structure', data);
            return [];
          }

          logger.debug('useQuery: Successfully loaded', subjectsArray.length, 'subjects');
          return subjectsArray;
        } catch (err) {
          console.error('useQuery error:', err.message);
          notification.error('Ошибка загрузки предметов преподавателя');
          throw err;
        }
      },
      enabled: !!teacherId,
      staleTime: 60000, // 1 minute
    });

    const error = queryError?.message || null;

    return {
      subjects,
      loading,
      error,
      hasError: !!error,
      refetch,
    };
  };

  /**
   * Hook to fetch current user's subjects (for logged-in teacher)
   */
  const useMySubjects = () => {
    const {
      data: subjects = [],
      isLoading: loading,
      error: queryError,
      refetch,
    } = useQuery({
      queryKey: ['subjects', 'my'],
      queryFn: async () => {
        logger.debug('useQuery: Fetching current user subjects');
        try {
          const data = await subjectsAPI.getCurrentUserSubjects();
          logger.debug('useQuery: Received my subjects:', data);

          // Normalize data
          const subjectsArray = Array.isArray(data) ? data : (data?.subjects || data?.data || []);

          if (!Array.isArray(subjectsArray)) {
            logger.warn('useMySubjects: Invalid data structure', data);
            return [];
          }

          logger.debug('useQuery: Successfully loaded', subjectsArray.length, 'subjects');
          return subjectsArray;
        } catch (err) {
          console.error('useQuery error:', err.message);
          notification.error('Ошибка загрузки ваших предметов');
          throw err;
        }
      },
      staleTime: 60000, // 1 minute
    });

    const error = queryError?.message || null;

    return {
      subjects,
      loading,
      error,
      hasError: !!error,
      refetch,
    };
  };

  /**
   * Mutation hook for assigning a subject to a teacher
   */
  const useAssignSubject = () => {
    const assignMutation = useMutation({
      mutationFn: async ({ teacherId, subjectId }) => {
        logger.debug('assignSubject: Starting with teacherId:', teacherId, 'subjectId:', subjectId);
        const result = await subjectsAPI.assignSubjectToTeacher(teacherId, subjectId);
        logger.debug('assignSubject: Assigned successfully:', result);
        return result;
      },
      onSuccess: (data, variables) => {
        notification.success('Предмет успешно назначен');
        // Invalidate teacher subjects and all subjects cache
        queryClient.invalidateQueries({ queryKey: ['subjects', 'teacher', variables.teacherId] });
        queryClient.invalidateQueries({ queryKey: ['subjects', 'my'] });
      },
      onError: (err) => {
        console.error('assignSubject error:', err.message);
        notification.error(err.message || 'Ошибка назначения предмета');
      },
    });

    const assign = useCallback(
      (teacherId, subjectId) => assignMutation.mutateAsync({ teacherId, subjectId }),
      [assignMutation]
    );

    return {
      assign,
      isAssigning: assignMutation.isPending,
      assignError: assignMutation.error?.message || null,
    };
  };

  /**
   * Mutation hook for removing a subject from a teacher
   */
  const useRemoveSubject = () => {
    const removeMutation = useMutation({
      mutationFn: async ({ teacherId, subjectId }) => {
        logger.debug('removeSubject: Starting with teacherId:', teacherId, 'subjectId:', subjectId);
        await subjectsAPI.removeSubjectFromTeacher(teacherId, subjectId);
        logger.debug('removeSubject: Removed successfully');
      },
      onSuccess: (data, variables) => {
        notification.success('Предмет успешно удален');
        // Invalidate teacher subjects and all subjects cache
        queryClient.invalidateQueries({ queryKey: ['subjects', 'teacher', variables.teacherId] });
        queryClient.invalidateQueries({ queryKey: ['subjects', 'my'] });
      },
      onError: (err) => {
        console.error('removeSubject error:', err.message);
        notification.error(err.message || 'Ошибка удаления предмета');
      },
    });

    const remove = useCallback(
      (teacherId, subjectId) => removeMutation.mutateAsync({ teacherId, subjectId }),
      [removeMutation]
    );

    return {
      remove,
      isRemoving: removeMutation.isPending,
      removeError: removeMutation.error?.message || null,
    };
  };

  return {
    useAllSubjects,
    useTeacherSubjects,
    useMySubjects,
    useAssignSubject,
    useRemoveSubject,
  };
};

export default useSubjects;
