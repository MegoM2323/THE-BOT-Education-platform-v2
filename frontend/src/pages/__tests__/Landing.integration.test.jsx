import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup, within } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

describe('Landing Page Integration Tests (T010)', () => {
  beforeEach(() => {
    // Silence console errors during tests
    global.consoleError = console.error;
    console.error = (...args) => {
      if (
        typeof args[0] === 'string' &&
        (args[0].includes('Warning: ReactDOM.render') ||
         args[0].includes('Not implemented: HTMLFormElement.prototype.submit'))
      ) {
        return;
      }
      global.consoleError(...args);
    };
  });

  afterEach(() => {
    console.error = global.consoleError;
    cleanup();
  });

  describe('Test 1: Page Rendering', () => {
    it('should render Landing page without errors', () => {
      expect(() => {
        render(
          <BrowserRouter>
            <Landing />
          </BrowserRouter>
        );
      }).not.toThrow();
    });

    it('should render main container with min-h-screen class', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );
      const mainDiv = container.firstChild;
      expect(mainDiv).toHaveClass('min-h-screen');
    });
  });

  describe('Test 2: Sections Presence', () => {
    it('should render Hero section with main heading', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const heading = screen.getByText(/Образовательная платформа нового поколения/i);
      expect(heading).toBeInTheDocument();
      expect(heading.tagName).toBe('H1');
    });

    it('should render Features section with heading', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const featuresHeading = screen.getByText('Возможности платформы');
      expect(featuresHeading).toBeInTheDocument();
      expect(featuresHeading.tagName).toBe('H2');
    });

    it('should render all 6 feature cards', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Features section has id="features"
      const featuresSection = container.querySelector('#features');
      expect(featuresSection).toBeInTheDocument();

      // Check for feature titles
      expect(screen.getByText('Трекинг прогресса')).toBeInTheDocument();
      expect(screen.getByText('Прямая связь')).toBeInTheDocument();
      expect(screen.getByText('Персональные отчёты')).toBeInTheDocument();
      expect(screen.getByText('Материалы и задания')).toBeInTheDocument();
      expect(screen.getByText('Многоролевая система')).toBeInTheDocument();
      expect(screen.getByText('Безопасность данных')).toBeInTheDocument();
    });

    it('should render Roles section with heading', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const rolesHeading = screen.getByText('Для всех участников образовательного процесса');
      expect(rolesHeading).toBeInTheDocument();
      expect(rolesHeading.tagName).toBe('H2');
    });

    it('should render all 4 role cards', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Roles section has id="roles"
      const rolesSection = container.querySelector('#roles');
      expect(rolesSection).toBeInTheDocument();

      // Check for role titles
      expect(screen.getByText('Ученики')).toBeInTheDocument();
      expect(screen.getByText('Родители')).toBeInTheDocument();
      expect(screen.getByText('Преподаватели')).toBeInTheDocument();
      expect(screen.getByText('Тьюторы')).toBeInTheDocument();
    });

    it('should render Application CTA section', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const ctaHeading = screen.getByText('Подать заявку на обучение');
      expect(ctaHeading).toBeInTheDocument();
      expect(ctaHeading.tagName).toBe('H2');
    });

    it('should render Footer section', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const footer = screen.getByText(/© 2024 THE BOT/i);
      expect(footer).toBeInTheDocument();

      const brandName = screen.getByText('THE BOT');
      expect(brandName).toBeInTheDocument();
    });
  });

  describe('Test 3: Navigation Links', () => {
    it('should have "Подать заявку" button in Hero linking to /login', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Find all links with text "Подать заявку"
      const applyButtons = screen.getAllByText('Подать заявку');
      expect(applyButtons.length).toBeGreaterThan(0);

      // First button should be in Hero section
      const heroButton = applyButtons[0];
      const link = heroButton.closest('a');
      expect(link).toHaveAttribute('href', '/login');
    });

    it('should have "Подать заявку" button in CTA section linking to /login', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const applyButtons = screen.getAllByText('Подать заявку');
      // CTA section has second "Подать заявку" button
      expect(applyButtons.length).toBeGreaterThanOrEqual(2);

      const ctaButton = applyButtons[1];
      const link = ctaButton.closest('a');
      expect(link).toHaveAttribute('href', '/login');
    });
  });

  describe('Test 4: Header Component Visibility', () => {
    it('should work with separate Header component (integration check)', () => {
      // Landing page doesn't have embedded header (removed in T004)
      // Header is rendered separately in App.jsx
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Landing should NOT have its own header
      const landingHeader = container.querySelector('.landing-header');
      expect(landingHeader).not.toBeInTheDocument();

      // Landing starts with Hero section
      const heroSection = container.querySelector('section');
      expect(heroSection).toBeInTheDocument();
    });
  });

  describe('Test 5: Responsive Layout', () => {
    it('should have responsive grid classes for features', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const featuresGrid = container.querySelector('#features .grid');
      expect(featuresGrid).toHaveClass('md:grid-cols-2');
      expect(featuresGrid).toHaveClass('lg:grid-cols-3');
    });

    it('should have responsive grid classes for roles', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const rolesGrid = container.querySelector('#roles .grid');
      expect(rolesGrid).toHaveClass('md:grid-cols-2');
    });

    it('should have responsive padding classes in Hero', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const heroSection = container.querySelector('section');
      expect(heroSection).toHaveClass('py-20');
      expect(heroSection).toHaveClass('md:py-32');
    });

    it('should have responsive flex direction for CTA buttons', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Hero buttons container
      const buttonsContainer = container.querySelector('.flex.flex-col.sm\\:flex-row');
      expect(buttonsContainer).toBeInTheDocument();
      expect(buttonsContainer).toHaveClass('sm:flex-row');
    });
  });

  describe('Test 6: Tailwind Styles Applied', () => {
    it('should have gradient text in Hero heading', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const heading = screen.getByText(/Образовательная платформа нового поколения/i);
      expect(heading).toHaveClass('gradient-text');
    });

    it('should have gradient-primary class on buttons', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const gradientButtons = container.querySelectorAll('.gradient-primary');
      expect(gradientButtons.length).toBeGreaterThan(0);
    });

    it('should have shadow-glow class on primary buttons', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const glowButtons = container.querySelectorAll('.shadow-glow');
      expect(glowButtons.length).toBeGreaterThan(0);
    });

    it('should have animate-slide-up class in Hero', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const animatedDiv = container.querySelector('.animate-slide-up');
      expect(animatedDiv).toBeInTheDocument();
    });

    it('should have bg-muted/30 background on alternating sections', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const mutedSections = container.querySelectorAll('.bg-muted\\/30');
      expect(mutedSections.length).toBeGreaterThanOrEqual(2);
    });

    it('should have hover effects on cards', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Features cards have hover:shadow-lg
      const featureCards = container.querySelectorAll('#features .hover\\:shadow-lg');
      expect(featureCards.length).toBe(6);

      // Role cards have hover:border-primary
      const roleCards = container.querySelectorAll('#roles .hover\\:border-primary');
      expect(roleCards.length).toBe(4);
    });
  });

  describe('Test 7: Icons Rendering', () => {
    it('should render lucide-react icons in features', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Icons are SVG elements, check for their presence
      const featureIcons = container.querySelectorAll('#features svg');
      expect(featureIcons.length).toBe(6);
    });

    it('should render lucide-react icons in roles', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const roleIcons = container.querySelectorAll('#roles svg');
      expect(roleIcons.length).toBe(4);
    });

    it('should render BookOpen icon in footer', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const footer = container.querySelector('footer');
      const footerIcon = footer.querySelector('svg');
      expect(footerIcon).toBeInTheDocument();
    });

    it('should have correct icon classes', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Feature icons have w-6 h-6 text-primary-foreground
      const featureIconContainers = container.querySelectorAll('#features .gradient-primary svg');
      featureIconContainers.forEach(icon => {
        expect(icon).toHaveClass('w-6');
        expect(icon).toHaveClass('h-6');
        expect(icon).toHaveClass('text-primary-foreground');
      });
    });
  });

  describe('Test 8: Content Accuracy', () => {
    it('should have correct Hero description text', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const description = screen.getByText(/Персонализированное обучение с трекингом прогресса/i);
      expect(description).toBeInTheDocument();
    });

    it('should have correct feature descriptions', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText(/Визуализация достижений и движения к образовательной цели/i)).toBeInTheDocument();
      expect(screen.getByText(/Общение с преподавателями и тьюторами напрямую через платформу/i)).toBeInTheDocument();
      expect(screen.getByText(/Подробная аналитика по каждому ученику/i)).toBeInTheDocument();
    });

    it('should have correct role features', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Students features
      expect(screen.getByText(/Доступ к образовательным материалам/i)).toBeInTheDocument();
      expect(screen.getByText(/Трекер прогресса с геймификацией/i)).toBeInTheDocument();

      // Teachers features
      expect(screen.getByText(/Создание и публикация материалов/i)).toBeInTheDocument();
      expect(screen.getByText(/Проверка домашних заданий/i)).toBeInTheDocument();
    });

    it('should have correct CTA description', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const ctaText = screen.getByText(/Заполните форму заявки, и наши специалисты свяжутся с вами/i);
      expect(ctaText).toBeInTheDocument();
    });
  });

  describe('Test 9: UI Components Integration', () => {
    it('should use Button component from shadcn/ui', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Buttons should have shadcn/ui classes
      const buttons = container.querySelectorAll('button');
      expect(buttons.length).toBeGreaterThan(0);

      // Check for button variants
      buttons.forEach(button => {
        // Should have inline-flex class (shadcn/ui Button default)
        expect(
          button.className.includes('inline-flex') ||
          button.className.includes('gradient-primary')
        ).toBeTruthy();
      });
    });

    it('should use Card component from shadcn/ui', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Cards should have shadcn/ui classes (rounded-xl, border, bg-card)
      const cards = container.querySelectorAll('.rounded-xl');
      expect(cards.length).toBeGreaterThan(0);
    });
  });

  describe('Test 10: Accessibility', () => {
    it('should have semantic HTML structure', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Should have sections
      const sections = container.querySelectorAll('section');
      expect(sections.length).toBeGreaterThanOrEqual(4);

      // Should have footer
      const footer = container.querySelector('footer');
      expect(footer).toBeInTheDocument();

      // Should have proper heading hierarchy
      const h1 = container.querySelector('h1');
      const h2 = container.querySelectorAll('h2');
      const h3 = container.querySelectorAll('h3');

      expect(h1).toBeInTheDocument();
      expect(h2.length).toBeGreaterThanOrEqual(3);
      expect(h3.length).toBeGreaterThanOrEqual(10);
    });

    it('should have proper link elements', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const links = container.querySelectorAll('a');
      expect(links.length).toBeGreaterThanOrEqual(2);

      links.forEach(link => {
        expect(link).toHaveAttribute('href');
      });
    });
  });

  describe('Test 11: Section IDs for Navigation', () => {
    it('should have id="features" on Features section', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const featuresSection = container.querySelector('#features');
      expect(featuresSection).toBeInTheDocument();
      expect(featuresSection.tagName).toBe('SECTION');
    });

    it('should have id="roles" on Roles section', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const rolesSection = container.querySelector('#roles');
      expect(rolesSection).toBeInTheDocument();
      expect(rolesSection.tagName).toBe('SECTION');
    });

    it('should have id="apply" on Application CTA section', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const applySection = container.querySelector('#apply');
      expect(applySection).toBeInTheDocument();
      expect(applySection.tagName).toBe('SECTION');
    });
  });

  describe('Test 12: No Console Errors', () => {
    it('should not log errors during render', () => {
      const errors = [];
      const originalError = console.error;
      console.error = (...args) => {
        errors.push(args);
      };

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      console.error = originalError;

      // Filter out known warnings
      const actualErrors = errors.filter(err => {
        const msg = String(err[0]);
        return !msg.includes('Warning: ReactDOM.render') &&
               !msg.includes('Not implemented: HTMLFormElement.prototype.submit');
      });

      expect(actualErrors).toHaveLength(0);
    });
  });
});
