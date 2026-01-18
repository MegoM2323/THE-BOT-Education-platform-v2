import { useQuery } from '@tanstack/react-query';
import * as bookingsAPI from '../api/bookings.js';

/**
 * Hook для получения списка уроков, от которых студент отписался
 * Используется для блокировки повторной записи
 */
export const useCancelledLessons = () => {
  const {
    data: cancelledLessonIds = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['cancelledLessons'],
    queryFn: async () => {
      try {
        const response = await bookingsAPI.getCancelledLessons();
        return Array.isArray(response) ? response : [];
      } catch (err) {
        console.error('useCancelledLessons error:', err);
        // Return empty array on error to prevent UI breaking
        return [];
      }
    },
    staleTime: 60000, // 1 minute
    refetchOnWindowFocus: true,
  });

  /**
   * Check if a specific lesson was cancelled by this student
   * @param {string} lessonId - UUID of the lesson
   * @returns {boolean}
   */
  const isLessonCancelled = (lessonId) => {
    return cancelledLessonIds.includes(lessonId);
  };

  return {
    cancelledLessonIds,
    isLessonCancelled,
    isLoading,
    error,
    refetch,
  };
};

export default useCancelledLessons;
