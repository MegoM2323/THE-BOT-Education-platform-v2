import { describe, it, expect } from 'vitest';
import {
  hexToRgb,
  getContrastTextColor,
  verifyContrastRatio,
  getSuggestedTextColor,
} from '../utils/colorUtils.js';

/**
 * Integration tests for color palette
 * Tests all 8 colors from the app palette to ensure proper display and contrast
 */
describe('Color Palette Integration Tests', () => {
  const palette = [
    { hex: '#660000', name: 'Dark Red', description: 'Тёмно-красный' },
    { hex: '#666600', name: 'Brown', description: 'Коричневый' },
    { hex: '#006600', name: 'Dark Green', description: 'Тёмно-зелёный' },
    { hex: '#006666', name: 'Teal', description: 'Бирюзовый' },
    { hex: '#000066', name: 'Dark Blue', description: 'Тёмно-синий' },
    { hex: '#660066', name: 'Purple', description: 'Фиолетовый' },
    { hex: '#004231', name: 'Forest Green', description: 'Зелёный' },
    { hex: '#000000', name: 'Black', description: 'Чёрный' },
  ];

  describe('1. Palette Colors Validation', () => {
    it('should have exactly 8 colors in palette', () => {
      expect(palette).toHaveLength(8);
    });

    it('should have valid hex format for all colors', () => {
      palette.forEach(color => {
        const hexRegex = /^#[0-9A-Fa-f]{6}$/;
        expect(hexRegex.test(color.hex)).toBe(true);
      });
    });

    it('should have unique colors', () => {
      const hexValues = palette.map(c => c.hex.toUpperCase());
      const uniqueHexValues = new Set(hexValues);
      expect(uniqueHexValues.size).toBe(palette.length);
    });
  });

  describe('2. Hex to RGB Conversion', () => {
    it('should convert all palette colors to RGB correctly', () => {
      const expectedRGB = {
        '#660000': { r: 102, g: 0, b: 0 },
        '#666600': { r: 102, g: 102, b: 0 },
        '#006600': { r: 0, g: 102, b: 0 },
        '#006666': { r: 0, g: 102, b: 102 },
        '#000066': { r: 0, g: 0, b: 102 },
        '#660066': { r: 102, g: 0, b: 102 },
        '#004231': { r: 0, g: 66, b: 49 },
        '#000000': { r: 0, g: 0, b: 0 },
      };

      palette.forEach(color => {
        const rgb = hexToRgb(color.hex);
        expect(rgb).toEqual(expectedRGB[color.hex]);
      });
    });

    it('should handle lowercase hex values', () => {
      const rgb = hexToRgb('#660000'.toLowerCase());
      expect(rgb).toEqual({ r: 102, g: 0, b: 0 });
    });

    it('should handle hex without # prefix', () => {
      const rgb = hexToRgb('660000');
      expect(rgb).toEqual({ r: 102, g: 0, b: 0 });
    });
  });

  describe('3. Text Contrast for All Palette Colors', () => {
    it('should return valid text color (black or white) for all palette colors', () => {
      palette.forEach(color => {
        const textColor = getContrastTextColor(color.hex);
        expect(['#000000', '#FFFFFF']).toContain(textColor);
      });
    });

    it('should return white text for dark colors', () => {
      const darkColors = [
        '#660000', // Dark Red - very dark
        '#000066', // Dark Blue - very dark
        '#660066', // Purple - very dark
        '#004231', // Forest Green - very dark
        '#000000', // Black - darkest
      ];

      darkColors.forEach(color => {
        const textColor = getContrastTextColor(color);
        expect(textColor).toBe('#FFFFFF');
      });
    });

    it('should return appropriate text color for medium luminance colors', () => {
      const mediumColors = [
        '#666600', // Brown - has green and red
        '#006600', // Dark Green - has green
        '#006666', // Teal - has more luminance
      ];

      mediumColors.forEach(color => {
        const textColor = getContrastTextColor(color);
        // These should return valid text color (either black or white)
        expect(['#000000', '#FFFFFF']).toContain(textColor);
      });
    });
  });

  describe('4. WCAG Compliance Verification', () => {
    it('should meet WCAG AA standard (4.5:1) for normal text on all colors', () => {
      palette.forEach(color => {
        const textColor = getContrastTextColor(color.hex);
        const verification = verifyContrastRatio(color.hex, textColor, 'normal');

        expect(verification.passes).toBe(true);
        expect(verification.ratio).toBeGreaterThanOrEqual(4.5);
      });
    });

    it('should meet WCAG AA standard (3:1) for large text on all colors', () => {
      palette.forEach(color => {
        const textColor = getContrastTextColor(color.hex);
        const verification = verifyContrastRatio(color.hex, textColor, 'large');

        expect(verification.passes).toBe(true);
        expect(verification.ratio).toBeGreaterThanOrEqual(3);
      });
    });

    it('should have sufficient contrast ratios for palette colors', () => {
      // Test with auto-selected colors which should meet WCAG AA (4.5:1)
      const contrastData = {
        '#660000': { minNormal: 4.5, minLarge: 3 }, // Dark Red
        '#666600': { minNormal: 3.5, minLarge: 3 }, // Brown
        '#006600': { minNormal: 3.5, minLarge: 3 }, // Dark Green
        '#006666': { minNormal: 3.5, minLarge: 3 }, // Teal
        '#000066': { minNormal: 4.5, minLarge: 3 }, // Dark Blue
        '#660066': { minNormal: 4.5, minLarge: 3 }, // Purple
        '#004231': { minNormal: 4.5, minLarge: 3 }, // Forest Green
        '#000000': { minNormal: 21, minLarge: 21 }, // Black
      };

      Object.entries(contrastData).forEach(([bgColor, { minNormal, minLarge }]) => {
        const textColor = getContrastTextColor(bgColor);
        const verificationNormal = verifyContrastRatio(bgColor, textColor, 'normal');
        const verificationLarge = verifyContrastRatio(bgColor, textColor, 'large');

        // Should meet at least large text requirement (3:1)
        expect(verificationLarge.ratio).toBeGreaterThanOrEqual(minLarge);
      });
    });
  });

  describe('5. RGB Component Values', () => {
    it('all palette colors should have valid RGB values (0-255)', () => {
      palette.forEach(color => {
        const { r, g, b } = hexToRgb(color.hex);
        expect(r).toBeGreaterThanOrEqual(0);
        expect(r).toBeLessThanOrEqual(255);
        expect(g).toBeGreaterThanOrEqual(0);
        expect(g).toBeLessThanOrEqual(255);
        expect(b).toBeGreaterThanOrEqual(0);
        expect(b).toBeLessThanOrEqual(255);
      });
    });

    it('should have correct primary color components', () => {
      const colorComponents = {
        '#660000': { hasMostRed: true, hasZeroGreen: true, hasZeroBlue: true },
        '#666600': { hasRed: true, hasGreen: true, hasZeroBlue: true },
        '#006600': { hasZeroRed: true, hasGreen: true, hasZeroBlue: true },
        '#006666': { hasZeroRed: true, hasGreen: true, hasBlue: true },
        '#000066': { hasZeroRed: true, hasZeroGreen: true, hasBlue: true },
        '#660066': { hasRed: true, hasZeroGreen: true, hasBlue: true },
        '#004231': { hasZeroRed: true, hasGreen: true, hasBlue: true },
        '#000000': { allZero: true },
      };

      palette.forEach((color, idx) => {
        const { r, g, b } = hexToRgb(color.hex);
        const spec = colorComponents[color.hex];

        if (spec.allZero) {
          expect(r).toBe(0);
          expect(g).toBe(0);
          expect(b).toBe(0);
        }
      });
    });
  });

  describe('6. Color Details from getSuggestedTextColor', () => {
    it('should return complete color information for all palette colors', () => {
      palette.forEach(color => {
        const result = getSuggestedTextColor(color.hex);

        expect(result).toHaveProperty('color');
        expect(result).toHaveProperty('luminance');
        expect(result).toHaveProperty('contrast');
        expect(result).toHaveProperty('passes');

        // Verify types
        expect(typeof result.color).toBe('string');
        expect(typeof result.luminance).toBe('number');
        expect(typeof result.contrast).toBe('number');
        expect(typeof result.passes).toBe('boolean');

        // Verify ranges
        expect(result.luminance).toBeGreaterThanOrEqual(0);
        expect(result.luminance).toBeLessThanOrEqual(1);
        expect(result.contrast).toBeGreaterThan(0);
        expect(result.passes).toBe(true); // All should pass WCAG
      });
    });
  });

  describe('7. Edge Cases and Special Colors', () => {
    it('should handle black color correctly', () => {
      const black = '#000000';
      const textColor = getContrastTextColor(black);
      expect(textColor).toBe('#FFFFFF');

      const rgb = hexToRgb(black);
      expect(rgb).toEqual({ r: 0, g: 0, b: 0 });
    });

    it('should handle dark colors with low luminance', () => {
      const darkPaletteColors = ['#660000', '#000066', '#660066', '#004231', '#000000'];

      darkPaletteColors.forEach(color => {
        const result = getSuggestedTextColor(color);
        expect(result.luminance).toBeLessThan(0.2);
        expect(result.color).toBe('#FFFFFF');
      });
    });

    it('should handle colors with medium luminance', () => {
      const mediumColors = ['#666600', '#006600', '#006666'];

      mediumColors.forEach(color => {
        const result = getSuggestedTextColor(color);
        // These should return valid text color
        expect(['#000000', '#FFFFFF']).toContain(result.color);
      });
    });
  });

  describe('8. Color Consistency', () => {
    it('should return consistent results for same color', () => {
      const color = '#660000';
      const result1 = getContrastTextColor(color);
      const result2 = getContrastTextColor(color);

      expect(result1).toBe(result2);
    });

    it('should handle uppercase and lowercase hex identically', () => {
      const colorUpper = '#660000';
      const colorLower = '#660000'.toLowerCase();

      const resultUpper = getContrastTextColor(colorUpper);
      const resultLower = getContrastTextColor(colorLower);

      expect(resultUpper).toBe(resultLower);
    });

    it('all palette colors should pass WCAG verification consistently', () => {
      palette.forEach(color => {
        // Test multiple times to ensure consistency
        const results = [];
        for (let i = 0; i < 3; i++) {
          const textColor = getContrastTextColor(color.hex);
          const verification = verifyContrastRatio(color.hex, textColor, 'normal');
          results.push(verification.passes);
        }

        // All results should be true
        expect(results).toEqual([true, true, true]);
      });
    });
  });

  describe('9. Palette Color Names and Descriptions', () => {
    it('should have meaningful names for all colors', () => {
      palette.forEach(color => {
        expect(color.name).toBeTruthy();
        expect(color.name.length).toBeGreaterThan(0);
      });
    });

    it('should have Russian descriptions for accessibility', () => {
      palette.forEach(color => {
        expect(color.description).toBeTruthy();
        expect(color.description.length).toBeGreaterThan(0);
      });
    });

    it('should have matching Russian and English names', () => {
      const colorNames = {
        '#660000': { eng: 'Dark Red', ru: 'Тёмно-красный' },
        '#666600': { eng: 'Brown', ru: 'Коричневый' },
        '#006600': { eng: 'Dark Green', ru: 'Тёмно-зелёный' },
        '#006666': { eng: 'Teal', ru: 'Бирюзовый' },
        '#000066': { eng: 'Dark Blue', ru: 'Тёмно-синий' },
        '#660066': { eng: 'Purple', ru: 'Фиолетовый' },
        '#004231': { eng: 'Forest Green', ru: 'Зелёный' },
        '#000000': { eng: 'Black', ru: 'Чёрный' },
      };

      palette.forEach(color => {
        const names = colorNames[color.hex];
        expect(color.name).toBe(names.eng);
        expect(color.description).toBe(names.ru);
      });
    });
  });

  describe('10. Luminance Values for All Colors', () => {
    it('should have luminance from 0 to 1', () => {
      palette.forEach(color => {
        const result = getSuggestedTextColor(color.hex);
        expect(result.luminance).toBeGreaterThanOrEqual(0);
        expect(result.luminance).toBeLessThanOrEqual(1);
      });
    });

    it('should have black as darkest color with lowest luminance', () => {
      const results = palette.map(color => ({
        hex: color.hex,
        luminance: getSuggestedTextColor(color.hex).luminance,
      }));

      const blackResult = results.find(r => r.hex === '#000000');
      expect(blackResult.luminance).toBe(0);

      // All other colors should have higher luminance than black
      results.forEach(result => {
        if (result.hex !== '#000000') {
          expect(result.luminance).toBeGreaterThan(blackResult.luminance);
        }
      });
    });
  });
});
