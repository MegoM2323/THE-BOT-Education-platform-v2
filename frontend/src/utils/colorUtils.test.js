/**
 * Tests for color utility functions
 */

import {
  hexToRgb,
  calculateLuminance,
  getContrastRatio,
  getContrastTextColor,
  verifyContrastRatio,
  getSuggestedTextColor,
} from './colorUtils.js';

describe('colorUtils', () => {
  describe('hexToRgb', () => {
    test('converts hex with # prefix', () => {
      const result = hexToRgb('#FF5733');
      expect(result).toEqual({ r: 255, g: 87, b: 51 });
    });

    test('converts hex without # prefix', () => {
      const result = hexToRgb('FF5733');
      expect(result).toEqual({ r: 255, g: 87, b: 51 });
    });

    test('handles white color', () => {
      const result = hexToRgb('#FFFFFF');
      expect(result).toEqual({ r: 255, g: 255, b: 255 });
    });

    test('handles black color', () => {
      const result = hexToRgb('#000000');
      expect(result).toEqual({ r: 0, g: 0, b: 0 });
    });

    test('handles invalid hex gracefully', () => {
      const result = hexToRgb('invalid');
      // Should fallback to default color #000000
      expect(result).toEqual({ r: 0, g: 0, b: 0 });
    });
  });

  describe('calculateLuminance', () => {
    test('white has luminance close to 1', () => {
      const luminance = calculateLuminance(255, 255, 255);
      expect(luminance).toBeCloseTo(1, 2);
    });

    test('black has luminance close to 0', () => {
      const luminance = calculateLuminance(0, 0, 0);
      expect(luminance).toBeCloseTo(0, 2);
    });

    test('calculates luminance for red', () => {
      const luminance = calculateLuminance(255, 0, 0);
      expect(luminance).toBeGreaterThan(0);
      expect(luminance).toBeLessThan(1);
    });

    test('calculates luminance for grey', () => {
      const luminance = calculateLuminance(128, 128, 128);
      expect(luminance).toBeGreaterThan(0);
      expect(luminance).toBeLessThan(1);
    });
  });

  describe('getContrastRatio', () => {
    test('white and black have maximum contrast', () => {
      const ratio = getContrastRatio(1, 0);
      expect(ratio).toBe(21);
    });

    test('same color has minimum contrast', () => {
      const ratio = getContrastRatio(0.5, 0.5);
      expect(ratio).toBe(1);
    });

    test('works regardless of parameter order', () => {
      const ratio1 = getContrastRatio(1, 0);
      const ratio2 = getContrastRatio(0, 1);
      expect(ratio1).toBe(ratio2);
    });
  });

  describe('getContrastTextColor', () => {
    test('white background returns black text', () => {
      expect(getContrastTextColor('#FFFFFF')).toBe('#000000');
    });

    test('black background returns white text', () => {
      expect(getContrastTextColor('#000000')).toBe('#FFFFFF');
    });

    test('light blue background returns black text', () => {
      expect(getContrastTextColor('#87CEEB')).toBe('#000000');
    });

    test('dark navy background returns white text', () => {
      expect(getContrastTextColor('#001F3F')).toBe('#FFFFFF');
    });

    test('red background returns appropriate text color', () => {
      const textColor = getContrastTextColor('#FF0000');
      expect(['#000000', '#FFFFFF']).toContain(textColor);
    });

    test('blue background returns appropriate text color', () => {
      const textColor = getContrastTextColor('#0000FF');
      expect(['#000000', '#FFFFFF']).toContain(textColor);
    });

    test('grey background returns appropriate text color', () => {
      const textColor = getContrastTextColor('#808080');
      expect(['#000000', '#FFFFFF']).toContain(textColor);
    });

    test('handles null/undefined gracefully', () => {
      // Default color #000000 (black) has zero luminance, so uses white text
      expect(getContrastTextColor(null)).toBe('#FFFFFF');
      expect(getContrastTextColor(undefined)).toBe('#FFFFFF');
    });
  });

  describe('verifyContrastRatio', () => {
    test('white on black passes WCAG AA', () => {
      const result = verifyContrastRatio('#000000', '#FFFFFF', 'normal');
      expect(result.passes).toBe(true);
      expect(result.ratio).toBeGreaterThanOrEqual(4.5);
    });

    test('light grey on white fails WCAG AA', () => {
      const result = verifyContrastRatio('#FFFFFF', '#DDDDDD', 'normal');
      expect(result.passes).toBe(false);
      expect(result.ratio).toBeLessThan(4.5);
    });

    test('large text has lower requirement', () => {
      const result = verifyContrastRatio('#FFFFFF', '#AAAAAA', 'large');
      expect(result.required).toBe(3);
    });

    test('normal text has higher requirement', () => {
      const result = verifyContrastRatio('#FFFFFF', '#AAAAAA', 'normal');
      expect(result.required).toBe(4.5);
    });
  });

  describe('getSuggestedTextColor', () => {
    test('returns color with contrast details', () => {
      const result = getSuggestedTextColor('#FFFFFF');
      expect(result).toHaveProperty('color');
      expect(result).toHaveProperty('luminance');
      expect(result).toHaveProperty('contrast');
      expect(result).toHaveProperty('passes');
    });

    test('suggested color passes WCAG AA', () => {
      const result = getSuggestedTextColor('#3B82F6');
      expect(result.passes).toBe(true);
      expect(result.contrast).toBeGreaterThanOrEqual(4.5);
    });

    test('luminance is between 0 and 1', () => {
      const result = getSuggestedTextColor('#FF5733');
      expect(result.luminance).toBeGreaterThanOrEqual(0);
      expect(result.luminance).toBeLessThanOrEqual(1);
    });
  });

  describe('Edge cases and real color combinations', () => {
    const testColors = [
      { bg: '#FFFFFF', expected: '#000000', name: 'Pure white' },
      { bg: '#000000', expected: '#FFFFFF', name: 'Pure black' },
      { bg: '#FF0000', expected: '#000000', name: 'Red (dark)' },
      { bg: '#00FF00', expected: '#000000', name: 'Green' },
      { bg: '#0000FF', expected: '#FFFFFF', name: 'Blue' },
      { bg: '#FFFF00', expected: '#000000', name: 'Yellow (light)' },
      { bg: '#FF00FF', expected: '#000000', name: 'Magenta' },
      { bg: '#00FFFF', expected: '#000000', name: 'Cyan (light)' },
      { bg: '#808080', expected: '#000000', name: 'Grey (mid luminance)' },
      { bg: '#3B82F6', expected: '#000000', name: 'Primary blue (dark)' },
    ];

    testColors.forEach(({ bg, expected, name }) => {
      test(`${name} (${bg}) returns ${expected}`, () => {
        const result = getContrastTextColor(bg);
        expect(result).toBe(expected);
      });
    });
  });

  describe('WCAG compliance verification', () => {
    test('all suggested colors meet WCAG AA standard', () => {
      const testBackgrounds = [
        '#FFFFFF', '#000000', '#FF0000', '#00FF00', '#0000FF',
        '#FFFF00', '#FF00FF', '#00FFFF', '#808080', '#3B82F6',
        '#EF4444', '#10B981', '#F59E0B', '#8B5CF6', '#EC4899',
      ];

      testBackgrounds.forEach(bg => {
        const textColor = getContrastTextColor(bg);
        const verification = verifyContrastRatio(bg, textColor, 'normal');
        expect(verification.passes).toBe(true);
        expect(verification.ratio).toBeGreaterThanOrEqual(4.5);
      });
    });
  });
});
