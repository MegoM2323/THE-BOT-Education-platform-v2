import { useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as bookingsAPI from '../api/bookings.js';
import { useNotification } from './useNotification.js';
import { invalidateCreditBalance } from './useCredits.js';

/**
 * Hook for managing student's bookings
 */
export const useMyBookings = () => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Fetch all bookings
  const {
    data: myBookings = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['myBookings'],
    queryFn: async () => {
      try {
        const response = await bookingsAPI.getMyBookings();
        const bookingsArray = Array.isArray(response) ? response : (response?.bookings || []);
        return bookingsArray;
      } catch (err) {
        console.error('useMyBookings error:', err);
        throw err;
      }
    },
    staleTime: 30000, // 30 seconds
    refetchOnWindowFocus: true,
  });

  // Split bookings into upcoming and past
  const { upcomingBookings, pastBookings } = useMemo(() => {
    const now = new Date();
    const upcoming = [];
    const past = [];

    myBookings.forEach(booking => {
      const lessonTime = booking.lesson?.start_time || booking.start_time;
      if (!lessonTime) return;

      const lessonDate = new Date(lessonTime);

      // Only show active bookings in upcoming
      if (lessonDate > now && booking.status === 'active') {
        upcoming.push(booking);
      } else if (lessonDate <= now || booking.status === 'cancelled') {
        past.push(booking);
      }
    });

    // Sort: upcoming by start_time ASC, past by start_time DESC
    upcoming.sort((a, b) => {
      const timeA = new Date(a.lesson?.start_time || a.start_time);
      const timeB = new Date(b.lesson?.start_time || b.start_time);
      return timeA - timeB;
    });

    past.sort((a, b) => {
      const timeA = new Date(a.lesson?.start_time || a.start_time);
      const timeB = new Date(b.lesson?.start_time || b.start_time);
      return timeB - timeA;
    });

    return { upcomingBookings: upcoming, pastBookings: past };
  }, [myBookings]);

  // Cancel booking mutation
  const cancelMutation = useMutation({
    mutationFn: async (bookingId) => {
      try {
        const response = await bookingsAPI.cancelBooking(bookingId);
        return response;
      } catch (err) {
        console.error('Cancel booking error:', err);
        throw err;
      }
    },
    onMutate: async (bookingId) => {
      // Cancel outgoing queries
      await queryClient.cancelQueries({ queryKey: ['myBookings'] });

      // Snapshot previous value
      const previousBookings = queryClient.getQueryData(['myBookings']);

      // Optimistically update by removing booking
      queryClient.setQueryData(['myBookings'], (old) => {
        if (!Array.isArray(old)) return old;
        return old.filter(b => b.id !== bookingId);
      });

      return { previousBookings };
    },
    onSuccess: () => {
      notification.success('Бронирование отменено');
      // Invalidate related queries для синхронизации всех представлений
      queryClient.invalidateQueries({ queryKey: ['myBookings'] });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['cancelledLessons'] });
      invalidateCreditBalance(queryClient);
    },
    onError: (err, bookingId, context) => {
      // Rollback on error
      if (context?.previousBookings) {
        queryClient.setQueryData(['myBookings'], context.previousBookings);
      }

      const errorMessage = err.message || 'Failed to cancel booking';
      notification.error(errorMessage);
    },
  });

  const cancelBooking = useCallback(
    (bookingId) => cancelMutation.mutateAsync(bookingId),
    [cancelMutation]
  );

  return {
    myBookings,
    upcomingBookings,
    pastBookings,
    isLoading,
    error,
    refetch,
    cancelBooking,
    isCancelling: cancelMutation.isPending,
  };
};

export default useMyBookings;
