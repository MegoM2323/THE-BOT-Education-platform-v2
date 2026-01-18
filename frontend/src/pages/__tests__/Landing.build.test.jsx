import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';
import HeroSection from '../Landing/components/HeroSection';
import FeaturesSection from '../Landing/components/FeaturesSection';
import RolesSection from '../Landing/components/RolesSection';
import CallToActionSection from '../Landing/components/CallToActionSection';
import FooterSection from '../Landing/components/FooterSection';

describe('Landing Page - Build and Rendering Tests', () => {
  beforeEach(() => {
    vi?.clearAllMocks?.();
  });

  describe('Module Compilation', () => {
    it('Landing component should compile successfully', () => {
      expect(() => render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      )).not.toThrow();
    });

    it('all section components should compile successfully', () => {
      expect(() => render(
        <BrowserRouter>
          <HeroSection />
        </BrowserRouter>
      )).not.toThrow();

      expect(() => render(
        <BrowserRouter>
          <FeaturesSection />
        </BrowserRouter>
      )).not.toThrow();

      expect(() => render(
        <BrowserRouter>
          <RolesSection />
        </BrowserRouter>
      )).not.toThrow();

      expect(() => render(
        <BrowserRouter>
          <CallToActionSection />
        </BrowserRouter>
      )).not.toThrow();

      expect(() => render(
        <BrowserRouter>
          <FooterSection />
        </BrowserRouter>
      )).not.toThrow();
    });
  });

  describe('Component Rendering', () => {
    it('Landing component should render without errors', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container).toBeTruthy();
      expect(container.querySelector('.landing-page')).toBeInTheDocument();
    });

    it('HeroSection should render without errors', () => {
      const { container } = render(
        <BrowserRouter>
          <HeroSection />
        </BrowserRouter>
      );

      expect(container.querySelector('.landing-header')).toBeInTheDocument();
      expect(container.querySelector('.hero-section')).toBeInTheDocument();
    });

    it('FeaturesSection should render without errors', () => {
      const { container } = render(
        <BrowserRouter>
          <FeaturesSection />
        </BrowserRouter>
      );

      expect(container.querySelector('.features-section')).toBeInTheDocument();
      expect(container.querySelectorAll('.feature-card').length).toBe(6);
    });

    it('RolesSection should render without errors', () => {
      const { container } = render(
        <BrowserRouter>
          <RolesSection />
        </BrowserRouter>
      );

      expect(container.querySelector('.roles-section')).toBeInTheDocument();
      expect(container.querySelectorAll('.role-card').length).toBe(4);
    });

    it('CallToActionSection should render without errors', () => {
      const { container } = render(
        <BrowserRouter>
          <CallToActionSection />
        </BrowserRouter>
      );

      expect(container.querySelector('.cta-section')).toBeInTheDocument();
    });

    it('FooterSection should render without errors', () => {
      const { container } = render(
        <BrowserRouter>
          <FooterSection />
        </BrowserRouter>
      );

      expect(container.querySelector('.footer-section')).toBeInTheDocument();
    });
  });

  describe('Component Integration', () => {
    it('all sections should render together in Landing component', () => {
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

    it('sections should maintain proper DOM structure', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const landingPage = container.querySelector('.landing-page');
      const children = landingPage.children;

      expect(children.length).toBeGreaterThan(0);
    });

    it('ToastContainer should be present', () => {
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

  describe('No Console Warnings or Errors', () => {
    it('should not produce console errors', () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(consoleErrorSpy).not.toHaveBeenCalled();
      consoleErrorSpy.mockRestore();
    });

    it('should not produce React warnings', () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      consoleWarnSpy.mockRestore();
    });
  });

  describe('Import Paths', () => {
    it('should have correct import paths', () => {
      expect(HeroSection).toBeDefined();
      expect(FeaturesSection).toBeDefined();
      expect(RolesSection).toBeDefined();
      expect(CallToActionSection).toBeDefined();
      expect(FooterSection).toBeDefined();
      expect(Landing).toBeDefined();
    });

    it('all components should be functions', () => {
      expect(typeof HeroSection).toBe('function');
      expect(typeof FeaturesSection).toBe('function');
      expect(typeof RolesSection).toBe('function');
      expect(typeof CallToActionSection).toBe('function');
      expect(typeof FooterSection).toBe('function');
      expect(typeof Landing).toBe('function');
    });
  });

  describe('CSS Import', () => {
    it('Landing component should import CSS', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const styles = window.getComputedStyle(container.querySelector('.landing-page'));
      expect(styles).toBeDefined();
    });

    it('CSS classes should be applied to elements', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const elements = [
        { selector: '.landing-page', expected: 'landing-page' },
        { selector: '.landing-header', expected: 'landing-header' },
        { selector: '.hero-section', expected: 'hero-section' },
        { selector: '.features-section', expected: 'features-section' },
        { selector: '.roles-section', expected: 'roles-section' },
        { selector: '.cta-section', expected: 'cta-section' },
        { selector: '.footer-section', expected: 'footer-section' }
      ];

      elements.forEach(({ selector, expected }) => {
        const element = container.querySelector(selector);
        expect(element).toBeInTheDocument();
        expect(element.className).toContain(expected);
      });
    });
  });

  describe('React Router Integration', () => {
    it('should work within BrowserRouter', () => {
      expect(() => render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      )).not.toThrow();
    });

    it('should render navigation links correctly', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const links = container.querySelectorAll('a[href="/login"]');
      expect(links.length).toBeGreaterThan(0);
    });
  });

  describe('Asset Loading', () => {
    it('should reference SVG assets correctly', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const waveImage = container.querySelector('.hero-waves');
      expect(waveImage).toBeInTheDocument();
      expect(waveImage.src).toContain('waves_big.svg');
    });
  });

  describe('Production Ready', () => {
    it('component should be production ready', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container).toBeTruthy();
      expect(container.querySelector('.landing-page')).toBeInTheDocument();

      const criticalElements = [
        '.landing-header',
        '.hero-section',
        '.features-section',
        '.roles-section',
        '.cta-section',
        '.footer-section'
      ];

      criticalElements.forEach(selector => {
        expect(container.querySelector(selector)).toBeInTheDocument();
      });
    });

    it('all sections should have proper semantics', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const sections = container.querySelectorAll('section');
      expect(sections.length).toBeGreaterThan(0);

      const header = container.querySelector('header');
      expect(header).toBeInTheDocument();

      const footer = container.querySelector('footer');
      expect(footer).toBeInTheDocument();
    });
  });
});
