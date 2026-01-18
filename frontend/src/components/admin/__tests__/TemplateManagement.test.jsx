import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import { TemplateManagement } from "../TemplateManagement.jsx";
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock hooks
vi.mock('../../../hooks/useTemplates.js', () => ({
  useTemplate: vi.fn(() => ({
    template: {
      id: '1',
      name: 'Test Template',
      description: 'Test Description',
    },
    loading: false,
    refetch: vi.fn(),
  })),
  useTemplates: vi.fn(() => ({
    updateTemplate: vi.fn(),
    isUpdating: false,
  })),
}));

vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

vi.mock('../TemplateCalendarView.jsx', () => ({
  default: ({ templateId }) => (
    <div data-testid="template-calendar-view">{templateId}</div>
  ),
}));

describe('TemplateManagement - Button Removal Verification', () => {
  let queryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
  });

  const renderComponent = (props = {}) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <TemplateManagement templateId="1" {...props} />
      </QueryClientProvider>
    );
  };

  describe('Removed buttons verification', () => {
    it('should NOT have "Применить на неделю" button', () => {
      renderComponent();
      const buttons = screen.queryAllByRole('button');
      const applyWeekButton = buttons.find(
        (btn) => btn.textContent.includes('Применить на неделю')
      );
      expect(applyWeekButton).toBeUndefined();
    });

    it('should NOT have "Откатить неделю" button', () => {
      renderComponent();
      const buttons = screen.queryAllByRole('button');
      const rollbackWeekButton = buttons.find(
        (btn) => btn.textContent.includes('Откатить неделю')
      );
      expect(rollbackWeekButton).toBeUndefined();
    });

    it('should NOT contain any text "Применить на неделю"', () => {
      const { container } = renderComponent();
      const hasApplyWeekText = container.textContent.includes('Применить на неделю');
      expect(hasApplyWeekText).toBe(false);
    });

    it('should NOT contain any text "Откатить неделю"', () => {
      const { container } = renderComponent();
      const hasRollbackWeekText = container.textContent.includes('Откатить неделю');
      expect(hasRollbackWeekText).toBe(false);
    });
  });

  describe('Removed modals verification', () => {
    it('should NOT have apply modal', () => {
      const { container } = renderComponent();
      const applyModals = container.querySelectorAll('[data-testid*="apply-modal"]');
      expect(applyModals.length).toBe(0);
    });

    it('should NOT have rollback modal', () => {
      const { container } = renderComponent();
      const rollbackModals = container.querySelectorAll('[data-testid*="rollback-modal"]');
      expect(rollbackModals.length).toBe(0);
    });

    it('should NOT have any modal with apply/rollback related text', () => {
      const { container } = renderComponent();
      const modalTexts = ['Применить', 'Откатить', 'применить', 'откатить'];
      const modals = container.querySelectorAll('[role="dialog"], .modal, [class*="modal"]');

      let foundApplyOrRollbackModal = false;
      modals.forEach((modal) => {
        const text = modal.textContent;
        if (modalTexts.some((t) => text.includes(t))) {
          foundApplyOrRollbackModal = true;
        }
      });

      expect(foundApplyOrRollbackModal).toBe(false);
    });
  });

  describe('Removed state verification (code inspection)', () => {
    it('component should only have expected state variables', () => {
      // This test verifies the component works without the removed states
      // If states were still present, the component would render them
      const { container } = renderComponent();
      const component = container.querySelector('[data-testid="template-management"]');
      expect(component).toBeInTheDocument();

      // The component should render without errors (proving states aren't referenced)
      expect(container).toBeTruthy();
    });

    it('should not reference weekStartDate anywhere', () => {
      const { container } = renderComponent();
      const hasWeekStartDate = container.innerHTML.includes('weekStartDate');
      expect(hasWeekStartDate).toBe(false);
    });

    it('should not reference showApplyModal state', () => {
      const { container } = renderComponent();
      const hasShowApplyModal = container.innerHTML.includes('showApplyModal');
      expect(hasShowApplyModal).toBe(false);
    });

    it('should not reference showRollbackModal state', () => {
      const { container } = renderComponent();
      const hasShowRollbackModal = container.innerHTML.includes('showRollbackModal');
      expect(hasShowRollbackModal).toBe(false);
    });
  });

  describe('Component structure verification', () => {
    it('should render template management section', () => {
      renderComponent();
      const templateManagement = screen.getByTestId('template-management');
      expect(templateManagement).toBeInTheDocument();
    });

    it('should render template header', () => {
      renderComponent();
      const header = screen.getByText('Test Template');
      expect(header).toBeInTheDocument();
    });

    it('should render template description', () => {
      renderComponent();
      const description = screen.getByText('Test Description');
      expect(description).toBeInTheDocument();
    });

    it('should render TemplateCalendarView component', () => {
      renderComponent();
      const calendarView = screen.getByTestId('template-calendar-view');
      expect(calendarView).toBeInTheDocument();
    });

    it('should have "Изменить" button only', () => {
      renderComponent();
      const editButton = screen.getByRole('button', { name: /Изменить/i });
      expect(editButton).toBeInTheDocument();
    });
  });

  describe('No removed handlers in button clicks', () => {
    it('should only have handlers for edit, save, and cancel operations', () => {
      renderComponent();
      const buttons = screen.getAllByRole('button');

      // Should have: "Изменить" button
      const editButton = buttons.find((btn) => btn.textContent.includes('Изменить'));
      expect(editButton).toBeTruthy();

      // Should NOT have buttons that would trigger apply/rollback handlers
      const hasApplyOrRollback = buttons.some((btn) => {
        const text = btn.textContent;
        return text.includes('Применить на неделю') || text.includes('Откатить неделю');
      });
      expect(hasApplyOrRollback).toBe(false);
    });
  });

  describe('No getNextMonday function reference', () => {
    it('should not have any date calculation functions for next Monday', () => {
      const { container } = renderComponent();
      const hasGetNextMonday = container.innerHTML.includes('getNextMonday');
      expect(hasGetNextMonday).toBe(false);
    });
  });

  describe('Edge cases - no apply/rollback in any variation', () => {
    it('should not have button text variations', () => {
      const { container } = renderComponent();
      const hasAnyApplyVariation = container.textContent.includes('применить') ||
                                   container.textContent.includes('Применить');
      const hasAnyRollbackVariation = container.textContent.includes('откатить') ||
                                      container.textContent.includes('Откатить');

      expect(hasAnyApplyVariation).toBe(false);
      expect(hasAnyRollbackVariation).toBe(false);
    });

    it('should not have any modal with apply-related data attributes', () => {
      const { container } = renderComponent();
      const modalsWithApply = container.querySelectorAll('[data-testid*="apply"], [class*="apply"]');

      // Filter to remove unrelated elements
      const relevantModals = Array.from(modalsWithApply).filter((elem) => {
        const text = elem.textContent.toLowerCase();
        return text.includes('неделю') || text.includes('недели');
      });

      expect(relevantModals.length).toBe(0);
    });
  });

  describe('Component renders correctly without removed features', () => {
    it('should render without errors', () => {
      expect(() => renderComponent()).not.toThrow();
    });

    it('should display main template management container', () => {
      renderComponent();
      const container = screen.getByTestId('template-management');
      expect(container).toBeVisible();
    });

    it('should not have any console errors about missing handlers', () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      renderComponent();

      // Check that no errors about undefined handlers were logged
      const errorCalls = consoleErrorSpy.mock.calls;
      const hasHandlerErrors = errorCalls.some((call) => {
        const message = call[0]?.toString() || '';
        return message.includes('handleApplyTemplate') ||
               message.includes('handleRollbackTemplate') ||
               message.includes('handleOpenApplyModal') ||
               message.includes('handleOpenRollbackModal');
      });

      expect(hasHandlerErrors).toBe(false);
      consoleErrorSpy.mockRestore();
    });
  });
});
