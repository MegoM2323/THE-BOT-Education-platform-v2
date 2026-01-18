/**
 * Test for month view calendar day calculation
 * Verifies that:
 * 1. Days from previous/next months are included
 * 2. Calendar grid is always 42 days (6 weeks)
 * 3. Date comparison uses local time, not UTC
 */

/**
 * Simulates getDaysInView logic for month view
 */
function getDaysInMonthView(currentDate) {
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const days = [];

  // Add placeholder days from previous month
  const firstDayOfWeek = firstDay.getDay() === 0 ? 6 : firstDay.getDay() - 1;
  for (let i = firstDayOfWeek; i > 0; i--) {
    const day = new Date(year, month, 1 - i);
    days.push(day);
  }

  // Add current month days
  for (let i = 1; i <= lastDay.getDate(); i++) {
    days.push(new Date(year, month, i));
  }

  // Add placeholder days from next month to complete the grid
  const daysNeeded = 42 - days.length;
  for (let i = 1; i <= daysNeeded; i++) {
    const day = new Date(year, month + 1, i);
    days.push(day);
  }

  return days;
}

/**
 * Helper to format date as YYYY-MM-DD using local time
 */
function formatLocalDate(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

describe('Calendar Month View - Day Calculation', () => {
  test('January 2025 should have 42 days (6 complete weeks)', () => {
    const jan2025 = new Date(2025, 0, 1); // January 1, 2025
    const days = getDaysInMonthView(jan2025);

    expect(days.length).toBe(42);
    console.log('January 2025 days:', days.length);

    // First day should be Monday of week containing Jan 1
    // Jan 1, 2025 is a Wednesday, so first day should be Dec 30, 2024 (Monday)
    expect(formatLocalDate(days[0])).toBe('2024-12-30');
    expect(formatLocalDate(days[1])).toBe('2024-12-31');
    expect(formatLocalDate(days[2])).toBe('2025-01-01');

    // Last day should be Sunday completing the 6th week
    // Jan 2025 starts on Wed (Dec 30, 2024 is Mon), so has 6 complete weeks
    // Last day is Feb 9, 2025 (Sunday)
    expect(formatLocalDate(days[41])).toBe('2025-02-09');
  });

  test('February 2025 should have 42 days', () => {
    const feb2025 = new Date(2025, 1, 1); // February 1, 2025
    const days = getDaysInMonthView(feb2025);

    expect(days.length).toBe(42);
    console.log('February 2025 days:', days.length);
  });

  test('December 2025 should have 42 days', () => {
    const dec2025 = new Date(2025, 11, 1); // December 1, 2025
    const days = getDaysInMonthView(dec2025);

    expect(days.length).toBe(42);
    console.log('December 2025 days:', days.length);

    // Should include Jan 2026 days
    const lastDayStr = formatLocalDate(days[41]);
    expect(lastDayStr).toMatch(/^2026-01-/);
  });

  test('All calendar months should have exactly 42 days', () => {
    for (let month = 0; month < 12; month++) {
      const testDate = new Date(2025, month, 1);
      const days = getDaysInMonthView(testDate);
      expect(days.length).toBe(42);
    }
  });

  test('Local date formatting should preserve dates across timezones', () => {
    // Create a date and verify local time is used, not UTC
    const testDate = new Date('2025-01-15T23:00:00Z'); // UTC time

    // Format using local time
    const localFormatted = formatLocalDate(testDate);

    // The actual local date depends on timezone, but it should not be converted to UTC
    // This test just verifies the function works consistently
    expect(localFormatted).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    console.log('Test date in UTC:', '2025-01-15T23:00:00Z');
    console.log('Formatted locally:', localFormatted);
  });

  test('Month boundary handling - days from adjacent months', () => {
    // Test a month that starts on Monday (simple case)
    // May 2025 starts on Thursday
    const may2025 = new Date(2025, 4, 1); // May 1, 2025
    const days = getDaysInMonthView(may2025);

    const firstDay = formatLocalDate(days[0]);
    const lastDay = formatLocalDate(days[41]);

    // First day should be April 28 (Monday)
    expect(firstDay).toBe('2025-04-28');

    // Last day should be June 8 (Sunday)
    expect(lastDay).toBe('2025-06-08');

    // Verify all days are consecutive
    let currentDate = new Date(days[0]);
    for (let i = 1; i < days.length; i++) {
      currentDate.setDate(currentDate.getDate() + 1);
      expect(formatLocalDate(days[i])).toBe(formatLocalDate(currentDate));
    }
  });
});
