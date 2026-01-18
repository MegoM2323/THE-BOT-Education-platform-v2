-- seed_templates.sql
-- Optional seed data for testing template functionality
-- Prerequisites: seed_test_data.sql must be run first (provides test users)
--
-- This seed creates:
-- 1. Sample template: "Week A - Standard Schedule" with 5 lessons
-- 2. Sample template: "Week B - Extended Schedule" with 7 lessons
-- 3. Pre-assigned students for individual lessons
--
-- To use: psql -U postgres -d tutoring_platform -f seed_templates.sql

BEGIN;

-- Get admin, teacher, and student IDs from seed data
-- Assuming seed_test_data.sql created:
-- - Admin: admin@test.com
-- - Teacher: teacher@test.com
-- - Students: student@test.com, student2@test.com

DO $$
DECLARE
    admin_user_id UUID;
    teacher_user_id UUID;
    student1_id UUID;
    student2_id UUID;
    template_a_id UUID;
    template_b_id UUID;
    tl_mon_morning UUID;
    tl_mon_afternoon UUID;
    tl_wed_morning UUID;
    tl_fri_morning UUID;
    tl_fri_afternoon UUID;
BEGIN
    -- Get user IDs
    SELECT id INTO admin_user_id FROM users WHERE email = 'admin@test.com' AND role = 'admin';
    SELECT id INTO teacher_user_id FROM users WHERE email = 'teacher@test.com' AND role = 'teacher';
    SELECT id INTO student1_id FROM users WHERE email = 'student@test.com' AND role = 'student';
    SELECT id INTO student2_id FROM users WHERE email = 'student2@test.com' AND role = 'student';

    IF admin_user_id IS NULL OR teacher_user_id IS NULL OR student1_id IS NULL THEN
        RAISE EXCEPTION 'Required test users not found. Please run seed_test_data.sql first.';
    END IF;

    -- ========================================================================
    -- TEMPLATE A: "Week A - Standard Schedule"
    -- ========================================================================

    template_a_id := uuid_generate_v4();

    INSERT INTO lesson_templates (id, admin_id, name, description)
    VALUES (
        template_a_id,
        admin_user_id,
        'Week A - Standard Schedule',
        'Standard weekly schedule with 5 lessons: 2 individual, 3 group lessons'
    );

    -- Monday 10:00-12:00 - Individual lesson (student1)
    tl_mon_morning := uuid_generate_v4();
    INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (
        tl_mon_morning,
        template_a_id,
        teacher_user_id,
        1, -- Monday
        '10:00:00',
        '12:00:00',
        'individual',
        1
    );

    INSERT INTO template_lesson_students (template_lesson_id, student_id)
    VALUES (tl_mon_morning, student1_id);

    -- Monday 14:00-16:00 - Group lesson (4 max)
    tl_mon_afternoon := uuid_generate_v4();
    INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (
        tl_mon_afternoon,
        template_a_id,
        teacher_user_id,
        1, -- Monday
        '14:00:00',
        '16:00:00',
        'group',
        4
    );

    -- Assign student1 to group lesson
    INSERT INTO template_lesson_students (template_lesson_id, student_id)
    VALUES (tl_mon_afternoon, student1_id);

    -- Wednesday 10:00-12:00 - Group lesson
    tl_wed_morning := uuid_generate_v4();
    INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (
        tl_wed_morning,
        template_a_id,
        teacher_user_id,
        3, -- Wednesday
        '10:00:00',
        '12:00:00',
        'group',
        4
    );

    INSERT INTO template_lesson_students (template_lesson_id, student_id)
    VALUES (tl_wed_morning, student1_id);

    IF student2_id IS NOT NULL THEN
        INSERT INTO template_lesson_students (template_lesson_id, student_id)
        VALUES (tl_wed_morning, student2_id);
    END IF;

    -- Friday 10:00-12:00 - Individual lesson (student2 if exists)
    tl_fri_morning := uuid_generate_v4();
    INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (
        tl_fri_morning,
        template_a_id,
        teacher_user_id,
        5, -- Friday
        '10:00:00',
        '12:00:00',
        'individual',
        1
    );

    IF student2_id IS NOT NULL THEN
        INSERT INTO template_lesson_students (template_lesson_id, student_id)
        VALUES (tl_fri_morning, student2_id);
    END IF;

    -- Friday 14:00-16:00 - Group lesson
    tl_fri_afternoon := uuid_generate_v4();
    INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (
        tl_fri_afternoon,
        template_a_id,
        teacher_user_id,
        5, -- Friday
        '14:00:00',
        '16:00:00',
        'group',
        4
    );

    INSERT INTO template_lesson_students (template_lesson_id, student_id)
    VALUES (tl_fri_afternoon, student1_id);

    -- ========================================================================
    -- TEMPLATE B: "Week B - Extended Schedule"
    -- ========================================================================

    template_b_id := uuid_generate_v4();

    INSERT INTO lesson_templates (id, admin_id, name, description)
    VALUES (
        template_b_id,
        admin_user_id,
        'Week B - Extended Schedule',
        'Extended schedule with 7 lessons covering all weekdays'
    );

    -- Monday 09:00-11:00 - Group
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 1, '09:00:00', '11:00:00', 'group', 4);

    -- Tuesday 10:00-12:00 - Group
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 2, '10:00:00', '12:00:00', 'group', 4);

    -- Wednesday 10:00-12:00 - Individual
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 3, '10:00:00', '12:00:00', 'individual', 1);

    -- Wednesday 14:00-16:00 - Group
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 3, '14:00:00', '16:00:00', 'group', 4);

    -- Thursday 10:00-12:00 - Group
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 4, '10:00:00', '12:00:00', 'group', 4);

    -- Friday 10:00-12:00 - Individual
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 5, '10:00:00', '12:00:00', 'individual', 1);

    -- Friday 15:00-17:00 - Group
    INSERT INTO template_lessons (template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students)
    VALUES (template_b_id, teacher_user_id, 5, '15:00:00', '17:00:00', 'group', 4);

    RAISE NOTICE 'Successfully created 2 templates with lesson entries';
    RAISE NOTICE 'Template A ID: %', template_a_id;
    RAISE NOTICE 'Template B ID: %', template_b_id;

END $$;

COMMIT;

-- Verification queries (optional - comment out or run separately)
/*
SELECT
    lt.name,
    lt.description,
    COUNT(tl.id) as lesson_count,
    COUNT(tls.id) as student_assignments
FROM lesson_templates lt
LEFT JOIN template_lessons tl ON lt.id = tl.template_id
LEFT JOIN template_lesson_students tls ON tl.id = tls.template_lesson_id
WHERE lt.deleted_at IS NULL
GROUP BY lt.id, lt.name, lt.description;
*/
