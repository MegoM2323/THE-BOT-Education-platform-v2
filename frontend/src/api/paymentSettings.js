/**
 * API endpoints для управления настройками платежей (admin only)
 */

import apiClient from './client.js';

/**
 * Получить список студентов с их статусом платежей
 * @param {Object} options - Параметры запроса
 * @param {boolean} [options.payment_enabled] - Фильтр по статусу платежей (true/false)
 * @param {string} [options.search] - Поиск по имени или email
 * @param {number} [options.limit] - Количество записей (по умолчанию: без ограничений)
 * @param {number} [options.offset] - Смещение для пагинации (по умолчанию: 0)
 * @returns {Promise<Array>} Список студентов: {id, full_name, email, payment_enabled, updated_at}
 * @throws {Error} Если возникла ошибка при получении списка
 */
export const getStudentsPaymentStatus = async (options = {}) => {
  try {
    const params = new URLSearchParams();

    // Фильтр по статусу платежей
    if (options.payment_enabled !== undefined) {
      params.append('payment_enabled', options.payment_enabled.toString());
    }

    // Поиск по имени/email
    if (options.search) {
      params.append('search', options.search);
    }

    // Пагинация
    if (options.limit) {
      params.append('limit', options.limit.toString());
    }

    if (options.offset !== undefined) {
      params.append('offset', options.offset.toString());
    }

    const query = params.toString();
    const response = await apiClient.get(`/admin/payment-settings${query ? `?${query}` : ''}`);

    // Backend возвращает {success: true, data: [...]} или {students: [...]}
    const studentsArray = response.students || response.data || response || [];

    return Array.isArray(studentsArray) ? studentsArray : [];
  } catch (error) {
    console.error('Error fetching students payment status:', error);
    throw error;
  }
};

/**
 * Обновить статус платежей для студента
 * @param {string} userId - ID студента (UUID)
 * @param {boolean} enabled - Новый статус платежей (true - разрешены, false - запрещены)
 * @returns {Promise<Object>} Обновленные данные студента: {id, full_name, email, payment_enabled, updated_at}
 * @throws {Error} Если возникла ошибка при обновлении статуса
 */
export const updatePaymentStatus = async (userId, enabled) => {
  try {
    const response = await apiClient.put(`/admin/users/${userId}/payment-settings`, {
      payment_enabled: enabled,
    });

    // Backend возвращает {success: true, data: user}
    return response.data || response;
  } catch (error) {
    console.error('Error updating payment status:', error);

    // Форматируем сообщение об ошибке
    let errorMessage = 'Failed to update payment status';

    if (error.status === 403) {
      errorMessage = 'Недостаточно прав для изменения настроек платежей';
    } else if (error.status === 404) {
      errorMessage = 'Пользователь не найден';
    } else if (error.status === 400) {
      errorMessage = error.message || 'Неверные данные запроса';
    } else if (error.message) {
      errorMessage = error.message;
    }

    throw new Error(errorMessage);
  }
};

export default {
  getStudentsPaymentStatus,
  updatePaymentStatus,
};
