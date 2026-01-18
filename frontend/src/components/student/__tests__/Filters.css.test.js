/**
 * CSS Tests for Filters.css - Filter Group and Checkbox Fixes
 * Test Scenarios:
 * 1. .filter-group has flex-shrink: 0 (not 1)
 * 2. .filter-group has white-space: nowrap
 * 3. .filter-checkbox-group has min-height: 40px
 */

import { readFileSync } from 'fs';
import { resolve } from 'path';

describe('Filters.css - Filter Group and Checkbox Fixes', () => {
  let cssContent;

  beforeAll(() => {
    const cssPath = resolve(__dirname, '../Filters.css');
    cssContent = readFileSync(cssPath, 'utf-8');
  });

  // ========== Filter Group Tests ==========

  test('filter-group CSS rule exists', () => {
    expect(cssContent).toContain('.filter-group');
  });

  test('filter-group should have flex-shrink: 0 (not 1)', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    const styles = filterGroupSection[1];

    // Must have flex-shrink: 0
    expect(styles).toContain('flex-shrink: 0');

    // Should NOT have flex-shrink: 1
    expect(styles).not.toContain('flex-shrink: 1');
  });

  test('filter-group should have display: flex', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    expect(filterGroupSection[1]).toContain('display: flex');
  });

  test('filter-group should have flex-direction: column', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    expect(filterGroupSection[1]).toContain('flex-direction: column');
  });

  test('filter-group should have white-space: nowrap', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    expect(filterGroupSection[1]).toContain('white-space: nowrap');
  });

  test('filter-group should have min-width: 180px', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    expect(filterGroupSection[1]).toContain('min-width: 180px');
  });

  test('filter-group should have max-width: 100%', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    expect(filterGroupSection[1]).toContain('max-width: 100%');
  });

  test('filter-group should have gap: 6px', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    expect(filterGroupSection[1]).toContain('gap: 6px');
  });

  test('filter-group prevents overflow with flex-shrink: 0 and white-space: nowrap', () => {
    const filterGroupSection = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroupSection).toBeTruthy();
    const styles = filterGroupSection[1];

    // Critical fix: flex-shrink: 0 prevents group from shrinking
    expect(styles).toContain('flex-shrink: 0');

    // white-space: nowrap prevents label wrapping
    expect(styles).toContain('white-space: nowrap');
  });

  // ========== Filter Checkbox Group Tests ==========

  test('filter-checkbox-group CSS rule exists', () => {
    expect(cssContent).toContain('.filter-checkbox-group');
  });

  test('filter-checkbox-group should have min-height: 40px', () => {
    const checkboxGroupSection = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroupSection).toBeTruthy();
    expect(checkboxGroupSection[1]).toContain('min-height: 40px');
  });

  test('filter-checkbox-group should have display: flex', () => {
    const checkboxGroupSection = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroupSection).toBeTruthy();
    expect(checkboxGroupSection[1]).toContain('display: flex');
  });

  test('filter-checkbox-group should have align-items: center', () => {
    const checkboxGroupSection = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroupSection).toBeTruthy();
    expect(checkboxGroupSection[1]).toContain('align-items: center');
  });

  test('filter-checkbox-group should have padding: 0', () => {
    const checkboxGroupSection = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroupSection).toBeTruthy();
    expect(checkboxGroupSection[1]).toContain('padding: 0');
  });

  test('filter-checkbox-group should have min-width: auto', () => {
    const checkboxGroupSection = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroupSection).toBeTruthy();
    expect(checkboxGroupSection[1]).toContain('min-width: auto');
  });

  test('filter-checkbox-group maintains minimum height for proper alignment', () => {
    const checkboxGroupSection = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroupSection).toBeTruthy();
    const styles = checkboxGroupSection[1];

    // min-height: 40px ensures touch targets are large enough
    expect(styles).toContain('min-height: 40px');

    // align-items: center ensures vertical centering
    expect(styles).toContain('align-items: center');
  });

  // ========== Filter Checkbox Tests ==========

  test('filter-checkbox CSS rule exists', () => {
    expect(cssContent).toContain('.filter-checkbox');
  });

  test('filter-checkbox should have correct dimensions', () => {
    const checkboxSection = cssContent.match(/\.filter-checkbox\s*{([^}]+)}/);
    expect(checkboxSection).toBeTruthy();
    const styles = checkboxSection[1];

    expect(styles).toContain('width: 18px');
    expect(styles).toContain('height: 18px');
  });

  test('filter-checkbox should have cursor: pointer', () => {
    const checkboxSection = cssContent.match(/\.filter-checkbox\s*{([^}]+)}/);
    expect(checkboxSection).toBeTruthy();
    expect(checkboxSection[1]).toContain('cursor: pointer');
  });

  test('filter-checkbox should have accent-color for styling', () => {
    const checkboxSection = cssContent.match(/\.filter-checkbox\s*{([^}]+)}/);
    expect(checkboxSection).toBeTruthy();
    expect(checkboxSection[1]).toContain('accent-color: var(--color-info)');
  });

  // ========== Filter Checkbox Label Tests ==========

  test('filter-checkbox-label CSS rule exists', () => {
    expect(cssContent).toContain('.filter-checkbox-label');
  });

  test('filter-checkbox-label should have display: flex', () => {
    const labelSection = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(labelSection).toBeTruthy();
    expect(labelSection[1]).toContain('display: flex');
  });

  test('filter-checkbox-label should have align-items: center', () => {
    const labelSection = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(labelSection).toBeTruthy();
    expect(labelSection[1]).toContain('align-items: center');
  });

  test('filter-checkbox-label should have gap: 8px', () => {
    const labelSection = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(labelSection).toBeTruthy();
    expect(labelSection[1]).toContain('gap: 8px');
  });

  test('filter-checkbox-label should have cursor: pointer', () => {
    const labelSection = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(labelSection).toBeTruthy();
    expect(labelSection[1]).toContain('cursor: pointer');
  });

  test('filter-checkbox-label should have user-select: none', () => {
    const labelSection = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(labelSection).toBeTruthy();
    expect(labelSection[1]).toContain('user-select: none');
  });

  // ========== Integration & Consistency Tests ==========

  test('all filter classes have proper flex properties', () => {
    // filter-group
    const filterGroup = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroup[1]).toContain('display: flex');
    expect(filterGroup[1]).toContain('flex-shrink: 0');

    // filter-checkbox-group
    const checkboxGroup = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroup[1]).toContain('display: flex');

    // filter-checkbox-label
    const label = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(label[1]).toContain('display: flex');
  });

  test('filter styling follows consistent spacing model', () => {
    // filter-group has gap: 6px
    const filterGroup = cssContent.match(/\.filter-group\s*{([^}]+)}/);
    expect(filterGroup[1]).toContain('gap: 6px');

    // filter-checkbox-label has gap: 8px
    const label = cssContent.match(/\.filter-checkbox-label\s*{([^}]+)}/);
    expect(label[1]).toContain('gap: 8px');

    // filter-checkbox-group has padding: 0
    const checkboxGroup = cssContent.match(/\.filter-checkbox-group\s*{([^}]+)}/);
    expect(checkboxGroup[1]).toContain('padding: 0');
  });

  test('responsive design includes mobile filter styles', () => {
    // Check for filter-group in 768px or 480px media query
    expect(cssContent).toContain('@media (max-width: 768px)');
    expect(cssContent).toContain('.filter-group');
  });
});
