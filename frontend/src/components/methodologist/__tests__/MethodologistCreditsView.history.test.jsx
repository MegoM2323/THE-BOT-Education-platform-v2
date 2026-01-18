import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { renderWithProviders } from '../../../test/test-utils.jsx';
import { MethodologistCreditsView } from '../MethodologistCreditsView.jsx';
import * as usersAPI from '../../../api/users.js';
import * as creditsAPI from '../../../api/credits.js';
import { useNotification } from '../../../hooks/useNotification.js';
import { useStudentCreditsHistory } from '../../../hooks/useStudentCreditsHistory.js';

vi.mock('../../../api/users.js');
vi.mock('../../../api/credits.js');
vi.mock('../../../hooks/useNotification.js');
vi.mock('../../../hooks/useStudentCreditsHistory.js');
vi.mock('../../common/StudentCreditsHistoryModal.jsx', () => ({
  default: ({ isOpen, onClose, student }) =>
    isOpen && student ? (
      <div data-testid="history-modal" role="dialog">
        <div>История кредитов: {student.full_name}</div>
        <button onClick={onClose} data-testid="close-modal">
          Закрыть
        </button>
      </div>
    ) : null,
}));

describe('MethodologistCreditsView - History Modal Integration', () => {
  const mockStudents = [
    { id: '1', full_name: 'Анна Иванова', email: 'anna@example.com', credits: 50 },
    { id: '2', full_name: 'Борис Петров', email: 'boris@example.com', credits: 100 },
    { id: '3', full_name: 'Виктор Сидоров', email: 'viktor@example.com', credits: 25 },
  ];

  const mockNotification = {
    error: vi.fn(),
    success: vi.fn(),
  };

  const defaultMockHistoryHook = {
    history: [],
    historyLoading: false,
    historyError: null,
    fetchHistory: vi.fn(),
    refresh: vi.fn(),
    getUserBalance: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useNotification.mockReturnValue(mockNotification);
    usersAPI.getStudentsAll.mockResolvedValue(mockStudents);
    creditsAPI.getUserCredits.mockImplementation((id) => {
      const student = mockStudents.find(s => s.id === id);
      return Promise.resolve({ balance: student?.credits || 0 });
    });
    vi.mocked(useStudentCreditsHistory).mockReturnValue(defaultMockHistoryHook);
  });

  describe('history button rendering', () => {
    it('should render "history" button for each student in table', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(3);
      });
    });

    it('should have correct aria-label on history button', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        historyButtons.forEach((btn) => {
          expect(btn).toHaveAttribute('aria-label', 'Просмотреть историю кредитов');
        });
      });
    });

    it('should display history button for only visible students', async () => {
      usersAPI.getStudentsAll.mockResolvedValue([mockStudents[0]]);

      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(1);
      });
    });

    it('should show correct button styling', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      historyButtons.forEach((btn) => {
        expect(btn).toHaveClass('history-action-btn');
      });
    });
  });

  describe('history modal opening', () => {
    it('should open modal when history button is clicked', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });
    });

    it('should NOT open modal initially', () => {
      renderWithProviders(<MethodologistCreditsView />);

      expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
    });

    it('should pass correct student object to modal', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Анна Иванова')).toBeInTheDocument();
      });
    });

    it('should pass correct student object for different students', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(3);
      });

      // Test second student
      let historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[1]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Борис Петров')).toBeInTheDocument();
      });

      // Close
      let closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // Test third student
      historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[2]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Виктор Сидоров')).toBeInTheDocument();
      });
    });

    it('should include student full_name in modal', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByText(/История кредитов: /)).toBeInTheDocument();
      });
    });

  });

  describe('modal closing', () => {
    it('should close modal when close button is clicked', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });

      const closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });
    });

    it('should be able to open modal again after closing', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });

      let closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // Open again
      const newHistoryButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(newHistoryButtons[0]);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });
    });

    it('should reset selectedStudentForHistory on close', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Анна Иванова')).toBeInTheDocument();
      });

      const closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // Open again with different student
      const newHistoryButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(newHistoryButtons[1]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Борис Петров')).toBeInTheDocument();
      });
    });
  });

  describe('StudentCreditsHistoryModal props', () => {
    it('should pass isOpen={true} when modal should be visible', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });
    });

    it('should pass isOpen={false} when modal should be hidden', () => {
      renderWithProviders(<MethodologistCreditsView />);

      expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
    });

    it('should pass onClose callback to modal', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });

      const closeButton = screen.getByTestId('close-modal');
      expect(closeButton).toBeInTheDocument();
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });
    });

    it('should pass student object with correct structure to modal', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        const modal = screen.getByTestId('history-modal');
        expect(modal.textContent).toContain('Анна Иванова');
      });
    });

    it('should pass null student when modal should be hidden', () => {
      renderWithProviders(<MethodologistCreditsView />);

      expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
    });

  });

  describe('multiple modal interactions', () => {
    it('should switch between different students history', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(3);
      });

      // Open first student
      let historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Анна Иванова')).toBeInTheDocument();
      });

      // Close first
      let closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // Open second student
      historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[1]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Борис Петров')).toBeInTheDocument();
      });

      // Close second
      closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // Open third student
      historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[2]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Виктор Сидоров')).toBeInTheDocument();
      });
    });

    it('should close modal without affecting other UI elements', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });

      // Close modal
      const closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // All students should still be visible
      expect(screen.getByText('Анна Иванова')).toBeInTheDocument();
      expect(screen.getByText('Борис Петров')).toBeInTheDocument();
      expect(screen.getByText('Виктор Сидоров')).toBeInTheDocument();
    });
  });

  describe('history button positioning', () => {
    it('should display history button in actions column', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(3);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      historyButtons.forEach((btn) => {
        // History button should have correct class
        expect(btn).toHaveClass('history-action-btn');
      });
    });

    it('should be the only button in actions column (no add/deduct)', async () => {
      const { container } = renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const rows = container.querySelectorAll('tbody tr');
      expect(rows.length).toBe(3);

      // Each row should have only 1 action button (history only)
      rows.forEach((row) => {
        const buttons = row.querySelectorAll('button');
        expect(buttons.length).toBe(1);
      });
    });

    it('should not show add/deduct buttons', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const addButtons = screen.queryAllByRole('button', { name: /добавить/i });
      const deductButtons = screen.queryAllByRole('button', { name: /списать/i });

      expect(addButtons).toHaveLength(0);
      expect(deductButtons).toHaveLength(0);
    });
  });

  describe('modal isolation', () => {
    it('should only show one modal at a time', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(3);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        const modals = screen.getAllByTestId('history-modal');
        expect(modals).toHaveLength(1);
      });
    });

    it('should not have credits modal', async () => {
      const { container } = renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      // Should not have any modal for credits operations
      const creditsModals = container.querySelectorAll('[data-testid="credits-form"]');
      expect(creditsModals).toHaveLength(0);
    });
  });

  describe('with empty students list', () => {
    it('should not show history buttons when students list is empty', async () => {
      usersAPI.getStudentsAll.mockResolvedValue([]);

      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        expect(screen.getByText('Студенты не найдены')).toBeInTheDocument();
      });

      const historyButtons = screen.queryAllByTestId('view-history-button');
      expect(historyButtons).toHaveLength(0);
    });
  });

  describe('sorting does not affect history modal', () => {
    it('should open history modal for sorted student', async () => {
      renderWithProviders(<MethodologistCreditsView />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(3);
      });

      // Click on balance header to sort
      const balanceHeader = screen.getByText(/Баланс/);
      fireEvent.click(balanceHeader);

      await waitFor(() => {
        // Table should still be rendered after sort
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(3);
      });

      // Now click on first visible student's history button
      const historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });
    });
  });
});
