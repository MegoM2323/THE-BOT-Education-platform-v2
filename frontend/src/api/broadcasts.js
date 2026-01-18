/**
 * API эндпоинты для работы с рассылками в Telegram (только для администраторов)
 */

import apiClient from './client.js';

/**
 * Получить всех пользователей с привязанным Telegram
 * @param {string} [role] - Фильтр по роли (student/teacher/admin)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список пользователей с привязанным Telegram
 */
export const getLinkedUsers = async (role, options = {}) => {
  const params = role ? `?role=${role}` : '';
  return await apiClient.get(`/admin/telegram/users${params}`, options);
};

/**
 * Создать список рассылки
 * @param {Object} data - Данные списка
 * @param {string} data.name - Название списка
 * @param {string} data.description - Описание списка
 * @param {Array<number>} data.user_ids - Массив ID пользователей
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Созданный список рассылки
 */
export const createBroadcastList = async (data, options = {}) => {
  return await apiClient.post('/admin/telegram/lists', data, options);
};

/**
 * Получить все списки рассылки
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список всех списков рассылки
 */
export const getBroadcastLists = async (options = {}) => {
  return await apiClient.get('/admin/telegram/lists', options);
};

/**
 * Получить конкретный список рассылки по ID
 * @param {number} id - ID списка
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Данные списка рассылки
 */
export const getBroadcastListById = async (id, options = {}) => {
  return await apiClient.get(`/admin/telegram/lists/${id}`, options);
};

/**
 * Обновить список рассылки
 * @param {number} id - ID списка
 * @param {Object} data - Обновленные данные
 * @param {string} [data.name] - Название списка
 * @param {string} [data.description] - Описание списка
 * @param {Array<number>} [data.user_ids] - Массив ID пользователей
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Обновленный список
 */
export const updateBroadcastList = async (id, data, options = {}) => {
  return await apiClient.put(`/admin/telegram/lists/${id}`, data, options);
};

/**
 * Удалить список рассылки
 * @param {number} id - ID списка
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Сообщение об успехе
 */
export const deleteBroadcastList = async (id, options = {}) => {
  return await apiClient.delete(`/admin/telegram/lists/${id}`, options);
};

/**
 * Отправить рассылку
 * @param {Object} data - Данные рассылки
 * @param {number} data.list_id - ID списка получателей
 * @param {string} data.message - Текст сообщения (макс. 4096 символов)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Данные отправленной рассылки
 */
export const sendBroadcast = async (data, options = {}) => {
  return await apiClient.post('/admin/telegram/broadcast', data, options);
};

/**
 * Получить историю рассылок
 * @param {Object} [params] - Параметры пагинации
 * @param {number} [params.limit=20] - Количество записей
 * @param {number} [params.offset=0] - Смещение
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Объект с рассылками и общим количеством
 * @property {Array} broadcasts - Массив рассылок
 * @property {number} total - Общее количество рассылок
 */
export const getBroadcasts = async (params = {}, options = {}) => {
  const { limit = 20, offset = 0 } = params;
  return await apiClient.get(`/admin/telegram/broadcasts?limit=${limit}&offset=${offset}`, options);
};

/**
 * Получить детальную информацию о рассылке
 * @param {number} id - ID рассылки
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Детали рассылки с информацией о доставках
 */
export const getBroadcastDetails = async (id, options = {}) => {
  return await apiClient.get(`/admin/telegram/broadcasts/${id}`, options);
};

/**
 * Отменить рассылку
 * @param {number} id - ID рассылки
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Сообщение об успехе
 */
export const cancelBroadcast = async (id, options = {}) => {
  return await apiClient.post(`/admin/telegram/broadcasts/${id}/cancel`, {}, options);
};

export default {
  getLinkedUsers,
  createBroadcastList,
  getBroadcastLists,
  getBroadcastListById,
  updateBroadcastList,
  deleteBroadcastList,
  sendBroadcast,
  getBroadcasts,
  getBroadcastDetails,
  cancelBroadcast,
};
