import { useMutation, useQueryClient } from '@tanstack/react-query';
import * as bookingsAPI from '../api/bookings.js';
import { useNotification } from './useNotification.js';
import { invalidateCreditBalance } from './useCredits.js';

/**
 * Hook for booking a lesson
 */
export const useBookLesson = () => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (lessonId) => {
      try {
        const response = await bookingsAPI.createBooking(lessonId);
        return response;
      } catch (err) {
        console.error('Book lesson error:', err);
        throw err;
      }
    },
    onMutate: async (lessonId) => {
      // Cancel outgoing queries
      await queryClient.cancelQueries({ queryKey: ['studentLessons'] });
      await queryClient.cancelQueries({ queryKey: ['myBookings'] });

      // Snapshot previous values
      const previousLessons = queryClient.getQueryData(['studentLessons']);

      // Optimistically update lessons (increment current_students)
      queryClient.setQueriesData({ queryKey: ['studentLessons'] }, (old) => {
        if (!Array.isArray(old)) return old;
        return old.map(lesson => {
          if (lesson.id === lessonId) {
            return {
              ...lesson,
              current_students: (lesson.current_students || 0) + 1,
            };
          }
          return lesson;
        });
      });

      return { previousLessons };
    },
    onSuccess: () => {
      notification.success('Занятие успешно забронировано');
      // Invalidate queries to refetch fresh data для синхронизации всех представлений
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myBookings'] });
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      invalidateCreditBalance(queryClient);
    },
    onError: (err, lessonId, context) => {
      // Rollback optimistic update
      if (context?.previousLessons) {
        queryClient.setQueriesData({ queryKey: ['studentLessons'] }, context.previousLessons);
      }

      const errorMessage = err.message || 'Ошибка при бронировании';
      notification.error(errorMessage);
    },
  });

  return mutation;
};

export default useBookLesson;
