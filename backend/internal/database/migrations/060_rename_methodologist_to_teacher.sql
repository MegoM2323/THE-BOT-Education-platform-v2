-- Rename role methodologist to teacher
UPDATE users SET role = 'teacher' WHERE role = 'methodologist';

-- Update check constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
  CHECK (role IN ('student', 'admin', 'teacher'));
