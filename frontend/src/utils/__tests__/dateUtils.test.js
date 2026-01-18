/**
 * Unit tests for date utility functions
 * Tests getMonday() and getDaysInView() logic
 */

/**
 * Get Monday of current week (ISO week numbering)
 * Monday = day 1, Sunday = day 0
 * Formula: (day + 6) % 7 gives days since Monday
 */
function getMonday(date = new Date()) {
  const d = new Date(date);
  const day = d.getDay();
  const daysSinceMonday = (day + 6) % 7;
  d.setDate(d.getDate() - daysSinceMonday);
  d.setHours(0, 0, 0, 0);
  return d;
}

/**
 * Get days in view (week or month)
 */
function getDaysInView(currentDate, view = 'week') {
  if (view === 'week') {
    const startOfWeek = new Date(currentDate);
    const day = startOfWeek.getDay();
    const daysSinceMonday = (day + 6) % 7;
    startOfWeek.setDate(currentDate.getDate() - daysSinceMonday);
    startOfWeek.setHours(0, 0, 0, 0);
    const days = [];
    for (let i = 0; i < 7; i++) {
      const day = new Date(startOfWeek);
      day.setDate(startOfWeek.getDate() + i);
      days.push(day);
    }
    return days;
  } else {
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const days = [];

    const firstDayOfWeek = firstDay.getDay() === 0 ? 6 : firstDay.getDay() - 1;
    for (let i = firstDayOfWeek; i > 0; i--) {
      const day = new Date(year, month, 1 - i);
      days.push(day);
    }

    for (let i = 1; i <= lastDay.getDate(); i++) {
      days.push(new Date(year, month, i));
    }

    const daysNeeded = 42 - days.length;
    for (let i = 1; i <= daysNeeded; i++) {
      const day = new Date(year, month + 1, i);
      days.push(day);
    }

    return days;
  }
}

describe('getMonday()', () => {
  test('Monday (day=1) input returns same Monday', () => {
    // Create a Monday: 2025-01-06
    const monday = new Date(2025, 0, 6); // January 6, 2025 is Monday
    const result = getMonday(monday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0); // January
    expect(result.getDate()).toBe(6);
    expect(result.getDay()).toBe(1); // Monday
  });

  test('Tuesday (day=2) input returns Monday of same week', () => {
    // Create a Tuesday: 2025-01-07
    const tuesday = new Date(2025, 0, 7); // January 7, 2025 is Tuesday
    const result = getMonday(tuesday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0);
    expect(result.getDate()).toBe(6); // Should be Monday
    expect(result.getDay()).toBe(1);
  });

  test('Wednesday (day=3) input returns Monday of same week', () => {
    // Create a Wednesday: 2025-01-08
    const wednesday = new Date(2025, 0, 8); // January 8, 2025 is Wednesday
    const result = getMonday(wednesday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0);
    expect(result.getDate()).toBe(6); // Should be Monday
    expect(result.getDay()).toBe(1);
  });

  test('Thursday (day=4) input returns Monday of same week', () => {
    // Create a Thursday: 2025-01-09
    const thursday = new Date(2025, 0, 9); // January 9, 2025 is Thursday
    const result = getMonday(thursday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0);
    expect(result.getDate()).toBe(6); // Should be Monday
    expect(result.getDay()).toBe(1);
  });

  test('Friday (day=5) input returns Monday of same week', () => {
    // Create a Friday: 2025-01-10
    const friday = new Date(2025, 0, 10); // January 10, 2025 is Friday
    const result = getMonday(friday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0);
    expect(result.getDate()).toBe(6); // Should be Monday
    expect(result.getDay()).toBe(1);
  });

  test('Saturday (day=6) input returns Monday of SAME week, NOT next', () => {
    // CRITICAL TEST: Saturday should go back to Monday of same week
    // Create a Saturday: 2025-01-11
    const saturday = new Date(2025, 0, 11); // January 11, 2025 is Saturday
    const result = getMonday(saturday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0);
    expect(result.getDate()).toBe(6); // Should be Monday, NOT January 13 (next Monday)
    expect(result.getDay()).toBe(1);
  });

  test('Sunday (day=0) input returns Monday of SAME week, NOT next', () => {
    // CRITICAL TEST: Sunday should go back to Monday of same week
    // Create a Sunday: 2025-01-12
    const sunday = new Date(2025, 0, 12); // January 12, 2025 is Sunday
    const result = getMonday(sunday);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0);
    expect(result.getDate()).toBe(6); // Should be Monday, NOT January 13 (next Monday)
    expect(result.getDay()).toBe(1);
  });

  test('Last day of month does not overflow to next month', () => {
    // January 31, 2025 is Friday
    const lastDay = new Date(2025, 0, 31);
    const result = getMonday(lastDay);

    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(0); // Should stay in January
    expect(result.getDate()).toBe(27); // Monday of same week
    expect(result.getDay()).toBe(1);
  });

  test('First day of month does not underflow to previous month', () => {
    // January 1, 2025 is Wednesday
    const firstDay = new Date(2025, 0, 1);
    const result = getMonday(firstDay);

    expect(result.getFullYear()).toBe(2024); // Monday is in previous year
    expect(result.getMonth()).toBe(11); // December
    expect(result.getDate()).toBe(30); // Monday before Jan 1
    expect(result.getDay()).toBe(1);
  });

  test('Returns date with time set to 00:00:00.000', () => {
    const date = new Date(2025, 0, 15, 14, 30, 45, 123);
    const result = getMonday(date);

    expect(result.getHours()).toBe(0);
    expect(result.getMinutes()).toBe(0);
    expect(result.getSeconds()).toBe(0);
    expect(result.getMilliseconds()).toBe(0);
  });

  test('Handles leap year correctly', () => {
    // 2024 is a leap year, February has 29 days
    // February 28, 2024 is Wednesday
    const date = new Date(2024, 1, 28);
    const result = getMonday(date);

    expect(result.getFullYear()).toBe(2024);
    expect(result.getMonth()).toBe(1); // February
    expect(result.getDate()).toBe(26); // Monday of same week
    expect(result.getDay()).toBe(1);
  });
});

describe('getDaysInView() - Week View', () => {
  test('Returns 7 days for week view', () => {
    const date = new Date(2025, 0, 15); // Wednesday
    const result = getDaysInView(date, 'week');

    expect(result).toHaveLength(7);
  });

  test('Week starts on Monday', () => {
    const date = new Date(2025, 0, 15); // Wednesday
    const result = getDaysInView(date, 'week');

    expect(result[0].getDay()).toBe(1); // Monday
  });

  test('Week ends on Sunday', () => {
    const date = new Date(2025, 0, 15); // Wednesday
    const result = getDaysInView(date, 'week');

    expect(result[6].getDay()).toBe(0); // Sunday
  });

  test('All days in week are consecutive', () => {
    const date = new Date(2025, 0, 15); // Wednesday
    const result = getDaysInView(date, 'week');

    for (let i = 1; i < 7; i++) {
      const dayDiff = (result[i].getTime() - result[i - 1].getTime()) / (24 * 60 * 60 * 1000);
      expect(dayDiff).toBe(1);
    }
  });

  test('Week view from Monday returns correct week', () => {
    const monday = new Date(2025, 0, 6); // Monday
    const result = getDaysInView(monday, 'week');

    expect(result[0].getDate()).toBe(6); // Monday
    expect(result[1].getDate()).toBe(7); // Tuesday
    expect(result[2].getDate()).toBe(8); // Wednesday
    expect(result[3].getDate()).toBe(9); // Thursday
    expect(result[4].getDate()).toBe(10); // Friday
    expect(result[5].getDate()).toBe(11); // Saturday
    expect(result[6].getDate()).toBe(12); // Sunday
  });

  test('Week view from Sunday returns previous Monday', () => {
    const sunday = new Date(2025, 0, 12); // Sunday
    const result = getDaysInView(sunday, 'week');

    expect(result[0].getDate()).toBe(6); // Previous Monday
    expect(result[6].getDate()).toBe(12); // Same Sunday
  });
});

describe('getDaysInView() - Month View', () => {
  test('Returns 42 days for month view', () => {
    const date = new Date(2025, 0, 15); // January 2025
    const result = getDaysInView(date, 'month');

    expect(result).toHaveLength(42);
  });

  test('Month view includes days from previous month', () => {
    const date = new Date(2025, 0, 15); // January 2025
    const result = getDaysInView(date, 'month');

    // January 1, 2025 is Wednesday, so previous month days are included
    expect(result[0].getMonth()).toBe(11); // December
  });

  test('Month view includes all days of current month', () => {
    const date = new Date(2025, 0, 15); // January 2025
    const result = getDaysInView(date, 'month');

    const januaryDays = result.filter(d => d.getMonth() === 0);
    expect(januaryDays.length).toBe(31);
  });

  test('Month view includes days from next month', () => {
    const date = new Date(2025, 0, 15); // January 2025
    const result = getDaysInView(date, 'month');

    const febDays = result.filter(d => d.getMonth() === 1);
    expect(febDays.length).toBeGreaterThan(0);
  });

  test('Month view grid always starts on Monday', () => {
    const date = new Date(2025, 0, 15); // January 2025
    const result = getDaysInView(date, 'month');

    expect(result[0].getDay()).toBe(1); // Monday
  });

  test('Month view grid always ends on Sunday', () => {
    const date = new Date(2025, 0, 15); // January 2025
    const result = getDaysInView(date, 'month');

    expect(result[41].getDay()).toBe(0); // Sunday
  });

  test('Month view for February (non-leap year)', () => {
    const date = new Date(2023, 1, 15); // February 2023
    const result = getDaysInView(date, 'month');

    expect(result).toHaveLength(42);
    const febDays = result.filter(d => d.getMonth() === 1);
    expect(febDays.length).toBe(28); // 2023 is not a leap year
  });

  test('Month view for February (leap year)', () => {
    const date = new Date(2024, 1, 15); // February 2024
    const result = getDaysInView(date, 'month');

    expect(result).toHaveLength(42);
    const febDays = result.filter(d => d.getMonth() === 1);
    expect(febDays.length).toBe(29); // 2024 is a leap year
  });
});
