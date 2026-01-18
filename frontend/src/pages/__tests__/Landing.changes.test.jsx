import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

// Mock notification functions
const mockNotification = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
};

vi.mock('../../hooks/useNotification.js', () => ({
  useNotification: () => mockNotification,
  default: () => mockNotification,
}));

describe('Landing Page - New Changes', () => {
  let consoleErrorSpy, consoleWarnSpy;

  beforeEach(() => {
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    mockNotification.success.mockClear();
    mockNotification.error.mockClear();
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
    consoleWarnSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe('Test 1: RecentLessonsSection удален', () => {
    it('should not render RecentLessonsSection title', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const recentLessonsTitle = screen.queryByText(/последние занятия/i);
      expect(recentLessonsTitle).not.toBeInTheDocument();
    });

    it('should not render RecentLessonsSection DOM element', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const recentLessonsSection = document.querySelector('.recent-lessons-section');
      expect(recentLessonsSection).not.toBeInTheDocument();
    });

    it('should render main sections in correct order', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('Кто я?')).toBeInTheDocument();
      expect(screen.getByText(/Записи на занятия/i)).toBeInTheDocument();
      expect(screen.getByText('THE BOT')).toBeInTheDocument();
    });
  });

  describe('Test 2: SwapSystemSection отображает 2x2 сетку', () => {
    it('should render 4 calendar-lesson cells', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarLessons = document.querySelectorAll('.calendar-lesson');
      expect(calendarLessons.length).toBe(4);
    });

    it('should display calendar-lesson cells in 2x2 grid on desktop (1200px+)', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarDemo = document.querySelector('.calendar-demo');
      expect(calendarDemo).toBeInTheDocument();

      const lessons = document.querySelectorAll('.calendar-lesson');
      expect(lessons.length).toBe(4);

      const computedStyle = window.getComputedStyle(calendarDemo);
      expect(computedStyle.display).toBe('grid');
    });

    it('should maintain 2x2 grid on tablet (768px)', () => {
      window.innerWidth = 768;
      window.innerHeight = 1024;
      window.dispatchEvent(new Event('resize'));

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarDemo = document.querySelector('.calendar-demo');
      const lessons = document.querySelectorAll('.calendar-lesson');

      expect(calendarDemo).toBeInTheDocument();
      expect(lessons.length).toBe(4);
      expect(window.getComputedStyle(calendarDemo).display).toBe('grid');
    });

    it('should change to 1 column on mobile (<480px)', () => {
      window.innerWidth = 480;
      window.innerHeight = 800;
      window.dispatchEvent(new Event('resize'));

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarDemo = document.querySelector('.calendar-demo');
      const lessons = document.querySelectorAll('.calendar-lesson');

      expect(calendarDemo).toBeInTheDocument();
      expect(lessons.length).toBe(4);
    });

    it('should have correct grid structure with 2 columns max', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarDemo = document.querySelector('.calendar-demo');
      expect(calendarDemo).toBeInTheDocument();

      const computedStyle = window.getComputedStyle(calendarDemo);
      expect(computedStyle.display).toBe('grid');

      const lessons = document.querySelectorAll('.calendar-lesson');
      expect(lessons.length).toBe(4);

      const lessonsArray = Array.from(lessons);
      lessonsArray.forEach((lesson) => {
        expect(lesson).toHaveClass('calendar-lesson');
      });
    });
  });

  describe('Test 3: Текст преподавателя обновлен', () => {
    it('should display "Мини группы" in first cell', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonTeachers = document.querySelectorAll('.calendar-lesson-teacher');
      expect(lessonTeachers[0].textContent).toContain('Мини группы');
    });

    it('should display "Продолжительность занятия" in second cell', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonTeachers = document.querySelectorAll('.calendar-lesson-teacher');
      expect(lessonTeachers[1].textContent).toContain('Продолжительность занятия');
    });

    it('should display "Цена фиксируется на старте" in third cell', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonTeachers = document.querySelectorAll('.calendar-lesson-teacher');
      expect(lessonTeachers[2].textContent).toContain('Цена фиксируется на старте');
    });

    it('should display "Мирослав Адаменко" in fourth cell', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonTeachers = document.querySelectorAll('.calendar-lesson-teacher');
      expect(lessonTeachers[3].textContent).toContain('Мирослав Адаменко');
    });

    it('should display all teacher texts correctly in order', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonTeachers = document.querySelectorAll('.calendar-lesson-teacher');

      expect(lessonTeachers.length).toBe(4);
      expect(lessonTeachers[0].textContent).toBe('Мини группы');
      expect(lessonTeachers[1].textContent).toBe('Продолжительность занятия');
      expect(lessonTeachers[2].textContent).toBe('Цена фиксируется на старте');
      expect(lessonTeachers[3].textContent).toBe('Мирослав Адаменко');
    });
  });

  describe('Test 4: Отсутствуют ошибки', () => {
    it('should not have console errors', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(consoleErrorSpy).not.toHaveBeenCalled();
    });

    it('should not have undefined variables in rendered content', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const pageHTML = document.documentElement.innerHTML;
      expect(pageHTML).not.toMatch(/\bundefined\b/);
      expect(pageHTML).not.toMatch(/\[object Object\]/);
      expect(pageHTML).not.toMatch(/\bNaN\b/);
    });

    it('should render page without React errors', () => {
      expect(() => {
        render(
          <BrowserRouter>
            <Landing />
          </BrowserRouter>
        );
      }).not.toThrow();
    });

    it('should render all calendar lessons without errors', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessons = document.querySelectorAll('.calendar-lesson');
      lessons.forEach((lesson) => {
        expect(lesson.querySelector('.calendar-lesson-subject')).toBeInTheDocument();
        expect(lesson.querySelector('.calendar-lesson-time')).toBeInTheDocument();
        expect(lesson.querySelector('.calendar-lesson-teacher')).toBeInTheDocument();
        expect(lesson.querySelector('.calendar-lesson-spots')).toBeInTheDocument();
      });
    });

    it('should not have console warnings', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(consoleWarnSpy).not.toHaveBeenCalled();
    });
  });
});
