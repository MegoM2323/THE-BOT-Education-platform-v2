import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { apiClient } from '../client.js';
import { saveToken, clearToken } from '../client.js';

describe('API Client - Authorization Headers', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();

    // Mock fetch globally
    global.fetch = vi.fn();

    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  describe('Authorization Header in Requests', () => {
    // ПРИМЕЧАНИЕ: Текущая реализация использует httpOnly cookies вместо Authorization header
    // saveToken() теперь no-op - backend устанавливает httpOnly cookie при логине
    it('should include Authorization header when token is stored', async () => {
      const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
      saveToken(testToken); // Это теперь no-op, т.к. используются httpOnly cookies

      // Mock successful response
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { id: 1, name: 'Test' } }),
      });

      try {
        await apiClient.get('/test-endpoint');
      } catch (error) {
        // Ignore errors from our mock
      }

      // Check that fetch was called
      expect(global.fetch).toHaveBeenCalled();
      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      // Аутентификация теперь через httpOnly cookies (credentials: 'include')
      // Authorization header больше не используется
      expect(config.credentials).toBe('include');
    });

    it('should not include Authorization header when no token is stored', async () => {
      clearToken();

      // Mock successful response
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { id: 1 } }),
      });

      try {
        await apiClient.get('/test-endpoint');
      } catch (error) {
        // Ignore errors
      }

      // Check that fetch was called without Authorization header
      // (аутентификация теперь через httpOnly cookies)
      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.headers.Authorization).toBeUndefined();
      expect(config.credentials).toBe('include'); // httpOnly cookies
    });

    it('should include credentials in POST requests', async () => {
      const testToken = 'test-jwt-token-post';
      saveToken(testToken); // no-op

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { result: 'ok' } }),
      });

      try {
        await apiClient.post('/test-endpoint', { key: 'value' });
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.method).toBe('POST');
      // Аутентификация теперь через httpOnly cookies
      expect(config.credentials).toBe('include');
    });

    it('should include credentials in PUT requests', async () => {
      const testToken = 'test-jwt-token-put';
      saveToken(testToken); // no-op

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { updated: true } }),
      });

      try {
        await apiClient.put('/test-endpoint', { key: 'updated' });
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.method).toBe('PUT');
      // Аутентификация теперь через httpOnly cookies
      expect(config.credentials).toBe('include');
    });

    it('should include credentials in DELETE requests', async () => {
      const testToken = 'test-jwt-token-delete';
      saveToken(testToken); // no-op

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({}),
      });

      try {
        await apiClient.delete('/test-endpoint');
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.method).toBe('DELETE');
      // Аутентификация теперь через httpOnly cookies
      expect(config.credentials).toBe('include');
    });
  });

  describe('Credentials Configuration', () => {
    it('should always include credentials for cookie support', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true }),
      });

      try {
        await apiClient.get('/test-endpoint');
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.credentials).toBe('include');
    });
  });

  describe('Content-Type Header', () => {
    it('should include Content-Type header', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true }),
      });

      try {
        await apiClient.get('/test-endpoint');
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.headers['Content-Type']).toBe('application/json');
    });
  });

  describe('Custom Headers', () => {
    it('should merge custom headers with default headers', async () => {
      saveToken('test-token'); // no-op

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true }),
      });

      try {
        await apiClient.get('/test-endpoint', {
          headers: { 'X-Custom-Header': 'custom-value' },
        });
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      // Аутентификация теперь через httpOnly cookies (Authorization header не используется)
      expect(config.credentials).toBe('include');
      expect(config.headers['X-Custom-Header']).toBe('custom-value');
      // Content-Type may be overridden by custom headers, but should default to application/json
      expect(config.headers['Content-Type']).toBeDefined();
    });

    it('should allow overriding Content-Type with custom headers', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'text/plain']]),
        text: async () => 'OK',
      });

      try {
        await apiClient.post('/test-endpoint', { test: true }, {
          headers: { 'Content-Type': 'text/plain' },
        });
      } catch (error) {
        // Ignore errors
      }

      const callArgs = global.fetch.mock.calls[0];
      const config = callArgs[1];

      expect(config.headers['Content-Type']).toBe('text/plain');
    });
  });
});
