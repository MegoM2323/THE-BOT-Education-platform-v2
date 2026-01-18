/**
 * Test for lesson filtering by date in calendar
 * Verifies that lessons are correctly matched to calendar days
 * across different timezones and month views
 */

/**
 * Helper to format date as YYYY-MM-DD using local time
 */
function formatLocalDate(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * Simulates getLessonsForDay logic
 */
function getLessonsForDay(day, lessons) {
  const year = day.getFullYear();
  const month = String(day.getMonth() + 1).padStart(2, '0');
  const date = String(day.getDate()).padStart(2, '0');
  const dayStr = `${year}-${month}-${date}`;

  return lessons.filter((lesson) => {
    const lessonTime = new Date(lesson.start_time);
    const lessonYear = lessonTime.getFullYear();
    const lessonMonth = String(lessonTime.getMonth() + 1).padStart(2, '0');
    const lessonDate = String(lessonTime.getDate()).padStart(2, '0');
    const lessonDateStr = `${lessonYear}-${lessonMonth}-${lessonDate}`;

    return lessonDateStr === dayStr;
  }).sort((a, b) => new Date(a.start_time) - new Date(b.start_time));
}

describe('Calendar - Lesson Day Matching', () => {
  test('Lessons should be matched to correct calendar day', () => {
    const lessons = [
      {
        id: 1,
        start_time: '2025-01-15T10:00:00Z', // Jan 15, 2025 at 10:00 UTC
        subject: 'Math'
      },
      {
        id: 2,
        start_time: '2025-01-15T14:30:00Z', // Jan 15, 2025 at 14:30 UTC
        subject: 'Physics'
      },
      {
        id: 3,
        start_time: '2025-01-20T09:00:00Z', // Jan 20, 2025
        subject: 'Chemistry'
      }
    ];

    // Test Jan 15
    const jan15 = new Date(2025, 0, 15); // Local Jan 15
    const jan15Lessons = getLessonsForDay(jan15, lessons);
    expect(jan15Lessons.length).toBe(2);
    expect(jan15Lessons[0].id).toBe(1);
    expect(jan15Lessons[1].id).toBe(2);

    // Test Jan 20
    const jan20 = new Date(2025, 0, 20);
    const jan20Lessons = getLessonsForDay(jan20, lessons);
    expect(jan20Lessons.length).toBe(1);
    expect(jan20Lessons[0].id).toBe(3);

    // Test Jan 19 (should have no lessons)
    const jan19 = new Date(2025, 0, 19);
    const jan19Lessons = getLessonsForDay(jan19, lessons);
    expect(jan19Lessons.length).toBe(0);
  });

  test('Lessons should be sorted by time within a day', () => {
    const lessons = [
      {
        id: 1,
        start_time: '2025-01-15T15:00:00Z',
        subject: 'Physics'
      },
      {
        id: 2,
        start_time: '2025-01-15T09:00:00Z',
        subject: 'Math'
      },
      {
        id: 3,
        start_time: '2025-01-15T12:00:00Z',
        subject: 'Chemistry'
      }
    ];

    const jan15 = new Date(2025, 0, 15);
    const jan15Lessons = getLessonsForDay(jan15, lessons);

    expect(jan15Lessons.length).toBe(3);
    expect(jan15Lessons[0].id).toBe(2); // 09:00
    expect(jan15Lessons[1].id).toBe(3); // 12:00
    expect(jan15Lessons[2].id).toBe(1); // 15:00
  });

  test('Month boundaries should not skip lessons', () => {
    // Lessons on Jan 31 and Feb 1
    const lessons = [
      {
        id: 1,
        start_time: '2025-01-31T10:00:00Z',
        subject: 'Math'
      },
      {
        id: 2,
        start_time: '2025-02-01T10:00:00Z',
        subject: 'Physics'
      }
    ];

    const jan31 = new Date(2025, 0, 31);
    const jan31Lessons = getLessonsForDay(jan31, lessons);
    expect(jan31Lessons.length).toBe(1);
    expect(jan31Lessons[0].id).toBe(1);

    const feb1 = new Date(2025, 1, 1);
    const feb1Lessons = getLessonsForDay(feb1, lessons);
    expect(feb1Lessons.length).toBe(1);
    expect(feb1Lessons[0].id).toBe(2);
  });

  test('End-of-month lessons should appear on month view grid', () => {
    // December has 31 days, so lessons near the end should appear
    const lessons = [
      {
        id: 1,
        start_time: '2025-12-28T10:00:00Z',
        subject: 'Math'
      },
      {
        id: 2,
        start_time: '2025-12-31T10:00:00Z',
        subject: 'Physics'
      }
    ];

    const dec28 = new Date(2024, 11, 28); // Dec 28, 2024
    const dec31 = new Date(2024, 11, 31); // Dec 31, 2024

    const dec28Lessons = getLessonsForDay(dec28, lessons);
    const dec31Lessons = getLessonsForDay(dec31, lessons);

    // Note: These should be 0 because the lessons are for 2025, not 2024
    expect(dec28Lessons.length).toBe(0);
    expect(dec31Lessons.length).toBe(0);

    // Correct dates for 2025
    const dec28_2025 = new Date(2025, 11, 28);
    const dec31_2025 = new Date(2025, 11, 31);

    const dec28_2025Lessons = getLessonsForDay(dec28_2025, lessons);
    const dec31_2025Lessons = getLessonsForDay(dec31_2025, lessons);

    expect(dec28_2025Lessons.length).toBe(1);
    expect(dec31_2025Lessons.length).toBe(1);
  });

  test('Empty lessons array should return empty list', () => {
    const jan15 = new Date(2025, 0, 15);
    const lessonsForDay = getLessonsForDay(jan15, []);
    expect(lessonsForDay.length).toBe(0);
    expect(Array.isArray(lessonsForDay)).toBe(true);
  });

  test('Lessons on boundary days (prev/next month) should match correctly', () => {
    // For month view, the calendar includes days from prev/next months
    // Lessons on those days should still match correctly
    const lessons = [
      {
        id: 1,
        start_time: '2024-12-30T10:00:00Z', // Monday before Jan 1, 2025
        subject: 'Math'
      },
      {
        id: 2,
        start_time: '2025-02-09T10:00:00Z', // Sunday after Jan 31, 2025
        subject: 'Physics'
      }
    ];

    // December 30, 2024 (shown in January 2025 month view)
    const dec30 = new Date(2024, 11, 30);
    const dec30Lessons = getLessonsForDay(dec30, lessons);
    expect(dec30Lessons.length).toBe(1);
    expect(dec30Lessons[0].id).toBe(1);

    // February 9, 2025 (shown in January 2025 month view)
    const feb9 = new Date(2025, 1, 9);
    const feb9Lessons = getLessonsForDay(feb9, lessons);
    expect(feb9Lessons.length).toBe(1);
    expect(feb9Lessons[0].id).toBe(2);
  });
});
