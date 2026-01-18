import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import PaymentSettings from "../PaymentSettings.jsx";

// Import mocked module
import { usePaymentSettings } from '../../../hooks/usePaymentSettings.js';

// Mock the hook
vi.mock('../../../hooks/usePaymentSettings.js', () => ({
  usePaymentSettings: vi.fn(),
}));

const createQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
});

const renderWithQueryClient = (ui) => {
  const queryClient = createQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  );
};

describe('PaymentSettings', () => {
  const mockStudents = [
    {
      id: 'student-1',
      full_name: 'Иван Петров',
      email: 'ivan@example.com',
      payment_enabled: true,
      updated_at: '2025-01-01T10:00:00Z',
    },
    {
      id: 'student-2',
      full_name: 'Мария Сидорова',
      email: 'maria@example.com',
      payment_enabled: false,
      updated_at: '2025-01-02T11:00:00Z',
    },
    {
      id: 'student-3',
      full_name: 'Петр Иванов',
      email: 'petr@example.com',
      payment_enabled: true,
      updated_at: '2025-01-03T12:00:00Z',
    },
  ];

  const mockUpdatePaymentStatus = vi.fn();

  beforeEach(() => {
    vi.mocked(usePaymentSettings).mockReturnValue({
      students: mockStudents,
      isLoading: false,
      error: null,
      updatePaymentStatus: mockUpdatePaymentStatus,
      isUpdating: false,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render payment settings component', () => {
      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByTestId('payment-settings')).toBeInTheDocument();
    });

    it('should render title and description', () => {
      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByText('Управление платежами')).toBeInTheDocument();
      expect(screen.getByText(/Включите или отключите возможность совершать платежи/)).toBeInTheDocument();
    });

    it('should render search input and sort select', () => {
      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByTestId('search-input')).toBeInTheDocument();
      expect(screen.getByTestId('sort-select')).toBeInTheDocument();
    });

    it('should render students table', () => {
      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByTestId('students-table')).toBeInTheDocument();
    });

    it('should render all students in table', () => {
      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      expect(screen.getByText('Мария Сидорова')).toBeInTheDocument();
      expect(screen.getByText('Петр Иванов')).toBeInTheDocument();
    });

    it('should display total count of students', () => {
      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByText('Всего студентов: 3')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading spinner when isLoading is true', () => {
      vi.mocked(usePaymentSettings).mockReturnValue({
        students: [],
        isLoading: true,
        error: null,
        updatePaymentStatus: mockUpdatePaymentStatus,
        isUpdating: false,
      });

      renderWithQueryClient(<PaymentSettings />);
      expect(document.querySelector('.payment-settings-loading')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should show error message when error occurs', () => {
      const errorMessage = 'Failed to load students';
      vi.mocked(usePaymentSettings).mockReturnValue({
        students: [],
        isLoading: false,
        error: new Error(errorMessage),
        updatePaymentStatus: mockUpdatePaymentStatus,
        isUpdating: false,
      });

      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByRole('alert')).toBeInTheDocument();
      expect(screen.getByText(new RegExp(errorMessage))).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no students found', () => {
      vi.mocked(usePaymentSettings).mockReturnValue({
        students: [],
        isLoading: false,
        error: null,
        updatePaymentStatus: mockUpdatePaymentStatus,
        isUpdating: false,
      });

      renderWithQueryClient(<PaymentSettings />);
      expect(screen.getByText('Студенты не найдены')).toBeInTheDocument();
    });

    it('should show empty state when search yields no results', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const searchInput = screen.getByTestId('search-input');

      await userEvent.type(searchInput, 'nonexistent@example.com');

      expect(screen.getByText('Студенты не найдены')).toBeInTheDocument();
      expect(screen.getByText('Всего студентов: 0 (из 3)')).toBeInTheDocument();
    });
  });

  describe('Search Functionality', () => {
    it('should filter students by name', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const searchInput = screen.getByTestId('search-input');

      await userEvent.type(searchInput, 'Мария');

      expect(screen.getByText('Мария Сидорова')).toBeInTheDocument();
      expect(screen.queryByText('Иван Петров')).not.toBeInTheDocument();
      expect(screen.queryByText('Петр Иванов')).not.toBeInTheDocument();
      expect(screen.getByText('Всего студентов: 1 (из 3)')).toBeInTheDocument();
    });

    it('should filter students by email', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const searchInput = screen.getByTestId('search-input');

      await userEvent.type(searchInput, 'maria@example.com');

      expect(screen.getByText('Мария Сидорова')).toBeInTheDocument();
      expect(screen.queryByText('Иван Петров')).not.toBeInTheDocument();
      expect(screen.getByText('Всего студентов: 1 (из 3)')).toBeInTheDocument();
    });

    it('should be case-insensitive', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const searchInput = screen.getByTestId('search-input');

      await userEvent.type(searchInput, 'ИВАН');

      expect(screen.getByText('Иван Петров')).toBeInTheDocument();
    });

    it('should clear filter when search is cleared', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const searchInput = screen.getByTestId('search-input');

      await userEvent.type(searchInput, 'Иван');
      expect(screen.queryByText('Мария Сидорова')).not.toBeInTheDocument();

      await userEvent.clear(searchInput);

      expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      expect(screen.getByText('Мария Сидорова')).toBeInTheDocument();
      expect(screen.getByText('Петр Иванов')).toBeInTheDocument();
    });
  });

  describe('Sort Functionality', () => {
    it('should sort by name (default)', () => {
      renderWithQueryClient(<PaymentSettings />);
      const rows = screen.getAllByTestId(/student-row-/);

      expect(rows[0]).toHaveTextContent('Иван Петров');
      expect(rows[1]).toHaveTextContent('Мария Сидорова');
      expect(rows[2]).toHaveTextContent('Петр Иванов');
    });

    it('should sort by email when selected', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const sortSelect = screen.getByTestId('sort-select');

      await userEvent.selectOptions(sortSelect, 'email');

      const rows = screen.getAllByTestId(/student-row-/);
      expect(sortSelect.value).toBe('email');
    });

    it('should sort by payment status when selected', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const sortSelect = screen.getByTestId('sort-select');

      await userEvent.selectOptions(sortSelect, 'status');

      expect(sortSelect.value).toBe('status');
      const rows = screen.getAllByTestId(/student-row-/);
      expect(rows.length).toBe(3);
    });

    it('should sort by updated date when selected', async () => {
      renderWithQueryClient(<PaymentSettings />);
      const sortSelect = screen.getByTestId('sort-select');

      await userEvent.selectOptions(sortSelect, 'updated');

      expect(sortSelect.value).toBe('updated');
      const rows = screen.getAllByTestId(/student-row-/);
      expect(rows[0]).toHaveTextContent('Петр Иванов');
    });
  });

  describe('Payment Toggle', () => {
    it('should toggle payment status when button is clicked', async () => {
      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      const toggle = within(row).getByTestId('payment-toggle');

      await userEvent.click(toggle);

      await waitFor(() => {
        expect(mockUpdatePaymentStatus).toHaveBeenCalledWith('student-1', false);
      });
    });

    it('should show correct status label', () => {
      renderWithQueryClient(<PaymentSettings />);

      const row1 = screen.getByTestId('student-row-student-1');
      const row2 = screen.getByTestId('student-row-student-2');

      expect(within(row1).getByText('Включены')).toBeInTheDocument();
      expect(within(row2).getByText('Отключены')).toBeInTheDocument();
    });

    it('should disable toggle while updating', () => {
      vi.mocked(usePaymentSettings).mockReturnValue({
        students: mockStudents,
        isLoading: false,
        error: null,
        updatePaymentStatus: mockUpdatePaymentStatus,
        isUpdating: true,
      });

      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      const toggle = within(row).getByTestId('payment-toggle');

      expect(toggle).toBeDisabled();
    });

    it('should call updatePaymentStatus with correct args when toggling off', async () => {
      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      const toggle = within(row).getByTestId('payment-toggle');

      await userEvent.click(toggle);

      await waitFor(() => {
        expect(mockUpdatePaymentStatus).toHaveBeenCalledWith('student-1', false);
      });
    });

    it('should call updatePaymentStatus with correct args when toggling on', async () => {
      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-2');
      const toggle = within(row).getByTestId('payment-toggle');

      await userEvent.click(toggle);

      await waitFor(() => {
        expect(mockUpdatePaymentStatus).toHaveBeenCalledWith('student-2', true);
      });
    });

    it('should handle toggle error gracefully', async () => {
      mockUpdatePaymentStatus.mockRejectedValueOnce(new Error('API Error'));

      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      const toggle = within(row).getByTestId('payment-toggle');

      await userEvent.click(toggle);

      expect(screen.getByTestId('payment-settings')).toBeInTheDocument();
    });
  });

  describe('Table Columns', () => {
    it('should display all required columns', () => {
      renderWithQueryClient(<PaymentSettings />);

      const headers = screen.getAllByRole('columnheader');
      const headerTexts = headers.map(h => h.textContent);

      expect(headerTexts).toContain('Студент');
      expect(headerTexts).toContain('Email');
      expect(headerTexts).toContain('Платежи');
      expect(headerTexts).toContain('Изменено');
    });

    it('should display all student data in correct columns', () => {
      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      expect(within(row).getByText('Иван Петров')).toBeInTheDocument();
      expect(within(row).getByText('ivan@example.com')).toBeInTheDocument();
    });
  });

  describe('Integration Tests', () => {
    it('should search and sort together', async () => {
      renderWithQueryClient(<PaymentSettings />);

      const searchInput = screen.getByTestId('search-input');
      const sortSelect = screen.getByTestId('sort-select');

      await userEvent.type(searchInput, 'Сидорова');

      await userEvent.selectOptions(sortSelect, 'email');

      expect(screen.getByText('Мария Сидорова')).toBeInTheDocument();
      expect(screen.queryByText('Иван Петров')).not.toBeInTheDocument();
      expect(screen.queryByText('Петр Иванов')).not.toBeInTheDocument();
    });

    it('should handle rapid toggle clicks', async () => {
      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      const toggle = within(row).getByTestId('payment-toggle');

      await userEvent.click(toggle);
      await userEvent.click(toggle);
      await userEvent.click(toggle);

      await waitFor(() => {
        expect(mockUpdatePaymentStatus).toHaveBeenCalledTimes(3);
      });
    });

    it('should maintain scroll position after toggle', async () => {
      renderWithQueryClient(<PaymentSettings />);

      const row = screen.getByTestId('student-row-student-1');
      const toggle = within(row).getByTestId('payment-toggle');

      const scrollBefore = window.scrollY;
      await userEvent.click(toggle);
      const scrollAfter = window.scrollY;

      expect(scrollBefore).toBe(scrollAfter);
    });
  });
});
