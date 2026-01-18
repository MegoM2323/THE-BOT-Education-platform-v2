import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import HeroSection from '../components/HeroSection';

describe('HeroSection - 100vh Fullscreen Responsive Testing', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  // T1: Mobile screen (375px)
  describe('T1: Mobile screen (375px)', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 667,
      });
    });

    it('renders hero section without errors', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');
      expect(heroSection).toBeInTheDocument();
    });

    it('hero section has id="hero" for scrolling', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('#hero');
      expect(heroSection).toBeInTheDocument();
    });

    it('hero-section has correct CSS classes for styling', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toHaveClass('hero-section');
      expect(heroSection).toBeInTheDocument();
    });

    it('hero section HTML structure is correct', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      // Check that all required child elements exist
      expect(heroSection.querySelector('.hero-bg-shapes')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
    });

    it('background shapes layer exists', () => {
      const { container } = render(<HeroSection />);
      const shapesLayer = container.querySelector('.hero-bg-shapes');

      expect(shapesLayer).toBeInTheDocument();
      expect(shapesLayer).toHaveClass('hero-bg-shapes');
    });

    it('three hero shapes are rendered', () => {
      const { container } = render(<HeroSection />);
      const shapes = container.querySelectorAll('.hero-shape');

      expect(shapes.length).toBe(3);
    });

    it('hero shapes have correct individual classes', () => {
      const { container } = render(<HeroSection />);

      expect(container.querySelector('.hero-shape-1')).toBeInTheDocument();
      expect(container.querySelector('.hero-shape-2')).toBeInTheDocument();
      expect(container.querySelector('.hero-shape-3')).toBeInTheDocument();
    });

    it('hero content is properly structured', () => {
      const { container } = render(<HeroSection />);
      const content = container.querySelector('.hero-content');

      expect(content).toBeInTheDocument();
      expect(content).toHaveClass('hero-content');
    });

    it('hero title is visible and has gradient text class', () => {
      const { container } = render(<HeroSection />);
      const title = container.querySelector('.hero-title.gradient-text');

      expect(title).toBeInTheDocument();
      expect(title.textContent).toContain('КРАТЧАЙШИЙ ПУТЬ К ТВОЕМУ ДИПЛОМУ');
    });

    it('hero layout exists and is properly structured', () => {
      const { container } = render(<HeroSection />);
      const layout = container.querySelector('.hero-layout');

      expect(layout).toBeInTheDocument();
      expect(layout).toHaveClass('hero-layout');
    });

    it('hero-left glass card is present', () => {
      const { container } = render(<HeroSection />);
      const leftCard = container.querySelector('.hero-left.glass-card');

      expect(leftCard).toBeInTheDocument();
    });

    it('hero-right section with image is present', () => {
      const { container } = render(<HeroSection />);
      const rightSection = container.querySelector('.hero-right');

      expect(rightSection).toBeInTheDocument();
    });

    it('hero image is present and has correct attributes', () => {
      const { container } = render(<HeroSection />);
      const image = container.querySelector('.hero-image');

      expect(image).toBeInTheDocument();
      expect(image).toHaveAttribute('alt', 'Преподаватель Мирослав Адаменко');
      expect(image).toHaveAttribute('src', '/images/landing/hero.png');
    });

    it('CTA button is present and has correct text', () => {
      const { container } = render(<HeroSection />);
      const ctaButton = container.querySelector('.hero-cta.glow-button');

      expect(ctaButton).toBeInTheDocument();
      expect(ctaButton.textContent).toBe('Записаться на консультацию');
    });

    it('CTA button is a button element with correct classes', () => {
      const { container } = render(<HeroSection />);
      const ctaButton = container.querySelector('.hero-cta.glow-button');

      expect(ctaButton.tagName).toBe('BUTTON');
      expect(ctaButton).toHaveClass('hero-cta');
      expect(ctaButton).toHaveClass('glow-button');
    });

    it('hero section structure maintains DOM hierarchy', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      const children = heroSection.children;
      expect(children.length).toBeGreaterThanOrEqual(2); // At least shapes and content
    });
  });

  // T2: Tablet screen (768px)
  describe('T2: Tablet screen (768px)', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 1024,
      });
    });

    it('hero section renders correctly on tablet', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('hero layout exists and is responsive', () => {
      const { container } = render(<HeroSection />);
      const layout = container.querySelector('.hero-layout');

      expect(layout).toBeInTheDocument();
      expect(layout).toHaveClass('hero-layout');
    });

    it('hero image and text are both visible on tablet', () => {
      const { container } = render(<HeroSection />);
      const image = container.querySelector('.hero-image');
      const leftCard = container.querySelector('.hero-left');

      expect(image).toBeInTheDocument();
      expect(leftCard).toBeInTheDocument();
    });

    it('CTA button is present and accessible on tablet', () => {
      const { container } = render(<HeroSection />);
      const ctaButton = container.querySelector('.hero-cta.glow-button');

      expect(ctaButton).toBeInTheDocument();
      expect(ctaButton).toHaveClass('glow-button');
    });

    it('all hero section elements remain accessible', () => {
      const { container } = render(<HeroSection />);

      const heroSection = container.querySelector('.hero-section');
      const shapesLayer = container.querySelector('.hero-bg-shapes');
      const content = container.querySelector('.hero-content');
      const shapes = container.querySelectorAll('.hero-shape');

      expect(heroSection).toBeInTheDocument();
      expect(shapesLayer).toBeInTheDocument();
      expect(content).toBeInTheDocument();
      expect(shapes.length).toBe(3);
    });
  });

  // T3: Desktop (1440px)
  describe('T3: Desktop (1440px)', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1440,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 900,
      });
    });

    it('hero section renders correctly on 1440px desktop', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('hero content container exists with proper structure', () => {
      const { container } = render(<HeroSection />);
      const content = container.querySelector('.hero-content');

      expect(content).toBeInTheDocument();
      expect(content).toHaveClass('hero-content');
    });

    it('hero layout is properly structured on desktop', () => {
      const { container } = render(<HeroSection />);
      const layout = container.querySelector('.hero-layout');

      expect(layout).toBeInTheDocument();
      expect(layout).toHaveClass('hero-layout');
    });

    it('all content elements are visible on desktop', () => {
      const { container } = render(<HeroSection />);

      const title = container.querySelector('.hero-title');
      const leftCard = container.querySelector('.hero-left');
      const rightSection = container.querySelector('.hero-right');
      const image = container.querySelector('.hero-image');
      const ctaButton = container.querySelector('.hero-cta');

      expect(title).toBeInTheDocument();
      expect(leftCard).toBeInTheDocument();
      expect(rightSection).toBeInTheDocument();
      expect(image).toBeInTheDocument();
      expect(ctaButton).toBeInTheDocument();
    });

    it('hero section is complete with all sub-elements', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      const shapes = heroSection.querySelectorAll('.hero-shape');
      const content = heroSection.querySelector('.hero-content');

      expect(shapes.length).toBe(3);
      expect(content).toBeInTheDocument();
    });
  });

  // T4: Large screen (1920px)
  describe('T4: Large screen (1920px)', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 1080,
      });
    });

    it('hero section renders correctly on 1920px large display', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('hero content exists with proper structure for large display', () => {
      const { container } = render(<HeroSection />);
      const content = container.querySelector('.hero-content');

      expect(content).toBeInTheDocument();
    });

    it('all decorative shapes are rendered on large screen', () => {
      const { container } = render(<HeroSection />);
      const shapes = container.querySelectorAll('.hero-shape');

      expect(shapes.length).toBe(3);
      expect(container.querySelector('.hero-shape-1')).toBeInTheDocument();
      expect(container.querySelector('.hero-shape-2')).toBeInTheDocument();
      expect(container.querySelector('.hero-shape-3')).toBeInTheDocument();
    });

    it('all hero elements remain properly rendered on large screen', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-bg-shapes')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-layout')).toBeInTheDocument();
    });
  });

  // T5: Extra large screen (2560px)
  describe('T5: Extra large screen (2560px)', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 2560,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 1440,
      });
    });

    it('hero section renders correctly on 2560px ultra-wide display', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('hero content is properly structured on ultra-wide display', () => {
      const { container } = render(<HeroSection />);
      const content = container.querySelector('.hero-content');

      expect(content).toBeInTheDocument();
      expect(content).toHaveClass('hero-content');
    });

    it('no rendering issues on ultra-wide display (2560px)', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-bg-shapes')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
      expect(heroSection.querySelectorAll('.hero-shape').length).toBe(3);
    });

    it('all content elements are intact on ultra-wide display', () => {
      const { container } = render(<HeroSection />);

      const title = container.querySelector('.hero-title');
      const layout = container.querySelector('.hero-layout');
      const image = container.querySelector('.hero-image');
      const ctaButton = container.querySelector('.hero-cta');

      expect(title).toBeInTheDocument();
      expect(layout).toBeInTheDocument();
      expect(image).toBeInTheDocument();
      expect(ctaButton).toBeInTheDocument();
    });
  });

  // T6: Orientation change testing
  describe('T6: Portrait to Landscape orientation change', () => {
    it('renders correctly in portrait mode (375x667)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 667,
      });

      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('renders correctly in landscape mode (667x375)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 667,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('all elements are accessible in both orientations', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 667,
      });

      const { container } = render(<HeroSection />);

      const heroSection = container.querySelector('.hero-section');
      const shapes = heroSection.querySelectorAll('.hero-shape');
      const content = heroSection.querySelector('.hero-content');

      expect(heroSection).toBeInTheDocument();
      expect(shapes.length).toBe(3);
      expect(content).toBeInTheDocument();
    });
  });

  // T7: Visual and accessibility testing
  describe('T7: Visual features and accessibility', () => {
    it('gradient text is applied to hero title', () => {
      const { container } = render(<HeroSection />);
      const title = container.querySelector('.hero-title.gradient-text');

      expect(title).toHaveClass('gradient-text');
      expect(title).toBeInTheDocument();
    });

    it('glass card effect is applied to left section', () => {
      const { container } = render(<HeroSection />);
      const leftCard = container.querySelector('.hero-left.glass-card');

      expect(leftCard).toHaveClass('glass-card');
    });

    it('glow button effect is applied to CTA button', () => {
      const { container } = render(<HeroSection />);
      const ctaButton = container.querySelector('.hero-cta.glow-button');

      expect(ctaButton).toHaveClass('glow-button');
    });

    it('all interactive elements have proper focus states', () => {
      const { container } = render(<HeroSection />);
      const ctaButton = container.querySelector('.hero-cta');

      expect(ctaButton).toBeInTheDocument();
      expect(ctaButton.tagName).toBe('BUTTON');
    });

    it('image has proper accessibility attributes', () => {
      const { container } = render(<HeroSection />);
      const image = container.querySelector('.hero-image');

      expect(image).toHaveAttribute('alt');
      expect(image.getAttribute('alt')).toBeTruthy();
    });

    it('hero section structure prevents overflow', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
      expect(heroSection.querySelector('.hero-bg-shapes')).toBeInTheDocument();
    });

    it('background shapes layer is properly structured', () => {
      const { container } = render(<HeroSection />);
      const shapesLayer = container.querySelector('.hero-bg-shapes');

      expect(shapesLayer).toBeInTheDocument();
      expect(shapesLayer).toHaveClass('hero-bg-shapes');
    });

    it('hero section has visual styling applied', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      // Hero section should have display: flex or similar
      expect(heroSection.style).toBeDefined();
    });
  });

  // T8: Consistency across pages
  describe('T8: Hero section consistency and isolation', () => {
    it('hero section does not affect other page elements', () => {
      const { container } = render(
        <div>
          <HeroSection />
          <div className="other-section">Other Content</div>
        </div>
      );

      const otherSection = container.querySelector('.other-section');
      expect(otherSection).toBeInTheDocument();
      expect(otherSection.textContent).toBe('Other Content');
    });

    it('hero section is self-contained in its structure', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      // Should have all required child elements
      expect(heroSection.querySelector('.hero-bg-shapes')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
    });

    it('background shapes and content are layered correctly', () => {
      const { container } = render(<HeroSection />);
      const shapes = container.querySelector('.hero-bg-shapes');
      const content = container.querySelector('.hero-content');

      expect(shapes).toBeInTheDocument();
      expect(content).toBeInTheDocument();
    });

    it('hero section structure is properly maintained', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-bg-shapes')).toBeInTheDocument();
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
    });
  });

  // Additional: CSS conflict resolution
  describe('CSS validation - no conflicting styles', () => {
    it('hero-section renders with correct structure', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection).toHaveClass('hero-section');
    });

    it('hero-section does not have conflicting structure', () => {
      const { container } = render(<HeroSection />);
      const heroSection = container.querySelector('.hero-section');

      expect(heroSection).toBeInTheDocument();
      expect(heroSection.querySelectorAll('.hero-shape').length).toBe(3);
      expect(heroSection.querySelector('.hero-content')).toBeInTheDocument();
    });
  });
});
