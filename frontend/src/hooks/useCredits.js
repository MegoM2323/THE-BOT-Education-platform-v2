import { useCallback, useState, useRef, useEffect } from "react";
import { logger } from "../utils/logger.js";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import * as creditsAPI from "../api/credits.js";
import { useNotification } from "./useNotification.js";

/**
 * Категоризирует ошибки для правильной обработки и логирования
 * Различает:
 * - Ошибки аутентификации (требуют повторного входа)
 * - Ошибки сети (требуют retry)
 * - Ошибки сервера (требуют уведомления пользователя)
 *
 * @param {Error} error - Объект ошибки
 * @returns {Object} Объект с категорией и описанием ошибки
 */
const categorizeError = (error) => {
  if (!error) return { category: "unknown", message: "Неизвестная ошибка" };

  const status = error?.status || error?.response?.status;
  const code = error?.code;
  const message = error?.message || "";

  if (status === 401) {
    return {
      category: "unauthorized",
      message: "Требуется повторный вход",
      userMessage: "Ваша сессия истекла. Пожалуйста, авторизуйтесь заново.",
    };
  }

  if (status === 403) {
    return {
      category: "forbidden",
      message: "Доступ запрещен",
      userMessage: "У вас нет прав для доступа к этому ресурсу.",
    };
  }

  if (status >= 500) {
    return {
      category: "server_error",
      message: `Ошибка сервера (${status})`,
      userMessage: "Сервер временно недоступен. Пожалуйста, попробуйте позже.",
    };
  }

  if (code === "NETWORK_ERROR" || message.toLowerCase().includes("network")) {
    return {
      category: "network_error",
      message: "Ошибка сети",
      userMessage: "Проблема с подключением. Попробуем снова автоматически.",
    };
  }

  if (status >= 400 && status < 500) {
    return {
      category: "client_error",
      message: `Ошибка запроса (${status})`,
      userMessage: "Некорректный запрос. Пожалуйста, попробуйте позже.",
    };
  }

  return {
    category: "unknown",
    message: message || "Неизвестная ошибка",
    userMessage:
      "Произошла ошибка при загрузке данных. Пожалуйста, попробуйте позже.",
  };
};

/**
 * Нормализует различные форматы ответа истории кредитов
 * Поддерживает:
 * 1. Массив: [...]
 * 2. Объект с data: { data: [...], pagination: {...} }
 * 3. Объект с transactions: { transactions: [...] }
 *
 * @param {any} response - Ответ от API
 * @returns {Array} Нормализованный массив истории
 */
const normalizeHistoryResponse = (response) => {
  // Обработка прямого массива
  if (Array.isArray(response)) {
    return response;
  }

  // Обработка объекта с data (пагинация)
  if (response?.data && Array.isArray(response.data)) {
    return response.data;
  }

  // Обработка объекта с transactions (трансакции)
  if (response?.transactions && Array.isArray(response.transactions)) {
    return response.transactions;
  }

  // Fallback: пустой массив
  return [];
};

/**
 * Кастомный хук для управления кредитами с React Query интеграцией
 * Автоматически участвует в query invalidation инфраструктуре
 *
 * @param {Object} options - Опции конфигурации
 * @param {number|false} options.refetchInterval - Интервал автоматической переполучения (мс) или false для отключения
 * @param {Object} options - Дополнительные опции для useQuery
 * @returns {Object} Объект с балансом, историей, методами и состоянием
 */
export const useCredits = (options = {}) => {
  const { refetchInterval = false, ...queryOptions } = options;
  const notification = useNotification();
  const queryClient = useQueryClient();
  const broadcastChannelRef = useRef(null);

  logger.debug("useCredits hook initialized", {
    refetchInterval,
    hasOptions: Object.keys(queryOptions).length > 0,
  });

  useEffect(() => {
    const initBroadcastChannel = () => {
      if (!broadcastChannelRef.current) {
        try {
          broadcastChannelRef.current = new BroadcastChannel("credits-updates");
          logger.debug("BroadcastChannel initialized for credits sync");

          const handleMessage = (event) => {
            if (event.data?.type === "CREDITS_UPDATED") {
              logger.debug(
                "BroadcastChannel: Received CREDITS_UPDATED from another tab",
                { userId: event.data.userId },
              );
              queryClient.invalidateQueries({
                queryKey: ["credits"],
                exact: true,
              });
              queryClient.invalidateQueries({
                queryKey: ["credits", "history"],
              });
            }
          };

          broadcastChannelRef.current.addEventListener(
            "message",
            handleMessage,
          );

          return () => {
            if (broadcastChannelRef.current) {
              broadcastChannelRef.current.removeEventListener(
                "message",
                handleMessage,
              );
            }
          };
        } catch (err) {
          logger.warn("BroadcastChannel not supported in this browser", err);
        }
      }
    };

    const cleanup = initBroadcastChannel();

    return () => {
      if (cleanup) {
        cleanup();
      }
      if (broadcastChannelRef.current) {
        broadcastChannelRef.current.close();
        broadcastChannelRef.current = null;
      }
    };
  }, [queryClient]);

  // useQuery для получения текущего баланса
  const {
    data: creditsData,
    isLoading: loading,
    error: queryError,
    refetch: refetchCredits,
  } = useQuery({
    queryKey: ["credits"],
    queryFn: async () => {
      logger.debug("useQuery: Fetching credits balance");
      try {
        const data = await creditsAPI.getCredits();
        logger.debug("useQuery: Credits data received:", data);
        return data;
      } catch (err) {
        const errorInfo = categorizeError(err);
        logger.error(
          `useQuery credits error [${errorInfo.category}]:`,
          errorInfo.message,
        );
        throw err;
      }
    },
    staleTime: 30000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchInterval: refetchInterval,
    retry: (failureCount, error) => {
      const errorInfo = categorizeError(error);

      if (
        errorInfo.category === "unauthorized" ||
        errorInfo.category === "forbidden"
      ) {
        logger.warn(`useQuery retry stopped: ${errorInfo.category} error`);
        return false;
      }

      if (failureCount < 3) {
        if (errorInfo.category === "network_error") {
          logger.warn(
            `useQuery: Network error retry attempt ${failureCount + 1}/3`,
          );
        } else {
          logger.warn(
            `useQuery: Retry attempt ${failureCount + 1}/3 for ${errorInfo.category}`,
          );
        }
        return true;
      }

      logger.error(
        `useQuery: Max retries (3) reached for ${errorInfo.category}`,
      );
      return false;
    },
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 30000),
    ...queryOptions,
  });

  // State to track if history has been requested
  const [historyRequested, setHistoryRequested] = useState(false);
  // State to track pagination filters
  const [historyFilters, setHistoryFilters] = useState({});

  // useQuery для истории кредитов
  const {
    data: historyData = [],
    isLoading: historyLoading,
    error: historyQueryError,
    refetch: refetchHistoryQuery,
  } = useQuery({
    queryKey: ["credits", "history", historyFilters],
    queryFn: async () => {
      logger.debug(
        "useQuery: Fetching credits history with filters:",
        historyFilters,
      );
      try {
        const data = await creditsAPI.getMyHistory(historyFilters);
        logger.debug("useQuery: Credits history received:", data);
        if (Array.isArray(data)) {
          return data;
        }
        if (Array.isArray(data?.transactions)) {
          return data.transactions;
        }
        const historyArray = Array.isArray(data?.data) ? data.data : [];
        if (data?.pagination) {
          return {
            data: historyArray,
            pagination: data.pagination,
          };
        }
        return historyArray;
      } catch (err) {
        const errorInfo = categorizeError(err);
        logger.error(
          `useQuery history error [${errorInfo.category}]:`,
          errorInfo.message,
        );
        throw err;
      }
    },
    enabled: historyRequested,
    staleTime: 30000,
    gcTime: 5 * 60 * 1000,
    retry: (failureCount, error) => {
      const errorInfo = categorizeError(error);

      if (
        errorInfo.category === "unauthorized" ||
        errorInfo.category === "forbidden"
      ) {
        logger.warn(
          `useQuery history retry stopped: ${errorInfo.category} error`,
        );
        return false;
      }

      if (failureCount < 3) {
        if (errorInfo.category === "network_error") {
          logger.warn(
            `useQuery history: Network error retry attempt ${failureCount + 1}/3`,
          );
        } else {
          logger.warn(
            `useQuery history: Retry attempt ${failureCount + 1}/3 for ${errorInfo.category}`,
          );
        }
        return true;
      }

      logger.error(
        `useQuery history: Max retries (3) reached for ${errorInfo.category}`,
      );
      return false;
    },
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 30000),
  });

  // Нормализуем историю (обработка всех форматов ответа API)
  const history = normalizeHistoryResponse(historyData);

  // Получаем информацию о пагинации если доступна
  const historyPagination = historyData?.pagination || null;

  // Store refetch function ref for use in fetchHistory
  const historyRefetchRef = useRef(refetchHistoryQuery);
  // Update ref when refetch function changes
  historyRefetchRef.current = refetchHistoryQuery;

  // Mutation для добавления кредитов (используется админом)
  const addCreditsMutation = useMutation({
    mutationFn: async ({ userId, amount, description }) => {
      logger.debug(
        "addCredits: Starting with userId:",
        userId,
        "amount:",
        amount,
      );
      try {
        const result = await creditsAPI.addCredits(userId, amount, description);
        logger.debug("addCredits: Success");
        return result;
      } catch (err) {
        const errorInfo = categorizeError(err);
        logger.error(
          `addCredits error [${errorInfo.category}]:`,
          errorInfo.message,
        );
        throw err;
      }
    },
  });

  // Mutation для списания кредитов (используется админом)
  const deductCreditsMutation = useMutation({
    mutationFn: async ({ userId, amount, description }) => {
      logger.debug(
        "deductCredits: Starting with userId:",
        userId,
        "amount:",
        amount,
      );
      try {
        const result = await creditsAPI.deductCredits(
          userId,
          amount,
          description,
        );
        logger.debug("deductCredits: Success");
        return result;
      } catch (err) {
        const errorInfo = categorizeError(err);
        logger.error(
          `deductCredits error [${errorInfo.category}]:`,
          errorInfo.message,
        );
        throw err;
      }
    },
  });

  // Обертки для удобного использования
  const addCredits = useCallback(
    async (userId, amount, description) => {
      try {
        logger.debug("addCredits wrapper: Starting transaction", {
          userId,
          amount,
        });
        const result = await addCreditsMutation.mutateAsync({
          userId,
          amount,
          description,
        });
        if (notification?.success) {
          notification.success(`Добавлено ${amount} кредитов`);
        }
        await queryClient.invalidateQueries({
          queryKey: ["credits"],
          exact: true,
        });
        if (broadcastChannelRef.current) {
          try {
            broadcastChannelRef.current.postMessage({
              type: "CREDITS_UPDATED",
              userId,
              newBalance: result?.balance,
              timestamp: Date.now(),
            });
            logger.debug("addCredits: BroadcastChannel message sent");
          } catch (err) {
            logger.warn(
              "addCredits: Failed to send BroadcastChannel message",
              err,
            );
          }
        }
        logger.debug("addCredits wrapper: Success");
        return result;
      } catch (error) {
        const errorInfo = categorizeError(error);
        logger.error(`addCredits wrapper error [${errorInfo.category}]`);
        if (notification?.error) {
          notification.error(
            errorInfo.userMessage || "Ошибка добавления кредитов",
          );
        }
        throw error;
      }
    },
    [addCreditsMutation, queryClient, notification],
  );

  const deductCredits = useCallback(
    async (userId, amount, description) => {
      try {
        logger.debug("deductCredits wrapper: Starting transaction", {
          userId,
          amount,
        });
        const result = await deductCreditsMutation.mutateAsync({
          userId,
          amount,
          description,
        });
        if (notification?.success) {
          notification.success(`Списано ${amount} кредитов`);
        }
        await queryClient.invalidateQueries({
          queryKey: ["credits"],
          exact: true,
        });
        if (broadcastChannelRef.current) {
          try {
            broadcastChannelRef.current.postMessage({
              type: "CREDITS_UPDATED",
              userId,
              newBalance: result?.balance,
              timestamp: Date.now(),
            });
            logger.debug("deductCredits: BroadcastChannel message sent");
          } catch (err) {
            logger.warn(
              "deductCredits: Failed to send BroadcastChannel message",
              err,
            );
          }
        }
        logger.debug("deductCredits wrapper: Success");
        return result;
      } catch (error) {
        const errorInfo = categorizeError(error);
        logger.error(`deductCredits wrapper error [${errorInfo.category}]`);
        if (notification?.error) {
          notification.error(
            errorInfo.userMessage || "Ошибка списания кредитов",
          );
        }
        throw error;
      }
    },
    [deductCreditsMutation, queryClient, notification],
  );

  const fetchCreditsCallback = useCallback(
    () => refetchCredits(),
    [refetchCredits],
  );

  const fetchHistoryCallback = useCallback(
    async (filters) => {
      // Update filters with pagination parameters
      if (filters) {
        setHistoryFilters(filters);
      }

      // Enable history query on first fetch
      if (!historyRequested) {
        setHistoryRequested(true);
        // Wait a tick for the query to be enabled, then refetch
        await new Promise((resolve) => setTimeout(resolve, 0));
        return historyRefetchRef.current();
      } else {
        // Already enabled, just refetch
        return historyRefetchRef.current();
      }
    },
    [historyRequested],
  );

  const refresh = useCallback(() => {
    // Инвалидировать кэш баланса для обновления всех компонентов
    queryClient.invalidateQueries({ queryKey: ["credits"], exact: true });
    // Only refresh history if it was already requested
    if (historyRequested) {
      queryClient.invalidateQueries({ queryKey: ["credits", "history"] });
    }
  }, [queryClient, historyRequested]);

  const error = queryError ? categorizeError(queryError) : null;
  const historyError = historyQueryError
    ? categorizeError(historyQueryError)
    : null;

  const balance = creditsData?.balance ?? 0;
  if (creditsData && typeof creditsData.balance !== "number") {
    logger.debug(
      "useCredits: API response missing balance field, using default: 0",
    );
  }

  return {
    balance,
    history,
    historyPagination,
    loading,
    error,
    historyLoading,
    historyError,
    fetchCredits: fetchCreditsCallback,
    fetchHistory: fetchHistoryCallback,
    addCredits,
    deductCredits,
    refresh,
    refetch: fetchCreditsCallback, // Алиас для совместимости с useSidebarCredits
    isAdding: addCreditsMutation.isPending,
    isDeducting: deductCreditsMutation.isPending,
  };
};

/**
 * Утилита для инвалидации баланса кредитов
 * Используется для синхронизации состояния между компонентами
 *
 * @param {QueryClient} queryClient - React Query client
 */
export const invalidateCreditBalance = (queryClient) => {
  queryClient.invalidateQueries({
    queryKey: ["credits"],
    exact: true, // Инвалидировать только ['credits'], НЕ ['credits', 'history']
  });
};

export default useCredits;
