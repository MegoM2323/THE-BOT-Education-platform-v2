-- 002_add_indexes.sql
-- Performance indexes for the tutoring platform

-- Users indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Credits indexes
CREATE INDEX idx_credits_user_id ON credits(user_id);

-- Credit transactions indexes
CREATE INDEX idx_credit_transactions_user_id ON credit_transactions(user_id);
CREATE INDEX idx_credit_transactions_created_at ON credit_transactions(created_at DESC);
CREATE INDEX idx_credit_transactions_booking_id ON credit_transactions(booking_id) WHERE booking_id IS NOT NULL;
CREATE INDEX idx_credit_transactions_operation_type ON credit_transactions(operation_type);

-- Lessons indexes
CREATE INDEX idx_lessons_teacher_id ON lessons(teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_lessons_start_time ON lessons(start_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_lessons_lesson_type ON lessons(lesson_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_lessons_deleted_at ON lessons(deleted_at);
CREATE INDEX idx_lessons_availability ON lessons(start_time, current_students, max_students)
    WHERE deleted_at IS NULL AND current_students < max_students;

-- Bookings indexes
CREATE INDEX idx_bookings_student_id ON bookings(student_id);
CREATE INDEX idx_bookings_lesson_id ON bookings(lesson_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_student_active ON bookings(student_id, status) WHERE status = 'active';
CREATE INDEX idx_bookings_lesson_active ON bookings(lesson_id, status) WHERE status = 'active';

-- Swaps indexes
CREATE INDEX idx_swaps_student_id ON swaps(student_id);
CREATE INDEX idx_swaps_created_at ON swaps(created_at DESC);
CREATE INDEX idx_swaps_old_lesson_id ON swaps(old_lesson_id);
CREATE INDEX idx_swaps_new_lesson_id ON swaps(new_lesson_id);

-- Sessions indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
-- Note: Removed partial index with CURRENT_TIMESTAMP as it requires IMMUTABLE functions
-- Application will filter expired sessions in code instead of at database level

-- Composite indexes for common queries
CREATE INDEX idx_lessons_teacher_time ON lessons(teacher_id, start_time, end_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_student_lesson ON bookings(student_id, lesson_id);
