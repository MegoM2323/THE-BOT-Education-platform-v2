/**
 * API endpoints для платежей
 */

import apiClient from './client.js';

/**
 * Создать платеж для покупки кредитов
 * @param {number} credits - Количество кредитов для покупки
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Данные платежа с confirmation_url
 */
export const createPayment = async (credits, options = {}) => {
  try {
    const response = await apiClient.post('/payments/create', { credits }, options);
    return response;
  } catch (error) {
    // Не логируем AbortError - это нормальное поведение при отмене запроса
    if (error.name !== 'AbortError') {
      console.error('Error creating payment:', error);
    }
    throw error;
  }
};

/**
 * Получить историю платежей текущего пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список платежей
 */
export const getPaymentHistory = async (options = {}) => {
  try {
    const response = await apiClient.get('/payments/history', options);
    // Обработка различных форматов ответа от backend
    if (Array.isArray(response)) {
      return response;
    }
    return response?.payments || response?.data || [];
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching payment history:', error);
    }
    throw error;
  }
};

export default {
  createPayment,
  getPaymentHistory,
};
