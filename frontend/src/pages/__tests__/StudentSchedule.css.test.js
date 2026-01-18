import { readFileSync } from 'fs';
import { resolve } from 'path';

describe('StudentSchedule.css - View Toggle', () => {
  let cssContent;

  beforeAll(() => {
    const cssPath = resolve(__dirname, '../StudentSchedule.css');
    cssContent = readFileSync(cssPath, 'utf-8');
  });

  test('view-toggle-btn CSS rule exists', () => {
    expect(cssContent).toContain('.view-toggle-btn');
  });

  test('student-schedule should have base styling', () => {
    const scheduleSection = cssContent.match(/\.student-schedule\s*{([^}]+)}/);
    expect(scheduleSection).toBeTruthy();
    expect(scheduleSection[1]).toContain('padding');
  });

  test('responsive design for mobile breakpoint exists', () => {
    const mediaRegex = /@media\s*\(max-width:\s*(480px|768px)\)/;
    const mediaMatch = cssContent.match(mediaRegex);
    expect(mediaMatch).toBeTruthy();
  });

  test('error-message styles should exist for error handling', () => {
    expect(cssContent).toContain('.error-message');
  });

  test('empty-state styles should exist for no content', () => {
    expect(cssContent).toContain('.empty-state');
  });

  test('lessons-grid styles should exist for layout', () => {
    expect(cssContent).toContain('.lessons-grid');
  });
});
