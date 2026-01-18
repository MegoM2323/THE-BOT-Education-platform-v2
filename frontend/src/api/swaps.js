/**
 * API эндпоинты обмена
 */

import apiClient from './client.js';

/**
 * Проверить возможность обмена
 * @param {string} oldLessonId - ID текущего занятия (которое хочет обменять)
 * @param {string} newLessonId - ID целевого занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Результат валидации
 */
export const validateSwap = async (oldLessonId, newLessonId, options = {}) => {
  return await apiClient.post('/swaps/validate', {
    old_lesson_id: oldLessonId,
    new_lesson_id: newLessonId,
  }, options);
};

/**
 * Выполнить операцию обмена
 * @param {string} oldLessonId - ID текущего занятия (которое хочет обменять)
 * @param {string} newLessonId - ID целевого занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Результат обмена с новым бронированием
 */
export const performSwap = async (oldLessonId, newLessonId, options = {}) => {
  return await apiClient.post('/swaps', {
    old_lesson_id: oldLessonId,
    new_lesson_id: newLessonId,
  }, options);
};

/**
 * Получить историю обменов для текущего пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список операций обмена
 */
export const getSwapHistory = async (options = {}) => {
  return await apiClient.get('/swaps/history', options);
};

/**
 * Получить все обмены (только администратор)
 * @param {Object} filters - Параметры фильтрации
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список обменов
 */
export const getAllSwaps = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  if (filters.student_id) params.append('student_id', filters.student_id);
  if (filters.start_date) params.append('start_date', filters.start_date);
  if (filters.end_date) params.append('end_date', filters.end_date);

  const query = params.toString();
  return await apiClient.get(`/swaps${query ? `?${query}` : ''}`, options);
};

export default {
  validateSwap,
  performSwap,
  getSwapHistory,
  getAllSwaps,
};
