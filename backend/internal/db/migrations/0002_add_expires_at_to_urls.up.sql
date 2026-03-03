ALTER TABLE urls
ADD COLUMN expires_at timestamp(3) NULL DEFAULT NULL AFTER updated_at,
ADD KEY ix_urls_expires_at (expires_at);