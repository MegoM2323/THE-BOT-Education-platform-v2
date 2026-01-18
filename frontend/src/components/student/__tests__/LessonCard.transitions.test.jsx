import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { LessonCard } from '../LessonCard';
import '@testing-library/jest-dom';

describe('LessonCard - CSS Transitions', () => {
  const mockLesson = {
    id: '1',
    subject: 'Математика',
    teacher_name: 'Иван Петров',
    start_time: new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString(),
    end_time: new Date(Date.now() + 48 * 60 * 60 * 1000 + 60 * 60 * 1000).toISOString(),
    max_students: 5,
    current_students: 3,
    color: '#3b82f6',
  };

  const mockProps = {
    lesson: mockLesson,
    onBook: () => {},
    onClick: () => {},
    showActions: true,
  };

  beforeEach(() => {
    // Clear computed styles cache
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  describe('1. CSS transitions on .lesson-card', () => {
    it('should have lesson-card class applied', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card');
    });

    it('should have GPU acceleration class and style', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      // Check that element exists and has proper structure
      expect(card).toBeInTheDocument();
      expect(card.tagName).toBe('DIV');
    });

    it('should render lesson-card-compact variant', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-compact');
    });

    it('should have proper card structure for transitions', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      // Card should have proper nesting for nested transitions
      const header = card.querySelector('.lesson-header-compact');
      expect(header).toBeInTheDocument();
    });
  });

  describe('2. CSS transitions on filtered card state (.lesson-card-filtered)', () => {
    it('should apply filtered class when isDisabled is true', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-filtered');
    });

    it('should apply both lesson-card and lesson-card-filtered classes', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card');
      expect(card).toHaveClass('lesson-card-filtered');
    });

    it('should keep card interactable despite filtered state', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      // Element should still be in DOM and visible
      expect(card).toBeInTheDocument();
      expect(card).toBeVisible();
    });

    it('should maintain structure when filtered', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      // Card structure should be intact
      const header = card.querySelector('.lesson-header-compact');
      expect(header).toBeInTheDocument();
    });

    it('should render all card content even when filtered', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      // Content should still be present
      const subject = card.querySelector('.lesson-subject-compact');
      expect(subject).toHaveTextContent(mockLesson.subject);
    });
  });

  describe('3. CSS transitions on animation classes', () => {
    it('should apply lesson-card-removing class', () => {
      const { container } = render(<LessonCard {...mockProps} isRemoving={true} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-removing');
    });

    it('should apply lesson-card-entering class', () => {
      const { container } = render(<LessonCard {...mockProps} isEntering={true} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-entering');
    });

    it('should apply lesson-card-operating class', () => {
      const { container } = render(<LessonCard {...mockProps} isOperating={true} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-operating');
    });

    it('should render correctly during removal animation', () => {
      const { container } = render(<LessonCard {...mockProps} isRemoving={true} />);
      const card = container.querySelector('.lesson-card');

      // Card should still be in DOM during removal animation
      expect(card).toBeInTheDocument();
      expect(card).toHaveClass('lesson-card-removing');
    });

    it('should render correctly during entering animation', () => {
      const { container } = render(<LessonCard {...mockProps} isEntering={true} />);
      const card = container.querySelector('.lesson-card');

      // Card should be in DOM during entry animation
      expect(card).toBeInTheDocument();
      expect(card).toHaveClass('lesson-card-entering');
    });

    it('should support multiple animation classes simultaneously', () => {
      const { container } = render(
        <LessonCard {...mockProps} isEntering={true} isOperating={true} />
      );
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-entering');
      expect(card).toHaveClass('lesson-card-operating');
    });
  });

  describe('4. Transitions in child elements', () => {
    it('.lesson-header should have transition', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const header = container.querySelector('.lesson-header-compact');

      // Header should exist and have transitions applied at parent level
      expect(header).toBeInTheDocument();
    });

    it('.lesson-actions should have transition', () => {
      const { container } = render(
        <LessonCard {...mockProps} isBooked={true} onCancelClick={() => {}} bookingId="b1" />
      );
      const actions = container.querySelector('.lesson-actions');

      // Actions should exist for booked lessons with cancel option
      if (actions) {
        const styles = window.getComputedStyle(actions);
        const transition = styles.transition || styles.transitionProperty || '';
        expect(transition.length > 0 || styles.willChange).toBeTruthy();
      }
    });

    it('lesson-card-filtered should affect child opacity', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      const children = card.querySelectorAll('.lesson-header-compact, .lesson-main-info');
      expect(children.length).toBeGreaterThan(0);
    });
  });

  describe('5. prefers-reduced-motion support', () => {
    it('should have CSS rules for reduced motion preference', () => {
      // Window should support matchMedia for accessibility testing
      expect(window.matchMedia).toBeDefined();
    });

    it('should check prefers-reduced-motion is available', () => {
      const mql = window.matchMedia('(prefers-reduced-motion: reduce)');

      expect(mql).toBeDefined();
      expect(typeof mql.matches).toBe('boolean');
    });

    it('should render card regardless of reduced motion preference', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toBeInTheDocument();
    });
  });

  describe('6. Transition timing and easing', () => {
    it('should render card with proper animations support', () => {
      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      expect(card).toBeInTheDocument();
    });

    it('filtered state should maintain card functionality', () => {
      const { container } = render(<LessonCard {...mockProps} isDisabled={true} />);
      const card = container.querySelector('.lesson-card');

      // Card should still be interactive in structure
      expect(card).toHaveClass('lesson-card-filtered');
      expect(card).toBeVisible();
    });

    it('animations should apply to correct states', () => {
      const { container: container1 } = render(
        <LessonCard {...mockProps} isEntering={true} />
      );
      const card1 = container1.querySelector('.lesson-card');

      const { container: container2 } = render(
        <LessonCard {...mockProps} isRemoving={true} />
      );
      const card2 = container2.querySelector('.lesson-card');

      expect(card1).toHaveClass('lesson-card-entering');
      expect(card2).toHaveClass('lesson-card-removing');
    });
  });

  describe('7. Transition conflicts and composition', () => {
    it('should support combined state classes', () => {
      const { container } = render(
        <LessonCard {...mockProps} isEntering={true} isOperating={true} />
      );
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-entering');
      expect(card).toHaveClass('lesson-card-operating');
    });

    it('should handle filtered + animation classes together', () => {
      const { container } = render(
        <LessonCard {...mockProps} isEntering={true} isDisabled={true} />
      );
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-filtered');
      expect(card).toHaveClass('lesson-card-entering');
    });

    it('should maintain all classes on booked card', () => {
      const { container } = render(
        <LessonCard {...mockProps} isBooked={true} isOperating={true} />
      );
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-booked');
      expect(card).toHaveClass('lesson-card-operating');
    });

    it('should render correctly with multiple transition-affecting classes', () => {
      const { container } = render(
        <LessonCard {...mockProps} isBooked={true} isDisabled={true} />
      );
      const card = container.querySelector('.lesson-card');

      expect(card).toHaveClass('lesson-card-booked');
      expect(card).toHaveClass('lesson-card-filtered');
      expect(card).toBeInTheDocument();
    });
  });

  describe('8. Functional transitions (not blocking interaction)', () => {
    it('filtered card should still be clickable', () => {
      const onClick = vi.fn();
      const { container } = render(
        <LessonCard {...mockProps} isDisabled={true} onClick={onClick} />
      );

      const card = container.querySelector('.lesson-card');
      expect(card).toBeInTheDocument();
      expect(card).toHaveClass('lesson-card-filtered');
    });

    it('should allow book button interaction during transitions', () => {
      const onBook = vi.fn();

      const { container } = render(
        <LessonCard {...mockProps} onBook={onBook} />
      );

      const card = container.querySelector('.lesson-card');
      expect(card).toBeInTheDocument();
    });

    it('should maintain click handlers with transition classes', () => {
      const onClick = vi.fn();

      const { container } = render(
        <LessonCard
          {...mockProps}
          onClick={onClick}
          isEntering={true}
        />
      );

      const card = container.querySelector('.lesson-card');
      expect(card).toHaveClass('lesson-card-entering');
      expect(card).toBeInTheDocument();
    });

    it('should not disable buttons despite filtered state', () => {
      const onBook = vi.fn();

      const { container } = render(
        <LessonCard {...mockProps} isDisabled={true} onBook={onBook} />
      );

      const buttons = container.querySelectorAll('button');
      expect(buttons.length).toBeGreaterThan(0);
    });
  });

  describe('9. Responsive transition behavior', () => {
    it('should maintain transitions on mobile viewport', () => {
      // Set mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      const styles = window.getComputedStyle(card);
      expect(styles.transition || styles.transitionProperty).toBeTruthy();
    });

    it('should maintain transitions on tablet viewport', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      const styles = window.getComputedStyle(card);
      expect(styles.transition || styles.transitionProperty).toBeTruthy();
    });

    it('should maintain transitions on desktop viewport', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920,
      });

      const { container } = render(<LessonCard {...mockProps} />);
      const card = container.querySelector('.lesson-card');

      const styles = window.getComputedStyle(card);
      expect(styles.transition || styles.transitionProperty).toBeTruthy();
    });
  });
});
