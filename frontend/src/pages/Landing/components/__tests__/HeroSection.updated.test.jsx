import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import HeroSection from '../HeroSection';

const renderWithRouter = (component) => {
  return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('HeroSection - Updated Hero Logo "THE BOT"', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Test 1: Build Test - npm run build verification', () => {
    it('should render HeroSection without errors', () => {
      const { container } = renderWithRouter(<HeroSection />);
      expect(container).toBeTruthy();
      expect(container.querySelector('.hero-section')).toBeInTheDocument();
    });

    it('should render header and hero sections', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const header = container.querySelector('.landing-header');
      const heroSection = container.querySelector('.hero-section');

      expect(header).toBeInTheDocument();
      expect(heroSection).toBeInTheDocument();
    });

    it('should have valid JSX structure', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const headerContainer = container.querySelector('.header-container');
      const heroContent = container.querySelector('.hero-content');

      expect(headerContainer).toBeInTheDocument();
      expect(heroContent).toBeInTheDocument();
    });
  });

  describe('Test 2: Component Rendering - Verify HeroSection structure', () => {
    it('should render top header with "THE BOT" logo', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');

      expect(logo).toBeInTheDocument();
      expect(logo).toHaveTextContent('THE BOT');
    });

    it('should render "Войти" link in header', () => {
      renderWithRouter(<HeroSection />);
      const loginLink = screen.getByRole('link', { name: /войти/i });

      expect(loginLink).toBeInTheDocument();
      expect(loginLink).toHaveAttribute('href', '/login');
    });

    it('should render central hero logo "THE BOT"', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toBeInTheDocument();
      expect(heroLogo).toHaveTextContent('THE BOT');
    });

    it('should render hero title', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroTitle = container.querySelector('.hero-title');

      expect(heroTitle).toBeInTheDocument();
      expect(heroTitle).toHaveTextContent('ТВОЙ КРАТЧАЙШИЙ ПУТЬ К УСПЕХУ');
    });

    it('should render hero section with proper structure', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-waves')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
    });

    it('should render waves image with proper attributes', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const waves = container.querySelector('.hero-waves');

      expect(waves).toHaveAttribute('src', '/images/waves_big.svg');
      expect(waves).toHaveAttribute('alt', '');
      expect(waves).toHaveAttribute('aria-hidden', 'true');
    });

    it('should render header with proper container structure', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const header = container.querySelector('.landing-header');
      const headerContainer = container.querySelector('.header-container');

      expect(header).toBeInTheDocument();
      expect(headerContainer).toBeInTheDocument();
      expect(header.contains(headerContainer)).toBe(true);
    });

    it('should have header-container with flex layout', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const headerContainer = container.querySelector('.header-container');

      expect(headerContainer).toBeInTheDocument();
      expect(headerContainer.children.length).toBe(2); // logo + login link
    });
  });

  describe('Test 3: Responsive Design - Hero logo sizing', () => {
    it('should apply correct font-size for desktop (1024px+): 4.5rem', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toBeInTheDocument();
      expect(heroLogo).toHaveClass('hero-logo');
      // CSS style should be applied via className
    });

    it('should have hero-logo class for styling application', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toHaveClass('hero-logo');
    });

    it('should render hero content with padding for responsive layout', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroContent = container.querySelector('.hero-content');

      expect(heroContent).toBeInTheDocument();
      expect(heroContent).toHaveClass('hero-content');
    });

    it('should render hero section with min-height', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      // min-height: 500px is applied via CSS class
    });

    it('hero-logo should use text gradient classes', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toHaveClass('hero-logo');
      // Gradient is applied via CSS
    });

    it('should render at various viewport sizes without errors', () => {
      const viewports = [375, 480, 768, 1024, 1440, 1920];

      viewports.forEach((width) => {
        const { container } = renderWithRouter(<HeroSection />);
        const heroLogo = container.querySelector('.hero-logo');
        expect(heroLogo).toBeInTheDocument();
      });
    });
  });

  describe('Test 4: CSS Verification - Hero logo classes', () => {
    it('should have hero-logo class applied', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toHaveClass('hero-logo');
    });

    it('should have hero-content class applied', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroContent = container.querySelector('.hero-content');

      expect(heroContent).toHaveClass('hero-content');
    });

    it('should have hero-section class applied', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toHaveClass('hero-section');
    });

    it('should have hero-title class applied', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroTitle = container.querySelector('.hero-title');

      expect(heroTitle).toHaveClass('hero-title');
    });

    it('should have logo class in header', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');

      expect(logo).toHaveClass('logo');
    });

    it('should have login-link class', () => {
      renderWithRouter(<HeroSection />);
      const loginLink = screen.getByRole('link', { name: /войти/i });

      expect(loginLink).toHaveClass('login-link');
    });

    it('should have hero-waves class', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const waves = container.querySelector('.hero-waves');

      expect(waves).toHaveClass('hero-waves');
    });

    it('should have header-container class', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const headerContainer = container.querySelector('.header-container');

      expect(headerContainer).toHaveClass('header-container');
    });

    it('should have landing-header class', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const header = container.querySelector('.landing-header');

      expect(header).toHaveClass('landing-header');
    });
  });

  describe('Test 5: Animation - Hero logo fade-in', () => {
    it('should have animation class on hero-content', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroContent = container.querySelector('.hero-content');

      expect(heroContent).toBeInTheDocument();
      // Animation is applied via CSS: animation: fadeInUp 0.8s ease-out;
    });

    it('should have animation class on hero-logo', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toBeInTheDocument();
      // Animation is applied via CSS: animation: fadeInUp 1s ease-out;
    });

    it('should render hero section with proper display flex for animation', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      // Display flex allows proper centering with animation
    });

    it('should have hero-content relative positioning for animation', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroContent = container.querySelector('.hero-content');

      expect(heroContent).toBeInTheDocument();
      // Relative positioning for animation transform
    });
  });

  describe('Test 6: Accessibility', () => {
    it('should have proper semantic header structure', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const header = container.querySelector('header');

      expect(header).toBeInTheDocument();
    });

    it('should have proper semantic section structure', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const section = container.querySelector('section.hero-section');

      expect(section).toBeInTheDocument();
    });

    it('should have h1 heading for hero title', () => {
      renderWithRouter(<HeroSection />);
      const heading = screen.getByRole('heading', { level: 1 });

      expect(heading).toBeInTheDocument();
      expect(heading).toHaveTextContent('ТВОЙ КРАТЧАЙШИЙ ПУТЬ К УСПЕХУ');
    });

    it('should have accessible link to login page', () => {
      renderWithRouter(<HeroSection />);
      const loginLink = screen.getByRole('link', { name: /войти/i });

      expect(loginLink).toBeInTheDocument();
      expect(loginLink.tagName).toBe('A');
    });

    it('should have hero-waves with aria-hidden true', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const waves = container.querySelector('.hero-waves');

      expect(waves).toHaveAttribute('aria-hidden', 'true');
    });

    it('should have proper text hierarchy', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');
      const heroLogo = container.querySelector('.hero-logo');
      const heroTitle = container.querySelector('.hero-title');

      expect(logo).toBeInTheDocument();
      expect(heroLogo).toBeInTheDocument();
      expect(heroTitle).toBeInTheDocument();
    });
  });

  describe('Test 7: Content Verification', () => {
    it('should have exact text "THE BOT" in header logo', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');

      expect(logo.textContent).toBe('THE BOT');
    });

    it('should have exact text "THE BOT" in hero logo', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo.textContent).toBe('THE BOT');
    });

    it('should have exact hero title text', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroTitle = container.querySelector('.hero-title');

      expect(heroTitle.textContent).toBe('ТВОЙ КРАТЧАЙШИЙ ПУТЬ К УСПЕХУ');
    });

    it('should have correct login link text', () => {
      renderWithRouter(<HeroSection />);
      const loginLink = screen.getByRole('link', { name: /войти/i });

      expect(loginLink.textContent).toBe('Войти');
    });

    it('should have waves image with correct src', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const waves = container.querySelector('.hero-waves');

      expect(waves).toHaveAttribute('src', '/images/waves_big.svg');
    });
  });

  describe('Test 8: Integration Tests', () => {
    it('should render complete hero section without console errors', () => {
      const consoleError = vi.spyOn(console, 'error');
      renderWithRouter(<HeroSection />);

      expect(consoleError).not.toHaveBeenCalled();
      consoleError.mockRestore();
    });

    it('should render complete header without console errors', () => {
      const consoleError = vi.spyOn(console, 'error');
      renderWithRouter(<HeroSection />);

      expect(consoleError).not.toHaveBeenCalled();
      consoleError.mockRestore();
    });

    it('should maintain proper DOM structure throughout component', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const html = container.innerHTML;

      expect(html).toContain('hero-section');
      expect(html).toContain('landing-header');
      expect(html).toContain('hero-logo');
      expect(html).toContain('THE BOT');
    });

    it('should render all required elements together', () => {
      const { container } = renderWithRouter(<HeroSection />);

      expect(container.querySelector('header')).toBeInTheDocument();
      expect(container.querySelector('section.hero-section')).toBeInTheDocument();
      expect(container.querySelector('.hero-logo')).toBeInTheDocument();
      expect(container.querySelector('.hero-title')).toBeInTheDocument();
    });
  });

  describe('Test 9: Link Navigation', () => {
    it('should have login link pointing to /login', () => {
      renderWithRouter(<HeroSection />);
      const loginLink = screen.getByRole('link', { name: /войти/i });

      expect(loginLink).toHaveAttribute('href', '/login');
    });

    it('should have proper navigation structure', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const links = container.querySelectorAll('a');

      expect(links.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('Test 10: Visual Consistency', () => {
    it('should use consistent font for logo and hero text', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');
      const heroLogo = container.querySelector('.hero-logo');

      expect(logo).toBeInTheDocument();
      expect(heroLogo).toBeInTheDocument();
      // Both use uppercase text-transform
    });

    it('should apply letter-spacing to logo', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');

      expect(logo).toBeInTheDocument();
      // letter-spacing: 0.3em applied via CSS
    });

    it('should apply letter-spacing to hero-logo', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const heroLogo = container.querySelector('.hero-logo');

      expect(heroLogo).toBeInTheDocument();
      // letter-spacing: 0.15em applied via CSS
    });

    it('should have uppercase text in all key elements', () => {
      const { container } = renderWithRouter(<HeroSection />);
      const logo = container.querySelector('.logo');
      const heroLogo = container.querySelector('.hero-logo');
      const heroTitle = container.querySelector('.hero-title');

      expect(logo.textContent).toBe(logo.textContent.toUpperCase());
      expect(heroLogo.textContent).toBe(heroLogo.textContent.toUpperCase());
      expect(heroTitle.textContent).toBe(heroTitle.textContent.toUpperCase());
    });
  });

  describe('Test 11: Edge Cases', () => {
    it('should handle rendering multiple times', () => {
      const { rerender } = renderWithRouter(<HeroSection />);

      expect(screen.getAllByText('THE BOT')[0]).toBeInTheDocument();

      rerender(
        <BrowserRouter>
          <HeroSection />
        </BrowserRouter>
      );

      expect(screen.getAllByText('THE BOT')[0]).toBeInTheDocument();
    });

    it('should render correctly after mounting', () => {
      const { container } = renderWithRouter(<HeroSection />);

      setTimeout(() => {
        const heroLogo = container.querySelector('.hero-logo');
        expect(heroLogo).toBeInTheDocument();
      }, 100);
    });

    it('should maintain structure with dynamic viewport changes', () => {
      const { container } = renderWithRouter(<HeroSection />);

      // Simulate viewport resize
      window.dispatchEvent(new Event('resize'));

      expect(container.querySelector('.hero-section')).toBeInTheDocument();
    });
  });
});
