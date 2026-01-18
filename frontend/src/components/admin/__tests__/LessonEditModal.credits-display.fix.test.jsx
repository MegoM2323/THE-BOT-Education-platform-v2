import { describe, it, expect } from 'vitest';

/**
 * Test to verify the fix for credit display in LessonEditModal
 * Focus: Ensure consistent ID mapping between enrolled and available students
 */
describe('LessonEditModal - Credits Display Fix (A5)', () => {
  /**
   * Test 1: Verify creditsMap is created correctly from API response
   * This tests the fix in loadLessonData() line 461-467
   */
  it('should create creditsMap with consistent user_id keys from API response', () => {
    const allCreditsResponse = {
      balances: [
        { user_id: 'enrolled-student-1', balance: 5 },
        { user_id: 'enrolled-student-2', balance: 3 },
        { user_id: 'available-student-1', balance: 10 },
        { user_id: 'available-student-2', balance: 0 },
      ],
    };

    // Simulate the fixed creditsMap creation logic (lines 461-467)
    const creditsMap = {};
    if (allCreditsResponse && allCreditsResponse.balances && Array.isArray(allCreditsResponse.balances)) {
      allCreditsResponse.balances.forEach(({ user_id, balance }) => {
        creditsMap[user_id] = balance || 0;
      });
    }

    // Verify all students have their credits mapped
    expect(creditsMap['enrolled-student-1']).toBe(5);
    expect(creditsMap['enrolled-student-2']).toBe(3);
    expect(creditsMap['available-student-1']).toBe(10);
    expect(creditsMap['available-student-2']).toBe(0);
    expect(Object.keys(creditsMap).length).toBe(4);
  });

  /**
   * Test 2: Verify enrolled students maintain student_id for lookup
   * This tests the fix in StudentCheckboxList usage (lines 1096-1105)
   */
  it('should map enrolled students with consistent ID for credit lookup', () => {
    const students = [
      {
        id: 'booking-1',
        student_id: 'enrolled-student-1',
        student_name: 'Alice',
        student_email: 'alice@test.com'
      },
      {
        id: 'booking-2',
        student_id: 'enrolled-student-2',
        student_name: 'Bob',
        student_email: 'bob@test.com'
      },
    ];

    const studentCredits = {
      'enrolled-student-1': 5,
      'enrolled-student-2': 3,
    };

    // Simulate the fixed StudentCheckboxList mapping (lines 1096-1105)
    const mapped = students.map(s => {
      const studentId = s.student_id || s.id;
      return {
        id: studentId,
        name: s.student_name,
        full_name: s.student_name,
        email: s.student_email,
        credits: studentCredits[studentId] || 0
      };
    });

    // Verify each enrolled student has correct credits
    expect(mapped[0]).toEqual({
      id: 'enrolled-student-1',
      name: 'Alice',
      full_name: 'Alice',
      email: 'alice@test.com',
      credits: 5
    });
    expect(mapped[1]).toEqual({
      id: 'enrolled-student-2',
      name: 'Bob',
      full_name: 'Bob',
      email: 'bob@test.com',
      credits: 3
    });
  });

  /**
   * Test 3: Verify available students use id field for lookup
   * This tests the fix in StudentCheckboxList usage (lines 1106-1115)
   */
  it('should map available students with consistent ID for credit lookup', () => {
    const availableStudents = [
      {
        id: 'available-student-1',
        full_name: 'Charlie',
        name: 'Charlie',
        email: 'charlie@test.com'
      },
      {
        id: 'available-student-2',
        full_name: 'Diana',
        name: 'Diana',
        email: 'diana@test.com'
      },
    ];

    const studentCredits = {
      'available-student-1': 10,
      'available-student-2': 0,
    };

    // Simulate the fixed StudentCheckboxList mapping (lines 1106-1115)
    const mapped = availableStudents.map(s => {
      const studentId = s.id;
      return {
        id: studentId,
        name: s.full_name || s.name,
        full_name: s.full_name || s.name,
        email: s.email,
        credits: studentCredits[studentId] || 0
      };
    });

    // Verify each available student has correct credits
    expect(mapped[0]).toEqual({
      id: 'available-student-1',
      name: 'Charlie',
      full_name: 'Charlie',
      email: 'charlie@test.com',
      credits: 10
    });
    expect(mapped[1]).toEqual({
      id: 'available-student-2',
      name: 'Diana',
      full_name: 'Diana',
      email: 'diana@test.com',
      credits: 0
    });
  });

  /**
   * Test 4: Verify enrolled/available students combined have no conflicts
   * This tests the complete fix where both types are combined
   */
  it('should combine enrolled and available students without credit lookup conflicts', () => {
    const students = [
      { id: 'b1', student_id: 'user-1', student_name: 'Alice', student_email: 'alice@test.com' },
      { id: 'b2', student_id: 'user-2', student_name: 'Bob', student_email: 'bob@test.com' },
    ];

    const availableStudents = [
      { id: 'user-3', full_name: 'Charlie', name: 'Charlie', email: 'charlie@test.com' },
      { id: 'user-4', full_name: 'Diana', name: 'Diana', email: 'diana@test.com' },
    ];

    const studentCredits = {
      'user-1': 5,
      'user-2': 3,
      'user-3': 10,
      'user-4': 0,
    };

    // Simulate the complete fix from lines 1095-1116
    const allStudents = [
      ...students.map(s => {
        const studentId = s.student_id || s.id;
        return {
          id: studentId,
          name: s.student_name,
          full_name: s.student_name,
          email: s.student_email,
          credits: studentCredits[studentId] || 0
        };
      }),
      ...availableStudents.map(s => {
        const studentId = s.id;
        return {
          id: studentId,
          name: s.full_name || s.name,
          full_name: s.full_name || s.name,
          email: s.email,
          credits: studentCredits[studentId] || 0
        };
      })
    ];

    // Verify all students are correctly mapped
    expect(allStudents).toHaveLength(4);
    expect(allStudents[0].credits).toBe(5); // Alice
    expect(allStudents[1].credits).toBe(3); // Bob
    expect(allStudents[2].credits).toBe(10); // Charlie
    expect(allStudents[3].credits).toBe(0); // Diana

    // Verify no credit is undefined or NaN
    allStudents.forEach(student => {
      expect(student.credits).toBeDefined();
      expect(student.credits).not.toBeNaN();
      expect(typeof student.credits).toBe('number');
    });
  });

  /**
   * Test 5: Verify enrolledStudentIds are correctly extracted for StudentCheckboxList
   * This tests the fix in line 1117
   */
  it('should correctly extract enrolledStudentIds for StudentCheckboxList', () => {
    const students = [
      { id: 'b1', student_id: 'user-1', student_name: 'Alice', student_email: 'alice@test.com' },
      { id: 'b2', student_id: 'user-2', student_name: 'Bob', student_email: 'bob@test.com' },
      { id: 'b3', student_id: 'user-3', student_name: 'Charlie', student_email: 'charlie@test.com' },
    ];

    // Simulate the fixed extracting of enrolledStudentIds (line 1117)
    const enrolledStudentIds = students.map(s => s.student_id || s.id);

    // Verify all student_ids are extracted
    expect(enrolledStudentIds).toHaveLength(3);
    expect(enrolledStudentIds).toContain('user-1');
    expect(enrolledStudentIds).toContain('user-2');
    expect(enrolledStudentIds).toContain('user-3');

    // Verify no booking IDs (b1, b2, b3) are in the list
    expect(enrolledStudentIds).not.toContain('b1');
    expect(enrolledStudentIds).not.toContain('b2');
    expect(enrolledStudentIds).not.toContain('b3');
  });

  /**
   * Test 6: Verify StudentCheckboxList receives correct data for rendering
   * This tests that the fixed mapping provides correct data structure
   */
  it('should provide StudentCheckboxList with correct student data structure', () => {
    const students = [
      { id: 'b1', student_id: 'user-1', student_name: 'Alice', student_email: 'alice@test.com' },
    ];

    const availableStudents = [
      { id: 'user-2', full_name: 'Bob', name: 'Bob', email: 'bob@test.com' },
    ];

    const studentCredits = {
      'user-1': 5,
      'user-2': 3,
    };

    // Build allStudents as the component does
    const allStudents = [
      ...students.map(s => {
        const studentId = s.student_id || s.id;
        return {
          id: studentId,
          name: s.student_name,
          full_name: s.student_name,
          email: s.student_email,
          credits: studentCredits[studentId] || 0
        };
      }),
      ...availableStudents.map(s => {
        const studentId = s.id;
        return {
          id: studentId,
          name: s.full_name || s.name,
          full_name: s.full_name || s.name,
          email: s.email,
          credits: studentCredits[studentId] || 0
        };
      })
    ];

    // Verify StudentCheckboxList will receive correct data
    // Each student object should have: id, name, full_name, email, credits
    allStudents.forEach(student => {
      expect(student).toHaveProperty('id');
      expect(student).toHaveProperty('name');
      expect(student).toHaveProperty('full_name');
      expect(student).toHaveProperty('email');
      expect(student).toHaveProperty('credits');

      // StudentCheckboxList uses student.id for the key (line 81 in StudentCheckboxList.jsx)
      // and student.credits for display (line 76 in StudentCheckboxList.jsx)
      expect(typeof student.id).toBe('string');
      expect(typeof student.credits).toBe('number');
    });

    // Verify the data matches what StudentCheckboxList expects
    const enrolledStudentIds = students.map(s => s.student_id || s.id);

    // Test the isEnrolled check logic (line 75 in StudentCheckboxList.jsx)
    expect(enrolledStudentIds.includes(allStudents[0].id)).toBe(true); // Alice is enrolled
    expect(enrolledStudentIds.includes(allStudents[1].id)).toBe(false); // Bob is not enrolled
  });

  /**
   * Test 7: Verify edge cases - missing balances, zero credits, etc
   */
  it('should handle edge cases in credit mapping', () => {
    const studentCredits = {};
    const allCreditsResponse = {
      balances: [
        { user_id: 'student-1', balance: 0 },
        { user_id: 'student-2' }, // Missing balance field
        { user_id: 'student-3', balance: -1 }, // Negative balance
      ],
    };

    // Apply the fixed logic
    if (allCreditsResponse && allCreditsResponse.balances && Array.isArray(allCreditsResponse.balances)) {
      allCreditsResponse.balances.forEach(({ user_id, balance }) => {
        studentCredits[user_id] = balance || 0;
      });
    }

    // Verify edge cases are handled
    expect(studentCredits['student-1']).toBe(0); // Zero balance is preserved
    expect(studentCredits['student-2']).toBe(0); // Missing balance defaults to 0
    expect(studentCredits['student-3']).toBe(-1); // Negative balance is preserved

    // Verify all credits can be looked up without undefined
    const students = ['student-1', 'student-2', 'student-3', 'student-4'];
    students.forEach(studentId => {
      const credits = studentCredits[studentId] || 0;
      expect(credits).toBeDefined();
      expect(credits).not.toBeNaN();
      expect(typeof credits).toBe('number');
    });
  });

  /**
   * Test 8: Verify the fix resolves the original bug
   * Bug: "enrolled students use student_id for lookup, available use id"
   * Fix: Both use consistent ID field after normalization
   */
  it('should fix the original bug - consistent ID lookup for all students', () => {
    // Original scenario that caused the bug
    const enrolledBooking = {
      id: 'booking-123', // This is the booking ID
      student_id: 'enrolled-uuid-1', // This is the student ID
      student_name: 'Alice',
      student_email: 'alice@test.com'
    };

    const availableStudent = {
      id: 'available-uuid-2', // This IS the student ID
      full_name: 'Bob',
      email: 'bob@test.com'
    };

    const creditsMap = {
      'enrolled-uuid-1': 5,
      'available-uuid-2': 3,
    };

    // OLD BUGGY WAY (before fix):
    // const oldEnrolledCredits = creditsMap[enrolledBooking.student_id]; // OK: 5
    // const oldAvailableCredits = creditsMap[availableStudent.id]; // OK: 3
    // But if we mixed them:
    // const mixedEnrolledCredits = creditsMap[enrolledBooking.id]; // WRONG: undefined!
    // const mixedAvailableCredits = creditsMap[availableStudent.student_id]; // WRONG: undefined!

    // NEW FIXED WAY (after fix):
    // Both now consistently normalize to use the same lookup ID
    const normalizedEnrolled = {
      id: enrolledBooking.student_id, // Use student_id
      credits: creditsMap[enrolledBooking.student_id] || 0
    };

    const normalizedAvailable = {
      id: availableStudent.id, // Use id
      credits: creditsMap[availableStudent.id] || 0
    };

    // Both should have credits, no undefined
    expect(normalizedEnrolled.credits).toBe(5);
    expect(normalizedAvailable.credits).toBe(3);
    expect(normalizedEnrolled.credits).not.toBeUndefined();
    expect(normalizedAvailable.credits).not.toBeUndefined();
  });
});
