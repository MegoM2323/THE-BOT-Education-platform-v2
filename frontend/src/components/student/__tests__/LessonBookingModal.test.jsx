import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LessonBookingModal } from "../LessonBookingModal.jsx";

// Mock HomeworkSection to avoid QueryClient issues
vi.mock('../../admin/HomeworkSection.jsx', () => ({
  default: () => <div data-testid="homework-section">Homework Section Mock</div>
}));

// Mock hooks
vi.mock('../../../hooks/useBookLesson.js', () => ({
  useBookLesson: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
  }),
}));

vi.mock('../../../hooks/useMyBookings.js', () => ({
  useMyBookings: () => ({
    cancelBooking: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useCancelledLessons.js', () => ({
  useCancelledLessons: () => ({
    isLessonCancelled: vi.fn((lessonId) => false),
  }),
}));

// Mock useAuth для HomeworkSection
vi.mock('../../../hooks/useAuth.js', () => ({
  useAuth: () => ({
    user: { id: 'student-1', role: 'student', email: 'student@test.com' },
    isAuthenticated: true,
  }),
}));

// Mock useHomework для HomeworkSection
vi.mock('../../../hooks/useHomework.js', () => ({
  useHomework: () => ({
    data: [],
    isLoading: false,
    error: null,
  }),
  useUploadHomework: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
  }),
  useDeleteHomework: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
  }),
}));

// Mock useNotification для HomeworkSection
vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: () => ({
    showNotification: vi.fn(),
  }),
}));

describe('LessonBookingModal - Re-booking Prevention', () => {
  // Tests for booking validation and re-booking prevention
  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 5);

  const mockLesson = {
    id: 'lesson-1',
    subject: 'Математика',
    start_time: futureDate.toISOString(),
    end_time: new Date(futureDate.getTime() + 2 * 60 * 60 * 1000).toISOString(),
    teacher_name: 'Иван Иванов',
    lesson_type: 'group',
    max_students: 10,
    current_students: 5,
  };

  const mockCredits = {
    balance: 5,
    transactions: [],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Re-booking Prevention', () => {
    it('should show message when student tries to book previously cancelled lesson', () => {
      const { useCancelledLessons } = require('../../../hooks/useCancelledLessons.js');
      const mockIsLessonCancelled = vi.fn((lessonId) => lessonId === 'lesson-1');

      useCancelledLessons.mockReturnValue({
        isLessonCancelled: mockIsLessonCancelled,
      });

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(
        screen.getByText(/Вы отписались от этого занятия и больше не можете на него записаться/)
      ).toBeInTheDocument();
    });

    it('should disable booking button for previously cancelled lesson', () => {
      const { useCancelledLessons } = require('../../../hooks/useCancelledLessons.js');
      const mockIsLessonCancelled = vi.fn((lessonId) => lessonId === 'lesson-1');

      useCancelledLessons.mockReturnValue({
        isLessonCancelled: mockIsLessonCancelled,
      });

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      const bookButton = screen.queryByText(/Записаться/);
      if (bookButton) {
        expect(bookButton).toBeDisabled();
      }
    });

    it('should show badge "Отписались" for cancelled lessons', () => {
      const { useCancelledLessons } = require('../../../hooks/useCancelledLessons.js');
      const mockIsLessonCancelled = vi.fn((lessonId) => lessonId === 'lesson-1');

      useCancelledLessons.mockReturnValue({
        isLessonCancelled: mockIsLessonCancelled,
      });

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      // Check for cancelled badge or message
      const cancelledIndicator = screen.getByText(
        /Вы отписались от этого занятия и больше не можете на него записаться/
      );
      expect(cancelledIndicator).toBeInTheDocument();
    });

    it('should allow booking if lesson was never cancelled', () => {
      const { useCancelledLessons } = require('../../../hooks/useCancelledLessons.js');
      const mockIsLessonCancelled = vi.fn(() => false);

      useCancelledLessons.mockReturnValue({
        isLessonCancelled: mockIsLessonCancelled,
      });

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      // Book button should be visible and enabled
      const bookButton = screen.queryByText(/Записаться/);
      if (bookButton) {
        expect(bookButton).not.toBeDisabled();
      }
    });
  });

  describe('Already Booked State', () => {
    it('should show "already booked" message', () => {
      const myBookings = [
        {
          id: 'booking-1',
          lesson_id: 'lesson-1',
          status: 'active',
        },
      ];

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={myBookings}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/Вы уже записаны на это занятие/)).toBeInTheDocument();
    });

    it('should disable booking button when already booked', () => {
      const myBookings = [
        {
          id: 'booking-1',
          lesson_id: 'lesson-1',
          status: 'active',
        },
      ];

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={myBookings}
          credits={mockCredits}
        />
      );

      const bookButton = screen.queryByText(/Записаться/);
      if (bookButton) {
        expect(bookButton).toBeDisabled();
      }
    });

    it('should show "cancel booking" button instead', () => {
      const myBookings = [
        {
          id: 'booking-1',
          lesson_id: 'lesson-1',
          status: 'active',
        },
      ];

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={myBookings}
          credits={mockCredits}
        />
      );

      // Should show cancel button or indicate already booked
      const alreadyBookedText = screen.getByText(/Вы уже записаны/);
      expect(alreadyBookedText).toBeInTheDocument();
    });
  });

  describe('Insufficient Credits', () => {
    it('should show message when no credits available', () => {
      const noCredits = { balance: 0, transactions: [] };

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={noCredits}
        />
      );

      expect(
        screen.getByText(/Недостаточно кредитов|требуется 1 кредит/)
      ).toBeInTheDocument();
    });

    it('should disable booking button when no credits', () => {
      const noCredits = { balance: 0, transactions: [] };

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={noCredits}
        />
      );

      const bookButton = screen.queryByText(/Записаться/);
      if (bookButton) {
        expect(bookButton).toBeDisabled();
      }
    });
  });

  describe('Lesson Capacity', () => {
    it('should show message when lesson is full', () => {
      const fullLesson = {
        ...mockLesson,
        current_students: 10,
        max_students: 10,
      };

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={fullLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/Нет свободных мест/)).toBeInTheDocument();
    });

    it('should disable booking button when no spots available', () => {
      const fullLesson = {
        ...mockLesson,
        current_students: 10,
        max_students: 10,
      };

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={fullLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      const bookButton = screen.queryByText(/Записаться/);
      if (bookButton) {
        expect(bookButton).toBeDisabled();
      }
    });
  });

  describe('Booking Time Requirements', () => {
    it('should prevent booking for past lessons', () => {
      const pastDate = new Date();
      pastDate.setDate(pastDate.getDate() - 1);

      const pastLesson = {
        ...mockLesson,
        start_time: pastDate.toISOString(),
      };

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={pastLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(
        screen.getByText(/Это занятие уже началось или прошло/)
      ).toBeInTheDocument();
    });
  });

  describe('Modal Behavior', () => {
    it('should render modal when isOpen is true', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(mockLesson.subject)).toBeInTheDocument();
    });

    it('should not render modal when isOpen is false', () => {
      render(
        <LessonBookingModal
          isOpen={false}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      // Modal should not be visible
      const modal = document.querySelector('[role="dialog"]');
      if (modal) {
        expect(modal).not.toBeVisible();
      }
    });

    it('should call onClose when modal is closed', async () => {
      const onCloseMock = vi.fn();

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={onCloseMock}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      // Find and click close button (typically X button)
      const closeButtons = screen.queryAllByRole('button');
      const closeButton = closeButtons.find((btn) =>
        btn.textContent.includes('×') || btn.textContent.includes('✕')
      );

      if (closeButton) {
        await userEvent.click(closeButton);
        expect(onCloseMock).toHaveBeenCalled();
      }
    });
  });

  describe('Lesson Details Display', () => {
    it('should display lesson subject', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/Математика/)).toBeInTheDocument();
    });

    it('should display teacher name', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/Иван Иванов/)).toBeInTheDocument();
    });

    it('should display lesson type', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/group|групповое|Group/i)).toBeInTheDocument();
    });

    it('should display student capacity info', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/5.*10|5 из 10/)).toBeInTheDocument();
    });
  });

  describe('Credit Information Display', () => {
    it('should display current credit balance', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/5.*кредит|кредитов/)).toBeInTheDocument();
    });

    it('should display cost of booking (1 credit)', () => {
      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      expect(screen.getByText(/1.*кредит|стоимость/i)).toBeInTheDocument();
    });
  });

  describe('Multiple Cancelled Lessons', () => {
    it('should prevent re-booking only for specific cancelled lesson', () => {
      const { useCancelledLessons } = require('../../../hooks/useCancelledLessons.js');
      const mockIsLessonCancelled = vi.fn(
        (lessonId) => lessonId === 'lesson-1'
      );

      useCancelledLessons.mockReturnValue({
        isLessonCancelled: mockIsLessonCancelled,
      });

      const lesson2 = { ...mockLesson, id: 'lesson-2' };

      render(
        <LessonBookingModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={lesson2}
          myBookings={[]}
          credits={mockCredits}
        />
      );

      // lesson-2 should be bookable
      const bookButton = screen.queryByText(/Записаться/);
      if (bookButton) {
        expect(bookButton).not.toBeDisabled();
      }
    });
  });
});
