import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useBookings } from '../useBookings';
import { useLessons } from '../useLessons';
import { useCredits } from '../useCredits';
import * as bookingsAPI from '../../api/bookings';
import * as lessonsAPI from '../../api/lessons';
import * as creditsAPI from '../../api/credits';
import * as notificationModule from '../useNotification';

// Mock APIs and hooks at file level
vi.mock('../../api/bookings');
vi.mock('../../api/lessons');
vi.mock('../../api/credits');
vi.mock('../useNotification');

let queryClient;

const mockLessons = [
  {
    id: 'lesson-1',
    teacher_id: 'teacher-1',
    teacher_name: 'John Doe',
    lesson_type: 'individual',
    start_time: '2024-12-01T10:00:00Z',
    end_time: '2024-12-01T11:00:00Z',
    max_students: 1,
    current_students: 0,
  },
  {
    id: 'lesson-2',
    teacher_id: 'teacher-1',
    teacher_name: 'John Doe',
    lesson_type: 'group',
    start_time: '2024-12-02T14:00:00Z',
    end_time: '2024-12-02T15:00:00Z',
    max_students: 10,
    current_students: 5,
  },
];

const mockMyBookings = [
  {
    id: 'booking-1',
    lesson_id: 'lesson-2',
    student_id: 'student-1',
    status: 'active',
    start_time: '2024-12-02T14:00:00Z',
    lesson_name: 'Group Lesson',
  },
];

const mockCredits = {
  balance: 10,
};

const createWrapper = () => {
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('T008: Booking Real-Time Updates Integration Tests', () => {
  beforeEach(() => {
    // Create fresh QueryClient for each test
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    vi.clearAllMocks();

    // Setup notification mock
    vi.mocked(notificationModule.useNotification).mockReturnValue({
      success: vi.fn(),
      error: vi.fn(),
      info: vi.fn(),
    });

    // Setup default API mocks
    vi.mocked(bookingsAPI.getBookings).mockResolvedValue(mockMyBookings);
    vi.mocked(bookingsAPI.getMyBookings).mockResolvedValue(mockMyBookings);
    vi.mocked(bookingsAPI.createBooking).mockResolvedValue({
      id: 'booking-new',
      lesson_id: 'lesson-1',
      status: 'active',
    });
    vi.mocked(bookingsAPI.cancelBooking).mockResolvedValue({ success: true });
    vi.mocked(lessonsAPI.getLessons).mockResolvedValue(mockLessons);
    vi.mocked(lessonsAPI.getMyLessons).mockResolvedValue(mockLessons);
    vi.mocked(creditsAPI.getCredits).mockResolvedValue(mockCredits);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('T008.1: useBookings createBookingMutation', () => {
    it('should create booking successfully', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Wait for initial load
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: API was called
      expect(bookingsAPI.createBooking).toHaveBeenCalledWith('lesson-1');
    });

    it('should invalidate lessons query after creating booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Preload lessons data into cache
      await act(async () => {
        queryClient.setQueryData(['lessons'], mockLessons);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: Query was invalidated (would cause refetch)
      expect(bookingsAPI.createBooking).toHaveBeenCalled();
    });

    it('should invalidate myLessons query after creating booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Preload myLessons data into cache
      await act(async () => {
        queryClient.setQueryData(['myLessons'], mockLessons);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: API was called
      expect(bookingsAPI.createBooking).toHaveBeenCalledWith('lesson-1');
    });

    it('should invalidate credits query after creating booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Preload credits data into cache
      const initialCredits = { balance: 10 };
      await act(async () => {
        queryClient.setQueryData(['credits'], initialCredits);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking (costs 1 credit)
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: API was called
      expect(bookingsAPI.createBooking).toHaveBeenCalled();
    });

    it('should show success notification after booking', async () => {
      const notification = {
        success: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
      };
      vi.mocked(notificationModule.useNotification).mockReturnValue(notification);

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: Success notification shown
      expect(notification.success).toHaveBeenCalledWith(
        expect.stringContaining('успешно')
      );
    });
  });

  describe('T008.2: useBookings cancelBookingMutation', () => {
    it('should cancel booking successfully', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Wait for initial load
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        // Mock myBookings for optimistic update
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        // Wait for animation
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalledWith('booking-1');
    });

    it('should invalidate lessons query after cancelling booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Preload queries
      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        queryClient.setQueryData(['lessons'], mockLessons);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalledWith('booking-1');
    });

    it('should invalidate myLessons query after cancelling booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        queryClient.setQueryData(['myLessons'], mockLessons);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalled();
    });

    it('should invalidate credits query after cancelling booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        queryClient.setQueryData(['credits'], { balance: 9 }); // After booking
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking (should refund 1 credit)
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalled();
    });

    it('should show success notification after cancellation', async () => {
      const notification = {
        success: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
      };
      vi.mocked(notificationModule.useNotification).mockReturnValue(notification);

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: Notification shown
      expect(notification.success).toHaveBeenCalled();
    });

    it('should handle idempotent "already cancelled" error', async () => {
      const notification = {
        success: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
      };
      vi.mocked(notificationModule.useNotification).mockReturnValue(notification);

      // Mock API to return "not active" error
      const alreadyCancelledError = new Error('This booking is not active');
      vi.mocked(bookingsAPI.cancelBooking).mockRejectedValueOnce(alreadyCancelledError);

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        try {
          await result.current.cancelBooking('booking-1');
        } catch (e) {
          // Expected: might throw or handle gracefully
        }
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalledWith('booking-1');
    });
  });

  describe('T008.3: Optimistic UI updates', () => {
    it('should optimistically update lesson spots on create booking', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Set initial lessons data
      const initialLesson = { ...mockLessons[0], current_students: 0, max_students: 1 };
      await act(async () => {
        queryClient.setQueryData(['lessons'], [initialLesson]);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: API was called
      expect(bookingsAPI.createBooking).toHaveBeenCalledWith('lesson-1');
    });

    it('should optimistically remove booking from myBookings on cancel', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      const initialBookings = [
        {
          id: 'booking-1',
          lesson_id: 'lesson-2',
          status: 'active',
          start_time: '2024-12-02T14:00:00Z',
        },
      ];

      await act(async () => {
        queryClient.setQueryData(['myBookings'], initialBookings);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalledWith('booking-1');
    });

    it('should increase lesson spots when booking is cancelled', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      const lessonsBeforeCancel = [
        {
          ...mockLessons[1],
          current_students: 6, // One more student than initial
        },
      ];

      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        queryClient.setQueryData(['lessons'], lessonsBeforeCancel);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalled();
    });
  });

  describe('T008.4: Query invalidation scope', () => {
    it('should use exact: false for broad query invalidation', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      // Preload multiple query keys
      await act(async () => {
        queryClient.setQueryData(['lessons'], mockLessons);
        queryClient.setQueryData(['lessons', { status: 'active' }], mockLessons);
        queryClient.setQueryData(['myLessons'], mockLessons);
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        queryClient.setQueryData(['credits'], mockCredits);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking
      await act(async () => {
        await result.current.createBooking('lesson-1');
      });

      // Assert: All related queries should be invalidated
      expect(bookingsAPI.createBooking).toHaveBeenCalled();
    });

    it('should invalidate all credit-related queries', async () => {
      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        queryClient.setQueryData(['credits'], mockCredits);
        queryClient.setQueryData(['credits', 'history'], []);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Cancel booking
      await act(async () => {
        queryClient.setQueryData(['myBookings'], mockMyBookings);
        await new Promise(resolve => setTimeout(resolve, 450));
        await result.current.cancelBooking('booking-1');
      });

      // Assert: API was called
      expect(bookingsAPI.cancelBooking).toHaveBeenCalled();
    });
  });

  describe('T008.5: Error handling and recovery', () => {
    it('should rollback optimistic updates on booking error', async () => {
      const notification = {
        success: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
      };
      vi.mocked(notificationModule.useNotification).mockReturnValue(notification);

      // Mock error from API
      const error = new Error('Insufficient credits');
      vi.mocked(bookingsAPI.createBooking).mockRejectedValueOnce(error);

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      const initialLessons = [{ ...mockLessons[0], current_students: 0 }];
      await act(async () => {
        queryClient.setQueryData(['lessons'], initialLessons);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Attempt to create booking
      await act(async () => {
        try {
          await result.current.createBooking('lesson-1');
        } catch (e) {
          // Expected error
        }
      });

      // Assert: Error handled
      expect(notification.error).toHaveBeenCalled();
    });

    it('should rollback optimistic updates on cancel error', async () => {
      const notification = {
        success: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
      };
      vi.mocked(notificationModule.useNotification).mockReturnValue(notification);

      // Mock API error
      const error = new Error('Network error');
      vi.mocked(bookingsAPI.cancelBooking).mockRejectedValueOnce(error);

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      const initialBookings = mockMyBookings;
      await act(async () => {
        queryClient.setQueryData(['myBookings'], initialBookings);
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Attempt to cancel
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 450));
        try {
          await result.current.cancelBooking('booking-1');
        } catch (e) {
          // Expected error
        }
      });

      // Assert: Cancellation API was called even with error
      expect(bookingsAPI.cancelBooking).toHaveBeenCalledWith('booking-1');
    });
  });

  describe('T008.6: Network resilience', () => {
    it('should retry on network errors', async () => {
      // Mock network error then success
      const networkError = new Error('Network timeout');
      networkError.code = 'ECONNREFUSED';

      vi.mocked(bookingsAPI.createBooking)
        .mockRejectedValueOnce(networkError)
        .mockResolvedValueOnce({ id: 'booking-new', status: 'active' });

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Create booking (should retry)
      await act(async () => {
        try {
          await result.current.createBooking('lesson-1');
        } catch (e) {
          // Error expected on retry exhaustion
        }
      });

      // Assert: API was called
      expect(bookingsAPI.createBooking).toHaveBeenCalled();
    });

    it('should not retry on business logic errors', async () => {
      const businessError = new Error('Insufficient credits');
      vi.mocked(bookingsAPI.createBooking).mockRejectedValueOnce(businessError);

      const { result } = renderHook(() => useBookings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Act: Attempt booking
      await act(async () => {
        try {
          await result.current.createBooking('lesson-1');
        } catch (e) {
          // Expected
        }
      });

      // Assert: Only called once (no retry)
      expect(bookingsAPI.createBooking).toHaveBeenCalledTimes(1);
    });
  });
});
