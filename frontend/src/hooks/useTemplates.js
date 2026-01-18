import { useCallback } from 'react';
import { logger } from '../utils/logger.js';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as templatesAPI from '../api/templates.js';
import { useNotification } from './useNotification.js';
import { invalidateCreditBalance } from './useCredits.js';

/**
 * Custom hook for managing lesson templates (legacy - multiple templates)
 */
export const useTemplates = () => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Query for fetching all templates
  const {
    data: templates = [],
    isLoading: loading,
    error: queryError,
    refetch: refetchQuery,
  } = useQuery({
    queryKey: ['templates'],
    queryFn: async () => {
      logger.debug('useQuery: Fetching templates');
      try {
        const data = await templatesAPI.getTemplates();
        logger.debug('useQuery: Successfully loaded', data?.length || 0, 'templates');
        return Array.isArray(data) ? data : [];
      } catch (err) {
        console.error('useQuery error:', err.message);
        notification.error('Не удалось загрузить шаблоны');
        throw err;
      }
    },
  });

  // Mutation for creating template
  const createTemplateMutation = useMutation({
    mutationFn: async (templateData) => {
      logger.debug('createTemplate: Starting with data:', templateData);
      const newTemplate = await templatesAPI.createTemplate(templateData);
      logger.debug('createTemplate: Created successfully:', newTemplate);
      return newTemplate;
    },
    onSuccess: () => {
      notification.success('Шаблон создан успешно');
      queryClient.invalidateQueries({ queryKey: ['templates'] });
    },
    onError: (err) => {
      console.error('createTemplate error:', err.message);
      notification.error(err.message || 'Не удалось создать шаблон');
    },
  });

  // Mutation for updating template
  const updateTemplateMutation = useMutation({
    mutationFn: async ({ templateId, updates }) => {
      logger.debug('updateTemplate: Starting with id:', templateId, 'updates:', updates);
      const updated = await templatesAPI.updateTemplate(templateId, updates);
      logger.debug('updateTemplate: Updated successfully:', updated);
      return { updated, templateId };
    },
    onSuccess: ({ templateId }) => {
      notification.success('Шаблон обновлен успешно');
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      queryClient.invalidateQueries({ queryKey: ['templates', templateId] });
    },
    onError: (err) => {
      console.error('updateTemplate error:', err.message);
      notification.error(err.message || 'Не удалось обновить шаблон');
    },
  });

  // Mutation for deleting template
  const deleteTemplateMutation = useMutation({
    mutationFn: async (templateId) => {
      logger.debug('deleteTemplate: Starting with id:', templateId);
      await templatesAPI.deleteTemplate(templateId);
      logger.debug('deleteTemplate: Deleted successfully');
    },
    onSuccess: () => {
      notification.success('Шаблон удален успешно');
      queryClient.invalidateQueries({ queryKey: ['templates'] });
    },
    onError: (err) => {
      console.error('deleteTemplate error:', err.message);
      notification.error(err.message || 'Не удалось удалить шаблон');
    },
  });

  // Mutation for applying template
  const applyTemplateMutation = useMutation({
    mutationFn: async ({ templateId, weekStartDate }) => {
      logger.debug('applyTemplate: Starting with id:', templateId, 'week:', weekStartDate);
      const result = await templatesAPI.applyTemplate(templateId, weekStartDate);
      logger.debug('applyTemplate: Applied successfully:', result);
      return result;
    },
    onSuccess: async (result) => {
      // Backend возвращает created_lessons_count (не lessons_created)
      const createdCount = result.created_lessons_count || 0;
      notification.success(`Шаблон применен: создано занятий - ${createdCount}`);
      // Инвалидация всех связанных кешей (занятия создались, бронирования и кредиты изменились)
      // Используем Promise.all вместо Promise.allSettled для гарантии завершения всех запросов
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false }),
        invalidateCreditBalance(queryClient),
        queryClient.invalidateQueries({ queryKey: ['templates'], exact: false }),
      ]);
    },
    onError: (err) => {
      console.error('applyTemplate error:', err.message);
      notification.error(err.message || 'Не удалось применить шаблон');
    },
  });

  // Mutation for rolling back template
  const rollbackTemplateMutation = useMutation({
    mutationFn: async ({ templateId, weekStartDate }) => {
      logger.debug('rollbackTemplate: Starting with id:', templateId, 'week:', weekStartDate);
      const result = await templatesAPI.rollbackTemplate(templateId, weekStartDate);
      logger.debug('rollbackTemplate: Rolled back successfully:', result);
      return result;
    },
    onSuccess: async (result) => {
      // Backend возвращает deleted_lessons (не lessons_deleted)
      const deletedCount = result.deleted_lessons || 0;
      notification.success(`Шаблон откачен: удалено занятий - ${deletedCount}`);
      // Инвалидация всех связанных кешей (занятия удалены, бронирования отменены, кредиты возвращены)
      // Используем Promise.all вместо Promise.allSettled для гарантии завершения всех запросов
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false }),
        queryClient.invalidateQueries({ queryKey: ['bookings'], exact: false }),
        invalidateCreditBalance(queryClient),
        queryClient.invalidateQueries({ queryKey: ['templates'], exact: false }),
      ]);
    },
    onError: (err) => {
      console.error('rollbackTemplate error:', err.message);
      notification.error(err.message || 'Не удалось откатить шаблон');
    },
  });

  // Wrapper functions for convenient usage
  const createTemplate = useCallback(
    (templateData) => createTemplateMutation.mutateAsync(templateData),
    [createTemplateMutation]
  );

  const updateTemplate = useCallback(
    (templateId, updates) => updateTemplateMutation.mutateAsync({ templateId, updates }),
    [updateTemplateMutation]
  );

  const deleteTemplate = useCallback(
    (templateId) => deleteTemplateMutation.mutateAsync(templateId),
    [deleteTemplateMutation]
  );

  const applyTemplate = useCallback(
    (templateId, weekStartDate) => applyTemplateMutation.mutateAsync({ templateId, weekStartDate }),
    [applyTemplateMutation]
  );

  const rollbackTemplate = useCallback(
    (templateId, weekStartDate) => rollbackTemplateMutation.mutateAsync({ templateId, weekStartDate }),
    [rollbackTemplateMutation]
  );

  const refresh = useCallback(() => {
    refetchQuery();
  }, [refetchQuery]);

  const error = queryError?.message || null;

  return {
    templates,
    loading,
    error,
    hasError: !!error,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    applyTemplate,
    rollbackTemplate,
    refresh,
    isCreating: createTemplateMutation.isPending,
    isUpdating: updateTemplateMutation.isPending,
    isDeleting: deleteTemplateMutation.isPending,
    isApplying: applyTemplateMutation.isPending,
    isRollingBack: rollbackTemplateMutation.isPending,
  };
};

/**
 * Custom hook for fetching a single template
 */
export const useTemplate = (templateId) => {
  const notification = useNotification();

  const {
    data: template = null,
    isLoading: loading,
    error: queryError,
    refetch,
  } = useQuery({
    queryKey: ['templates', templateId],
    queryFn: async () => {
      logger.debug('useQuery: Fetching template:', templateId);
      try {
        const data = await templatesAPI.getTemplate(templateId);
        logger.debug('useQuery: Successfully loaded template:', data);
        return data;
      } catch (err) {
        console.error('useQuery error:', err.message);
        notification.error('Не удалось загрузить шаблон');
        throw err;
      }
    },
    enabled: !!templateId,
  });

  const error = queryError?.message || null;

  return {
    template,
    loading,
    error,
    hasError: !!error,
    refetch,
  };
};

export default useTemplates;
