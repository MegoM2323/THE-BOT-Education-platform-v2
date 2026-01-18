-- Migration 021: Add cancelled_bookings table for re-booking prevention
-- Date: 2025-12-08
-- Purpose: Prevent students from rebooking cancelled lessons

-- Create cancelled_bookings table
CREATE TABLE IF NOT EXISTS cancelled_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    cancelled_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_student_lesson UNIQUE(student_id, lesson_id)
);

-- Create index for fast lookups (primary use case)
CREATE INDEX IF NOT EXISTS idx_cancelled_bookings_student_lesson 
ON cancelled_bookings(student_id, lesson_id);

-- Create index for listing by student (admin feature)
CREATE INDEX IF NOT EXISTS idx_cancelled_bookings_student 
ON cancelled_bookings(student_id, cancelled_at DESC);

-- Add comments for documentation
COMMENT ON TABLE cancelled_bookings IS 
  'Stores records of cancelled bookings to prevent students from rebooking the same lesson';

COMMENT ON COLUMN cancelled_bookings.booking_id IS 
  'UUID of the original booking that was cancelled';

COMMENT ON COLUMN cancelled_bookings.student_id IS 
  'UUID of the student who cancelled the booking';

COMMENT ON COLUMN cancelled_bookings.lesson_id IS 
  'UUID of the lesson that was cancelled';

COMMENT ON COLUMN cancelled_bookings.cancelled_at IS 
  'Timestamp when the booking was cancelled';

-- Display success message
DO $$
BEGIN
    RAISE NOTICE 'Migration 021: cancelled_bookings table created successfully';
END $$;
