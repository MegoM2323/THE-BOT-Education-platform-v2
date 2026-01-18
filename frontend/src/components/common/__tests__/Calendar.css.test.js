import '@testing-library/jest-dom';

/**
 * Calendar Component CSS Tests
 * Verifies CSS styles for lesson cards and text overflow handling
 */

describe('Calendar.css - CSS Styles', () => {
  let styleSheet;

  beforeAll(() => {
    // Load the CSS file for testing
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = '/components/common/Calendar.css';
    document.head.appendChild(link);

    // Also try to load CSS rules programmatically for testing
    try {
      const style = document.createElement('style');
      style.textContent = require('../Calendar.css').default || '';
      document.head.appendChild(style);
    } catch (e) {
      // CSS file may not be importable as module, that's ok
    }
  });

  describe('Scenario 1: .calendar-lesson min-height verification', () => {
    it('should have min-height: 70px on .calendar-lesson', () => {
      const element = document.createElement('div');
      element.className = 'calendar-lesson';
      document.body.appendChild(element);

      // Get computed style
      const computedStyle = window.getComputedStyle(element);
      const minHeight = computedStyle.minHeight;

      // Note: In testing environment, CSS may not load, so we verify the CSS text instead
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toContain('.calendar-lesson');
      expect(cssText).toMatch(/\.calendar-lesson\s*\{[\s\S]*?min-height:\s*70px/);

      document.body.removeChild(element);
    });
  });

  describe('Scenario 2: .calendar-lesson-time text overflow styles', () => {
    it('should have overflow: hidden on .calendar-lesson-time', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toContain('.calendar-lesson-time');
      expect(cssText).toMatch(/\.calendar-lesson-time\s*\{[\s\S]*?overflow:\s*hidden/);
    });

    it('should have text-overflow: ellipsis on .calendar-lesson-time', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toMatch(/\.calendar-lesson-time\s*\{[\s\S]*?text-overflow:\s*ellipsis/);
    });

    it('should have white-space: nowrap on .calendar-lesson-time', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toMatch(/\.calendar-lesson-time\s*\{[\s\S]*?white-space:\s*nowrap/);
    });
  });

  describe('Scenario 3: .calendar-lesson-teacher text overflow styles', () => {
    it('should have overflow: hidden on .calendar-lesson-teacher', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toContain('.calendar-lesson-teacher');
      expect(cssText).toMatch(/\.calendar-lesson-teacher\s*\{[\s\S]*?overflow:\s*hidden/);
    });

    it('should have text-overflow: ellipsis on .calendar-lesson-teacher', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toMatch(/\.calendar-lesson-teacher\s*\{[\s\S]*?text-overflow:\s*ellipsis/);
    });

    it('should have white-space: nowrap on .calendar-lesson-teacher', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toMatch(/\.calendar-lesson-teacher\s*\{[\s\S]*?white-space:\s*nowrap/);
    });
  });

  describe('Scenario 4: .calendar-lesson-spots text overflow styles', () => {
    it('should have overflow: hidden on .calendar-lesson-spots', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toContain('.calendar-lesson-spots');
      expect(cssText).toMatch(/\.calendar-lesson-spots\s*\{[\s\S]*?overflow:\s*hidden/);
    });

    it('should have text-overflow: ellipsis on .calendar-lesson-spots', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toMatch(/\.calendar-lesson-spots\s*\{[\s\S]*?text-overflow:\s*ellipsis/);
    });

    it('should have white-space: nowrap on .calendar-lesson-spots', () => {
      const cssText = require('fs').readFileSync(
        require('path').join(__dirname, '../Calendar.css'),
        'utf8'
      );

      expect(cssText).toMatch(/\.calendar-lesson-spots\s*\{[\s\S]*?white-space:\s*nowrap/);
    });
  });

  describe('Scenario 5: CSS file compilation verification', () => {
    it('should load Calendar.css without errors', () => {
      // Verify the file exists and is readable
      const fs = require('fs');
      const path = require('path');
      const cssPath = path.join(__dirname, '../Calendar.css');

      expect(() => {
        fs.readFileSync(cssPath, 'utf8');
      }).not.toThrow();
    });

    it('should have valid CSS syntax', () => {
      const fs = require('fs');
      const path = require('path');
      const cssPath = path.join(__dirname, '../Calendar.css');
      const cssText = fs.readFileSync(cssPath, 'utf8');

      // Basic CSS validation: check for balanced braces
      const openBraces = (cssText.match(/\{/g) || []).length;
      const closeBraces = (cssText.match(/\}/g) || []).length;

      expect(openBraces).toBe(closeBraces);
    });

    it('should contain all required lesson-related selectors', () => {
      const fs = require('fs');
      const path = require('path');
      const cssPath = path.join(__dirname, '../Calendar.css');
      const cssText = fs.readFileSync(cssPath, 'utf8');

      const requiredSelectors = [
        '.calendar-lesson',
        '.calendar-lesson-time',
        '.calendar-lesson-teacher',
        '.calendar-lesson-spots'
      ];

      requiredSelectors.forEach(selector => {
        expect(cssText).toContain(selector);
      });
    });
  });

  describe('Integration Tests: Overflow styles consistency', () => {
    it('should have consistent overflow properties across text elements', () => {
      const fs = require('fs');
      const path = require('path');
      const cssPath = path.join(__dirname, '../Calendar.css');
      const cssText = fs.readFileSync(cssPath, 'utf8');

      // Extract the overflow pattern
      const overflowElements = [
        '.calendar-lesson-time',
        '.calendar-lesson-teacher',
        '.calendar-lesson-spots'
      ];

      const hasRequiredProperties = (selector) => {
        const selectorRegex = new RegExp(
          `${selector}\\s*\\{[\\s\\S]*?overflow:\\s*hidden[\\s\\S]*?text-overflow:\\s*ellipsis[\\s\\S]*?white-space:\\s*nowrap`,
          'g'
        );
        return selectorRegex.test(cssText);
      };

      overflowElements.forEach(element => {
        expect(hasRequiredProperties(element)).toBe(true);
      });
    });

    it('should have min-height set on main lesson container', () => {
      const fs = require('fs');
      const path = require('path');
      const cssPath = path.join(__dirname, '../Calendar.css');
      const cssText = fs.readFileSync(cssPath, 'utf8');

      const lessonMinHeightRegex = /\.calendar-lesson\s*\{[\s\S]*?min-height:\s*70px/;
      expect(cssText).toMatch(lessonMinHeightRegex);
    });
  });
});
