-- 038_subjects_table.sql
-- Purpose: Create subjects table for managing available subjects in the platform

BEGIN;

-- Subjects table: Defines available subjects that teachers can teach
CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Teacher subjects: Links teachers with the subjects they teach
-- This allows many-to-many relationship between teachers and subjects
CREATE TABLE teacher_subjects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(teacher_id, subject_id)
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Subjects indexes
CREATE INDEX idx_subjects_name ON subjects(name);
CREATE INDEX idx_subjects_active ON subjects(id) WHERE deleted_at IS NULL;

-- Teacher subjects indexes
CREATE INDEX idx_teacher_subjects_teacher_id ON teacher_subjects(teacher_id);
CREATE INDEX idx_teacher_subjects_subject_id ON teacher_subjects(subject_id);
CREATE INDEX idx_teacher_subjects_composite ON teacher_subjects(teacher_id, subject_id);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp for subjects
CREATE TRIGGER update_subjects_updated_at
    BEFORE UPDATE ON subjects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE subjects IS 'Available subjects in the platform (Math, English, Russian, etc.)';
COMMENT ON TABLE teacher_subjects IS 'Mapping of teachers to subjects they teach';

COMMENT ON COLUMN subjects.id IS 'Unique subject identifier';
COMMENT ON COLUMN subjects.name IS 'Subject name (e.g., "Mathematics", "English", "Russian")';
COMMENT ON COLUMN subjects.description IS 'Optional description of the subject';
COMMENT ON COLUMN subjects.created_at IS 'Timestamp when subject was created';
COMMENT ON COLUMN subjects.updated_at IS 'Timestamp when subject was last updated';
COMMENT ON COLUMN subjects.deleted_at IS 'Soft delete timestamp (NULL = active)';

COMMENT ON COLUMN teacher_subjects.teacher_id IS 'Teacher who teaches this subject';
COMMENT ON COLUMN teacher_subjects.subject_id IS 'Subject being taught';
COMMENT ON COLUMN teacher_subjects.assigned_at IS 'Timestamp when subject was assigned to teacher';

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration, run the following commands:
/*
BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS update_subjects_updated_at ON subjects;

-- Drop indexes
DROP INDEX IF EXISTS idx_subjects_name;
DROP INDEX IF EXISTS idx_subjects_active;
DROP INDEX IF EXISTS idx_teacher_subjects_teacher_id;
DROP INDEX IF EXISTS idx_teacher_subjects_subject_id;
DROP INDEX IF EXISTS idx_teacher_subjects_composite;

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS teacher_subjects CASCADE;
DROP TABLE IF EXISTS subjects CASCADE;

COMMIT;
*/
