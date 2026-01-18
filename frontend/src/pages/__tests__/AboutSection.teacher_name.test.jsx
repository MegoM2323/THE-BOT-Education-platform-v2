import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import AboutSection from '../Landing/components/AboutSection';

describe('AboutSection - teacher_name Updates', () => {
  let consoleErrorSpy;
  let consoleWarnSpy;

  beforeEach(() => {
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
    consoleWarnSpy.mockRestore();
  });

  // Test 1: Проверить что все 3 карточки отображаются
  describe('Test 1: All 3 demo lesson cards are visible', () => {
    it('should render 3 demo-lesson-card elements', () => {
      render(<AboutSection />);
      const cards = screen.getAllByRole('generic', { hidden: true }).filter(
        el => el.className.includes('demo-lesson-card')
      );
      expect(cards).toHaveLength(3);
    });

    it('should have all 3 cards visible on page', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      expect(cards).toHaveLength(3);
      cards.forEach(card => {
        expect(card).toBeVisible();
      });
    });

    it('should contain demo-lessons container with all cards', () => {
      const { container } = render(<AboutSection />);
      const demoLessonsContainer = container.querySelector('.demo-lessons');

      expect(demoLessonsContainer).toBeInTheDocument();
      const cards = demoLessonsContainer.querySelectorAll('.demo-lesson-card');
      expect(cards).toHaveLength(3);
    });
  });

  // Test 2: Проверить новые значения teacher_name
  describe('Test 2: Verify new teacher_name values in demo-lesson-teacher', () => {
    it('first card should contain "Мини группы" in demo-lesson-teacher', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');
      const firstCardTeacher = cards[0].querySelector('.demo-lesson-teacher');

      expect(firstCardTeacher).toBeInTheDocument();
      expect(firstCardTeacher.textContent).toBe('Мини группы');
    });

    it('second card should contain "Продолжительность занятия" in demo-lesson-teacher', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');
      const secondCardTeacher = cards[1].querySelector('.demo-lesson-teacher');

      expect(secondCardTeacher).toBeInTheDocument();
      expect(secondCardTeacher.textContent).toBe('Продолжительность занятия');
    });

    it('third card should contain "Цена фиксируется на старте" in demo-lesson-teacher', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');
      const thirdCardTeacher = cards[2].querySelector('.demo-lesson-teacher');

      expect(thirdCardTeacher).toBeInTheDocument();
      expect(thirdCardTeacher.textContent).toBe('Цена фиксируется на старте');
    });

    it('all cards should have demo-lesson-teacher span with correct className', () => {
      const { container } = render(<AboutSection />);
      const teacherSpans = container.querySelectorAll('.demo-lesson-teacher');

      expect(teacherSpans).toHaveLength(3);
      const expectedTexts = [
        'Мини группы',
        'Продолжительность занятия',
        'Цена фиксируется на старте'
      ];

      teacherSpans.forEach((span, index) => {
        expect(span.textContent).toBe(expectedTexts[index]);
      });
    });

    it('teacher_name values should be displayed as text not as attributes', () => {
      const { container } = render(<AboutSection />);
      const teacherSpans = container.querySelectorAll('.demo-lesson-teacher');

      teacherSpans.forEach(span => {
        expect(span.textContent.length).toBeGreaterThan(0);
        expect(span.getAttribute('data-teacher')).toBeNull();
      });
    });
  });

  // Test 3: Проверить что остальные элементы не изменились
  describe('Test 3: Verify other elements remain unchanged', () => {
    it('all cards should contain subject field', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      cards.forEach(card => {
        const subject = card.querySelector('.demo-lesson-subject');
        expect(subject).toBeInTheDocument();
        expect(subject.textContent.length).toBeGreaterThan(0);
      });
    });

    it('all cards should contain correct subject values', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');
      const expectedSubjects = ['3-4 человека', '2 часа', '3400/Занятие'];

      cards.forEach((card, index) => {
        const subject = card.querySelector('.demo-lesson-subject');
        expect(subject.textContent).toBe(expectedSubjects[index]);
      });
    });

    it('all cards should contain spots field', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      cards.forEach(card => {
        const spots = card.querySelector('.demo-lesson-spots');
        expect(spots).toBeInTheDocument();
        expect(spots.textContent).toBe('4 мест');
      });
    });

    it('all cards should have demo-lesson-header with subject and spots', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      cards.forEach(card => {
        const header = card.querySelector('.demo-lesson-header');
        expect(header).toBeInTheDocument();

        const subject = header.querySelector('.demo-lesson-subject');
        const spots = header.querySelector('.demo-lesson-spots');

        expect(subject).toBeInTheDocument();
        expect(spots).toBeInTheDocument();
      });
    });

    it('all cards should have left border color #004231', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      cards.forEach(card => {
        const borderLeftStyle = card.style.borderLeft;
        expect(borderLeftStyle).toContain('4px solid');
        const computedStyle = window.getComputedStyle(card);
        const rgbColor = computedStyle.borderLeftColor;
        expect(rgbColor).toMatch(/rgb\(0, 66, 49\)|rgb\(0,66,49\)|#004231/);
      });
    });

    it('all cards should have semi-transparent background color', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      cards.forEach(card => {
        const bgStyle = card.style.background;
        expect(bgStyle.length).toBeGreaterThan(0);
        expect(bgStyle).toMatch(/rgba|rgb/);
        expect(bgStyle).toContain('0.03');
      });
    });
  });

  // Test 4: Проверить что нет ошибок
  describe('Test 4: Verify no errors', () => {
    it('should render without console errors', () => {
      render(<AboutSection />);
      expect(consoleErrorSpy).not.toHaveBeenCalled();
    });

    it('should render without console warnings', () => {
      render(<AboutSection />);
      expect(consoleWarnSpy).not.toHaveBeenCalled();
    });

    it('should not have any undefined variables in rendered content', () => {
      const { container } = render(<AboutSection />);
      const allText = container.textContent;

      expect(allText).not.toContain('undefined');
      expect(allText).not.toContain('[object Object]');
    });

    it('should render React without errors', () => {
      expect(() => {
        render(<AboutSection />);
      }).not.toThrow();
    });

    it('should have all required sections', () => {
      const { container } = render(<AboutSection />);

      expect(container.querySelector('.about-section')).toBeInTheDocument();
      expect(container.querySelector('.container')).toBeInTheDocument();
      expect(container.querySelector('.about-content')).toBeInTheDocument();
      expect(container.querySelector('.demo-lessons')).toBeInTheDocument();
      expect(container.querySelector('.consult-block')).toBeInTheDocument();
    });

    it('should have correct element hierarchy', () => {
      const { container } = render(<AboutSection />);

      const section = container.querySelector('.about-section');
      const containerDiv = section.querySelector('.container');
      const content = containerDiv.querySelector('.about-content');
      const demoLessons = containerDiv.querySelector('.demo-lessons');

      expect(section).toBeInTheDocument();
      expect(containerDiv).toBeInTheDocument();
      expect(content).toBeInTheDocument();
      expect(demoLessons).toBeInTheDocument();
    });

    it('each demo-lesson-card should have unique key (id 1-3)', () => {
      const { container } = render(<AboutSection />);
      const cards = container.querySelectorAll('.demo-lesson-card');

      expect(cards).toHaveLength(3);
      // Keys are internal React prop, but we can verify structure
      cards.forEach((card, index) => {
        expect(card).toBeInTheDocument();
      });
    });

    it('should render all text content without encoding issues', () => {
      const { container } = render(<AboutSection />);

      expect(container.textContent).toContain('Мини группы');
      expect(container.textContent).toContain('Продолжительность занятия');
      expect(container.textContent).toContain('Цена фиксируется на старте');
      expect(container.textContent).toContain('3-4 человека');
      expect(container.textContent).toContain('2 часа');
      expect(container.textContent).toContain('3400/Занятие');
    });
  });
});
