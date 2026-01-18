import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { TemplateLessonEditorModal } from '../TemplateLessonEditorModal.jsx';
import * as usersAPI from '../../../api/users.js';
import apiClient from '../../../api/client.js';

vi.mock('../../../api/users.js');
vi.mock('../../../api/client.js');
vi.mock('../../common/Modal.jsx', () => ({
  default: ({ children, isOpen }) => isOpen ? <div data-testid="modal">{children}</div> : null,
}));
vi.mock('../../common/Button.jsx', () => ({
  default: ({ children, onClick, type = 'button', disabled = false }) => (
    <button type={type} onClick={onClick} disabled={disabled}>{children}</button>
  ),
}));
vi.mock('../../common/Spinner.jsx', () => ({
  default: () => <div data-testid="spinner">Loading...</div>,
}));
vi.mock('../../common/ColorPicker.jsx', () => ({
  default: ({ value, onChange, disabled }) => (
    <div data-testid="color-picker">
      <input
        data-testid="color-input"
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
      />
    </div>
  ),
}));

const mockNotification = {
  success: vi.fn(),
  error: vi.fn(),
};

vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: () => mockNotification,
}));

const mockTeachers = [
  { id: '1', full_name: 'Учитель 1' },
];

const mockStudents = [
  { id: '1', full_name: 'Студент 1', email: 'student1@example.com' },
];

describe('TemplateLessonEditorModal - Autosave Color', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNotification.success.mockClear();
    mockNotification.error.mockClear();

    usersAPI.getTeachersAll.mockResolvedValue(mockTeachers);
    usersAPI.getStudentsAll.mockResolvedValue(mockStudents);
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('1. Debounce Functionality', () => {
    it('should send API request with final color after debounce period', async () => {
      apiClient.patch.mockResolvedValue({ color: '#8B5CF6' });

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');

      // Make multiple rapid color changes
      fireEvent.change(colorInput, { target: { value: '#EF4444' } });
      fireEvent.change(colorInput, { target: { value: '#10B981' } });
      fireEvent.change(colorInput, { target: { value: '#F59E0B' } });
      fireEvent.change(colorInput, { target: { value: '#8B5CF6' } });

      // Wait for debounce to trigger - only final color should be saved
      await waitFor(() => {
        expect(apiClient.patch).toHaveBeenCalledTimes(1);
      }, { timeout: 1000 });

      expect(apiClient.patch).toHaveBeenCalledWith(
        '/templates/template-1/lessons/lesson-1',
        { color: '#8B5CF6' }
      );
    });
  });

  describe('2. Autosave Only in Edit Mode', () => {

    it('should disable ColorPicker in create mode', async () => {
      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="create"
          templateId="template-1"
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      expect(colorInput).toBeDisabled();
    });

    it('should enable ColorPicker in edit mode', async () => {
      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      expect(colorInput).not.toBeDisabled();
    });
  });

  describe('3. API Request Verification', () => {
    it('should send only color in payload', async () => {
      apiClient.patch.mockResolvedValue({ color: '#10B981' });

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      fireEvent.change(colorInput, { target: { value: '#10B981' } });

      await waitFor(() => {
        expect(apiClient.patch).toHaveBeenCalled();
      }, { timeout: 1000 });

      const callArgs = apiClient.patch.mock.calls[0];
      expect(callArgs[1]).toEqual({ color: '#10B981' });
      expect(Object.keys(callArgs[1])).toEqual(['color']);
    });

    it('should show success notification on successful save', async () => {
      apiClient.patch.mockResolvedValue({ color: '#EF4444' });

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      fireEvent.change(colorInput, { target: { value: '#EF4444' } });

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalledWith('Цвет занятия сохранён');
      });
    });

    it('should show error notification on failed save', async () => {
      apiClient.patch.mockRejectedValue(new Error('Network error'));

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      fireEvent.change(colorInput, { target: { value: '#EF4444' } });

      await waitFor(() => {
        expect(mockNotification.error).toHaveBeenCalledWith('Не удалось сохранить цвет');
      });
    });
  });

  describe('4. Cleanup on Unmount', () => {
    it('should properly cleanup when component unmounts', async () => {
      apiClient.patch.mockResolvedValue({ color: '#EF4444' });

      const { unmount } = render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      // Unmount should not cause any errors
      expect(() => {
        unmount();
      }).not.toThrow();
    });
  });

  describe('5. UI Loading Indicators', () => {
    it('should show saving indicator text during save', async () => {
      apiClient.patch.mockImplementation(() => new Promise(() => {}));

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      fireEvent.change(colorInput, { target: { value: '#EF4444' } });

      await waitFor(() => {
        const indicator = screen.queryByText(/сохраняется/i);
        expect(indicator).toBeTruthy();
      });
    });

    it('should hide saving indicator after successful save', async () => {
      apiClient.patch.mockResolvedValue({ color: '#EF4444' });

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      fireEvent.change(colorInput, { target: { value: '#EF4444' } });

      await waitFor(() => {
        const indicator = screen.queryByText(/сохраняется/i);
        expect(indicator).toBeFalsy();
      });
    });

    it('should hide saving indicator after failed save', async () => {
      apiClient.patch.mockRejectedValue(new Error('Network error'));

      render(
        <TemplateLessonEditorModal
          isOpen={true}
          mode="edit"
          templateId="template-1"
          lessonId="lesson-1"
          prefilledData={{
            day_of_week: 1,
            start_time: '10:00',
            teacher_id: '1',
            max_students: 1,
            student_ids: [],
            color: '#3b82f6',
          }}
          onClose={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.queryByTestId('color-input')).toBeInTheDocument();
      });

      const colorInput = screen.getByTestId('color-input');
      fireEvent.change(colorInput, { target: { value: '#EF4444' } });

      await waitFor(() => {
        const indicator = screen.queryByText(/сохраняется/i);
        expect(indicator).toBeFalsy();
      });
    });
  });

});
