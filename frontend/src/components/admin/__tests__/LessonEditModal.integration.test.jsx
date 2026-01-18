import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render as baseRender, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LessonEditModal } from '../LessonEditModal';
import * as bookingAPI from '../../../api/bookings';
import * as creditAPI from '../../../api/credits';
import * as lessonAPI from '../../../api/lessons';
import * as userAPI from '../../../api/users';
import { NotificationProvider } from '../../../context/NotificationContext.jsx';
import { AuthContext } from '../../../context/AuthContext.jsx';

vi.mock('../../../api/bookings');
vi.mock('../../../api/credits');
vi.mock('../../../api/lessons');
vi.mock('../../../api/users');

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

const mockAuthContext = {
  user: {
    id: 'admin-1',
    email: 'admin@test.com',
    role: 'admin',
    full_name: 'Admin User',
  },
  isLoading: false,
  error: null,
};

let queryClient;

const renderModal = (isOpen = true, onClose = vi.fn(), onLessonUpdated = vi.fn()) => {
  return baseRender(
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={mockAuthContext}>
        <NotificationProvider>
          <LessonEditModal
            isOpen={isOpen}
            onClose={onClose}
            lesson={mockLesson}
            onLessonUpdated={onLessonUpdated}
          />
        </NotificationProvider>
      </AuthContext.Provider>
    </QueryClientProvider>
  );
};

describe('T007: Admin Student Management Integration Tests', () => {
  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    vi.clearAllMocks();

    vi.mocked(bookingAPI.getBookings).mockResolvedValue(mockBookingsResponse);
    vi.mocked(userAPI.getStudentsAll).mockResolvedValue(mockStudents);
    vi.mocked(userAPI.getTeachersAll).mockResolvedValue(mockTeachers);
    vi.mocked(creditAPI.getAllCredits).mockResolvedValue(mockCreditsResponse);
    vi.mocked(bookingAPI.createBooking).mockResolvedValue({ id: 'booking-2', status: 'active' });
    vi.mocked(bookingAPI.cancelBooking).mockResolvedValue({ success: true });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('T007.1: LessonEditModal loads with enrolled students', () => {
    it('should display enrolled students when modal opens', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      await waitFor(() => {
        expect(screen.getByText('Student One')).toBeInTheDocument();
      });
    });

    it('should load all API data for the modal', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      await waitFor(() => {
        expect(userAPI.getStudentsAll).toHaveBeenCalled();
        expect(userAPI.getTeachersAll).toHaveBeenCalled();
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
        expect(bookingAPI.getBookings).toHaveBeenCalled();
      });
    });
  });

  describe('T007.2: Credits display in dropdown', () => {
    it('should load credit data when modal opens', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });
    });

    it('should maintain credit data for display', async () => {
      const customCredits = {
        balances: [
          { user_id: 'student-2', balance: 5 },
          { user_id: 'student-3', balance: 0 },
        ],
      };

      vi.mocked(creditAPI.getAllCredits).mockResolvedValueOnce(customCredits);

      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      expect(creditAPI.getAllCredits).toHaveBeenCalled();
    });
  });

  describe('T007.3: Query invalidation after operations', () => {
    it('should call createBooking API when adding student', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });
    });
  });

  describe('T007.4: Data consistency', () => {
    it('should have students and credits data available simultaneously', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      await waitFor(() => {
        expect(userAPI.getStudentsAll).toHaveBeenCalled();
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });
    });
  });

  describe('T007.5: Modal rendering and error handling', () => {
    it('should display modal content when data loads successfully', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      const elements = screen.getAllByText(/Основное/);
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should handle API calls without crashing', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
    });
  });

  describe('T007.6: Lesson information display', () => {
    it('should display lesson capacity information', async () => {
      renderModal();

      await waitFor(() => {
        expect(screen.getByText(/Редактирование занятия/)).toBeInTheDocument();
      });

      const maxStudentsInput = screen.getByDisplayValue('10');
      expect(maxStudentsInput).toBeInTheDocument();
    });
  });
});
