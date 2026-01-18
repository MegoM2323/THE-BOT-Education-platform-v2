import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { renderWithProviders } from '../../../test/test-utils.jsx';
import { CreditsManagement } from '../CreditsManagement.jsx';
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

describe('CreditsManagement - History Modal Integration', () => {
  const mockStudents = [
    { id: '1', full_name: 'Иван Петров', email: 'ivan@example.com', credits: 50 },
    { id: '2', full_name: 'Мария Иванова', email: 'maria@example.com', credits: 100 },
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
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(2);
      });
    });

    it('should have correct aria-label on history button', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        historyButtons.forEach((btn) => {
          expect(btn).toHaveAttribute('aria-label', 'Просмотреть историю кредитов');
        });
      });
    });

    it('should display history button for only visible students', async () => {
      usersAPI.getStudentsAll.mockResolvedValue([mockStudents[0]]);

      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(1);
      });
    });
  });

  describe('history modal opening', () => {
    it('should open modal when history button is clicked', async () => {
      renderWithProviders(
        <CreditsManagement />);

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
      renderWithProviders(
        <CreditsManagement />);

      expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
    });

    it('should pass correct student object to modal', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Иван Петров')).toBeInTheDocument();
      });
    });

    it('should pass correct student object for second student', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(2);
      });

      const secondHistoryButton = screen.getAllByTestId('view-history-button')[1];
      fireEvent.click(secondHistoryButton);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Мария Иванова')).toBeInTheDocument();
      });
    });

    it('should include student full_name in modal', async () => {
      renderWithProviders(
        <CreditsManagement />);

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
      renderWithProviders(
        <CreditsManagement />);

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
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        expect(screen.getByTestId('history-modal')).toBeInTheDocument();
      });

      const closeButton = screen.getByTestId('close-modal');
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
  });

  describe('StudentCreditsHistoryModal props', () => {
    it('should pass isOpen={true} when modal should be visible', async () => {
      const { container } = renderWithProviders(
        <CreditsManagement />);

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
      renderWithProviders(
        <CreditsManagement />);

      // Modal should not be rendered initially
      expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
    });

    it('should pass onClose callback to modal', async () => {
      renderWithProviders(
        <CreditsManagement />);

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

      // Modal should close
      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });
    });

    it('should pass student object with correct structure to modal', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const firstHistoryButton = screen.getAllByTestId('view-history-button')[0];
      fireEvent.click(firstHistoryButton);

      await waitFor(() => {
        const modal = screen.getByTestId('history-modal');
        expect(modal.textContent).toContain('Иван Петров');
      });
    });

    it('should pass null student when modal should be hidden', () => {
      renderWithProviders(
        <CreditsManagement />);

      // No modal should be visible initially
      expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
    });
  });

  describe('multiple modal interactions', () => {
    it('should switch between different students history', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(2);
      });

      // Open first student
      let historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Иван Петров')).toBeInTheDocument();
      });

      // Close first
      const closeButton = screen.getByTestId('close-modal');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('history-modal')).not.toBeInTheDocument();
      });

      // Open second student
      historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[1]);

      await waitFor(() => {
        expect(screen.getByText('История кредитов: Мария Иванова')).toBeInTheDocument();
      });
    });

    it('should close modal without affecting other UI elements', async () => {
      renderWithProviders(
        <CreditsManagement />);

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

      // All action buttons should still be present
      const addButtons = screen.getAllByTestId('add-credits-button');
      expect(addButtons.length).toBeGreaterThan(0);
    });
  });

  describe('history button positioning', () => {
    it('should display history button in actions column', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons).toHaveLength(2);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      historyButtons.forEach((btn) => {
        // History button should have correct class
        expect(btn).toHaveClass('credits-action-btn');
      });
    });

    it('should be in correct order with add/deduct buttons', async () => {
      const { container } = renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBeGreaterThan(0);
      });

      const rows = container.querySelectorAll('tbody tr');
      expect(rows.length).toBe(2);

      // Each row should have 3 action buttons
      rows.forEach((row) => {
        const buttons = row.querySelectorAll('button');
        expect(buttons.length).toBe(3);
      });
    });
  });

  describe('modal isolation', () => {
    it('should only show one modal at a time', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const historyButtons = screen.getAllByTestId('view-history-button');
        expect(historyButtons.length).toBe(2);
      });

      const historyButtons = screen.getAllByTestId('view-history-button');
      fireEvent.click(historyButtons[0]);

      await waitFor(() => {
        const modals = screen.getAllByTestId('history-modal');
        expect(modals).toHaveLength(1);
      });
    });

    it('should not interfere with credits modal', async () => {
      renderWithProviders(
        <CreditsManagement />);

      await waitFor(() => {
        const addButtons = screen.getAllByTestId('add-credits-button');
        expect(addButtons.length).toBeGreaterThan(0);
      });

      const addButton = screen.getAllByTestId('add-credits-button')[0];
      fireEvent.click(addButton);

      // Credits modal might open but history modal should not
      const historyModals = screen.queryAllByTestId('history-modal');
      expect(historyModals).toHaveLength(0);
    });
  });
});
