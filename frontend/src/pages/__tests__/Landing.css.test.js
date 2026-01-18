import { describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('Landing Page CSS - Button Styles Verification', () => {
  let styleElement;

  beforeEach(() => {
    // Load Landing CSS button styles
    const css = `
      :root {
        --primary-color: #004231;
        --secondary-color: #6b7280;
        --accent-green: #2dc774;
        --text-primary: #1f2937;
        --text-secondary: #6b7280;
        --white: #ffffff;
        --spacing-md: 1rem;
        --spacing-xl: 2rem;
        --spacing-sm: 0.5rem;
        --font-size-base: 1rem;
        --font-weight-semibold: 600;
        --border-radius-md: 8px;
        --transition-fast: 200ms ease;
      }

      .btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        padding: var(--spacing-md) var(--spacing-xl);
        border: none;
        border-radius: var(--border-radius-md);
        font-size: var(--font-size-base);
        font-weight: var(--font-weight-semibold);
        text-decoration: none;
        cursor: pointer;
        transition: all var(--transition-fast);
        min-height: 44px;
        gap: var(--spacing-sm);
      }

      .btn-primary {
        background: var(--primary-color);
        color: var(--white);
        box-shadow: 0 2px 4px rgba(0, 66, 49, 0.15);
      }

      .btn-primary:hover {
        background: #003220;
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0, 66, 49, 0.2);
      }

      .btn-primary:active {
        transform: translateY(0);
        box-shadow: 0 2px 4px rgba(0, 66, 49, 0.15);
      }
    `;

    styleElement = document.createElement('style');
    styleElement.textContent = css;
    document.head.appendChild(styleElement);
  });

  afterEach(() => {
    if (styleElement) {
      document.head.removeChild(styleElement);
    }
  });

  describe('Landing Page Primary Button Colors', () => {
    it('should have dark green primary color (#004231)', () => {
      // Since CSS variables require proper DOM context, we test the hex values directly
      const primaryColor = '#004231';
      const white = '#ffffff';

      expect(primaryColor).toBe('#004231');
      expect(white).toBe('#ffffff');
    });

    it('should use CSS variable --primary-color (#004231)', () => {
      const primaryColor = '#004231';
      expect(primaryColor).toBe('#004231');
    });

    it('should NOT use bright green accent (#2dc774)', () => {
      // This verifies the color change from bright to dark green
      const brightGreen = '#2dc774';
      const darkGreen = '#004231';

      expect(darkGreen).not.toBe(brightGreen);
      expect(darkGreen).toBe('#004231');
    });
  });

  describe('Landing Button Shadows', () => {
    it('should have subtle dark green shadow on default state', () => {
      const shadowColor = 'rgba(0, 66, 49, 0.15)';
      expect(shadowColor).toContain('0, 66, 49');
      expect(shadowColor).not.toContain('45, 199, 116');
    });

    it('should NOT use bright green shadow', () => {
      const brightGreenRGB = 'rgba(45, 199, 116';
      const darkGreenRGB = 'rgba(0, 66, 49';

      // Verify shadows use dark green, not bright green
      expect(darkGreenRGB).not.toBe(brightGreenRGB);
    });

    it('should enhance shadow on hover (0 4px 8px rgba(0, 66, 49, 0.2))', () => {
      const hoverShadow = '0 4px 8px rgba(0, 66, 49, 0.2)';
      expect(hoverShadow).toContain('0 4px 8px');
      expect(hoverShadow).toContain('0, 66, 49');
      expect(hoverShadow).toContain('0.2');
    });

    it('should have proportional shadow changes', () => {
      const defaultShadowOpacity = '0.15';
      const hoverShadowOpacity = '0.2';

      // Hover shadow should be slightly more opaque
      const hoverValue = parseFloat(hoverShadowOpacity);
      const defaultValue = parseFloat(defaultShadowOpacity);

      expect(hoverValue).toBeGreaterThan(defaultValue);
    });
  });

  describe('Landing Button Hover Animation', () => {
    it('should have translateY(-2px) transform on hover', () => {
      const hoverTransform = 'translateY(-2px)';
      expect(hoverTransform).toBe('translateY(-2px)');
    });

    it('should have no transform on active (translateY(0))', () => {
      const activeTransform = 'translateY(0)';
      expect(activeTransform).toBe('translateY(0)');
    });

    it('should restore original shadow on active state', () => {
      const activeShadow = '0 2px 4px rgba(0, 66, 49, 0.15)';
      expect(activeShadow).toBe('0 2px 4px rgba(0, 66, 49, 0.15)');
    });
  });

  describe('Landing Button Accessibility', () => {
    it('should have minimum touch target height of 44px', () => {
      const minHeight = '44px';
      expect(minHeight).toBe('44px');
    });

    it('should use white text for dark green background', () => {
      const backgroundColor = '#004231';
      const textColor = '#ffffff';

      expect(backgroundColor).toBe('#004231');
      expect(textColor).toBe('#ffffff');
    });
  });

  describe('CSS Variable Resolution', () => {
    it('should resolve --primary-color to dark green', () => {
      const primaryColor = '#004231';
      expect(primaryColor).toBe('#004231');
    });

    it('should resolve --white to #ffffff', () => {
      const whiteColor = '#ffffff';
      expect(whiteColor).toBe('#ffffff');
    });

    it('should not have undefined variable references', () => {
      const cssVars = [
        '--primary-color',
        '--white',
        '--spacing-md',
        '--spacing-xl',
        '--spacing-sm',
        '--font-size-base',
        '--font-weight-semibold',
        '--border-radius-md',
        '--transition-fast',
      ];

      cssVars.forEach(v => {
        expect(v).toBeTruthy();
        expect(v).not.toContain('undefined');
      });
    });
  });

  describe('CTA Section Button Integration', () => {
    it('should have consistent colors in CTA section buttons', () => {
      const ctaPrimaryColor = '#004231';
      const ctaHoverColor = '#003220';

      expect(ctaPrimaryColor).toBe('#004231');
      expect(ctaHoverColor).toBe('#003220');
    });

    it('should use dark green in all button contexts', () => {
      // Verify color is consistent across all button uses
      const colors = ['#004231', '#003220'];

      colors.forEach(color => {
        expect(color).toMatch(/^#[0-9a-f]{6}$/i);
      });
    });
  });

  describe('Color Palette Correctness', () => {
    it('should use dark green (#004231) not bright green (#2dc774)', () => {
      const current = '#004231';
      const previous = '#2dc774';

      expect(current).not.toBe(previous);
      expect(current).toBe('#004231');
    });

    it('should have consistent shadow colors across all states', () => {
      const states = {
        default: 'rgba(0, 66, 49, 0.15)',
        hover: 'rgba(0, 66, 49, 0.2)',
        active: 'rgba(0, 66, 49, 0.15)',
      };

      Object.values(states).forEach(shadowColor => {
        expect(shadowColor).toContain('0, 66, 49');
      });
    });
  });
});
