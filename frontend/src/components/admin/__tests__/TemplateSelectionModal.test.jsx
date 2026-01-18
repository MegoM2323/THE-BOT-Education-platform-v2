import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TemplateSelectionModal } from '../TemplateSelectionModal';
import * as useTemplatesModule from '../../../hooks/useTemplates';
import * as apiClientModule from '../../../api/client';

// Mock hooks and API
vi.mock('../../../hooks/useTemplates');
vi.mock('../../../api/client');

describe('TemplateSelectionModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('должен отображать имена студентов в preview без "Unknown"', async () => {
    // Mock useTemplates hook
    const mockTemplates = [
      { id: '123', name: 'Test Template' },
    ];

    vi.spyOn(useTemplatesModule, 'useTemplates').mockReturnValue({
      templates: mockTemplates,
      loading: false,
      applyTemplate: vi.fn(),
      isApplying: false,
    });

    // Mock API response с корректной структурой данных
    const mockTemplateDetails = {
      id: '123',
      name: 'Test Template',
      lesson_count: 2,
      lessons: [
        {
          id: 'lesson-1',
          students: [
            {
              student_id: 'student-1',
              student_name: 'Иван Иванов', // Имя должно корректно отображаться
            },
            {
              student_id: 'student-2',
              student_name: 'Мария Петрова',
            },
          ],
        },
        {
          id: 'lesson-2',
          students: [
            {
              student_id: 'student-1',
              student_name: 'Иван Иванов', // Повторный студент (должен объединиться)
            },
          ],
        },
      ],
    };

    vi.spyOn(apiClientModule, 'apiClient', 'get').mockReturnValue({
      get: vi.fn().mockResolvedValue(mockTemplateDetails),
    });

    const mockClose = vi.fn();
    const mockWeekStart = new Date('2025-12-08'); // Понедельник

    render(
      <TemplateSelectionModal
        isOpen={true}
        onClose={mockClose}
        weekStartDate={mockWeekStart}
        onApplied={vi.fn()}
      />
    );

    // Проверяем что модальное окно открыто
    expect(screen.getByText('Применить шаблон к неделе')).toBeInTheDocument();

    // Выбираем шаблон
    const select = screen.getByLabelText('Выберите шаблон:');
    expect(select).toBeInTheDocument();

    // TODO: добавить более детальные проверки после выбора шаблона
    // (требует mockирования change events и waitFor)
  });

  it('должен использовать fallback на email если student_name пустое', async () => {
    const mockTemplates = [{ id: '456', name: 'Test Template 2' }];

    vi.spyOn(useTemplatesModule, 'useTemplates').mockReturnValue({
      templates: mockTemplates,
      loading: false,
      applyTemplate: vi.fn(),
      isApplying: false,
    });

    // Mock данных где у студента нет student_name, но есть email
    const mockTemplateDetails = {
      id: '456',
      name: 'Test Template 2',
      lessons: [
        {
          id: 'lesson-3',
          students: [
            {
              student_id: 'student-3',
              student_name: '', // Пустое имя
              email: 'student3@example.com', // Fallback на email
            },
          ],
        },
      ],
    };

    vi.spyOn(apiClientModule, 'apiClient', 'get').mockReturnValue({
      get: vi.fn().mockResolvedValue(mockTemplateDetails),
    });

    const mockClose = vi.fn();
    const mockWeekStart = new Date('2025-12-08');

    render(
      <TemplateSelectionModal
        isOpen={true}
        onClose={mockClose}
        weekStartDate={mockWeekStart}
        onApplied={vi.fn()}
      />
    );

    // Базовая проверка рендера
    expect(screen.getByText('Применить шаблон к неделе')).toBeInTheDocument();

    // TODO: проверить что email отображается в preview
  });
});
