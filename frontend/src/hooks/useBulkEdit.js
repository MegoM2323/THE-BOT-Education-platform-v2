/**
 * Custom hook for bulk edit operations on lessons
 * Handles "Apply to all subsequent" functionality
 */

import { useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '../api/client.js';
import { invalidateCreditBalance } from './useCredits.js';

/**
 * Apply modification to all subsequent lessons matching the pattern
 * @returns {Object} Mutation object with mutateAsync, isLoading, error
 */
export const useApplyToAllSubsequent = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ lessonId, modificationType, ...params }) => {
      console.log('[useBulkEdit] Starting mutation with:', { lessonId, modificationType, params });

      const payload = {
        lesson_id: lessonId,
        modification_type: modificationType,
      };

      // Add modification-specific parameters
      if (modificationType === 'add_student' || modificationType === 'remove_student') {
        payload.student_id = params.studentId;
      }

      if (modificationType === 'change_teacher') {
        payload.teacher_id = params.teacherId;
      }

      if (modificationType === 'change_time') {
        payload.new_start_time = params.newStartTime;
      }

      if (modificationType === 'change_capacity') {
        payload.new_max_students = params.newMaxStudents;
      }

      console.log('[useBulkEdit] Payload:', payload);

      const response = await apiClient.post(
        `/lessons/${lessonId}/apply-to-all`,
        payload
      );

      console.log('[useBulkEdit] Response:', response);

      return response;
    },
    onSuccess: () => {
      // Invalidate all lesson-related queries to refresh calendar
      // exact: false для инвалидации всех вложенных ключей (например ['lessons', filters])
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
      invalidateCreditBalance(queryClient);
    },
  });
};

/**
 * Get preview of affected lessons before applying modification
 * @param {string} lessonId - ID of the lesson to use as pattern
 * @param {string} modificationType - Type of modification
 * @returns {Object} Query object with data, isLoading, error
 */
export const useGetAffectedLessons = (_lessonId, _modificationType) => {
  return {
    // For now, we don't have a preview endpoint
    // The modal will show general information about the modification
    // Backend will return affected_lessons_count after applying
    data: null,
    isLoading: false,
    error: null,
  };
};

export default {
  useApplyToAllSubsequent,
  useGetAffectedLessons,
};
