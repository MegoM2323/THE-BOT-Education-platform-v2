-- 001_init_schema.sql
-- Initial database schema for Tutoring Platform

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('student', 'teacher', 'admin')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Credits table
CREATE TABLE credits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Credit transactions table
CREATE TABLE credit_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    operation_type VARCHAR(50) NOT NULL CHECK (operation_type IN ('add', 'deduct', 'refund')),
    reason TEXT,
    performed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    booking_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Lessons table
CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_type VARCHAR(50) NOT NULL CHECK (lesson_type IN ('individual', 'group')),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    max_students INTEGER NOT NULL CHECK (max_students > 0),
    current_students INTEGER NOT NULL DEFAULT 0 CHECK (current_students >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT lesson_time_valid CHECK (end_time > start_time),
    CONSTRAINT lesson_capacity_valid CHECK (current_students <= max_students)
);

-- Bookings table
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('active', 'cancelled')) DEFAULT 'active',
    booked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Swaps table (history of lesson swaps)
CREATE TABLE swaps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    new_lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    old_booking_id UUID NOT NULL,
    new_booking_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Comments for tables
COMMENT ON TABLE users IS 'User accounts (students, teachers, admins)';
COMMENT ON TABLE credits IS 'Credit balance for each user';
COMMENT ON TABLE credit_transactions IS 'History of all credit operations';
COMMENT ON TABLE lessons IS 'Available lessons created by teachers';
COMMENT ON TABLE bookings IS 'Student bookings for lessons';
COMMENT ON TABLE swaps IS 'History of lesson swap operations';
COMMENT ON TABLE sessions IS 'Active user sessions for authentication';

-- Comments for important columns
COMMENT ON COLUMN lessons.current_students IS 'Current number of students booked (must not exceed max_students)';
COMMENT ON COLUMN credit_transactions.performed_by IS 'Admin who performed the operation (NULL for system operations)';
COMMENT ON COLUMN bookings.status IS 'Booking status: active or cancelled';
