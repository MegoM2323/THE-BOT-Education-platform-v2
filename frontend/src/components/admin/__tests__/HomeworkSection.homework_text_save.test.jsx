import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { HomeworkSection } from '../HomeworkSection';
import * as lessonAPI from '../../../api/lessons';
import { useAuth } from '../../../hooks/useAuth';
import { useNotification } from '../../../hooks/useNotification';
import { useHomework, useUploadHomework, useDeleteHomework } from '../../../hooks/useHomework';

vi.mock('../../../api/lessons');
vi.mock('../../../hooks/useAuth');
vi.mock('../../../hooks/useNotification');
vi.mock('../../../hooks/useHomework');

describe('HomeworkSection - Homework Text Save (T406)', () => {
  let queryClient;
  const mockLesson = {
    id: 'lesson-1',
    homework_text: 'Existing homework text',
    teacher_id: 'teacher-1',
  };

  const mockUser = {
    id: 'teacher-1',
    role: 'methodologist',
  };

  const mockNotification = {
    showNotification: vi.fn(),
  };

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    vi.clearAllMocks();

    vi.mocked(useAuth).mockReturnValue({ user: mockUser });
    vi.mocked(useNotification).mockReturnValue(mockNotification);
    vi.mocked(useHomework).mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
      isFetching: false,
    });
    vi.mocked(useUploadHomework).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isError: false,
      error: null,
    });
    vi.mocked(useDeleteHomework).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isError: false,
      error: null,
    });

    vi.mocked(lessonAPI.updateLesson).mockResolvedValue({ ...mockLesson });
  });

  const renderComponent = (props = {}) => {
    const defaultProps = {
      lessonId: 'lesson-1',
      lesson: mockLesson,
      onHomeworkCountChange: vi.fn(),
    };

    return render(
      <QueryClientProvider client={queryClient}>
        <HomeworkSection {...defaultProps} {...props} />
      </QueryClientProvider>
    );
  };

  it('should display homework text from lesson prop', () => {
    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    expect(textarea.value).toBe('Existing homework text');
  });

  it.skip('should save homework text with debounce when teacher changes it', async () => {
    vi.useFakeTimers();
    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');

    fireEvent.change(textarea, { target: { value: 'New homework text' } });
    expect(textarea.value).toBe('New homework text');

    expect(lessonAPI.updateLesson).not.toHaveBeenCalled();

    vi.advanceTimersByTime(500);

    await waitFor(() => {
      expect(lessonAPI.updateLesson).toHaveBeenCalledWith('lesson-1', {
        homework_text: 'New homework text',
      });
    });

    vi.useRealTimers();
  });

  it.skip('should show saving indicator while saving', async () => {
    vi.useFakeTimers();

    vi.mocked(lessonAPI.updateLesson).mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(() => resolve({ ...mockLesson }), 1000);
        })
    );

    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    fireEvent.change(textarea, { target: { value: 'New text' } });

    vi.advanceTimersByTime(500);

    expect(screen.getByText('Сохранение...')).toBeInTheDocument();

    vi.advanceTimersByTime(1000);

    await waitFor(() => {
      expect(screen.queryByText('Сохранение...')).not.toBeInTheDocument();
    });

    expect(screen.getByText(/Сохранено/)).toBeInTheDocument();

    vi.useRealTimers();
  });

  it.skip('should show error notification if save fails', async () => {
    vi.useFakeTimers();

    const errorMsg = 'Network error';
    vi.mocked(lessonAPI.updateLesson).mockRejectedValue(new Error(errorMsg));

    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    fireEvent.change(textarea, { target: { value: 'New text' } });

    vi.advanceTimersByTime(500);

    await waitFor(() => {
      expect(mockNotification.showNotification).toHaveBeenCalledWith(
        'Не удалось сохранить описание',
        'error'
      );
    });

    vi.useRealTimers();
  });

  it.skip('should cancel save if lesson changes during debounce', async () => {
    vi.useFakeTimers();
    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    fireEvent.change(textarea, { target: { value: 'Text for lesson 1' } });

    vi.advanceTimersByTime(300);

    renderComponent({
      lessonId: 'lesson-2',
      lesson: { ...mockLesson, id: 'lesson-2', homework_text: 'Lesson 2 text' },
    });

    vi.advanceTimersByTime(250);

    expect(lessonAPI.updateLesson).not.toHaveBeenCalled();

    vi.useRealTimers();
  });

  it.skip('should invalidate React Query cache after successful save', async () => {
    vi.useFakeTimers();

    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    fireEvent.change(textarea, { target: { value: 'New text' } });

    vi.advanceTimersByTime(500);

    await waitFor(() => {
      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: ['lessons'],
      });
      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: ['myLessons'],
      });
    });

    vi.useRealTimers();
  });

  it.skip('should handle empty homework text correctly', async () => {
    vi.useFakeTimers();
    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');

    fireEvent.change(textarea, { target: { value: '' } });
    expect(textarea.value).toBe('');

    vi.advanceTimersByTime(500);

    await waitFor(() => {
      expect(lessonAPI.updateLesson).toHaveBeenCalledWith('lesson-1', {
        homework_text: '',
      });
    });

    vi.useRealTimers();
  });

  it('should not save if teacher cannot edit lesson', async () => {
    vi.mocked(useAuth).mockReturnValue({
      user: { id: 'other-teacher', role: 'methodologist' },
    });

    renderComponent();

    const textarea = screen.queryByPlaceholderText('Введите текст домашнего задания...');

    expect(textarea).not.toBeInTheDocument();
  });

  it('should allow admin to view homework text editor', async () => {
    vi.mocked(useAuth).mockReturnValue({ user: { id: 'admin-1', role: 'admin' } });

    renderComponent();

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    expect(textarea).toBeInTheDocument();
  });

  it('should update local state when homework text is loaded from lesson prop', () => {
    renderComponent({
      lesson: { ...mockLesson, homework_text: 'Updated from parent' },
    });

    const textarea = screen.getByPlaceholderText('Введите текст домашнего задания...');
    expect(textarea.value).toBe('Updated from parent');
  });
});
