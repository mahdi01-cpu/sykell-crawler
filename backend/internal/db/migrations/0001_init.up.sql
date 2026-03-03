CREATE TABLE IF NOT EXISTS urls (
  id bigint unsigned NOT NULL auto_increment,
  -- Core
  url varchar(2048) NOT NULL,
  url_hash binary(32) NOT NULL,
  status varchar(32) NOT NULL,
  -- Crawled Data
  html_version varchar(64) NOT NULL DEFAULT '',
  page_title varchar(1024) NOT NULL DEFAULT '',
  links_count int unsigned NOT NULL DEFAULT 0,
  internal_links_count int unsigned NOT NULL DEFAULT 0,
  external_links_count int unsigned NOT NULL DEFAULT 0,
  inaccessible_links_count int unsigned NOT NULL DEFAULT 0,
  has_login_form boolean NOT NULL DEFAULT false,
  h1_count int unsigned NOT NULL DEFAULT 0,
  h2_count int unsigned NOT NULL DEFAULT 0,
  h3_count int unsigned NOT NULL DEFAULT 0,
  h4_count int unsigned NOT NULL DEFAULT 0,
  h5_count int unsigned NOT NULL DEFAULT 0,
  h6_count int unsigned NOT NULL DEFAULT 0,
  -- Timestamps
  created_at timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY ux_urls_hash (url_hash),
  KEY ix_urls_status (status),
  KEY ix_urls_created_at (created_at),
  KEY ix_urls_internal_links (internal_links_count),
  KEY ix_urls_external_links (external_links_count),
  KEY ix_urls_inaccessible_links (inaccessible_links_count)
) engine = innodb DEFAULT charset = utf8mb4 COLLATE = utf8mb4_unicode_ci;
