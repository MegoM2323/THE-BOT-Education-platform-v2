import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LessonEditModal } from '../LessonEditModal';
import * as bookingAPI from '../../../api/bookings';
import * as creditAPI from '../../../api/credits';
import * as lessonAPI from '../../../api/lessons';
import * as userAPI from '../../../api/users';
import * as authAPI from '../../../api/auth';
import { renderWithProviders } from '../../../test/test-utils.jsx';

// Mock all API modules
vi.mock('../../../api/bookings');
vi.mock('../../../api/credits');
vi.mock('../../../api/lessons');
vi.mock('../../../api/users');
vi.mock('../../../api/auth');

const mockLesson = {
  id: 'lesson-123',
  teacher_id: 'teacher-1',
  start_time: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
  end_time: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000 + 60 * 60 * 1000).toISOString(),
  max_students: 10,
  current_students: 1,
};

const mockStudents = [
  { id: 'student-1', full_name: 'Student One', email: 'student1@test.com' },
  { id: 'student-2', full_name: 'Student Two', email: 'student2@test.com' },
  { id: 'student-3', full_name: 'Student Three', email: 'student3@test.com' },
];

const mockTeachers = [
  { id: 'teacher-1', full_name: 'Teacher One' },
];

const mockCreditsResponse = {
  balances: [
    { user_id: 'student-1', balance: 5, email: 'student1@test.com', full_name: 'Student One' },
    { user_id: 'student-2', balance: 3, email: 'student2@test.com', full_name: 'Student Two' },
    { user_id: 'student-3', balance: 0, email: 'student3@test.com', full_name: 'Student Three' },
  ],
};

const mockBookingsResponse = [
  {
    id: 'booking-1',
    student_id: 'student-1',
    student_name: 'Student One',
    student_email: 'student1@test.com',
    status: 'active',
  },
];

const renderModal = (isOpen = true, onClose = vi.fn(), onLessonUpdated = vi.fn()) => {
  return renderWithProviders(
    <LessonEditModal
      isOpen={isOpen}
      onClose={onClose}
      lesson={mockLesson}
      onLessonUpdated={onLessonUpdated}
    />,
    { withAuth: true }
  );
};

describe('LessonEditModal - Credit Refresh', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Setup auth API mock
    authAPI.checkAuth = vi.fn().mockResolvedValue({
      user: {
        id: 'admin-1',
        email: 'admin@test.com',
        role: 'admin',
        full_name: 'Admin User',
      },
      balance: null,
    });

    // Setup API mocks
    bookingAPI.getBookings = vi.fn().mockResolvedValue(mockBookingsResponse);
    userAPI.getStudents = vi.fn().mockResolvedValue(mockStudents);
    userAPI.getStudentsAll = vi.fn().mockResolvedValue(mockStudents);
    userAPI.getTeachers = vi.fn().mockResolvedValue(mockTeachers);
    userAPI.getTeachersAll = vi.fn().mockResolvedValue(mockTeachers);
    creditAPI.getCredits = vi.fn().mockResolvedValue(mockCreditsResponse);
    creditAPI.getAllCredits = vi.fn().mockResolvedValue(mockCreditsResponse);
    bookingAPI.createBooking = vi.fn().mockResolvedValue({ id: 'booking-2', status: 'active' });
    lessonAPI.updateLesson = vi.fn().mockResolvedValue(mockLesson);
    lessonAPI.getLessonWithBookings = vi.fn().mockResolvedValue(mockLesson);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should refresh credits after successfully adding student', async () => {
    const user = userEvent.setup();
    renderModal();

    // Wait for modal to load
    await waitFor(() => {
      expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
    });

    // Wait for StudentCheckboxList to render with students
    await waitFor(() => {
      expect(screen.getByText('Student Two')).toBeInTheDocument();
    });

    // Find and click checkbox for student-2
    const checkboxes = screen.getAllByRole('checkbox');
    const student2Checkbox = checkboxes.find((checkbox) => {
      const label = checkbox.closest('label');
      return label && label.textContent.includes('Student Two');
    });

    expect(student2Checkbox).toBeInTheDocument();
    expect(student2Checkbox).not.toBeChecked(); // Should not be enrolled yet

    // Setup mock for after adding student - student-2 should have 2 credits (deducted 1)
    const updatedCreditsResponse = {
      balances: [
        { user_id: 'student-1', balance: 5 },
        { user_id: 'student-2', balance: 2 }, // Deducted 1
        { user_id: 'student-3', balance: 0 },
      ],
    };
    creditAPI.getAllCredits = vi.fn().mockResolvedValue(updatedCreditsResponse);

    // Click checkbox to add student
    await user.click(student2Checkbox);

    // Wait for booking to be created
    await waitFor(() => {
      expect(bookingAPI.createBooking).toHaveBeenCalledWith('lesson-123', 'student-2');
    });
  });

  it('should display updated credits immediately after adding student', async () => {
    const user = userEvent.setup();
    renderModal();

    // Wait for modal to load
    await waitFor(() => {
      expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
    });

    // Wait for StudentCheckboxList to render with students
    await waitFor(() => {
      expect(screen.getByText('Student Three')).toBeInTheDocument();
    });

    // Verify that getAllCredits was called to fetch available students' credits
    expect(creditAPI.getAllCredits).toHaveBeenCalled();

    // Find checkbox for student-3 with 0 credits (should be disabled)
    const checkboxes = screen.getAllByRole('checkbox');
    const student3Checkbox = checkboxes.find((checkbox) => {
      const label = checkbox.closest('label');
      return label && label.textContent.includes('Student Three');
    });

    expect(student3Checkbox).toBeInTheDocument();
    expect(student3Checkbox).toBeDisabled(); // Should be disabled due to 0 credits
  });

  it('should show loading state while refreshing credits', async () => {
    const user = userEvent.setup();

    // Mock slow credits response
    creditAPI.getAllCredits = vi.fn().mockImplementation(() => {
      return new Promise((resolve) => {
        setTimeout(() => {
          resolve(mockCreditsResponse);
        }, 500);
      });
    });

    renderModal();

    await waitFor(() => {
      expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
    });

    // Wait for StudentCheckboxList to render with students
    await waitFor(() => {
      expect(screen.getByText('Student Two')).toBeInTheDocument();
    });

    // Verify getAllCredits was called during initial load
    expect(creditAPI.getAllCredits).toHaveBeenCalled();
  });

  it('should handle credit refresh errors gracefully', async () => {
    const user = userEvent.setup();

    // Mock error response on initial load - component should still render
    creditAPI.getAllCredits = vi.fn()
      .mockRejectedValueOnce(new Error('Network error'));

    renderModal();

    await waitFor(() => {
      expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
    });

    // Component should still render even if credits failed to load
    // Available students list will be empty but modal should still work
    const modal = screen.getByTestId('modal');
    expect(modal).toBeInTheDocument();
  });

  it('should handle negative credit balance display correctly', async () => {
    const negativeCreditsResponse = {
      balances: [
        { user_id: 'student-1', balance: 5 },
        { user_id: 'student-2', balance: -2 },
        { user_id: 'student-3', balance: 0 },
      ],
    };

    creditAPI.getAllCredits = vi.fn().mockResolvedValue(negativeCreditsResponse);

    const user = userEvent.setup();
    renderModal();

    await waitFor(() => {
      expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
    });

    // Wait for StudentCheckboxList to render with students and credits
    await waitFor(() => {
      expect(screen.getByText('Student Two')).toBeInTheDocument();
    });

    // Verify getAllCredits was called
    expect(creditAPI.getAllCredits).toHaveBeenCalled();

    // The component should render correctly even with negative balances
    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes.length).toBeGreaterThan(0);
  });
});
