/**
 * API эндпоинты для работы с Telegram интеграцией
 */

import apiClient from './client.js';

/**
 * Генерация токена для привязки Telegram
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Объект с токеном и именем бота
 * @property {string} token - Уникальный токен для привязки
 * @property {string} bot_username - Имя пользователя Telegram бота
 * @throws {Error} Ошибка при генерации токена
 */
export const generateLinkToken = async (options = {}) => {
  try {
    const response = await apiClient.get('/telegram/link-token', options);
    if (!response.token || !response.bot_username) {
      throw new Error('Некорректный ответ от сервера: отсутствуют обязательные поля');
    }
    return response;
  } catch (error) {
    // FIX T003: Preserve AbortError to avoid showing "Request cancelled" to user on unmount
    if (error.name === 'AbortError') {
      throw error;
    }
    const statusCode = error?.response?.status || 'unknown';
    const message = error?.response?.data?.message || error?.response?.data?.error || error?.message || 'Ошибка при генерации токена';
    console.error(`Telegram API error (${statusCode}):`, message);
    throw new Error(message);
  }
};

/**
 * Получение статуса привязки Telegram текущего пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Статус привязки
 * @property {boolean} linked - Привязан ли Telegram
 * @property {Object} [telegram] - Данные Telegram (если привязан)
 * @property {string} [telegram.telegram_id] - ID пользователя в Telegram
 * @property {string} [telegram.username] - Username в Telegram
 * @property {string} [telegram.first_name] - Имя в Telegram
 * @property {string} [telegram.last_name] - Фамилия в Telegram
 * @property {string} [telegram.linked_at] - Дата привязки
 * @throws {Error} Ошибка при получении статуса
 */
export const getMyTelegramLink = async (options = {}) => {
  try {
    const response = await apiClient.get('/telegram/me', options);
    return response;
  } catch (error) {
    // FIX T003: Preserve AbortError to avoid showing "Request cancelled" to user on unmount
    if (error.name === 'AbortError') {
      throw error;
    }
    const statusCode = error?.response?.status || 'unknown';
    const message = error?.response?.data?.message || error?.response?.data?.error || error?.message || 'Ошибка при получении статуса привязки';
    console.error(`Telegram API error (${statusCode}):`, message);
    throw new Error(message);
  }
};

/**
 * Отвязка Telegram от аккаунта
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Сообщение об успехе
 * @property {string} message - Сообщение об успешной отвязке
 * @throws {Error} Ошибка при отвязке
 */
export const unlinkTelegram = async (options = {}) => {
  try {
    const response = await apiClient.delete('/telegram/link', options);
    return response;
  } catch (error) {
    // FIX T003: Preserve AbortError to avoid showing "Request cancelled" to user on unmount
    if (error.name === 'AbortError') {
      throw error;
    }
    const statusCode = error?.response?.status || 'unknown';
    const message = error?.response?.data?.message || error?.response?.data?.error || error?.message || 'Ошибка при отвязке Telegram';
    console.error(`Telegram API error (${statusCode}):`, message);
    throw new Error(message);
  }
};

/**
 * Подписка на Telegram уведомления (альтернативный способ)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Сообщение об успехе
 * @throws {Error} Ошибка при подписке
 */
export const subscribeToNotifications = async (options = {}) => {
  try {
    const response = await apiClient.post('/telegram/subscribe', {}, options);
    return response;
  } catch (error) {
    // FIX T003: Preserve AbortError to avoid showing "Request cancelled" to user on unmount
    if (error.name === 'AbortError') {
      throw error;
    }
    const statusCode = error?.response?.status || 'unknown';
    const message = error?.response?.data?.message || error?.response?.data?.error || error?.message || 'Ошибка при подписке на уведомления';
    console.error(`Telegram API error (${statusCode}):`, message);
    throw new Error(message);
  }
};

/**
 * Отписка от Telegram уведомлений (альтернативный способ)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Сообщение об успехе
 * @throws {Error} Ошибка при отписке
 */
export const unsubscribeFromNotifications = async (options = {}) => {
  try {
    const response = await apiClient.post('/telegram/unsubscribe', {}, options);
    return response;
  } catch (error) {
    // FIX T003: Preserve AbortError to avoid showing "Request cancelled" to user on unmount
    if (error.name === 'AbortError') {
      throw error;
    }
    const statusCode = error?.response?.status || 'unknown';
    const message = error?.response?.data?.message || error?.response?.data?.error || error?.message || 'Ошибка при отписке от уведомлений';
    console.error(`Telegram API error (${statusCode}):`, message);
    throw new Error(message);
  }
};

export default {
  generateLinkToken,
  getMyTelegramLink,
  unlinkTelegram,
  subscribeToNotifications,
  unsubscribeFromNotifications,
};
