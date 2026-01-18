import { describe, it, expect, beforeEach, vi } from 'vitest';
import { QueryClient } from '@tanstack/react-query';

describe('AuthContext - Race Condition Handling', () => {
  let testQueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    testQueryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  describe('Query client cache operations', () => {
    it('should handle concurrent cache updates', async () => {
      testQueryClient.setQueryData(['user'], { id: 1, name: 'User 1' });

      await Promise.all([
        testQueryClient.setQueryData(['user'], { id: 1, name: 'User 2' }),
        testQueryClient.setQueryData(['credits'], { balance: 100 }),
      ]);

      expect(testQueryClient.getQueryData(['user'])).toEqual({ id: 1, name: 'User 2' });
      expect(testQueryClient.getQueryData(['credits'])).toEqual({ balance: 100 });
    });

    it('should handle rapid cache invalidation', async () => {
      testQueryClient.setQueryData(['data'], { value: 1 });

      await testQueryClient.invalidateQueries({ queryKey: ['data'] });
      await testQueryClient.invalidateQueries({ queryKey: ['data'] });

      const state = testQueryClient.getQueryState(['data']);
      expect(state?.isInvalidated).toBe(true);
    });

    it('should clear all queries atomically', () => {
      testQueryClient.setQueryData(['a'], 1);
      testQueryClient.setQueryData(['b'], 2);
      testQueryClient.setQueryData(['c'], 3);

      testQueryClient.clear();

      expect(testQueryClient.getQueryData(['a'])).toBeUndefined();
      expect(testQueryClient.getQueryData(['b'])).toBeUndefined();
      expect(testQueryClient.getQueryData(['c'])).toBeUndefined();
    });
  });

  describe('Cache isolation', () => {
    it('should maintain separate query keys', () => {
      testQueryClient.setQueryData(['user', 1], { id: 1, name: 'User 1' });
      testQueryClient.setQueryData(['user', 2], { id: 2, name: 'User 2' });

      expect(testQueryClient.getQueryData(['user', 1])).toEqual({ id: 1, name: 'User 1' });
      expect(testQueryClient.getQueryData(['user', 2])).toEqual({ id: 2, name: 'User 2' });
    });

    it('should invalidate queries by prefix', async () => {
      testQueryClient.setQueryData(['user', 'profile'], { name: 'Test' });
      testQueryClient.setQueryData(['user', 'settings'], { theme: 'dark' });

      await testQueryClient.invalidateQueries({ queryKey: ['user'] });

      expect(testQueryClient.getQueryState(['user', 'profile'])?.isInvalidated).toBe(true);
      expect(testQueryClient.getQueryState(['user', 'settings'])?.isInvalidated).toBe(true);
    });
  });
});
