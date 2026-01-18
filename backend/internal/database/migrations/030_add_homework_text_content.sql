-- 030_add_homework_text_content.sql
-- Add text_content column to lesson_homework table to support text descriptions

-- Add text_content column to lesson_homework table
ALTER TABLE lesson_homework
ADD COLUMN IF NOT EXISTS text_content TEXT;

-- Add comment for the column
COMMENT ON COLUMN lesson_homework.text_content IS 'Текстовое описание или инструкции по выполнению домашнего задания';
