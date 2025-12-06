-- Performance Index Migration for Logs Table (MySQL Version)
-- Run this on existing MySQL databases to improve query performance

-- MySQL Version
CREATE INDEX idx_type_created ON logs (type, created_at);
CREATE INDEX idx_user_created ON logs (user_id, created_at);

-- Analyze the table to update statistics
ANALYZE TABLE logs;

-- Check index usage after a few hours:
-- SELECT TABLE_NAME, INDEX_NAME, CARDINALITY
-- FROM information_schema.STATISTICS
-- WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs'
-- ORDER BY CARDINALITY DESC;
