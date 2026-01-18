/**
 * API endpoints for subjects management
 */

import apiClient from './client.js';

/**
 * Fetch all available subjects
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} List of all subjects
 */
export const fetchAllSubjects = async (options = {}) => {
  try {
    const response = await apiClient.get('/subjects', options);
    // Обработка различных форматов ответа
    if (Array.isArray(response)) {
      return response;
    }
    return response?.subjects || response?.data || [];
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching subjects:', error);
    }
    throw error;
  }
};

/**
 * Fetch subjects for a specific teacher
 * @param {string} teacherId - Teacher's UUID
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} List of teacher's subjects
 */
export const fetchTeacherSubjects = async (teacherId, options = {}) => {
  try {
    const response = await apiClient.get(`/teachers/${teacherId}/subjects`, options);
    if (Array.isArray(response)) {
      return response;
    }
    return response?.subjects || response?.data || [];
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching teacher subjects:', error);
    }
    throw error;
  }
};

/**
 * Assign a subject to a teacher (admin only)
 * @param {string} teacherId - Teacher's UUID
 * @param {string} subjectId - Subject's UUID
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Assignment result
 */
export const assignSubjectToTeacher = async (teacherId, subjectId, options = {}) => {
  try {
    const response = await apiClient.post(`/teachers/${teacherId}/subjects`, {
      subject_id: subjectId,
    }, options);
    return response;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error assigning subject to teacher:', error);
    }
    throw error;
  }
};

/**
 * Remove a subject from a teacher (admin only)
 * @param {string} teacherId - Teacher's UUID
 * @param {string} subjectId - Subject's UUID
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const removeSubjectFromTeacher = async (teacherId, subjectId, options = {}) => {
  try {
    await apiClient.delete(`/teachers/${teacherId}/subjects/${subjectId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error removing subject from teacher:', error);
    }
    throw error;
  }
};

/**
 * Get current logged-in teacher's subjects
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} List of current teacher's subjects
 */
export const getCurrentUserSubjects = async (options = {}) => {
  try {
    const response = await apiClient.get('/my-subjects', options);
    if (Array.isArray(response)) {
      return response;
    }
    return response?.subjects || response?.data || [];
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching current user subjects:', error);
    }
    throw error;
  }
};

export default {
  fetchAllSubjects,
  fetchTeacherSubjects,
  assignSubjectToTeacher,
  removeSubjectFromTeacher,
  getCurrentUserSubjects,
};
