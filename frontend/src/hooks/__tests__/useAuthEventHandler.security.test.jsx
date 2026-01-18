/**
 * Security tests for useAuthEventHandler
 * Verifies protection against Open Redirect vulnerabilities
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { useAuthEventHandler } from '../useAuthEventHandler.js';
import { AuthProvider } from '../../context/AuthContext.jsx';

// Mock useNavigate and useLocation
const mockNavigate = vi.fn();
const mockLocation = { pathname: '/dashboard' };

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useLocation: () => mockLocation,
  };
});

describe('useAuthEventHandler Security', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  const wrapper = ({ children }) => (
    <BrowserRouter>
      <AuthProvider>{children}</AuthProvider>
    </BrowserRouter>
  );

  it('should ALWAYS navigate to /login (hardcoded)', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    // Dispatch auth:unauthorized event
    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: { url: '/api/test', method: 'GET', status: 401 },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/login',
        expect.objectContaining({ replace: true })
      );
    });

    // Verify it's the hardcoded /login, not from event
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[0]).toBe('/login');
  });

  it('should NOT use redirect URL from event detail', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    // Dispatch event with malicious redirect
    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: {
          url: '/api/test',
          method: 'GET',
          status: 401,
          redirect: 'https://evil.com/phishing',
          location: 'javascript:alert(1)',
        },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // Verify navigation is to hardcoded /login
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[0]).toBe('/login');
    expect(navigateCall[0]).not.toContain('evil.com');
    expect(navigateCall[0]).not.toContain('javascript:');
  });

  it('should only preserve CURRENT location.pathname in state', async () => {
    mockLocation.pathname = '/dashboard/student/bookings';
    renderHook(() => useAuthEventHandler(), { wrapper });

    // Dispatch event
    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: { url: '/api/test', method: 'GET', status: 401 },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // Verify state contains current pathname (safe)
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[1].state.from).toBe('/dashboard/student/bookings');
  });

  it('should NOT use any external data in navigation', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    // Dispatch event with various malicious fields
    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: {
          url: '/api/test',
          method: 'GET',
          status: 401,
          redirect_url: 'https://evil.com',
          next: '/admin',
          goto: 'javascript:alert(1)',
        },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // Verify no external data used
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[0]).toBe('/login');
    expect(JSON.stringify(navigateCall)).not.toContain('evil.com');
    expect(JSON.stringify(navigateCall)).not.toContain('/admin');
    expect(JSON.stringify(navigateCall)).not.toContain('javascript:');
  });

  it('should validate navigation path is same-origin (React Router)', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: { url: '/api/test', method: 'GET', status: 401 },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // React Router navigate() only accepts relative paths
    // This is a security feature - absolute URLs would require window.location
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[0]).toBe('/login');
    expect(navigateCall[0]).not.toMatch(/^https?:\/\//);
  });

  it('should use React Router replace mode to prevent back navigation', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: { url: '/api/test', method: 'GET', status: 401 },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // Verify replace: true (security: prevents going back to authenticated page)
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[1].replace).toBe(true);
  });

  it('should NOT execute if event detail contains XSS payload', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    // Dispatch event with XSS attempt
    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: {
          url: '/api/test',
          method: 'GET',
          status: 401,
          xss: '<script>alert(document.cookie)</script>',
          injection: '"; alert(1); //',
        },
      })
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // Verify safe navigation
    const navigateCall = mockNavigate.mock.calls[0];
    expect(navigateCall[0]).toBe('/login');
    expect(JSON.stringify(navigateCall)).not.toContain('<script>');
    expect(JSON.stringify(navigateCall)).not.toContain('alert(');
  });

  it('should ignore malformed event objects', async () => {
    renderHook(() => useAuthEventHandler(), { wrapper });

    // Dispatch malformed event
    window.dispatchEvent(
      new CustomEvent('auth:unauthorized', {
        detail: null, // Malformed
      })
    );

    // Should not throw error
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/login',
        expect.objectContaining({ replace: true })
      );
    });
  });
});
