/**
 * API эндпоинты занятий
 */

import apiClient from "./client.js";
import { toLocalDateString } from '../utils/dateFormat.js';

/**
 * Получить все занятия с опциональными фильтрами
 * @param {Object} filters - Параметры фильтрации
 *   - teacher_id: UUID фильтра по преподавателю
 *   - date: дата в формате YYYY-MM-DD для фильтра по конкретному дню
 *   - available: true/false для фильтра по доступности
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список занятий с информацией о bookings
 * @note Backend accepts: teacher_id, date, available
 *       start_date/end_date/subject не поддерживаются бэкендом
 */
export const getLessons = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  // Поддерживаемые параметры бэкенда:
  // 1. teacher_id - UUID преподавателя
  if (filters.teacher_id) params.append("teacher_id", filters.teacher_id);

  // 2. date - конкретная дата (формат YYYY-MM-DD)
  // Примечание: Бэкенд преобразует это в интервал от начала до конца дня
  if (filters.date) params.append("date", filters.date);

  // 3. available - фильтр по доступности (true/false)
  if (filters.available !== undefined && filters.available !== null) {
    params.append("available", String(filters.available));
  }

  // НЕ ИСПОЛЬЗУЕМЫЕ параметры (удалены, т.к. бэкенд их не поддерживает):
  // - start_date / end_date (используйте date вместо)
  // - subject (не поддерживается бэкендом)

  const query = params.toString();

  try {
    const response = await apiClient.get(
      `/lessons${query ? `?${query}` : ""}`,
      options,
    );
    // Backend возвращает {success, data: lessons[]}, извлекаем массив lessons
    // Handle null/undefined response - return as is for test compatibility
    if (response === null) return null;
    if (response === undefined) return undefined;

    let lessons = [];
    // Backend structure: {success, data: [...]} или массив напрямую
    if (Array.isArray(response)) {
      lessons = response;
    }
    // Если response.data - массив, возвращаем его
    else if ("data" in response && Array.isArray(response.data)) {
      lessons = response.data;
    }
    // Если есть response.lessons массив, возвращаем его
    else if (response?.lessons && Array.isArray(response.lessons)) {
      lessons = response.lessons;
    }
    // Для любого другого формата возвращаем response.data или сам response
    else if ("data" in response) {
      lessons = response.data;
    }
    else {
      lessons = response;
    }

    return lessons;
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error fetching lessons:", error);
    }
    throw error;
  }
};

/**
 * Получить занятия для текущего пользователя (студент или преподаватель)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список занятий пользователя
 */
export const getMyLessons = async (options = {}) => {
  try {
    const response = await apiClient.get("/lessons/my", options);
    // Handle null/undefined response
    if (response === null) return null;
    if (response === undefined) return undefined;
    // Backend возвращает {success, data: lessons[]}, извлекаем массив lessons
    if (Array.isArray(response)) {
      return response;
    }
    // Если response.data - массив, возвращаем его
    if ("data" in response && Array.isArray(response.data)) {
      return response.data;
    }
    // Если есть response.lessons массив, возвращаем его
    if (response?.lessons && Array.isArray(response.lessons)) {
      return response.lessons;
    }
    // Для любого другого формата возвращаем response.data или сам response
    if ("data" in response) {
      return response.data;
    }
    return response;
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error fetching my lessons:", error);
    }
    throw error;
  }
};

/**
 * Получить расписание преподавателя с дополнительной информацией
 * @param {Object} params - Параметры запроса
 *   - start_date: начало периода (ISO string или YYYY-MM-DD)
 *   - end_date: конец периода (ISO string или YYYY-MM-DD)
 *   - teacher_id: UUID преподавателя (только для админа)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} { lessons: Array, count: number }
 */
export const getTeacherSchedule = async (params = {}, options = {}) => {
  const queryParams = new URLSearchParams();

  // Конвертируем даты в YYYY-MM-DD формат в ЛОКАЛЬНОМ часовом поясе
  // (toISOString конвертирует в UTC, что смещает дату на 1 день для UTC+N часовых поясов)
  if (params.start_date) {
    queryParams.append("start_date", toLocalDateString(params.start_date));
  }

  if (params.end_date) {
    queryParams.append("end_date", toLocalDateString(params.end_date));
  }

  // teacher_id опционален (только для админа)
  if (params.teacher_id) {
    queryParams.append("teacher_id", params.teacher_id);
  }

  const query = queryParams.toString();

  try {
    const response = await apiClient.get(
      `/teacher/schedule${query ? `?${query}` : ""}`,
      options,
    );

    if (response === null) return null;
    if (response === undefined) return undefined;

    // Backend возвращает {success, data: {lessons: [...], count: N}}
    if (response?.data) {
      return response.data;
    }

    if (Array.isArray(response)) {
      return { lessons: response, count: response.length };
    }

    if (response?.lessons && Array.isArray(response.lessons)) {
      return {
        lessons: response.lessons,
        count: response.count || response.lessons.length,
      };
    }

    return response;
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error fetching teacher schedule:", error);
    }
    throw error;
  }
};

/**
 * Получить занятие по ID
 * @param {number} lessonId - ID занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Детали занятия
 */
export const getLessonById = async (lessonId, options = {}) => {
  try {
    return await apiClient.get(`/lessons/${lessonId}`, options);
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error fetching lesson:", error);
    }
    throw error;
  }
};

/**
 * Создать новое занятие (только преподаватель/администратор)
 * @param {Object} lessonData - Данные занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Созданное занятие
 */
export const createLesson = async (lessonData, options = {}) => {
  try {
    return await apiClient.post("/lessons", lessonData, options);
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error creating lesson:", error);
    }
    throw error;
  }
};

/**
 * Обновить занятие (только преподаватель/администратор)
 * @param {number} lessonId - ID занятия
 * @param {Object} updates - Обновляемые поля
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Обновлённое занятие
 */
export const updateLesson = async (lessonId, updates, options = {}) => {
  if (!lessonId) {
    throw new Error('lessonId is required');
  }
  if (!updates || typeof updates !== 'object') {
    throw new Error('updates must be a non-null object');
  }
  try {
    return await apiClient.put(`/lessons/${lessonId}`, updates, options);
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error updating lesson:", error);
    }
    throw error;
  }
};

/**
 * Удалить занятие (только администратор)
 * @param {number} lessonId - ID занятия
 * @param {boolean} [deleteSeries=false] - Удалить всю серию повторяющихся занятий
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const deleteLesson = async (lessonId, deleteSeries = false, options = {}) => {
  try {
    const url = deleteSeries
      ? `/lessons/${lessonId}?delete_series=true`
      : `/lessons/${lessonId}`;
    return await apiClient.delete(url, options);
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error deleting lesson:", error);
    }
    throw error;
  }
};

/**
 * Получить доступные слоты для бронирования
 * @param {Object} filters - Параметры фильтрации
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Доступные слоты занятий
 */
export const getAvailableSlots = async (filters = {}, options = {}) => {
  return await getLessons({ ...filters, available: true }, options);
};

/**
 * Получить список студентов, записанных на занятие
 * @param {string} lessonId - ID занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Объект с массивом студентов
 */
export const getLessonStudents = async (lessonId, options = {}) => {
  try {
    const response = await apiClient.get(
      `/lessons/${lessonId}/students`,
      options,
    );
    // Обработка различных форматов ответа
    if (Array.isArray(response)) {
      return { students: response };
    }
    return response;
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error fetching lesson students:", error);
    }
    throw error;
  }
};

/**
 * Отправить отчет о занятии родителям студентов
 * @param {string} lessonId - ID занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Результат отправки { sent, failed, total_students, errors }
 */
export const sendReportToParents = async (lessonId, options = {}) => {
  try {
    return await apiClient.post(
      `/lessons/${lessonId}/report/send-to-parents`,
      {},
      options,
    );
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error sending report to parents:", error);
    }
    throw error;
  }
};

/**
 * Создать серию повторяющихся занятий
 * @param {string} lessonId - ID исходного занятия
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} { count: number, lessons: [] }
 * Backend автоматически создаст серию на 4 месяца (семестр)
 */
export const createRecurringSeries = async (lessonId, options = {}) => {
  try {
    return await apiClient.post(
      `/lessons/${lessonId}/recurring`,
      {}, // Backend вычисляет количество недель автоматически
      options,
    );
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error creating recurring series:", error);
    }
    throw error;
  }
};

/**
 * Отменить серию повторяющихся занятий (удалить будущие)
 * @param {string} lessonId - ID занятия из серии
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} { message: string, deleted_count: number }
 */
export const cancelRecurringSeries = async (lessonId, options = {}) => {
  try {
    return await apiClient.delete(`/lessons/${lessonId}/recurring`, options);
  } catch (error) {
    if (error.name !== "AbortError") {
      console.error("Error cancelling recurring series:", error);
    }
    throw error;
  }
};

export default {
  getLessons,
  getMyLessons,
  getTeacherSchedule,
  getLessonById,
  createLesson,
  updateLesson,
  deleteLesson,
  getAvailableSlots,
  getLessonStudents,
  sendReportToParents,
  createRecurringSeries,
  cancelRecurringSeries,
};
