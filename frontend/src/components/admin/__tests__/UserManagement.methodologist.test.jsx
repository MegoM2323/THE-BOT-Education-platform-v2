import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import UserManagement from '../UserManagement';
import * as usersAPI from '../../../api/users';

// Mock API
vi.mock('../../../api/users', () => ({
  getUsersAll: vi.fn(),
  createUser: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
}));

// Mock notification hook
vi.mock('../../../hooks/useNotification', () => ({
  useNotification: () => ({
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  }),
}));

// Mock auth hook
vi.mock('../../../hooks/useAuth', () => ({
  useAuth: () => ({
    user: { id: '1', role: 'admin', email: 'admin@test.com' },
    isAuthenticated: true,
    loading: false,
  }),
}));

// Mock slow connection hook
vi.mock('../../../hooks/useSlowConnection', () => ({
  useSlowConnection: () => false,
}));

// Helper to wrap component with providers
const renderWithProviders = (component) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        {component}
      </BrowserRouter>
    </QueryClientProvider>
  );
};

describe('UserManagement - Методист роль', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('должен отображать роль Методист в таблице', async () => {
    const mockUsers = [
      {
        id: '1',
        email: 'methodologist@test.com',
        full_name: 'Test Methodologist',
        role: 'methodologist',
      },
      {
        id: '2',
        email: 'student@test.com',
        full_name: 'Test Student',
        role: 'student',
      },
    ];

    usersAPI.getUsersAll.mockResolvedValue(mockUsers);

    renderWithProviders(<UserManagement />);

    // Ждем загрузки данных
    await waitFor(() => {
      expect(screen.getByText('Test Methodologist')).toBeInTheDocument();
    });

    // Проверяем что getRoleLabel правильно отображает роль
    expect(screen.getByText('Методист')).toBeInTheDocument();
    expect(screen.getByText('Студент')).toBeInTheDocument();
  });

  it('методист должен быть в списке ролей для создания', async () => {
    // Создаем хотя бы одного пользователя чтобы избежать скелетона
    const mockUsers = [
      {
        id: '1',
        email: 'admin@test.com',
        full_name: 'Admin',
        role: 'admin',
      },
    ];
    usersAPI.getUsersAll.mockResolvedValue(mockUsers);

    renderWithProviders(<UserManagement />);

    // Ждем загрузки
    await waitFor(() => {
      expect(screen.getByText('Admin')).toBeInTheDocument();
    });

    // Открываем форму создания
    const createButton = screen.getByTestId('create-user-button');
    await userEvent.click(createButton);

    // Проверяем что есть опция "Методист"
    const roleSelect = screen.getByTestId('role-select');
    expect(roleSelect).toBeInTheDocument();

    const options = Array.from(roleSelect.options).map(opt => opt.text);
    expect(options).toContain('Методист');
  });

  it('методист должен иметь корректное значение в select', async () => {
    const mockUsers = [
      {
        id: '1',
        email: 'admin@test.com',
        full_name: 'Admin',
        role: 'admin',
      },
    ];
    usersAPI.getUsersAll.mockResolvedValue(mockUsers);

    renderWithProviders(<UserManagement />);

    await waitFor(() => {
      expect(screen.getByText('Admin')).toBeInTheDocument();
    });

    // Открываем форму создания
    const createButton = screen.getByTestId('create-user-button');
    await userEvent.click(createButton);

    // Выбираем роль "Методист"
    const roleSelect = screen.getByTestId('role-select');
    await userEvent.selectOptions(roleSelect, 'methodologist');

    // Проверяем что значение установлено
    expect(roleSelect.value).toBe('methodologist');
  });

  it('создание пользователя с ролью методист должно работать', async () => {
    const mockUsers = [
      {
        id: '1',
        email: 'admin@test.com',
        full_name: 'Admin',
        role: 'admin',
      },
    ];
    usersAPI.getUsersAll.mockResolvedValue(mockUsers);
    usersAPI.createUser.mockResolvedValue({
      id: '3',
      email: 'newmethodologist@test.com',
      full_name: 'New Methodologist',
      role: 'methodologist',
    });

    renderWithProviders(<UserManagement />);

    await waitFor(() => {
      expect(screen.getByText('Admin')).toBeInTheDocument();
    });

    // Открываем форму создания
    const createButton = screen.getByTestId('create-user-button');
    await userEvent.click(createButton);

    // Заполняем форму
    const emailInput = screen.getByTestId('email-input');
    const fullNameInput = screen.getByTestId('full-name-input');
    const passwordInput = screen.getByTestId('password-input');
    const roleSelect = screen.getByTestId('role-select');

    await userEvent.type(emailInput, 'newmethodologist@test.com');
    await userEvent.type(fullNameInput, 'New Methodologist');
    await userEvent.type(passwordInput, 'password123');
    await userEvent.selectOptions(roleSelect, 'methodologist');

    // Отправляем форму
    const submitButton = screen.getByText('Создать');
    await userEvent.click(submitButton);

    // Проверяем что API вызвался с правильными данными
    await waitFor(() => {
      expect(usersAPI.createUser).toHaveBeenCalledWith({
        email: 'newmethodologist@test.com',
        full_name: 'New Methodologist',
        password: 'password123',
        role: 'methodologist',
      });
    });
  });

  it('редактирование пользователя с изменением роли на методист должно работать', async () => {
    const mockUsers = [
      {
        id: '2',
        email: 'student@test.com',
        full_name: 'Test Student',
        role: 'student',
      },
    ];

    usersAPI.getUsersAll.mockResolvedValue(mockUsers);
    usersAPI.updateUser.mockResolvedValue({
      id: '2',
      email: 'student@test.com',
      full_name: 'Test Student',
      role: 'methodologist',
    });

    renderWithProviders(<UserManagement />);

    // Ждем загрузки данных
    await waitFor(() => {
      expect(screen.getByText('Test Student')).toBeInTheDocument();
    });

    // Открываем форму редактирования
    const editButtons = screen.getAllByText('Редактировать');
    await userEvent.click(editButtons[0]);

    // Изменяем роль на методист
    const roleSelect = screen.getByTestId('role-select');
    await userEvent.selectOptions(roleSelect, 'methodologist');

    // Сохраняем изменения
    const saveButton = screen.getByText('Сохранить');
    await userEvent.click(saveButton);

    // Проверяем что API вызвался с правильными данными
    await waitFor(() => {
      expect(usersAPI.updateUser).toHaveBeenCalledWith(
        '2',
        expect.objectContaining({
          role: 'methodologist',
        })
      );
    });
  });

  it('должен отображать несколько методистов в таблице', async () => {
    const mockUsers = [
      {
        id: '1',
        email: 'methodologist1@test.com',
        full_name: 'First Methodologist',
        role: 'methodologist',
      },
      {
        id: '2',
        email: 'methodologist2@test.com',
        full_name: 'Second Methodologist',
        role: 'methodologist',
      },
      {
        id: '3',
        email: 'admin@test.com',
        full_name: 'Admin User',
        role: 'admin',
      },
    ];

    usersAPI.getUsersAll.mockResolvedValue(mockUsers);

    renderWithProviders(<UserManagement />);

    // Ждем загрузки данных
    await waitFor(() => {
      expect(screen.getByText('First Methodologist')).toBeInTheDocument();
      expect(screen.getByText('Second Methodologist')).toBeInTheDocument();
    });

    // Проверяем что оба методиста отображаются с правильной ролью
    const methodologistLabels = screen.getAllByText('Методист');
    expect(methodologistLabels).toHaveLength(2);
  });

  it('getRoleLabel должен корректно возвращать "Методист" для роли methodologist', async () => {
    const mockUsers = [
      {
        id: '1',
        email: 'methodologist@test.com',
        full_name: 'Test Methodologist',
        role: 'methodologist',
      },
    ];

    usersAPI.getUsersAll.mockResolvedValue(mockUsers);

    renderWithProviders(<UserManagement />);

    await waitFor(() => {
      const methodologistLabel = screen.getByText('Методист');
      expect(methodologistLabel).toBeInTheDocument();
    });
  });
});
