/**
 * API endpoints for lesson templates
 */

import apiClient from './client.js';

/**
 * Get all templates
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} List of templates
 */
export const getTemplates = async (options = {}) => {
  try {
    const response = await apiClient.get('/templates', options);
    console.log('Raw API response for templates:', response);
    if (Array.isArray(response)) {
      return response;
    }
    const templates = response?.templates || response?.data || [];
    console.log('Parsed templates array:', templates);
    return templates;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching templates:', error);
    }
    throw error;
  }
};

/**
 * Get single template with lessons
 * @param {string} templateId - Template ID
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Template with lessons
 */
export const getTemplate = async (templateId, options = {}) => {
  try {
    return await apiClient.get(`/templates/${templateId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error fetching template:', error);
    }
    throw error;
  }
};

/**
 * Create new template
 * @param {Object} templateData - Template data (name, description)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Created template
 */
export const createTemplate = async (templateData, options = {}) => {
  try {
    const response = await apiClient.post('/templates', templateData, options);
    return response;
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error creating template:', error);
    }
    throw error;
  }
};

/**
 * Update template
 * @param {string} templateId - Template ID
 * @param {Object} updates - Fields to update
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Updated template
 */
export const updateTemplate = async (templateId, updates, options = {}) => {
  try {
    return await apiClient.put(`/templates/${templateId}`, updates, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error updating template:', error);
    }
    throw error;
  }
};

/**
 * Delete template
 * @param {string} templateId - Template ID
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const deleteTemplate = async (templateId, options = {}) => {
  try {
    return await apiClient.delete(`/templates/${templateId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error deleting template:', error);
    }
    throw error;
  }
};

/**
 * Apply template to a week
 * @param {string} templateId - Template ID
 * @param {string} weekStartDate - Week start date (YYYY-MM-DD)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Result with created lessons
 */
export const applyTemplate = async (templateId, weekStartDate, options = {}) => {
  try {
    return await apiClient.post(`/templates/${templateId}/apply`, {
      week_start_date: weekStartDate,
    }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error applying template:', error);
    }
    throw error;
  }
};

/**
 * Rollback week to template
 * @param {string} templateId - Template ID
 * @param {string} weekStartDate - Week start date (YYYY-MM-DD)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Result with deleted/created lessons
 */
export const rollbackTemplate = async (templateId, weekStartDate, options = {}) => {
  try {
    return await apiClient.post(`/templates/${templateId}/rollback`, {
      week_start_date: weekStartDate,
    }, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error rolling back template:', error);
    }
    throw error;
  }
};

/**
 * Create template lesson
 * @param {string} templateId - Template ID
 * @param {Object} lessonData - Lesson data
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Created template lesson
 */
export const createTemplateLesson = async (templateId, lessonData, options = {}) => {
  try {
    return await apiClient.post(`/templates/${templateId}/lessons`, lessonData, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error creating template lesson:', error);
    }
    throw error;
  }
};

/**
 * Update template lesson
 * @param {string} templateId - Template ID
 * @param {string} lessonId - Lesson ID
 * @param {Object} updates - Fields to update
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Object>} Updated template lesson
 */
export const updateTemplateLesson = async (templateId, lessonId, updates, options = {}) => {
  try {
    return await apiClient.put(`/templates/${templateId}/lessons/${lessonId}`, updates, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error updating template lesson:', error);
    }
    throw error;
  }
};

/**
 * Delete template lesson
 * @param {string} templateId - Template ID
 * @param {string} lessonId - Lesson ID
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<void>}
 */
export const deleteTemplateLesson = async (templateId, lessonId, options = {}) => {
  try {
    return await apiClient.delete(`/templates/${templateId}/lessons/${lessonId}`, options);
  } catch (error) {
    if (error.name !== 'AbortError') {
      console.error('Error deleting template lesson:', error);
    }
    throw error;
  }
};

export default {
  getTemplates,
  getTemplate,
  createTemplate,
  updateTemplate,
  deleteTemplate,
  applyTemplate,
  rollbackTemplate,
  createTemplateLesson,
  updateTemplateLesson,
  deleteTemplateLesson,
};
