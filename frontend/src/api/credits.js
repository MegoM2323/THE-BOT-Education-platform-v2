/**
 * API эндпоинты кредитов
 *
 * Все функции работают с нормализованным ответом от apiClient,
 * который автоматически извлекает data из { success, data } обёртки.
 *
 * @see ../api/client.js request() function для понимания трансформации ответов
 */

import apiClient from "./client.js";
import { logger } from "../utils/logger.js";

/**
 * Get current user's credit balance
 *
 * Endpoint: GET /credits
 *
 * Backend Response Format:
 * ```json
 * {
 *   "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *   "balance": 10,
 *   "created_at": "2025-11-26T10:30:00Z",
 *   "updated_at": "2025-11-26T10:30:00Z"
 * }
 * ```
 *
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<{balance: number}>} User's credit balance
 */
export const getCredits = async (options = {}) => {
  logger.debug("[Credits API] Fetching student credits...");
  try {
    const response = await apiClient.get("/credits", options);

    if (response && typeof response.balance === "number") {
      logger.debug("[Credits API] Received credits:", {
        balance: response.balance,
      });
    } else {
      logger.warn(
        "[Credits API] Response missing balance field, using default value",
      );
    }

    return {
      balance: response?.balance ?? 0,
    };
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to fetch credits:", {
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

/**
 * Get all students' credit balances (admin only)
 *
 * Endpoint: GET /credits/all
 *
 * Backend Response Format:
 * ```json
 * {
 *   "balances": [
 *     {
 *       "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *       "balance": 10
 *     }
 *   ]
 * }
 * ```
 *
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<{balances: Array<{user_id: string, balance: number}>}>}
 */
export const getAllCredits = async (options = {}) => {
  logger.debug("[Credits API] Fetching all students credits...");
  try {
    const response = await apiClient.get("/credits/all", options);

    let balances = [];
    let format = "unknown";

    logger.debug("[Credits API] Response format analysis:", {
      isArray: Array.isArray(response),
      hasData: !!response?.data,
      hasBalances: !!response?.balances,
      dataType: typeof response?.data,
    });

    if (Array.isArray(response)) {
      balances = response;
      format = "array";
    } else if (response?.data) {
      if (Array.isArray(response.data)) {
        balances = response.data;
        format = "paginated";
      } else if (response.data?.data && Array.isArray(response.data.data)) {
        balances = response.data.data;
        format = "nested_paginated";
      }
    } else if (response?.balances && Array.isArray(response.balances)) {
      balances = response.balances;
      format = "balances";
    }

    logger.debug("[Credits API] Received all credits response:", {
      format,
      count: balances.length,
      hasMeta: !!response?.meta,
    });

    return {
      balances,
    };
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to fetch all credits:", {
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

/**
 * Get specific user's credit balance (admin only)
 *
 * Endpoint: GET /credits/user/{id}
 *
 * Backend Response Format:
 * ```json
 * {
 *   "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *   "balance": 10,
 *   "created_at": "2025-11-26T10:30:00Z",
 *   "updated_at": "2025-11-26T10:30:00Z"
 * }
 * ```
 *
 * @param {string} userId - User ID (UUID)
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<{balance: number}>} User's credit balance
 */
export const getUserCredits = async (userId, options = {}) => {
  const maskedUserId = userId ? "****" + userId.slice(-4) : "****";
  logger.debug("[Credits API] Fetching user credits for:", {
    userId: maskedUserId,
  });
  try {
    const response = await apiClient.get(`/credits/user/${userId}`, options);

    if (response && typeof response.balance === "number") {
      logger.debug("[Credits API] Received user credits:", {
        userId: maskedUserId,
        balance: response.balance,
      });
    } else {
      logger.warn("[Credits API] Response missing balance field for user:", {
        userId: maskedUserId,
      });
    }

    return {
      balance: response?.balance ?? 0,
    };
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to fetch user credits:", {
        userId: maskedUserId,
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

/**
 * Add credits to a user account (admin only)
 *
 * Endpoint: POST /users/{id}/credits
 *
 * Backend Response Format:
 * ```json
 * {
 *   "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *   "balance": 20,
 *   "transaction_id": "850e8400-e29b-41d4-a716-446655440003"
 * }
 * ```
 *
 * @param {string} userId - User ID (UUID)
 * @param {number} amount - Number of credits to add
 * @param {string} [description=''] - Transaction reason/description
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<{user_id: string, balance: number, transaction_id: string}>} Updated balance
 */
export const addCredits = async (
  userId,
  amount,
  description = "",
  options = {},
) => {
  const maskedUserId = userId ? "****" + userId.slice(-4) : "****";
  logger.debug("[Credits API] Adding credits:", {
    userId: maskedUserId,
    amount,
    description,
  });
  try {
    const response = await apiClient.post(
      `/users/${userId}/credits`,
      {
        amount,
        reason: description,
      },
      options,
    );

    logger.debug("[Credits API] Successfully added credits:", {
      userId: maskedUserId,
      newBalance: response?.balance,
      transactionId: response?.transaction_id,
    });

    return response;
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to add credits:", {
        userId: maskedUserId,
        amount,
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

/**
 * Deduct credits from a user account (admin only)
 *
 * Endpoint: POST /users/{id}/credits with negative amount
 *
 * Backend Response Format:
 * ```json
 * {
 *   "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *   "balance": 5,
 *   "transaction_id": "850e8400-e29b-41d4-a716-446655440004"
 * }
 * ```
 *
 * @param {string} userId - User ID (UUID)
 * @param {number} amount - Number of credits to deduct (positive number)
 * @param {string} [description=''] - Transaction reason/description
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<{user_id: string, balance: number, transaction_id: string}>} Updated balance
 */
export const deductCredits = async (
  userId,
  amount,
  description = "",
  options = {},
) => {
  const maskedUserId = userId ? "****" + userId.slice(-4) : "****";
  logger.debug("[Credits API] Deducting credits:", {
    userId: maskedUserId,
    amount,
    description,
  });
  try {
    const response = await apiClient.post(
      `/users/${userId}/credits`,
      {
        amount: -amount,
        reason: description,
      },
      options,
    );

    logger.debug("[Credits API] Successfully deducted credits:", {
      userId: maskedUserId,
      newBalance: response?.balance,
      transactionId: response?.transaction_id,
    });

    return response;
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to deduct credits:", {
        userId: maskedUserId,
        amount,
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

/**
 * Get credit transaction history with optional filters
 *
 * Endpoint: GET /credits/history
 *
 * Backend Response Format:
 * ```json
 * [
 *   {
 *     "id": "850e8400-e29b-41d4-a716-446655440001",
 *     "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *     "amount": -1,
 *     "operation_type": "deduct",
 *     "reason": "Booking for lesson",
 *     "booking_id": "750e8400-e29b-41d4-a716-446655440001",
 *     "created_at": "2025-11-20T10:00:00Z"
 *   }
 * ]
 * ```
 *
 * @param {Object} [filters={}] - Optional filter parameters
 * @param {string} [filters.user_id] - Filter by user ID (admin only)
 * @param {string} [filters.start_date] - Filter by start date (YYYY-MM-DD)
 * @param {string} [filters.end_date] - Filter by end date (YYYY-MM-DD)
 * @param {string} [filters.type] - Filter by operation type (add/deduct/refund)
 * @param {number} [filters.limit] - Number of records per page
 * @param {number} [filters.offset] - Pagination offset
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Array of transaction objects
 */
export const getHistory = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  if (filters.user_id) params.append("user_id", filters.user_id);
  if (filters.start_date) params.append("start_date", filters.start_date);
  if (filters.end_date) params.append("end_date", filters.end_date);
  if (filters.type) params.append("type", filters.type);
  if (filters.limit) params.append("limit", String(filters.limit));
  if (filters.offset) params.append("offset", String(filters.offset));

  const query = params.toString();

  logger.debug("[Credits API] Fetching credit history (admin):", {
    hasFilters: Object.keys(filters).length > 0,
  });

  try {
    const response = await apiClient.get(
      `/credits/history${query ? `?${query}` : ""}`,
      options,
    );
    // Backend может возвращать разные форматы в зависимости от версии API:
    // - Массив напрямую (legacy): [...]
    // - Объект с transactions (v1): { transactions: [...] }
    // - Объект с history (v1-alt): { history: [...] }
    // - Пагинированный ответ (v2): { data: [...], meta: {...} }

    let history;
    let format;
    if (Array.isArray(response)) {
      history = response;
      format = "array";
    } else {
      history =
        response?.transactions || response?.history || response?.data || [];
      format = response?.transactions
        ? "transactions"
        : response?.history
          ? "history"
          : response?.data
            ? "paginated"
            : "unknown";
    }

    logger.debug("[Credits API] Received history response:", {
      format,
      count: history.length,
    });

    return history;
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to fetch credit history:", {
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

/**
 * Get current user's credit transaction history
 *
 * Endpoint: GET /credits/history
 *
 * Backend Response Format:
 * ```json
 * [
 *   {
 *     "id": "850e8400-e29b-41d4-a716-446655440001",
 *     "user_id": "550e8400-e29b-41d4-a716-446655440002",
 *     "amount": -1,
 *     "operation_type": "deduct",
 *     "reason": "Booking for lesson",
 *     "booking_id": "750e8400-e29b-41d4-a716-446655440001",
 *     "created_at": "2025-11-20T10:00:00Z"
 *   }
 * ]
 * ```
 *
 * @param {Object} [filters={}] - Optional filter parameters
 * @param {string} [filters.start_date] - Filter by start date (YYYY-MM-DD)
 * @param {string} [filters.end_date] - Filter by end date (YYYY-MM-DD)
 * @param {string} [filters.type] - Filter by operation type (add/deduct/refund)
 * @param {number} [filters.limit] - Number of records per page
 * @param {number} [filters.offset] - Pagination offset
 * @param {Object} [options] - Опции запроса (включая signal для отмены)
 * @returns {Promise<Array>} Array of transaction objects
 */
export const getMyHistory = async (filters = {}, options = {}) => {
  const params = new URLSearchParams();

  if (filters.start_date) params.append("start_date", filters.start_date);
  if (filters.end_date) params.append("end_date", filters.end_date);
  if (filters.type) params.append("type", filters.type);
  if (filters.limit) params.append("limit", String(filters.limit));
  if (filters.offset) params.append("offset", String(filters.offset));

  const query = params.toString();

  logger.debug("[Credits API] Fetching my credit history:", {
    hasFilters: Object.keys(filters).length > 0,
  });

  try {
    const response = await apiClient.get(
      `/credits/history${query ? `?${query}` : ""}`,
      options,
    );
    // Backend может возвращать разные форматы в зависимости от версии API:
    // - Массив напрямую (legacy): [...]
    // - Объект с transactions (v1): { transactions: [...] }
    // - Объект с history (v1-alt): { history: [...] }
    // - Пагинированный ответ (v2): { data: [...], meta: {...} }

    let history;
    let format;
    if (Array.isArray(response)) {
      history = response;
      format = "array";
    } else {
      history =
        response?.transactions || response?.history || response?.data || [];
      format = response?.transactions
        ? "transactions"
        : response?.history
          ? "history"
          : response?.data
            ? "paginated"
            : "unknown";
    }

    logger.debug("[Credits API] Received my history response:", {
      format,
      count: history.length,
    });

    return history;
  } catch (error) {
    if (error.name !== "AbortError") {
      logger.error("[Credits API] Failed to fetch my credit history:", {
        status: error.status || error.response?.status,
        message: error.message,
        name: error.name,
      });
    }
    throw error;
  }
};

export default {
  getCredits,
  getAllCredits,
  getUserCredits,
  addCredits,
  deductCredits,
  getHistory,
  getMyHistory,
};
