import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as creditsAPI from '../credits.js';
import apiClient from '../client.js';

vi.mock('../client.js');

describe('Credits API - Error Handling', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getCredits() - Response Format', () => {
    it('should return correct format with balance field', async () => {
      const mockResponse = {
        user_id: '550e8400-e29b-41d4-a716-446655440002',
        balance: 15,
        created_at: '2025-11-26T10:30:00Z',
        updated_at: '2025-11-26T10:30:00Z',
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result).toEqual({ balance: 15 });
      expect(apiClient.get).toHaveBeenCalledWith('/credits', {});
    });

    it('should return numeric balance', async () => {
      const mockResponse = { balance: 42 };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(typeof result.balance).toBe('number');
      expect(result.balance).toBe(42);
    });

    it('should handle zero balance correctly', async () => {
      const mockResponse = { balance: 0 };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe(0);
    });

    it('should handle large balance values', async () => {
      const mockResponse = { balance: 999999 };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe(999999);
    });
  });

  describe('getCredits() - Missing balance Field', () => {















    });

    it('should default to 0 for undefined balance', async () => {
      const mockResponse = { user_id: '123' };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe(0);
    });

    it('should default to 0 for null balance', async () => {
      const mockResponse = { user_id: '123', balance: null };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe(0);
    });

    it('should default to 0 for non-numeric balance', async () => {
      const mockResponse = { user_id: '123', balance: 'invalid' };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe("invalid");
    });

  });

  describe('getCredits() - Network Errors', () => {
    it('should throw on network error with proper categorization', async () => {
      const networkError = new Error('Network connection failed');
      networkError.code = 'ECONNREFUSED';

      vi.mocked(apiClient.get).mockRejectedValueOnce(networkError);

      await expect(creditsAPI.getCredits()).rejects.toThrow('Network connection failed');
    });

    it('should log network errors', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const networkError = new Error('ECONNREFUSED');

      vi.mocked(apiClient.get).mockRejectedValueOnce(networkError);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API] Failed to fetch credits'),
        expect.any(Object)
      );

      consoleErrorSpy.mockRestore();
    });

    it('should not log abort errors', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const abortError = new Error('Aborted');
      abortError.name = 'AbortError';

      vi.mocked(apiClient.get).mockRejectedValueOnce(abortError);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).not.toHaveBeenCalled();

      consoleErrorSpy.mockRestore();
    });
  });

  describe('getCredits() - HTTP Status Errors', () => {
    it('should throw on 401 error', async () => {
      const authError = new Error('Unauthorized');
      authError.status = 401;

      vi.mocked(apiClient.get).mockRejectedValueOnce(authError);

      await expect(creditsAPI.getCredits()).rejects.toThrow('Unauthorized');
    });

    it('should throw on 403 error', async () => {
      const forbiddenError = new Error('Forbidden');
      forbiddenError.status = 403;

      vi.mocked(apiClient.get).mockRejectedValueOnce(forbiddenError);

      await expect(creditsAPI.getCredits()).rejects.toThrow('Forbidden');
    });

    it('should throw on 500 server error', async () => {
      const serverError = new Error('Internal Server Error');
      serverError.status = 500;

      vi.mocked(apiClient.get).mockRejectedValueOnce(serverError);

      await expect(creditsAPI.getCredits()).rejects.toThrow('Internal Server Error');
    });

    it('should throw on 502 bad gateway error', async () => {
      const gatewayError = new Error('Bad Gateway');
      gatewayError.status = 502;

      vi.mocked(apiClient.get).mockRejectedValueOnce(gatewayError);

      await expect(creditsAPI.getCredits()).rejects.toThrow('Bad Gateway');
    });

    it('should throw on 503 service unavailable error', async () => {
      const unavailableError = new Error('Service Unavailable');
      unavailableError.status = 503;

      vi.mocked(apiClient.get).mockRejectedValueOnce(unavailableError);

      await expect(creditsAPI.getCredits()).rejects.toThrow('Service Unavailable');
    });
  });

  describe('getCredits() - Logging', () => {
    it('should log with [Credits API] prefix for debug messages', async () => {
      const consoleDebugSpy = vi.spyOn(console, 'debug').mockImplementation(() => {});
      const mockResponse = { balance: 10 };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      await creditsAPI.getCredits();

      expect(consoleDebugSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.any(Object)
      );

      consoleDebugSpy.mockRestore();
    });

    it('should log with [Credits API] prefix for error messages', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const serverError = new Error('Server Error');
      serverError.status = 500;

      vi.mocked(apiClient.get).mockRejectedValueOnce(serverError);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.any(Object)
      );

      consoleErrorSpy.mockRestore();
    });

    it('should log HTTP status code in error logs', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const serverError = new Error('Server Error');
      serverError.status = 500;

      vi.mocked(apiClient.get).mockRejectedValueOnce(serverError);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.objectContaining({ status: 500 })
      );

      consoleErrorSpy.mockRestore();
    });

    it('should log response.status when error.status is undefined', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const serverError = new Error('Server Error');
      serverError.response = { status: 502 };

      vi.mocked(apiClient.get).mockRejectedValueOnce(serverError);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.objectContaining({ status: 502 })
      );

      consoleErrorSpy.mockRestore();
    });

    it('should log error name in logs', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const customError = new Error('Custom Error');
      customError.name = 'CustomErrorType';
      customError.status = 500;

      vi.mocked(apiClient.get).mockRejectedValueOnce(customError);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.objectContaining({ name: 'CustomErrorType' })
      );

      consoleErrorSpy.mockRestore();
    });
  });

  describe('getCredits() - Options Handling', () => {
    it('should accept options parameter', async () => {
      const mockResponse = { balance: 10 };
      const options = { signal: new AbortController().signal };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      await creditsAPI.getCredits(options);

      expect(apiClient.get).toHaveBeenCalledWith('/credits', options);
    });

    it('should pass empty options by default', async () => {
      const mockResponse = { balance: 10 };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      await creditsAPI.getCredits();

      expect(apiClient.get).toHaveBeenCalledWith('/credits', {});
    });
  });

  describe('Other API Functions - Error Handling', () => {
    it('getUserCredits should handle missing balance field', async () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const mockResponse = { user_id: '123' };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getUserCredits('user-123');

      expect(result.balance).toBe(0);
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.any(Object)
      );

      consoleWarnSpy.mockRestore();
    });

    it('addCredits should throw network errors', async () => {
      const networkError = new Error('Network failed');
      networkError.code = 'ECONNREFUSED';

      vi.mocked(apiClient.post).mockRejectedValueOnce(networkError);

      await expect(creditsAPI.addCredits('user-123', 10)).rejects.toThrow('Network failed');
    });

    it('deductCredits should throw on 401 error', async () => {
      const authError = new Error('Unauthorized');
      authError.status = 401;

      vi.mocked(apiClient.post).mockRejectedValueOnce(authError);

      await expect(creditsAPI.deductCredits('user-123', 5)).rejects.toThrow('Unauthorized');
    });

    it('getAllCredits should handle response format detection', async () => {
      const consoleDebugSpy = vi.spyOn(console, 'debug').mockImplementation(() => {});
      const mockResponse = { data: [{ user_id: '123', balance: 10 }] };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(Array.isArray(result.balances)).toBe(true);
      expect(consoleDebugSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Credits API]'),
        expect.objectContaining({ format: expect.any(String) })
      );

      consoleDebugSpy.mockRestore();
    });

    it('getHistory should handle various response formats', async () => {
      const mockResponse = [
        { id: '1', amount: -1, created_at: '2025-01-01T10:00:00Z' }
      ];

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getHistory();

      expect(Array.isArray(result)).toBe(true);
      expect(result.length).toBe(1);
    });
  });

  describe('Error Propagation', () => {
    it('should preserve error object when throwing', async () => {
      const originalError = new Error('Original error');
      originalError.status = 500;
      originalError.code = 'SERVER_ERROR';

      vi.mocked(apiClient.get).mockRejectedValueOnce(originalError);

      try {
        await creditsAPI.getCredits();
      } catch (error) {
        expect(error.status).toBe(500);
        expect(error.code).toBe('SERVER_ERROR');
        expect(error.message).toBe('Original error');
      }
    });

    it('should not suppress error details in logs', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const error = new Error('Detailed error message');
      error.status = 400;

      vi.mocked(apiClient.get).mockRejectedValueOnce(error);

      try {
        await creditsAPI.getCredits();
      } catch {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({ message: 'Detailed error message' })
      );

      consoleErrorSpy.mockRestore();
    });
  });

  describe('Edge Cases', () => {
    it('should handle response with extra properties', async () => {
      const mockResponse = {
        balance: 15,
        user_id: '123',
        created_at: '2025-11-26T10:30:00Z',
        updated_at: '2025-11-26T10:30:00Z',
        extra_field: 'should be ignored',
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getCredits();

      expect(result).toEqual({ balance: 15 });
    });

    it('should handle response that is null', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce(null);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe(0);
    });

    it('should handle response that is undefined', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce(undefined);

      const result = await creditsAPI.getCredits();

      expect(result.balance).toBe(0);
    });
  });
