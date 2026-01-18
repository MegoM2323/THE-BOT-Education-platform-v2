/**
 * Centralized utility for React Query cache invalidation
 * Ensures consistent cache invalidation patterns across the application
 */

/**
 * Invalidate lesson-related queries
 * @param {QueryClient} queryClient - React Query client instance
 */
export const invalidateLessonData = (queryClient) => {
  queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false });
};

/**
 * Invalidate booking-related queries
 * @param {QueryClient} queryClient - React Query client instance
 */
export const invalidateBookingData = (queryClient) => {
  queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['myBookings'], exact: false });
};

/**
 * Invalidate credit-related queries
 * @param {QueryClient} queryClient - React Query client instance
 */
export const invalidateCreditData = (queryClient) => {
  // Инвалидируем только баланс кредитов, не историю
  // Ключ ['credits'] используется в useCredits.js для баланса
  queryClient.invalidateQueries({ queryKey: ['credits'], exact: true });
};

/**
 * Invalidate user/student-related queries
 * @param {QueryClient} queryClient - React Query client instance
 */
export const invalidateUserData = (queryClient) => {
  queryClient.invalidateQueries({ queryKey: ['users'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['students'], exact: false });
  queryClient.invalidateQueries({ queryKey: ['teachers'], exact: false });
};

/**
 * Invalidate all lesson and booking related data
 * Used when changes affect multiple query caches
 * @param {QueryClient} queryClient - React Query client instance
 */
export const invalidateAllLessonAndBookingData = (queryClient) => {
  invalidateLessonData(queryClient);
  invalidateBookingData(queryClient);
  invalidateCreditData(queryClient);
};

/**
 * Invalidate all data related to user operations
 * Used after creating/updating/deleting users
 * @param {QueryClient} queryClient - React Query client instance
 */
export const invalidateAllUserRelatedData = (queryClient) => {
  invalidateUserData(queryClient);
  invalidateCreditData(queryClient);
};
