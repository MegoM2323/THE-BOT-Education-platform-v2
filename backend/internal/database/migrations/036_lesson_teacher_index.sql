-- 036_lesson_teacher_index.sql
-- Composite index optimization for teacher schedule queries
-- Addresses B1: Missing index for teacher schedule queries

-- CURRENT QUERY PATTERN (GetTeacherSchedule):
-- SELECT l.id, l.teacher_id, l.start_time, l.end_time, l.max_students, l.current_students, ...
-- FROM lessons l
-- WHERE l.teacher_id = $1
--   AND l.start_time >= $2
--   AND l.start_time <= $3
--   AND l.deleted_at IS NULL
-- ORDER BY l.start_time ASC

-- INDEX STRATEGY:
-- Primary composite index: idx_lessons_teacher_time (teacher_id, start_time, end_time)
-- - Allows: index skip scan on teacher_id, range scan on start_time, index-only scan for end_time
-- - Deleted_at IS NULL filter: partial index to exclude soft-deleted rows
-- - Already exists in migration 002, optimized for B1 use case

-- VERIFICATION: Run EXPLAIN ANALYZE to confirm index usage:
-- EXPLAIN ANALYZE
-- SELECT l.id, l.teacher_id, l.start_time, l.end_time,
--        l.max_students, l.current_students, l.color, l.subject
-- FROM lessons l
-- WHERE l.teacher_id = 'uuid-value'
--   AND l.start_time >= '2024-01-01'
--   AND l.start_time <= '2024-12-31'
--   AND l.deleted_at IS NULL
-- ORDER BY l.start_time ASC;

-- EXPECTED PLAN:
-- Index Scan using idx_lessons_teacher_time on lessons l
--   Index Cond: (teacher_id = 'uuid-value') AND (start_time >= '2024-01-01') AND (start_time <= '2024-12-31')
--   Filter: (deleted_at IS NULL)

-- Index exists, no new index needed. This migration documents the optimization strategy.

-- ADD QUERY PERFORMANCE COMMENT
COMMENT ON INDEX idx_lessons_teacher_time IS
'Composite index for teacher schedule queries. Optimizes GetTeacherSchedule() queries that filter by
teacher_id and date range (start_time >= date AND start_time <= date). Index is partial
(WHERE deleted_at IS NULL) to reduce size and improve scan performance. Column order
(teacher_id, start_time, end_time) enables: (1) equality match on teacher_id, (2) range scan
on start_time for date filtering, (3) index-only scan for end_time. Related to finding B1.';

-- ADDITIONAL INDEX FOR COVERAGE (Optional performance improvement for COUNT queries)
-- If needed, create a covering index with INCLUDE for COUNT(*) queries
-- Note: PostgreSQL 11+ only
-- CREATE INDEX idx_lessons_teacher_time_count ON lessons(teacher_id, start_time)
--   INCLUDE (max_students, current_students)
--   WHERE deleted_at IS NULL;

-- Future monitoring: Track index bloat with:
-- SELECT schemaname, tablename, indexname,
--        pg_size_pretty(pg_relation_size(indexrelid)) as size,
--        idx_blks_read, idx_blks_hit
-- FROM pg_stat_user_indexes
-- WHERE indexname = 'idx_lessons_teacher_time';
