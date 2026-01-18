import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

describe('Landing Page - Branding Tests', () => {
  beforeEach(() => {
    vi?.clearAllMocks?.();
  });

  describe('THE BOT Branding', () => {
    it('should display THE BOT in header logo', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const logoElements = container.querySelectorAll('.logo, .logo-text');
      let foundCount = 0;
      logoElements.forEach(el => {
        if (el.textContent.includes('THE BOT')) foundCount++;
      });
      expect(foundCount).toBeGreaterThanOrEqual(2);
    });

    it('should have THE BOT in footer', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const footer = container.querySelector('.footer-section');
      expect(footer.textContent).toContain('THE BOT');
    });

    it('should have THE BOT in copyright text', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText(/2024-2025 THE BOT/)).toBeInTheDocument();
    });

    it('footer should have logo icon', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const footerLogo = container.querySelector('.footer-logo');
      expect(footerLogo).toBeInTheDocument();
      expect(footerLogo.textContent).toContain('📚');
    });

    it('footer logo should have text class', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const logoText = container.querySelector('.logo-text');
      expect(logoText).toBeInTheDocument();
      expect(logoText.textContent).toBe('THE BOT');
    });

    it('header logo should be a div element', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const logo = container.querySelector('.landing-header .logo');
      expect(logo).toBeInTheDocument();
      expect(logo.tagName).toBe('DIV');
      expect(logo.textContent).toBe('THE BOT');
    });

    it('logo should have navigation links', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const loginLinks = container.querySelectorAll('a[href="/login"]');
      expect(loginLinks.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('No DIPLOMA References (except header)', () => {
    it('landing page should not contain DIPLOMA text', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const landingPage = container.querySelector('.landing-page');
      const text = landingPage.textContent;

      expect(text).not.toContain('DIPLOMA');
    });

    it('should not reference DIPLOMA.svg in landing components', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const images = container.querySelectorAll('img');
      images.forEach(img => {
        if (img.classList.contains('hero-waves') || img.classList.contains('waves_big')) {
          expect(img.src).not.toContain('DIPLOMA');
        }
      });
    });
  });

  describe('Title and Hero Text', () => {
    it('should have hero title with diploma reference in content', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText(/КРАТЧАЙШИЙ ПУТЬ К ДИПЛОМУ/i)).toBeInTheDocument();
    });

    it('hero title should be h1', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const heroTitle = container.querySelector('.hero-title');
      expect(heroTitle?.tagName).toBe('H1');
    });

    it('should have section titles', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('Возможности платформы')).toBeInTheDocument();
      expect(screen.getByText('Для всех участников образовательного процесса')).toBeInTheDocument();
      expect(screen.getByText('Подать заявку на обучение')).toBeInTheDocument();
    });
  });

  describe('Navigation Links', () => {
    it('should have login link in header', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const headerContainer = container.querySelector('.landing-header');
      const loginLink = headerContainer.querySelector('a[href="/login"]');
      expect(loginLink).toBeInTheDocument();
      expect(loginLink.textContent).toBe('Войти');
    });

    it('should have login link in CTA section', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const ctaSection = container.querySelector('.cta-section');
      const ctaLink = ctaSection?.querySelector('a[href="/login"]');
      expect(ctaLink).toBeInTheDocument();
    });

    it('CTA button should have proper classes', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const ctaSection = container.querySelector('.cta-section');
      const ctaLink = ctaSection?.querySelector('a[href="/login"]');

      expect(ctaLink?.classList.contains('btn')).toBeTruthy();
      expect(ctaLink?.classList.contains('btn-primary')).toBeTruthy();
    });
  });

  describe('Content Structure', () => {
    it('should have proper container layout', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const containers = container.querySelectorAll('.container');
      expect(containers.length).toBeGreaterThan(0);
    });

    it('all major sections should be inside landing-page div', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const landingPage = container.querySelector('.landing-page');
      expect(landingPage.querySelector('.hero-section')).toBeInTheDocument();
      expect(landingPage.querySelector('.features-section')).toBeInTheDocument();
      expect(landingPage.querySelector('.roles-section')).toBeInTheDocument();
      expect(landingPage.querySelector('.cta-section')).toBeInTheDocument();
      expect(landingPage.querySelector('.footer-section')).toBeInTheDocument();
    });
  });
});
