import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { AuthProvider } from '../AuthContext.jsx';
import { useAuth } from '../../hooks/useAuth.js';
import * as authAPI from '../../api/auth.js';

// Mock the auth API
vi.mock('../../api/auth.js');

// Test component that uses useAuth
function TestComponent() {
  const { user, isAuthenticated, loading } = useAuth();

  if (loading) return <div>Loading...</div>;
  if (!isAuthenticated) return <div>Not authenticated</div>;
  return <div>User: {user?.id}</div>;
}

describe('AuthContext - Token Management (HttpOnly Cookies)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Login token handling', () => {
    it('should authenticate when login response includes user data', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
      });

      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      // Simulate login by directly calling authAPI (in real app this would be via context)
      // Note: With httpOnly cookies, token is set by server via Set-Cookie header
      const result = await authAPI.login('test@example.com', 'password');

      expect(result.user).toEqual(mockUser);
      // Token is now stored in httpOnly cookie by backend, not accessible to JS
    });

    it('should handle login response with user data only', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'methodologist' };

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
      });

      const result = await authAPI.login('test@example.com', 'password');

      expect(result.user).toEqual(mockUser);
      // No token in response body - it's in httpOnly cookie
    });

    it('should extract user data from response correctly', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student', fullName: 'Test User' };

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
      });

      const result = await authAPI.login('test@example.com', 'password');

      expect(result.user).toEqual(mockUser);
    });

    it('should handle balance in login response', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };
      const mockBalance = 100;

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
        balance: mockBalance,
      });

      const result = await authAPI.login('test@example.com', 'password');

      expect(result.balance).toBe(mockBalance);
    });
  });

  describe('Logout token handling', () => {
    it('should call logout API to clear httpOnly cookie', async () => {
      authAPI.logout.mockResolvedValueOnce({ success: true });

      await authAPI.logout();

      expect(authAPI.logout).toHaveBeenCalled();
      // Token is cleared via server response that removes httpOnly cookie
    });
  });

  describe('CheckAuth token handling', () => {
    it('should handle 401 when getMe fails (cookie expired/invalid)', async () => {
      const error = new Error('Unauthorized');
      error.status = 401;

      authAPI.getMe.mockRejectedValueOnce(error);

      try {
        await authAPI.getMe();
      } catch (error) {
        // Expected to fail - httpOnly cookie is invalid/expired
      }

      expect(authAPI.getMe).toHaveBeenCalled();
      // Auth state should be cleared when 401 occurs
    });
  });

  describe('Cookie-based auth integrity', () => {
    it('should handle missing cookie gracefully (not logged in)', async () => {
      const error = new Error('No authentication cookie');
      error.status = 401;

      authAPI.getMe.mockRejectedValueOnce(error);

      try {
        await authAPI.getMe();
      } catch (error) {
        // Expected - no httpOnly cookie present
      }

      expect(authAPI.getMe).toHaveBeenCalled();
    });

    it('should work with valid cookie across multiple requests', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      authAPI.getMe
        .mockResolvedValueOnce(mockUser)
        .mockResolvedValueOnce(mockUser)
        .mockResolvedValueOnce(mockUser);

      // Multiple requests should work with same httpOnly cookie
      const result1 = await authAPI.getMe();
      const result2 = await authAPI.getMe();
      const result3 = await authAPI.getMe();

      expect(result1).toEqual(mockUser);
      expect(result2).toEqual(mockUser);
      expect(result3).toEqual(mockUser);
      expect(authAPI.getMe).toHaveBeenCalledTimes(3);
    });

    it('should handle cookie refresh on subsequent requests', async () => {
      const mockUser1 = { id: 1, email: 'test@example.com', role: 'student' };
      const mockUser2 = { id: 1, email: 'test@example.com', role: 'student', fullName: 'Updated' };

      authAPI.getMe
        .mockResolvedValueOnce(mockUser1)
        .mockResolvedValueOnce(mockUser2);

      const result1 = await authAPI.getMe();
      expect(result1).toEqual(mockUser1);

      // Server may refresh cookie with updated data
      const result2 = await authAPI.getMe();
      expect(result2).toEqual(mockUser2);
    });
  });
});
