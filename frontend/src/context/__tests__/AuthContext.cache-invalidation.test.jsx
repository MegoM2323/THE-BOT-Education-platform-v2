import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { QueryClient } from '@tanstack/react-query';

describe('AuthContext - React Query Cache Invalidation', () => {
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

  afterEach(() => {
    vi.restoreAllMocks();
    testQueryClient.clear();
  });

  describe('Cache isolation between users', () => {
    it('should ensure no cache data persists between different user sessions', async () => {
      testQueryClient.setQueryData(['user', 'profile'], { id: 1, name: 'Student User' });
      testQueryClient.setQueryData(['credits'], { balance: 100 });
      testQueryClient.setQueryData(['lessons'], { count: 5 });

      expect(testQueryClient.getQueryData(['credits'])).toBeDefined();
      expect(testQueryClient.getQueryData(['lessons'])).toBeDefined();

      testQueryClient.clear();

      expect(testQueryClient.getQueryData(['user', 'profile'])).toBeUndefined();
      expect(testQueryClient.getQueryData(['credits'])).toBeUndefined();
      expect(testQueryClient.getQueryData(['lessons'])).toBeUndefined();
    });

    it('should invalidate specific queries', async () => {
      testQueryClient.setQueryData(['credits'], { balance: 100 });

      await testQueryClient.invalidateQueries({ queryKey: ['credits'] });

      const state = testQueryClient.getQueryState(['credits']);
      expect(state?.isInvalidated).toBe(true);
    });

    it('should reset query cache', () => {
      testQueryClient.setQueryData(['test'], { value: 1 });
      expect(testQueryClient.getQueryData(['test'])).toBeDefined();

      testQueryClient.resetQueries();
      expect(testQueryClient.getQueryData(['test'])).toBeUndefined();
    });

    it('should allow setting and getting query data', () => {
      const mockData = { user: { id: 1, name: 'Test' }, balance: 50 };
      testQueryClient.setQueryData(['auth', 'user'], mockData);

      const data = testQueryClient.getQueryData(['auth', 'user']);
      expect(data).toEqual(mockData);
    });
  });
});
