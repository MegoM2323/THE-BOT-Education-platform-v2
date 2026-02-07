-- 011_advanced_scheduling_schema.sql
-- Purpose: Add advanced scheduling features (templates, bulk edits, visibility)
-- Created: 2025-12-01
--
-- This migration adds:
-- 1. Lesson templates system (admin-created weekly schedule templates)
-- 2. Template application tracking (history of template usage)
-- 3. Lesson modifications audit trail (bulk edit operations)
-- 4. Enhanced lesson visibility (individual lesson privacy)
-- 5. Auto time calculation support (end_time = start_time + 2h)

BEGIN;

-- ============================================================================
-- TABLE DEFINITIONS
-- ============================================================================

-- Lesson templates: Admin-created weekly schedule templates
CREATE TABLE lesson_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
    -- Note: Admin role validation should be done in application layer
);

-- Template lessons: Individual lesson entries within templates
-- These are NOT actual lessons, but template definitions
-- day_of_week: 0 = Sunday, 1 = Monday, ..., 6 = Saturday (ISO standard)
CREATE TABLE template_lessons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES lesson_templates(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    lesson_type VARCHAR(50) NOT NULL CHECK (lesson_type IN ('individual', 'group')),
    max_students INTEGER NOT NULL DEFAULT 4 CHECK (max_students > 0),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT template_lessons_time_valid CHECK (end_time > start_time)
    -- Note: Teacher role validation should be done in application layer
);

-- Template lesson students: Pre-assigned students for template lessons
CREATE TABLE template_lesson_students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_lesson_id UUID NOT NULL REFERENCES template_lessons(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_lesson_id, student_id)
    -- Note: Student role validation should be done in application layer
);

-- Template applications: History of when templates were applied to calendar weeks
CREATE TABLE template_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES lesson_templates(id) ON DELETE RESTRICT,
    applied_by_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    week_start_date DATE NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'applied' CHECK (status IN ('applied', 'rolled_back')),
    rolled_back_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT template_applications_valid_status CHECK (status IN ('applied', 'rolled_back')),
    CONSTRAINT template_applications_rollback_check CHECK ((status = 'rolled_back' AND rolled_back_at IS NOT NULL) OR (status = 'applied' AND rolled_back_at IS NULL))
);

-- Lesson modifications: Audit trail for bulk edit operations
CREATE TABLE lesson_modifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    performed_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    modification_type VARCHAR(50) NOT NULL CHECK (modification_type IN ('add_student', 'remove_student', 'change_teacher', 'change_time', 'change_type')),
    source_lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    affected_lessons_count INTEGER NOT NULL DEFAULT 0,
    old_value TEXT,
    new_value TEXT,
    applied_from_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- ALTER EXISTING TABLES
-- ============================================================================

-- Add template tracking columns to lessons table
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS applied_from_template BOOLEAN DEFAULT false NOT NULL;
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS template_application_id UUID REFERENCES template_applications(id) ON DELETE SET NULL;

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Lesson templates indexes
CREATE INDEX idx_lesson_templates_admin_id ON lesson_templates(admin_id);
CREATE INDEX idx_lesson_templates_created_at ON lesson_templates(created_at DESC);
CREATE INDEX idx_lesson_templates_active ON lesson_templates(id) WHERE deleted_at IS NULL;

-- Template lessons indexes
CREATE INDEX idx_template_lessons_template_id ON template_lessons(template_id);
CREATE INDEX idx_template_lessons_teacher_id ON template_lessons(teacher_id);
CREATE INDEX idx_template_lessons_day_of_week ON template_lessons(day_of_week);
CREATE INDEX idx_template_lessons_start_time ON template_lessons(start_time);
CREATE INDEX idx_template_lessons_type ON template_lessons(lesson_type);
CREATE INDEX idx_template_lessons_composite ON template_lessons(template_id, day_of_week, start_time);

-- Template lesson students indexes
CREATE INDEX idx_template_lesson_students_template_lesson_id ON template_lesson_students(template_lesson_id);
CREATE INDEX idx_template_lesson_students_student_id ON template_lesson_students(student_id);

-- Template applications indexes
CREATE INDEX idx_template_applications_template_id ON template_applications(template_id);
CREATE INDEX idx_template_applications_applied_by_id ON template_applications(applied_by_id);
CREATE INDEX idx_template_applications_week_start ON template_applications(week_start_date);
CREATE INDEX idx_template_applications_applied_at ON template_applications(applied_at DESC);
CREATE INDEX idx_template_applications_active ON template_applications(id) WHERE status = 'applied';

-- Lesson modifications indexes
CREATE INDEX idx_lesson_modifications_performed_by ON lesson_modifications(performed_by);
CREATE INDEX idx_lesson_modifications_source_lesson_id ON lesson_modifications(source_lesson_id);
CREATE INDEX idx_lesson_modifications_type ON lesson_modifications(modification_type);
CREATE INDEX idx_lesson_modifications_created_at ON lesson_modifications(created_at DESC);
CREATE INDEX idx_lesson_modifications_applied_from_date ON lesson_modifications(applied_from_date);

-- Lessons table new column indexes
CREATE INDEX idx_lessons_template_application_id ON lessons(template_application_id);
CREATE INDEX idx_lessons_applied_from_template ON lessons(applied_from_template) WHERE applied_from_template = true;

-- ============================================================================
-- FUNCTIONS AND TRIGGERS
-- ============================================================================

-- Function: Auto-update updated_at timestamp for lesson_templates
CREATE TRIGGER update_lesson_templates_updated_at
    BEFORE UPDATE ON lesson_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function: Auto-update updated_at timestamp for template_lessons
CREATE TRIGGER update_template_lessons_updated_at
    BEFORE UPDATE ON template_lessons
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function: Enforce individual lessons have max_students = 1
CREATE OR REPLACE FUNCTION enforce_individual_lesson_constraint()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.lesson_type = 'individual' AND NEW.max_students != 1 THEN
        RAISE EXCEPTION 'Individual lessons must have max_students = 1, got %', NEW.max_students;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_individual_constraint_template_lessons
    BEFORE INSERT OR UPDATE ON template_lessons
    FOR EACH ROW
    WHEN (NEW.lesson_type = 'individual')
    EXECUTE FUNCTION enforce_individual_lesson_constraint();

-- Function: Validate student role before adding to template
CREATE OR REPLACE FUNCTION validate_student_role()
RETURNS TRIGGER AS $$
DECLARE
    user_role VARCHAR(50);
BEGIN
    SELECT role INTO user_role FROM users WHERE id = NEW.student_id;
    IF user_role != 'student' THEN
        RAISE EXCEPTION 'User % is not a student (role: %)', NEW.student_id, user_role;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_student_role_before_insert
    BEFORE INSERT ON template_lesson_students
    FOR EACH ROW
    EXECUTE FUNCTION validate_student_role();

-- Function: Validate teacher role before adding to template lesson
CREATE OR REPLACE FUNCTION validate_teacher_role()
RETURNS TRIGGER AS $$
DECLARE
    user_role VARCHAR(50);
BEGIN
    SELECT role INTO user_role FROM users WHERE id = NEW.teacher_id;
    IF user_role != 'teacher' THEN
        RAISE EXCEPTION 'User % is not a teacher (role: %)', NEW.teacher_id, user_role;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_teacher_role_before_insert
    BEFORE INSERT OR UPDATE ON template_lessons
    FOR EACH ROW
    EXECUTE FUNCTION validate_teacher_role();

-- Function: Validate template lesson capacity doesn't exceed student count
CREATE OR REPLACE FUNCTION validate_template_lesson_capacity()
RETURNS TRIGGER AS $$
DECLARE
    student_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO student_count
    FROM template_lesson_students
    WHERE template_lesson_id = NEW.id;

    IF NEW.max_students < student_count THEN
        RAISE EXCEPTION 'Cannot set max_students to % when % students are already assigned',
            NEW.max_students, student_count;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_template_lesson_capacity_on_update
    BEFORE UPDATE ON template_lessons
    FOR EACH ROW
    WHEN (NEW.max_students IS DISTINCT FROM OLD.max_students)
    EXECUTE FUNCTION validate_template_lesson_capacity();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE lesson_templates IS 'Admin-created weekly schedule templates containing lesson definitions';
COMMENT ON TABLE template_lessons IS 'Individual lesson entries within templates (not actual lessons, just definitions)';
COMMENT ON TABLE template_lesson_students IS 'Pre-assigned students for template lessons';
COMMENT ON TABLE template_applications IS 'History of when templates were applied to calendar weeks with credit tracking';
COMMENT ON TABLE lesson_modifications IS 'Audit trail for bulk edit operations (apply to all subsequent lessons)';

COMMENT ON COLUMN lesson_templates.admin_id IS 'Admin who created this template';
COMMENT ON COLUMN lesson_templates.name IS 'Human-readable template name (e.g., "Week A Standard Schedule")';
COMMENT ON COLUMN lesson_templates.description IS 'Optional description of template purpose';
COMMENT ON COLUMN lesson_templates.deleted_at IS 'Soft delete timestamp (NULL = active)';

COMMENT ON COLUMN template_lessons.template_id IS 'Parent template this lesson belongs to';
COMMENT ON COLUMN template_lessons.teacher_id IS 'Teacher assigned to this lesson in the template';
COMMENT ON COLUMN template_lessons.day_of_week IS 'Day of week (0=Sunday, 1=Monday, ..., 6=Saturday) per ISO 8601';
COMMENT ON COLUMN template_lessons.start_time IS 'Start time (TIME type, no date component)';
COMMENT ON COLUMN template_lessons.end_time IS 'End time (should be start_time + 2 hours by default)';
COMMENT ON COLUMN template_lessons.lesson_type IS 'Individual or group lesson type';
COMMENT ON COLUMN template_lessons.max_students IS 'Maximum students for this lesson (1 for individual, typically 4 for group)';

COMMENT ON COLUMN template_lesson_students.template_lesson_id IS 'Template lesson this student is pre-assigned to';
COMMENT ON COLUMN template_lesson_students.student_id IS 'Student pre-assigned to this template lesson';

COMMENT ON COLUMN template_applications.template_id IS 'Template that was applied';
COMMENT ON COLUMN template_applications.applied_by_id IS 'Admin who applied the template';
COMMENT ON COLUMN template_applications.week_start_date IS 'Monday of the week this template was applied to (ISO 8601)';
COMMENT ON COLUMN template_applications.applied_at IS 'Timestamp when template was applied';
COMMENT ON COLUMN template_applications.status IS 'Application status: applied or rolled_back';
COMMENT ON COLUMN template_applications.rolled_back_at IS 'Timestamp when this application was rolled back (NULL = still active)';

COMMENT ON COLUMN lesson_modifications.performed_by IS 'User who performed the bulk modification';
COMMENT ON COLUMN lesson_modifications.modification_type IS 'Type of modification (add_student, remove_student, change_teacher, etc.)';
COMMENT ON COLUMN lesson_modifications.source_lesson_id IS 'Original lesson that triggered the bulk edit';
COMMENT ON COLUMN lesson_modifications.affected_lessons_count IS 'Number of lessons modified in this operation';
COMMENT ON COLUMN lesson_modifications.old_value IS 'Previous value (JSON or text representation)';
COMMENT ON COLUMN lesson_modifications.new_value IS 'New value after modification';
COMMENT ON COLUMN lesson_modifications.applied_from_date IS 'Date from which modifications were applied (all subsequent lessons)';
COMMENT ON COLUMN lesson_modifications.notes IS 'Optional notes about the modification';

COMMENT ON COLUMN lessons.applied_from_template IS 'True if this lesson was created from a template application';
COMMENT ON COLUMN lessons.template_application_id IS 'Reference to template application that created this lesson (NULL if manually created)';

COMMENT ON FUNCTION enforce_individual_lesson_constraint() IS 'Ensures individual lessons always have exactly 1 max_students';
COMMENT ON FUNCTION validate_student_role() IS 'Validates user is a student before adding to template lesson';
COMMENT ON FUNCTION validate_teacher_role() IS 'Validates user is a teacher before assigning to template lesson';
COMMENT ON FUNCTION validate_template_lesson_capacity() IS 'Prevents reducing max_students below current student count';

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration, run the following commands:
/*
BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS enforce_individual_constraint_template_lessons ON template_lessons;
DROP TRIGGER IF EXISTS validate_student_role_before_insert ON template_lesson_students;
DROP TRIGGER IF EXISTS validate_teacher_role_before_insert ON template_lessons;
DROP TRIGGER IF EXISTS validate_template_lesson_capacity_on_update ON template_lessons;
DROP TRIGGER IF EXISTS update_lesson_templates_updated_at ON lesson_templates;
DROP TRIGGER IF EXISTS update_template_lessons_updated_at ON template_lessons;

-- Drop functions
DROP FUNCTION IF EXISTS enforce_individual_lesson_constraint();
DROP FUNCTION IF EXISTS validate_student_role();
DROP FUNCTION IF EXISTS validate_teacher_role();
DROP FUNCTION IF EXISTS validate_template_lesson_capacity();

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS lesson_modifications CASCADE;
DROP TABLE IF EXISTS template_applications CASCADE;
DROP TABLE IF EXISTS template_lesson_students CASCADE;
DROP TABLE IF EXISTS template_lessons CASCADE;
DROP TABLE IF EXISTS lesson_templates CASCADE;

-- Remove added columns from lessons table
ALTER TABLE lessons DROP COLUMN IF EXISTS applied_from_template;
ALTER TABLE lessons DROP COLUMN IF EXISTS template_application_id;

COMMIT;
*/
