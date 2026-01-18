/**
 * Example usage of TemplateLessonEditorModal component
 *
 * This file demonstrates how to integrate the TemplateLessonEditorModal
 * in your admin components for creating and editing template lessons.
 */

import { useState } from 'react';
import TemplateLessonEditorModal from './TemplateLessonEditorModal.jsx';
import Button from "../common/Button.jsx";
import { logger } from '../../utils/logger.js';

export const TemplateLessonEditorExample = () => {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const templateId = 'your-template-id-here'; // Replace with actual template ID

  // Example data for edit mode
  const existingLesson = {
    id: 'lesson-id-123',
    day_of_week: 1, // Monday
    start_time: '10:00',
    teacher_id: 'teacher-id-456',
    lesson_type: 'group',
    max_students: 4,
    assigned_students: ['student-1', 'student-2'], // Array of student IDs
  };

  // Handle successful save
  const handleSave = (savedLesson) => {
    logger.debug('Lesson saved:', savedLesson);
    // Refresh your template data here
    // For example: refetchTemplate();
  };

  // Handle errors
  const handleError = (error) => {
    logger.error('Error saving lesson:', error);
  };

  return (
    <div>
      <h2>Template Lesson Editor Examples</h2>

      {/* Example 1: Create new lesson */}
      <div style={{ marginBottom: '20px' }}>
        <h3>Create Mode</h3>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          Add New Template Lesson
        </Button>

        <TemplateLessonEditorModal
          isOpen={isCreateModalOpen}
          mode="create"
          templateId={templateId}
          onClose={() => setIsCreateModalOpen(false)}
          onSave={handleSave}
          onError={handleError}
        />
      </div>

      {/* Example 2: Edit existing lesson */}
      <div style={{ marginBottom: '20px' }}>
        <h3>Edit Mode</h3>
        <Button onClick={() => setIsEditModalOpen(true)}>
          Edit Existing Template Lesson
        </Button>

        <TemplateLessonEditorModal
          isOpen={isEditModalOpen}
          mode="edit"
          templateId={templateId}
          lessonId={existingLesson.id}
          prefilledData={{
            day_of_week: existingLesson.day_of_week,
            start_time: existingLesson.start_time,
            teacher_id: existingLesson.teacher_id,
            lesson_type: existingLesson.lesson_type,
            max_students: existingLesson.max_students,
            assigned_students: existingLesson.assigned_students,
          }}
          onClose={() => setIsEditModalOpen(false)}
          onSave={handleSave}
          onError={handleError}
        />
      </div>
    </div>
  );
};

/**
 * Integration example for template management page
 */
export const TemplateManagementIntegration = () => {
  const [modalState, setModalState] = useState({
    isOpen: false,
    mode: 'create',
    lessonData: null,
  });

  const templateId = 'template-123';

  // Open modal for creating new lesson
  const openCreateModal = () => {
    setModalState({
      isOpen: true,
      mode: 'create',
      lessonData: null,
    });
  };

  // Open modal for editing lesson
  const openEditModal = (lesson) => {
    setModalState({
      isOpen: true,
      mode: 'edit',
      lessonData: {
        id: lesson.id,
        day_of_week: lesson.day_of_week,
        start_time: lesson.start_time,
        teacher_id: lesson.teacher_id,
        lesson_type: lesson.lesson_type,
        max_students: lesson.max_students,
        assigned_students: lesson.student_ids || [],
      },
    });
  };

  // Close modal
  const closeModal = () => {
    setModalState({
      isOpen: false,
      mode: 'create',
      lessonData: null,
    });
  };

  // Handle save
  const handleSave = (savedLesson) => {
    logger.debug('Lesson saved:', savedLesson);
    closeModal();
    // Refresh template data
  };

  return (
    <div>
      <Button onClick={openCreateModal}>Add Template Lesson</Button>

      {/* Example lesson list with edit buttons */}
      <div className="lesson-list">
        {/* Replace with your actual lesson list */}
        <div onClick={() => openEditModal({
          id: 'lesson-1',
          day_of_week: 1,
          start_time: '10:00',
          teacher_id: 'teacher-1',
          lesson_type: 'group',
          max_students: 4,
          student_ids: ['student-1', 'student-2'],
        })}>
          Click to edit lesson
        </div>
      </div>

      {/* Single modal handles both create and edit */}
      <TemplateLessonEditorModal
        isOpen={modalState.isOpen}
        mode={modalState.mode}
        templateId={templateId}
        lessonId={modalState.lessonData?.id}
        prefilledData={modalState.lessonData}
        onClose={closeModal}
        onSave={handleSave}
      />
    </div>
  );
};

/**
 * Props Interface (for TypeScript reference)
 *
 * interface TemplateLessonEditorModalProps {
 *   isOpen: boolean;                    // Modal visibility
 *   mode: 'create' | 'edit';            // Determines behavior
 *   templateId: string;                 // Required for API calls
 *   lessonId?: string;                  // Required for edit mode
 *   prefilledData?: {                   // Pre-fill data for edit mode
 *     day_of_week: number;              // 0-6 (Sun-Sat), 1=Monday
 *     start_time: string;               // HH:MM format
 *     teacher_id: string;               // Teacher UUID
 *     lesson_type: 'individual' | 'group';
 *     max_students: number;             // 1 for individual, 4+ for group
 *     assigned_students: string[];      // Array of student UUIDs
 *   };
 *   onClose: () => void;                // Close handler
 *   onSave: (lesson: any) => void;      // Save success handler
 *   onError?: (error: Error) => void;   // Optional error handler
 * }
 */

export default TemplateLessonEditorExample;
