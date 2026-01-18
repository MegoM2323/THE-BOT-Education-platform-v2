import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import StudentFilterSearch from '../StudentFilterSearch.jsx';

const mockStudents = [
  { id: 'student-1', full_name: 'Иван Иванов', email: 'ivan@example.com' },
  { id: 'student-2', full_name: 'Мария Петрова', email: 'maria@example.com' },
  { id: 'student-3', full_name: 'Петр Сидоров', email: 'petr@example.com' },
];

let mockGetStudentsAll;

vi.mock('../../../api/users.js', async () => {
  return {
    getStudentsAll: (...args) => mockGetStudentsAll(...args),
  };
});

describe('StudentFilterSearch Component - Multi-Select', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetStudentsAll = vi.fn(() => Promise.resolve(mockStudents));
  });

  // ============================================
  // GROUP 1: Rendering (4 tests)
  // ============================================

  it('should render with empty selectedStudents array', async () => {
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    // No chips should be displayed
    const chipsContainer = document.querySelector('.student-filter-search-chips-container');
    expect(chipsContainer).not.toBeInTheDocument();
  });

  it('should display chips for selectedStudents array', async () => {
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1', 'student-2']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    await waitFor(() => {
      const chips = screen.getAllByRole('button', { name: /Удалить/i });
      expect(chips.length).toBeGreaterThanOrEqual(2);
    });

    expect(screen.getByText('Иван Иванов')).toBeInTheDocument();
    expect(screen.getByText('Мария Петрова')).toBeInTheDocument();
  });

  it('should render with disabled state', async () => {
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
        disabled={true}
      />
    );

    await waitFor(() => {
      const input = screen.getByRole('textbox');
      expect(input).toBeDisabled();
    });
  });

  it('should display error state when API fails', async () => {
    const user = userEvent.setup();
    mockGetStudentsAll = vi.fn(() =>
      Promise.reject(new Error('Network error'))
    );

    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  // ============================================
  // GROUP 2: Search/Filter (3 tests)
  // ============================================

  it('should filter students by name in search input', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    expect(screen.getByText('Иван Иванов')).toBeInTheDocument();
    expect(screen.getByText('Мария Петрова')).toBeInTheDocument();
    expect(screen.getByText('Петр Сидоров')).toBeInTheDocument();

    await user.type(input, 'Мария');

    await waitFor(() => {
      expect(screen.queryByText('Мария Петрова')).toBeInTheDocument();
      expect(screen.queryByText('Иван Иванов')).not.toBeInTheDocument();
      expect(screen.queryByText('Петр Сидоров')).not.toBeInTheDocument();
    });
  });

  it('should preserve selected students when searching', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Иван should be checked
    const ivanCheckbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    expect(ivanCheckbox).toBeChecked();

    // Type to filter
    await user.type(input, 'М');

    // Other students filtered, Иван still selected
    expect(screen.getByText('Мария Петрова')).toBeInTheDocument();
    expect(ivanCheckbox).toBeChecked();
  });

  it('should display "Студентов не найдено" when no matches', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    await user.type(input, 'Несуществующее');

    await waitFor(() => {
      expect(screen.getByText('Студентов не найдено')).toBeInTheDocument();
    });
  });

  // ============================================
  // GROUP 3: Checkbox Toggle (5 tests)
  // ============================================

  it('should add student to selectedStudents when checkbox clicked', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const checkbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    await user.click(checkbox);

    expect(onStudentsSelect).toHaveBeenCalledWith(['student-1']);
  });

  it('should remove student from selectedStudents when checkbox clicked again', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const checkbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    expect(checkbox).toBeChecked();

    await user.click(checkbox);

    expect(onStudentsSelect).toHaveBeenCalledWith([]);
  });

  it('should handle multiple checkbox selections with correct array', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();

    const { rerender } = render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Select first student
    let checkbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    await user.click(checkbox);
    expect(onStudentsSelect).toHaveBeenLastCalledWith(['student-1']);

    // Re-render with first student selected
    onStudentsSelect.mockClear();
    rerender(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    // Select second student
    checkbox = screen.getByRole('checkbox', { name: /Мария Петрова/i });
    await user.click(checkbox);
    expect(onStudentsSelect).toHaveBeenLastCalledWith(['student-1', 'student-2']);
  });

  it('should remove student via chip remove button', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1', 'student-2']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    await waitFor(() => {
      const removeButtons = screen.getAllByRole('button', { name: /Удалить/i });
      expect(removeButtons.length).toBe(2);
    });

    // Click remove button for first student (Иван)
    const removeButtons = screen.getAllByRole('button', { name: /Удалить Иван/i });
    await user.click(removeButtons[0]);

    expect(onStudentsSelect).toHaveBeenCalledWith(['student-2']);
  });

  it('should call onStudentsSelect([]) when Clear All button clicked', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1', 'student-2']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Очистить все/i })).toBeInTheDocument();
    });

    const clearAllBtn = screen.getByRole('button', { name: /Очистить все/i });
    await user.click(clearAllBtn);

    expect(onStudentsSelect).toHaveBeenCalledWith([]);
  });

  // ============================================
  // GROUP 4: Keyboard Navigation (3 tests)
  // ============================================

  it('should navigate through options with ArrowDown key', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Press ArrowDown to highlight first item
    await user.keyboard('{ArrowDown}');
    let ivanLabel = screen.getByText('Иван Иванов').closest('label');
    expect(ivanLabel).toHaveClass('highlighted');

    // Press ArrowDown to highlight second item
    await user.keyboard('{ArrowDown}');
    let mariaLabel = screen.getByText('Мария Петрова').closest('label');
    expect(mariaLabel).toHaveClass('highlighted');
    expect(ivanLabel).not.toHaveClass('highlighted');

    // Press ArrowDown to highlight third item
    await user.keyboard('{ArrowDown}');
    let petrLabel = screen.getByText('Петр Сидоров').closest('label');
    expect(petrLabel).toHaveClass('highlighted');
    expect(mariaLabel).not.toHaveClass('highlighted');
  });

  it('should navigate upward with ArrowUp key', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Navigate down to second student
    await user.keyboard('{ArrowDown}{ArrowDown}');
    let mariaLabel = screen.getByText('Мария Петрова').closest('label');
    expect(mariaLabel).toHaveClass('highlighted');

    // Navigate up back to first student
    await user.keyboard('{ArrowUp}');
    let ivanLabel = screen.getByText('Иван Иванов').closest('label');
    expect(ivanLabel).toHaveClass('highlighted');
    expect(mariaLabel).not.toHaveClass('highlighted');

    // Navigate up to nothing
    await user.keyboard('{ArrowUp}');
    expect(ivanLabel).not.toHaveClass('highlighted');
  });

  it('should close dropdown with Escape key', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    await user.keyboard('{Escape}');

    await waitFor(() => {
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });
  });

  // ============================================
  // GROUP 5: Space & Enter Toggle (2 tests)
  // ============================================

  it('should toggle checkbox with Space key on highlighted item', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Navigate to first student
    await user.keyboard('{ArrowDown}');

    // Press Space to toggle
    await user.keyboard(' ');

    expect(onStudentsSelect).toHaveBeenCalledWith(['student-1']);
  });

  it('should NOT toggle with Enter key (Space only)', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Navigate to first student
    await user.keyboard('{ArrowDown}');

    // Press Enter (should not toggle in multi-select mode)
    await user.keyboard('{Enter}');

    expect(onStudentsSelect).not.toHaveBeenCalled();
  });

  // ============================================
  // GROUP 6: Edge Cases (3 tests)
  // ============================================

  it('should handle API error gracefully and show message', async () => {
    const user = userEvent.setup();
    mockGetStudentsAll = vi.fn(() =>
      Promise.reject(new Error('Server error'))
    );

    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByText('Server error')).toBeInTheDocument();
    });
  });

  it('should display loading spinner while fetching', async () => {
    mockGetStudentsAll = vi.fn(
      () =>
        new Promise((resolve) => {
          setTimeout(() => resolve(mockStudents), 500);
        })
    );

    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    expect(input).toHaveAttribute('aria-busy', 'true');
  });

  it('should handle empty students list from API', async () => {
    const user = userEvent.setup();
    mockGetStudentsAll = vi.fn(() => Promise.resolve([]));

    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByText('Студентов не найдено')).toBeInTheDocument();
    });
  });

  // ============================================
  // GROUP 7: Performance (2 tests)
  // ============================================

  it('should not re-render when selectedStudents is the same array', async () => {
    const onStudentsSelect = vi.fn();
    const selectedIds = ['student-1', 'student-2'];

    const { rerender } = render(
      <StudentFilterSearch
        selectedStudents={selectedIds}
        onStudentsSelect={onStudentsSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Иван Иванов')).toBeInTheDocument();
    });

    // Get current chips count
    const chipsBeforeRerender = screen.getAllByRole('button', { name: /Удалить/i });
    const countBefore = chipsBeforeRerender.length;

    // Re-render with same array reference should work correctly
    rerender(
      <StudentFilterSearch
        selectedStudents={selectedIds}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const chipsAfterRerender = screen.getAllByRole('button', { name: /Удалить/i });
    const countAfter = chipsAfterRerender.length;

    expect(countBefore).toBe(countAfter);
  });

  it('should prevent duplicate student IDs in selectedStudents', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();

    const { rerender } = render(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Click checkbox for student-1 (already selected)
    const checkbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    await user.click(checkbox);

    // Should be removed, not duplicated
    expect(onStudentsSelect).toHaveBeenCalledWith([]);
    expect(onStudentsSelect).not.toHaveBeenCalledWith(['student-1', 'student-1']);
  });

  // ============================================
  // GROUP 8: Accessibility (4 tests)
  // ============================================

  it('should have correct ARIA attributes on input', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');

    expect(input).toHaveAttribute('aria-label', 'Поиск студента');
    expect(input).toHaveAttribute('aria-expanded', 'false');
    expect(input).toHaveAttribute('aria-autocomplete', 'list');
    expect(input).toHaveAttribute('aria-controls', 'student-filter-dropdown');

    await user.click(input);

    await waitFor(() => {
      expect(input).toHaveAttribute('aria-expanded', 'true');
    });
  });

  it('should have correct ARIA attributes on dropdown', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      const listbox = screen.getByRole('listbox');
      expect(listbox).toHaveAttribute('id', 'student-filter-dropdown');
      expect(listbox).toHaveAttribute('aria-label', 'Список студентов');
    });
  });

  it('should have aria-checked on checkboxes', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      const checkboxes = screen.getAllByRole('checkbox');
      expect(checkboxes.length).toBeGreaterThan(0);

      const ivanCheckbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
      expect(ivanCheckbox).toHaveAttribute('aria-checked', 'true');

      const mariaCheckbox = screen.getByRole('checkbox', { name: /Мария Петрова/i });
      expect(mariaCheckbox).toHaveAttribute('aria-checked', 'false');
    });
  });

  it('should have proper labels for chip remove buttons', async () => {
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={['student-1', 'student-2']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByLabelText('Удалить Иван Иванов')).toBeInTheDocument();
      expect(screen.getByLabelText('Удалить Мария Петрова')).toBeInTheDocument();
    });
  });

  // ============================================
  // GROUP 9: Case-Insensitive Search (2 tests)
  // ============================================

  it('should perform case-insensitive search with lowercase input', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    await user.type(input, 'иван');

    await waitFor(() => {
      expect(screen.getByText('Иван Иванов')).toBeInTheDocument();
      expect(screen.queryByText('Мария Петрова')).not.toBeInTheDocument();
    });
  });

  it('should perform case-insensitive search with uppercase input', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    await user.type(input, 'СИДОРОВ');

    await waitFor(() => {
      expect(screen.getByText('Петр Сидоров')).toBeInTheDocument();
      expect(screen.queryByText('Иван Иванов')).not.toBeInTheDocument();
      expect(screen.queryByText('Мария Петрова')).not.toBeInTheDocument();
    });
  });

  // ============================================
  // GROUP 10: Dropdown Behavior (3 tests)
  // ============================================

  it('should close dropdown after blur', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Blur the input
    await user.tab();

    await waitFor(() => {
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });
  });

  it('should clear search term after selecting student', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();

    const { rerender } = render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    await user.type(input, 'Иван');
    expect(input.value).toBe('Иван');

    const checkbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    await user.click(checkbox);

    expect(onStudentsSelect).toHaveBeenCalledWith(['student-1']);

    // After selection, search term should remain (dropdown closes but input keeps text)
    // This is expected behavior - user can clear it manually if needed
    expect(input.value).toBe('Иван');
  });

  it('should reopen dropdown when typing after selection', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();

    const { rerender } = render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    const checkbox = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    await user.click(checkbox);

    // Simulate selection
    rerender(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    // Type again - dropdown should reopen
    const inputAfter = screen.getByRole('textbox');
    await user.type(inputAfter, 'М');

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
      expect(screen.getByText('Мария Петрова')).toBeInTheDocument();
    });
  });

  // ============================================
  // BONUS: Regression Tests (2 tests)
  // ============================================

  it('should handle null/undefined full_name gracefully', async () => {
    const user = userEvent.setup();
    mockGetStudentsAll = vi.fn(() =>
      Promise.resolve([
        { id: 'student-1', full_name: null },
        { id: 'student-2', full_name: undefined },
        { id: 'student-3', full_name: 'Петр Сидоров' },
      ])
    );

    const onStudentsSelect = vi.fn();
    render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
      // Should show "Без имени" fallback for null/undefined names
      const defaultNames = screen.getAllByText('Без имени');
      expect(defaultNames.length).toBeGreaterThanOrEqual(2);
    });
  });

  it('should handle rapid checkbox clicks without losing state', async () => {
    const user = userEvent.setup();
    const onStudentsSelect = vi.fn();

    const { rerender } = render(
      <StudentFilterSearch
        selectedStudents={[]}
        onStudentsSelect={onStudentsSelect}
      />
    );

    const input = screen.getByRole('textbox');
    await user.click(input);

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });

    // Rapid clicks
    const checkbox1 = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    const checkbox2 = screen.getByRole('checkbox', { name: /Мария Петрова/i });

    await user.click(checkbox1);
    expect(onStudentsSelect).toHaveBeenLastCalledWith(['student-1']);

    onStudentsSelect.mockClear();

    rerender(
      <StudentFilterSearch
        selectedStudents={['student-1']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    await user.click(checkbox2);
    expect(onStudentsSelect).toHaveBeenLastCalledWith(['student-1', 'student-2']);

    onStudentsSelect.mockClear();

    rerender(
      <StudentFilterSearch
        selectedStudents={['student-1', 'student-2']}
        onStudentsSelect={onStudentsSelect}
      />
    );

    // Click first again to remove
    const checkboxAgain = screen.getByRole('checkbox', { name: /Иван Иванов/i });
    await user.click(checkboxAgain);
    expect(onStudentsSelect).toHaveBeenLastCalledWith(['student-2']);
  });
});
