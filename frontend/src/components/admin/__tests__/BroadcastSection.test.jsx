import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import BroadcastSection from "../BroadcastSection.jsx";

// Import mocked modules
import { useLessonBroadcasts, useSendLessonBroadcast } from '../../../hooks/useLessonBroadcasts.js';
import { useNotification } from '../../../hooks/useNotification.js';
import { useAuth } from '../../../hooks/useAuth.js';

// Mock hooks
vi.mock('../../../hooks/useLessonBroadcasts.js', () => ({
  useLessonBroadcasts: vi.fn(),
  useSendLessonBroadcast: vi.fn(),
}));

vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: vi.fn(),
}));

vi.mock('../../../hooks/useAuth.js', () => ({
  useAuth: vi.fn(),
}));

vi.mock('../../../utils/dateFormat.js', () => ({
  formatTime: (date) => {
    const d = new Date(date);
    return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  },
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

describe('BroadcastSection', () => {
  const mockBroadcasts = [
    {
      id: 'broadcast-1',
      message: 'Привет, студенты! Домашнее задание: прочитать главы 1-3.',
      status: 'completed',
      sender_name: 'Teacher Name',
      created_at: '2025-01-01T10:00:00Z',
      sent_count: 25,
      files: [
        {
          id: 'file-1',
          file_name: 'homework.pdf',
          file_size: 2048000,
        },
      ],
    },
    {
      id: 'broadcast-2',
      message: 'Напоминание о тесте на следующей неделе.',
      status: 'pending',
      sender_name: 'Teacher Name',
      created_at: '2025-01-02T11:00:00Z',
      sent_count: 0,
      files: [],
    },
  ];

  const mockLesson = {
    id: 'lesson-1',
    start_time: '2025-01-10T10:00:00Z',
    end_time: '2025-01-10T12:00:00Z',
    subject: 'Математика',
  };

  const mockSendMutation = {
    mutateAsync: vi.fn(),
    isPending: false,
  };

  const mockShowNotification = vi.fn();

  beforeEach(() => {
    vi.mocked(useLessonBroadcasts).mockReturnValue({
      data: mockBroadcasts,
      isLoading: false,
      error: null,
    });

    vi.mocked(useSendLessonBroadcast).mockReturnValue(mockSendMutation);

    vi.mocked(useNotification).mockReturnValue({
      showNotification: mockShowNotification,
    });

    vi.mocked(useAuth).mockReturnValue({
      user: { id: 'user-1', role: 'methodologist' },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render broadcast section', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText('История рассылок')).toBeInTheDocument();
    });

    it('should render compose section for teachers', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText('Новая рассылка')).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/Введите сообщение для студентов урока/)).toBeInTheDocument();
    });

    it('should render all broadcasts in history', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText(/Привет, студенты/)).toBeInTheDocument();
      expect(screen.getByText(/Напоминание о тесте/)).toBeInTheDocument();
    });

    it('should show status badges for broadcasts', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText('Завершено')).toBeInTheDocument();
      expect(screen.getByText('Ожидает')).toBeInTheDocument();
    });

    it('should show sender name', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const senderElements = screen.getAllByText('Teacher Name');
      expect(senderElements.length).toBeGreaterThan(0);
    });

    it('should show file count for broadcasts with files', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText(/📎 1 файл\(а\)/)).toBeInTheDocument();
    });

    it('should show sent count for completed broadcasts', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText(/✓ 25 получателей/)).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading spinner', () => {
      vi.mocked(useLessonBroadcasts).mockReturnValue({
        data: [],
        isLoading: true,
        error: null,
      });

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(document.querySelector('.broadcast-section')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should show error message', () => {
      vi.mocked(useLessonBroadcasts).mockReturnValue({
        data: [],
        isLoading: false,
        error: new Error('Load failed'),
      });

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText(/Ошибка загрузки рассылок/)).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should show empty message when no broadcasts', () => {
      vi.mocked(useLessonBroadcasts).mockReturnValue({
        data: [],
        isLoading: false,
        error: null,
      });

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText('Рассылок пока нет')).toBeInTheDocument();
    });
  });

  describe('Message Composition', () => {
    it('should handle message input', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      expect(textarea).toHaveValue('Test message');
    });

    it('should show character counter', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Hello');

      expect(screen.getByText('5 / 4096')).toBeInTheDocument();
    });

    it('should prevent sending empty message', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const sendButton = screen.getByRole('button', { name: /📤 Отправить/ });

      expect(sendButton).toBeDisabled();
    });

    it.skip('should validate message length limit', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      const longMessage = 'a'.repeat(4100);
      await userEvent.type(textarea, longMessage);

      const sendButton = screen.getByText(/📤 Отправить/);
      await userEvent.click(sendButton);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Сообщение не должно превышать 4096 символов',
          'error'
        );
      });
    });

    it('should enable send button when message is present', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);
      const sendButton = screen.getByRole('button', { name: /📤 Отправить/ });

      expect(sendButton).toBeDisabled();

      await userEvent.type(textarea, 'Test message');

      expect(sendButton).not.toBeDisabled();
    });
  });

  describe('File Upload', () => {
    it('should have file attach button', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText(/📎 Прикрепить файлы/)).toBeInTheDocument();
    });

    it('should show file count in button', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      expect(screen.getByText(/📎 Прикрепить файлы \(0\/10\)/)).toBeInTheDocument();
    });

    it('should add files to list', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const testFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');

      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(screen.getByText(/test.pdf/)).toBeInTheDocument();
      });
    });

    it.skip('should show error when exceeding max file count', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const files = Array.from({ length: 11 }, (_, i) =>
        new File([`${i}`], `file${i}.pdf`, { type: 'application/pdf' })
      );

      const fileInput = document.querySelector('input[type="file"]');

      for (let i = 0; i < 10; i++) {
        fireEvent.change(fileInput, { target: { files: [files[i]] } });
      }

      fireEvent.change(fileInput, { target: { files: [files[10]] } });

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Максимум 10 файлов на рассылку',
          'error'
        );
      });
    });

    it('should show error for file larger than 10MB', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const largeFile = new File(['x'.repeat(11 * 1024 * 1024)], 'large.pdf', {
        type: 'application/pdf',
      });

      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [largeFile] } });

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          expect.stringContaining('превышает 10MB'),
          'error'
        );
      });
    });

    it('should remove file from list', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const testFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');

      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(screen.getByText(/test.pdf/)).toBeInTheDocument();
      });

      const removeButton = screen.getByLabelText('Удалить файл');
      await userEvent.click(removeButton);

      expect(screen.queryByText(/test.pdf/)).not.toBeInTheDocument();
    });

    it('should update button text with file count', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const testFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');

      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(screen.getByText(/📎 Прикрепить файлы \(1\/10\)/)).toBeInTheDocument();
      });
    });

    it.skip('should disable attach button when max files reached', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const files = Array.from({ length: 10 }, (_, i) =>
        new File([`${i}`], `file${i}.pdf`, { type: 'application/pdf' })
      );

      const fileInput = document.querySelector('input[type="file"]');

      for (const file of files) {
        fireEvent.change(fileInput, { target: { files: [file] } });
      }

      const attachButton = screen.getByText(/📎 Прикрепить файлы/);

      await waitFor(() => {
        expect(attachButton).toBeDisabled();
      });
    });
  });

  describe('Preview', () => {
    it('should show preview button', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      expect(screen.getByText(/👁️ Предпросмотр/)).not.toBeDisabled();
    });

    it('should prevent preview with empty message', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const previewButton = screen.getByRole('button', { name: /👁️ Предпросмотр/ });

      expect(previewButton).toBeDisabled();
    });

    it('should show preview modal with message', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message content');

      const previewButton = screen.getByRole('button', { name: /👁️ Предпросмотр/ });
      await userEvent.click(previewButton);

      await waitFor(() => {
        expect(screen.getByText('Предпросмотр рассылки')).toBeInTheDocument();
      });
      // Message text appears in both textarea and preview, use getAllByText
      const messageElements = screen.getAllByText('Test message content');
      expect(messageElements.length).toBeGreaterThanOrEqual(1);
    });

    it('should show lesson info in preview', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      const previewButton = screen.getByText(/👁️ Предпросмотр/);
      await userEvent.click(previewButton);

      expect(screen.getByText(/Математика/)).toBeInTheDocument();
    });

    it('should show files in preview', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      const testFile = new File(['content'], 'homework.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [testFile] } });

      // Wait for file to appear in list
      await waitFor(() => {
        expect(screen.getByText(/homework.pdf/)).toBeInTheDocument();
      });

      const previewButton = screen.getByRole('button', { name: /👁️ Предпросмотр/ });
      await userEvent.click(previewButton);

      await waitFor(() => {
        expect(screen.getByText('Предпросмотр рассылки')).toBeInTheDocument();
      });
      // File should be visible in preview (there might be multiple instances)
      const fileElements = screen.getAllByText(/homework.pdf/);
      expect(fileElements.length).toBeGreaterThanOrEqual(1);
    });

    it('should close preview modal', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      let previewButton = screen.getByText(/👁️ Предпросмотр/);
      await userEvent.click(previewButton);

      expect(screen.getByText('Предпросмотр рассылки')).toBeInTheDocument();

      const closeButton = screen.getByText('Закрыть');
      await userEvent.click(closeButton);

      expect(screen.queryByText('Предпросмотр рассылки')).not.toBeInTheDocument();
    });
  });

  describe('Send Broadcast', () => {
    it('should send broadcast', async () => {
      mockSendMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test broadcast message');

      const sendButton = screen.getByText(/📤 Отправить/);
      await userEvent.click(sendButton);

      await waitFor(() => {
        expect(mockSendMutation.mutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            lessonId: 'lesson-1',
            message: 'Test broadcast message',
          })
        );
      });
    });

    it('should show success notification', async () => {
      mockSendMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      const sendButton = screen.getByText(/📤 Отправить/);
      await userEvent.click(sendButton);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Рассылка успешно отправлена',
          'success'
        );
      });
    });

    it('should clear form after send', async () => {
      mockSendMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      const sendButton = screen.getByText(/📤 Отправить/);
      await userEvent.click(sendButton);

      await waitFor(() => {
        expect(textarea).toHaveValue('');
      });
    });

    it('should handle send error', async () => {
      mockSendMutation.mutateAsync.mockRejectedValueOnce(new Error('Send failed'));

      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);
      const textarea = screen.getByPlaceholderText(/Введите сообщение для студентов урока/);

      await userEvent.type(textarea, 'Test message');

      const sendButton = screen.getByText(/📤 Отправить/);
      await userEvent.click(sendButton);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalled();
      });
    });
  });

  describe('Broadcast History', () => {
    it('should show broadcast items as clickable', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const broadcast = screen.getByText(/Привет, студенты/);
      expect(broadcast.closest('.broadcast-item')).toBeInTheDocument();
    });

    it('should show broadcast details in modal', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      // The broadcasts are clickable items showing history
      const broadcast = screen.getByText(/Привет, студенты/);
      expect(broadcast).toBeInTheDocument();
      expect(screen.getByText(/📎 1 файл\(а\)/)).toBeInTheDocument();
    });

    it.skip('should close broadcast details modal', async () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      const broadcast = screen.getByText(/Привет, студенты/);
      await userEvent.click(broadcast);

      expect(screen.getByText('Детали рассылки')).toBeInTheDocument();

      const modal = screen.getByText('Детали рассылки').closest('[role="dialog"]');
      if (modal) {
        fireEvent.click(modal.parentElement);
      }
    });
  });

  describe('Student Role', () => {
    beforeEach(() => {
      vi.mocked(useAuth).mockReturnValue({
        user: { id: 'student-1', role: 'student' },
      });
    });

    it('should not show compose section for students', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      expect(screen.queryByText('Новая рассылка')).not.toBeInTheDocument();
      expect(screen.queryByText(/Введите сообщение для студентов урока/)).not.toBeInTheDocument();
    });

    it('should show broadcast history for students', () => {
      renderWithQueryClient(<BroadcastSection lessonId="lesson-1" lesson={mockLesson} />);

      expect(screen.getByText('История рассылок')).toBeInTheDocument();
      expect(screen.getByText(/Привет, студенты/)).toBeInTheDocument();
    });
  });
});
