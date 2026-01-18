/**
 * API эндпоинты для заявок на пробные занятия
 */

import apiClient from './client.js';

/**
 * Получить все заявки на пробные занятия (только администратор)
 * @param {Object} filters - Параметры фильтрации
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список заявок
 */
export const getTrialRequests = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  if (filters.status) params.append('status', filters.status);
  if (filters.search) params.append('search', filters.search);
  if (filters.start_date) params.append('start_date', filters.start_date);
  if (filters.end_date) params.append('end_date', filters.end_date);

  const query = params.toString();

  try {
    const response = await apiClient.get(`/trial-requests${query ? `?${query}` : ''}`, options);
    // Backend возвращает {trial_requests} объект или массив напрямую
    if (Array.isArray(response)) {
      return response;
    }
    return response?.trial_requests || response?.data || [];
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching trial requests:', error);
    }
    throw error;
  }
};

/**
 * Получить заявку по ID (только администратор)
 * @param {number} requestId - ID заявки
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Детали заявки
 */
export const getTrialRequestById = async (requestId, options = {}) => {
  try {
    return await apiClient.get(`/trial-requests/${requestId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching trial request:', error);
    }
    throw error;
  }
};

/**
 * Обновить статус заявки (только администратор)
 * @param {number} requestId - ID заявки
 * @param {string} status - Новый статус (pending, contacted, scheduled, completed)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Обновлённая заявка
 */
export const updateTrialRequest = async (requestId, status, options = {}) => {
  try {
    return await apiClient.patch(`/trial-requests/${requestId}`, { status }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error updating trial request:', error);
    }
    throw error;
  }
};

/**
 * Удалить заявку (только администратор)
 * @param {number} requestId - ID заявки
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const deleteTrialRequest = async (requestId, options = {}) => {
  try {
    return await apiClient.delete(`/trial-requests/${requestId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error deleting trial request:', error);
    }
    throw error;
  }
};

export default {
  getTrialRequests,
  getTrialRequestById,
  updateTrialRequest,
  deleteTrialRequest,
};
