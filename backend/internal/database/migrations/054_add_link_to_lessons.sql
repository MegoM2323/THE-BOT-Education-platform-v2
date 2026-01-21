-- Migration: Add link field to lessons table
-- Description: Adds optional link field for storing meeting URLs or resource links

-- +migrate Up
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS link TEXT;

-- +migrate Down
ALTER TABLE lessons DROP COLUMN IF EXISTS link;

