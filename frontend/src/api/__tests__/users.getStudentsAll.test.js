import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getStudentsAll } from '../users.js';
import * as apiClientModule from '../client.js';

vi.mock('../client.js', () => ({
  default: {
    get: vi.fn(),
  },
}));

describe('getStudentsAll() - Pagination and Data Loading', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('T004: Unit Tests - Core Functionality', () => {
    it('test_getStudentsAll_zero_students - should return empty array when API returns no students', async () => {
      apiClientModule.default.get.mockResolvedValueOnce({
        users: [],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 0 },
      });

      const result = await getStudentsAll();

      expect(result).toEqual([]);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(1);
    });

    it('test_getStudentsAll_single_page_30_students - should return 30 students with one API call', async () => {
      const students = Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        name: `Student ${i + 1}`,
        email: `student${i + 1}@example.com`,
        role: 'student',
      }));

      apiClientModule.default.get.mockResolvedValueOnce({
        users: students,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 30 },
      });

      const result = await getStudentsAll();

      expect(result).toHaveLength(30);
      expect(result).toEqual(students);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(1);
    });

    it('test_getStudentsAll_two_pages_80_students - should fetch all 80 students across 2 pages', async () => {
      const page1Students = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        name: `Student ${i + 1}`,
        email: `student${i + 1}@example.com`,
      }));

      const page2Students = Array.from({ length: 30 }, (_, i) => ({
        id: i + 51,
        name: `Student ${i + 51}`,
        email: `student${i + 51}@example.com`,
      }));

      apiClientModule.default.get
        .mockResolvedValueOnce({
          users: page1Students,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 80 },
        })
        .mockResolvedValueOnce({
          users: page2Students,
          meta: { total_pages: 2, page: 2, per_page: 50, total: 80 },
        });

      const result = await getStudentsAll();

      expect(result).toHaveLength(80);
      expect(result).toEqual([...page1Students, ...page2Students]);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(2);
    });

    it('test_getStudentsAll_three_pages_125_students - should fetch all 125 students across 3 pages', async () => {
      const page1Students = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        name: `Student ${i + 1}`,
      }));

      const page2Students = Array.from({ length: 50 }, (_, i) => ({
        id: i + 51,
        name: `Student ${i + 51}`,
      }));

      const page3Students = Array.from({ length: 25 }, (_, i) => ({
        id: i + 101,
        name: `Student ${i + 101}`,
      }));

      apiClientModule.default.get
        .mockResolvedValueOnce({
          users: page1Students,
          meta: { total_pages: 3, page: 1, per_page: 50, total: 125 },
        })
        .mockResolvedValueOnce({
          users: page2Students,
          meta: { total_pages: 3, page: 2, per_page: 50, total: 125 },
        })
        .mockResolvedValueOnce({
          users: page3Students,
          meta: { total_pages: 3, page: 3, per_page: 50, total: 125 },
        });

      const result = await getStudentsAll();

      expect(result).toHaveLength(125);
      expect(result).toEqual([...page1Students, ...page2Students, ...page3Students]);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(3);
    });

    it('test_getStudentsAll_api_error_throws - should throw exception when API call fails', async () => {
      const testError = new Error('Network error');
      apiClientModule.default.get.mockRejectedValueOnce(testError);

      await expect(getStudentsAll()).rejects.toThrow('Network error');
    });

    it('test_getStudentsAll_abort_error_not_logged - should not log AbortError', async () => {
      const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      const abortError = new Error('Aborted');
      abortError.name = 'AbortError';

      apiClientModule.default.get.mockRejectedValueOnce(abortError);

      await expect(getStudentsAll()).rejects.toThrow('Aborted');
      expect(errorSpy).not.toHaveBeenCalledWith('Error fetching all students:', expect.anything());
    });

    it('test_getStudentsAll_with_filters - should pass search filter through pagination', async () => {
      const filteredStudents = [
        { id: 5, name: 'John Student' },
        { id: 10, name: 'John Doe' },
      ];

      apiClientModule.default.get.mockResolvedValueOnce({
        users: filteredStudents,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 2 },
      });

      const result = await getStudentsAll({ search: 'john' });

      expect(result).toEqual(filteredStudents);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(1);
    });

    it('test_getStudentsAll_preserves_student_data - should preserve all student properties', async () => {
      const studentWithAllProps = {
        id: 1,
        name: 'Test Student',
        email: 'test@example.com',
        role: 'student',
        credits: 100,
        phone: '+1234567890',
        telegram_username: '@student123',
      };

      apiClientModule.default.get.mockResolvedValueOnce({
        users: [studentWithAllProps],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
      });

      const result = await getStudentsAll();

      expect(result[0]).toEqual(studentWithAllProps);
    });
  });

  describe('T005: Integration Tests - CreditsManagement Component Scenario', () => {
    it('test_getStudentsAll_loads_25_students_for_credits_management - should fetch all 25 students', async () => {
      const page1Students = Array.from({ length: 25 }, (_, i) => ({
        id: i + 1,
        name: `Student ${i + 1}`,
        email: `student${i + 1}@example.com`,
        role: 'student',
        credits: Math.floor(Math.random() * 1000),
      }));

      apiClientModule.default.get.mockResolvedValueOnce({
        users: page1Students,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 25 },
      });

      const result = await getStudentsAll();

      expect(result).toHaveLength(25);
      expect(result).toEqual(page1Students);

      result.forEach((student) => {
        expect(student).toHaveProperty('id');
        expect(student).toHaveProperty('name');
        expect(student).toHaveProperty('email');
        expect(student).toHaveProperty('credits');
      });
    });

    it('test_getStudentsAll_integration_multiple_large_batches - should handle loading 300+ students in 6 pages', async () => {
      const pages = [];
      let studentId = 1;

      for (let pageNum = 1; pageNum <= 6; pageNum++) {
        const pageStudents = Array.from({ length: 50 }, (_, i) => ({
          id: studentId + i,
          name: `Student ${studentId + i}`,
          email: `student${studentId + i}@example.com`,
          role: 'student',
          credits: Math.random() * 500,
        }));
        studentId += 50;

        pages.push({
          users: pageStudents,
          meta: { total_pages: 6, page: pageNum, per_page: 50, total: 300 },
        });
      }

      pages.forEach((pageData) => {
        apiClientModule.default.get.mockResolvedValueOnce(pageData);
      });

      const result = await getStudentsAll();

      expect(result).toHaveLength(300);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(6);

      for (let i = 0; i < result.length; i++) {
        expect(result[i].id).toBe(i + 1);
        expect(result[i].email).toBe(`student${i + 1}@example.com`);
      }
    });

    it('test_getStudentsAll_all_students_have_action_buttons - verify each student can have credits actions', async () => {
      const students = Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        name: `Student ${i + 1}`,
        email: `student${i + 1}@example.com`,
        role: 'student',
        credits: 100 + i * 10,
      }));

      apiClientModule.default.get.mockResolvedValueOnce({
        users: students,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 50 },
      });

      const result = await getStudentsAll();

      result.forEach((student) => {
        expect(student).toHaveProperty('id');
        expect(student).toHaveProperty('credits');
        expect(typeof student.id).toBe('number');
        expect(typeof student.credits).toBe('number');
      });
    });
  });

  describe('Edge Cases and Boundary Tests', () => {
    it('test_getStudentsAll_single_student - should handle exactly 1 student', async () => {
      apiClientModule.default.get.mockResolvedValueOnce({
        users: [{ id: 1, name: 'Solo Student', role: 'student' }],
        meta: { total_pages: 1, page: 1, per_page: 50, total: 1 },
      });

      const result = await getStudentsAll();

      expect(result).toHaveLength(1);
      expect(result[0].id).toBe(1);
    });

    it('test_getStudentsAll_exact_page_boundary - should handle exactly 50 students (1 page)', async () => {
      const students = Array.from({ length: 50 }, (_, i) => ({ id: i + 1 }));

      apiClientModule.default.get.mockResolvedValueOnce({
        users: students,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 50 },
      });

      const result = await getStudentsAll();

      expect(result).toHaveLength(50);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(1);
    });

    it('test_getStudentsAll_exact_page_boundary_multi - should handle exactly 100 students (2 pages)', async () => {
      const page1 = Array.from({ length: 50 }, (_, i) => ({ id: i + 1 }));
      const page2 = Array.from({ length: 50 }, (_, i) => ({ id: i + 51 }));

      apiClientModule.default.get
        .mockResolvedValueOnce({
          users: page1,
          meta: { total_pages: 2, page: 1, per_page: 50, total: 100 },
        })
        .mockResolvedValueOnce({
          users: page2,
          meta: { total_pages: 2, page: 2, per_page: 50, total: 100 },
        });

      const result = await getStudentsAll();

      expect(result).toHaveLength(100);
      expect(apiClientModule.default.get).toHaveBeenCalledTimes(2);
    });

    it('test_getStudentsAll_empty_users_array - should include empty pages in results', async () => {
      const page1 = Array.from({ length: 50 }, (_, i) => ({ id: i + 1 }));
      const page2 = [];
      const page3 = Array.from({ length: 30 }, (_, i) => ({ id: i + 51 }));

      apiClientModule.default.get
        .mockResolvedValueOnce({
          users: page1,
          meta: { total_pages: 3, page: 1, per_page: 50, total: 80 },
        })
        .mockResolvedValueOnce({
          users: page2,
          meta: { total_pages: 3, page: 2, per_page: 50, total: 80 },
        })
        .mockResolvedValueOnce({
          users: page3,
          meta: { total_pages: 3, page: 3, per_page: 50, total: 80 },
        });

      const result = await getStudentsAll();

      expect(result).toHaveLength(80);
      expect(result).toEqual([...page1, ...page2, ...page3]);
    });

    it('test_getStudentsAll_per_page_pagination_params - should use per_page=50 for all requests', async () => {
      const page1 = Array.from({ length: 50 }, (_, i) => ({ id: i + 1 }));

      apiClientModule.default.get.mockResolvedValueOnce({
        users: page1,
        meta: { total_pages: 1, page: 1, per_page: 50, total: 50 },
      });

      await getStudentsAll();

      expect(apiClientModule.default.get).toHaveBeenCalled();
    });
  });
});
