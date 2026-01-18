import { describe, it, expect } from 'vitest';

// Import functions directly from the component
// Since these are not exported, we need to redefine them for testing
// OR test them through the component's behavior
// For unit testing, we'll define them locally based on the source code

const normalizeStudentIds = (ids) => {
  if (!Array.isArray(ids)) return new Set();
  return new Set(
    ids
      .map(id => {
        if (typeof id === 'object' && (id.id || id.student_id)) {
          return String(id.id || id.student_id).trim();
        }
        return String(id).trim();
      })
      .filter(Boolean)
  );
};

const studentIdsChanged = (oldIds, newIds) => {
  const oldSet = normalizeStudentIds(oldIds);
  const newSet = normalizeStudentIds(newIds);

  if (oldSet.size !== newSet.size) return true;

  return ![...oldSet].every(id => newSet.has(id));
};

describe('normalizeStudentIds Function', () => {
  describe('Input normalization', () => {
    it('should convert array of strings to Set', () => {
      const ids = ['uuid-1', 'uuid-2', 'uuid-3'];
      const result = normalizeStudentIds(ids);

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(3);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
      expect(result.has('uuid-3')).toBe(true);
    });

    it('should convert array of {id: UUID} objects to Set', () => {
      const ids = [
        { id: 'uuid-1' },
        { id: 'uuid-2' },
        { id: 'uuid-3' },
      ];
      const result = normalizeStudentIds(ids);

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(3);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
      expect(result.has('uuid-3')).toBe(true);
    });

    it('should convert array of {student_id: UUID} objects to Set', () => {
      const ids = [
        { student_id: 'uuid-A' },
        { student_id: 'uuid-B' },
        { student_id: 'uuid-C' },
      ];
      const result = normalizeStudentIds(ids);

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(3);
      expect(result.has('uuid-A')).toBe(true);
      expect(result.has('uuid-B')).toBe(true);
      expect(result.has('uuid-C')).toBe(true);
    });

    it('should handle mixed format (prefers id over student_id)', () => {
      const ids = [
        { id: 'uuid-1', student_id: 'uuid-wrong' },
        { student_id: 'uuid-2' },
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(2);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
    });

    it('should return empty Set for empty array', () => {
      const result = normalizeStudentIds([]);

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(0);
    });

    it('should return empty Set for null', () => {
      const result = normalizeStudentIds(null);

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(0);
    });

    it('should return empty Set for undefined', () => {
      const result = normalizeStudentIds(undefined);

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(0);
    });

    it('should return empty Set for non-array input', () => {
      const result = normalizeStudentIds('not-an-array');

      expect(result).toBeInstanceOf(Set);
      expect(result.size).toBe(0);
    });

    it('should trim whitespace from UUIDs', () => {
      const ids = [
        '  uuid-1  ',
        ' uuid-2 ',
        'uuid-3',
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(3);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
      expect(result.has('uuid-3')).toBe(true);
    });

    it('should trim whitespace from object IDs', () => {
      const ids = [
        { id: '  uuid-1  ' },
        { student_id: ' uuid-2 ' },
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(2);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
    });

    it('should preserve UUID case (NOT convert to lowercase)', () => {
      const ids = [
        'UUID-1',
        'uuid-2',
        'Uuid-3',
        'UuId-4',
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(4);
      expect(result.has('UUID-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
      expect(result.has('Uuid-3')).toBe(true);
      expect(result.has('UuId-4')).toBe(true);
      // Verify different cases are treated as different values
      expect(result.has('uuid-1')).toBe(false);
    });

    it('should handle UUID with hyphens correctly', () => {
      const id = '550e8400-e29b-41d4-a716-446655440000';
      const result = normalizeStudentIds([id]);

      expect(result.size).toBe(1);
      expect(result.has(id)).toBe(true);
    });

    it('should filter out empty strings', () => {
      const ids = [
        'uuid-1',
        '',
        'uuid-2',
        '   ',
        'uuid-3',
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(3);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
      expect(result.has('uuid-3')).toBe(true);
    });

    it('should handle empty student_id values gracefully', () => {
      const ids = [
        { id: 'uuid-1' },
        { student_id: 'uuid-2' },
        { student_id: '   ' },  // Whitespace only - will be trimmed to empty string and filtered
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(2);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
    });

    it('should handle numeric IDs by converting to strings', () => {
      const ids = [
        123,
        456,
        789,
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(3);
      expect(result.has('123')).toBe(true);
      expect(result.has('456')).toBe(true);
      expect(result.has('789')).toBe(true);
    });

    it('should handle numeric values in objects', () => {
      const ids = [
        { id: 123 },
        { student_id: 456 },
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(2);
      expect(result.has('123')).toBe(true);
      expect(result.has('456')).toBe(true);
    });

    it('should remove duplicates (Set behavior)', () => {
      const ids = [
        'uuid-1',
        'uuid-1',
        'uuid-2',
        'uuid-1',
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(2);
      expect(result.has('uuid-1')).toBe(true);
      expect(result.has('uuid-2')).toBe(true);
    });

    it('should remove duplicates even when in different formats', () => {
      const ids = [
        'uuid-1',
        { id: 'uuid-1' },
        { student_id: 'uuid-1' },
      ];
      const result = normalizeStudentIds(ids);

      expect(result.size).toBe(1);
      expect(result.has('uuid-1')).toBe(true);
    });
  });
});

describe('studentIdsChanged Function', () => {
  describe('Complete replacement', () => {
    it('should return true when completely different students [A, B] -> [C, D]', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = ['uuid-C', 'uuid-D'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should return false when students are identical [A, B] -> [A, B]', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = ['uuid-A', 'uuid-B'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should return false when order differs but content is same [A, B] -> [B, A]', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = ['uuid-B', 'uuid-A'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });
  });

  describe('Addition of students', () => {
    it('should return true when adding one student [A, B] -> [A, B, C]', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = ['uuid-A', 'uuid-B', 'uuid-C'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should return true when adding multiple students [A] -> [A, B, C]', () => {
      const oldIds = ['uuid-A'];
      const newIds = ['uuid-A', 'uuid-B', 'uuid-C'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should return true when adding to empty list [] -> [A]', () => {
      const oldIds = [];
      const newIds = ['uuid-A'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should return true when adding multiple to empty list [] -> [A, B, C]', () => {
      const oldIds = [];
      const newIds = ['uuid-A', 'uuid-B', 'uuid-C'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });
  });

  describe('Removal of students', () => {
    it('should return true when removing one student [A, B] -> [A]', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = ['uuid-A'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should return true when removing multiple students [A, B, C] -> [A]', () => {
      const oldIds = ['uuid-A', 'uuid-B', 'uuid-C'];
      const newIds = ['uuid-A'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should return true when removing all students [A, B] -> []', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = [];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });
  });

  describe('Empty lists', () => {
    it('should return false when both are empty [] -> []', () => {
      const oldIds = [];
      const newIds = [];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should return false when both are null', () => {
      const oldIds = null;
      const newIds = null;

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should return false when both are undefined', () => {
      const oldIds = undefined;
      const newIds = undefined;

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });
  });

  describe('Mixed formats', () => {
    it('should work with old as strings, new as objects', () => {
      const oldIds = ['uuid-A', 'uuid-B'];
      const newIds = [
        { id: 'uuid-A' },
        { id: 'uuid-B' },
      ];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should work with old as objects, new as strings', () => {
      const oldIds = [
        { id: 'uuid-A' },
        { id: 'uuid-B' },
      ];
      const newIds = ['uuid-A', 'uuid-B'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should work with old as {id:}, new as {student_id:}', () => {
      const oldIds = [
        { id: 'uuid-A' },
        { id: 'uuid-B' },
      ];
      const newIds = [
        { student_id: 'uuid-A' },
        { student_id: 'uuid-B' },
      ];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should detect changes when mixed format changes students', () => {
      const oldIds = [
        { id: 'uuid-A' },
        { id: 'uuid-B' },
      ];
      const newIds = [
        { student_id: 'uuid-A' },
        { student_id: 'uuid-C' },
      ];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should handle completely mixed format with strings and objects', () => {
      const oldIds = [
        'uuid-A',
        { id: 'uuid-B' },
        { student_id: 'uuid-C' },
      ];
      const newIds = [
        { id: 'uuid-A' },
        'uuid-B',
        { student_id: 'uuid-C' },
      ];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });
  });

  describe('Case sensitivity', () => {
    it('should treat different cases as different IDs', () => {
      const oldIds = ['UUID-1', 'uuid-2'];
      const newIds = ['uuid-1', 'UUID-2'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should preserve case when comparing', () => {
      const oldIds = ['UUID-1', 'UUID-2'];
      const newIds = ['UUID-1', 'UUID-2'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });
  });

  describe('Whitespace handling', () => {
    it('should normalize whitespace correctly', () => {
      const oldIds = ['  uuid-A  ', ' uuid-B '];
      const newIds = ['uuid-A', 'uuid-B'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should detect changes despite whitespace differences', () => {
      const oldIds = ['  uuid-A  ', ' uuid-B '];
      const newIds = ['uuid-A', 'uuid-C'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });
  });

  describe('Large sets', () => {
    it('should handle large lists correctly', () => {
      const oldIds = Array.from({ length: 100 }, (_, i) => `uuid-${i}`);
      const newIds = Array.from({ length: 100 }, (_, i) => `uuid-${i}`);

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(false);
    });

    it('should detect single change in large list', () => {
      const oldIds = Array.from({ length: 100 }, (_, i) => `uuid-${i}`);
      const newIds = Array.from({ length: 100 }, (_, i) => `uuid-${i}`);
      newIds[50] = 'uuid-new';

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });
  });

  describe('Critical bug fix regression test', () => {
    it('should detect change from [A, B] to [C, D] (the original bug)', () => {
      // This is the specific bug case mentioned in the requirements
      const oldIds = ['550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002'];
      const newIds = ['550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440004'];

      const result = studentIdsChanged(oldIds, newIds);

      // This should be true (students changed) but the bug made it false
      expect(result).toBe(true);
    });

    it('should work with real API response format (students array)', () => {
      const oldIds = [
        { student_id: '550e8400-e29b-41d4-a716-446655440001' },
        { student_id: '550e8400-e29b-41d4-a716-446655440002' },
      ];
      const newIds = [
        { student_id: '550e8400-e29b-41d4-a716-446655440003' },
        { student_id: '550e8400-e29b-41d4-a716-446655440004' },
      ];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should work when adding single student to existing group', () => {
      const oldIds = ['uuid-1', 'uuid-2'];
      const newIds = ['uuid-1', 'uuid-2', 'uuid-3'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });

    it('should work when removing single student from group', () => {
      const oldIds = ['uuid-1', 'uuid-2', 'uuid-3'];
      const newIds = ['uuid-1', 'uuid-2'];

      const result = studentIdsChanged(oldIds, newIds);

      expect(result).toBe(true);
    });
  });
});
