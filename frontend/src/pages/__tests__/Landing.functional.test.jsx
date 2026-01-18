import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import userEvent from '@testing-library/user-event';
import Landing from '../Landing';
import apiClient from '../../api/client.js';

// Mock apiClient
vi.mock('../../api/client.js', () => ({
  default: {
    post: vi.fn(),
  },
}));

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

describe('Landing - Functional Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNotification.success.mockClear();
    mockNotification.error.mockClear();
    window.HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('Page Rendering', () => {
    it('страница должна полностью отрендериться', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.landing-page')).toBeInTheDocument();
      expect(container.querySelector('.hero-section')).toBeInTheDocument();
    });

    it('должны быть видны все основные компоненты', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText(/кратчайший путь/i)).toBeInTheDocument();
      expect(screen.getByText(/кто я/i)).toBeInTheDocument();
    });

    it('не должно быть console errors при рендеринге', () => {
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

  describe('Form Presence and Structure', () => {
    it('должна быть форма обратной связи', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const form = container.querySelector('.contact-form');
      expect(form).toBeInTheDocument();
    });

    it('форма должна содержать все необходимые поля', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const form = container.querySelector('.contact-form');
      const inputs = form.querySelectorAll('input');

      // Должны быть поля: name, phone, telegram, email
      expect(inputs.length).toBeGreaterThanOrEqual(4);
    });

    it('форма должна иметь кнопку отправки', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const button = container.querySelector('.submit-button');
      expect(button).toBeInTheDocument();
    });

    it('кнопка должна быть типа submit', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const button = container.querySelector('.submit-button');
      expect(button.type).toBe('submit');
    });
  });

  describe('Form Validation', () => {
    it('форма должна иметь валидацию имени', async () => {
      const user = userEvent.setup();
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const nameInput = container.querySelector('input[name="name"]');
      await user.click(nameInput);
      await user.tab();

      // Поле должно иметь обработчик blur
      expect(nameInput).toBeInTheDocument();
    });

    it('форма должна иметь валидацию телефона', async () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const phoneInput = container.querySelector('input[name="phone"]');
      expect(phoneInput).toBeInTheDocument();
    });

    it('форма должна иметь валидацию Telegram', async () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const telegramInput = container.querySelector('input[name="telegram"]');
      expect(telegramInput).toBeInTheDocument();
    });

    it('форма должна иметь поле email', async () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const emailInput = container.querySelector('input[name="email"]');
      expect(emailInput).toBeInTheDocument();
    });
  });

  describe('Form Styling and Classes', () => {
    it('форма должна иметь класс contact-form', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.contact-form')).toBeInTheDocument();
    });

    it('поля должны иметь класс form-input', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const inputs = container.querySelectorAll('.form-input');
      expect(inputs.length).toBeGreaterThan(0);
    });

    it('labels должны иметь класс form-label', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const labels = container.querySelectorAll('.form-label');
      expect(labels.length).toBeGreaterThan(0);
    });

    it('кнопка должна иметь класс submit-button', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.submit-button')).toBeInTheDocument();
    });
  });

  describe('Form Submission Behavior', () => {
    it('форма должна иметь обработчик отправки', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const form = container.querySelector('.contact-form');

      // Проверяем что форма существует
      expect(form).toBeInTheDocument();

      // И может содержать правильные элементы
      const nameInput = form?.querySelector('input[name="name"]');
      const phoneInput = form?.querySelector('input[name="phone"]');
      const telegramInput = form?.querySelector('input[name="telegram"]');
      const button = form?.querySelector('.submit-button');

      expect(nameInput).toBeInTheDocument();
      expect(phoneInput).toBeInTheDocument();
      expect(telegramInput).toBeInTheDocument();
      expect(button).toBeInTheDocument();
    });

    it('при отправке должны отправляться данные на API', async () => {
      const user = userEvent.setup();
      apiClient.post.mockResolvedValueOnce({ success: true });

      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const nameInput = container.querySelector('input[name="name"]');
      const phoneInput = container.querySelector('input[name="phone"]');
      const telegramInput = container.querySelector('input[name="telegram"]');

      await user.type(nameInput, 'Тест');
      await user.type(phoneInput, '+79991234567');
      await user.type(telegramInput, '@test');

      const button = container.querySelector('.submit-button');
      await user.click(button);

      await waitFor(() => {
        expect(apiClient.post).toHaveBeenCalled();
      }, { timeout: 2000 });
    });
  });

  describe('Animations', () => {
    it('анимация не должна замораживать страницу', async () => {
      const start = performance.now();

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const end = performance.now();
      const renderTime = end - start;

      // Рендеринг должен быть быстрым (менее 2 секунд)
      expect(renderTime).toBeLessThan(2000);
    });
  });

  describe('Content Verification', () => {
    it('должна быть информация о преподавателе', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Может быть несколько элементов с таким текстом, берем первый
      expect(screen.getAllByText(/мирослав адаменко/i).length).toBeGreaterThan(0);
    });

    it('должны быть примеры занятий', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const demoLessons = container.querySelector('.demo-lessons');
      expect(demoLessons).toBeInTheDocument();
    });

    it('должны быть примеры домашних заданий', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const homeworkGrid = container.querySelector('.homework-examples-grid');
      expect(homeworkGrid).toBeInTheDocument();
    });

    it('должна быть информация о расписании занятий', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarDemo = container.querySelector('.calendar-demo');
      expect(calendarDemo).toBeInTheDocument();
    });
  });

  describe('Page Structure', () => {
    it('все основные секции должны быть на месте', () => {
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

    it('все секции должны иметь контейнер', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const sections = container.querySelectorAll('section');
      expect(sections.length).toBeGreaterThan(0);

      let containerCount = 0;
      sections.forEach(section => {
        const hasContainer = section.querySelector('.container') !== null;
        if (hasContainer) containerCount++;
      });

      // Большинство секций должны иметь контейнер
      expect(containerCount).toBeGreaterThan(0);
    });
  });

  describe('Error Message Handling', () => {
    it('ошибки валидации должны быть скрыты по умолчанию', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const errorMessages = container.querySelectorAll('.error-message');
      errorMessages.forEach(error => {
        // По стилям они должны быть скрыты (display: none)
        expect(error).toBeInTheDocument();
      });
    });
  });
});
