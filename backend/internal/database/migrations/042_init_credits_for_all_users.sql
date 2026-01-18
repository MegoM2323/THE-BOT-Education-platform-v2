-- 042_init_credits_for_all_users.sql
-- Initialize credits for existing admins and teachers who don't have credit records
-- Ensures all users (admin, teacher, student) have credit accounts

INSERT INTO credits (id, user_id, balance, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    0,
    NOW(),
    NOW()
FROM users u
LEFT JOIN credits c ON u.id = c.user_id
WHERE c.id IS NULL
    AND u.deleted_at IS NULL
    AND u.role IN ('admin', 'teacher', 'student')
ON CONFLICT (user_id) DO NOTHING;

-- Verification: Check if all non-deleted users now have credits
-- SELECT COUNT(*) as total_users,
--        SUM(CASE WHEN c.id IS NOT NULL THEN 1 ELSE 0 END) as users_with_credits
-- FROM users u
-- LEFT JOIN credits c ON u.id = c.user_id
-- WHERE u.deleted_at IS NULL;
