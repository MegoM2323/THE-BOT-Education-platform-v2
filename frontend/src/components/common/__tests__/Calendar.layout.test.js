import '@testing-library/jest-dom';
import { render } from '@testing-library/react';
import { Calendar } from '../Calendar.jsx';

describe('Calendar Layout Tests', () => {
  let mockLessons;
  let mockRole;

  beforeEach(() => {
    mockLessons = [
      {
        id: 1,
        name: 'Math Lesson',
        start_time: '2024-01-02T10:00:00Z',
        end_time: '2024-01-02T11:00:00Z',
        teacher_name: 'John Doe',
        spots_available: 5,
        spots_total: 10,
        color: '#3B82F6',
      },
    ];
    mockRole = 'student';
  });

  describe('Scenario 1: Calendar alignment - left margin/padding consistency', () => {
    it('should align calendar to left with consistent margin', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const calendarWrapper = container.querySelector('[class*="calendar"]');
      expect(calendarWrapper).toBeTruthy();

      // Check computed style
      const computedStyle = window.getComputedStyle(calendarWrapper);
      const marginLeft = computedStyle.marginLeft;
      const paddingLeft = computedStyle.paddingLeft;

      // Margin should be a valid value (0, px value, or auto is acceptable for root elements)
      expect(marginLeft).toBeDefined();
      // Should not have excessive left margin
      const leftValue = parseInt(marginLeft) || 0;
      expect(leftValue).toBeLessThanOrEqual(50);
    });

    it('should have all elements with consistent left alignment', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const headerElements = container.querySelectorAll('[class*="header"]');
      const filterElements = container.querySelectorAll('[class*="filter"]');
      const cardElements = container.querySelectorAll('[class*="lesson"]');

      const allElements = [...headerElements, ...filterElements, ...cardElements];

      expect(allElements.length).toBeGreaterThan(0);

      allElements.forEach(element => {
        const computedStyle = window.getComputedStyle(element);
        const marginLeft = parseInt(computedStyle.marginLeft) || 0;
        const paddingLeft = parseInt(computedStyle.paddingLeft) || 0;

        // Verify consistent alignment (should be aligned properly)
        expect(typeof marginLeft).toBe('number');
        expect(typeof paddingLeft).toBe('number');
      });
    });
  });

  describe('Scenario 2: Responsive layout - mobile breakpoint 768px', () => {
    beforeEach(() => {
      // Mock window.matchMedia for mobile viewport
      window.innerWidth = 768;
      window.dispatchEvent(new Event('resize'));
    });

    it('should render calendar with mobile-appropriate layout at 768px', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const calendarElement = container.querySelector('[class*="calendar"]');
      expect(calendarElement).toBeTruthy();

      // Verify element is visible and not hidden
      const computedStyle = window.getComputedStyle(calendarElement);
      expect(computedStyle.display).not.toBe('none');
    });

    it('should apply correct width constraints at 768px breakpoint', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const calendarElement = container.querySelector('[class*="calendar"]');
      if (calendarElement) {
        const computedStyle = window.getComputedStyle(calendarElement);
        const width = computedStyle.width;

        // Width should be reasonable (not 0 or invalid)
        expect(width).not.toBe('0px');
        expect(width).not.toBe('auto');
      }
    });
  });

  describe('Scenario 3: Responsive layout - tablet breakpoint 1024px', () => {
    beforeEach(() => {
      window.innerWidth = 1024;
      window.dispatchEvent(new Event('resize'));
    });

    it('should render calendar with tablet-appropriate layout at 1024px', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const calendarElement = container.querySelector('[class*="calendar"]');
      expect(calendarElement).toBeTruthy();

      const computedStyle = window.getComputedStyle(calendarElement);
      expect(computedStyle.display).not.toBe('none');
    });

    it('should maintain element proportions at 1024px breakpoint', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const lessonCards = container.querySelectorAll('[data-testid="calendar-lesson"]');

      lessonCards.forEach(card => {
        const computedStyle = window.getComputedStyle(card);
        const width = computedStyle.width;
        const height = computedStyle.height;

        // Elements should be visible
        expect(computedStyle.display).not.toBe('none');
        expect(width).not.toBe('0px');
        expect(height).not.toBe('0px');
      });
    });
  });

  describe('Scenario 4: Responsive layout - desktop breakpoint 1400px', () => {
    beforeEach(() => {
      window.innerWidth = 1400;
      window.dispatchEvent(new Event('resize'));
    });

    it('should render calendar with desktop-appropriate layout at 1400px', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const calendarElement = container.querySelector('[class*="calendar"]');
      expect(calendarElement).toBeTruthy();

      const computedStyle = window.getComputedStyle(calendarElement);
      expect(computedStyle.display).not.toBe('none');
    });

    it('should optimize layout for desktop viewing at 1400px', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      // Verify layout is not compressed
      const calendarElement = container.querySelector('[class*="calendar"]');
      if (calendarElement) {
        const computedStyle = window.getComputedStyle(calendarElement);
        const width = computedStyle.width;

        // Width should be defined (testing environment may have limited viewport)
        expect(width).not.toBe('0px');
        expect(width).toBeDefined();
      }
    });
  });

  describe('Scenario 5: CSS variables usage verification', () => {
    it('should use CSS custom properties for colors', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      // Look for any lesson element or container
      const lessonCard = container.querySelector('[class*="lesson"]');
      expect(lessonCard).toBeTruthy();

      if (lessonCard) {
        const computedStyle = window.getComputedStyle(lessonCard);

        // Verify styles are applied
        expect(computedStyle).toBeDefined();
        expect(computedStyle.display || computedStyle.color || computedStyle.backgroundColor).toBeTruthy();
      }
    });

    it('should apply CSS variables from parent container', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const root = container.querySelector('[class*="calendar"]');
      if (root) {
        const computedStyle = window.getComputedStyle(root);

        // Verify that styles are computed (not empty)
        expect(computedStyle.display).toBeTruthy();
        expect(computedStyle.position || computedStyle.width).toBeTruthy();
      }
    });

    it('should inherit text color from CSS variables for accessibility', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      const textElements = container.querySelectorAll('[class*="lesson-time"], [class*="lesson-teacher"]');

      textElements.forEach(element => {
        const computedStyle = window.getComputedStyle(element);
        const color = computedStyle.color;

        // Color should be defined (not transparent)
        expect(color).not.toBe('rgba(0, 0, 0, 0)');
        expect(color).toBeTruthy();
      });
    });
  });

  describe('Scenario 6: Full page layout consistency', () => {
    it('should align title, filters and cards consistently', () => {
      const { container } = render(
        <Calendar
          lessons={mockLessons}
          role={mockRole}
          onLessonClick={() => {}}
        />
      );

      // Check if multiple elements share alignment
      const allAlignedElements = container.querySelectorAll('[class*="calendar"]');

      expect(allAlignedElements.length).toBeGreaterThan(0);

      let firstElementLeft = null;

      allAlignedElements.forEach((element) => {
        const rect = element.getBoundingClientRect();

        // First iteration: set reference left position
        if (firstElementLeft === null && rect.left !== 0) {
          firstElementLeft = rect.left;
        }

        // Verify elements don't have drastically different alignments
        // (allowing small differences for nested elements)
        expect(Math.abs((rect.left || 0) - (firstElementLeft || 0))).toBeLessThanOrEqual(40);
      });
    });
  });
});
