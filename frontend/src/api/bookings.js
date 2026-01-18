/**
 * API эндпоинты бронирований
 */

import apiClient from './client.js';
import { logger } from '../utils/logger.js';

/**
 * Валидирует структуру booking объекта
 * @param {Object} booking - Объект бронирования для валидации
 * @returns {boolean} true если booking содержит все необходимые поля
 */
const isValidBooking = (booking) => {
  if (!booking || typeof booking !== 'object') {
    logger.warn('Invalid booking: not an object', booking);
    return false;
  }

  // Проверяем обязательные поля
  const hasId = booking.id || booking.booking_id;
  const hasLessonId = booking.lesson_id;
  const hasValidTime = booking.start_time || (booking.lesson && booking.lesson.start_time);

  if (!hasId) {
    logger.warn('Invalid booking: missing id/booking_id', booking);
    return false;
  }

  if (!hasLessonId) {
    logger.warn('Invalid booking: missing lesson_id', booking);
    return false;
  }

  if (!hasValidTime) {
    logger.warn('Invalid booking: missing start_time', booking);
    return false;
  }

  // Проверяем что поля имеют правильный тип
  if (hasValidTime && typeof hasValidTime !== 'string' && !(hasValidTime instanceof Date)) {
    logger.warn('Invalid booking: start_time has wrong type', booking);
    return false;
  }

  return true;
};

/**
 * Нормализует массив bookings, фильтруя невалидные записи
 * @param {Array} bookingsArray - Массив бронирований
 * @returns {Array} Отфильтрованный массив валидных бронирований
 */
const normalizeBookingsArray = (bookingsArray) => {
  if (!Array.isArray(bookingsArray)) {
    logger.warn('normalizeBookingsArray: input is not an array', bookingsArray);
    return [];
  }

  const validBookings = [];
  const invalidCount = { filtered: 0 };

  bookingsArray.forEach((booking, index) => {
    if (isValidBooking(booking)) {
      validBookings.push(booking);
    } else {
      invalidCount.filtered++;
      logger.debug(`Filtered out invalid booking at index ${index}`, booking);
    }
  });

  if (invalidCount.filtered > 0) {
    logger.info(`Filtered ${invalidCount.filtered} invalid bookings from ${bookingsArray.length} total`);
  }

  return validBookings;
};

/**
 * Получить все бронирования (только администратор) или бронирования пользователя
 * @param {Object} filters - Параметры фильтрации
 * @returns {Promise<Array>} Список бронирований
 */
export const getBookings = async (filters = {}) => {
  const params = new URLSearchParams();

  if (filters.student_id) params.append('student_id', filters.student_id);
  if (filters.lesson_id) params.append('lesson_id', filters.lesson_id);
  if (filters.status) params.append('status', filters.status);

  const query = params.toString();
  const response = await apiClient.get(`/bookings${query ? `?${query}` : ''}`);

  // Backend возвращает {data: {bookings: [...]}, meta: {...}}, извлекаем массив bookings
  let bookingsArray = response.data?.bookings || response.bookings || response || [];

  // Нормализуем и валидируем полученный массив
  bookingsArray = normalizeBookingsArray(bookingsArray);

  return bookingsArray;
};

/**
 * Получить бронирования пользователя
 * @returns {Promise<Array>} Список бронирований пользователя
 */
export const getMyBookings = async () => {
  const response = await apiClient.get('/bookings/my');

  // Backend возвращает {data: {bookings: [...]}, meta: {...}} или массив
  let bookingsArray = Array.isArray(response) ? response : (response.data?.bookings || response?.bookings || response || []);

  // Нормализуем и валидируем полученный массив
  bookingsArray = normalizeBookingsArray(bookingsArray);

  return bookingsArray;
};

/**
 * Получить бронирование по ID
 * @param {number} bookingId - ID бронирования
 * @returns {Promise<Object>} Детали бронирования
 */
export const getBookingById = async (bookingId) => {
  return await apiClient.get(`/bookings/${bookingId}`);
};

/**
 * Создать новое бронирование (студент) или добавить студента на занятие (админ)
 * @param {string} lessonId - ID занятия для бронирования (UUID)
 * @param {string} [studentId] - ID студента (только для администраторов)
 * @returns {Promise<Object>} Созданное бронирование
 */
export const createBooking = async (lessonId, studentId) => {
  const payload = { lesson_id: lessonId };
  if (studentId) {
    payload.student_id = studentId;
  }
  const response = await apiClient.post('/bookings', payload);
  // Backend возвращает {booking: {...}}, извлекаем объект booking
  return response.booking || response;
};

/**
 * Отменить бронирование
 * @param {number} bookingId - ID бронирования для отмены
 * @returns {Promise<Object>} Обновлённое бронирование
 */
export const cancelBooking = async (bookingId) => {
  return await apiClient.delete(`/bookings/${bookingId}`);
};

/**
 * Получить статус бронирования
 * @param {number} bookingId - ID бронирования
 * @returns {Promise<Object>} Объект со статусом: {status: 'active'|'cancelled'}
 */
export const getBookingStatus = async (bookingId) => {
  return await apiClient.get(`/bookings/${bookingId}/status`);
};

/**
 * Обновить статус бронирования (только администратор)
 * @param {number} bookingId - ID бронирования
 * @param {string} status - Новый статус
 * @returns {Promise<Object>} Обновлённое бронирование
 */
export const updateBookingStatus = async (bookingId, status) => {
  return await apiClient.patch(`/bookings/${bookingId}`, { status });
};

/**
 * Получить список уроков, от которых студент отписался (для блокировки повторной записи)
 * @returns {Promise<Array>} Массив lesson_id, которые студент отменил
 */
export const getCancelledLessons = async () => {
  const response = await apiClient.get('/bookings/cancelled-lessons');
  return response.lesson_ids || response || [];
};

export default {
  getBookings,
  getMyBookings,
  getBookingById,
  createBooking,
  cancelBooking,
  getBookingStatus,
  updateBookingStatus,
  getCancelledLessons,
};
