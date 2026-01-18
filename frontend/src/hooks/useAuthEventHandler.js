/**
 * Hook to handle auth events (401 unauthorized)
 * Listens for custom 'auth:unauthorized' events from API client
 * and handles logout + navigation using React Router
 */
import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from './useAuth.js';
import { logger } from '../utils/logger.js';

export const useAuthEventHandler = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout, isAuthenticated } = useAuth();

  useEffect(() => {
    const handleUnauthorized = async (event) => {
      logger.warn('Received auth:unauthorized event', {
        detail: event.detail,
        currentPath: location.pathname,
        isAuthenticated,
      });

      // Clear authentication state
      try {
        await logout();
      } catch (error) {
        logger.error('Error during logout on 401:', error);
      }

      // Public routes that should not redirect to login
      const publicRoutes = ['/', '/login', '/telegram-register'];
      const isPublicRoute = publicRoutes.some(
        (route) =>
          location.pathname === route ||
          (route === '/telegram-register' &&
            location.pathname.startsWith('/telegram-register'))
      );

      // Navigate to login only if not on a public route
      if (!isPublicRoute) {
        logger.debug('Navigating to login page', {
          from: location.pathname,
        });
        navigate('/login', {
          replace: true,
          state: { from: location.pathname },
        });
      }
    };

    // Add event listener
    if (typeof window !== 'undefined') {
      window.addEventListener('auth:unauthorized', handleUnauthorized);
    }

    // Cleanup
    return () => {
      if (typeof window !== 'undefined') {
        window.removeEventListener('auth:unauthorized', handleUnauthorized);
      }
    };
  }, [navigate, location, logout, isAuthenticated]);
};

export default useAuthEventHandler;
