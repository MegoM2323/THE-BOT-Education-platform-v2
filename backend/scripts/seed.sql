-- Seed data for Tutoring Platform
-- This file creates test users, lessons, and initial data

-- Clean existing data (optional, comment out if not needed)
-- TRUNCATE users, credits, lessons, bookings, swaps, sessions CASCADE;

-- Insert test users
-- Password for all users: "password123"
-- Hash: $2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy (bcrypt cost 10)
INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@example.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Admin User', 'admin', NOW(), NOW()),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'teacher1@example.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'John Teacher', 'teacher', NOW(), NOW()),
    ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'teacher2@example.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Jane Teacher', 'teacher', NOW(), NOW()),
    ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'student1@example.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Alice Student', 'student', NOW(), NOW()),
    ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'student2@example.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Bob Student', 'student', NOW(), NOW()),
    ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'student3@example.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Charlie Student', 'student', NOW(), NOW())
ON CONFLICT (email) WHERE deleted_at IS NULL DO NOTHING;

-- Add credits to students (trigger should auto-create, but we can set initial balance)
UPDATE credits SET balance = 10 WHERE user_id = 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44';
UPDATE credits SET balance = 5 WHERE user_id = 'e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55';
UPDATE credits SET balance = 3 WHERE user_id = 'f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66';

-- Insert sample lessons (starting tomorrow and later)
-- max_students определяет вместимость: 1 = индивидуальное, >1 = групповое
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, created_at, updated_at) VALUES
    -- Teacher 1 lessons
    ('10000000-0000-0000-0000-000000000001', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        NOW() + INTERVAL '2 days', NOW() + INTERVAL '2 days' + INTERVAL '1 hour', 1, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000002', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        NOW() + INTERVAL '3 days', NOW() + INTERVAL '3 days' + INTERVAL '1.5 hours', 5, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000003', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        NOW() + INTERVAL '4 days', NOW() + INTERVAL '4 days' + INTERVAL '1 hour', 1, NOW(), NOW()),

    -- Teacher 2 lessons
    ('20000000-0000-0000-0000-000000000001', 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
        NOW() + INTERVAL '2 days', NOW() + INTERVAL '2 days' + INTERVAL '2 hours', 10, NOW(), NOW()),
    ('20000000-0000-0000-0000-000000000002', 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
        NOW() + INTERVAL '5 days', NOW() + INTERVAL '5 days' + INTERVAL '1 hour', 1, NOW(), NOW()),
    ('20000000-0000-0000-0000-000000000003', 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
        NOW() + INTERVAL '6 days', NOW() + INTERVAL '6 days' + INTERVAL '1.5 hours', 8, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Record initial credit transactions
INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, created_at) VALUES
    (gen_random_uuid(), 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 10, 'add', 'Initial credits', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NOW()),
    (gen_random_uuid(), 'e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 5, 'add', 'Initial credits', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NOW()),
    (gen_random_uuid(), 'f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 3, 'add', 'Initial credits', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NOW());

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✓ Seed data inserted successfully!';
    RAISE NOTICE '';
    RAISE NOTICE 'Test accounts:';
    RAISE NOTICE '  Admin:    admin@example.com / password123';
    RAISE NOTICE '  Teacher1: teacher1@example.com / password123';
    RAISE NOTICE '  Teacher2: teacher2@example.com / password123';
    RAISE NOTICE '  Student1: student1@example.com / password123 (10 credits)';
    RAISE NOTICE '  Student2: student2@example.com / password123 (5 credits)';
    RAISE NOTICE '  Student3: student3@example.com / password123 (3 credits)';
END $$;
