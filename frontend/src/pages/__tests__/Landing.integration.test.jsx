import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

describe('Landing Page - Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('Component Rendering', () => {
    it('should render Landing page without crashing', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );
      expect(container.querySelector('.landing-page')).toBeInTheDocument();
    });

    it('should render all major sections', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.landing-header')).toBeInTheDocument();
      expect(container.querySelector('.hero-section')).toBeInTheDocument();
      expect(container.querySelector('.features-section')).toBeInTheDocument();
      expect(container.querySelector('.roles-section')).toBeInTheDocument();
      expect(container.querySelector('.cta-section')).toBeInTheDocument();
      expect(container.querySelector('.footer-section')).toBeInTheDocument();
    });

    it('should render HeroSection with THE BOT logo and login link', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const logoElements = container.querySelectorAll('.logo, .logo-text');
      let foundLogo = false;
      logoElements.forEach(el => {
        if (el.textContent.includes('THE BOT')) foundLogo = true;
      });
      expect(foundLogo).toBeTruthy();
      expect(screen.getByText('Войти')).toBeInTheDocument();
    });

    it('should render FeaturesSection with all features', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('Возможности платформы')).toBeInTheDocument();
      expect(screen.getByText('Трекинг прогресса')).toBeInTheDocument();
      expect(screen.getByText('Прямая связь')).toBeInTheDocument();
      expect(screen.getByText('Персональные отчёты')).toBeInTheDocument();
      expect(screen.getByText('Материалы и задания')).toBeInTheDocument();
      expect(screen.getByText('Многоролевая система')).toBeInTheDocument();
      expect(screen.getByText('Безопасность данных')).toBeInTheDocument();
    });

    it('should render RolesSection with all roles', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('Для всех участников образовательного процесса')).toBeInTheDocument();
      expect(screen.getByText('Ученики')).toBeInTheDocument();
      expect(screen.getByText('Родители')).toBeInTheDocument();
      expect(screen.getByText('Преподаватели')).toBeInTheDocument();
      expect(screen.getByText('Тьюторы')).toBeInTheDocument();
    });

    it('should render CallToActionSection with button', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('Подать заявку на обучение')).toBeInTheDocument();
      const buttons = screen.getAllByText('Подать заявку');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should render FooterSection with THE BOT branding', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const footer = container.querySelector('.footer-section');
      expect(footer).toBeInTheDocument();
      expect(footer.textContent).toContain('THE BOT');
      expect(footer.textContent).toContain('2024-2025');
    });

    it('should render ToastContainer', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const toastContainer = container.querySelector('[aria-label="Notifications Alt+T"]') ||
                            container.querySelector('.Toastify');
      expect(toastContainer).toBeTruthy();
    });
  });

  describe('Hero Section', () => {
    it('should have hero title', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('ТВОЙ КРАТЧАЙШИЙ ПУТЬ К ДИПЛОМУ')).toBeInTheDocument();
    });

    it('should have hero waves image', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const waveImage = container.querySelector('.hero-waves');
      expect(waveImage).toBeInTheDocument();
      expect(waveImage.src).toContain('waves_big.svg');
    });

    it('should have header with proper classes', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const header = container.querySelector('.landing-header');
      expect(header).toBeInTheDocument();
      expect(header.querySelector('.logo')).toBeInTheDocument();
    });
  });

  describe('Features Section', () => {
    it('should render 6 feature cards', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const featureCards = container.querySelectorAll('.feature-card');
      expect(featureCards.length).toBe(6);
    });

    it('each feature card should have icon and description', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const featureCards = container.querySelectorAll('.feature-card');
      featureCards.forEach(card => {
        expect(card.querySelector('.feature-icon')).toBeInTheDocument();
        expect(card.querySelector('h3')).toBeInTheDocument();
        expect(card.querySelector('p')).toBeInTheDocument();
      });
    });

    it('should have features grid layout', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.features-grid')).toBeInTheDocument();
    });
  });

  describe('Roles Section', () => {
    it('should render 4 role cards', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const roleCards = container.querySelectorAll('.role-card');
      expect(roleCards.length).toBe(4);
    });

    it('each role card should have icon, title, and features list', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const roleCards = container.querySelectorAll('.role-card');
      roleCards.forEach(card => {
        expect(card.querySelector('.role-icon')).toBeInTheDocument();
        expect(card.querySelector('h3')).toBeInTheDocument();
        expect(card.querySelector('ul')).toBeInTheDocument();
        expect(card.querySelectorAll('li').length).toBeGreaterThan(0);
      });
    });

    it('should have roles grid layout', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.roles-grid')).toBeInTheDocument();
    });
  });

  describe('Call To Action Section', () => {
    it('should have proper structure', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.cta-section')).toBeInTheDocument();
      expect(container.querySelector('.cta-content')).toBeInTheDocument();
    });

    it('should have CTA content text', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(
        screen.getByText(/Заполните форму заявки/i)
      ).toBeInTheDocument();
    });
  });

  describe('Footer Section', () => {
    it('should have footer content', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.footer-content')).toBeInTheDocument();
    });

    it('should have footer logo', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.footer-logo')).toBeInTheDocument();
    });

    it('should have copyright info', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText(/2024-2025 THE BOT/i)).toBeInTheDocument();
    });
  });

  describe('No Console Errors', () => {
    it('should not have console errors during render', () => {
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
