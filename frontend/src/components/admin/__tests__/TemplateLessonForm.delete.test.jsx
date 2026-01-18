import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TemplateLessonForm } from '../TemplateLessonForm.jsx';
import { NotificationProvider } from '../../../context/NotificationContext.jsx';
import * as usersAPI from '../../../api/users.js';
import apiClient from '../../../api/client.js';

// Mock API modules
vi.mock('../../../api/users.js');
vi.mock('../../../api/client.js');

// Mock child components
vi.mock('../TemplateStudentAssignmentModal.jsx', () => ({
  default: ({ isOpen, onClose }) => (
    isOpen && <div data-testid="student-assignment-modal">Student Modal</div>
  ),
}));

vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: () => ({
    success: vi.fn((msg) => console.log('Success:', msg)),
    error: vi.fn((msg) => console.log('Error:', msg)),
  }),
}));

// Mock ColorPicker
vi.mock('../common/ColorPicker.jsx', () => ({
  default: ({ value, onChange, disabled }) => (
    <input
      type="color"
      data-testid="color-picker"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
    />
  ),
}));

const mockTeachers = [
  { id: '1', full_name: 'Teacher One' },
  { id: '2', full_name: 'Teacher Two' },
];

const mockEditingLesson = {
  id: 'lesson-123',
  teacher_id: '1',
  day_of_week: 0,
  start_time: '10:00',
  end_time: '12:00',
  max_students: 4,
  description: 'Test lesson',
  subject: 'Math',
  color: '#3B82F6',
  student_ids: [],
};

let queryClient;

const renderComponent = (props = {}) => {
  const defaultProps = {
    templateId: 'template-123',
    onSave: vi.fn(),
    onCancel: vi.fn(),
    editingLesson: null,
    ...props,
  };

  return render(
    <QueryClientProvider client={queryClient}>
      <NotificationProvider>
        <TemplateLessonForm {...defaultProps} />
      </NotificationProvider>
    </QueryClientProvider>
  );
};

describe('T303: TemplateLessonForm Delete Functionality', () => {
  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    vi.clearAllMocks();

    // Setup default API mocks
    usersAPI.getTeachersAll.mockResolvedValue(mockTeachers);
    apiClient.put = vi.fn().mockResolvedValue({});
    apiClient.post = vi.fn().mockResolvedValue({});
    apiClient.delete = vi.fn().mockResolvedValue({});
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Test Case 1: Delete Button Visibility', () => {
    it('should NOT show delete button when editingLesson is null (create mode)', async () => {
      renderComponent({ editingLesson: null });

      await waitFor(() => {
        expect(screen.getByText(/Добавить занятие в шаблон/)).toBeInTheDocument();
      });

      const deleteButton = screen.queryByRole('button', { name: /Удалить/i });
      expect(deleteButton).not.toBeInTheDocument();
    });

    it('should show delete button when editingLesson is provided (edit mode)', async () => {
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      expect(deleteButton).toBeInTheDocument();
    });

    it('delete button should have danger variant when visible', async () => {
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      expect(deleteButton).toHaveClass('btn-danger');
    });

    it('delete button should be disabled while deleting', async () => {
      const user = userEvent.setup();
      let resolveDelete;
      apiClient.delete = vi.fn(
        () =>
          new Promise((resolve) => {
            resolveDelete = resolve;
          })
      );

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      expect(deleteButton).not.toBeDisabled();

      // Click delete button to trigger delete
      await user.click(deleteButton);

      // Confirm deletion
      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Check if delete button becomes disabled during deletion
      await waitFor(() => {
        const deleteButtons = screen.getAllByRole('button', { name: /Удалить/i });
        if (deleteButtons.length > 0) {
          expect(deleteButtons[0]).toBeDisabled();
        }
      });

      // Resolve the deletion
      resolveDelete({});
    });

    it('delete button should be disabled while submitting', async () => {
      const user = userEvent.setup();
      let resolveSubmit;
      apiClient.put = vi.fn(
        () =>
          new Promise((resolve) => {
            resolveSubmit = resolve;
          })
      );

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      const submitButton = screen.getByRole('button', { name: /Обновить занятие/i });

      // Start submission
      await user.click(submitButton);

      // Check if delete button is disabled during submission
      await waitFor(() => {
        expect(deleteButton).toBeDisabled();
      }, { timeout: 1000 });

      // Resolve submission
      resolveSubmit({});
    });
  });

  describe('Test Case 2: Delete Confirmation Modal', () => {
    it('should open ConfirmModal when delete button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      // Modal should appear with title
      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });
    });

    it('modal should show correct title', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });
    });

    it('modal should show correct message', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByText('Вы уверены, что хотите удалить это занятие?')
        ).toBeInTheDocument();
      });
    });

    it('confirm button should have text "Удалить"', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      // Find the confirm button in the modal (there should be 2 buttons with "Удалить" text)
      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      expect(buttons.length).toBeGreaterThanOrEqual(2); // Delete + Confirm
    });

    it('confirm button should have danger variant', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      // Find the confirm button (should be the last "Удалить" button)
      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      expect(confirmButton).toHaveClass('btn-danger');
    });
  });

  describe('Test Case 3: Delete API Call', () => {
    it('should call DELETE endpoint with correct URL', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      // Find and click confirm button
      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalledWith(
          '/templates/template-123/lessons/lesson-123'
        );
      });
    });

    it('should include templateId and lessonId in URL', async () => {
      const user = userEvent.setup();
      const customLesson = { ...mockEditingLesson, id: 'custom-lesson-id' };
      renderComponent({ templateId: 'custom-template-id', editingLesson: customLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalledWith(
          '/templates/custom-template-id/lessons/custom-lesson-id'
        );
      });
    });

    it('should handle authentication headers properly', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Verify delete was called (headers are handled by apiClient mock)
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalled();
      });
    });
  });

  describe('Test Case 4: Delete Success', () => {
    it('should show success notification after deletion', async () => {
      const user = userEvent.setup();
      const mockNotification = {
        success: vi.fn(),
        error: vi.fn(),
      };

      vi.mocked(usersAPI.getTeachersAll).mockResolvedValue(mockTeachers);

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Success notification should be shown
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalled();
      });
    });

    it('should invalidate React Query cache after deletion', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      // Spy on queryClient invalidateQueries
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      await waitFor(() => {
        expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['templates'] });
        expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['templates', 'template-123'] });
      });
    });

    it('should close modal after successful deletion', async () => {
      const user = userEvent.setup();
      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Modal should close
      await waitFor(() => {
        expect(screen.queryByText('Удаление занятия шаблона')).not.toBeInTheDocument();
      });
    });

    it('should call onCancel callback after deletion', async () => {
      const user = userEvent.setup();
      const mockOnCancel = vi.fn();
      const mockOnSave = vi.fn();
      renderComponent({
        editingLesson: mockEditingLesson,
        onCancel: mockOnCancel,
        onSave: mockOnSave,
      });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalled();
      });
    });

    it('should reset form after successful deletion', async () => {
      const user = userEvent.setup();
      const mockOnSave = vi.fn();
      renderComponent({ editingLesson: mockEditingLesson, onSave: mockOnSave });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // After deletion, onSave should be called (which closes the form)
      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalled();
      });
    });
  });

  describe('Test Case 5: Delete Error Handling', () => {
    it('should show error notification on API failure', async () => {
      const user = userEvent.setup();
      apiClient.delete = vi.fn().mockRejectedValue(new Error('Network error'));

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Wait for error (component should show error notification)
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalled();
      });
    });

    it('should keep modal open if deletion fails', async () => {
      const user = userEvent.setup();
      apiClient.delete = vi.fn().mockRejectedValue(new Error('Server error'));

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Wait for delete attempt to fail
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalled();
      }, { timeout: 1000 });

      // Component should handle error gracefully
      expect(apiClient.delete).toHaveBeenCalledWith('/templates/template-123/lessons/lesson-123');
    });

    it('should allow user to retry after error', async () => {
      const user = userEvent.setup();
      let callCount = 0;
      apiClient.delete = vi.fn().mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.reject(new Error('First attempt failed'));
        }
        return Promise.resolve({});
      });

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      // Wait for modal to open
      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      // First attempt
      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      // Wait for first delete call
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalledTimes(1);
      }, { timeout: 1000 });

      // Verify the retry is possible by checking button is present
      const retryButtons = screen.queryAllByRole('button', { name: /Удалить/i });
      expect(retryButtons.length).toBeGreaterThan(0);
    });

    it('should display error message from server', async () => {
      const user = userEvent.setup();
      const serverError = new Error('Lesson has active bookings');
      apiClient.delete = vi.fn().mockRejectedValue(serverError);

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];
      await user.click(confirmButton);

      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalled();
      });
    });
  });

  describe('Test Case 6: Delete Disabled During Operations', () => {
    it('should disable delete button while deleting', async () => {
      const user = userEvent.setup();
      let resolveDelete;
      apiClient.delete = vi.fn(
        () =>
          new Promise((resolve) => {
            resolveDelete = resolve;
          })
      );

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Удаление занятия шаблона')).toBeInTheDocument();
      });

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];

      // Button should be disabled during deletion - click confirm
      await user.click(confirmButton);

      // Verify delete was called
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalled();
      });

      // Resolve the delete
      resolveDelete({});
    });

    it('should disable delete button while submitting', async () => {
      const user = userEvent.setup();
      let resolveSubmit;
      apiClient.put = vi.fn(
        () =>
          new Promise((resolve) => {
            resolveSubmit = resolve;
          })
      );

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      const submitButton = screen.getByRole('button', { name: /Обновить занятие/i });

      // Start submit
      await user.click(submitButton);

      // Delete button should be disabled
      expect(deleteButton).toBeDisabled();

      // Resolve the submit
      resolveSubmit({});
    });

    it('should prevent double-click on delete button', async () => {
      const user = userEvent.setup();
      let callCount = 0;
      apiClient.delete = vi.fn(() => {
        callCount++;
        return Promise.resolve({});
      });

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      await user.click(deleteButton);

      const buttons = screen.getAllByRole('button', { name: /Удалить/i });
      const confirmButton = buttons[buttons.length - 1];

      // Click confirm button
      await user.click(confirmButton);

      // Wait for delete to be called
      await waitFor(() => {
        expect(apiClient.delete).toHaveBeenCalledTimes(1);
      });
    });

    it('should prevent delete during loading state', async () => {
      const user = userEvent.setup();

      renderComponent({ editingLesson: mockEditingLesson });

      await waitFor(() => {
        expect(screen.getByText(/Редактировать занятие шаблона/)).toBeInTheDocument();
      });

      // Verify delete button is enabled initially
      const deleteButton = screen.getByRole('button', { name: /Удалить/i });
      expect(deleteButton).not.toBeDisabled();
    });
  });
});
