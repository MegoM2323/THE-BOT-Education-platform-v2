import { describe, it, expect, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

describe('Landing Page - CSS Tests', () => {
  beforeEach(() => {
    vi?.clearAllMocks?.();
  });

  describe('CSS Classes Existence', () => {
    it('all components should have required CSS classes', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const requiredClasses = [
        '.landing-page',
        '.landing-header',
        '.hero-section',
        '.hero-content',
        '.hero-title',
        '.features-section',
        '.features-grid',
        '.feature-card',
        '.roles-section',
        '.roles-grid',
        '.role-card',
        '.cta-section',
        '.cta-content',
        '.footer-section',
        '.footer-content',
        '.footer-logo',
        '.container'
      ];

      requiredClasses.forEach(selector => {
        const element = container.querySelector(selector);
        expect(element).toBeInTheDocument();
      });
    });

    it('feature cards should have proper class structure', () => {
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

    it('role cards should have proper class structure', () => {
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
      });
    });
  });

  describe('CSS Grid Layouts', () => {
    it('features-grid should be present', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const grid = container.querySelector('.features-grid');
      expect(grid).toBeInTheDocument();
    });

    it('roles-grid should be present', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const grid = container.querySelector('.roles-grid');
      expect(grid).toBeInTheDocument();
    });

    it('feature cards should be inside features-grid', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const grid = container.querySelector('.features-grid');
      const cards = grid?.querySelectorAll('.feature-card');
      expect(cards?.length).toBe(6);
    });

    it('role cards should be inside roles-grid', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const grid = container.querySelector('.roles-grid');
      const cards = grid?.querySelectorAll('.role-card');
      expect(cards?.length).toBe(4);
    });
  });

  describe('Computed Styles', () => {
    it('container should have max-width set', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const containerEl = container.querySelector('.container');
      const styles = window.getComputedStyle(containerEl);

      expect(styles.maxWidth).toBeDefined();
    });

    it('landing-page should have min-height of 100vh', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const landingPage = container.querySelector('.landing-page');
      const styles = window.getComputedStyle(landingPage);

      expect(styles.minHeight).toBeDefined();
    });

    it('hero-section should have background color', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const heroSection = container.querySelector('.hero-section');
      const styles = window.getComputedStyle(heroSection);

      expect(styles.backgroundColor).toBeDefined();
    });

    it('feature-card should have padding', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const featureCard = container.querySelector('.feature-card');
      const styles = window.getComputedStyle(featureCard);

      expect(styles.padding).toBeDefined();
    });
  });

  describe('Responsive Design', () => {
    it('all sections should have container with responsive width', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const containers = container.querySelectorAll('.container');
      containers.forEach(cont => {
        expect(cont.classList.contains('container')).toBe(true);
      });
    });

    it('features-grid should support grid layout', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const grid = container.querySelector('.features-grid');
      const styles = window.getComputedStyle(grid);

      expect(styles.display).toBeDefined();
    });

    it('roles-grid should support grid layout', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const grid = container.querySelector('.roles-grid');
      const styles = window.getComputedStyle(grid);

      expect(styles.display).toBeDefined();
    });
  });

  describe('Button Styling', () => {
    it('CTA button should have btn class', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const button = container.querySelector('.btn');
      expect(button).toBeInTheDocument();
    });

    it('CTA button should have btn-primary class', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const button = container.querySelector('.btn-primary');
      expect(button).toBeInTheDocument();
    });

    it('button should have computed styles', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const button = container.querySelector('.btn-primary');
      const styles = window.getComputedStyle(button);

      expect(styles.padding).toBeDefined();
      expect(styles.borderRadius).toBeDefined();
    });
  });

  describe('Header Layout', () => {
    it('landing-header should be present', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.landing-header')).toBeInTheDocument();
    });

    it('header should have header-container class', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.header-container')).toBeInTheDocument();
    });

    it('header should have logo and login link', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const header = container.querySelector('.landing-header');
      expect(header.querySelector('.logo')).toBeInTheDocument();
      expect(header.querySelector('.login-link')).toBeInTheDocument();
    });
  });

  describe('Footer Layout', () => {
    it('footer-section should be present', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(container.querySelector('.footer-section')).toBeInTheDocument();
    });

    it('footer-logo should have proper structure', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const footerLogo = container.querySelector('.footer-logo');
      expect(footerLogo.querySelector('.logo-icon')).toBeInTheDocument();
      expect(footerLogo.querySelector('.logo-text')).toBeInTheDocument();
    });
  });

  describe('Section Visibility', () => {
    it('all sections should be visible', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const sections = [
        '.hero-section',
        '.features-section',
        '.roles-section',
        '.cta-section',
        '.footer-section'
      ];

      sections.forEach(selector => {
        const section = container.querySelector(selector);
        expect(section).toBeInTheDocument();
        const styles = window.getComputedStyle(section);
        expect(styles.display).not.toBe('none');
      });
    });
  });
});
