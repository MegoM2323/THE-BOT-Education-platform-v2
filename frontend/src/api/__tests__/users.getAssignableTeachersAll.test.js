import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getAssignableTeachersAll } from '../users.js';
import * as apiClientModule from '../client.js';

vi.mock('../client.js', () => ({
  default: {
    get: vi.fn(),
  },
}));

describe('getAssignableTeachersAll() - Assignable Teachers (teachers, admins, methodologists)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('T001: Deduplication Tests', () => {
    // TEST 1 (FAILING): getAssignableTeachersAll should deduplicate users
    // Bug: Function merges teachers + admins + methodologists but doesn't deduplicate
    // If same user appears in multiple arrays, they'll be duplicated
    //
    // Current behavior: Returns duplicated users
    // Expected behavior: Each unique user ID should appear only once
    it('test_getAssignableTeachersAll_should_deduplicate_users', async () => {
      // Create overlapping users: some users have multiple roles
      // In reality, a user can only have one role, but could be fetched multiple times
      // if backend has inconsistencies, or we're testing cross-role assignments

      // Shared user appears in teacher and admin lists
      const sharedUser1 = {
        id: 'user-1',
        full_name: 'Alice Johnson',
        email: 'alice@example.com',
        role: 'teacher',
      };

      const sharedUser2 = {
        id: 'user-2',
        full_name: 'Bob Smith',
        email: 'bob@example.com',
        role: 'admin',
      };

      const teacherOnly = {
        id: 'user-3',
        full_name: 'Carol White',
        email: 'carol@example.com',
        role: 'teacher',
      };

      const adminOnly = {
        id: 'user-4',
        full_name: 'David Brown',
        email: 'david@example.com',
        role: 'admin',
      };

      const methodologistOnly = {
        id: 'user-5',
        full_name: 'Eve Green',
        email: 'eve@example.com',
        role: 'methodologist',
      };

      // Mock API responses for three separate calls (teachers, admins, methodologists)
      // Teachers list (includes sharedUser1 and teacherOnly)
      const teachersResponse = {
        users: [sharedUser1, teacherOnly],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 2 },
      };

      // Admins list (includes sharedUser1, sharedUser2, and adminOnly)
      // Note: sharedUser1 appears again here - this creates the duplicate!
      const adminsResponse = {
        users: [sharedUser1, sharedUser2, adminOnly],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 3 },
      };

      // Methodologists list (includes sharedUser2 and methodologistOnly)
      // Note: sharedUser2 appears again here - another duplicate!
      const methodologistsResponse = {
        users: [sharedUser2, methodologistOnly],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 2 },
      };

      // Setup mocks for Promise.all([teachers, admins, methodologists])
      // First call gets teachers
      // Second call gets admins
      // Third call gets methodologists
      let callCount = 0;
      apiClientModule.default.get.mockImplementation((url, options) => {
        callCount++;
        if (callCount === 1) {
          return Promise.resolve(teachersResponse);
        } else if (callCount === 2) {
          return Promise.resolve(adminsResponse);
        } else if (callCount === 3) {
          return Promise.resolve(methodologistsResponse);
        }
        return Promise.reject(new Error('Unexpected call'));
      });

      // Execute
      const result = await getAssignableTeachersAll();

      // Assert: verify NO duplicates exist
      // With current bug: result will have 7 items (2 + 3 + 2) - FAILS
      // After fix: result should have 5 unique items (user-1, user-2, user-3, user-4, user-5)

      expect(result).toHaveLength(5, 'Should have exactly 5 unique users after deduplication');

      // Verify each unique ID appears only once
      const userIds = result.map(u => u.id);
      const uniqueIds = new Set(userIds);
      expect(uniqueIds.size).toBe(5, 'All user IDs should be unique');

      // Verify specific users are present
      expect(userIds).toContain('user-1', 'Should include Alice');
      expect(userIds).toContain('user-2', 'Should include Bob');
      expect(userIds).toContain('user-3', 'Should include Carol');
      expect(userIds).toContain('user-4', 'Should include David');
      expect(userIds).toContain('user-5', 'Should include Eve');

      // Count occurrences - each should be exactly 1
      const user1Count = result.filter(u => u.id === 'user-1').length;
      const user2Count = result.filter(u => u.id === 'user-2').length;
      expect(user1Count).toBe(1, 'User-1 should appear exactly once');
      expect(user2Count).toBe(1, 'User-2 should appear exactly once');
    });

    it('test_getAssignableTeachersAll_no_duplicates_identical_users', async () => {
      // Test case: exact same user object in multiple lists
      const user = {
        id: 'shared-user',
        full_name: 'Shared Teacher',
        email: 'shared@example.com',
        role: 'teacher',
      };

      // Mock all three calls returning the same user
      const teachersResponse = {
        users: [user],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
      };

      const adminsResponse = {
        users: [user],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
      };

      const methodologistsResponse = {
        users: [user],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
      };

      let callCount = 0;
      apiClientModule.default.get.mockImplementation(() => {
        callCount++;
        if (callCount === 1) return Promise.resolve(teachersResponse);
        if (callCount === 2) return Promise.resolve(adminsResponse);
        if (callCount === 3) return Promise.resolve(methodologistsResponse);
        return Promise.reject(new Error('Unexpected call'));
      });

      const result = await getAssignableTeachersAll();

      // Should have exactly 1 user, not 3 duplicates
      expect(result).toHaveLength(1, 'Should deduplicate identical users from all three lists');
      expect(result[0].id).toBe('shared-user');
    });

    it('test_getAssignableTeachersAll_preserves_unique_users', async () => {
      // Test: all users are unique across lists
      const teachers = [
        { id: 't1', full_name: 'Teacher One', role: 'teacher' },
        { id: 't2', full_name: 'Teacher Two', role: 'teacher' },
      ];

      const admins = [
        { id: 'a1', full_name: 'Admin One', role: 'admin' },
        { id: 'a2', full_name: 'Admin Two', role: 'admin' },
      ];

      const methodologists = [
        { id: 'm1', full_name: 'Methodologist One', role: 'methodologist' },
      ];

      let callCount = 0;
      apiClientModule.default.get.mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.resolve({
            users: teachers,
            meta: { total_pages: 1, page: 1, per_page: 50, total: 2 },
          });
        }
        if (callCount === 2) {
          return Promise.resolve({
            users: admins,
            meta: { total_pages: 1, page: 1, per_page: 50, total: 2 },
          });
        }
        if (callCount === 3) {
          return Promise.resolve({
            users: methodologists,
            meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
          });
        }
        return Promise.reject(new Error('Unexpected call'));
      });

      const result = await getAssignableTeachersAll();

      // Should have 5 unique users
      expect(result).toHaveLength(5, 'Should have all 5 unique users');
      expect(result.map(u => u.id).sort()).toEqual(['a1', 'a2', 'm1', 't1', 't2'].sort());
    });

    it('test_getAssignableTeachersAll_empty_lists_with_duplicates', async () => {
      // Edge case: some lists empty, but results still have duplicates
      const user = {
        id: 'only-user',
        full_name: 'Only User',
        role: 'admin',
      };

      let callCount = 0;
      apiClientModule.default.get.mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.resolve({
            users: [user], // Teachers has the user
            meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
          });
        }
        if (callCount === 2) {
          return Promise.resolve({
            users: [user], // Admins also has the same user
            meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
          });
        }
        if (callCount === 3) {
          return Promise.resolve({
            users: [], // Methodologists empty
            meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
          });
        }
        return Promise.reject(new Error('Unexpected call'));
      });

      const result = await getAssignableTeachersAll();

      // Should have 1 user (deduplicated), not 2
      expect(result).toHaveLength(1, 'Should deduplicate and return only 1 user');
      expect(result[0].id).toBe('only-user');
    });
  });

  describe('T002: Sorting and Display', () => {
    it('test_getAssignableTeachersAll_sorted_by_full_name_russian', async () => {
      const users = [
        { id: '3', full_name: 'Вячеслав Петров', role: 'teacher' },
        { id: '1', full_name: 'Алексей Иванов', role: 'teacher' },
        { id: '2', full_name: 'Борис Сидоров', role: 'teacher' },
      ];

      let callCount = 0;
      apiClientModule.default.get.mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.resolve({
            users: users,
            meta: { total_pages: 1, page: 1, per_page: 50, total: 3 },
          });
        }
        if (callCount === 2) {
          return Promise.resolve({
            users: [],
            meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
          });
        }
        if (callCount === 3) {
          return Promise.resolve({
            users: [],
            meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
          });
        }
        return Promise.reject(new Error('Unexpected call'));
      });

      const result = await getAssignableTeachersAll();

      // Should be sorted by full_name in Russian locale
      expect(result[0].full_name).toBe('Алексей Иванов', 'Should start with Алексей');
      expect(result[1].full_name).toBe('Борис Сидоров', 'Should have Борис second');
      expect(result[2].full_name).toBe('Вячеслав Петров', 'Should end with Вячеслав');
    });
  });

  describe('T003: API Interaction', () => {
    it('test_getAssignableTeachersAll_makes_three_parallel_calls', async () => {
      let callCount = 0;
      apiClientModule.default.get.mockImplementation(() => {
        callCount++;
        return Promise.resolve({
          users: [],
          meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
        });
      });

      await getAssignableTeachersAll();

      // Should make exactly 3 calls: one for teachers, one for admins, one for methodologists
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(3);
    });

    it('test_getAssignableTeachersAll_passes_filters_to_all_calls', async () => {
      let callCount = 0;
      apiClientModule.default.get.mockImplementation((url) => {
        callCount++;
        // Check that filters are being passed in URL
        if (url.includes('search=john')) {
          // Filter passed correctly
        }
        return Promise.resolve({
          users: [],
          meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
        });
      });

      await getAssignableTeachersAll({ search: 'john' });

      // Should pass search filter to all 3 role queries
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(3);
    });

    it('test_getAssignableTeachersAll_propagates_errors', async () => {
      const testError = new Error('Network error');
      apiClientModule.default.get.mockRejectedValueOnce(testError);

      await expect(getAssignableTeachersAll()).rejects.toThrow('Network error');
    });
  });
});
