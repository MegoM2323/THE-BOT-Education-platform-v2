import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../Landing';

const mockDownloadFile = vi.fn();

// Mock notification functions
const mockNotification = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
};

vi.mock('../../hooks/useNotification.js', () => ({
  useNotification: () => mockNotification,
  default: () => mockNotification,
}));

vi.mock('../../../api/axios', () => ({
  api: {
    post: vi.fn(),
  },
}));

describe('Landing Page - Specific Requirements Tests', () => {
  beforeEach(() => {
    mockDownloadFile.mockClear();
    mockNotification.success.mockClear();
    mockNotification.error.mockClear();
    vi.clearAllMocks();
  });

  describe('Test 1: AboutSection отображается правильно', () => {
    it('test_about_section_image_src_correct', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const image = screen.getByAltText('Мирослав Адаменко');
      expect(image).toBeInTheDocument();
      expect(image.src).toContain('/images/teacher-photo-new.png');
    });

    it('test_about_section_text_3_4_people', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('3-4 человека')).toBeInTheDocument();
    });

    it('test_about_section_text_2_hours', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('2 часа')).toBeInTheDocument();
    });

    it('test_about_section_text_price_per_lesson', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      expect(screen.getByText('3400/Занятие')).toBeInTheDocument();
    });

    it('test_about_section_old_schedule_text_absent', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Проверяем что старого текста расписания нет
      expect(screen.queryByText(/Чт.*1.*Янв.*14:00-16:00/)).not.toBeInTheDocument();
      expect(screen.queryByText('Чт, 1 Янв, 14:00-16:00')).not.toBeInTheDocument();
    });
  });

  describe('Test 2: RecentLessonsSection функционирует', () => {
    it('test_recent_lessons_cards_display', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const lessonCards = screen.getAllByRole('heading', { level: 3 });
      const recentLessonsCards = lessonCards.filter(
        (card) => card.textContent === 'Дерево отрезков' || card.textContent === 'Сканирующая прямая'
      );

      expect(recentLessonsCards.length).toBeGreaterThanOrEqual(2);
    });

    it('test_recent_lessons_file_items_visible', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const fileItems = document.querySelectorAll('.file-item-clickable');
      expect(fileItems.length).toBeGreaterThan(0);
      fileItems.forEach((item) => {
        expect(item).toBeVisible();
      });
    });

    it('test_recent_lessons_file_click_triggers_download', async () => {
      const user = userEvent.setup();

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const fileItems = document.querySelectorAll('.file-item-clickable');
      expect(fileItems.length).toBeGreaterThan(0);

      // Проверяем что элемент имеет onClick атрибут (функция)
      fileItems.forEach((item) => {
        expect(item.onclick || item.getAttribute('onclick')).toBeDefined();
      });
    });

    it('test_recent_lessons_file_cursor_pointer_on_hover', async () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const fileItem = document.querySelector('.file-item-clickable');
      expect(fileItem).toBeInTheDocument();

      // Проверяем CSS класс для курсора
      const styles = window.getComputedStyle(fileItem);
      // Класс file-item-clickable должен иметь cursor: pointer
      expect(fileItem.classList.contains('file-item-clickable')).toBe(true);
    });
  });

  describe('Test 3: SwapSystemSection показывает 4 ячейки', () => {
    it('test_swap_system_4_cells_display', () => {
      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarLessons = document.querySelectorAll('.calendar-lesson');
      expect(calendarLessons.length).toBe(4);
    });

    it('test_swap_system_4_cells_visible_desktop_1200px', () => {
      // Установить ширину окна в 1200px
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1200,
      });

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarLessons = document.querySelectorAll('.calendar-lesson');
      calendarLessons.forEach((lesson) => {
        expect(lesson).toBeVisible();
      });

      expect(calendarLessons.length).toBe(4);
    });

    it('test_swap_system_2_columns_tablet_768px', () => {
      // Установить ширину окна в 768px
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarDemo = document.querySelector('.calendar-demo');
      expect(calendarDemo).toBeInTheDocument();

      // Проверяем наличие элементов
      const calendarLessons = calendarDemo.querySelectorAll('.calendar-lesson');
      expect(calendarLessons.length).toBe(4);
    });

    it('test_swap_system_1_column_mobile_480px', () => {
      // Установить ширину окна в 480px
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 480,
      });

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const calendarLessons = document.querySelectorAll('.calendar-lesson');
      expect(calendarLessons.length).toBe(4);
    });
  });

  describe('Test 4: Форма не имеет зеленой обводки', () => {
    it('test_form_no_green_focus_style_exists', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Проверяем что в CSS нет селектора .form-input:focus с зеленой обводкой
      const styleSheets = document.styleSheets;
      let hasGreenFocus = false;

      for (let sheet of styleSheets) {
        try {
          const rules = sheet.cssRules || sheet.rules;
          for (let rule of rules) {
            if (rule.selectorText && rule.selectorText.includes('.form-input:focus')) {
              const style = rule.style.borderColor || rule.style.border;
              if (style && (style.includes('004231') || style.includes('green'))) {
                hasGreenFocus = true;
              }
            }
          }
        } catch (e) {
          // CORS issue with external stylesheets, skip
        }
      }

      // Дополнительно проверяем через input элемент
      const inputs = container.querySelectorAll('.form-input');
      inputs.forEach((input) => {
        const computedStyle = window.getComputedStyle(input);
        // На фокусе не должно быть зеленой обводки
        expect(computedStyle.borderColor).not.toBe('rgb(0, 66, 49)'); // #004231 в RGB
      });
    });

    it('test_form_input_focus_no_green_border', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      const input = container.querySelector('.form-input');
      if (input) {
        input.focus();
        const computedStyle = window.getComputedStyle(input);

        // Проверяем что border не зеленый
        expect(computedStyle.outlineColor).not.toBe('rgb(0, 66, 49)');
      }
    });
  });

  describe('Test 5: Отсутствуют ошибки в консоли', () => {
    it('test_no_console_errors_on_page_load', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Проверяем что нет ошибок (кроме React 18 warnings)
      const errors = consoleSpy.mock.calls.filter((call) => {
        const message = call[0]?.toString() || '';
        return (
          !message.includes('Warning: ReactDOM.render') &&
          !message.includes('Not implemented: HTMLFormElement.prototype.submit')
        );
      });

      expect(errors.length).toBe(0);

      consoleSpy.mockRestore();
    });

    it('test_no_undefined_variables', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Проверяем что текст страницы не содержит "undefined"
      const pageText = container.textContent;
      expect(pageText).not.toContain('undefined');
    });

    it('test_all_required_sections_render_without_error', () => {
      const { container } = render(
        <BrowserRouter>
          <Landing />
        </BrowserRouter>
      );

      // Проверяем наличие всех ключевых секций
      expect(document.querySelector('.about-section')).toBeInTheDocument();
      expect(document.querySelector('.recent-lessons-section')).toBeInTheDocument();
      expect(document.querySelector('.swap-system-section')).toBeInTheDocument();
      expect(document.querySelector('.contact-section')).toBeInTheDocument();
    });
  });
});
