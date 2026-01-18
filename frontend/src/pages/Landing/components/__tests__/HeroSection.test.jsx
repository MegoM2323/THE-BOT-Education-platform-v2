import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import HeroSection from '../HeroSection';

describe('HeroSection', () => {
  beforeEach(() => {
    // Мокаем scrollIntoView
    window.HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  describe('Rendering', () => {
    it('должен отображать заголовок', () => {
      render(<HeroSection />);

      const title = screen.getByRole('heading', { level: 1 });
      expect(title).toHaveTextContent(/кратчайший путь к твоему диплому/i);
    });

    it('должен отображать подзаголовок с описанием учителя', () => {
      render(<HeroSection />);

      expect(
        screen.getByRole('heading', { level: 2, name: /кто я\?/i })
      ).toBeInTheDocument();
      expect(
        screen.getByText(/мирослав адаменко/i)
      ).toBeInTheDocument();
    });

    it('должен отображать фото преподавателя', () => {
      render(<HeroSection />);

      const image = screen.getByAltText(/преподаватель мирослав адаменко/i);
      expect(image).toBeInTheDocument();
    });

    it('должен иметь правильную структуру секции', () => {
      const { container } = render(<HeroSection />);

      const section = container.querySelector('.hero-section');
      expect(section).toBeInTheDocument();
      expect(section).toHaveAttribute('id', 'hero');
    });

    it('должен иметь AOS атрибуты для анимации', () => {
      const { container } = render(<HeroSection />);

      const title = container.querySelector('[data-aos="fade-up"]');
      expect(title).toBeInTheDocument();
    });
  });

  describe('Functionality', () => {
    it('должен корректно отображать макет с двумя колонками', () => {
      const { container } = render(<HeroSection />);

      const layout = container.querySelector('.hero-layout');
      expect(layout).toBeInTheDocument();

      const leftColumn = container.querySelector('.hero-left');
      const rightColumn = container.querySelector('.hero-right');
      expect(leftColumn).toBeInTheDocument();
      expect(rightColumn).toBeInTheDocument();
    });

    it('должен отображать фоновые декоративные элементы', () => {
      const { container } = render(<HeroSection />);

      const bgShapes = container.querySelector('.hero-bg-shapes');
      expect(bgShapes).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('должен иметь семантически правильную структуру', () => {
      const { container } = render(<HeroSection />);

      // Проверка, что это секция
      expect(container.querySelector('section')).toBeInTheDocument();

      // Проверка наличия заголовка h1
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();

      // Проверка наличия заголовка h2 "Кто я?"
      expect(screen.getByRole('heading', { level: 2 })).toBeInTheDocument();
    });

    it('должен содержать изображение с правильным alt текстом', () => {
      render(<HeroSection />);

      const image = screen.getByAltText(/преподаватель мирослав адаменко/i);
      expect(image).toHaveAttribute('src', '/images/landing/hero.png');
    });
  });

  describe('Content', () => {
    it('должен содержать правильный текст заголовка', () => {
      render(<HeroSection />);

      expect(screen.getByText('КРАТЧАЙШИЙ ПУТЬ К ТВОЕМУ ДИПЛОМУ')).toBeInTheDocument();
    });

    it('должен содержать правильный текст "Кто я?"', () => {
      render(<HeroSection />);

      expect(
        screen.getByRole('heading', { level: 2, name: /кто я\?/i })
      ).toBeInTheDocument();
    });

    it('должен содержать информацию о Мирославе Адаменко', () => {
      render(<HeroSection />);

      expect(screen.getByText(/мирослав адаменко/i)).toBeInTheDocument();
      expect(screen.getByText(/увлеченный преподаватель/i)).toBeInTheDocument();
    });
  });
});
