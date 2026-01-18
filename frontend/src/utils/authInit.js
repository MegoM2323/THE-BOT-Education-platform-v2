import { logger } from '../utils/logger.js';

/**
 * DEPRECATED: Auth initialization is no longer needed
 * Authentication now uses httpOnly cookies managed by the backend
 * Token storage in localStorage removed for XSS protection
 */

/**
 * Инициализирует состояние аутентификации при загрузке приложения
 * DEPRECATED: Auth state is now managed via httpOnly cookies (backend)
 * This function kept for backward compatibility but does nothing
 */
export const initializeAuth = () => {
  logger.debug('initializeAuth called - no action needed (using httpOnly cookies)');
  // Auth state automatically restored via session cookie on backend
  // No client-side token management needed
  return false; // Always return false as there's no localStorage token
};

export default initializeAuth;
