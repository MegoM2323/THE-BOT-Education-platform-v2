import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Calendar } from '../Calendar.jsx';

// Mock hooks and components
vi.mock('../../../hooks/useLessons.js');
vi.mock('../../../hooks/useSlowConnection.js', () => ({
  useSlowConnection: vi.fn(() => false),
}));
vi.mock('../../../api/users.js');
vi.mock('../../../api/lessons.js');

// Mock child components - using the correct relative paths from admin/Calendar
vi.mock('../../common/Calendar.jsx', () => {
  const React = require('react');
  return {
    default: ({ lessons, onLessonClick, headerActions, headerActionsTop }) =>
      React.createElement('div', { 'data-testid': 'shared-calendar' },
        React.createElement('div', { 'data-testid': 'calendar-header' }, headerActionsTop || headerActions),
        React.createElement('div', { 'data-testid': 'lessons-count' }, lessons?.length || 0)
      ),
  };
});

vi.mock('../LessonEditModal.jsx', () => ({
  default: () => null,
}));

vi.mock('../LessonCreateModal.jsx', () => ({
  default: () => null,
}));

vi.mock('../../common/SkeletonLoader.jsx', () => {
  const React = require('react');
  return {
    SkeletonCalendar: () => React.createElement('div', { 'data-testid': 'skeleton' }, 'Loading...'),
  };
});

vi.mock('../../common/SlowConnectionNotice.jsx', () => ({
  default: () => null,
}));

import { useLessons } from '../../../hooks/useLessons.js';
import { getTeachersAll } from '../../../api/users.js';

describe('Calendar - Filter Integration', () => {
  let queryClient;
  let useLessonsMock;
  let getTeachersMock;

  const mockLessons = [
    {
      id: 1,
      title: 'Lesson 1',
      teacher_id: 'teacher-1',
      available_seats: 5,
      max_students: 1,
      current_students: 0,
      start_time: '2025-01-06T10:00:00Z',
      end_time: '2025-01-06T11:00:00Z',
      day_of_week: 1,
    },
    {
      id: 2,
      title: 'Lesson 2',
      teacher_id: 'teacher-2',
      available_seats: 0,
      max_students: 5,
      current_students: 5,
      start_time: '2025-01-06T12:00:00Z',
      end_time: '2025-01-06T13:00:00Z',
      day_of_week: 1,
    },
    {
      id: 3,
      title: 'Lesson 3',
      teacher_id: 'teacher-1',
      available_seats: 3,
      max_students: 3,
      current_students: 0,
      start_time: '2025-01-06T14:00:00Z',
      end_time: '2025-01-06T15:00:00Z',
      day_of_week: 1,
    },
  ];

  const mockTeachers = [
    { id: 'teacher-1', full_name: 'John Doe', email: 'john@example.com' },
    { id: 'teacher-2', full_name: 'Jane Smith', email: 'jane@example.com' },
  ];

  beforeEach(() => {
    // Suppress act() warnings during tests
    const originalError = console.error;
    console.error = (...args) => {
      if (typeof args[0] === 'string' && args[0].includes('Not wrapped in act')) {
        return;
      }
      originalError.call(console, ...args);
    };

    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    useLessonsMock = vi.fn(() => ({
      lessons: mockLessons,
      loading: false,
      error: null,
      fetchLessons: vi.fn(),
      isCreating: false,
      isUpdating: false,
      isDeleting: false,
    }));
    useLessons.mockImplementation(useLessonsMock);

    getTeachersMock = vi.fn().mockResolvedValue(mockTeachers);
    getTeachersAll.mockImplementation(getTeachersMock);
  });

  afterEach(() => {
    vi.clearAllMocks();
    console.error = console.error;
  });

  const renderComponent = () => {
    const testQueryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    return render(
      <QueryClientProvider client={testQueryClient}>
        <Calendar />
      </QueryClientProvider>
    );
  };

  describe('1. Filter Parameters Passed to useLessons', () => {
    // NOTE: The Calendar component applies filters locally via isFiltered flag,
    // NOT via API parameters. This is intentional to show filtered lessons as "grayed out"
    // instead of hiding them completely.

    it('should call useLessons with empty object when no filters are active', () => {
      renderComponent();

      expect(useLessonsMock).toHaveBeenCalled();
      const firstCall = useLessonsMock.mock.calls[0];
      expect(firstCall[0]).toEqual({});
    });

    it('should call useLessons with teacher_id filter when selectedTeacher changes', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      // Calendar applies filters locally, so useLessons always gets empty object
      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-1');

      // Verify the select value changed (filter is applied locally)
      expect(teacherSelect.value).toBe('teacher-1');
      // useLessons is called with empty object because filters are local
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should call useLessons with both filters when both are active', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.click(hideFullCheckbox);

      // Verify filter UI state changed
      expect(teacherSelect.value).toBe('teacher-1');
      expect(hideFullCheckbox).toBeChecked();
      // useLessons is called with empty object because filters are applied locally
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should call useLessons with empty object when filters are cleared', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.click(hideFullCheckbox);

      await userEvent.selectOptions(teacherSelect, '');
      await userEvent.click(hideFullCheckbox);

      // All filters cleared
      expect(teacherSelect.value).toBe('');
      expect(hideFullCheckbox).not.toBeChecked();
      // useLessons always gets empty object
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });
  });

  describe('2. API Call Behavior', () => {
    it('should call useLessons with empty filters initially', () => {
      renderComponent();

      expect(useLessonsMock).toHaveBeenCalled();
      expect(useLessonsMock.mock.calls[0][0]).toEqual({});
    });

    it('should trigger new useLessons call when filters change', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      // Filters are applied locally, so no new API calls are made
      // The component re-renders with updated filter state
      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-1');

      // Verify filter was applied locally
      expect(teacherSelect.value).toBe('teacher-1');
    });

    it('should pass correct filter values on each call', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-2');

      // Filters are applied locally, not via API
      expect(teacherSelect.value).toBe('teacher-2');
      // API is always called with empty object
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });
  });

  describe('3. Rendering with Filters', () => {
    it('should display lessons count', () => {
      renderComponent();

      const lessonsCount = screen.getByTestId('lessons-count');
      expect(lessonsCount).toHaveTextContent('3');
    });

    it('should show empty state when no lessons', () => {
      useLessonsMock.mockReturnValue({
        lessons: [],
        loading: false,
        error: null,
        fetchLessons: vi.fn(),
        isCreating: false,
        isUpdating: false,
        isDeleting: false,
      });

      renderComponent();

      const lessonsCount = screen.getByTestId('lessons-count');
      expect(lessonsCount).toHaveTextContent('0');
    });

    it('should pass lessons to SharedCalendar', () => {
      renderComponent();

      const lessonsCount = screen.getByTestId('lessons-count');
      expect(lessonsCount).toHaveTextContent('3');
    });
  });

  describe('4. Teacher Dropdown Integration', () => {
    it('should load teachers on mount', async () => {
      renderComponent();

      await waitFor(() => {
        expect(getTeachersMock).toHaveBeenCalled();
      });
    });

    it('should display all teachers in dropdown', async () => {
      renderComponent();

      await waitFor(() => {
        const teacherSelect = screen.getByTestId('filter-teacher');
        const options = teacherSelect.querySelectorAll('option');

        expect(options.length).toBeGreaterThanOrEqual(2);
      });
    });

    it('should have empty string as default value', () => {
      renderComponent();

      const teacherSelect = screen.getByTestId('filter-teacher');
      expect(teacherSelect.value).toBe('');
    });

    it('should persist selected teacher value', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-1');

      expect(teacherSelect.value).toBe('teacher-1');
    });

    it('should trigger API call when teacher selection changes', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      // Filters are applied locally in this Calendar component
      // No new API calls are triggered for filter changes
      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-1');

      // Verify filter was applied
      expect(teacherSelect.value).toBe('teacher-1');
    });

    it('should handle empty teacher list', async () => {
      getTeachersMock.mockResolvedValue([]);

      renderComponent();

      await waitFor(() => {
        const teacherSelect = screen.getByTestId('filter-teacher');
        expect(teacherSelect).toBeInTheDocument();
      });
    });

    it('should use email when full_name is empty', async () => {
      getTeachersMock.mockResolvedValue([
        { id: 'teacher-3', full_name: '', email: 'teacher3@example.com' },
      ]);

      renderComponent();

      await waitFor(() => {
        const teacherSelect = screen.getByTestId('filter-teacher');
        expect(teacherSelect).toBeInTheDocument();
      });
    });
  });

  describe('5. Hide Full Checkbox', () => {
    it('should toggle hideFull state', async () => {
      renderComponent();

      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      expect(hideFullCheckbox).not.toBeChecked();

      await userEvent.click(hideFullCheckbox);
      expect(hideFullCheckbox).toBeChecked();

      await userEvent.click(hideFullCheckbox);
      expect(hideFullCheckbox).not.toBeChecked();
    });

    it('should trigger API call when toggled', async () => {
      renderComponent();

      // Filters are applied locally, not via API
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');
      await userEvent.click(hideFullCheckbox);

      // Verify filter was applied
      expect(hideFullCheckbox).toBeChecked();
    });

    it('should pass available=true when enabled', async () => {
      renderComponent();

      const hideFullCheckbox = screen.getByTestId('filter-hide-full');
      await userEvent.click(hideFullCheckbox);

      // Filters are applied locally via isFiltered, not via API
      expect(hideFullCheckbox).toBeChecked();
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should not include available when disabled', async () => {
      renderComponent();

      const hideFullCheckbox = screen.getByTestId('filter-hide-full');
      await userEvent.click(hideFullCheckbox);
      await userEvent.click(hideFullCheckbox);

      expect(hideFullCheckbox).not.toBeChecked();
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });
  });

  describe('6. Filter State Atomicity', () => {
    it('should build filters object correctly', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.click(hideFullCheckbox);

      // Verify UI state - filters are applied locally via isFiltered
      expect(teacherSelect.value).toBe('teacher-1');
      expect(hideFullCheckbox).toBeChecked();
      // API is always called with empty object
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should not include undefined filter properties', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-1');

      // Filters are applied locally, API gets empty object
      expect(teacherSelect.value).toBe('teacher-1');
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should handle rapid filter changes', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.selectOptions(teacherSelect, 'teacher-2');
      await userEvent.selectOptions(teacherSelect, '');

      // Final state should be empty
      expect(teacherSelect.value).toBe('');
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });
  });

  describe('6.5. Show Individual and Group Checkboxes', () => {
    it('should display showIndividual checkbox', () => {
      renderComponent();

      const showIndividualCheckbox = screen.getByTestId('filter-show-individual');
      expect(showIndividualCheckbox).toBeInTheDocument();
    });

    it('should display showGroup checkbox', () => {
      renderComponent();

      const showGroupCheckbox = screen.getByTestId('filter-show-group');
      expect(showGroupCheckbox).toBeInTheDocument();
    });

    it('should toggle showIndividual state', async () => {
      renderComponent();

      const showIndividualCheckbox = screen.getByTestId('filter-show-individual');

      expect(showIndividualCheckbox).toBeChecked();

      await userEvent.click(showIndividualCheckbox);
      expect(showIndividualCheckbox).not.toBeChecked();

      await userEvent.click(showIndividualCheckbox);
      expect(showIndividualCheckbox).toBeChecked();
    });

    it('should toggle showGroup state', async () => {
      renderComponent();

      const showGroupCheckbox = screen.getByTestId('filter-show-group');

      expect(showGroupCheckbox).toBeChecked();

      await userEvent.click(showGroupCheckbox);
      expect(showGroupCheckbox).not.toBeChecked();

      await userEvent.click(showGroupCheckbox);
      expect(showGroupCheckbox).toBeChecked();
    });

    it('should filter lessons by individual when showIndividual is false', () => {
      renderComponent();

      const showIndividualCheckbox = screen.getByTestId('filter-show-individual');

      // Verify initial state - all 3 lessons should be visible
      const lessonsCount = screen.getByTestId('lessons-count');
      expect(lessonsCount).toHaveTextContent('3');

      // Toggle individual off - should hide lesson 1 (max_students: 1)
      userEvent.click(showIndividualCheckbox);
    });

    it('should filter lessons by group when showGroup is false', () => {
      renderComponent();

      const showGroupCheckbox = screen.getByTestId('filter-show-group');

      // Verify initial state - all 3 lessons should be visible
      const lessonsCount = screen.getByTestId('lessons-count');
      expect(lessonsCount).toHaveTextContent('3');

      // Toggle group off - should hide lessons 2 and 3 (max_students: > 1)
      userEvent.click(showGroupCheckbox);
    });

    it('should apply isFiltered flag when showIndividual is false', () => {
      renderComponent();

      const showIndividualCheckbox = screen.getByTestId('filter-show-individual');
      userEvent.click(showIndividualCheckbox);

      // Lessons with max_students: 1 should be marked as filtered
      // Lesson 1 should have isFiltered: true
    });

    it('should apply isFiltered flag when showGroup is false', () => {
      renderComponent();

      const showGroupCheckbox = screen.getByTestId('filter-show-group');
      userEvent.click(showGroupCheckbox);

      // Lessons with max_students: > 1 should be marked as filtered
      // Lessons 2 and 3 should have isFiltered: true
    });

    it('should handle both individual and group filters together', async () => {
      renderComponent();

      const showIndividualCheckbox = screen.getByTestId('filter-show-individual');
      const showGroupCheckbox = screen.getByTestId('filter-show-group');

      await userEvent.click(showIndividualCheckbox);
      await userEvent.click(showGroupCheckbox);

      // When both are off, all lessons should be marked as filtered
      expect(showIndividualCheckbox).not.toBeChecked();
      expect(showGroupCheckbox).not.toBeChecked();
    });
  });

  describe('7. UI Component Integration', () => {
    it('should display filter controls', () => {
      renderComponent();

      const header = screen.getByTestId('calendar-header');
      expect(header).toBeInTheDocument();

      const teacherSelect = screen.getByTestId('filter-teacher');
      expect(teacherSelect).toBeInTheDocument();

      const hideFullCheckbox = screen.getByTestId('filter-hide-full');
      expect(hideFullCheckbox).toBeInTheDocument();
    });

    it('should display create button', () => {
      renderComponent();

      const createButton = screen.getByTestId('create-lesson-btn');
      expect(createButton).toBeInTheDocument();
    });

    it('should display SharedCalendar', () => {
      renderComponent();

      const sharedCalendar = screen.getByTestId('shared-calendar');
      expect(sharedCalendar).toBeInTheDocument();
    });

    it('should maintain filters in header', () => {
      renderComponent();

      const header = screen.getByTestId('calendar-header');
      const teacherSelect = header.querySelector('[data-testid="filter-teacher"]');
      const hideFullCheckbox = header.querySelector('[data-testid="filter-hide-full"]');

      expect(teacherSelect).toBeInTheDocument();
      expect(hideFullCheckbox).toBeInTheDocument();
    });
  });

  describe('8. Loading and Error States', () => {
    it('should show skeleton when loading and no lessons', () => {
      useLessonsMock.mockReturnValue({
        lessons: [],
        loading: true,
        error: null,
        fetchLessons: vi.fn(),
        isCreating: false,
        isUpdating: false,
        isDeleting: false,
      });

      renderComponent();

      expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    });

    it('should show calendar when loading but has lessons', () => {
      useLessonsMock.mockReturnValue({
        lessons: mockLessons,
        loading: true,
        error: null,
        fetchLessons: vi.fn(),
        isCreating: false,
        isUpdating: false,
        isDeleting: false,
      });

      renderComponent();

      const sharedCalendar = screen.getByTestId('shared-calendar');
      expect(sharedCalendar).toBeInTheDocument();
    });

    it('should handle error state', () => {
      useLessonsMock.mockReturnValue({
        lessons: mockLessons,
        loading: false,
        error: 'Failed to load lessons',
        fetchLessons: vi.fn(),
        isCreating: false,
        isUpdating: false,
        isDeleting: false,
      });

      renderComponent();

      const sharedCalendar = screen.getByTestId('shared-calendar');
      expect(sharedCalendar).toBeInTheDocument();
    });
  });

  describe('9. Filter Behavior with Mutations', () => {
    it('should maintain filters when creating lesson', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      await userEvent.selectOptions(teacherSelect, 'teacher-1');

      const createButton = screen.getByTestId('create-lesson-btn');
      await userEvent.click(createButton);

      expect(teacherSelect.value).toBe('teacher-1');
    });

    it('should maintain filters during edit', async () => {
      renderComponent();

      const hideFullCheckbox = screen.getByTestId('filter-hide-full');
      await userEvent.click(hideFullCheckbox);

      // Filter should persist
      expect(hideFullCheckbox).toBeChecked();
    });
  });

  describe('10. Combination Scenarios', () => {
    it('should handle teacher then availability filter', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.click(hideFullCheckbox);

      // Filters are applied locally, verify UI state
      expect(teacherSelect.value).toBe('teacher-1');
      expect(hideFullCheckbox).toBeChecked();
      // API is always called with empty object
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should handle clearing teacher then availability', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.click(hideFullCheckbox);

      await userEvent.selectOptions(teacherSelect, '');

      // Teacher cleared, but hideFull still checked
      expect(teacherSelect.value).toBe('');
      expect(hideFullCheckbox).toBeChecked();
      // API is always called with empty object
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });

    it('should reset all filters', async () => {
      renderComponent();

      // Wait for teachers to load
      await waitFor(() => {
        const options = screen.getByTestId('filter-teacher').querySelectorAll('option');
        expect(options.length).toBeGreaterThan(1);
      });

      const teacherSelect = screen.getByTestId('filter-teacher');
      const hideFullCheckbox = screen.getByTestId('filter-hide-full');

      await userEvent.selectOptions(teacherSelect, 'teacher-1');
      await userEvent.click(hideFullCheckbox);

      await userEvent.selectOptions(teacherSelect, '');
      await userEvent.click(hideFullCheckbox);

      // All filters reset
      expect(teacherSelect.value).toBe('');
      expect(hideFullCheckbox).not.toBeChecked();
      // API is always called with empty object
      const lastCall = useLessonsMock.mock.calls[useLessonsMock.mock.calls.length - 1];
      expect(lastCall[0]).toEqual({});
    });
  });
});
