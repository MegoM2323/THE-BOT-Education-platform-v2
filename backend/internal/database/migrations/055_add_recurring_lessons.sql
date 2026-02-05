-- +migrate Up
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT false;
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS recurring_group_id UUID;
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS recurring_end_date DATE;

CREATE INDEX idx_lessons_recurring_group_id ON lessons(recurring_group_id) WHERE recurring_group_id IS NOT NULL;
CREATE INDEX idx_lessons_is_recurring ON lessons(is_recurring) WHERE is_recurring = true;

-- +migrate Down
DROP INDEX IF EXISTS idx_lessons_is_recurring;
DROP INDEX IF EXISTS idx_lessons_recurring_group_id;
ALTER TABLE lessons DROP COLUMN IF EXISTS recurring_end_date;
ALTER TABLE lessons DROP COLUMN IF EXISTS recurring_group_id;
ALTER TABLE lessons DROP COLUMN IF EXISTS is_recurring;
