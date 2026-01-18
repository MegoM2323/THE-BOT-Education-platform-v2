import { useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as paymentSettingsAPI from '../api/paymentSettings.js';
import { useNotification } from './useNotification.js';

/**
 * Hook для управления настройками платежей студентов (admin only)
 */
export const usePaymentSettings = () => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Fetch students with payment status
  const {
    data: students = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['paymentSettings'],
    queryFn: async () => {
      try {
        const response = await paymentSettingsAPI.getStudentsPaymentStatus();
        return Array.isArray(response) ? response : [];
      } catch (err) {
        console.error('usePaymentSettings error:', err);
        throw err;
      }
    },
    staleTime: 60000, // 1 minute
    refetchOnWindowFocus: true,
  });

  // Update payment status mutation
  const updateMutation = useMutation({
    mutationFn: async ({ userId, enabled }) => {
      const response = await paymentSettingsAPI.updatePaymentStatus(userId, enabled);
      return { userId, enabled, response };
    },
    onMutate: async ({ userId, enabled }) => {
      // Cancel outgoing queries
      await queryClient.cancelQueries({ queryKey: ['paymentSettings'] });

      // Snapshot previous value
      const previousStudents = queryClient.getQueryData(['paymentSettings']);

      // Optimistically update
      queryClient.setQueryData(['paymentSettings'], (old) => {
        if (!Array.isArray(old)) return old;
        return old.map((student) =>
          student.id === userId
            ? { ...student, payment_enabled: enabled, updated_at: new Date().toISOString() }
            : student
        );
      });

      return { previousStudents };
    },
    onSuccess: (data) => {
      const status = data.enabled ? 'включены' : 'отключены';
      notification.success(`Платежи ${status}`);
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['paymentSettings'] });
    },
    onError: (err, variables, context) => {
      // Rollback on error
      if (context?.previousStudents) {
        queryClient.setQueryData(['paymentSettings'], context.previousStudents);
      }

      const errorMessage = err.message || 'Ошибка обновления статуса платежей';
      notification.error(errorMessage);
    },
  });

  const updatePaymentStatus = useCallback(
    (userId, enabled) => updateMutation.mutateAsync({ userId, enabled }),
    [updateMutation]
  );

  return {
    students,
    isLoading,
    error,
    refetch,
    updatePaymentStatus,
    isUpdating: updateMutation.isPending,
  };
};

export default usePaymentSettings;
