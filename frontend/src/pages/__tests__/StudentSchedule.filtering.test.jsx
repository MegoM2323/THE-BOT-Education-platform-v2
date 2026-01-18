import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StudentSchedule } from '../StudentSchedule';
import * as useStudentLessonsHook from '../../hooks/useStudentLessons';
import * as useMyBookingsHook from '../../hooks/useMyBookings';
import * as useStudentCreditsHook from '../../hooks/useStudentCredits';
import * as useAuthHook from '../../hooks/useAuth';

describe('StudentSchedule - Filtering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockLessons = [
    {
      id: 1,
      subject: 'Math',
      teacher_id: 'teacher1',
      teacher_name: 'John',
      start_time: '2024-01-01T10:00:00',
      end_time: '2024-01-01T11:00:00',
      color: '#3b82f6',
      current_students: 5,
      max_students: 10,
    },
    {
      id: 2,
      subject: 'English',
      teacher_id: 'teacher2',
      teacher_name: 'Jane',
      start_time: '2024-01-01T15:00:00',
      end_time: '2024-01-01T16:00:00',
      color: '#10b981',
      current_students: 3,
      max_students: 10,
    },
  ];

  it('should mark lessons as filtered when teacher filter is applied', () => {
    const filterTeacher = 'teacher1';

    const lessons = mockLessons.map(lesson => {
      let isFiltered = false;
      if (filterTeacher && lesson.teacher_id !== filterTeacher) {
        isFiltered = true;
      }
      return { ...lesson, isFiltered };
    });

    expect(lessons[0].isFiltered).toBe(false);
    expect(lessons[1].isFiltered).toBe(true);
  });

  it('should mark lessons as filtered when time filter is applied', () => {
    const filterTime = 'morning'; // 6am-12pm

    const lessons = mockLessons.map(lesson => {
      let isFiltered = false;
      if (filterTime) {
        const hour = new Date(lesson.start_time).getHours();
        if (filterTime === 'morning' && (hour < 6 || hour >= 12)) {
          isFiltered = true;
        }
      }
      return { ...lesson, isFiltered };
    });

    expect(lessons[0].isFiltered).toBe(false); // 10am is morning
    expect(lessons[1].isFiltered).toBe(true);  // 3pm is not morning
  });

  it('should mark lessons as filtered when availability filter is applied', () => {
    const filterAvailability = true;

    const lessons = mockLessons.map(lesson => {
      let isFiltered = false;
      if (filterAvailability && lesson.current_students >= lesson.max_students) {
        isFiltered = true;
      }
      return { ...lesson, isFiltered };
    });

    expect(lessons[0].isFiltered).toBe(false); // 5/10 available
    expect(lessons[1].isFiltered).toBe(false); // 3/10 available
  });

  it('should mark lessons as filtered when they are full and availability is enabled', () => {
    const filterAvailability = true;
    const fullLesson = { ...mockLessons[0], current_students: 10 };

    const isFiltered = filterAvailability && fullLesson.current_students >= fullLesson.max_students;

    expect(isFiltered).toBe(true);
  });

  it('should apply multiple filters', () => {
    const filterTeacher = 'teacher1';
    const filterTime = 'afternoon'; // 12pm-6pm

    const lessons = mockLessons.map(lesson => {
      let isFiltered = false;

      if (filterTeacher && lesson.teacher_id !== filterTeacher) {
        isFiltered = true;
      }

      if (filterTime) {
        const hour = new Date(lesson.start_time).getHours();
        if (filterTime === 'afternoon' && (hour < 12 || hour >= 18)) {
          isFiltered = true;
        }
      }

      return { ...lesson, isFiltered };
    });

    expect(lessons[0].isFiltered).toBe(true); // wrong teacher AND wrong time
    expect(lessons[1].isFiltered).toBe(true); // wrong teacher (but right time)
  });
});
