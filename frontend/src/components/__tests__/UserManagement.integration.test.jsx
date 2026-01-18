import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import UserManagement from '../admin/UserManagement.jsx';
import * as usersAPI from '../../api/users.js';

vi.mock('../../api/users.js');
vi.mock('../../hooks/useNotification.js', () => ({
  useNotification: () => ({
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  }),
}));

vi.mock('../../hooks/useSlowConnection.js', () => ({
  useSlowConnection: vi.fn(() => false),
}));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderUserManagement = () => {
  return render(
    <QueryClientProvider client={queryClient}>
      <UserManagement />
    </QueryClientProvider>
  );
};

describe('UserManagement Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('T1: Load 25+ users across multiple pages', () => {
    it('should display all 25 users from multiple API pages', async () => {
      const mockUsers = Array.from({ length: 25 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: i % 2 === 0 ? 'student' : 'teacher',
      }));

      usersAPI.getUsersAll.mockResolvedValue(mockUsers);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByText('Управление пользователями')).toBeInTheDocument();
      }, { timeout: 5000 });

      await waitFor(() => {
        const userRows = screen.getAllByTestId('user-row');
        expect(userRows).toHaveLength(25);
      }, { timeout: 5000 });

      mockUsers.slice(0, 3).forEach(user => {
        expect(screen.getByText(user.full_name)).toBeInTheDocument();
      });
    });
  });

  describe('T2: Filter by role=student', () => {
    it('should fetch and display only student users', async () => {
      const mockStudents = Array.from({ length: 20 }, (_, i) => ({
        id: i + 1,
        email: `student${i + 1}@test.com`,
        full_name: `Student ${i + 1}`,
        role: 'student',
      }));

      usersAPI.getUsersAll.mockResolvedValue(mockStudents);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByTestId('user-management')).toBeInTheDocument();
      }, { timeout: 5000 });

      const roleFilter = screen.getByTestId('role-filter');
      fireEvent.change(roleFilter, { target: { value: 'student' } });

      await waitFor(() => {
        const userRows = screen.getAllByTestId('user-row');
        expect(userRows).toHaveLength(20);
      }, { timeout: 5000 });

      expect(usersAPI.getUsersAll).toHaveBeenCalledWith({ role: 'student' });
      expect(screen.getByText('Student 1')).toBeInTheDocument();
    });
  });

  describe('T3: Filter by role=teacher', () => {
    it('should fetch and display only teacher users', async () => {
      const mockTeachers = Array.from({ length: 10 }, (_, i) => ({
        id: i + 1,
        email: `teacher${i + 1}@test.com`,
        full_name: `Teacher ${i + 1}`,
        role: 'teacher',
      }));

      usersAPI.getUsersAll.mockResolvedValue(mockTeachers);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByTestId('user-management')).toBeInTheDocument();
      }, { timeout: 5000 });

      const roleFilter = screen.getByTestId('role-filter');
      fireEvent.change(roleFilter, { target: { value: 'teacher' } });

      await waitFor(() => {
        const userRows = screen.getAllByTestId('user-row');
        expect(userRows).toHaveLength(10);
      }, { timeout: 5000 });

      expect(usersAPI.getUsersAll).toHaveBeenCalledWith({ role: 'teacher' });
      expect(screen.getByText('Teacher 1')).toBeInTheDocument();
    });
  });

  describe('T4: Edit button is available for users', () => {
    it('should display edit button for each user', async () => {
      const mockUsers = [
        {
          id: 1,
          email: 'john@test.com',
          full_name: 'John Doe',
          role: 'student',
        },
      ];

      usersAPI.getUsersAll.mockResolvedValue(mockUsers);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument();
      }, { timeout: 5000 });

      const editButton = screen.getByTestId('edit-user-button');
      expect(editButton).toBeInTheDocument();
      expect(editButton.textContent).toContain('Редактировать');
    });
  });

  describe('T5: Create button available in header', () => {
    it('should display create user button', async () => {
      usersAPI.getUsersAll.mockResolvedValue([]);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByTestId('user-management')).toBeInTheDocument();
      }, { timeout: 5000 });

      const createButton = screen.getByTestId('create-user-button');
      expect(createButton).toBeInTheDocument();
      expect(createButton.textContent).toContain('Добавить пользователя');
    });
  });

  describe('T6: Role filter dropdown available', () => {
    it('should display role filter dropdown', async () => {
      usersAPI.getUsersAll.mockResolvedValue([]);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByTestId('user-management')).toBeInTheDocument();
      }, { timeout: 5000 });

      const roleFilter = screen.getByTestId('role-filter');
      expect(roleFilter).toBeInTheDocument();
      expect(roleFilter).toHaveValue('');
    });
  });

  describe('T7: Table structure is correct', () => {
    it('should display table with correct columns', async () => {
      const mockUsers = [
        {
          id: 1,
          email: 'user1@test.com',
          full_name: 'User One',
          role: 'student',
        },
      ];

      usersAPI.getUsersAll.mockResolvedValue(mockUsers);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByText('Имя')).toBeInTheDocument();
      }, { timeout: 5000 });

      expect(screen.getByText('Email')).toBeInTheDocument();
      expect(screen.getByText('Роль')).toBeInTheDocument();
      expect(screen.getByText('Действия')).toBeInTheDocument();
    });
  });

  describe('T8: Delete button is available', () => {
    it('should have delete button for each user', async () => {
      const mockUsers = Array.from({ length: 3 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      usersAPI.getUsersAll.mockResolvedValue(mockUsers);

      renderUserManagement();

      await waitFor(() => {
        const deleteButtons = screen.getAllByTestId('delete-user-button');
        expect(deleteButtons).toHaveLength(3);
      }, { timeout: 5000 });
    });
  });

  describe('T9: Component mounts without errors', () => {
    it('should render component successfully', async () => {
      usersAPI.getUsersAll.mockResolvedValue([]);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByTestId('user-management')).toBeInTheDocument();
      }, { timeout: 5000 });

      expect(usersAPI.getUsersAll).toHaveBeenCalled();
    });
  });

  describe('T10: Empty users message displays correctly', () => {
    it('should display empty message when no users found', async () => {
      usersAPI.getUsersAll.mockResolvedValue([]);

      renderUserManagement();

      await waitFor(() => {
        expect(screen.getByText('Пользователи не найдены')).toBeInTheDocument();
      }, { timeout: 5000 });
    });
  });
});
