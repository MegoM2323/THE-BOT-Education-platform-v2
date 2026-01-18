/**
 * API эндпоинты пользователей (функциональность администратора)
 */

import apiClient from './client.js';

/**
 * Получить всех пользователей (только администратор)
 * @param {Object} filters - Параметры фильтрации
 * @param {string} [filters.role] - Фильтр по роли (student, teacher, admin)
 * @param {string} [filters.search] - Поиск по имени или email
 * @param {boolean} [filters.active] - Фильтр по активности
 * @param {number} [filters.page] - Номер страницы (начиная с 1)
 * @param {number} [filters.per_page] - Количество на странице (макс 100)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список пользователей
 */
export const getUsers = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  if (filters.role) params.append('role', filters.role);
  if (filters.search) params.append('search', filters.search);
  if (filters.active !== undefined) params.append('active', String(filters.active));
  if (filters.page) params.append('page', String(filters.page));
  if (filters.per_page) params.append('per_page', String(filters.per_page));

  const query = params.toString();

  try {
    const response = await apiClient.get(`/users${query ? `?${query}` : ''}`, options);
    // apiClient извлекает data из {success, data} обертки
    // Backend возвращает {users: [...], meta: {...}} после извлечения
    // Обрабатываем разные форматы ответа
    if (Array.isArray(response)) {
      return response;
    }
    // Стандартный формат: { users: [...], meta: {...} }
    if (response?.users && Array.isArray(response.users)) {
      return response.users;
    }
    // Вложенный формат: { data: { users: [...] } }
    if (response?.data?.users && Array.isArray(response.data.users)) {
      return response.data.users;
    }
    // Если ничего не подходит, возвращаем пустой массив
    console.warn('[getUsers] Unexpected response format:', response);
    return [];
  } catch (error) {
    // Не логируем AbortError - это нормальное поведение при отмене запроса
    if (error.name !== 'AbortError') {
      console.error('Error fetching users:', error);
    }
    throw error;
  }
};

/**
 * Получить пользователя по ID (только администратор)
 * @param {number} userId - ID пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Детали пользователя
 */
export const getUserById = async (userId, options = {}) => {
  try {
    return await apiClient.get(`/users/${userId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching user:', error);
    }
    throw error;
  }
};

/**
 * Создать нового пользователя (только администратор)
 * @param {Object} userData - Данные пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Созданный пользователь
 */
export const createUser = async (userData, options = {}) => {
  try {
    return await apiClient.post('/users', userData, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error creating user:', error);
    }
    throw error;
  }
};

/**
 * Обновить пользователя (только администратор)
 * @param {number} userId - ID пользователя
 * @param {Object} updates - Обновляемые поля
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Обновлённый пользователь
 */
export const updateUser = async (userId, updates, options = {}) => {
  try {
    return await apiClient.put(`/users/${userId}`, updates, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error updating user:', error);
    }
    throw error;
  }
};

/**
 * Удалить пользователя (только администратор)
 * @param {number} userId - ID пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const deleteUser = async (userId, options = {}) => {
  try {
    return await apiClient.delete(`/users/${userId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error deleting user:', error);
    }
    throw error;
  }
};

/**
 * Получить статистику пользователя (только администратор)
 * @param {number} userId - ID пользователя
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Статистика пользователя
 */
export const getUserStats = async (userId, options = {}) => {
  try {
    return await apiClient.get(`/users/${userId}/stats`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching user stats:', error);
    }
    throw error;
  }
};

/**
 * Получить список учителей (методистов)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список учителей
 */
export const getTeachers = async (options = {}) => {
  return await getUsers({ role: 'methodologist' }, options);
};

/**
 * Получить список студентов
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Список студентов
 */
export const getStudents = async (options = {}) => {
  return await getUsers({ role: 'student' }, options);
};

/**
 * Получить пользователей с мета-данными пагинации
 * @param {Object} filters - Параметры фильтрации (см. getUsers)
 * @param {Object} [options] - Опции запроса
 * @returns {Promise<{users: Array, meta: Object}>} Пользователи и мета-данные
 */
export const getUsersWithPagination = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  if (filters.role) params.append('role', filters.role);
  if (filters.search) params.append('search', filters.search);
  if (filters.active !== undefined) params.append('active', String(filters.active));
  if (filters.page) params.append('page', String(filters.page));
  if (filters.per_page) params.append('per_page', String(filters.per_page));

  const query = params.toString();

  try {
    const response = await apiClient.get(`/users${query ? `?${query}` : ''}`, options);

    // Извлекаем пользователей
    let users = [];
    if (Array.isArray(response)) {
      users = response;
    } else if (response?.users && Array.isArray(response.users)) {
      users = response.users;
    } else if (response?.data?.users && Array.isArray(response.data.users)) {
      users = response.data.users;
    }

    // Извлекаем мета-данные пагинации
    const meta = response?.meta || response?.data?.meta || {
      page: filters.page || 1,
      per_page: filters.per_page || 20,
      total: users.length,
      total_pages: 1,
    };

    return { users, meta };
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching users with pagination:', error);
    }
    throw error;
  }
};

/**
 * Получить студентов с мета-данными пагинации
 * @param {Object} filters - Параметры фильтрации (page, per_page, search)
 * @param {Object} [options] - Опции запроса
 * @returns {Promise<{users: Array, meta: Object}>} Студенты и мета-данные
 */
export const getStudentsWithPagination = async (filters = {}, options = {}) => {
  return await getUsersWithPagination({ ...filters, role: 'student' }, options);
};

/**
 * Получить ВСЕ студентов с автоматической пагинацией через все страницы
 * (в отличие от getStudents которая получает только первые 20)
 * @param {Object} filters - Параметры фильтрации (search и т.д., без page/per_page)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Полный список всех студентов
 */
export const getStudentsAll = async (filters = {}, options = {}) => {
  try {
    const allStudents = [];
    let page = 1;
    const per_page = 50;

    while (true) {
      const result = await getUsersWithPagination(
        { ...filters, role: 'student', page, per_page },
        options
      );

      if (!result || !result.meta || !Array.isArray(result.users)) {
        console.error('Invalid API response structure');
        break;
      }

      allStudents.push(...result.users);

      // Останавливаемся если мы на последней странице
      if (page >= result.meta.total_pages) {
        break;
      }

      page += 1;
    }

    return allStudents;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching all students:', error);
    }
    throw error;
  }
};

/**
 * Получить ВСЕ учителей (методистов) с автоматической пагинацией через все страницы
 * (в отличие от getTeachers которая получает только первые 20)
 * @param {Object} filters - Параметры фильтрации (search и т.д., без page/per_page)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Полный список всех учителей
 */
export const getTeachersAll = async (filters = {}, options = {}) => {
  try {
    const allTeachers = [];
    let page = 1;
    const per_page = 50;

    while (true) {
      const result = await getUsersWithPagination(
        { ...filters, role: 'methodologist', page, per_page },
        options
      );

      if (!result || !result.meta || !Array.isArray(result.users)) {
        console.error('Invalid API response structure');
        break;
      }

      allTeachers.push(...result.users);

      if (page >= result.meta.total_pages) {
        break;
      }

      page += 1;
    }

    return allTeachers;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching all teachers:', error);
    }
    throw error;
  }
};

/**
 * Получить ВСЕ пользователей с автоматической пагинацией через все страницы
 * (универсальная функция для всех типов пользователей - студентов, преподавателей, администраторов)
 * @param {Object} filters - Параметры фильтрации (role, search, active и т.д., без page/per_page)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Полный список всех пользователей
 */
export const getUsersAll = async (filters = {}, options = {}) => {
  try {
    const allUsers = [];
    let page = 1;
    const per_page = 50;

    while (true) {
      const result = await getUsersWithPagination(
        { ...filters, page, per_page },
        options
      );

      if (!result || !result.meta || !Array.isArray(result.users)) {
        console.error('Invalid API response structure');
        break;
      }

      allUsers.push(...result.users);

      // Останавливаемся если мы на последней странице
      if (page >= result.meta.total_pages) {
        break;
      }

      page += 1;
    }

    return allUsers;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching all users:', error);
    }
    throw error;
  }
};

/**
 * Получить ВСЕ пользователей, которые могут быть назначены учителями
 * (учителя/методисты и администраторы)
 * @param {Object} filters - Параметры фильтрации (search и т.д., без page/per_page)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Полный список всех пользователей, которые могут быть учителями
 */
export const getAssignableTeachersAll = async (filters = {}, options = {}) => {
  try {
    const [methodologists, admins] = await Promise.all([
      getUsersAll({ ...filters, role: 'methodologist' }, options),
      getUsersAll({ ...filters, role: 'admin' }, options),
    ]);

    const allAssignableTeachers = [...methodologists, ...admins];

    const seen = new Set();
    const deduplicated = [];
    for (const user of allAssignableTeachers) {
      if (!seen.has(user.id)) {
        seen.add(user.id);
        deduplicated.push(user);
      }
    }

    deduplicated.sort((a, b) =>
      (a.full_name || '').localeCompare(b.full_name || '', 'ru')
    );

    return deduplicated;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching assignable teachers:', error);
    }
    throw error;
  }
};

/**
 * Обновить профиль текущего пользователя
 * @param {Object} updates - Поля для обновления (например, telegram_username)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Обновлённые данные пользователя
 */
export const updateMyProfile = async (updates, options = {}) => {
  try {
    return await apiClient.put('/auth/profile', updates, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error updating profile:', error);
    }
    throw error;
  }
};

export default {
  getUsers,
  getUserById,
  createUser,
  updateUser,
  deleteUser,
  getUserStats,
  getTeachers,
  getStudents,
  getStudentsAll,
  getTeachersAll,
  getUsersAll,
  getAssignableTeachersAll,
  getUsersWithPagination,
  getStudentsWithPagination,
  updateMyProfile,
};
