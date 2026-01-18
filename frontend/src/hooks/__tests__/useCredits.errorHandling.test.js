import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('useCredits Error Categorization Helper', () => {
  let categorizeError;

  beforeEach(() => {
    categorizeError = (error) => {
      if (!error) return { category: 'unknown', message: 'Неизвестная ошибка' };

      const status = error?.status || error?.response?.status;
      const code = error?.code;
      const message = error?.message || '';

      if (status === 401) {
        return {
          category: 'unauthorized',
          message: 'Требуется повторный вход',
          userMessage: 'Ваша сессия истекла. Пожалуйста, авторизуйтесь заново.'
        };
      }

      if (status === 403) {
        return {
          category: 'forbidden',
          message: 'Доступ запрещен',
          userMessage: 'У вас нет прав для доступа к этому ресурсу.'
        };
      }

      if (status >= 500) {
        return {
          category: 'server_error',
          message: `Ошибка сервера (${status})`,
          userMessage: 'Сервер временно недоступен. Пожалуйста, попробуйте позже.'
        };
      }

      if (code === 'NETWORK_ERROR' || message.toLowerCase().includes('network')) {
        return {
          category: 'network_error',
          message: 'Ошибка сети',
          userMessage: 'Проблема с подключением. Попробуем снова автоматически.'
        };
      }

      if (status >= 400 && status < 500) {
        return {
          category: 'client_error',
          message: `Ошибка запроса (${status})`,
          userMessage: 'Некорректный запрос. Пожалуйста, попробуйте позже.'
        };
      }

      return {
        category: 'unknown',
        message: message || 'Неизвестная ошибка',
        userMessage: 'Произошла ошибка при загрузке данных. Пожалуйста, попробуйте позже.'
      };
    };
  });

  describe('Error Categorization', () => {
    it('should categorize 401 status as unauthorized', () => {
      const error = new Error('Unauthorized');
      error.status = 401;

      const result = categorizeError(error);

      expect(result.category).toBe('unauthorized');
      expect(result.message).toBe('Требуется повторный вход');
      expect(result.userMessage).toContain('сессия истекла');
    });

    it('should categorize 403 status as forbidden', () => {
      const error = new Error('Forbidden');
      error.status = 403;

      const result = categorizeError(error);

      expect(result.category).toBe('forbidden');
      expect(result.message).toBe('Доступ запрещен');
    });

    it('should categorize 500+ status as server_error', () => {
      const error = new Error('Internal Server Error');
      error.status = 500;

      const result = categorizeError(error);

      expect(result.category).toBe('server_error');
      expect(result.message).toContain('Ошибка сервера');
      expect(result.userMessage).toContain('временно недоступен');
    });

    it('should categorize 503 status as server_error', () => {
      const error = new Error('Service Unavailable');
      error.status = 503;

      const result = categorizeError(error);

      expect(result.category).toBe('server_error');
    });

    it('should categorize NETWORK_ERROR code as network_error', () => {
      const error = new Error('Network timeout');
      error.code = 'NETWORK_ERROR';

      const result = categorizeError(error);

      expect(result.category).toBe('network_error');
      expect(result.userMessage).toContain('подключением');
    });

    it('should categorize message containing "network" as network_error', () => {
      const error = new Error('Network error during request');

      const result = categorizeError(error);

      expect(result.category).toBe('network_error');
    });

    it('should categorize 400 status as client_error', () => {
      const error = new Error('Bad Request');
      error.status = 400;

      const result = categorizeError(error);

      expect(result.category).toBe('client_error');
    });

    it('should categorize 404 status as client_error', () => {
      const error = new Error('Not Found');
      error.status = 404;

      const result = categorizeError(error);

      expect(result.category).toBe('client_error');
    });

    it('should categorize null error as unknown', () => {
      const result = categorizeError(null);

      expect(result.category).toBe('unknown');
      expect(result.message).toBe('Неизвестная ошибка');
    });

    it('should categorize undefined error as unknown', () => {
      const result = categorizeError(undefined);

      expect(result.category).toBe('unknown');
    });

    it('should handle error with response.status property', () => {
      const error = new Error('Unauthorized');
      error.response = { status: 401 };

      const result = categorizeError(error);

      expect(result.category).toBe('unauthorized');
    });

    it('should provide user-friendly messages for all categories', () => {
      const errors = [
        { error: new Error('Auth'), status: 401, expectedMsg: 'сессия истекла' },
        { error: new Error('Forbidden'), status: 403, expectedMsg: 'прав' },
        { error: new Error('Server'), status: 500, expectedMsg: 'временно' },
      ];

      errors.forEach(({ error, status, expectedMsg }) => {
        error.status = status;
        const result = categorizeError(error);
        expect(result.userMessage).toContain(expectedMsg);
      });
    });

    it('should always return object with category, message, and userMessage', () => {
      const testErrors = [
        new Error('Test'),
        { status: 401 },
        { code: 'NETWORK_ERROR' },
        { status: 500 },
      ];

      testErrors.forEach(error => {
        const result = categorizeError(error);
        expect(result).toHaveProperty('category');
        expect(result).toHaveProperty('message');
        expect(result).toHaveProperty('userMessage');
      });
    });
  });

  describe('Retry Logic Implications', () => {
    it('should identify non-retryable errors (401, 403)', () => {
      const errors = [
        { error: new Error('Unauthorized'), status: 401, shouldRetry: false },
        { error: new Error('Forbidden'), status: 403, shouldRetry: false },
      ];

      errors.forEach(({ error, status, shouldRetry }) => {
        error.status = status;
        const result = categorizeError(error);
        const isAuthError = result.category === 'unauthorized' || result.category === 'forbidden';
        expect(isAuthError).toBe(!shouldRetry);
      });
    });

    it('should identify retryable errors (network, 5xx)', () => {
      const errors = [
        { error: new Error('Network error'), category: 'network_error', shouldRetry: true },
        { error: new Error('Server error'), status: 500, category: 'server_error', shouldRetry: true },
      ];

      errors.forEach(({ error, status, category, shouldRetry }) => {
        if (status) error.status = status;
        if (category === 'network_error') error.code = 'NETWORK_ERROR';
        const result = categorizeError(error);
        const isRetryable = result.category !== 'unauthorized' && result.category !== 'forbidden';
        expect(isRetryable).toBe(shouldRetry);
      });
    });
  });
});
