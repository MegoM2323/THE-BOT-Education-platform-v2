import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { logger } from '../utils/logger.js';
import * as paymentsAPI from '../api/payments.js';
import { useNotification } from './useNotification.js';

/**
 * Hook для работы с платежами
 */
export const usePayments = () => {
  const notification = useNotification();
  const queryClient = useQueryClient();

  // Получение истории платежей
  const {
    data: payments = [],
    isLoading: isLoadingHistory,
    error: historyError,
    refetch: refetchHistory,
  } = useQuery({
    queryKey: ['payments', 'history'],
    queryFn: async () => {
      try {
        const data = await paymentsAPI.getPaymentHistory();
        return data;
      } catch (err) {
        console.error('usePayments error:', err);
        notification.error('Ошибка загрузки истории платежей');
        throw err;
      }
    },
    staleTime: 60000, // 1 минута
    refetchOnWindowFocus: true,
  });

  // Создание платежа
  const createPaymentMutation = useMutation({
    mutationFn: async (credits) => {
      return await paymentsAPI.createPayment(credits);
    },
    onSuccess: () => {
      // Обновляем историю платежей
      queryClient.invalidateQueries({ queryKey: ['payments', 'history'] });
      // Инвалидируем баланс кредитов после успешного платежа
      queryClient.invalidateQueries({ queryKey: ['credits'], exact: true });
      // Платеж создан, redirect произойдет в компоненте
    },
    onError: (error) => {
      console.error('Create payment error:', error);
      notification.error(error.message || 'Ошибка создания платежа');
    },
  });

  return {
    payments,
    isLoadingHistory,
    historyError,
    refetchHistory,
    createPayment: createPaymentMutation.mutateAsync,
    isCreatingPayment: createPaymentMutation.isPending,
  };
};

export default usePayments;
