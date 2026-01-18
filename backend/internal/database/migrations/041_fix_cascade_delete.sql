-- Migration 041: Fix CASCADE DELETE to prevent accidental data loss
-- Problem: When a user or lesson is deleted, all related data cascades
-- Solution: Change CASCADE to RESTRICT for critical foreign keys

-- ============================================
-- LESSONS TABLE
-- ============================================

-- lessons.teacher_id: Prevent deleting teacher if they have lessons
ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_teacher_id_fkey;
ALTER TABLE lessons ADD CONSTRAINT lessons_teacher_id_fkey
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ============================================
-- BOOKINGS TABLE
-- ============================================

-- bookings.lesson_id: Prevent deleting lesson if it has bookings
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_lesson_id_fkey;
ALTER TABLE bookings ADD CONSTRAINT bookings_lesson_id_fkey
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE RESTRICT;

-- bookings.student_id: Prevent deleting student if they have bookings
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_student_id_fkey;
ALTER TABLE bookings ADD CONSTRAINT bookings_student_id_fkey
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ============================================
-- CREDIT TRANSACTIONS (Audit Trail - never delete)
-- ============================================

ALTER TABLE credit_transactions DROP CONSTRAINT IF EXISTS credit_transactions_user_id_fkey;
ALTER TABLE credit_transactions ADD CONSTRAINT credit_transactions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ============================================
-- PAYMENTS (Financial Records - never delete)
-- ============================================

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_user_id_fkey;
ALTER TABLE payments ADD CONSTRAINT payments_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ============================================
-- CREDITS (Balance - soft delete only)
-- ============================================

ALTER TABLE credits DROP CONSTRAINT IF EXISTS credits_user_id_fkey;
ALTER TABLE credits ADD CONSTRAINT credits_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ============================================
-- CHAT ROOMS (Preserve chat history)
-- ============================================

ALTER TABLE chat_rooms DROP CONSTRAINT IF EXISTS chat_rooms_teacher_id_fkey;
ALTER TABLE chat_rooms ADD CONSTRAINT chat_rooms_teacher_id_fkey
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE chat_rooms DROP CONSTRAINT IF EXISTS chat_rooms_student_id_fkey;
ALTER TABLE chat_rooms ADD CONSTRAINT chat_rooms_student_id_fkey
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================
-- MESSAGES (Preserve message history)
-- ============================================

ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_sender_id_fkey;
ALTER TABLE messages ADD CONSTRAINT messages_sender_id_fkey
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================
-- SWAPS (Preserve swap history)
-- ============================================

ALTER TABLE swaps DROP CONSTRAINT IF EXISTS swaps_student_id_fkey;
ALTER TABLE swaps ADD CONSTRAINT swaps_student_id_fkey
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ============================================
-- CANCELLED BOOKINGS (Audit history)
-- ============================================

ALTER TABLE cancelled_bookings DROP CONSTRAINT IF EXISTS cancelled_bookings_student_id_fkey;
ALTER TABLE cancelled_bookings ADD CONSTRAINT cancelled_bookings_student_id_fkey
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE cancelled_bookings DROP CONSTRAINT IF EXISTS cancelled_bookings_lesson_id_fkey;
ALTER TABLE cancelled_bookings ADD CONSTRAINT cancelled_bookings_lesson_id_fkey
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE SET NULL;

-- ============================================
-- LESSON HOMEWORK (Preserve if lesson exists)
-- ============================================

ALTER TABLE lesson_homework DROP CONSTRAINT IF EXISTS lesson_homework_lesson_id_fkey;
ALTER TABLE lesson_homework ADD CONSTRAINT lesson_homework_lesson_id_fkey
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE RESTRICT;

ALTER TABLE lesson_homework DROP CONSTRAINT IF EXISTS lesson_homework_created_by_fkey;
ALTER TABLE lesson_homework ADD CONSTRAINT lesson_homework_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================
-- LESSON BROADCASTS (Preserve if lesson exists)
-- ============================================

ALTER TABLE lesson_broadcasts DROP CONSTRAINT IF EXISTS lesson_broadcasts_lesson_id_fkey;
ALTER TABLE lesson_broadcasts ADD CONSTRAINT lesson_broadcasts_lesson_id_fkey
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE RESTRICT;

ALTER TABLE lesson_broadcasts DROP CONSTRAINT IF EXISTS lesson_broadcasts_sender_id_fkey;
ALTER TABLE lesson_broadcasts ADD CONSTRAINT lesson_broadcasts_sender_id_fkey
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================
-- Add index for soft delete queries
-- ============================================

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lessons_deleted_at ON lessons(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status) WHERE status = 'active';

COMMENT ON CONSTRAINT lessons_teacher_id_fkey ON lessons IS 'RESTRICT: Must soft-delete lessons before deleting teacher';
COMMENT ON CONSTRAINT bookings_lesson_id_fkey ON bookings IS 'RESTRICT: Must cancel bookings before deleting lesson';
COMMENT ON CONSTRAINT bookings_student_id_fkey ON bookings IS 'RESTRICT: Must cancel bookings before deleting student';
