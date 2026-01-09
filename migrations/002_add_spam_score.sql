-- Migration 002: Add Spam Score
-- HeadlessForms v1.2.0
-- Created: 2026-01-10

-- Add spam_score column to submissions for detailed spam analysis
ALTER TABLE submissions ADD COLUMN spam_score REAL DEFAULT 0;
