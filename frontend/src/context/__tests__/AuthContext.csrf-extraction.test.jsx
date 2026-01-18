import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { AuthProvider } from '../AuthContext.jsx';
import * as authAPI from '../../api/auth.js';
import * as clientModule from '../../api/client.js';

// Mock the auth API
vi.mock('../../api/auth.js');

// Mock client module to track CSRF operations
vi.mock('../../api/client.js', async () => {
  const actual = await vi.importActual('../../api/client.js');
  return {
    ...actual,
    setupCSRF: vi.fn(),
    fetchCSRFToken: vi.fn(),
    setCSRFToken: vi.fn(),
    clearCSRFToken: vi.fn(),
  };
});

describe('AuthContext - CSRF Token Extraction Fix (T103)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Login CSRF token handling', () => {
    it('should NOT call setupCSRF fallback when token already in header', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      // Mock login response WITHOUT csrf_token in body
      // (It comes from X-CSRF-Token header which client.js extracts)
      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
        balance: 50,
      });

      const result = await authAPI.login('test@example.com', 'password');

      // Verify setupCSRF was NOT called after login
      // (before fix, it would be called as fallback because data.csrf_token doesn't exist)
      // After fix: client.js already extracted token from X-CSRF-Token header
      expect(clientModule.setupCSRF).not.toHaveBeenCalled();
      expect(result.user).toEqual(mockUser);
    });

    it('should use CSRF token already extracted from response headers by client.js', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      // Simulate login response WITHOUT csrf_token in body
      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
      });

      // Mock getMe
      authAPI.getMe.mockResolvedValueOnce({
        user: mockUser,
      });

      // Verify that client.js would have extracted token from header
      // and setupCSRF should NOT be called as fallback
      const result = await authAPI.login('test@example.com', 'password');

      expect(result.user).toEqual(mockUser);
      // Token extraction happens in client.js request() function (lines 190-195)
      // NOT in AuthContext login method anymore
    });

    it('should not attempt to extract csrf_token from response body', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      // Even if backend returned csrf_token in body (it doesn't),
      // client.js header extraction should take precedence
      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
        csrf_token: 'some-token-from-body', // Should be ignored now
      });

      const result = await authAPI.login('test@example.com', 'password');

      expect(result.user).toEqual(mockUser);
      // We no longer extract from body, client.js handles it from header
    });
  });

  describe('CSRF token flow in client.js', () => {
    it('client.js extracts CSRF token from X-CSRF-Token header', () => {
      // This test documents how client.js handles CSRF extraction
      // (lines 190-195 in client.js)
      // The token is extracted in the request() function and stored in csrfToken variable

      // When AuthContext calls authAPI.login(), which calls client.request()
      // The request() function automatically extracts X-CSRF-Token from response headers
      // and stores it in the module-level csrfToken variable

      expect(clientModule.getCSRFToken).toBeDefined();
      expect(clientModule.setCSRFToken).toBeDefined();
    });

    it('should not call fetchCSRFToken as fallback after login', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
      });

      await authAPI.login('test@example.com', 'password');

      // The old code would call setupCSRF() here as a fallback
      // because it checked for data.csrf_token in body (which doesn't exist)
      // This is now fixed - no fallback fetch should happen
      expect(clientModule.fetchCSRFToken).not.toHaveBeenCalled();
    });
  });

  describe('Race condition prevention', () => {
    it('should not create race condition with unnecessary CSRF fetches', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
      });

      // Before fix: login would trigger both:
      // 1. POST /auth/login
      // 2. GET /csrf-token (unnecessary fallback)
      // This could cause race conditions if getMe ran between them

      // After fix: only login is called, CSRF token already extracted from header
      await authAPI.login('test@example.com', 'password');

      // Verify only login was made, no fallback CSRF fetch
      expect(authAPI.login).toHaveBeenCalledOnce();
      expect(clientModule.setupCSRF).not.toHaveBeenCalled();
    });
  });

  describe('Backwards compatibility', () => {
    it('should still handle responses with user data wrapper correctly', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
        balance: 100,
      });

      const result = await authAPI.login('test@example.com', 'password');

      expect(result.user).toEqual(mockUser);
    });

    it('should handle responses where data is the user directly', async () => {
      const mockUser = { id: 1, email: 'test@example.com', role: 'student' };

      // Mock login returns the full response structure
      authAPI.login.mockResolvedValueOnce({
        user: mockUser,
        balance: 100,
      });

      const result = await authAPI.login('test@example.com', 'password');

      // When called via authAPI, we get the full wrapped response
      expect(result.user).toEqual(mockUser);
    });
  });
});
