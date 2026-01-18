import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getLessons, getMyLessons, getLessonById, createLesson, updateLesson, deleteLesson, getAvailableSlots, getLessonStudents } from '../lessons.js';

describe('Lessons API', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Mock fetch globally
    global.fetch = vi.fn();

    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('getLessons - Filter Parameter Handling', () => {
    it('should call /lessons without query params when no filters provided', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons();

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // URL should contain /lessons endpoint
      expect(url).toContain('/lessons');
      // URL should NOT have query parameters
      expect(url).not.toContain('?');
    });

    it('should pass teacher_id filter as query parameter', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ teacher_id: 5 });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('?teacher_id=5');
    });

    it('should pass date filter as query parameter', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ date: '2025-01-01' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('?date=2025-01-01');
    });

    it('should pass available filter as query parameter', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ available: true });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('?available=true');
    });

    it('should combine multiple filters correctly', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({
        teacher_id: 3,
        date: '2025-01-01',
        available: true,
      });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('teacher_id=3');
      expect(url).toContain('date=2025-01-01');
      expect(url).toContain('available=true');
      expect(url).toContain('&');
    });

    it('should NOT send subject parameter (unsupported by backend)', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ subject: 'Math' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).not.toContain('subject');
      expect(url).not.toContain('Math');
    });

    it('should NOT send start_date parameter (unsupported by backend)', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ start_date: '2025-01-01' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).not.toContain('start_date');
      expect(url).not.toContain('2025-01-01');
    });

    it('should NOT send end_date parameter (unsupported by backend)', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ end_date: '2025-12-31' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).not.toContain('end_date');
      expect(url).not.toContain('2025-12-31');
    });

    it('should ignore unsupported parameters when mixed with supported ones', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({
        teacher_id: 5,
        subject: 'Math',
        start_date: '2025-01-01',
        end_date: '2025-12-31',
        available: true,
      });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // Should contain supported parameters
      expect(url).toContain('teacher_id=5');
      expect(url).toContain('available=true');

      // Should NOT contain unsupported parameters
      expect(url).not.toContain('subject');
      expect(url).not.toContain('start_date');
      expect(url).not.toContain('end_date');
    });

    it('should handle available=false filter correctly', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ available: false });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('?available=false');
    });

    it('should handle teacher_id as UUID correctly', async () => {
      const uuidValue = '550e8400-e29b-41d4-a716-446655440000';
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({ teacher_id: uuidValue });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain(`teacher_id=${uuidValue}`);
    });

    it('should encode special characters in parameter values', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      // Using a date with special format
      await getLessons({ date: '2025-01-01 10:30' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // Special characters should be encoded
      expect(url).toContain('date=');
      // The URL should still be valid
      expect(url).toContain('?');
    });

    it('should ignore undefined filter values', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });

      await getLessons({
        teacher_id: 3,
        date: undefined,
      });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('teacher_id=3');
      expect(url).not.toContain('date=undefined');
    });

    it('should extract lessons array from response', async () => {
      const mockLessons = [
        { id: 1, subject: 'Math', teacher_id: 5 },
        { id: 2, subject: 'Physics', teacher_id: 3 },
      ];

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockLessons }),
      });

      const result = await getLessons();

      expect(result).toEqual(mockLessons);
    });

    it('should handle response without nested structure', async () => {
      const mockLessons = [
        { id: 1, subject: 'Math' },
      ];

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockLessons }),
      });

      const result = await getLessons();

      expect(result).toEqual(mockLessons);
    });
  });

  describe('getMyLessons - Response Handling', () => {
    it('should call /lessons/my endpoint without parameters', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getMyLessons();

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons/my');
      expect(url).not.toContain('?');
    });

    it('should extract lessons array from nested response.data', async () => {
      const mockLessons = [
        { id: 1, subject: 'Math' },
      ];

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockLessons }),
      });

      const result = await getMyLessons();

      expect(result).toEqual(mockLessons);
      expect(Array.isArray(result)).toBe(true);
    });

    it('should handle response with direct array format', async () => {
      const mockLessons = [
        { id: 1, subject: 'Math' },
        { id: 2, subject: 'Physics' },
      ];

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => mockLessons,
      });

      const result = await getMyLessons();

      expect(result).toEqual(mockLessons);
    });

    it('should handle multiple response formats correctly', async () => {
      const mockLessons = [
        { id: 1, subject: 'Math' },
        { id: 2, subject: 'Physics' },
      ];

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockLessons }),
      });

      const result = await getMyLessons();

      expect(result).toEqual(mockLessons);
      expect(result.length).toBe(2);
      expect(result[0].id).toBe(1);
    });

    it('should handle getMyLessons with empty array', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const result = await getMyLessons();

      expect(result).toEqual([]);
      expect(Array.isArray(result)).toBe(true);
    });
  });

  describe('getLessonById', () => {
    it('should call /lessons/:id endpoint', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { id: 42, subject: 'Math' } }),
      });

      await getLessonById(42);

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons/42');
    });

    it('should return lesson data', async () => {
      const mockLesson = { id: 42, subject: 'Math', teacher_id: 5 };

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockLesson }),
      });

      const result = await getLessonById(42);

      expect(result).toEqual(mockLesson);
    });
  });

  describe('createLesson', () => {
    it('should POST to /lessons endpoint', async () => {
      const lessonData = {
        subject: 'Math',
        teacher_id: 5,
        start_time: '2025-01-15T10:00:00Z',
        duration_minutes: 60,
      };

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 201,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { id: 1, ...lessonData } }),
      });

      await createLesson(lessonData);

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];
      const config = callArgs[1];

      expect(url).toContain('/lessons');
      expect(config.method).toBe('POST');
      expect(config.body).toBe(JSON.stringify(lessonData));
    });
  });

  describe('updateLesson', () => {
    it('should PUT to /lessons/:id endpoint', async () => {
      const updates = { subject: 'Advanced Math' };

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { id: 42, ...updates } }),
      });

      await updateLesson(42, updates);

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];
      const config = callArgs[1];

      expect(url).toContain('/lessons/42');
      expect(config.method).toBe('PUT');
      expect(config.body).toBe(JSON.stringify(updates));
    });
  });

  describe('deleteLesson', () => {
    it('should DELETE to /lessons/:id endpoint', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({}),
      });

      await deleteLesson(42);

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];
      const config = callArgs[1];

      expect(url).toContain('/lessons/42');
      expect(config.method).toBe('DELETE');
    });
  });

  describe('getAvailableSlots', () => {
    it('should call getLessons with available=true', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getAvailableSlots();

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('available=true');
    });

    it('should merge additional filters with available=true', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getAvailableSlots({ teacher_id: 5, date: '2025-01-01' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('available=true');
      expect(url).toContain('teacher_id=5');
      expect(url).toContain('date=2025-01-01');
    });
  });

  describe('getLessonStudents', () => {
    it('should call /lessons/:id/students endpoint', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { students: [] } }),
      });

      await getLessonStudents(42);

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons/42/students');
    });

    it('should return students data', async () => {
      const mockData = {
        students: [
          { id: 1, name: 'Student 1' },
          { id: 2, name: 'Student 2' },
        ],
      };

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockData }),
      });

      const result = await getLessonStudents(42);

      expect(result).toEqual(mockData);
    });
  });

  describe('Error Handling and Response Validation', () => {
    it('should handle API errors in getLessons', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: false, error: { message: 'Internal Server Error' } }),
      });

      await expect(getLessons()).rejects.toThrow();
    });

    it('should handle network errors', async () => {
      global.fetch.mockRejectedValueOnce(new TypeError('Network error'));

      await expect(getLessons()).rejects.toThrow();
    });

    it('should handle null response gracefully in getLessons', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => null,
      });

      const result = await getLessons();
      expect(result).toBeNull();
    });

    it('should handle undefined response gracefully in getLessons', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => undefined,
      });

      const result = await getLessons();
      expect(result).toBeUndefined();
    });

    it('should handle null response gracefully in getMyLessons', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => null,
      });

      const result = await getMyLessons();
      expect(result).toBeNull();
    });

    it('should handle undefined response gracefully in getMyLessons', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => undefined,
      });

      const result = await getMyLessons();
      expect(result).toBeUndefined();
    });

    it('should properly throw error responses', async () => {
      const errorResponse = {
        success: false,
        error: { message: 'Unauthorized', code: 401 }
      };

      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => errorResponse,
      });

      await expect(getLessons()).rejects.toThrow();
    });
  });

  describe('Query String Construction and Parameter Encoding', () => {
    it('should create clean URL without parameters when filters are empty', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({});

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons');
      expect(url).not.toContain('?');
    });

    it('should not add query string separator when all filters are undefined', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ teacher_id: undefined, available: undefined });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).not.toContain('?');
    });

    it('should properly separate multiple parameters with ampersand', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({
        teacher_id: 123,
        date: '2025-06-15',
      });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // Should have exactly one & separating the two parameters
      const ampCount = (url.match(/&/g) || []).length;
      expect(ampCount).toBe(1);
      expect(url).toContain('?');
    });

    it('should properly encode date parameter with dashes', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ date: '2025-12-31' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('date=2025-12-31');
    });

    it('should handle numeric teacher_id properly', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ teacher_id: 12345 });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('teacher_id=12345');
    });

    it('should properly encode boolean values', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ available: true });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('available=true');
    });

    it('should encode false boolean value correctly', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ available: false });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('available=false');
    });

    it('should not include parameter if value is null', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ teacher_id: null, available: true });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('available=true');
      expect(url).not.toContain('teacher_id');
    });

    it('should handle only available parameter without other filters', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ available: true });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons?available=true');
      expect(url).not.toContain('teacher_id');
      expect(url).not.toContain('date');
    });

    it('should handle only date parameter without other filters', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ date: '2025-01-01' });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons?date=2025-01-01');
      expect(url).not.toContain('teacher_id');
      expect(url).not.toContain('available');
    });

    it('should handle only teacher_id parameter without other filters', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ teacher_id: 999 });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      expect(url).toContain('/lessons?teacher_id=999');
      expect(url).not.toContain('date');
      expect(url).not.toContain('available');
    });
  });

  describe('Edge Cases - Response Handling', () => {
    it('should handle null response data', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: null }),
      });

      const result = await getLessons();
      expect(result).toBeNull();
    });

    it('should handle undefined response data', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: undefined }),
      });

      const result = await getLessons();
      // Should return the undefined data field explicitly
      expect(result).toBeUndefined();
    });

    it('should handle empty lessons array in nested response', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const result = await getLessons();
      expect(result).toEqual([]);
    });

    it('should handle response without count field', async () => {
      const mockLessons = [{ id: 1, subject: 'Math' }];
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: mockLessons }),
      });

      const result = await getLessons();
      expect(result).toEqual(mockLessons);
    });

    it('should handle response with missing lessons field', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { count: 5 } }),
      });

      const result = await getLessons();
      // Should return the object as-is when data doesn't parse to lessons array
      expect(result).toEqual({ count: 5 });
    });

    it('should handle lessons with missing optional fields', async () => {
      const lessonsWithMissingFields = [
        { id: 1, subject: 'Math' }, // missing teacher_id, teacher_name
        { id: 2, teacher_id: 5 }, // missing subject
      ];

      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: lessonsWithMissingFields }),
      });

      const result = await getLessons();
      expect(result).toHaveLength(2);
      expect(result[0].id).toBe(1);
      expect(result[1].id).toBe(2);
    });

    it('should handle getMyLessons with null response', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: null }),
      });

      const result = await getMyLessons();
      expect(result).toBeNull();
    });

    it('should handle getLessonById with null response', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: null }),
      });

      const result = await getLessonById(999);
      expect(result).toBeNull();
    });

    it('should handle getLessonStudents with empty students array', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: { students: [] } }),
      });

      const result = await getLessonStudents(42);
      expect(result.students).toEqual([]);
    });

    it('should handle getLessonStudents with missing students field', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: {} }),
      });

      const result = await getLessonStudents(42);
      expect(result).toEqual({});
    });
  });

  describe('getLessons - StudentIDs Filter Parameter Handling', () => {
    it('should pass student_ids as comma-separated array in URL', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const uuid1 = '550e8400-e29b-41d4-a716-446655440000';
      const uuid2 = '660e8400-e29b-41d4-a716-446655440001';

      await getLessons({ student_ids: [uuid1, uuid2] });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // URL should contain student_ids with both UUIDs
      // Comma may be URL-encoded as %2C
      expect(url).toContain('student_ids=');
      expect(url).toContain(uuid1);
      expect(url).toContain(uuid2);
      // Check for comma or encoded comma
      expect(url).toMatch(/[,|%2C]/);
    });

    it('should handle empty student_ids array (no parameter in URL)', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      await getLessons({ student_ids: [] });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // Empty array should NOT append student_ids parameter
      expect(url).not.toContain('student_ids');
    });

    it('should combine student_ids with other filters correctly', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const teacherUUID = '550e8400-e29b-41d4-a716-446655440000';
      const studentUUID1 = '660e8400-e29b-41d4-a716-446655440001';
      const studentUUID2 = '770e8400-e29b-41d4-a716-446655440002';

      await getLessons({
        teacher_id: teacherUUID,
        student_ids: [studentUUID1, studentUUID2],
        date: '2026-01-11',
      });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // All three filters should be in URL
      expect(url).toContain(`teacher_id=${teacherUUID}`);
      expect(url).toContain('student_ids=');
      expect(url).toContain(studentUUID1);
      expect(url).toContain(studentUUID2);
      expect(url).toContain('date=2026-01-11');
      expect(url).toContain('&');
    });

    it('should handle single student_id in array correctly', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const uuid = '550e8400-e29b-41d4-a716-446655440000';

      await getLessons({ student_ids: [uuid] });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // Single UUID should be passed without trailing comma
      expect(url).toContain(`student_ids=${uuid}`);
      expect(url).not.toContain(`${uuid},`);
    });

    it('should NOT pass student_ids parameter if null or undefined', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      // Test with null
      await getLessons({ student_ids: null });
      let url = global.fetch.mock.calls[0][0];
      expect(url).not.toContain('student_ids');

      global.fetch.mockClear();
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      // Test with undefined
      await getLessons({ student_ids: undefined });
      url = global.fetch.mock.calls[0][0];
      expect(url).not.toContain('student_ids');
    });

    it('should properly URL-encode student_ids with special characters', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const uuid = '550e8400-e29b-41d4-a716-446655440000';

      await getLessons({ student_ids: [uuid] });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // UUIDs should be passed as-is (no encoding of hyphens in standard UUIDs)
      expect(url).toContain(`student_ids=${uuid}`);
    });

    it('should work with student_ids and available filter together', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ success: true, data: [] }),
      });

      const uuid1 = '550e8400-e29b-41d4-a716-446655440000';
      const uuid2 = '660e8400-e29b-41d4-a716-446655440001';

      await getLessons({
        student_ids: [uuid1, uuid2],
        available: true,
      });

      const callArgs = global.fetch.mock.calls[0];
      const url = callArgs[0];

      // Both filters should be present
      expect(url).toContain('student_ids=');
      expect(url).toContain(uuid1);
      expect(url).toContain(uuid2);
      expect(url).toContain('available=true');
      expect(url).toContain('&');
    });
  });
});
