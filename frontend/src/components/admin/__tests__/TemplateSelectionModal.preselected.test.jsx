import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TemplateSelectionModal } from '../TemplateSelectionModal';
import * as useTemplatesModule from '../../../hooks/useTemplates';
import { apiClient } from '../../../api/client';
import * as creditsAPI from '../../../api/credits';

// Mock dependencies
vi.mock('../../../hooks/useTemplates');
vi.mock('../../../api/client');
vi.mock('../../../api/credits');

describe('TemplateSelectionModal - preselectedTemplateId', () => {
  const mockTemplates = [
    { id: 1, name: 'Template 1' },
    { id: 2, name: 'Template 2' },
    { id: 3, name: 'Template 3' },
  ];

  const mockTemplateDetails = {
    id: 2,
    name: 'Template 2',
    lesson_count: 3,
    lessons: [
      {
        id: 'lesson-1',
        day_of_week: 0,
        start_time: '10:00:00',
        end_time: '11:00:00',
        students: [
          { student_id: 1, student_name: 'Иван Иванов' },
        ],
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();

    vi.spyOn(useTemplatesModule, 'useTemplates').mockReturnValue({
      templates: mockTemplates,
      loading: false,
      applyTemplate: vi.fn().mockResolvedValue({ success: true }),
      isApplying: false,
    });

    apiClient.get = vi.fn().mockResolvedValue(mockTemplateDetails);
    apiClient.get.mockImplementation((url) => {
      if (url.includes('/bookings')) {
        return Promise.resolve({ data: { bookings: [] } });
      }
      return Promise.resolve(mockTemplateDetails);
    });

    creditsAPI.getUserCredits = vi.fn().mockResolvedValue({ balance: 10 });
  });

  describe('Без preselectedTemplateId', () => {
    it('должен показать dropdown "Выберите шаблон"', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={null}
        />
      );

      expect(screen.getByLabelText('Выберите шаблон:')).toBeInTheDocument();
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('должен показать placeholder в dropdown', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={null}
        />
      );

      expect(screen.getByText('-- Выберите шаблон --')).toBeInTheDocument();
    });

    it('должен показать все шаблоны в dropdown', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={null}
        />
      );

      const select = screen.getByRole('combobox');
      const options = Array.from(select.options);

      expect(options).toHaveLength(4); // Placeholder + 3 templates
      expect(options[1].textContent).toBe('Template 1');
      expect(options[2].textContent).toBe('Template 2');
      expect(options[3].textContent).toBe('Template 3');
    });

    it('должен иметь пустой selectedTemplateId по умолчанию', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={null}
        />
      );

      const select = screen.getByRole('combobox');
      expect(select.value).toBe('');
    });
  });

  describe('С preselectedTemplateId', () => {
    it('должен скрыть dropdown и показать info блок', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      expect(screen.queryByLabelText('Выберите шаблон:')).not.toBeInTheDocument();
      expect(screen.getByText('Шаблон:')).toBeInTheDocument();
    });

    it('должен показать название preselected шаблона в info блоке', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Template 2')).toBeInTheDocument();
      });
    });

    it('должен автоматически загрузить preview для preselected шаблона', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(apiClient.get).toHaveBeenCalledWith('/templates/2');
      });

      await waitFor(() => {
        expect(screen.getByText('Предпросмотр')).toBeInTheDocument();
      });
    });

    it('должен автоматически установить selectedTemplateId', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Иван Иванов')).toBeInTheDocument();
      });
    });

    it('должен принимать preselectedTemplateId как строку', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId="2"
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Template 2')).toBeInTheDocument();
      });
    });

    it('должен принимать preselectedTemplateId как число', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Template 2')).toBeInTheDocument();
      });
    });
  });

  describe('Изменение preselectedTemplateId', () => {
    it('должен обновить selectedTemplateId при изменении prop', async () => {
      const { rerender } = render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Template 2')).toBeInTheDocument();
      });

      // Clear mocks before rerender
      vi.clearAllMocks();

      // Mock new template response
      const newMockTemplateDetails = {
        id: 3,
        name: 'Template 3',
        lesson_count: 2,
        lessons: [
          {
            id: 'lesson-3',
            day_of_week: 2,
            start_time: '12:00:00',
            end_time: '13:00:00',
            students: [
              { student_id: 3, student_name: 'Сергей Сергеев' },
            ],
          },
        ],
      };

      apiClient.get = vi.fn().mockImplementation((url) => {
        if (url.includes('/bookings')) {
          return Promise.resolve({ data: { bookings: [] } });
        }
        return Promise.resolve(newMockTemplateDetails);
      });

      // Изменяем preselectedTemplateId
      rerender(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={3}
        />
      );

      // Ожидаем обновления шаблона через useEffect
      await waitFor(() => {
        expect(apiClient.get).toHaveBeenCalledWith('/templates/3');
      }, { timeout: 3000 });
    });

    it('должен загрузить preview для нового шаблона', async () => {
      const { rerender } = render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(apiClient.get).toHaveBeenCalledWith('/templates/2');
      });

      // Clear all mocks including console.debug
      vi.clearAllMocks();

      // Mock new template response
      const newMockTemplateDetails = {
        id: 3,
        name: 'Template 3',
        lesson_count: 2,
        lessons: [
          {
            id: 'lesson-2',
            day_of_week: 1,
            start_time: '11:00:00',
            end_time: '12:00:00',
            students: [
              { student_id: 2, student_name: 'Петр Петров' },
            ],
          },
        ],
      };

      apiClient.get = vi.fn().mockImplementation((url) => {
        if (url.includes('/bookings')) {
          return Promise.resolve({ data: { bookings: [] } });
        }
        return Promise.resolve(newMockTemplateDetails);
      });

      rerender(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={3}
        />
      );

      await waitFor(() => {
        expect(apiClient.get).toHaveBeenCalledWith('/templates/3');
      });
    });
  });

  describe('Preview загружается', () => {
    it('должен показать spinner при загрузке preview', async () => {
      apiClient.get = vi.fn().mockImplementation(() => new Promise(() => {}));

      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Загрузка предпросмотра...')).toBeInTheDocument();
      });
    });

    it('должен показать preview данные для preselected шаблона', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Предпросмотр')).toBeInTheDocument();
        // "Template 2" appears twice: in preselected info and in preview
        expect(screen.getAllByText('Template 2')).toHaveLength(2);
        expect(screen.getByText('Будет создано занятий:')).toBeInTheDocument();
        expect(screen.getByText('3')).toBeInTheDocument();
      });
    });

    it('должен показать список студентов в preview', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Иван Иванов')).toBeInTheDocument();
        expect(screen.getByText('Влияние на кредиты:')).toBeInTheDocument();
      });
    });

    it('должен показать валидацию кредитов', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(creditsAPI.getUserCredits).toHaveBeenCalledWith(1);
      });
    });

    it('должен показать ошибку при загрузке preview', async () => {
      apiClient.get = vi.fn().mockRejectedValue(new Error('Failed to load template'));

      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        // Error is displayed in .template-selection-error div
        const errorElement = screen.getByText('Failed to load template');
        expect(errorElement).toBeInTheDocument();
      }, { timeout: 3000 });
    });
  });

  describe('Применение шаблона', () => {
    it('должен применить preselected шаблон', async () => {
      const mockApplyTemplate = vi.fn().mockResolvedValue({ success: true });
      vi.spyOn(useTemplatesModule, 'useTemplates').mockReturnValue({
        templates: mockTemplates,
        loading: false,
        applyTemplate: mockApplyTemplate,
        isApplying: false,
      });

      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          onApplied={vi.fn()}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Применить шаблон')).toBeInTheDocument();
      });

      const applyButton = screen.getByText('Применить шаблон');
      fireEvent.click(applyButton);

      await waitFor(() => {
        expect(mockApplyTemplate).toHaveBeenCalledWith(2, '2026-01-19');
      });
    });

    it('должен вызвать onApplied callback после успешного применения', async () => {
      const onAppliedMock = vi.fn();

      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          onApplied={onAppliedMock}
          preselectedTemplateId={2}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Применить шаблон')).toBeInTheDocument();
      });

      const applyButton = screen.getByText('Применить шаблон');
      fireEvent.click(applyButton);

      await waitFor(() => {
        expect(onAppliedMock).toHaveBeenCalledWith({ success: true });
      });
    });
  });

  describe('Edge cases', () => {
    it('должен обработать несуществующий preselectedTemplateId', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={999}
        />
      );

      // Should show "Загрузка..." in the preselected template info section
      await waitFor(() => {
        expect(screen.getByText('Шаблон:')).toBeInTheDocument();
        // Check for loading preview spinner instead
        expect(screen.getByText('Загрузка предпросмотра...')).toBeInTheDocument();
      });
    });

    it('должен обработать preselectedTemplateId = 0', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={0}
        />
      );

      expect(screen.queryByText('Предпросмотр')).not.toBeInTheDocument();
    });

    it('должен обработать пустую строку как preselectedTemplateId', () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId=""
        />
      );

      expect(screen.getByLabelText('Выберите шаблон:')).toBeInTheDocument();
    });

    it('должен показать "Загрузка..." если шаблон не найден в списке', async () => {
      render(
        <TemplateSelectionModal
          isOpen={true}
          onClose={vi.fn()}
          weekStartDate={new Date('2026-01-19')}
          preselectedTemplateId={999}
        />
      );

      // Check for the specific "Загрузка..." text in preselected template info
      await waitFor(() => {
        const infoSection = screen.getByText('Шаблон:').closest('.preselected-template-info');
        expect(infoSection).toHaveTextContent('Загрузка...');
      });
    });
  });
});
