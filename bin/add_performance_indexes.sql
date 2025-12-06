-- Performance Index Migration for Logs Table
-- Run this on existing databases to improve query performance
-- This addresses gradual CPU increase over hours caused by full table scans

-- PostgreSQL Version (use CONCURRENTLY to avoid blocking writes)
-- Note: CONCURRENTLY cannot run inside a transaction block
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_type_created ON logs (type, created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_created ON logs (user_id, created_at);

-- Analyze the table to update statistics for the query planner
ANALYZE logs;

-- Check index usage after a few hours:
-- SELECT indexrelname, idx_scan, idx_tup_read, idx_tup_fetch 
-- FROM pg_stat_user_indexes 
-- WHERE schemaname = 'public' AND tablename = 'logs'
-- ORDER BY idx_scan DESC;
