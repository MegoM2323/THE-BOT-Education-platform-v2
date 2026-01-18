import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import StudentCreditsHistoryModal from '../StudentCreditsHistoryModal.jsx';
import { useStudentCreditsHistory } from '../../../hooks/useStudentCreditsHistory.js';

vi.mock('../../../hooks/useStudentCreditsHistory.js');
vi.mock('../Modal.jsx', () => ({
  default: ({ isOpen, onClose, title, children, loading }) =>
    isOpen ? (
      <div data-testid="modal" role="dialog">
        <div className="modal-header">
          <h2>{title}</h2>
          <button onClick={onClose} aria-label="Close">×</button>
        </div>
        <div className="modal-body">
          {loading && <div data-testid="modal-spinner">Loading...</div>}
          {children}
        </div>
      </div>
    ) : null,
}));
vi.mock('../Spinner.jsx', () => ({
  default: () => <div data-testid="spinner">Spinner</div>,
}));

describe('StudentCreditsHistoryModal', () => {
  const mockStudent = {
    id: '123',
    full_name: 'Иван Петров',
    credits: 50,
  };

  const mockHistory = [
    {
      id: 1,
      operation_type: 'add',
      amount: 10,
      reason: 'Booking lesson',
      created_at: '2024-01-15T10:30:00Z',
    },
    {
      id: 2,
      operation_type: 'deduct',
      amount: -5,
      reason: 'Lesson cancelled',
      created_at: '2024-01-14T15:45:00Z',
    },
  ];

  const defaultMockHook = {
    history: mockHistory,
    historyLoading: false,
    historyError: null,
    fetchHistory: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useStudentCreditsHistory).mockReturnValue(defaultMockHook);
  });

  describe('rendering - isOpen prop', () => {
    it('should render modal when isOpen is true', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const modal = screen.getByTestId('modal');
      expect(modal).toBeInTheDocument();
    });

    it('should NOT render modal when isOpen is false', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={false}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const modal = screen.queryByTestId('modal');
      expect(modal).not.toBeInTheDocument();
    });

    it('should not render when student prop is null', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={null}
        />
      );

      const modal = screen.queryByTestId('modal');
      expect(modal).not.toBeInTheDocument();
    });

    it('should not render when student prop is undefined', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={undefined}
        />
      );

      const modal = screen.queryByTestId('modal');
      expect(modal).not.toBeInTheDocument();
    });
  });

  describe('modal title', () => {
    it('should display title with student full_name', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText('История кредитов: Иван Петров')).toBeInTheDocument();
    });

    it('should include student name in title for different student', () => {
      const anotherStudent = { id: '456', full_name: 'Мария Иванова', credits: 100 };

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={anotherStudent}
        />
      );

      expect(screen.getByText('История кредитов: Мария Иванова')).toBeInTheDocument();
    });
  });

  describe('balance display', () => {
    it('should display current student balance', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText('50 кредитов')).toBeInTheDocument();
    });

    it('should show loading state when credits is undefined', () => {
      const studentWithoutCredits = { id: '123', full_name: 'Иван Петров' };

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={studentWithoutCredits}
        />
      );

      expect(screen.getByText('Загрузка...')).toBeInTheDocument();
    });

    it('should display different balance values', () => {
      const studentWithDifferentCredits = { id: '123', full_name: 'Иван Петров', credits: 250 };

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={studentWithDifferentCredits}
        />
      );

      expect(screen.getByText('250 кредитов')).toBeInTheDocument();
    });
  });

  describe('operation type filter', () => {
    it('should display filter dropdown with operation type options', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const filterSelect = screen.getByDisplayValue('Все');
      expect(filterSelect).toBeInTheDocument();
    });

    it('should have correct filter options', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const options = screen.getAllByRole('option');
      const optionTexts = options.map(opt => opt.textContent);

      expect(optionTexts).toContain('Все');
      expect(optionTexts).toContain('Начисление');
      expect(optionTexts).toContain('Списание');
      expect(optionTexts).toContain('Возврат');
    });

    it('should filter history when filter is applied', async () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const filterSelect = screen.getByDisplayValue('Все');

      // Change filter to 'add' (Начисление)
      fireEvent.change(filterSelect, { target: { value: 'add' } });

      // Wait for table to update
      await waitFor(() => {
        const transactions = screen.getAllByTestId('transaction');
        // Should only show the 'add' transaction
        expect(transactions).toHaveLength(1);
        expect(transactions[0].textContent).toContain('Начисление');
      });
    });

    it('should show all items when filter is empty', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const filterSelect = screen.getByDisplayValue('Все');
      fireEvent.change(filterSelect, { target: { value: '' } });

      const transactions = screen.getAllByTestId('transaction');
      expect(transactions).toHaveLength(2);
    });
  });

  describe('table structure', () => {
    it('should display table with correct column headers', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText('Дата и время')).toBeInTheDocument();
      expect(screen.getByText('Операция')).toBeInTheDocument();
      expect(screen.getByText('Количество')).toBeInTheDocument();
      expect(screen.getByText('Причина')).toBeInTheDocument();
    });

    it('should display all transactions in table', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const transactions = screen.getAllByTestId('transaction');
      expect(transactions).toHaveLength(2);
    });

    it('should display transaction details correctly', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const transactions = screen.getAllByTestId('transaction');

      // First transaction (add)
      expect(transactions[0].textContent).toContain('Начисление');
      expect(transactions[0].textContent).toContain('+10');

      // Second transaction (deduct)
      expect(transactions[1].textContent).toContain('Списание');
      expect(transactions[1].textContent).toContain('-5');
    });
  });

  describe('loading state', () => {
    it('should show Spinner when historyLoading is true', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        historyLoading: true,
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByTestId('modal-spinner')).toBeInTheDocument();
    });

    it('should not show Spinner when historyLoading is false', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        historyLoading: false,
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.queryByTestId('modal-spinner')).not.toBeInTheDocument();
    });

    it('should disable filter select during loading', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        historyLoading: true,
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const filterSelect = screen.getByDisplayValue('Все');
      expect(filterSelect).toBeDisabled();
    });
  });

  describe('error state', () => {
    it('should display error message when historyError exists', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        historyError: {
          userMessage: 'Ошибка загрузки данных',
          message: 'API Error',
        },
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText(/Ошибка загрузки истории:/)).toBeInTheDocument();
      expect(screen.getByText(/Ошибка загрузки данных/)).toBeInTheDocument();
    });

    it('should display retry button when error occurs', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        historyError: {
          userMessage: 'Ошибка загрузки данных',
          message: 'API Error',
        },
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const retryButton = screen.getByRole('button', { name: /повторить/i });
      expect(retryButton).toBeInTheDocument();
    });

    it('should call fetchHistory when retry button is clicked', async () => {
      const mockFetchHistory = vi.fn();
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        history: [],
        historyLoading: false,
        historyError: {
          userMessage: 'Ошибка загрузки данных',
          message: 'API Error',
        },
        fetchHistory: mockFetchHistory,
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const retryButton = screen.getByRole('button', { name: /повторить/i });
      fireEvent.click(retryButton);

      expect(mockFetchHistory).toHaveBeenCalled();
    });

    it('should not show error when historyError is null', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        historyError: null,
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
  });

  describe('empty history state', () => {
    it('should display empty message when history is empty', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText('История операций пуста')).toBeInTheDocument();
    });

    it('should not display table when history is empty', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const transactions = screen.queryAllByTestId('transaction');
      expect(transactions).toHaveLength(0);
    });

    it('should show filter no-match message when filter results are empty', () => {
      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const filterSelect = screen.getByDisplayValue('Все');
      // Set filter to 'refund' which has no items
      fireEvent.change(filterSelect, { target: { value: 'refund' } });

      expect(
        screen.getByText('Нет операций, соответствующих выбранному фильтру')
      ).toBeInTheDocument();
    });
  });

  describe('close button', () => {
    it('should call onClose when close button is clicked', () => {
      const mockOnClose = vi.fn();

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={mockOnClose}
          student={mockStudent}
        />
      );

      const closeButton = screen.getByLabelText('Close');
      fireEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should not render modal after close', () => {
      const { rerender } = render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByTestId('modal')).toBeInTheDocument();

      rerender(
        <StudentCreditsHistoryModal
          isOpen={false}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.queryByTestId('modal')).not.toBeInTheDocument();
    });
  });

  describe('helper functions', () => {
    it('should call translateReason for transaction reason', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [
          {
            id: 1,
            operation_type: 'add',
            amount: 10,
            reason: 'Booking lesson',
            created_at: '2024-01-15T10:30:00Z',
          },
        ],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      // 'Booking lesson' should be translated to 'Запись на занятие'
      expect(screen.getByText('Запись на занятие')).toBeInTheDocument();
    });

    it('should display getOperationTypeLabel for operation type', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [
          {
            id: 1,
            operation_type: 'add',
            amount: 10,
            reason: 'Test',
            created_at: '2024-01-15T10:30:00Z',
          },
        ],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const transactions = screen.getAllByTestId('transaction');
      expect(transactions[0]).toHaveTextContent('Начисление');
    });

    it('should handle different operation type labels', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [
          {
            id: 1,
            operation_type: 'deduct',
            amount: -5,
            reason: 'Test',
            created_at: '2024-01-15T10:30:00Z',
          },
          {
            id: 2,
            operation_type: 'refund',
            amount: 10,
            reason: 'Test',
            created_at: '2024-01-14T10:30:00Z',
          },
        ],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const transactions = screen.getAllByTestId('transaction');
      expect(transactions[0]).toHaveTextContent('Списание');
      expect(transactions[1]).toHaveTextContent('Возврат');
    });
  });

  describe('transaction amount signs', () => {
    it('should show positive sign for positive amounts', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [
          {
            id: 1,
            operation_type: 'add',
            amount: 10,
            reason: 'Test',
            created_at: '2024-01-15T10:30:00Z',
          },
        ],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText('+10')).toBeInTheDocument();
    });

    it('should show negative sign for negative amounts', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [
          {
            id: 1,
            operation_type: 'deduct',
            amount: -5,
            reason: 'Test',
            created_at: '2024-01-15T10:30:00Z',
          },
        ],
      });

      render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      expect(screen.getByText('-5')).toBeInTheDocument();
    });
  });

  describe('css classes', () => {
    it('should apply correct CSS classes to amount', () => {
      vi.mocked(useStudentCreditsHistory).mockReturnValue({
        ...defaultMockHook,
        history: [
          {
            id: 1,
            operation_type: 'add',
            amount: 10,
            reason: 'Test',
            created_at: '2024-01-15T10:30:00Z',
          },
          {
            id: 2,
            operation_type: 'deduct',
            amount: -5,
            reason: 'Test',
            created_at: '2024-01-14T10:30:00Z',
          },
        ],
      });

      const { container } = render(
        <StudentCreditsHistoryModal
          isOpen={true}
          onClose={vi.fn()}
          student={mockStudent}
        />
      );

      const amounts = container.querySelectorAll('.transaction-amount');
      expect(amounts[0]).toHaveClass('positive');
      expect(amounts[1]).toHaveClass('negative');
    });
  });
});
