-- Fix first_name and last_name columns to be NOT NULL with empty string default
-- This prevents NULL scan errors when reading user records

-- Update any existing NULL values to empty strings
UPDATE users SET first_name = '' WHERE first_name IS NULL;
UPDATE users SET last_name = '' WHERE last_name IS NULL;

-- Set default values and NOT NULL constraints
ALTER TABLE users ALTER COLUMN first_name SET DEFAULT '';
ALTER TABLE users ALTER COLUMN last_name SET DEFAULT '';
ALTER TABLE users ALTER COLUMN first_name SET NOT NULL;
ALTER TABLE users ALTER COLUMN last_name SET NOT NULL;

COMMENT ON COLUMN users.first_name IS 'User first name, cannot be NULL (empty string allowed)';
COMMENT ON COLUMN users.last_name IS 'User last name, cannot be NULL (empty string allowed)';
