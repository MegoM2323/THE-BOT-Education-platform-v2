/**
 * API эндпоинты для работы учителя
 */

import apiClient from './client.js';

/**
 * Отправить рассылку студентам занятия
 * @param {string} lessonId - UUID занятия
 * @param {string} message - Текст сообщения (макс. 4096 символов)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Результат отправки рассылки
 */
export const sendLessonBroadcast = async (lessonId, message, options = {}) => {
  try {
    return await apiClient.post(`/teacher/lessons/${lessonId}/broadcast`, { message }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error sending lesson broadcast:', error);
    }
    throw error;
  }
};

export default {
  sendLessonBroadcast,
};
