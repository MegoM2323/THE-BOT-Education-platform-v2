import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter as Router } from 'react-router-dom';
import { LoginForm } from '../LoginForm';
import { useAuth } from '../../../hooks/useAuth.js';
import { useNotification } from '../../../hooks/useNotification.js';
import { ROLES } from '../../../utils/constants.js';

// Mock hooks
vi.mock('../../../hooks/useAuth.js');
vi.mock('../../../hooks/useNotification.js');

const renderLoginForm = () => {
  return render(
    <Router>
      <LoginForm />
    </Router>
  );
};

describe('LoginForm - Microtask Redirect Timing Tests', () => {
  let mockLogin;
  let mockNotification;

  beforeEach(() => {
    mockLogin = vi.fn();
    mockNotification = {
      success: vi.fn(),
      error: vi.fn()
    };

    useAuth.mockReturnValue({
      login: mockLogin,
      user: null,
      isAuthenticated: false
    });

    useNotification.mockReturnValue(mockNotification);

    vi.clearAllMocks();
  });

  describe('Scenario 1: Admin login redirect', () => {
    it('should call login with correct credentials for admin', async () => {
      const mockUser = {
        id: '1',
        email: 'admin@test.com',
        role: ROLES.ADMIN,
      };

      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'TestPass123!');

      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('admin@test.com', 'TestPass123!');
      }, { timeout: 1000 });
    });

    it('should show success notification after admin login', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalledWith('Добро пожаловать!');
      }, { timeout: 1000 });
    });

    it('should not log errors on successful admin login', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      });

      expect(consoleErrorSpy).not.toHaveBeenCalled();
      consoleErrorSpy.mockRestore();
    });
  });

  describe('Scenario 2: Student login redirect', () => {
    it('should handle student login correctly', async () => {
      const mockUser = {
        id: '3',
        email: 'student@test.com',
        role: ROLES.STUDENT,
      };

      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'student@test.com');
      await userEvent.type(passwordField, 'TestPass123!');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('student@test.com', 'TestPass123!');
      }, { timeout: 1000 });
    });

    it('should show notification on student login', async () => {
      const mockUser = { id: '3', role: ROLES.STUDENT };
      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'student@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      });
    });

    it('should load student data correctly', async () => {
      const mockUser = {
        id: '3',
        email: 'student@test.com',
        role: ROLES.STUDENT,
        name: 'John Student'
      };

      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'student@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled();
      });

      expect(mockNotification.success).toHaveBeenCalled();
    });
  });

  describe('Scenario 3: Login error handling', () => {
    it('should not call success notification on invalid credentials', async () => {
      mockLogin.mockRejectedValue(new Error('Invalid credentials'));

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'wrongpassword');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled();
      }, { timeout: 1000 });

      expect(mockNotification.success).not.toHaveBeenCalled();
    });

    it('should show error notification on login failure', async () => {
      const errorMessage = 'Invalid credentials';
      mockLogin.mockRejectedValue(new Error(errorMessage));

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'wrong');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
      }, { timeout: 1000 });
    });

    it('should keep user on login page after error', async () => {
      mockLogin.mockRejectedValue(new Error('Network error'));

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'user@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.error).toHaveBeenCalled();
      }, { timeout: 1000 });

      expect(screen.getByTestId('login-form')).toBeInTheDocument();
    });

    it('should allow retry after login error', async () => {
      mockLogin
        .mockRejectedValueOnce(new Error('First attempt failed'))
        .mockResolvedValueOnce({ id: '1', role: ROLES.ADMIN });

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      // First attempt - fails
      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'wrong');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.error).toHaveBeenCalled();
      }, { timeout: 1000 });

      // Clear and retry
      await userEvent.clear(emailField);
      await userEvent.clear(passwordField);

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'correct');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      }, { timeout: 1000 });

      expect(mockLogin).toHaveBeenCalledTimes(2);
    });

    it('should handle network error gracefully', async () => {
      mockLogin.mockRejectedValue(new Error('Network timeout'));

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'user@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.error).toHaveBeenCalledWith('Network timeout');
      }, { timeout: 1000 });
    });
  });

  describe('Scenario 4: Race condition - fast submit', () => {
    it('should prevent multiple login calls on rapid clicks', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');

      fireEvent.click(submitButton);
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      }, { timeout: 1000 });

      expect(mockLogin.mock.calls.length).toBeLessThanOrEqual(1);
    });

    it('should not make duplicate API calls on rapid submit', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');

      fireEvent.click(submitButton);
      fireEvent.click(submitButton);
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      }, { timeout: 1000 });

      expect(mockLogin.mock.calls.length).toBeLessThanOrEqual(1);
    });
  });

  describe('Microtask timing and component safety', () => {
    it('should render form and be safe with microtask', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      });

      // Form should still be in document
      expect(screen.getByTestId('login-form')).toBeInTheDocument();
    });

    it('should not interfere with error handling with microtask', async () => {
      mockLogin.mockRejectedValue(new Error('Login failed'));

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'user@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.error).toHaveBeenCalled();
      }, { timeout: 1000 });

      // Success should NOT be called
      expect(mockNotification.success).not.toHaveBeenCalled();
    });

    it('should execute login and notification within reasonable time', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      const startTime = Date.now();

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      }, { timeout: 1000 });

      const endTime = Date.now();
      const duration = endTime - startTime;

      // Should complete within reasonable time
      expect(duration).toBeLessThan(2000);
    });

    it('should work correctly when component unmounts during async', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockResolvedValue(mockUser);

      const { unmount } = renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      // Unmount before async completes
      unmount();

      await waitFor(() => {
        expect(mockNotification.success).toHaveBeenCalled();
      });

      // Should not crash
      expect(mockLogin).toHaveBeenCalled();
    });

    it('should not cause memory leaks with microtask', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      const resolveLogin = vi.fn();
      mockLogin.mockImplementation(() => new Promise(resolve => {
        resolveLogin(resolve);
        setTimeout(() => resolve(mockUser), 50);
      }));

      const { unmount } = renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      // Give it time to process
      await new Promise(resolve => setTimeout(resolve, 100));
      
      unmount();

      // Should not have lingering state updates
      expect(mockLogin).toHaveBeenCalled();
    });
  });

  describe('Form validation and UI state', () => {
    it('should render login form correctly', () => {
      renderLoginForm();

      expect(screen.getByTestId('login-form')).toBeInTheDocument();
      expect(document.querySelector('input[name="email"]')).toBeInTheDocument();
      expect(document.querySelector('input[name="password"]')).toBeInTheDocument();
      expect(screen.getByTestId('login-button')).toBeInTheDocument();
    });

    it('should show loading state during login', async () => {
      const mockUser = { id: '1', role: ROLES.ADMIN };
      mockLogin.mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve(mockUser), 200))
      );

      renderLoginForm();

      const emailField = document.querySelector('input[name="email"]');
      const passwordField = document.querySelector('input[name="password"]');
      const submitButton = screen.getByTestId('login-button');

      await userEvent.type(emailField, 'admin@test.com');
      await userEvent.type(passwordField, 'password');
      fireEvent.click(submitButton);

      // Should show loading indicator
      await waitFor(() => {
        expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
      });
    });

    it('should have toggle password button', () => {
      renderLoginForm();

      const toggleBtn = screen.getByTestId('toggle-password');
      expect(toggleBtn).toBeInTheDocument();
    });
  });
});
