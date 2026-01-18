/**
 * API эндпоинты для рассылок уроков
 */

import apiClient from './client.js';

/**
 * Создать рассылку для студентов урока
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} message - Текст сообщения (макс 4096 символов)
 * @param {File[]} files - Массив файлов для прикрепления (опционально)
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Object>} Информация о созданной рассылке: {id, lesson_id, message, status, created_at, files}
 * @throws {Error} Если возникла ошибка при создании рассылки
 */
export const createLessonBroadcast = async (lessonId, message, files = [], signal) => {
  try {
    const formData = new FormData();
    formData.append('message', message);

    // Добавляем файлы если есть
    if (files && files.length > 0) {
      files.forEach(file => formData.append('files', file));
    }

    // Используем apiClient для поддержки AbortController
    return await apiClient.post(`/lessons/${lessonId}/broadcasts`, formData, {
      signal, // Передаем signal для отмены запроса
    });
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== 'AbortError') {
      console.error('Error creating lesson broadcast:', error);
    }
    throw error;
  }
};

/**
 * Получить список рассылок для урока
 * @param {string} lessonId - ID урока (UUID)
 * @param {Object} options - Параметры запроса
 * @param {string} [options.status] - Фильтр по статусу ('pending', 'sent', 'failed')
 * @param {number} [options.limit] - Количество записей
 * @param {number} [options.offset] - Смещение для пагинации
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Array>} Массив рассылок с файлами
 * @throws {Error} Если возникла ошибка при получении списка рассылок
 */
export const getLessonBroadcasts = async (lessonId, options = {}, signal) => {
  try {
    const params = new URLSearchParams();
    if (options.status) params.append('status', options.status);
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());

    const queryString = params.toString();
    const endpoint = `/lessons/${lessonId}/broadcasts${queryString ? `?${queryString}` : ''}`;

    // Используем apiClient для поддержки AbortController
    const data = await apiClient.get(endpoint, {
      signal, // Передаем signal для отмены запроса
    });

    // Backend возвращает {broadcasts: [...], count, total, ...} через apiClient auto-unwrap
    const broadcastsArray = data?.broadcasts || data || [];

    return Array.isArray(broadcastsArray) ? broadcastsArray : [];
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== 'AbortError') {
      console.error('Error fetching lesson broadcasts:', error);
    }
    throw error;
  }
};

/**
 * Получить детальную информацию о рассылке
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} broadcastId - ID рассылки (UUID)
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Object>} Детали рассылки с полной информацией о файлах
 * @throws {Error} Если возникла ошибка при получении деталей рассылки
 */
export const getLessonBroadcast = async (lessonId, broadcastId, signal) => {
  try {
    // Используем apiClient для поддержки AbortController
    return await apiClient.get(`/lessons/${lessonId}/broadcasts/${broadcastId}`, {
      signal, // Передаем signal для отмены запроса
    });
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== 'AbortError') {
      console.error('Error fetching broadcast details:', error);
    }
    throw error;
  }
};

/**
 * Скачать файл из рассылки (открывает в новом окне)
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} broadcastId - ID рассылки (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @returns {void}
 */
export const getBroadcastFile = (lessonId, broadcastId, fileId) => {
  const baseURL = import.meta.env.VITE_API_URL || '/api';
  const downloadURL = `${baseURL}/v1/lessons/${lessonId}/broadcasts/${broadcastId}/files/${fileId}/download`;

  // Открываем URL в новом окне для скачивания
  // Браузер автоматически включает cookies (включая session ID) для same-origin запросов
  window.open(downloadURL, '_blank');
};

/**
 * Получить URL для скачивания файла из рассылки
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} broadcastId - ID рассылки (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @returns {string} URL для скачивания
 */
export const getBroadcastFileDownloadURL = (lessonId, broadcastId, fileId) => {
  const baseURL = import.meta.env.VITE_API_URL || '/api';
  return `${baseURL}/v1/lessons/${lessonId}/broadcasts/${broadcastId}/files/${fileId}/download`;
};

export default {
  createLessonBroadcast,
  getLessonBroadcasts,
  getLessonBroadcast,
  getBroadcastFile,
  getBroadcastFileDownloadURL,
};
