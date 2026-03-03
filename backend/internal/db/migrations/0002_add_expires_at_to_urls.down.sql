ALTER TABLE urls DROP INDEX ix_urls_expires_at;
ALTER TABLE urls DROP COLUMN expires_at;