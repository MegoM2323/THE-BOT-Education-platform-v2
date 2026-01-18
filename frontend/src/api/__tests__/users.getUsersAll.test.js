import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as usersAPI from '../users.js';

vi.mock('../client.js', () => ({
  default: {
    get: vi.fn(),
  },
}));

import apiClient from '../client.js';

describe('getUsersAll()', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('T1: Zero users (empty result)', () => {
    it('should return empty array when API returns no users', async () => {
      apiClient.get.mockResolvedValueOnce({
        users: [],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
      });

      const result = await usersAPI.getUsersAll();

      expect(result).toEqual([]);
      expect(apiClient.get).toHaveBeenCalledTimes(1);
    });
  });

  describe('T2: Single page (30 users)', () => {
    it('should return 30 users from single page', async () => {
      const mockUsers = Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      apiClient.get.mockResolvedValueOnce({
        users: mockUsers,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 30 },
      });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(30);
      expect(result).toEqual(mockUsers);
      expect(apiClient.get).toHaveBeenCalledTimes(1);
    });
  });

  describe('T3: Two pages (80 users)', () => {
    it('should return 80 users from two pages', async () => {
      const page1Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      const page2Users = Array.from({ length: 30 }, (_, i) => ({
        id: i + 51,
        email: `user${i + 51}@test.com`,
        full_name: `User ${i + 51}`,
        role: 'student',
      }));

      apiClient.get
        .mockResolvedValueOnce({
          users: page1Users,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 80 },
        })
        .mockResolvedValueOnce({
          users: page2Users,
          meta: { total_pages: 2, page: 2, per_page: 50, total: 80 },
        });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(80);
      expect(result).toEqual([...page1Users, ...page2Users]);
      expect(apiClient.get).toHaveBeenCalledTimes(2);
    });
  });

  describe('T4: Three pages (125 users)', () => {
    it('should return 125 users from three pages', async () => {
      const page1Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      const page2Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 51,
        email: `user${i + 51}@test.com`,
        full_name: `User ${i + 51}`,
        role: 'student',
      }));

      const page3Users = Array.from({ length: 25 }, (_, i) => ({
        id: i + 101,
        email: `user${i + 101}@test.com`,
        full_name: `User ${i + 101}`,
        role: 'student',
      }));

      apiClient.get
        .mockResolvedValueOnce({
          users: page1Users,
          meta: { total_pages: 3, page: 1, per_page: 50, total: 125 },
        })
        .mockResolvedValueOnce({
          users: page2Users,
          meta: { total_pages: 3, page: 2, per_page: 50, total: 125 },
        })
        .mockResolvedValueOnce({
          users: page3Users,
          meta: { total_pages: 3, page: 3, per_page: 50, total: 125 },
        });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(125);
      expect(result).toEqual([...page1Users, ...page2Users, ...page3Users]);
      expect(apiClient.get).toHaveBeenCalledTimes(3);
    });
  });

  describe('T5: Filter by role=student', () => {
    it('should pass role filter to API and return student users', async () => {
      const mockStudents = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `student${i + 1}@test.com`,
        full_name: `Student ${i + 1}`,
        role: 'student',
      }));

      apiClient.get.mockResolvedValueOnce({
        users: mockStudents,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 50 },
      });

      const result = await usersAPI.getUsersAll({ role: 'student' });

      expect(result).toHaveLength(50);
      expect(result.every(u => u.role === 'student')).toBe(true);
      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('role=student'),
        expect.any(Object)
      );
    });
  });

  describe('T6: Filter by role=teacher', () => {
    it('should pass role filter to API and return teacher users', async () => {
      const mockTeachers = Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        email: `teacher${i + 1}@test.com`,
        full_name: `Teacher ${i + 1}`,
        role: 'teacher',
      }));

      apiClient.get.mockResolvedValueOnce({
        users: mockTeachers,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 30 },
      });

      const result = await usersAPI.getUsersAll({ role: 'teacher' });

      expect(result).toHaveLength(30);
      expect(result.every(u => u.role === 'teacher')).toBe(true);
      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('role=teacher'),
        expect.any(Object)
      );
    });
  });

  describe('T7: Filter by search=Alex', () => {
    it('should pass search filter to API and return matching users', async () => {
      const mockResults = Array.from({ length: 10 }, (_, i) => ({
        id: i + 1,
        email: `alex${i + 1}@test.com`,
        full_name: `Alex ${i + 1}`,
        role: 'student',
      }));

      apiClient.get.mockResolvedValueOnce({
        users: mockResults,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 10 },
      });

      const result = await usersAPI.getUsersAll({ search: 'Alex' });

      expect(result).toHaveLength(10);
      expect(result.every(u => u.full_name.includes('Alex'))).toBe(true);
      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('search=Alex'),
        expect.any(Object)
      );
    });
  });

  describe('T8: Combined filters role=student + search=John', () => {
    it('should pass combined filters to API', async () => {
      const mockResults = Array.from({ length: 5 }, (_, i) => ({
        id: i + 1,
        email: `john.student${i + 1}@test.com`,
        full_name: `John Student ${i + 1}`,
        role: 'student',
      }));

      apiClient.get.mockResolvedValueOnce({
        users: mockResults,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 5 },
      });

      const result = await usersAPI.getUsersAll({ role: 'student', search: 'John' });

      expect(result).toHaveLength(5);
      expect(result.every(u => u.role === 'student' && u.full_name.includes('John'))).toBe(true);
      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('role=student'),
        expect.any(Object)
      );
      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('search=John'),
        expect.any(Object)
      );
    });
  });

  describe('T9: API error (500)', () => {
    it('should throw error when API returns 500', async () => {
      const error = new Error('Server error');
      error.response = { status: 500 };
      apiClient.get.mockRejectedValueOnce(error);

      await expect(usersAPI.getUsersAll()).rejects.toThrow('Server error');
      expect(apiClient.get).toHaveBeenCalledTimes(1);
    });
  });

  describe('T10: AbortError not logged', () => {
    it('should not log AbortError when getUsersWithPagination throws', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error');
      const abortError = new Error('Aborted');
      abortError.name = 'AbortError';

      apiClient.get.mockRejectedValueOnce(abortError);

      await expect(usersAPI.getUsersAll()).rejects.toThrow('Aborted');
      expect(consoleErrorSpy).not.toHaveBeenCalledWith(
        expect.stringContaining('Error fetching'),
        expect.any(Error)
      );

      consoleErrorSpy.mockRestore();
    });
  });

  describe('T11: Concurrent API calls across pages', () => {
    it('should make separate API calls for each page', async () => {
      const page1Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      const page2Users = Array.from({ length: 30 }, (_, i) => ({
        id: i + 51,
        email: `user${i + 51}@test.com`,
        full_name: `User ${i + 51}`,
        role: 'student',
      }));

      apiClient.get
        .mockResolvedValueOnce({
          users: page1Users,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 80 },
        })
        .mockResolvedValueOnce({
          users: page2Users,
          meta: { total_pages: 2, page: 2, per_page: 50, total: 80 },
        });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(80);
      expect(apiClient.get).toHaveBeenCalledTimes(2);
      expect(apiClient.get).toHaveBeenNthCalledWith(
        1,
        expect.stringContaining('page=1'),
        expect.any(Object)
      );
      expect(apiClient.get).toHaveBeenNthCalledWith(
        2,
        expect.stringContaining('page=2'),
        expect.any(Object)
      );
    });
  });

  describe('T12: User data preservation', () => {
    it('should preserve all user fields when paginating', async () => {
      const page1Users = [
        {
          id: 1,
          email: 'user1@test.com',
          full_name: 'User One',
          role: 'student',
          created_at: '2024-01-01',
          updated_at: '2024-01-02',
          is_active: true,
          phone: '+1234567890',
        },
      ];

      const page2Users = [
        {
          id: 51,
          email: 'user51@test.com',
          full_name: 'User Fifty-One',
          role: 'teacher',
          created_at: '2024-01-03',
          updated_at: '2024-01-04',
          is_active: false,
          phone: '+9876543210',
        },
      ];

      apiClient.get
        .mockResolvedValueOnce({
          users: page1Users,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 100 },
        })
        .mockResolvedValueOnce({
          users: page2Users,
          meta: { total_pages: 2, page: 2, per_page: 50, total: 100 },
        });

      const result = await usersAPI.getUsersAll();

      expect(result[0]).toEqual(page1Users[0]);
      expect(result[1]).toEqual(page2Users[0]);
      expect(result[0].phone).toBe('+1234567890');
      expect(result[1].is_active).toBe(false);
    });
  });

  describe('T13: Single user edge case', () => {
    it('should handle single user correctly', async () => {
      const singleUser = {
        id: 1,
        email: 'user@test.com',
        full_name: 'Solo User',
        role: 'admin',
      };

      apiClient.get.mockResolvedValueOnce({
        users: [singleUser],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
      });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual(singleUser);
    });
  });

  describe('T14: Exact page boundary (50)', () => {
    it('should handle exactly 50 users (one page)', async () => {
      const mockUsers = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      apiClient.get.mockResolvedValueOnce({
        users: mockUsers,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 50 },
      });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(50);
      expect(apiClient.get).toHaveBeenCalledTimes(1);
    });
  });

  describe('T15: Exact page boundary (100)', () => {
    it('should handle exactly 100 users (two pages)', async () => {
      const page1Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      const page2Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 51,
        email: `user${i + 51}@test.com`,
        full_name: `User ${i + 51}`,
        role: 'student',
      }));

      apiClient.get
        .mockResolvedValueOnce({
          users: page1Users,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 100 },
        })
        .mockResolvedValueOnce({
          users: page2Users,
          meta: { total_pages: 2, page: 2, per_page: 50, total: 100 },
        });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(100);
      expect(apiClient.get).toHaveBeenCalledTimes(2);
    });
  });

  describe('T16: Empty pages handling', () => {
    it('should handle empty last page gracefully', async () => {
      const page1Users = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        email: `user${i + 1}@test.com`,
        full_name: `User ${i + 1}`,
        role: 'student',
      }));

      apiClient.get
        .mockResolvedValueOnce({
          users: page1Users,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 50 },
        })
        .mockResolvedValueOnce({
          users: [],
          meta: { total_pages: 2, page: 2, per_page: 50, total: 50 },
        });

      const result = await usersAPI.getUsersAll();

      expect(result).toHaveLength(50);
      expect(result).toEqual(page1Users);
    });
  });
});
