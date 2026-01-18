/**
 * API эндпоинты для аутентификации
 */

import apiClient, { saveToken, clearToken, getStoredToken } from './client.js';

/**
 * Вход пользователя
 * @param {string} email - Email пользователя
 * @param {string} password - Пароль пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Данные пользователя
 */
export const login = async (email, password, options = {}) => {
  try {
    // Login might need more time due to session creation and race condition retries
    // Increased timeout to 15 seconds for better reliability
    return await apiClient.post('/auth/login', { email, password }, { timeout: 15000, ...options });
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error during login:', error);
    }
    throw error;
  }
};

/**
 * Выход текущего пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const logout = async (options = {}) => {
  try {
    return await apiClient.post('/auth/logout', {}, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error during logout:', error);
    }
    throw error;
  }
};

/**
 * Получить текущего аутентифицированного пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Данные пользователя
 */
export const getMe = async (options = {}) => {
  try {
    return await apiClient.get('/auth/me', options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching current user:', error);
    }
    throw error;
  }
};

/**
 * Запрос на сброс пароля
 * @param {string} email - Email пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>}
 */
export const requestPasswordReset = async (email, options = {}) => {
  try {
    return await apiClient.post('/auth/password-reset/request', { email }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error requesting password reset:', error);
    }
    throw error;
  }
};

/**
 * Сброс пароля с помощью токена
 * @param {string} token - Токен сброса
 * @param {string} newPassword - Новый пароль
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>}
 */
export const resetPassword = async (token, newPassword, options = {}) => {
  try {
    return await apiClient.post('/auth/password-reset/confirm', {
      token,
      new_password: newPassword,
    }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error resetting password:', error);
    }
    throw error;
  }
};

/**
 * Регистрация нового пользователя через Telegram username
 * @param {string} telegramUsername - Telegram username без @
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Данные созданного пользователя
 */
export const registerViaTelegram = async (telegramUsername, options = {}) => {
  try {
    return await apiClient.post('/auth/register-telegram', { telegram_username: telegramUsername }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error registering via Telegram:', error);
    }
    throw error;
  }
};

/**
 * Смена пароля текущего пользователя
 * @param {string} oldPassword - Текущий пароль
 * @param {string} newPassword - Новый пароль
 * @param {Object} [options] - Опции запроса (включая signal для отмены, timeout)
 * @returns {Promise<Object>} Ответ сервера
 */
export const changePassword = async (oldPassword, newPassword, options = {}) => {
  try {
    return await apiClient.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      let errorMessage = 'Error changing password:';

      if (error.status === 401) {
        errorMessage = 'Incorrect current password';
      } else if (error.status === 400) {
        errorMessage = 'Invalid password: weak password or same as old password';
      } else if (error.status === 500) {
        errorMessage = 'Server error during password change';
      }

      console.error(errorMessage, error.message);
    }
    throw error;
  }
};

// DEPRECATED: saveToken, clearToken, getStoredToken are no-ops now
// Kept for backward compatibility only - auth uses httpOnly cookies
export default {
  login,
  logout,
  getMe,
  requestPasswordReset,
  resetPassword,
  registerViaTelegram,
  changePassword,
  saveToken,
  clearToken,
  getStoredToken,
};
