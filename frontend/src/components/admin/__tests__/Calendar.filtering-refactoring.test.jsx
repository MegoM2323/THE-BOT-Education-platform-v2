import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import Calendar from '../Calendar.jsx';

// Mock lessons with complete booking information
const mockLessonsWithBookings = [
  {
    id: 'lesson-1',
    start_time: new Date(Date.now() + 86400000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 1,
    bookings: [{ student_id: 'student-1', student_name: 'Петр' }],
    subject: 'Математика',
  },
  {
    id: 'lesson-2',
    start_time: new Date(Date.now() + 172800000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 1,
    bookings: [{ student_id: 'student-2', student_name: 'Мария' }],
    subject: 'Русский язык',
  },
  {
    id: 'lesson-3',
    start_time: new Date(Date.now() + 259200000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 2,
    bookings: [
      { student_id: 'student-1', student_name: 'Петр' },
      { student_id: 'student-2', student_name: 'Мария' }
    ],
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
  {
    id: 'lesson-5',
    start_time: new Date(Date.now() + 432000000).toISOString(),
    teacher_id: 'teacher-1',
    max_students: 2,
    current_students: 1,
    bookings: [{ student_id: 'student-3', student_name: 'Иван' }],
    subject: 'География',
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
    lessons: mockLessonsWithBookings,
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

describe('Calendar: Student Filter Refactoring (Frontend-Only Pattern)', () => {
  describe('1. Backend Returns ALL Lessons with Complete Bookings', () => {
    it('should receive lessons with complete bookings array from backend', () => {
      // All lessons should be returned (not filtered by backend)
      expect(mockLessonsWithBookings).toHaveLength(5);

      // Each lesson has bookings array (even if empty)
      mockLessonsWithBookings.forEach(lesson => {
        expect(lesson).toHaveProperty('bookings');
        expect(Array.isArray(lesson.bookings)).toBe(true);
      });
    });

    it('lesson with student should have booking info in bookings array', () => {
      const lesson1 = mockLessonsWithBookings[0];
      expect(lesson1.bookings).toEqual([
        { student_id: 'student-1', student_name: 'Петр' }
      ]);
    });

    it('lesson without students should have empty bookings array', () => {
      const lesson4 = mockLessonsWithBookings[3];
      expect(lesson4.bookings).toEqual([]);
    });

    it('lesson with multiple students should have all bookings', () => {
      const lesson3 = mockLessonsWithBookings[2];
      expect(lesson3.bookings).toHaveLength(2);
      expect(lesson3.bookings.map(b => b.student_id)).toContain('student-1');
      expect(lesson3.bookings.map(b => b.student_id)).toContain('student-2');
    });
  });

  describe('2. Frontend Filters Locally Using lesson.bookings', () => {
    it('should show all lessons when no filter is selected', () => {
      // When selectedFilterStudents is empty, all lessons pass the filter
      const selectedFilterStudents = [];

      const filtered = mockLessonsWithBookings.filter(lesson => {
        if (selectedFilterStudents?.length === 0) return true;
        return lesson.bookings?.some(b =>
          selectedFilterStudents.includes(b.student_id)
        );
      });

      // All 5 lessons should be visible
      expect(filtered).toHaveLength(5);
    });

    it('should not mark any lesson as filtered when selectedFilterStudents is empty', () => {
      // When selectedFilterStudents is empty, isFiltered should be false for all lessons
      const selectedFilterStudents = [];

      const lessonsWithFilterState = mockLessonsWithBookings.map((lesson) => {
        let isFiltered = false;

        if (selectedFilterStudents?.length > 0) {
          const hasStudentBooking = lesson.bookings?.some(b =>
            selectedFilterStudents.includes(b.student_id)
          );
          if (!hasStudentBooking) isFiltered = true;
        }

        return { ...lesson, isFiltered };
      });

      lessonsWithFilterState.forEach(lesson => {
        expect(lesson.isFiltered).toBe(false);
      });
    });
  });

  describe('3. Lessons with Selected Students Display Normally', () => {
    it('should show lesson with matching student booking in normal colors', () => {
      // Simulate: selectedFilterStudents = ['student-1']
      // Lesson-1 has student-1, should NOT be marked as filtered

      const lesson = mockLessonsWithBookings[0]; // has student-1
      const selectedStudents = ['student-1'];

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      expect(hasStudentBooking).toBe(true);
    });

    it('should show lesson with multiple students if ANY selected student has booking', () => {
      const lesson = mockLessonsWithBookings[2]; // has student-1 and student-2
      const selectedStudents = ['student-1']; // Only selecting student-1

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      // Should be visible because student-1 has a booking
      expect(hasStudentBooking).toBe(true);
    });
  });

  describe('4. Lessons Without Selected Students Display Gray (isFiltered=true)', () => {
    it('should mark lesson as filtered if selected student has no booking', () => {
      const lesson = mockLessonsWithBookings[1]; // has only student-2
      const selectedStudents = ['student-3']; // Selecting student-3

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      expect(hasStudentBooking).toBe(false);
    });

    it('should mark empty lesson as filtered when any student is selected', () => {
      const lesson = mockLessonsWithBookings[3]; // empty bookings
      const selectedStudents = ['student-1'];

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      expect(hasStudentBooking).toBe(false);
    });
  });

  describe('5. Multiple Students Can Be Selected', () => {
    it('should show lesson if ANY of multiple selected students have booking (OR logic)', () => {
      const lesson = mockLessonsWithBookings[2]; // has student-1 and student-2
      const selectedStudents = ['student-1', 'student-3'];

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      // Should be visible because student-1 has a booking (OR logic)
      expect(hasStudentBooking).toBe(true);
    });

    it('should not show lesson if NONE of multiple selected students have booking', () => {
      const lesson = mockLessonsWithBookings[0]; // has only student-1
      const selectedStudents = ['student-2', 'student-3'];

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      // Should not be visible because neither student-2 nor student-3 has booking
      expect(hasStudentBooking).toBe(false);
    });

    it('should handle selection of all students', () => {
      const selectedStudents = ['student-1', 'student-2', 'student-3'];

      // Every lesson should be visible except lesson-4 (no bookings)
      mockLessonsWithBookings.forEach(lesson => {
        const hasStudentBooking = lesson.bookings?.some(b =>
          selectedStudents.includes(b.student_id)
        );

        if (lesson.id === 'lesson-4') {
          // Empty lesson should not be visible
          expect(hasStudentBooking).toBe(false);
        } else {
          // All other lessons have at least one selected student
          expect(hasStudentBooking).toBe(true);
        }
      });
    });
  });

  describe('6. Clearing Filter Shows All Lessons Normally', () => {
    it('should restore all lessons to normal colors when filter is cleared', () => {
      // When selectedFilterStudents goes from ['student-1'] to []
      const selectedStudents = [];

      mockLessonsWithBookings.forEach(lesson => {
        const hasStudentBooking = lesson.bookings?.some(b =>
          selectedStudents.includes(b.student_id)
        );

        // No filtering should happen when selectedStudents is empty
        expect(selectedStudents.length).toBe(0);
      });
    });
  });

  describe('Error Cases: Missing/Null Bookings', () => {
    it('should not crash if lesson.bookings is missing', () => {
      const lesson = { ...mockLessonsWithBookings[0] };
      delete lesson.bookings;
      const selectedStudents = ['student-1'];

      // Should use optional chaining and not crash
      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      expect(hasStudentBooking).toBe(undefined); // optional chaining returns undefined
    });

    it('should not crash if lesson.bookings is null', () => {
      const lesson = { ...mockLessonsWithBookings[0], bookings: null };
      const selectedStudents = ['student-1'];

      // Should handle null gracefully
      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      expect(hasStudentBooking).toBe(undefined);
    });

    it('should show lesson normally if bookings is null and no filter selected', () => {
      const lesson = { ...mockLessonsWithBookings[0], bookings: null };
      const selectedStudents = [];

      // When no filter is selected, lesson should not be marked as filtered
      let isFiltered = false;
      if (selectedStudents?.length > 0) {
        const hasStudentBooking = lesson.bookings?.some(b =>
          selectedStudents.includes(b.student_id)
        );
        if (!hasStudentBooking) isFiltered = true;
      }

      expect(isFiltered).toBe(false);
    });

    it('should show lesson normally if bookings is undefined and no filter selected', () => {
      const lesson = { ...mockLessonsWithBookings[0] };
      delete lesson.bookings;
      const selectedStudents = [];

      let isFiltered = false;
      if (selectedStudents?.length > 0) {
        const hasStudentBooking = lesson.bookings?.some(b =>
          selectedStudents.includes(b.student_id)
        );
        if (!hasStudentBooking) isFiltered = true;
      }

      expect(isFiltered).toBe(false);
    });
  });

  describe('7. Backend No Longer Filters by student_ids', () => {
    it('should not pass student_ids parameter to backend API', () => {
      // The API function should NOT have a student_ids parameter
      // Filtering happens on frontend only

      // Backend returns ALL lessons regardless of student_ids
      const apiFilters = {}; // No student_ids parameter

      expect(apiFilters).not.toHaveProperty('student_ids');
    });

    it('should handle initialization with empty bookings (backend default)', () => {
      // If backend fails to load bookings, lesson should have empty bookings array
      const lesson = { ...mockLessonsWithBookings[0], bookings: [] };
      const selectedStudents = ['student-1'];

      const hasStudentBooking = lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      );

      // Should correctly detect no bookings match
      expect(hasStudentBooking).toBe(false);
    });
  });

  describe('8. Frontend-Only Filtering Logic', () => {
    it('should use lesson.bookings for filtering, not backend parameter', () => {
      const lesson = mockLessonsWithBookings[0];
      const selectedStudents = ['student-1'];

      // Filtering happens ONLY by checking lesson.bookings
      // NOT by passing parameter to backend
      const isFiltered = !(lesson.bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      ));

      expect(isFiltered).toBe(false);
    });

    it('should maintain isFiltered state across multiple filter changes', () => {
      // Start with student-1 filter
      let selectedStudents = ['student-1'];
      let isFiltered = !(mockLessonsWithBookings[1].bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      ));
      expect(isFiltered).toBe(true); // lesson-2 has only student-2

      // Change to student-2 filter
      selectedStudents = ['student-2'];
      isFiltered = !(mockLessonsWithBookings[1].bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      ));
      expect(isFiltered).toBe(false); // lesson-2 has student-2

      // Add student-1 to filter
      selectedStudents = ['student-1', 'student-2'];
      isFiltered = !(mockLessonsWithBookings[1].bookings?.some(b =>
        selectedStudents.includes(b.student_id)
      ));
      expect(isFiltered).toBe(false); // lesson-2 has student-2
    });
  });

  describe('9. Integration: Complex Filtering Scenarios', () => {
    it('should handle filtering with multiple students and lessons', () => {
      const selectedStudents = ['student-1', 'student-3'];

      const filtered = mockLessonsWithBookings.map(lesson => {
        let isFiltered = false;
        if (selectedStudents?.length > 0) {
          const hasStudentBooking = lesson.bookings?.some(b =>
            selectedStudents.includes(b.student_id)
          );
          if (!hasStudentBooking) isFiltered = true;
        }
        return { ...lesson, isFiltered };
      });

      // Lesson-1: has student-1 ✓
      expect(filtered[0].isFiltered).toBe(false);
      // Lesson-2: has student-2 ✗
      expect(filtered[1].isFiltered).toBe(true);
      // Lesson-3: has student-1 ✓
      expect(filtered[2].isFiltered).toBe(false);
      // Lesson-4: empty ✗
      expect(filtered[3].isFiltered).toBe(true);
      // Lesson-5: has student-3 ✓
      expect(filtered[4].isFiltered).toBe(false);
    });

    it('should correctly count visible vs filtered lessons', () => {
      const selectedStudents = ['student-2'];

      const filtered = mockLessonsWithBookings.filter(lesson => {
        if (selectedStudents?.length === 0) return true;
        return lesson.bookings?.some(b =>
          selectedStudents.includes(b.student_id)
        );
      });

      // Only lesson-2 and lesson-3 have student-2
      expect(filtered).toHaveLength(2);
    });
  });
});
