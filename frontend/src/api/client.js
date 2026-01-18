/**
 * Базовый API клиент с обёрткой fetch и обработкой ошибок
 */

import { logger } from '../utils/logger.js';

// Определяем базовый URL для API
// Используем переменную окружения VITE_API_URL если доступна
// Иначе используем относительный путь (через Vite прокси в dev) или текущий хост для production
const getAPIBaseURL = () => {
  // Проверяем переменную окружения VITE_API_URL
  if (import.meta?.env?.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }

  // Для development используем относительный URL (работает через Vite прокси)
  if (import.meta?.env?.DEV) {
    return '/api/v1';
  }

  // Для тестов используем mock URL если window не доступен
  if (!globalThis.window || !globalThis.window.location) {
    return '/api/v1';
  }

  // Для production используем текущий хост
  return `${window.location.origin}/api/v1`;
};

const API_BASE_URL = getAPIBaseURL();

// CSRF token хранится в памяти (не в localStorage для безопасности)
let csrfToken = null;

// Флаг для предотвращения множественных попыток refresh
let isRefreshingSession = false;
let refreshPromise = null;

/**
 * DEPRECATED: Token storage removed for security (XSS protection)
 * Auth now uses httpOnly cookies set by backend
 * These functions kept for API compatibility but do nothing
 */
export const getStoredToken = () => {
  // Token stored in httpOnly cookie, not accessible via JS
  return null;
};

/**
 * DEPRECATED: Token storage removed for security (XSS protection)
 * @deprecated Backend sets httpOnly cookie automatically
 */
// eslint-disable-next-line no-unused-vars
export const saveToken = (token) => {
  // No-op: Backend sets httpOnly cookie, frontend doesn't store tokens
  logger.debug('saveToken called but ignored - using httpOnly cookies');
};

/**
 * DEPRECATED: Token storage removed for security (XSS protection)
 * @deprecated Backend clears httpOnly cookie on logout
 */
export const clearToken = () => {
  // No-op: Backend clears cookie, frontend doesn't manage tokens
  logger.debug('clearToken called but ignored - using httpOnly cookies');
};

/**
 * Получить CSRF token с сервера
 * Вызывается после успешного логина
 */
export const fetchCSRFToken = async () => {
  try {
    const url = `${API_BASE_URL}/csrf-token`;
    logger.debug('Fetching CSRF token from:', url);

    const response = await fetch(url, {
      method: 'GET',
      credentials: 'include', // Важно для session cookies
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      logger.error('Failed to fetch CSRF token:', response.status);
      return;
    }

    const data = await response.json();
    if (data.success && data.data && data.data.csrf_token) {
      csrfToken = data.data.csrf_token;
      logger.debug('CSRF token fetched successfully from /csrf-token endpoint');
    } else {
      logger.error('Invalid CSRF token response format:', {
        success: data.success,
        hasData: !!data.data,
        hasCsrfToken: !!(data.data && data.data.csrf_token),
        response: data,
      });
    }
  } catch (error) {
    logger.error('Error fetching CSRF token:', error);
  }
};

/**
 * Получить текущий CSRF token
 */
export const getCSRFToken = () => csrfToken;

/**
 * Установить CSRF token вручную (из ответа логина)
 */
export const setCSRFToken = (token) => {
  if (token) {
    csrfToken = token;
    logger.debug('CSRF token set manually');
  }
};

/**
 * Очистить CSRF token (при logout)
 */
export const clearCSRFToken = () => {
  csrfToken = null;
};

/**
 * Инициализировать CSRF защиту
 * Вызывается после логина
 */
export const setupCSRF = async () => {
  await fetchCSRFToken();
};

/**
 * Попытаться обновить сессию через /auth/me
 * При успехе backend автоматически продлит сессию и обновит cookie
 * @returns {Promise<boolean>} true если сессия успешно обновлена
 */
const tryRefreshSession = async () => {
  if (isRefreshingSession) {
    return refreshPromise;
  }

  isRefreshingSession = true;
  refreshPromise = (async () => {
    try {
      const url = `${API_BASE_URL}/auth/me`;
      logger.debug('[Session Refresh] Attempting to refresh session via /auth/me');

      const response = await fetch(url, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        logger.info('[Session Refresh] Session refreshed successfully');
        await fetchCSRFToken();
        return true;
      }

      logger.warn('[Session Refresh] Failed to refresh session:', response.status);
      return false;
    } catch (error) {
      logger.error('[Session Refresh] Error refreshing session:', error);
      return false;
    } finally {
      isRefreshingSession = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
};

/**
 * Пользовательский класс ошибок для API ошибок
 */
export class APIError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }
}

/**
 * Базовая обёртка fetch с общей конфигурацией и детальным логированием
 * Поддерживает отмену запросов через AbortSignal для предотвращения утечек памяти
 * При CSRF ошибке автоматически обновляет токен и повторяет запрос
 */
async function request(endpoint, options = {}, isRetry = false) {
  const url = `${API_BASE_URL}${endpoint}`;
  const method = options.method || 'GET';
  const startTime = performance.now();

  // Проверяем, является ли body FormData
  const isFormData = options.body instanceof FormData;

  // Объединяем headers правильно чтобы custom headers не перезаписывали defaults
  // Для FormData НЕ устанавливаем Content-Type - браузер сам добавит с boundary
  const defaultHeaders = {};
  if (!isFormData) {
    defaultHeaders['Content-Type'] = 'application/json';
  }

  // Добавляем CSRF token для state-changing requests
  if (csrfToken && method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    defaultHeaders['X-CSRF-Token'] = csrfToken;
  }

  // Поддержка AbortSignal для отмены запросов
  // Если передан signal, используем его, иначе используем опции как есть
  const config = {
    credentials: 'include', // Важно для cookies (httpOnly session cookie)
    ...options,
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
    // signal уже может быть в options, мы его не перезаписываем
  };

  // Логирование отправляемого запроса
  // Для FormData не пытаемся парсить как JSON
  let logData;
  if (isFormData) {
    logData = '[FormData]';
  } else if (config.body) {
    try {
      logData = JSON.parse(config.body);
    } catch {
      logData = config.body;
    }
  }
  logger.debug('API Request:', {
    method,
    url,
    headers: config.headers,
    data: logData,
    timestamp: new Date().toISOString(),
  });

  try {
    const response = await fetch(url, config);
    const duration = performance.now() - startTime;

    // Разбор ответа
    let data = null;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    }

    // Извлекаем CSRF token из заголовка если есть (для login/register)
    const csrfTokenHeader = response.headers.get('X-CSRF-Token');
    if (csrfTokenHeader) {
      csrfToken = csrfTokenHeader;
      logger.debug('CSRF token extracted from response header:', csrfTokenHeader.substring(0, 10) + '...');
    } else if (method === 'POST' && (endpoint.includes('/auth/login') || endpoint.includes('/auth/register'))) {
      // Логируем если токен ожидался но не получен (для отладки)
      logger.warn('Expected X-CSRF-Token header not found in login/register response');
    }

    // Логирование успешного ответа
    logger.debug('API Response:', {
      status: response.status,
      statusText: response.statusText,
      url,
      method,
      duration: `${duration.toFixed(2)}ms`,
      data,
      timestamp: new Date().toISOString(),
    });

    // Обработка 401 Unauthorized
    if (response.status === 401) {
      const isLogoutEndpoint = endpoint.includes('/auth/logout');
      const isAuthMeEndpoint = endpoint.includes('/auth/me');

      // Не пытаемся refresh для logout и auth/me endpoints (избегаем циклов)
      if (!isRetry && !isLogoutEndpoint && !isAuthMeEndpoint) {
        logger.warn('API Error: Unauthorized (401), attempting session refresh...', { url, method });

        // Пытаемся обновить сессию
        const refreshed = await tryRefreshSession();
        if (refreshed) {
          logger.info('Session refreshed, retrying original request...');
          return request(endpoint, options, true);
        }
      }

      logger.error('API Error: Unauthorized (401), session refresh failed or not applicable', {
        url,
        method,
        data,
      });
      // Очищаем CSRF token при 401
      clearCSRFToken();

      // Dispatch custom event instead of direct window.location redirect
      // This allows AuthContext to handle logout and navigation properly
      if (typeof window !== 'undefined' && !isLogoutEndpoint) {
        window.dispatchEvent(new CustomEvent('auth:unauthorized', {
          detail: { url, method, status: 401 }
        }));
      }

      throw new APIError('Unauthorized', 401, null);
    }

    // Обработка 403 Forbidden (может быть CSRF ошибка)
    if (response.status === 403) {
      // Проверяем если это CSRF ошибка
      // Backend возвращает error.code = 'INVALID_CSRF'
      if (data?.error?.code === 'INVALID_CSRF' && !isRetry) {
        logger.warn('CSRF token invalid, refreshing and retrying request...');
        // Пытаемся обновить CSRF token
        await fetchCSRFToken();
        // Повторяем запрос с новым токеном (один раз)
        return request(endpoint, options, true);
      }
    }

    // Обработка не-OK ответов
    if (!response.ok) {
      let message = `HTTP ${response.status}`;

      // Проверить наличие ошибки в массиве errors
      if (Array.isArray(data?.errors) && data.errors.length > 0) {
        message = data.errors[0];
      } else if (data?.error?.message) {
        message = data.error.message;
      } else if (data?.message) {
        message = data.message;
      } else if (data?.error) {
        message = typeof data.error === 'string' ? data.error : JSON.stringify(data.error);
      }

      // Логирование ошибок ответа
      logger.error('API Error:', {
        url,
        method,
        status: response.status,
        statusText: response.statusText,
        message,
        data,
        duration: `${duration.toFixed(2)}ms`,
        timestamp: new Date().toISOString(),
      });

      throw new APIError(message, response.status, data);
    }

    // Возврат data.data если успешный ответ, иначе возврат сырых данных
    if (data && data.success && data.data !== undefined) {
      return data.data;
    }

    return data;
  } catch (error) {
    const duration = performance.now() - startTime;

    if (error instanceof APIError) {
      // APIError уже залогирована выше, просто выбрасываем
      throw error;
    }

    // Обработка AbortError - не показываем пользователю, это нормальное поведение
    if (error.name === 'AbortError') {
      logger.debug('API Request cancelled:', {
        url,
        method,
        duration: `${duration.toFixed(2)}ms`,
        timestamp: new Date().toISOString(),
      });
      // Выбрасываем специальный тип ошибки который можно отфильтровать
      const abortError = new Error('Request cancelled');
      abortError.name = 'AbortError';
      throw abortError;
    }

    // Определение типа ошибки
    let errorType = 'Unknown Error';
    if (error instanceof TypeError) {
      errorType = 'Network Error';
    } else if (error instanceof SyntaxError) {
      errorType = 'Parse Error';
    }

    // Сетевые или другие ошибки
    logger.error('API Error:', {
      url,
      method,
      errorType,
      errorMessage: error.message,
      duration: `${duration.toFixed(2)}ms`,
      timestamp: new Date().toISOString(),
      stack: error.stack,
    });

    throw new APIError(
      error.message || 'Network error occurred',
      0,
      null
    );
  }
}

/**
 * HTTP методы с логированием
 */
export const apiClient = {
  get: (endpoint, options = {}) => {
    logger.debug('GET Request:', { endpoint, options });
    return request(endpoint, { ...options, method: 'GET' });
  },

  post: (endpoint, body, options = {}) => {
    logger.debug('POST Request:', { endpoint, body, options });

    // Поддержка FormData для загрузки файлов
    if (body instanceof FormData) {
      // Для FormData НЕ устанавливаем Content-Type - браузер сам добавит с boundary
      const { headers, ...restOptions } = options;
      const filteredHeaders = { ...headers };
      delete filteredHeaders['Content-Type']; // Удаляем Content-Type для FormData

      return request(endpoint, {
        ...restOptions,
        method: 'POST',
        body: body, // FormData напрямую
        headers: filteredHeaders,
      });
    }

    // Для обычных объектов используем JSON
    return request(endpoint, {
      ...options,
      method: 'POST',
      body: JSON.stringify(body),
    });
  },

  put: (endpoint, body, options = {}) => {
    logger.debug('PUT Request:', { endpoint, body, options });

    // Поддержка FormData для загрузки файлов
    if (body instanceof FormData) {
      const { headers, ...restOptions } = options;
      const filteredHeaders = { ...headers };
      delete filteredHeaders['Content-Type'];

      return request(endpoint, {
        ...restOptions,
        method: 'PUT',
        body: body,
        headers: filteredHeaders,
      });
    }

    return request(endpoint, {
      ...options,
      method: 'PUT',
      body: JSON.stringify(body),
    });
  },

  patch: (endpoint, body, options = {}) => {
    logger.debug('PATCH Request:', { endpoint, body, options });

    // Поддержка FormData для загрузки файлов
    if (body instanceof FormData) {
      const { headers, ...restOptions } = options;
      const filteredHeaders = { ...headers };
      delete filteredHeaders['Content-Type'];

      return request(endpoint, {
        ...restOptions,
        method: 'PATCH',
        body: body,
        headers: filteredHeaders,
      });
    }

    return request(endpoint, {
      ...options,
      method: 'PATCH',
      body: JSON.stringify(body),
    });
  },

  delete: (endpoint, options = {}) => {
    logger.debug('DELETE Request:', { endpoint, options });
    return request(endpoint, { ...options, method: 'DELETE' });
  },
};

export default apiClient;
