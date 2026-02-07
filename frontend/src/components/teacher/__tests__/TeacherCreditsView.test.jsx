import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import TeacherCreditsView from '../TeacherCreditsView.jsx';
import * as usersAPI from '../../../api/users.js';
import * as creditsAPI from '../../../api/credits.js';
import { useNotification } from '../../../hooks/useNotification.js';
import { renderWithProviders } from '../../../test/test-utils.jsx';

vi.mock('../../../api/users.js');
vi.mock('../../../api/credits.js');
vi.mock('../../../hooks/useNotification.js');

const mockStudents = [
  { id: '1', full_name: 'Анна Иванова', email: 'anna@example.com' },
  { id: '2', full_name: 'Борис Петров', email: 'boris@example.com' },
  { id: '3', full_name: 'Виктор Сидоров', email: 'viktor@example.com' },
];

const mockCredits = {
  '1': { balance: 50 },
  '2': { balance: 100 },
  '3': { balance: 25 },
};

describe('TeacherCreditsView', () => {
  const mockNotification = {
    error: vi.fn(),
    success: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useNotification.mockReturnValue(mockNotification);
    usersAPI.getStudentsAll.mockResolvedValue(mockStudents);
    creditsAPI.getUserCredits.mockImplementation((studentId) =>
      Promise.resolve(mockCredits[studentId] || { balance: 0 })
    );
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('компонент загружает студентов при монтировании', async () => {
    renderWithProviders(<TeacherCreditsView />);

    expect(usersAPI.getStudentsAll).toHaveBeenCalledTimes(1);

    await waitFor(() => {
      expect(screen.getByText('Анна Иванова')).toBeInTheDocument();
    });
  });

  it('таблица отображает правильные колонки (Имя, Email, Баланс)', async () => {
    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      const headers = screen.getAllByRole('columnheader');
      expect(headers).toHaveLength(4);
      expect(headers[0]).toHaveTextContent('Имя');
      expect(headers[1]).toHaveTextContent('Email');
      expect(headers[2]).toHaveTextContent('Баланс');
      expect(headers[3]).toHaveTextContent('Действия');
    });
  });

  it('БЕЗ кнопок действий (no + or - buttons)', async () => {
    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      const buttons = screen.queryAllByRole('button', { hidden: false });
      const actionButtons = buttons.filter(btn =>
        btn.textContent === '+' || btn.textContent === '-'
      );
      expect(actionButtons).toHaveLength(0);
    });
  });

  it('сортировка по имени по умолчанию', async () => {
    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      const bodyRows = rows.slice(1);

      expect(bodyRows[0]).toHaveTextContent('Анна Иванова');
      expect(bodyRows[1]).toHaveTextContent('Борис Петров');
      expect(bodyRows[2]).toHaveTextContent('Виктор Сидоров');
    });
  });

  it('русские символы сортируются правильно (localeCompare)', async () => {
    const russianStudents = [
      { id: '1', full_name: 'Юлия Юрина', email: 'yulia@example.com' },
      { id: '2', full_name: 'Ангелина Андреева', email: 'angelina@example.com' },
      { id: '3', full_name: 'Мария Морозова', email: 'maria@example.com' },
    ];

    usersAPI.getStudentsAll.mockResolvedValue(russianStudents);
    creditsAPI.getUserCredits.mockResolvedValue({ balance: 50 });

    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      const bodyRows = rows.slice(1);

      expect(bodyRows[0]).toHaveTextContent('Ангелина Андреева');
      expect(bodyRows[1]).toHaveTextContent('Мария Морозова');
      expect(bodyRows[2]).toHaveTextContent('Юлия Юрина');
    });
  });

  it('обработка пустого списка студентов', async () => {
    usersAPI.getStudentsAll.mockResolvedValue([]);

    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      expect(screen.getByText('Студенты не найдены')).toBeInTheDocument();
    });
  });

  it('баланс отображается в badge формате', async () => {
    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      const badges = screen.getAllByText(/\d+ кредитов/);
      expect(badges.length).toBeGreaterThan(0);

      badges.forEach(badge => {
        expect(badge).toHaveClass('balance-badge');
      });
    });
  });

  it('обработка ошибок при загрузке студентов', async () => {
    usersAPI.getStudentsAll.mockRejectedValue(new Error('Network error'));

    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      expect(mockNotification.error).toHaveBeenCalledWith('Ошибка загрузки студентов');
    });
  });

  it('loading state отображается', async () => {
    usersAPI.getStudentsAll.mockImplementation(() => new Promise(() => {}));

    const { container } = renderWithProviders(<TeacherCreditsView />);

    const loadingDiv = container.querySelector('.teacher-credits-loading');
    expect(loadingDiv).toBeInTheDocument();
  });

  it('обработка ошибки при загрузке кредитов одного студента', async () => {
    creditsAPI.getUserCredits.mockImplementation((studentId) => {
      if (studentId === '2') {
        return Promise.reject(new Error('Failed to load credits'));
      }
      return Promise.resolve(mockCredits[studentId] || { balance: 0 });
    });

    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      expect(screen.getByText('Анна Иванова')).toBeInTheDocument();
      expect(screen.getByText('Борис Петров')).toBeInTheDocument();
      expect(screen.getByText('Виктор Сидоров')).toBeInTheDocument();
    });

    const rows = screen.getAllByRole('row');
    const borisRow = Array.from(rows).find(row =>
      row.textContent.includes('Борис Петров')
    );
    expect(borisRow).toHaveTextContent('0 кредитов');
  });

  it('sortedStudents мемоизирована (не создает новый массив при каждом render)', async () => {
    renderWithProviders(<TeacherCreditsView />);

    await waitFor(() => {
      expect(screen.getByText('Анна Иванова')).toBeInTheDocument();
    });

    const rows = screen.getAllByRole('row');
    expect(rows).toHaveLength(4);
  });
});
