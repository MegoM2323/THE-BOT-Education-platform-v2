import { useCallback, useState, useRef, useEffect } from "react";
import { logger } from "../utils/logger.js";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import * as creditsAPI from "../api/credits.js";
import { useNotification } from "./useNotification.js";

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

const normalizeHistoryResponse = (response) => {
  if (Array.isArray(response)) {
    return response;
  }

  if (response?.data && Array.isArray(response.data)) {
    return response.data;
  }

  if (response?.transactions && Array.isArray(response.transactions)) {
    return response.transactions;
  }

  return [];
};

export const useStudentCreditsHistory = (userId, options = {}) => {
  const { refetchInterval = false, ...queryOptions } = options;
  const notification = useNotification();
  const queryClient = useQueryClient();
  const historyRefetchRef = useRef(null);

  const maskedUserId = userId ? "****" + userId.slice(-4) : "****";
  logger.debug("useStudentCreditsHistory hook initialized", {
    userId: maskedUserId,
    refetchInterval,
    hasOptions: Object.keys(queryOptions).length > 0,
  });

  const [historyFilters, setHistoryFilters] = useState({});

  // Сбрасываем дополнительные фильтры при изменении userId
  useEffect(() => {
    setHistoryFilters({});
  }, [userId]);

  const {
    data: historyData = [],
    isLoading: historyLoading,
    error: historyQueryError,
    refetch: refetchHistoryQuery,
  } = useQuery({
    queryKey: ["credits", "student-history", userId],
    queryFn: async () => {
      if (!userId) {
        return [];
      }
      logger.debug(
        "useStudentCreditsHistory: Fetching history with userId:",
        maskedUserId,
      );
      try {
        // Для базового запроса используем только user_id без дополнительных фильтров
        const data = await creditsAPI.getHistory({ user_id: userId });
        logger.debug(
          "useStudentCreditsHistory: History received for userId:",
          maskedUserId,
        );
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
          `useStudentCreditsHistory error [${errorInfo.category}]:`,
          errorInfo.message,
        );
        throw err;
      }
    },
    staleTime: 30000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchInterval: refetchInterval,
    enabled: !!userId,
    retry: (failureCount, error) => {
      const errorInfo = categorizeError(error);

      if (
        errorInfo.category === "unauthorized" ||
        errorInfo.category === "forbidden"
      ) {
        logger.warn(
          `useStudentCreditsHistory retry stopped: ${errorInfo.category} error`,
        );
        return false;
      }

      if (failureCount < 3) {
        if (errorInfo.category === "network_error") {
          logger.warn(
            `useStudentCreditsHistory: Network error retry attempt ${failureCount + 1}/3`,
          );
        } else {
          logger.warn(
            `useStudentCreditsHistory: Retry attempt ${failureCount + 1}/3 for ${errorInfo.category}`,
          );
        }
        return true;
      }

      logger.error(
        `useStudentCreditsHistory: Max retries (3) reached for ${errorInfo.category}`,
      );
      return false;
    },
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 30000),
    ...queryOptions,
  });

  const history = !userId ? [] : normalizeHistoryResponse(historyData);
  const historyPagination = !userId ? null : (historyData?.pagination || null);

  historyRefetchRef.current = refetchHistoryQuery;

  const fetchHistory = useCallback(
    async (filters = {}) => {
      if (!userId) {
        return [];
      }
      const mergedFilters = { user_id: userId, ...filters };
      setHistoryFilters(mergedFilters);

      await new Promise((resolve) => setTimeout(resolve, 0));
      return historyRefetchRef.current();
    },
    [userId],
  );

  const refresh = useCallback(() => {
    if (!userId) {
      return;
    }
    queryClient.invalidateQueries({
      queryKey: ["credits", "student-history", userId],
    });
  }, [queryClient, userId]);

  const historyError = !userId
    ? {
        category: "missing_user_id",
        userMessage: 'ID студента не указан',
        message: 'Missing userId parameter'
      }
    : historyQueryError
    ? categorizeError(historyQueryError)
    : null;

  const getUserBalance = useCallback(async () => {
    if (!userId) {
      return null;
    }
    try {
      logger.debug("useStudentCreditsHistory: Fetching user balance");
      const data = await creditsAPI.getUserCredits(userId);
      return data?.balance ?? 0;
    } catch (error) {
      const errorInfo = categorizeError(error);
      logger.error(
        `useStudentCreditsHistory: Failed to fetch user balance [${errorInfo.category}]`,
      );
      if (notification?.error) {
        notification.error(errorInfo.userMessage || "Ошибка загрузки баланса");
      }
      throw error;
    }
  }, [userId, notification]);

  return {
    history,
    historyPagination,
    historyLoading: !userId ? false : historyLoading,
    historyError,
    fetchHistory,
    refresh,
    getUserBalance,
  };
};

export default useStudentCreditsHistory;
