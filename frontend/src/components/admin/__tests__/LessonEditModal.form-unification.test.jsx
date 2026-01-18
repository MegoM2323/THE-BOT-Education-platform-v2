import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LessonEditModal } from "../LessonEditModal.jsx";

vi.mock('../../../api/lessons.js', () => ({
  updateLesson: vi.fn(() => Promise.resolve({ success: true })),
  deleteLesson: vi.fn(() => Promise.resolve({ success: true }))
}));

vi.mock('../../../api/bookings.js', () => ({
  getBookings: vi.fn(() => Promise.resolve([])),
  createBooking: vi.fn(() => Promise.resolve({ success: true })),
  cancelBooking: vi.fn(() => Promise.resolve({ success: true }))
}));

vi.mock('../../../api/users.js', () => ({
  getStudents: vi.fn(() => Promise.resolve([])),
  getTeachers: vi.fn(() => Promise.resolve([
    { id: 'teacher-123', full_name: 'John Doe' },
    { id: 'teacher-456', full_name: 'Jane Smith' }
  ]))
}));

vi.mock('../../../api/credits.js', () => ({
  getAllCredits: vi.fn(() => Promise.resolve({ balances: [] }))
}));

vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: () => ({ showNotification: vi.fn() })
}));

vi.mock('../../../hooks/useBulkEdit.js', () => ({
  useApplyToAllSubsequent: () => ({
    mutateAsync: vi.fn(() => Promise.resolve({ affected_lessons_count: 1 })),
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

vi.mock('../../../hooks/useAuth.js', () => ({
  useAuth: () => ({
    user: { id: 'admin-user-id', role: 'admin', full_name: 'Admin User' }
  })
}));

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
  default: () => null
}));

vi.mock('../../common/ColorPicker.jsx', () => ({
  default: () => <div data-testid="color-picker">Color Picker</div>
}));

vi.mock('../HomeworkSection.jsx', () => ({
  default: () => <div data-testid="homework-section">Homework Section</div>
}));

vi.mock('../BroadcastSection.jsx', () => ({
  default: () => <div data-testid="broadcast-section">Broadcast Section</div>
}));

vi.mock('../StudentCheckboxList.jsx', () => ({
  default: () => <div data-testid="student-checkbox-list">Student Checkbox List</div>
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false }
    }
  });

  return ({ children }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

const mockLesson = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  start_time: '2024-01-15T10:00:00Z',
  end_time: '2024-01-15T11:00:00Z',
  max_students: 5,
  subject: 'Mathematics',
  teacher_id: 'teacher-123',
  color: '#3B82F6'
};

describe('LessonEditModal - Form Unification', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Class Names Verification', () => {
    it('should render form-input class on date input', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const dateInput = container.querySelector('input[type="date"]');
      expect(dateInput).toBeInTheDocument();
      expect(dateInput).toHaveClass('form-input');
    });

    it('should render form-input class on time inputs', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const timeInputs = container.querySelectorAll('input[type="time"]');
      expect(timeInputs.length).toBe(2);
      timeInputs.forEach(input => {
        expect(input).toHaveClass('form-input');
      });
    });

    it('should render form-input class on max students number input', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const numberInput = container.querySelector('input[type="number"]');
      expect(numberInput).toBeInTheDocument();
      expect(numberInput).toHaveClass('form-input');
    });

    it('should render form-input class on subject text input', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const textInput = container.querySelector('input[type="text"]');
      expect(textInput).toBeInTheDocument();
      expect(textInput).toHaveClass('form-input');
    });

    it('should render form-select class on teacher select', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const select = container.querySelector('select');
      expect(select).toBeInTheDocument();
      expect(select).toHaveClass('form-select');
    });

    it('should NOT use old info-* class names anywhere in form', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const elementsWithInfoInput = container.querySelectorAll('.info-input');
      const elementsWithInfoSelect = container.querySelectorAll('.info-select');

      expect(elementsWithInfoInput.length).toBe(0);
      expect(elementsWithInfoSelect.length).toBe(0);
    });
  });

  describe('Form Input Consistency', () => {
    it('all form inputs should have consistent styling', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const formInputs = container.querySelectorAll('.form-input');
      expect(formInputs.length).toBeGreaterThan(0);

      formInputs.forEach(input => {
        expect(input).toHaveClass('form-input');
      });
    });

    it('form-input and form-select should exist', async () => {
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const formInputs = container.querySelectorAll('.form-input');
      const formSelects = container.querySelectorAll('.form-select');

      expect(formInputs.length).toBeGreaterThan(0);
      expect(formSelects.length).toBeGreaterThan(0);
    });
  });

  describe('Form Functionality After Unification', () => {
    it('form inputs should still be editable', async () => {
      const user = userEvent.setup();
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const subjectInput = container.querySelector('input[type="text"]');
      expect(subjectInput).toBeInTheDocument();

      await user.clear(subjectInput);
      await user.type(subjectInput, 'Physics');

      expect(subjectInput).toHaveValue('Physics');
    });

    it('form inputs should be focusable', async () => {
      const user = userEvent.setup();
      const { container } = render(
        <LessonEditModal
          isOpen={true}
          onClose={vi.fn()}
          lesson={mockLesson}
          onLessonUpdated={vi.fn()}
        />,
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(screen.queryByTestId('spinner')).not.toBeInTheDocument();
      });

      const dateInput = container.querySelector('input[type="date"]');
      expect(dateInput).toBeInTheDocument();

      await user.click(dateInput);
      expect(dateInput).toHaveFocus();
    });
  });
});
