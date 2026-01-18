/**
 * Utility functions for detecting modifications in lesson edit form
 * Used by bulk edit "Apply to all subsequent" feature
 */

/**
 * Detect what type of modification was made to a lesson
 * @param {Object} originalLesson - Original lesson data
 * @param {Object} editedLesson - Modified lesson data
 * @param {Array} originalStudents - Original list of students
 * @param {Array} editedStudents - Modified list of students
 * @returns {string|null} Modification type or null if no change detected
 */
export const detectModificationType = (
  originalLesson,
  editedLesson,
  originalStudents = [],
  editedStudents = []
) => {
  // Priority order: students > teacher > time > capacity
  // Only return one modification type at a time

  // Check for student additions
  if (hasStudentAdditions(originalStudents, editedStudents)) {
    return 'add_student';
  }

  // Check for student removals
  if (hasStudentRemovals(originalStudents, editedStudents)) {
    return 'remove_student';
  }

  // Check for teacher change
  if (hasTeacherChange(originalLesson, editedLesson)) {
    return 'change_teacher';
  }

  // Check for time change
  if (hasTimeChange(originalLesson, editedLesson)) {
    return 'change_time';
  }

  // Check for capacity change
  if (hasCapacityChange(originalLesson, editedLesson)) {
    return 'change_capacity';
  }

  return null;
};

/**
 * Check if students were added
 */
export const hasStudentAdditions = (originalStudents, editedStudents) => {
  const originalIds = new Set(
    originalStudents.map((s) => s.student_id || s.id)
  );
  const editedIds = new Set(editedStudents.map((s) => s.student_id || s.id));

  // Find students in edited but not in original
  for (const id of editedIds) {
    if (!originalIds.has(id)) {
      return true;
    }
  }

  return false;
};

/**
 * Check if students were removed
 */
export const hasStudentRemovals = (originalStudents, editedStudents) => {
  const originalIds = new Set(
    originalStudents.map((s) => s.student_id || s.id)
  );
  const editedIds = new Set(editedStudents.map((s) => s.student_id || s.id));

  // Find students in original but not in edited
  for (const id of originalIds) {
    if (!editedIds.has(id)) {
      return true;
    }
  }

  return false;
};

/**
 * Check if teacher changed
 */
export const hasTeacherChange = (originalLesson, editedLesson) => {
  return (
    originalLesson.teacher_id !== editedLesson.teacher_id &&
    editedLesson.teacher_id !== undefined
  );
};

/**
 * Check if time changed
 */
export const hasTimeChange = (originalLesson, editedLesson) => {
  // Compare start_time (end_time auto-calculated)
  if (!editedLesson.start_time) return false;

  const originalTime = new Date(originalLesson.start_time).getTime();
  const editedTime = new Date(editedLesson.start_time).getTime();

  return originalTime !== editedTime;
};

/**
 * Check if capacity changed
 */
export const hasCapacityChange = (originalLesson, editedLesson) => {
  return (
    originalLesson.max_students !== editedLesson.max_students &&
    editedLesson.max_students !== undefined
  );
};

/**
 * Get added students
 */
export const getAddedStudents = (originalStudents, editedStudents) => {
  const originalIds = new Set(
    originalStudents.map((s) => s.student_id || s.id)
  );

  return editedStudents.filter(
    (s) => !originalIds.has(s.student_id || s.id)
  );
};

/**
 * Get removed students
 */
export const getRemovedStudents = (originalStudents, editedStudents) => {
  const editedIds = new Set(editedStudents.map((s) => s.student_id || s.id));

  return originalStudents.filter((s) => !editedIds.has(s.student_id || s.id));
};

/**
 * Get modification details for confirmation dialog
 * @param {string} modificationType - Type of modification
 * @param {Object} originalLesson - Original lesson
 * @param {Object} editedLesson - Edited lesson
 * @param {Array} originalStudents - Original students
 * @param {Array} editedStudents - Edited students
 * @param {Array} availableTeachers - List of teachers for lookup
 * @returns {Object} Modification details
 */
export const getModificationDetails = (
  modificationType,
  originalLesson,
  editedLesson,
  originalStudents,
  editedStudents,
  availableTeachers = []
) => {
  const details = {};

  switch (modificationType) {
    case 'add_student': {
      const added = getAddedStudents(originalStudents, editedStudents);
      if (added.length > 0) {
        const student = added[0]; // Only support one at a time for bulk edit
        details.studentId = student.student_id || student.id;
        details.studentName = student.student_name || student.full_name;
      }
      break;
    }

    case 'remove_student': {
      const removed = getRemovedStudents(originalStudents, editedStudents);
      if (removed.length > 0) {
        const student = removed[0]; // Only support one at a time
        details.studentId = student.student_id || student.id;
        details.studentName = student.student_name || student.full_name;
      }
      break;
    }

    case 'change_teacher': {
      details.teacherId = editedLesson.teacher_id;
      const teacher = availableTeachers.find((t) => t.id === editedLesson.teacher_id);
      details.teacherName = teacher ? teacher.full_name : 'Unknown Teacher';
      break;
    }

    case 'change_time': {
      details.newStartTime = editedLesson.start_time;
      break;
    }

    case 'change_capacity': {
      details.newMaxStudents = editedLesson.max_students;
      break;
    }

    default:
      break;
  }

  return details;
};

export default {
  detectModificationType,
  hasStudentAdditions,
  hasStudentRemovals,
  hasTeacherChange,
  hasTimeChange,
  hasCapacityChange,
  getAddedStudents,
  getRemovedStudents,
  getModificationDetails,
};
