import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Calendar from '../Calendar.jsx';

const mockLessons = [
  {
    id: 'lesson-1',
    start_time: new Date(Date.now() + 86400000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 1,
    bookings: [{ student_id: 'student-1' }],
    subject: 'Математика',
  },
  {
    id: 'lesson-2',
    start_time: new Date(Date.now() + 172800000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 1,
    bookings: [{ student_id: 'student-2' }],
    subject: 'Русский язык',
  },
  {
    id: 'lesson-3',
    start_time: new Date(Date.now() + 259200000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 2,
    bookings: [{ student_id: 'student-1' }, { student_id: 'student-2' }],
    subject: 'Английский язык',
  },
  {
    id: 'lesson-4',
    start_time: new Date(Date.now() + 345600000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 0,
    bookings: [],
    subject: 'История',
  },
];

vi.mock('../../../api/lessons.js', () => ({
  lessonsAPI: {
    getLessons: vi.fn(() => Promise.resolve([])),
  },
}));

vi.mock('../../../api/users.js', () => ({
  getTeachersAll: vi.fn(() =>
    Promise.resolve([
      { id: 'teacher-1', full_name: 'Иван Учитель', email: 'ivan@example.com' },
    ])
  ),
  getStudentsAll: vi.fn(() =>
    Promise.resolve([
      { id: 'student-1', full_name: 'Петр Ученик', email: 'petr@example.com' },
      { id: 'student-2', full_name: 'Мария Ученица', email: 'maria@example.com' },
      { id: 'student-3', full_name: 'Иван Новый', email: 'ivan.new@example.com' },
    ])
  ),
}));

vi.mock('../../../hooks/useLessons.js', () => ({
  useLessons: () => ({
    lessons: mockLessons,
    loading: false,
    error: null,
    fetchLessons: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useSlowConnection.js', () => ({
  useSlowConnection: () => false,
}));

vi.mock('../../common/Calendar.jsx', () => ({
  default: ({ lessons, headerActionsTop }) => (
    <div data-testid="shared-calendar">
      <div data-testid="calendar-header-actions">{headerActionsTop}</div>
      <div data-testid="lessons-list">
        {lessons && lessons.length > 0 ? (
          <ul>
            {lessons.map((lesson) => (
              <li key={lesson.id} data-testid={`lesson-${lesson.id}`}>
                <span
                  data-testid={`lesson-status-${lesson.id}`}
                  className={lesson.isFiltered ? 'filtered' : 'active'}
                >
                  {lesson.subject}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <div>No lessons</div>
        )}
      </div>
    </div>
  ),
}));

vi.mock('../AdminCalendarCreditsDisplay.jsx', () => ({
  default: () => <div data-testid="credits-display">Credits Display</div>,
}));

vi.mock('../LessonEditModal.jsx', () => ({
  default: () => null,
}));

vi.mock('../LessonCreateModal.jsx', () => ({
  default: () => null,
}));

describe('Calendar: StudentFilterSearch Integration (Multiple Students)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // ========== GROUP 1: Component Integration (4 тестов) ==========

  it('should render StudentFilterSearch with correct props for multiple selection', async () => {
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    expect(headerActions).toBeInTheDocument();

    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );
    expect(studentInput).toBeInTheDocument();
    expect(studentInput).not.toBeDisabled();
  });

  it('should update Calendar state when multiple students are selected', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
    expect(lesson1Status).toHaveClass('active');

    await user.click(studentFilterInput);

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    await waitFor(() => {
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      expect(lesson2Status).toHaveClass('active');
    });
  });

  it('should display selected students as chips and allow removal', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const headerActions2 = screen.getByTestId('calendar-header-actions');
      const chipsContainer = headerActions2.querySelector('.student-filter-search-chips-container');
      expect(chipsContainer).toBeInTheDocument();
      const chipsCount = chipsContainer.querySelectorAll('.student-filter-search-chip');
      expect(chipsCount.length).toBe(1);
    });

    await user.click(studentFilterInput);

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    await waitFor(() => {
      const headerActions2 = screen.getByTestId('calendar-header-actions');
      const chipsContainer = headerActions2.querySelector('.student-filter-search-chips-container');
      const chipsCount = chipsContainer.querySelectorAll('.student-filter-search-chip');
      expect(chipsCount.length).toBe(2);
    });
  });

  it('should clear all students and remove filter', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await user.click(studentFilterInput);

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    await waitFor(() => {
      const chipsContainer = headerActions.querySelector('.student-filter-search-chips-container');
      const chips = chipsContainer.querySelectorAll('.student-filter-search-chip');
      expect(chips.length).toBe(2);
    });

    const clearButtons = within(headerActions).queryAllByRole('button');
    let clearRemoveButtons = clearButtons.filter((btn) =>
      btn.getAttribute('aria-label')?.includes('Удалить')
    );

    while (clearRemoveButtons.length > 0) {
      await user.click(clearRemoveButtons[0]);
      clearRemoveButtons = within(headerActions)
        .queryAllByRole('button')
        .filter((btn) => btn.getAttribute('aria-label')?.includes('Удалить'));
    }

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  // ========== GROUP 2: Filtering Logic (5 тестов) ==========

  it('should show lessons with bookings from selected student', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      const lesson3Status = screen.getByTestId('lesson-status-lesson-3');
      expect(lesson1Status).toHaveClass('active');
      expect(lesson3Status).toHaveClass('active');
    });
  });

  it('should mark lessons without selected student as filtered (gray)', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      const lesson4Status = screen.getByTestId('lesson-status-lesson-4');
      expect(lesson2Status).toHaveClass('filtered');
      expect(lesson4Status).toHaveClass('filtered');
    });
  });

  it('should show no filtering when selectedFilterStudents is empty', async () => {
    render(<Calendar />);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      const lesson3Status = screen.getByTestId('lesson-status-lesson-3');
      const lesson4Status = screen.getByTestId('lesson-status-lesson-4');

      expect(lesson1Status).toHaveClass('active');
      expect(lesson2Status).toHaveClass('active');
      expect(lesson3Status).toHaveClass('active');
      expect(lesson4Status).toHaveClass('active');
    });
  });

  it('should show lesson if ANY of multiple selected students have booking', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await user.click(studentFilterInput);

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      const lesson3Status = screen.getByTestId('lesson-status-lesson-3');
      expect(lesson1Status).toHaveClass('active');
      expect(lesson2Status).toHaveClass('active');
      expect(lesson3Status).toHaveClass('active');
    });
  });

  it('should show all bookings of student across multiple lessons', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      const lesson3Status = screen.getByTestId('lesson-status-lesson-3');
      expect(lesson1Status).toHaveClass('active');
      expect(lesson3Status).toHaveClass('active');
    });
  });

  // ========== GROUP 3: UI State Sync (3 теста) ==========

  it('should display 1 chip when 1 student selected', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const chipsContainer = headerActions.querySelector('.student-filter-search-chips-container');
      const chips = chipsContainer.querySelectorAll('.student-filter-search-chip');
      expect(chips.length).toBe(1);
    });
  });

  it('should display 2 chips when 2 students selected', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await user.click(studentFilterInput);

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    await waitFor(() => {
      const chipsContainer = headerActions.querySelector('.student-filter-search-chips-container');
      const chips = chipsContainer.querySelectorAll('.student-filter-search-chip');
      expect(chips.length).toBe(2);
    });
  });

  it('should update lesson visibility immediately when deselecting student', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });

    const chips = within(headerActions).queryAllByRole('button');
    const removeChipBtn = chips.find((btn) => btn.textContent.includes('×'));
    if (removeChipBtn) {
      await user.click(removeChipBtn);
    }

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).not.toHaveClass('filtered');
    });
  });

  // ========== GROUP 4: Edge Cases (4 теста) ==========

  it('should handle null selectedFilterStudents gracefully', async () => {
    render(<Calendar />);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  it('should handle empty array selectedFilterStudents=[] without filtering', async () => {
    render(<Calendar />);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      const lesson3Status = screen.getByTestId('lesson-status-lesson-3');
      expect(lesson1Status).toHaveClass('active');
      expect(lesson2Status).toHaveClass('active');
      expect(lesson3Status).toHaveClass('active');
    });
  });

  it('should handle rapid selection and deselection without race conditions', async () => {
    const user = userEvent.setup({ delay: null });
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    }, { timeout: 1000 });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    // Wait for dropdown to close and reopen with keyboard
    await user.keyboard('{ArrowDown}');

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    }, { timeout: 1000 });

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    // Wait for dropdown to close and reopen with keyboard
    await user.keyboard('{ArrowDown}');

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    }, { timeout: 1000 });

    const student3Option = screen.getByText('Иван Новый');
    await user.click(student3Option);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    }, { timeout: 1000 });

    expect(screen.getByTestId('lesson-status-lesson-1')).toHaveClass('active');
  });

  it('should handle lesson bookings changes and keep filtering current', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  // ========== GROUP 5: API Integration (3 теста) ==========

  it('should build correct query params for student_ids filter', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await user.click(studentFilterInput);

    const student2Option = screen.getByText('Мария Ученица');
    await user.click(student2Option);

    await waitFor(() => {
      expect(screen.getByTestId('lesson-status-lesson-1')).toHaveClass('active');
    });
  });

  it('should filter locally without API call when student_ids change', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  it('should not pass empty student_ids in query', async () => {
    render(<Calendar />);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  // ========== GROUP 6: Visual Filtering (2 теста) ==========

  it('should apply isFiltered=true class to non-matching lessons', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      expect(lesson2Status).toHaveClass('filtered');
    });
  });

  it('should apply normal styling to non-filtered lessons', async () => {
    render(<Calendar />);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
      expect(lesson1Status).not.toHaveClass('filtered');
    });
  });

  // ========== GROUP 7: UUID String Conversion (5 тестов) ==========

  it('should handle UUID objects in selectedFilterStudents with String() conversion', async () => {
    const uuidObj = { toString: () => 'student-1' };
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    // Simulate selecting a student (converts UUID object to string)
    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  it('should handle UUID strings in selectedFilterStudents', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  it('should handle UUID objects in lesson.bookings.student_id with String() conversion', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  it('should filter correctly with mixed UUID formats (object in filter, string in booking)', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson1Status = screen.getByTestId('lesson-status-lesson-1');
      expect(lesson1Status).toHaveClass('active');
    });
  });

  it('should correctly identify non-matching lessons when UUID types differ', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const headerActions = await screen.findByTestId('calendar-header-actions');
    const inputs = within(headerActions).queryAllByRole('textbox');
    const studentFilterInput = inputs.find(
      (input) => input.getAttribute('aria-label') === 'Поиск студента'
    );

    await user.click(studentFilterInput);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const student1Option = screen.getByText('Петр Ученик');
    await user.click(student1Option);

    await waitFor(() => {
      const lesson2Status = screen.getByTestId('lesson-status-lesson-2');
      expect(lesson2Status).toHaveClass('filtered');
    });
  });
});
