import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useAuthEventHandler } from '../useAuthEventHandler.js';
import * as useAuthModule from '../useAuth.js';

const mockNavigate = vi.fn();
let mockLogout;

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock('../useAuth.js');

describe('useAuthEventHandler', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
    mockLogout = vi.fn().mockResolvedValue(undefined);

    vi.mocked(useAuthModule.useAuth).mockReturnValue({
      user: { id: 1, email: 'test@test.com', role: 'student' },
      isAuthenticated: true,
      logout: mockLogout,
      login: vi.fn(),
      loading: false,
      balance: 10,
      updateUser: vi.fn(),
      refreshUser: vi.fn(),
      checkAuth: vi.fn(),
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const createWrapper = (initialEntries = ['/dashboard']) => {
    return ({ children }) => (
      <MemoryRouter initialEntries={initialEntries}>
        {children}
      </MemoryRouter>
    );
  };

  it('should listen for auth:unauthorized events', async () => {
    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/test', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalled();
    });

    unmount();
  });

  it('should call logout on auth:unauthorized event', async () => {
    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/users/me', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalledTimes(1);
    });

    unmount();
  });

  it('should navigate to /login after logout', async () => {
    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(['/dashboard']),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/test', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/login',
        expect.objectContaining({
          replace: true,
        })
      );
    });

    unmount();
  });

  it('should NOT navigate if already on /login', async () => {
    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(['/login']),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/test', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalled();
    });

    expect(mockNavigate).not.toHaveBeenCalled();

    unmount();
  });

  it('should preserve current location for redirect after login', async () => {
    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(['/dashboard/student/bookings']),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/bookings', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/login',
        expect.objectContaining({
          state: { from: '/dashboard/student/bookings' },
        })
      );
    });

    unmount();
  });

  it('should handle logout errors gracefully', async () => {
    mockLogout.mockRejectedValueOnce(new Error('Logout failed'));

    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(['/dashboard']),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/test', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/login',
        expect.any(Object)
      );
    });

    unmount();
  });

  it('should cleanup event listener on unmount', () => {
    const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener');

    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(),
    });

    unmount();

    expect(removeEventListenerSpy).toHaveBeenCalledWith(
      'auth:unauthorized',
      expect.any(Function)
    );

    removeEventListenerSpy.mockRestore();
  });

  it('should handle multiple events correctly', async () => {
    const { unmount } = renderHook(() => useAuthEventHandler(), {
      wrapper: createWrapper(['/dashboard']),
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/test1', method: 'GET', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalledTimes(1);
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { url: '/api/test2', method: 'POST', status: 401 },
        })
      );
    });

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalledTimes(2);
    });

    unmount();
  });
});
