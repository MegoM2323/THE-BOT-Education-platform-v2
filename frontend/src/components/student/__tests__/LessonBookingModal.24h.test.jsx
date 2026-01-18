import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LessonBookingModal } from '../LessonBookingModal';
import { vi } from 'vitest';

// Mock hooks
vi.mock('../../../hooks/useBookLesson.js');
vi.mock('../../../hooks/useMyBookings.js');
vi.mock('../../../hooks/useCancelledLessons.js');

import { useBookLesson } from '../../../hooks/useBookLesson.js';
import { useMyBookings } from '../../../hooks/useMyBookings.js';
import { useCancelledLessons } from '../../../hooks/useCancelledLessons.js';

describe('LessonBookingModal - 24-hour Restriction Tests', () => {
  let queryClient;

  const mockLesson = {
    id: 1,
    subject: 'Mathematics',
    teacher_name: 'John Doe',
    start_time: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(), // 2 hours from now
    end_time: new Date(Date.now() + 3 * 60 * 60 * 1000).toISOString(),
    max_students: 5,
    current_students: 2,
    credits_cost: 1,
    color: '#004231'
  };

  const mockCredits = {
    balance: 10
  };

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false }
      }
    });

    useBookLesson.mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isLoading: false
    });

    useMyBookings.mockReturnValue({
      cancelBooking: vi.fn().mockResolvedValue({})
    });

    useCancelledLessons.mockReturnValue({
      isLessonCancelled: vi.fn().mockReturnValue(false)
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const renderModal = (lesson = mockLesson, myBookings = [], credits = mockCredits) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={lesson}
          myBookings={myBookings}
          credits={credits}
        />
      </QueryClientProvider>
    );
  };

  test('allows booking lesson less than 24 hours before start', () => {
    // Lesson in 2 hours (< 24h)
    const lessonSoon = {
      ...mockLesson,
      start_time: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString()
    };

    renderModal(lessonSoon, [], mockCredits);

    // Book button should be ENABLED
    const bookButton = screen.getByTestId('book-lesson-button');
    expect(bookButton).toBeInTheDocument();
    expect(bookButton).not.toBeDisabled();

    // No 24-hour error message
    expect(screen.queryByText(/24 часа/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/менее чем за 24/i)).not.toBeInTheDocument();
  });

  test('allows booking lesson 1 hour before start', () => {
    // Lesson in 1 hour (< 24h, should still be allowed)
    const lessonVerySoon = {
      ...mockLesson,
      start_time: new Date(Date.now() + 1 * 60 * 60 * 1000).toISOString()
    };

    renderModal(lessonVerySoon, [], mockCredits);

    const bookButton = screen.getByTestId('book-lesson-button');
    expect(bookButton).not.toBeDisabled();
  });

  test('allows booking lesson 30 minutes before start', () => {
    // Lesson in 30 minutes (< 24h, edge case)
    const lessonImminent = {
      ...mockLesson,
      start_time: new Date(Date.now() + 30 * 60 * 1000).toISOString()
    };

    renderModal(lessonImminent, [], mockCredits);

    const bookButton = screen.getByTestId('book-lesson-button');
    expect(bookButton).not.toBeDisabled();
  });

  test('shows cancel button if already booked (not book button)', () => {
    const myBookings = [
      { lesson_id: 1, status: 'active', id: 100 }
    ];

    renderModal(mockLesson, myBookings, mockCredits);

    // Should NOT show book button (already booked)
    expect(screen.queryByTestId('book-lesson-button')).not.toBeInTheDocument();

    // Should show cancel button instead
    expect(screen.getByTestId('cancel-booking-button')).toBeInTheDocument();
  });

  test('blocks booking if insufficient credits (not 24h related)', () => {
    const lowCredits = { balance: 0 };

    renderModal(mockLesson, [], lowCredits);

    // Should show credit error, NOT 24h error
    expect(screen.getByText(/Недостаточно кредитов/i)).toBeInTheDocument();
    expect(screen.queryByText(/24 часа/i)).not.toBeInTheDocument();
  });

  test('blocks booking if lesson is full (not 24h related)', () => {
    const fullLesson = {
      ...mockLesson,
      current_students: 5,
      max_students: 5
    };

    renderModal(fullLesson, [], mockCredits);

    expect(screen.getByText(/Нет свободных мест/i)).toBeInTheDocument();
    expect(screen.queryByText(/24 часа/i)).not.toBeInTheDocument();
  });

  test('blocks cancellation less than 24 hours before start', () => {
    // Booked lesson in 2 hours
    const lessonSoon = {
      ...mockLesson,
      start_time: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString()
    };

    const myBookings = [
      { lesson_id: 1, status: 'active', id: 100 }
    ];

    renderModal(lessonSoon, myBookings, mockCredits);

    // Cancel button should exist but be DISABLED
    const cancelButton = screen.getByTestId('cancel-booking-button');
    expect(cancelButton).toBeInTheDocument();
    expect(cancelButton).toBeDisabled();

    // Should show 24-hour cancellation restriction message
    expect(screen.getByText(/Отмена возможна только за 24 часа/i)).toBeInTheDocument();
  });

  test('allows cancellation more than 24 hours before start', () => {
    // Booked lesson in 48 hours (> 24h)
    const lessonLater = {
      ...mockLesson,
      start_time: new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString()
    };

    const myBookings = [
      { lesson_id: 1, status: 'active', id: 100 }
    ];

    renderModal(lessonLater, myBookings, mockCredits);

    // Cancel button should be ENABLED
    const cancelButton = screen.getByTestId('cancel-booking-button');
    expect(cancelButton).toBeInTheDocument();
    expect(cancelButton).not.toBeDisabled();

    // No 24-hour restriction message
    expect(screen.queryByText(/Отмена возможна только за 24 часа/i)).not.toBeInTheDocument();
  });

  test('allows cancellation exactly at 24-hour boundary', () => {
    // Lesson in exactly 24 hours + 1 second (to ensure >= 24)
    const lessonAt24h = {
      ...mockLesson,
      start_time: new Date(Date.now() + (24 * 60 * 60 * 1000) + 1000).toISOString()
    };

    const myBookings = [
      { lesson_id: 1, status: 'active', id: 100 }
    ];

    renderModal(lessonAt24h, myBookings, mockCredits);

    // At >= 24h, should be allowed (condition is < 24)
    const cancelButton = screen.getByTestId('cancel-booking-button');
    expect(cancelButton).not.toBeDisabled();

    // No 24-hour restriction message
    expect(screen.queryByText(/Отмена возможна только за 24 часа/i)).not.toBeInTheDocument();
  });

  test('blocks cancellation 23.5 hours before start', () => {
    // Lesson in 23.5 hours (< 24h)
    const lessonAt23h = {
      ...mockLesson,
      start_time: new Date(Date.now() + 23.5 * 60 * 60 * 1000).toISOString()
    };

    const myBookings = [
      { lesson_id: 1, status: 'active', id: 100 }
    ];

    renderModal(lessonAt23h, myBookings, mockCredits);

    const cancelButton = screen.getByTestId('cancel-booking-button');
    expect(cancelButton).toBeDisabled();
    expect(screen.getByText(/Отмена возможна только за 24 часа/i)).toBeInTheDocument();
  });

  test('booking eligibility does not check time restriction', () => {
    // This test verifies that checkBookingEligibility function
    // does NOT have any 24-hour restriction logic

    // Test at various time intervals
    const timeIntervals = [
      30 * 60 * 1000,    // 30 minutes
      1 * 60 * 60 * 1000,  // 1 hour
      2 * 60 * 60 * 1000,  // 2 hours
      12 * 60 * 60 * 1000, // 12 hours
      23 * 60 * 60 * 1000, // 23 hours
      48 * 60 * 60 * 1000  // 48 hours
    ];

    timeIntervals.forEach((interval) => {
      const lesson = {
        ...mockLesson,
        start_time: new Date(Date.now() + interval).toISOString()
      };

      const { unmount } = renderModal(lesson, [], mockCredits);

      // All should show enabled book button (no time restriction)
      const bookButton = screen.getByTestId('book-lesson-button');
      expect(bookButton).not.toBeDisabled();

      // No 24-hour message for booking
      expect(screen.queryByText(/запись.*24 часа/i)).not.toBeInTheDocument();

      unmount();
    });
  });

  test('only valid booking restrictions apply', () => {
    // Lesson in 1 hour (< 24h)
    const lessonSoon = {
      ...mockLesson,
      start_time: new Date(Date.now() + 1 * 60 * 60 * 1000).toISOString()
    };

    renderModal(lessonSoon, [], mockCredits);

    // Book button enabled (no 24h restriction)
    const bookButton = screen.getByTestId('book-lesson-button');
    expect(bookButton).not.toBeDisabled();

    // Only these errors should be possible:
    // - "Вы уже записаны"
    // - "Нет свободных мест"
    // - "Недостаточно кредитов"
    // - "уже началось или прошло"
    // - "отписались от этого занятия"
    // NOT "24 часа"

    const validErrors = [
      /уже записаны/i,
      /свободных мест/i,
      /кредитов/i,
      /началось/i,
      /отписались/i
    ];

    const errorElement = screen.queryByRole('alert');
    if (errorElement) {
      const errorText = errorElement.textContent;
      expect(errorText).not.toMatch(/24 часа/i);
    }
  });
});
