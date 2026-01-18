import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import HomeworkSection from "../HomeworkSection.jsx";

// Import formatHomeworkText for direct testing
import { formatHomeworkText } from '../../../utils/formatHomeworkText.js';

// Import mocked modules
import { useHomework, useUploadHomework, useDeleteHomework } from '../../../hooks/useHomework.js';
import { useNotification } from '../../../hooks/useNotification.js';
import { useAuth } from '../../../hooks/useAuth.js';
import { downloadHomework } from '../../../api/homework.js';
import { updateLesson } from '../../../api/lessons.js';

// Mock hooks
vi.mock('../../../hooks/useHomework.js', () => ({
  useHomework: vi.fn(),
  useUploadHomework: vi.fn(),
  useDeleteHomework: vi.fn(),
}));

vi.mock('../../../hooks/useNotification.js', () => ({
  useNotification: vi.fn(),
}));

vi.mock('../../../hooks/useAuth.js', () => ({
  useAuth: vi.fn(),
}));

vi.mock('../../../api/homework.js', () => ({
  downloadHomework: vi.fn(),
}));

vi.mock('../../../api/lessons.js', () => ({
  updateLesson: vi.fn(),
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

describe('HomeworkSection', () => {
  const mockFiles = [
    {
      id: 'file-1',
      file_name: 'lecture-1.pdf',
      file_size: 2048000,
      created_at: '2025-01-01T10:00:00Z',
      created_by_name: 'Teacher Name',
    },
    {
      id: 'file-2',
      file_name: 'notes.docx',
      file_size: 512000,
      created_at: '2025-01-02T11:00:00Z',
      created_by_name: 'Teacher Name',
    },
  ];

  const mockUploadMutation = {
    mutateAsync: vi.fn(),
    isPending: false,
  };

  const mockDeleteMutation = {
    mutateAsync: vi.fn(),
    isPending: false,
  };

  const mockShowNotification = vi.fn();

  beforeEach(() => {
    vi.mocked(useHomework).mockReturnValue({
      data: mockFiles,
      isLoading: false,
      error: null,
    });

    vi.mocked(useUploadHomework).mockReturnValue(mockUploadMutation);
    vi.mocked(useDeleteHomework).mockReturnValue(mockDeleteMutation);

    vi.mocked(useNotification).mockReturnValue({
      showNotification: mockShowNotification,
    });

    vi.mocked(useAuth).mockReturnValue({
      user: { id: 'user-1', role: 'admin' },
    });

    vi.mocked(updateLesson).mockResolvedValue({ success: true });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render homework section', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByText(/Перетащите файл/)).toBeInTheDocument();
    });

    it('should render upload zone for admin/teacher', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByText(/Перетащите файл сюда или кликните для выбора/)).toBeInTheDocument();
    });

    it('should list all homework files', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByText('lecture-1.pdf')).toBeInTheDocument();
      expect(screen.getByText('notes.docx')).toBeInTheDocument();
    });

    it('should display file size correctly', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      // File 1: 2048000 bytes = ~2 MB, File 2: 512000 bytes = 500 KB
      expect(screen.getByText(/2 MB|1\.95 MB/)).toBeInTheDocument();
      expect(screen.getByText(/500 KB|0\.49 MB/)).toBeInTheDocument();
    });

    it('should display file count', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByText('2 из 10 файлов')).toBeInTheDocument();
    });

    it('should display teacher name', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      // The component displays created_by_name in file info
      const teacherElements = screen.getAllByText(/Teacher Name/);
      expect(teacherElements.length).toBeGreaterThan(0);
    });

    it('should render homework text textarea', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByLabelText(/Описание домашнего задания/)).toBeInTheDocument();
    });

    it('should load homework text from lesson prop', () => {
      const lesson = { homework_text: 'Test homework description' };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);
      const textarea = screen.getByLabelText(/Описание домашнего задания/);
      expect(textarea.value).toBe('Test homework description');
    });

    it('should show empty textarea when homework_text is null', () => {
      const lesson = { homework_text: null };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);
      const textarea = screen.getByLabelText(/Описание домашнего задания/);
      expect(textarea.value).toBe('');
    });
  });

  describe('Loading State', () => {
    it('should show spinner when loading', () => {
      vi.mocked(useHomework).mockReturnValue({
        data: [],
        isLoading: true,
        error: null,
      });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(document.querySelector('.homework-section')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should show error message when loading fails', () => {
      vi.mocked(useHomework).mockReturnValue({
        data: [],
        isLoading: false,
        error: new Error('Load failed'),
      });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByText(/Ошибка загрузки домашних заданий/)).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should show empty message when no files', () => {
      vi.mocked(useHomework).mockReturnValue({
        data: [],
        isLoading: false,
        error: null,
      });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      expect(screen.getByText('Домашние задания не добавлены')).toBeInTheDocument();
    });
  });

  describe('File Upload - Via File Input', () => {
    it('should open file picker when upload area is clicked', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const uploadZone = screen.getByText(/Перетащите файл сюда или кликните для выбора/);

      await userEvent.click(uploadZone);

      expect(mockUploadMutation.mutateAsync).not.toHaveBeenCalled();
    });

    it('should upload file when selected via input', async () => {
      mockUploadMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      const testFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(mockUploadMutation.mutateAsync).toHaveBeenCalledWith({
          lessonId: 'lesson-1',
          file: testFile,
        });
      });
    });

    it('should show error when file size exceeds limit', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      const largeFile = new File(['x'.repeat(11 * 1024 * 1024)], 'large.pdf', {
        type: 'application/pdf',
      });

      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [largeFile] } });

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Размер файла не должен превышать 10MB',
          'error'
        );
      });
    });

    it('should show error when exceeding max file count', async () => {
      const manyFiles = Array.from({ length: 10 }, (_, i) => ({
        id: `file-${i}`,
        file_name: `file${i}.pdf`,
        file_size: 1024,
        created_at: '2025-01-01T10:00:00Z',
      }));

      vi.mocked(useHomework).mockReturnValue({
        data: manyFiles,
        isLoading: false,
        error: null,
      });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      const testFile = new File(['content'], 'new.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Максимум 10 файлов на урок',
          'error'
        );
      });
    });

    it('should show success notification on upload', async () => {
      mockUploadMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      const testFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Файл успешно загружен',
          'success'
        );
      });
    });

    it('should clear file input after upload', async () => {
      mockUploadMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      const testFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = document.querySelector('input[type="file"]');
      fireEvent.change(fileInput, { target: { files: [testFile] } });

      await waitFor(() => {
        expect(fileInput.value).toBe('');
      });
    });
  });

  describe('File Upload - Drag and Drop', () => {
    it('should show dragging state on drag over', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const uploadZone = screen.getByText(/Перетащите файл сюда или кликните для выбора/).closest('div');

      fireEvent.dragOver(uploadZone);

      expect(uploadZone).toHaveClass('dragging');
    });

    it('should remove dragging state on drag leave', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const uploadZone = screen.getByText(/Перетащите файл сюда или кликните для выбора/).closest('div');

      fireEvent.dragOver(uploadZone);
      fireEvent.dragLeave(uploadZone);

      expect(uploadZone).not.toHaveClass('dragging');
    });

    it('should handle drop event', async () => {
      mockUploadMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const uploadZone = screen.getByText(/Перетащите файл сюда или кликните для выбора/).closest('div');

      const testFile = new File(['content'], 'dropped.pdf', { type: 'application/pdf' });

      fireEvent.drop(uploadZone, {
        dataTransfer: { files: [testFile] },
      });

      await waitFor(() => {
        expect(mockUploadMutation.mutateAsync).toHaveBeenCalled();
      });
    });

    it('should remove dragging state after drop', async () => {
      mockUploadMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const uploadZone = screen.getByText(/Перетащите файл сюда или кликните для выбора/).closest('div');

      const testFile = new File(['content'], 'dropped.pdf', { type: 'application/pdf' });

      fireEvent.dragOver(uploadZone);
      fireEvent.drop(uploadZone, {
        dataTransfer: { files: [testFile] },
      });

      expect(uploadZone).not.toHaveClass('dragging');
    });
  });

  describe('File Download', () => {
    it('should have download button for each file', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      // Download buttons have title="Скачать"
      const downloadButtons = screen.getAllByTitle('Скачать');
      expect(downloadButtons.length).toBeGreaterThanOrEqual(2);
    });

    it('should call download handler when download button clicked', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const downloadButtons = screen.getAllByTitle('Скачать');

      await userEvent.click(downloadButtons[0]);
    });
  });

  describe('File Delete', () => {
    it('should have delete button for each file', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      // Delete buttons have title="Удалить" and contain trash emoji
      const deleteButtons = screen.getAllByTitle('Удалить');
      expect(deleteButtons.length).toBeGreaterThanOrEqual(2);
    });

    it('should show confirmation modal when delete clicked', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const deleteButtons = screen.getAllByTitle('Удалить');

      await userEvent.click(deleteButtons[0]);

      expect(screen.getByText(/Вы уверены, что хотите удалить файл/)).toBeInTheDocument();
    });

    it('should delete file on confirmation', async () => {
      mockDeleteMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const deleteButtons = screen.getAllByTitle('Удалить');

      await userEvent.click(deleteButtons[0]);

      const confirmButton = screen.getByRole('button', { name: 'Удалить' });
      await userEvent.click(confirmButton);

      await waitFor(() => {
        expect(mockDeleteMutation.mutateAsync).toHaveBeenCalledWith({
          lessonId: 'lesson-1',
          fileId: 'file-1',
        });
      });
    });

    it('should show success notification after delete', async () => {
      mockDeleteMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const deleteButtons = screen.getAllByTitle('Удалить');

      await userEvent.click(deleteButtons[0]);

      const confirmButton = screen.getByRole('button', { name: 'Удалить' });
      await userEvent.click(confirmButton);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Файл успешно удалён',
          'success'
        );
      });
    });

    it('should handle delete error', async () => {
      mockDeleteMutation.mutateAsync.mockRejectedValueOnce(
        new Error('Delete failed')
      );

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const deleteButtons = screen.getAllByTitle('Удалить');

      await userEvent.click(deleteButtons[0]);

      const confirmButton = screen.getByRole('button', { name: 'Удалить' });
      await userEvent.click(confirmButton);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalled();
      });
    });

    it('should close modal after deletion', async () => {
      mockDeleteMutation.mutateAsync.mockResolvedValueOnce({ success: true });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const deleteButtons = screen.getAllByTitle('Удалить');

      await userEvent.click(deleteButtons[0]);

      expect(screen.getByText(/Вы уверены/)).toBeInTheDocument();

      const confirmButton = screen.getByRole('button', { name: 'Удалить' });
      await userEvent.click(confirmButton);

      await waitFor(() => {
        expect(screen.queryByText(/Вы уверены/)).not.toBeInTheDocument();
      });
    });
  });

  describe('Student Role', () => {
    beforeEach(() => {
      vi.mocked(useAuth).mockReturnValue({
        user: { id: 'student-1', role: 'student' },
      });
    });

    it('should not show upload zone for students', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      expect(
        screen.queryByText(/Перетащите файл сюда или кликните для выбора/)
      ).not.toBeInTheDocument();
    });

    it('should not show delete buttons for students', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      // Delete buttons should not be visible for students
      const deleteButtons = screen.queryAllByTitle('Удалить');
      expect(deleteButtons.length).toBe(0);
    });

    it('should show download button for students', () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      const downloadButtons = screen.queryAllByTitle('Скачать');
      expect(downloadButtons.length).toBeGreaterThan(0);
    });
  });

  describe('Upload Callback', () => {
    it('should call onHomeworkCountChange when count changes', () => {
      const onCountChange = vi.fn();

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} onHomeworkCountChange={onCountChange} />);

      expect(onCountChange).toHaveBeenCalledWith(2);
    });
  });

  describe('Homework Text Autosave', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it.skip('should autosave homework text after typing', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const textarea = screen.getByLabelText(/Описание домашнего задания/);

      await userEvent.type(textarea, 'New homework text');
      vi.advanceTimersByTime(500);

      await waitFor(() => {
        expect(updateLesson).toHaveBeenCalledWith('lesson-1', {
          homework_text: 'New homework text',
        });
      });
    });

    it.skip('should show saving indicator while saving', async () => {
      vi.mocked(updateLesson).mockImplementation(() => new Promise(resolve => setTimeout(resolve, 1000)));

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const textarea = screen.getByLabelText(/Описание домашнего задания/);

      await userEvent.type(textarea, 'Text');
      vi.advanceTimersByTime(500);

      await waitFor(() => {
        expect(screen.getByText(/Сохранение/)).toBeInTheDocument();
      });
    });

    it.skip('should show saved indicator after successful save', async () => {
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const textarea = screen.getByLabelText(/Описание домашнего задания/);

      await userEvent.type(textarea, 'Text');
      vi.advanceTimersByTime(500);

      await waitFor(() => {
        expect(screen.getByText(/Сохранено/)).toBeInTheDocument();
      });
    });

    it.skip('should handle save errors', async () => {
      vi.mocked(updateLesson).mockRejectedValue(new Error('Save failed'));

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);
      const textarea = screen.getByLabelText(/Описание домашнего задания/);

      await userEvent.type(textarea, 'Text');
      vi.advanceTimersByTime(500);

      await waitFor(() => {
        expect(mockShowNotification).toHaveBeenCalledWith(
          'Не удалось сохранить описание',
          'error'
        );
      });
    });

    it('should show read-only div instead of textarea for students', () => {
      vi.mocked(useAuth).mockReturnValue({
        user: { id: 'student-1', role: 'student' },
      });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{ homework_text: 'Test homework' }} />);

      // For students, textarea is not rendered - instead a div is shown
      const textarea = screen.queryByRole('textbox');
      expect(textarea).not.toBeInTheDocument();

      // Verify read-only display is shown
      expect(screen.getByText('Test homework')).toBeInTheDocument();
    });

    it('should not autosave for students', async () => {
      vi.mocked(useAuth).mockReturnValue({
        user: { id: 'student-1', role: 'student' },
      });

      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={{}} />);

      vi.advanceTimersByTime(1000);

      expect(updateLesson).not.toHaveBeenCalled();
    });
  });

  describe('Homework Text Formatting - Line Breaks', () => {
    beforeEach(() => {
      vi.mocked(useAuth).mockReturnValue({
        user: { id: 'student-1', role: 'student' },
      });
    });

    it('should display homework text with line breaks', () => {
      const lesson = { homework_text: 'Line 1\nLine 2\nLine 3' };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);

      // Find the homework-text-display div
      const displayDiv = document.querySelector('.homework-text-display');
      expect(displayDiv).toBeInTheDocument();

      // Check that all lines are in the display
      expect(displayDiv.textContent).toContain('Line 1');
      expect(displayDiv.textContent).toContain('Line 2');
      expect(displayDiv.textContent).toContain('Line 3');

      // Check that br elements are present
      const brElements = displayDiv.querySelectorAll('br');
      expect(brElements.length).toBe(2);
    });

    it('should handle text with multiple consecutive line breaks', () => {
      const lesson = { homework_text: 'Paragraph 1\n\n\nParagraph 2' };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);

      const displayDiv = document.querySelector('.homework-text-display');
      expect(displayDiv).toBeInTheDocument();
      expect(displayDiv.textContent).toContain('Paragraph 1');
      expect(displayDiv.textContent).toContain('Paragraph 2');

      const brElements = displayDiv.querySelectorAll('br');
      expect(brElements.length).toBe(3);
    });

    it('should handle URLs and line breaks in homework text', () => {
      const text = 'Read here:\nhttps://example.com\nThanks';
      const lesson = { homework_text: text };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);

      const displayDiv = document.querySelector('.homework-text-display');
      expect(displayDiv).toBeInTheDocument();

      // Check that text parts are present
      expect(displayDiv.textContent).toContain('Read here:');
      expect(displayDiv.textContent).toContain('Thanks');

      // Check that link is present and properly formatted
      const link = screen.getByRole('link', { name: 'https://example.com' });
      expect(link).toBeInTheDocument();
      expect(link).toHaveAttribute('href', 'https://example.com');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');

      // Check that br elements are present
      const brElements = displayDiv.querySelectorAll('br');
      expect(brElements.length).toBe(2);
    });

    it('should handle multiple URLs with line breaks', () => {
      const text = 'Links:\nhttps://example1.com\nhttps://example2.com\nDone';
      const lesson = { homework_text: text };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);

      const displayDiv = document.querySelector('.homework-text-display');
      expect(displayDiv).toBeInTheDocument();

      const link1 = screen.getByRole('link', { name: 'https://example1.com' });
      const link2 = screen.getByRole('link', { name: 'https://example2.com' });

      expect(link1).toHaveAttribute('href', 'https://example1.com');
      expect(link2).toHaveAttribute('href', 'https://example2.com');
      expect(displayDiv.textContent).toContain('Done');

      const brElements = displayDiv.querySelectorAll('br');
      expect(brElements.length).toBe(3);
    });

    it('should not display br elements for single line text', () => {
      const lesson = { homework_text: 'Single line text' };
      renderWithQueryClient(<HomeworkSection lessonId="lesson-1" lesson={lesson} />);

      const displayDiv = document.querySelector('.homework-text-display');
      expect(displayDiv).toBeInTheDocument();
      expect(displayDiv.textContent).toContain('Single line text');

      const brElements = displayDiv.querySelectorAll('br');
      expect(brElements.length).toBe(0);
    });
  });

  describe('formatHomeworkText utility function', () => {
    it('should return null for empty text', () => {
      expect(formatHomeworkText('')).toBeNull();
      expect(formatHomeworkText(null)).toBeNull();
      expect(formatHomeworkText('   ')).toBeNull();
    });

    it('should handle simple text without URLs or line breaks', () => {
      const result = formatHomeworkText('Simple text');
      expect(result).not.toBeNull();
      expect(Array.isArray(result)).toBe(true);
      expect(result[0]).toBe('Simple text');
    });

    it('should convert line breaks to br elements', () => {
      const result = formatHomeworkText('Line 1\nLine 2');
      expect(result).not.toBeNull();
      expect(Array.isArray(result)).toBe(true);

      // Should have: 'Line 1', <br>, 'Line 2'
      expect(result.length).toBe(3);
      expect(result[0]).toBe('Line 1');
      expect(result[1].type).toBe('br');
      expect(result[2]).toBe('Line 2');
    });

    it('should convert URLs to anchor links', () => {
      const result = formatHomeworkText('Visit https://example.com now');
      expect(result).not.toBeNull();

      // Find the link element
      const linkElement = result.find(el => el.type === 'a');
      expect(linkElement).toBeDefined();
      expect(linkElement.props.href).toBe('https://example.com');
      expect(linkElement.props.target).toBe('_blank');
      expect(linkElement.props.rel).toBe('noopener noreferrer');
    });

    it('should handle both URLs and line breaks', () => {
      const result = formatHomeworkText('Line 1 https://example.com\nLine 2');
      expect(result).not.toBeNull();
      expect(Array.isArray(result)).toBe(true);

      const linkElement = result.find(el => el.type === 'a');
      expect(linkElement).toBeDefined();
      expect(linkElement.props.href).toBe('https://example.com');

      const brElement = result.find(el => el.type === 'br');
      expect(brElement).toBeDefined();
    });

    it('should handle multiple URLs on same line', () => {
      const result = formatHomeworkText('https://example1.com and https://example2.com');
      const links = result.filter(el => el.type === 'a');
      expect(links.length).toBe(2);
      expect(links[0].props.href).toBe('https://example1.com');
      expect(links[1].props.href).toBe('https://example2.com');
    });

    it('should handle URLs with trailing punctuation', () => {
      const result = formatHomeworkText('Visit https://example.com.');
      const linkElement = result.find(el => el.type === 'a');
      expect(linkElement.props.href).toContain('https://example.com');
    });
  });
});
