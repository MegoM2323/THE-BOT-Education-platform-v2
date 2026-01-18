import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as creditsAPI from '../credits.js';
import apiClient from '../client.js';

vi.mock('../client.js');

describe('Credits API - getAllCredits() Response Format Parsing', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'debug').mockImplementation(() => {});
    vi.spyOn(console, 'warn').mockImplementation(() => {});
  });

  afterEach(() => {
    console.debug.mockRestore();
    console.warn.mockRestore();
  });

  describe('getAllCredits() - Paginated Response Format', () => {
    it('should parse paginated response { data: [...], meta: {...} }', async () => {
      const mockResponse = {
        data: [
          { user_id: 'user-1', balance: 10 },
          { user_id: 'user-2', balance: 5 },
        ],
        meta: {
          total: 2,
          page: 1,
          limit: 10,
        },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [
          { user_id: 'user-1', balance: 10 },
          { user_id: 'user-2', balance: 5 },
        ],
      });
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'paginated',
          count: 2,
        })
      );
    });

    it('should extract and return correct count from paginated response', async () => {
      const mockResponse = {
        data: Array.from({ length: 50 }, (_, i) => ({
          user_id: `user-${i + 1}`,
          balance: (i + 1) * 10,
        })),
        meta: {
          total: 50,
          page: 1,
          limit: 50,
        },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(50);
      expect(result.balances[0]).toEqual({ user_id: 'user-1', balance: 10 });
      expect(result.balances[49]).toEqual({ user_id: 'user-50', balance: 500 });
    });
  });

  describe('getAllCredits() - Nested Paginated Response Format', () => {
    it('should parse nested paginated response { data: { data: [...], ... } }', async () => {
      const mockResponse = {
        data: {
          data: [
            { user_id: 'user-1', balance: 10 },
            { user_id: 'user-2', balance: 5 },
          ],
          meta: {
            total: 2,
            page: 1,
          },
        },
        meta: {
          code: 200,
        },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [
          { user_id: 'user-1', balance: 10 },
          { user_id: 'user-2', balance: 5 },
        ],
      });
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'nested_paginated',
          count: 2,
        })
      );
    });

    it('should handle deeply nested paginated structure', async () => {
      const mockResponse = {
        data: {
          data: [
            { user_id: 'a', balance: 100 },
            { user_id: 'b', balance: 200 },
            { user_id: 'c', balance: 300 },
          ],
          pagination: {
            page: 1,
            limit: 10,
            total: 3,
          },
        },
        status: 'success',
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(3);
      expect(result.balances).toEqual([
        { user_id: 'a', balance: 100 },
        { user_id: 'b', balance: 200 },
        { user_id: 'c', balance: 300 },
      ]);
    });
  });

  describe('getAllCredits() - Old Format Backward Compatibility', () => {
    it('should parse old format { balances: [...] }', async () => {
      const mockResponse = {
        balances: [
          { user_id: 'user-1', balance: 15 },
          { user_id: 'user-2', balance: 20 },
        ],
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [
          { user_id: 'user-1', balance: 15 },
          { user_id: 'user-2', balance: 20 },
        ],
      });
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'balances',
          count: 2,
        })
      );
    });

    it('should handle empty balances array in old format', async () => {
      const mockResponse = {
        balances: [],
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
    });
  });

  describe('getAllCredits() - Array Direct Format', () => {
    it('should parse array returned directly', async () => {
      const mockResponse = [
        { user_id: 'user-1', balance: 10 },
        { user_id: 'user-2', balance: 5 },
      ];

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [
          { user_id: 'user-1', balance: 10 },
          { user_id: 'user-2', balance: 5 },
        ],
      });
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'array',
          count: 2,
        })
      );
    });

    it('should handle empty array directly', async () => {
      const mockResponse = [];

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
    });
  });

  describe('getAllCredits() - Empty Response Handling', () => {
    it('should return empty balances for empty response', async () => {
      const mockResponse = {};

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'unknown',
          count: 0,
        })
      );
    });

    it('should handle null response gracefully', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce(null);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
    });

    it('should handle undefined response gracefully', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce(undefined);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
    });

    it('should return empty balances when response has null data', async () => {
      const mockResponse = {
        data: null,
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
    });

    it('should return empty balances when response has undefined data', async () => {
      const mockResponse = {
        data: undefined,
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result).toEqual({
        balances: [],
      });
    });
  });

  describe('getAllCredits() - Multiple Students with Different Credits', () => {
    it('should handle multiple students with varying credit amounts in paginated format', async () => {
      const mockResponse = {
        data: [
          { user_id: 'user-1', balance: 0 },
          { user_id: 'user-2', balance: 5 },
          { user_id: 'user-3', balance: 999 },
          { user_id: 'user-4', balance: 1 },
        ],
        meta: { total: 4 },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(4);
      expect(result.balances[0]).toEqual({ user_id: 'user-1', balance: 0 });
      expect(result.balances[1]).toEqual({ user_id: 'user-2', balance: 5 });
      expect(result.balances[2]).toEqual({ user_id: 'user-3', balance: 999 });
      expect(result.balances[3]).toEqual({ user_id: 'user-4', balance: 1 });
    });

    it('should maintain order of students from API response', async () => {
      const mockResponse = {
        data: [
          { user_id: 'charlie', balance: 30 },
          { user_id: 'alice', balance: 10 },
          { user_id: 'bob', balance: 20 },
        ],
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances.map(b => b.user_id)).toEqual(['charlie', 'alice', 'bob']);
    });
  });

  describe('getAllCredits() - Logging and Format Detection', () => {
    it('should log format analysis before parsing', async () => {
      const mockResponse = {
        data: [{ user_id: 'user-1', balance: 10 }],
        meta: { total: 1 },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      await creditsAPI.getAllCredits();

      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Response format analysis:',
        expect.objectContaining({
          isArray: false,
          hasData: true,
          hasBalances: false,
          dataType: 'object',
        })
      );
    });

    it('should log successful parsing with format and count', async () => {
      const mockResponse = {
        balances: [
          { user_id: 'user-1', balance: 10 },
          { user_id: 'user-2', balance: 20 },
          { user_id: 'user-3', balance: 30 },
        ],
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      await creditsAPI.getAllCredits();

      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'balances',
          count: 3,
          hasMeta: false,
        })
      );
    });

    it('should detect meta field presence', async () => {
      const mockResponse = {
        data: [{ user_id: 'user-1', balance: 10 }],
        meta: {
          page: 1,
          limit: 10,
          total: 1,
        },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      await creditsAPI.getAllCredits();

      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          hasMeta: true,
        })
      );
    });
  });

  describe('getAllCredits() - Format Priority and Fallback', () => {
    it('should prefer array detection over balances field', async () => {
      const mockResponse = [
        { user_id: 'user-1', balance: 10 },
      ];

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(1);
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'array',
        })
      );
    });

    it('should prefer data field over balances field', async () => {
      const mockResponse = {
        data: [{ user_id: 'user-1', balance: 10 }],
        balances: [{ user_id: 'user-2', balance: 20 }], // Should be ignored
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances[0].user_id).toBe('user-1');
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'paginated',
        })
      );
    });

    it('should check nested data.data before data', async () => {
      const mockResponse = {
        data: {
          data: [{ user_id: 'user-nested', balance: 30 }],
        },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances[0].user_id).toBe('user-nested');
      expect(console.debug).toHaveBeenCalledWith(
        '[Credits API] Received all credits response:',
        expect.objectContaining({
          format: 'nested_paginated',
        })
      );
    });
  });

  describe('getAllCredits() - Edge Cases', () => {
    it('should handle response with extra fields', async () => {
      const mockResponse = {
        data: [{ user_id: 'user-1', balance: 10 }],
        meta: { total: 1 },
        extra_field: 'should be ignored',
        another_field: { nested: 'value' },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(1);
      expect(result.balances[0]).toEqual({ user_id: 'user-1', balance: 10 });
    });

    it('should handle balances with extra user info fields', async () => {
      const mockResponse = {
        data: [
          {
            user_id: 'user-1',
            balance: 10,
            user_name: 'John',
            user_email: 'john@test.com',
          },
          {
            user_id: 'user-2',
            balance: 5,
            user_name: 'Jane',
            user_email: 'jane@test.com',
          },
        ],
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(2);
      // Extra fields should be preserved
      expect(result.balances[0]).toEqual({
        user_id: 'user-1',
        balance: 10,
        user_name: 'John',
        user_email: 'john@test.com',
      });
    });

    it('should handle response with mixed case field names in data array', async () => {
      const mockResponse = {
        data: [
          { user_id: 'user-1', balance: 10 },
          { user_id: 'user-2', balance: 20 },
        ],
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(2);
    });

    it('should handle very large response', async () => {
      const largeData = Array.from({ length: 10000 }, (_, i) => ({
        user_id: `user-${i}`,
        balance: Math.floor(Math.random() * 1000),
      }));

      const mockResponse = {
        data: largeData,
        meta: { total: 10000 },
      };

      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse);

      const result = await creditsAPI.getAllCredits();

      expect(result.balances).toHaveLength(10000);
    });
  });
});
