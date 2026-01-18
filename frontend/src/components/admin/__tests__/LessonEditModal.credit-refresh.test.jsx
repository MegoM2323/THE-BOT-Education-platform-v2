import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LessonEditModal } from '../LessonEditModal';
import * as bookingAPI from '../../../api/bookings';
import * as creditAPI from '../../../api/credits';
import * as userAPI from '../../../api/users';
import * as lessonAPI from '../../../api/lessons';
import * as notificationModule from '../../../hooks/useNotification';

vi.mock('../../../api/bookings');
vi.mock('../../../api/credits');
vi.mock('../../../api/users');
vi.mock('../../../api/lessons');
vi.mock('../../../hooks/useNotification');

vi.mock('../../common/Modal.jsx', () => ({
  default: ({ children, isOpen, title, onClose, footer }) =>
    isOpen ? (
      <div data-testid="modal">
        <h2>{title}</h2>
        {children}
        <div data-testid="modal-footer">{footer}</div>
      </div>
    ) : null
}));

vi.mock('../../common/Button.jsx', () => ({
  default: ({ children, onClick, loading, disabled, ...props }) => (
    <button onClick={onClick} disabled={disabled || loading} {...props}>
      {loading ? 'Loading...' : children}
    </button>
  )
}));

vi.mock('../../common/Spinner.jsx', () => ({
  default: ({ size }) => <div data-testid="spinner">{size || 'default'} spinner</div>
}));

vi.mock('../../common/ConfirmModal.jsx', () => ({
  default: ({ isOpen, onClose, onConfirm, title, message }) =>
    isOpen ? (
      <div data-testid="confirm-modal">
        <h3>{title}</h3>
        <p>{message}</p>
        <button onClick={onConfirm}>Confirm</button>
        <button onClick={onClose}>Cancel</button>
      </div>
    ) : null
}));

vi.mock('../../common/ColorPicker.jsx', () => ({
  default: () => <div data-testid="color-picker">Color Picker</div>
}));

vi.mock('../../../hooks/useAuth.js', () => ({
  useAuth: () => ({
    user: {
      id: 'admin-user-id',
      role: 'admin',
      full_name: 'Admin User'
    }
  })
}));

vi.mock('../../../hooks/useBulkEdit.js', () => ({
  useApplyToAllSubsequent: () => ({
    mutateAsync: vi.fn(() => Promise.resolve({ affected_lessons_count: 0 })),
    isPending: false
  })
}));

vi.mock('../../../hooks/useAutosave.js', () => ({
  useAutosave: () => ({
    isSaving: false,
    saveNow: vi.fn(() => Promise.resolve()),
    error: null,
    lastSaved: null
  })
}));

vi.mock('../HomeworkSection.jsx', () => ({
  default: () => <div data-testid="homework-section">Homework Section</div>
}));

vi.mock('../BroadcastSection.jsx', () => ({
  default: () => <div data-testid="broadcast-section">Broadcast Section</div>
}));

vi.mock('../StudentCheckboxList.jsx', () => ({
  default: ({ allStudents, enrolledStudentIds, onToggle }) => (
    <div data-testid="student-checkbox-list">
      {allStudents.map(s => (
        <div key={s.id}>
          {s.name} - {s.credits} credits
          <input
            type="checkbox"
            checked={enrolledStudentIds.includes(s.id)}
            onChange={(e) => onToggle(s.id, e.target.checked)}
          />
        </div>
      ))}
    </div>
  )
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false }
    }
  });

  return ({ children }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('LessonEditModal - Credit Refresh Logic', () => {
  const mockLesson = {
    id: 'lesson-123',
    teacher_id: 'teacher-1',
    start_time: '2025-12-20T10:00:00Z',
    end_time: '2025-12-20T11:00:00Z',
    max_students: 5,
    current_students: 2
  };

  const mockStudent1 = {
    id: 'student-1',
    full_name: 'John Doe',
    email: 'john@example.com'
  };

  const mockStudent2 = {
    id: 'student-2',
    full_name: 'Jane Smith',
    email: 'jane@example.com'
  };

  const mockStudent3 = {
    id: 'student-3',
    full_name: 'Bob Johnson',
    email: 'bob@example.com'
  };

  let mockShowNotification;

  beforeEach(() => {
    vi.clearAllMocks();

    mockShowNotification = vi.fn();
    vi.mocked(notificationModule.useNotification).mockReturnValue({
      showNotification: mockShowNotification
    });

    vi.mocked(bookingAPI.getBookings).mockResolvedValue([]);
    vi.mocked(creditAPI.getAllCredits).mockResolvedValue({
      balances: [
        { user_id: 'student-1', balance: 5 },
        { user_id: 'student-2', balance: 3 },
        { user_id: 'student-3', balance: 10 }
      ]
    });
    vi.mocked(userAPI.getStudentsAll).mockResolvedValue([mockStudent1, mockStudent2, mockStudent3]);
    vi.mocked(userAPI.getTeachersAll).mockResolvedValue([{ id: 'teacher-1', full_name: 'Test Teacher' }]);
    vi.mocked(bookingAPI.createBooking).mockResolvedValue({ id: 'booking-new' });
    vi.mocked(bookingAPI.cancelBooking).mockResolvedValue({ success: true });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Credit refresh after adding student', () => {
    it('should call getAllCredits when loading lesson data', async () => {
      render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });
    });

    it('should display students with their credit balances', async () => {
      render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.getByTestId('student-checkbox-list')).toBeInTheDocument();
      });

      expect(screen.getByText(/10 credits/)).toBeInTheDocument();
    });

    it('should refresh credits after successfully adding student', async () => {
      render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.getByTestId('student-checkbox-list')).toBeInTheDocument();
      });

      const initialCallCount = vi.mocked(creditAPI.getAllCredits).mock.calls.length;

      const checkbox = screen.getAllByRole('checkbox')[2];
      await userEvent.click(checkbox);

      await waitFor(() => {
        expect(bookingAPI.createBooking).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalledTimes(initialCallCount + 1);
      });
    });
  });

  describe('Negative balance display', () => {
    it('should prevent adding student with 0 credits (show error)', async () => {
      vi.mocked(creditAPI.getAllCredits).mockResolvedValue({
        balances: [
          { user_id: 'student-1', balance: 5 },
          { user_id: 'student-2', balance: 3 },
          { user_id: 'student-3', balance: 0 }
        ]
      });

      render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.getByTestId('student-checkbox-list')).toBeInTheDocument();
      });

      const checkbox = screen.getAllByRole('checkbox')[2];
      await userEvent.click(checkbox);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          expect.stringContaining('недостаточно кредитов'),
          'error'
        );
      });

      expect(bookingAPI.createBooking).not.toHaveBeenCalled();
    });
  });

  describe('Error handling during credit refresh', () => {
    it('should not crash UI if credit refresh fails', async () => {
      vi.mocked(creditAPI.getAllCredits)
        .mockResolvedValueOnce({
          balances: [{ user_id: 'student-3', balance: 10 }]
        })
        .mockRejectedValueOnce(new Error('Credit API error'));

      render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.getByTestId('student-checkbox-list')).toBeInTheDocument();
      });

      expect(screen.getByTestId('modal')).toBeInTheDocument();
    });
  });

  describe('Multiple students credit handling', () => {
    it('should load credits for all students', async () => {
      vi.mocked(creditAPI.getAllCredits).mockReset();
      vi.mocked(creditAPI.getAllCredits).mockResolvedValue({
        balances: [
          { user_id: 'student-1', balance: 5 },
          { user_id: 'student-2', balance: 3 },
          { user_id: 'student-3', balance: 10 }
        ]
      });

      render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.getByTestId('student-checkbox-list')).toBeInTheDocument();
      });

      expect(screen.getByText(/5 credits/)).toBeInTheDocument();
      expect(screen.getByText(/3 credits/)).toBeInTheDocument();
      expect(screen.getByText(/10 credits/)).toBeInTheDocument();
    });
  });
});
