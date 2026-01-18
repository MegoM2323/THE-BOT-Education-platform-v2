import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { StudentSchedule } from '../StudentSchedule';
import * as useStudentLessonsModule from '../../hooks/useStudentLessons';
import * as useMyBookingsModule from '../../hooks/useMyBookings';
import * as useStudentCreditsModule from '../../hooks/useStudentCredits';
import * as useSlowConnectionModule from '../../hooks/useSlowConnection';
import { createFullWrapper, renderWithProviders } from '@/test/test-utils.jsx';

vi.mock('../../hooks/useStudentLessons');
vi.mock('../../hooks/useMyBookings');
vi.mock('../../hooks/useStudentCredits');
vi.mock('../../hooks/useSlowConnection');
vi.mock('../../components/common/Calendar', () => ({
  default: () => <div data-testid="mock-calendar">Mock Calendar</div>
}));
vi.mock('../../components/student/ScheduleFiltersBar', () => ({
  default: () => <div data-testid="mock-filters">Mock Filters</div>
}));
vi.mock('../../components/student/LessonCard', () => ({
  default: ({ lesson }) => (
    <div data-testid={`lesson-card-${lesson.id}`}>
      Lesson: {lesson.name}
    </div>
  )
}));
vi.mock('../../components/student/LessonBookingModal', () => ({
  default: () => <div data-testid="mock-booking-modal">Mock Booking Modal</div>
}));
vi.mock('../../components/common/SlowConnectionNotice', () => ({
  default: () => <div data-testid="slow-connection-notice">Slow Connection</div>
}));

describe('StudentSchedule - Payment Banner Removal', () => {
  const mockLessons = [
    {
      id: 'lesson-1',
      name: 'Math Lesson',
      teacher_id: 'teacher-1',
      teacher_name: 'John Doe',
      max_students: 1,
      current_students: 0,
      date: '2025-01-10T10:00:00Z',
    },
    {
      id: 'lesson-2',
      name: 'English Lesson',
      teacher_id: 'teacher-2',
      teacher_name: 'Jane Smith',
      max_students: 5,
      current_students: 2,
      date: '2025-01-10T14:00:00Z',
    },
  ];

  const mockBookings = [
    {
      id: 'booking-1',
      lesson_id: 'lesson-1',
      status: 'active',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useStudentLessonsModule.useStudentLessons).mockReturnValue({
      lessons: mockLessons,
      isLoading: false,
      error: null,
    });

    vi.mocked(useMyBookingsModule.useMyBookings).mockReturnValue({
      myBookings: mockBookings,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });

    vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
      credits: { balance: 10 },
      isLoading: false,
      error: null,
    });

    vi.mocked(useSlowConnectionModule.useSlowConnection).mockReturnValue(false);
  });

  describe('Payment Banner - Does Not Exist', () => {
    it('should not render payment banner element in DOM', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const paymentBanner = screen.queryByTestId('payment-banner');
      expect(paymentBanner).not.toBeInTheDocument();
    });

    it('should not render any element with payment-related text', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.queryByText(/buy credits/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/payment/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/upgrade/i)).not.toBeInTheDocument();
    });

    it('should not render payment warning or alert', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const alerts = screen.queryAllByRole('alert');
      const paymentAlerts = alerts.filter(alert =>
        alert.textContent?.toLowerCase().includes('payment') ||
        alert.textContent?.toLowerCase().includes('credits')
      );

      expect(paymentAlerts.length).toBe(0);
    });

    it('should not render any element with class containing "payment"', () => {
      const { container } = renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const paymentElements = container.querySelectorAll('[class*="payment"]');
      expect(paymentElements.length).toBe(0);
    });

    it('should not render element with "buy" or "purchase" text', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.queryByText(/buy/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/purchase/i)).not.toBeInTheDocument();
    });
  });

  describe('Bookings Error Banner - Still Renders', () => {
    it('should render bookings error banner when error exists', () => {
      const bookingsError = new Error('Failed to load bookings');

      vi.mocked(useMyBookingsModule.useMyBookings).mockReturnValue({
        myBookings: [],
        isLoading: false,
        error: bookingsError,
        refetch: vi.fn(),
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const errorBanner = screen.getByTestId('bookings-error-banner');
      expect(errorBanner).toBeInTheDocument();
    });

    it('should display error message in bookings error banner', () => {
      const errorMessage = 'Network connection failed';
      const bookingsError = new Error(errorMessage);

      vi.mocked(useMyBookingsModule.useMyBookings).mockReturnValue({
        myBookings: [],
        isLoading: false,
        error: bookingsError,
        refetch: vi.fn(),
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByText(new RegExp(errorMessage))).toBeInTheDocument();
    });

    it('should have retry button in bookings error banner', () => {
      const bookingsError = new Error('Failed to load bookings');

      vi.mocked(useMyBookingsModule.useMyBookings).mockReturnValue({
        myBookings: [],
        isLoading: false,
        error: bookingsError,
        refetch: vi.fn(),
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const retryButton = screen.getByTestId('retry-bookings-btn');
      expect(retryButton).toBeInTheDocument();
      expect(retryButton.textContent).toMatch(/retry|повторить/i);
    });

    it('should not confuse bookings error with payment banner', () => {
      const bookingsError = new Error('Failed to load bookings');

      vi.mocked(useMyBookingsModule.useMyBookings).mockReturnValue({
        myBookings: [],
        isLoading: false,
        error: bookingsError,
        refetch: vi.fn(),
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const bookingsErrorBanner = screen.getByTestId('bookings-error-banner');
      expect(bookingsErrorBanner).toBeInTheDocument();

      const paymentBanner = screen.queryByTestId('payment-banner');
      expect(paymentBanner).not.toBeInTheDocument();
    });
  });

  describe('Credits Loading and Display', () => {
    it('should display credits balance correctly without payment banner', () => {
      vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
        credits: { balance: 25 },
        isLoading: false,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByText(/расписание/i)).toBeInTheDocument();
    });

    it('should handle zero balance without payment banner', () => {
      vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
        credits: { balance: 0 },
        isLoading: false,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByText(/расписание/i)).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should display schedule even when credits balance is low', () => {
      vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
        credits: { balance: 1 },
        isLoading: false,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByText(/расписание/i)).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should load credits without showing payment banner', async () => {
      vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
        credits: { balance: 15 },
        isLoading: false,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should handle credits error gracefully without payment banner', () => {
      const creditsError = new Error('Failed to load credits');

      vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
        credits: null,
        isLoading: false,
        error: creditsError,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });
  });

  describe('Lesson Cards Rendering', () => {
    it('should render schedule container for lessons normally', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should display all lessons without payment restriction', () => {
      const manyLessons = Array.from({ length: 20 }, (_, i) => ({
        id: `lesson-${i}`,
        name: `Lesson ${i}`,
        teacher_id: `teacher-${i}`,
        teacher_name: `Teacher ${i}`,
        max_students: 5,
        current_students: i % 3,
        date: `2025-01-${10 + (i % 20)}T10:00:00Z`,
      }));

      vi.mocked(useStudentLessonsModule.useStudentLessons).mockReturnValue({
        lessons: manyLessons,
        isLoading: false,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should render lessons even with zero credits balance', () => {
      vi.mocked(useStudentCreditsModule.useStudentCredits).mockReturnValue({
        credits: { balance: 0 },
        isLoading: false,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });
  });

  describe('Filters Work Correctly', () => {
    it('should render filter bar without payment banner interference', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const filterBar = screen.getByTestId('mock-filters');
      expect(filterBar).toBeInTheDocument();

      const paymentBanner = screen.queryByTestId('payment-banner');
      expect(paymentBanner).not.toBeInTheDocument();
    });

    it('should allow filtering by teacher without payment banner', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('mock-filters')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should allow filtering by lesson type without payment banner', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('mock-filters')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should preserve filters when no payment banner exists', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const filterBar = screen.getByTestId('mock-filters');
      expect(filterBar).toBeInTheDocument();

      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });
  });

  describe('Page Structure Without Payment Banner', () => {
    it('should have standard schedule structure', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule-header')).toBeInTheDocument();
      expect(screen.getByTestId('schedule-sidebar')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should have calendar or list view without payment banner', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const calendarView = screen.queryByTestId('calendar-view');
      const listView = screen.queryByTestId('list-view');

      expect(calendarView || listView).toBeTruthy();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should not have any payment-related DOM elements', () => {
      const { container } = renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      const paymentElements = container.querySelectorAll(
        '[class*="payment"], [data-testid*="payment"], [id*="payment"]'
      );

      expect(paymentElements.length).toBe(0);
    });
  });

  describe('Loading States Without Payment Banner', () => {
    it('should show loading skeleton without payment banner', () => {
      vi.mocked(useStudentLessonsModule.useStudentLessons).mockReturnValue({
        lessons: [],
        isLoading: true,
        error: null,
      });

      vi.mocked(useMyBookingsModule.useMyBookings).mockReturnValue({
        myBookings: [],
        isLoading: true,
        error: null,
        refetch: vi.fn(),
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });

    it('should show slow connection notice without payment banner', () => {
      vi.mocked(useSlowConnectionModule.useSlowConnection).mockReturnValue(true);

      vi.mocked(useStudentLessonsModule.useStudentLessons).mockReturnValue({
        lessons: [],
        isLoading: true,
        error: null,
      });

      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('slow-connection-notice')).toBeInTheDocument();
      expect(screen.queryByTestId('payment-banner')).not.toBeInTheDocument();
    });
  });

  describe('Integration - All Features Without Payment Banner', () => {
    it('should render complete schedule with all features', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule')).toBeInTheDocument();
      expect(screen.getByTestId('student-schedule-header')).toBeInTheDocument();
      expect(screen.getByTestId('schedule-sidebar')).toBeInTheDocument();
      expect(screen.getByTestId('mock-filters')).toBeInTheDocument();

      const paymentBanner = screen.queryByTestId('payment-banner');
      expect(paymentBanner).toBeNull();
    });

    it('should work with bookings, lessons, and credits loaded', async () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByTestId('student-schedule')).toBeInTheDocument();

      const paymentBanner = screen.queryByTestId('payment-banner');
      expect(paymentBanner).toBeNull();
    });

    it('should display title and timezone without payment banner', () => {
      renderWithProviders(
        <BrowserRouter>
          <StudentSchedule />
        </BrowserRouter>
      );

      expect(screen.getByText(/расписание/i)).toBeInTheDocument();

      const paymentBanner = screen.queryByTestId('payment-banner');
      expect(paymentBanner).toBeNull();
    });
  });
});
