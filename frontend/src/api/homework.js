/**
 * API эндпоинты для домашних заданий
 *
 * TODO: Backend endpoints needed for homework text functionality:
 *
 * 1. GET /api/lessons/:lessonId/homework/text
 *    - Get homework text description for a lesson
 *    - Response: { success: true, data: { text: string } }
 *
 * 2. PUT /api/lessons/:lessonId/homework/text
 *    - Save/update homework text description
 *    - Body: { text: string }
 *    - Response: { success: true, data: { text: string, updated_at: string } }
 */

import apiClient from './client.js';

/**
 * Загрузить файл домашнего задания для урока
 * @param {string} lessonId - ID урока (UUID)
 * @param {File} file - Файл для загрузки
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Object>} Информация о загруженном файле: {id, file_name, file_size, mime_type, created_at}
 * @throws {Error} Если возникла ошибка при загрузке файла
 */
export const uploadHomework = async (lessonId, file, signal) => {
  try {
    const formData = new FormData();
    formData.append('file', file);

    // Используем apiClient для поддержки AbortController
    return await apiClient.post(`/lessons/${lessonId}/homework`, formData, {
      signal, // Передаем signal для отмены запроса
    });
  } catch (error) {
    // Не логируем AbortError как ошибку - это нормальное поведение при размонтировании
    if (error.name !== 'AbortError') {
      console.error('Error uploading homework:', error);
    }
    throw error;
  }
};

/**
 * Получить список файлов домашнего задания для урока
 * @param {string} lessonId - ID урока (UUID)
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Array>} Массив файлов, отсортированный по дате создания (новые первые)
 * @throws {Error} Если возникла ошибка при получении списка файлов
 */
export const getHomework = async (lessonId, signal) => {
  try {
    // Используем apiClient для поддержки AbortController
    const data = await apiClient.get(`/lessons/${lessonId}/homework`, {
      signal, // Передаем signal для отмены запроса
    });

    // Backend возвращает массив через apiClient auto-unwrap
    const homeworkArray = Array.isArray(data) ? data : (data?.data || []);

    // Сортируем по дате создания (новые первые)
    return Array.isArray(homeworkArray)
      ? homeworkArray.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
      : [];
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== 'AbortError') {
      console.error('Error fetching homework:', error);
    }
    throw error;
  }
};

/**
 * Удалить файл домашнего задания
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<void>}
 * @throws {Error} Если возникла ошибка при удалении файла
 */
export const deleteHomework = async (lessonId, fileId, signal) => {
  try {
    // Используем apiClient для поддержки AbortController
    return await apiClient.delete(`/lessons/${lessonId}/homework/${fileId}`, {
      signal, // Передаем signal для отмены запроса
    });
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== 'AbortError') {
      console.error('Error deleting homework:', error);
    }
    throw error;
  }
};

/**
 * Скачать файл домашнего задания (открывает в новом окне)
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @returns {void}
 */
export const downloadHomework = (lessonId, fileId) => {
  // Используем правильный базовый URL с /v1
  let baseURL = import.meta.env.VITE_API_URL;
  if (!baseURL) {
    // В dev режиме используем /api/v1, в production - текущий хост + /api/v1
    if (import.meta.env.DEV) {
      baseURL = '/api/v1';
    } else if (typeof window !== 'undefined' && window.location) {
      baseURL = `${window.location.origin}/api/v1`;
    } else {
      baseURL = '/api/v1';
    }
  }
  const downloadURL = `${baseURL}/lessons/${lessonId}/homework/${fileId}/download`;

  // Открываем URL в новом окне для скачивания
  // Cookies будут отправлены автоматически благодаря same-site policy
  window.open(downloadURL, '_blank');
};

/**
 * Получить URL для скачивания файла домашнего задания
 * @param {string} lessonId - ID урока (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @returns {string} URL для скачивания
 */
export const getHomeworkDownloadURL = (lessonId, fileId) => {
  const baseURL = import.meta.env.VITE_API_URL || '/api';
  return `${baseURL}/lessons/${lessonId}/homework/${fileId}/download`;
};

export default {
  uploadHomework,
  getHomework,
  deleteHomework,
  downloadHomework,
  getHomeworkDownloadURL,
};
