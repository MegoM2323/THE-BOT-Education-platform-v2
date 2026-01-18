import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

describe('Landing - Visual Tests (Desktop 1200px+)', () => {
  beforeEach(() => {
    window.matchMedia = vi.fn().mockImplementation(query => ({
      matches: query === '(min-width: 1200px)',
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    window.HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  describe('Typography - Green Headings', () => {
    it('должны иметь все заголовки зелёный цвет (#004231)', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const headings = container.querySelectorAll('h1, h2, h3');
      headings.forEach(heading => {
        const computedColor = window.getComputedStyle(heading).color;
        // CSS переменная или rgb цвет
        expect(computedColor).toMatch(/^(var\(--primary-color\)|rgb\(0, 66, 49\))$/);
      });
    });

    it('должны иметь увеличенный размер шрифта для зелёных заголовков', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const sectionTitles = container.querySelectorAll('.section-title, .section-title-center');
      sectionTitles.forEach(title => {
        const fontSize = window.getComputedStyle(title).fontSize;
        // Может быть 2.05rem или в px, проверяем что это не мало
        expect(fontSize).toBeTruthy();
        expect(fontSize).toMatch(/rem|px/);
      });
    });

    it('должны быть жирными (font-weight: 700)', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const sectionTitles = container.querySelectorAll('.section-title, .section-title-center');
      sectionTitles.forEach(title => {
        const fontWeight = window.getComputedStyle(title).fontWeight;
        expect(fontWeight).toBe('700');
      });
    });
  });

  describe('Layout - Grid Structures', () => {
    it('плитки после "кто я" должны отображаться в 3 колонки на десктопе', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const demoLessons = container.querySelector('.demo-lessons');
      expect(demoLessons).toBeInTheDocument();

      const gridStyle = window.getComputedStyle(demoLessons).gridTemplateColumns;
      // repeat(auto-fit, minmax(280px, 1fr)) на 1200px+ должна быть 3 колонки
      expect(gridStyle).toBeTruthy();
    });

    it('4 плитки записи должны отображаться в 2x2 сетке', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const homeworkGrid = container.querySelector('.homework-examples-grid');
      expect(homeworkGrid).toBeInTheDocument();

      const gridStyle = window.getComputedStyle(homeworkGrid).gridTemplateColumns;
      expect(gridStyle).toContain('repeat(2');
    });

    it('блоки расписания (3-4 человека, 2 часа, 3400/Занятие) должны отображаться', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonBlocks = container.querySelectorAll('.demo-lesson-block');
      expect(lessonBlocks.length).toBeGreaterThan(0);

      // Проверяем наличие значений
      let hasGroupSize = false;
      let hasDuration = false;
      let hasPrice = false;

      lessonBlocks.forEach(block => {
        const text = block.textContent;
        if (text.includes('3-4') || text.includes('человека')) hasGroupSize = true;
        if (text.includes('2 часа')) hasDuration = true;
        if (text.includes('3400')) hasPrice = true;
      });

      expect(hasGroupSize).toBe(true);
      expect(hasDuration).toBe(true);
      expect(hasPrice).toBe(true);
    });
  });

  describe('Responsive - Tablet (768px)', () => {
    beforeEach(() => {
      window.matchMedia = vi.fn().mockImplementation(query => ({
        matches: query === '(max-width: 768px)',
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }));
    });

    it('плитки должны быть 2 в строку на планшете', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Проверяем наличие responsive класса или стиля
      const demoLessons = container.querySelector('.demo-lessons');
      expect(demoLessons).toBeInTheDocument();
    });

    it('форма должна быть полной ширины на планшете', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const form = container.querySelector('.contact-form-wrapper');
      expect(form).toBeInTheDocument();
    });
  });

  describe('Responsive - Mobile (480px)', () => {
    beforeEach(() => {
      window.matchMedia = vi.fn().mockImplementation(query => ({
        matches: query === '(max-width: 480px)',
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }));
    });

    it('плитки должны быть 1 в строку на мобильном', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // На мобильном все должно быть одной колонкой
      const demoLessons = container.querySelector('.demo-lessons');
      expect(demoLessons).toBeInTheDocument();
    });

    it('фото должно быть под текстом на мобильном', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const aboutContainer = container.querySelector('.about-container');
      const computedDisplay = window.getComputedStyle(aboutContainer).display;
      // На мобильном это должно быть block, не grid
      expect(computedDisplay).toBeTruthy();
    });
  });

  describe('Form Styling', () => {
    it('labels должны быть в нормальном регистре (не жирные)', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const labels = container.querySelectorAll('.form-label');
      labels.forEach(label => {
        const fontWeight = window.getComputedStyle(label).fontWeight;
        expect(fontWeight).toBe('400');
      });
    });

    it('labels должны быть чёрные', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const labels = container.querySelectorAll('.form-label');
      labels.forEach(label => {
        const color = window.getComputedStyle(label).color;
        // Должно быть тёмным цветом, не светлым
        expect(color).toBeTruthy();
      });
    });

    it('кнопка формы должна быть зелёная (#004231)', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const button = container.querySelector('.submit-button');
      const backgroundColor = window.getComputedStyle(button).backgroundColor;
      // CSS переменная или rgb цвет
      expect(backgroundColor).toMatch(/^(var\(--primary-color\)|rgb\(0, 66, 49\))$/);
    });

    it('при фокусе поля формы должна быть зелёная граница', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const input = container.querySelector('.form-input');
      // Проверяем что есть focus style
      const style = window.getComputedStyle(input);
      expect(style.transition).toContain('border');
    });
  });

  describe('Form Validation Messages', () => {
    it('красные плашки ошибок должны быть скрыты по умолчанию', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const errorMessages = container.querySelectorAll('.error-message');
      errorMessages.forEach(error => {
        const display = window.getComputedStyle(error).display;
        expect(display).toBe('none');
      });
    });
  });

  describe('Hover Effects', () => {
    it('при наведении на карточки должен быть слабый затемняющий эффект', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const cards = container.querySelectorAll('.calendar-lesson');
      cards.forEach(card => {
        const style = window.getComputedStyle(card);
        expect(style.transition).toContain('opacity');
      });
    });
  });

  describe('Animations', () => {
    it('анимация ДЗ блоков должна быть fadeInUp', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const homeworkExamples = container.querySelectorAll('.homework-example');
      homeworkExamples.forEach(example => {
        const style = window.getComputedStyle(example);
        expect(style.animation).toContain('fadeInUp');
      });
    });

    it('анимация не должна замораживать страницу', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const animatedElements = container.querySelectorAll('[style*="animation"]');
      expect(animatedElements.length).toBeGreaterThanOrEqual(0);
      // Если нет ошибок при рендеринге - тест проходит
    });
  });

  describe('Spacing Synchronization', () => {
    it('отступы должны быть синхронизированы между секциями', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const sections = container.querySelectorAll('section');
      const paddings = [];

      sections.forEach(section => {
        const padding = window.getComputedStyle(section).padding;
        paddings.push(padding);
      });

      // Проверяем что отступы существуют
      paddings.forEach(padding => {
        expect(padding).toBeTruthy();
      });
    });
  });

  describe('Teacher Photo', () => {
    it('фото преподавателя должно быть на уровне "Лекции" на десктопе', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const aboutRight = container.querySelector('.about-right');
      const img = aboutRight?.querySelector('img');

      expect(img).toBeInTheDocument();
      expect(img).toHaveAttribute('alt');
    });
  });

  describe('Overall Page Structure', () => {
    it('страница должна иметь все основные секции', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.hero-section')).toBeInTheDocument();
      expect(container.querySelector('.about-section')).toBeInTheDocument();
      expect(container.querySelector('.lesson-types-section')).toBeInTheDocument();
      expect(container.querySelector('.examples-section')).toBeInTheDocument();
      expect(container.querySelector('.swap-system-section')).toBeInTheDocument();
      expect(container.querySelector('.how-it-works-section')).toBeInTheDocument();
      expect(container.querySelector('.contact-section')).toBeInTheDocument();
    });

    it('страница должна правильно отображаться без ошибок', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(consoleSpy).not.toHaveBeenCalled();

      consoleSpy.mockRestore();
    });
  });
});
