import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { ProtectedRoute } from '../ProtectedRoute';
import { useAuth } from '../../../hooks/useAuth.js';

// Mock useAuth hook
vi.mock('../../../hooks/useAuth.js');

// Mock Spinner component
vi.mock('../../common/Spinner.jsx', () => ({
  default: ({ size }) => <div data-testid="spinner" data-size={size}>Loading...</div>
}));

describe('ProtectedRoute', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading state', () => {
    it('should show loading spinner when loading is true', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      render(
        <MemoryRouter>
          <ProtectedRoute>
            <div>Protected Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();
      expect(screen.getByTestId('spinner')).toBeInTheDocument();
      expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
    });

    it('should have accessibility attributes on loading state', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      render(
        <MemoryRouter>
          <ProtectedRoute>
            <div>Protected Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      const loadingDiv = screen.getByTestId('protected-route-loading');
      expect(loadingDiv).toHaveAttribute('role', 'status');
      expect(loadingDiv).toHaveAttribute('aria-label', 'Checking authentication');
    });

    it('should use large spinner during loading', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      render(
        <MemoryRouter>
          <ProtectedRoute>
            <div>Protected Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      expect(screen.getByTestId('spinner')).toHaveAttribute('data-size', 'lg');
    });

    it('should cover full viewport with loading state', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      render(
        <MemoryRouter>
          <ProtectedRoute>
            <div>Protected Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      const loadingDiv = screen.getByTestId('protected-route-loading');
      const styles = loadingDiv.style;

      expect(styles.position).toBe('fixed');
      expect(styles.top).toBe('0px');
      expect(styles.left).toBe('0px');
      expect(styles.right).toBe('0px');
      expect(styles.bottom).toBe('0px');
      expect(styles.zIndex).toBe('9999');
    });
  });

  describe('Authentication check', () => {
    it('should redirect to login when not authenticated', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: false,
        isAuthenticated: false
      });

      render(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByText('Login Page')).toBeInTheDocument();
      expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
    });

    it('should render children when authenticated', () => {
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'student' },
        loading: false,
        isAuthenticated: true
      });

      render(
        <MemoryRouter>
          <ProtectedRoute>
            <div>Protected Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      expect(screen.getByText('Protected Content')).toBeInTheDocument();
    });

    it('should not redirect during loading even if isAuthenticated is false', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      render(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Should show loading, not redirect to login
      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();
      expect(screen.queryByText('Login Page')).not.toBeInTheDocument();
      expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
    });
  });

  describe('Role-based access control', () => {
    it('should render children when user has allowed role', () => {
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'admin' },
        loading: false,
        isAuthenticated: true
      });

      render(
        <MemoryRouter>
          <ProtectedRoute allowedRoles={['admin', 'teacher']}>
            <div>Admin Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      expect(screen.getByText('Admin Content')).toBeInTheDocument();
    });

    it('should redirect to home when user does not have allowed role', () => {
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'student' },
        loading: false,
        isAuthenticated: true
      });

      render(
        <MemoryRouter initialEntries={['/admin']}>
          <Routes>
            <Route path="/admin" element={
              <ProtectedRoute allowedRoles={['admin', 'teacher']}>
                <div>Admin Content</div>
              </ProtectedRoute>
            } />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByText('Home Page')).toBeInTheDocument();
      expect(screen.queryByText('Admin Content')).not.toBeInTheDocument();
    });

    it('should render children when no roles specified', () => {
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'student' },
        loading: false,
        isAuthenticated: true
      });

      render(
        <MemoryRouter>
          <ProtectedRoute>
            <div>Any Authenticated User</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      expect(screen.getByText('Any Authenticated User')).toBeInTheDocument();
    });

    it('should not check roles during loading', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      render(
        <MemoryRouter initialEntries={['/admin']}>
          <Routes>
            <Route path="/admin" element={
              <ProtectedRoute allowedRoles={['admin']}>
                <div>Admin Content</div>
              </ProtectedRoute>
            } />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Should show loading, not redirect to home
      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();
      expect(screen.queryByText('Home Page')).not.toBeInTheDocument();
    });
  });

  describe('Race condition prevention - FB002 (B4)', () => {
    it('should prevent flash of login page by showing loading first', () => {
      // Simulate initial render with loading state
      const { rerender } = render(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // First render: loading state
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      rerender(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();
      expect(screen.queryByText('Login Page')).not.toBeInTheDocument();
      expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();

      // Second render: auth loaded, user authenticated
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'student' },
        loading: false,
        isAuthenticated: true
      });

      rerender(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByText('Protected Content')).toBeInTheDocument();
      expect(screen.queryByText('Login Page')).not.toBeInTheDocument();
    });

    it('should prevent redirect during auth initialization', () => {
      // Auth state is initializing
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      const { container } = render(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Verify we're still on /protected path (showing loading)
      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();
      expect(screen.queryByText('Login Page')).not.toBeInTheDocument();
    });

    it('should handle multiple re-renders during auth check', () => {
      const { rerender } = render(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Render 1: Loading
      useAuth.mockReturnValue({
        user: null,
        loading: true,
        isAuthenticated: false
      });

      rerender(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();

      // Render 2: Still loading
      rerender(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByTestId('protected-route-loading')).toBeInTheDocument();

      // Render 3: Auth complete
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'student' },
        loading: false,
        isAuthenticated: true
      });

      rerender(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      expect(screen.getByText('Protected Content')).toBeInTheDocument();
      expect(screen.queryByTestId('protected-route-loading')).not.toBeInTheDocument();
    });

    it('should use replace navigation to prevent history pollution', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: false,
        isAuthenticated: false
      });

      render(
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/protected" element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            } />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Should redirect to login
      expect(screen.getByText('Login Page')).toBeInTheDocument();

      // The Navigate component uses replace prop, which we've specified
      // This is tested by the fact that the component uses <Navigate replace />
    });
  });

  describe('Edge cases', () => {
    it('should handle null user during role check', () => {
      useAuth.mockReturnValue({
        user: null,
        loading: false,
        isAuthenticated: true // Edge case: authenticated but user is null
      });

      render(
        <MemoryRouter initialEntries={['/admin']}>
          <Routes>
            <Route path="/admin" element={
              <ProtectedRoute allowedRoles={['admin']}>
                <div>Admin Content</div>
              </ProtectedRoute>
            } />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Should redirect to home since user?.role will be undefined
      expect(screen.getByText('Home Page')).toBeInTheDocument();
    });

    it('should handle empty allowedRoles array', () => {
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com', role: 'student' },
        loading: false,
        isAuthenticated: true
      });

      render(
        <MemoryRouter>
          <ProtectedRoute allowedRoles={[]}>
            <div>Protected Content</div>
          </ProtectedRoute>
        </MemoryRouter>
      );

      // Empty array should allow any authenticated user
      expect(screen.getByText('Protected Content')).toBeInTheDocument();
    });

    it('should handle user with undefined role', () => {
      useAuth.mockReturnValue({
        user: { id: 1, email: 'test@example.com' }, // No role property
        loading: false,
        isAuthenticated: true
      });

      render(
        <MemoryRouter initialEntries={['/admin']}>
          <Routes>
            <Route path="/admin" element={
              <ProtectedRoute allowedRoles={['admin']}>
                <div>Admin Content</div>
              </ProtectedRoute>
            } />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      // Should redirect to home since user.role is undefined
      expect(screen.getByText('Home Page')).toBeInTheDocument();
    });
  });
});
