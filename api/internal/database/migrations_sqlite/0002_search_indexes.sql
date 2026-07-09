-- Public article search, archive filters, category/tag filters, and alternate sort modes.
CREATE INDEX IF NOT EXISTS idx_posts_status_view_count ON posts (status, view_count DESC, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_status_comment_count ON posts (status, comment_count DESC, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_status_like_count ON posts (status, like_count DESC, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_status_category_published_at ON posts (status, category_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_slug_status ON posts (slug, status);
CREATE INDEX IF NOT EXISTS idx_posts_status_author_name ON posts (status, lower(author_name), published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_author_slug_lower ON posts (lower(replace(author_name, ' ', '-')));
CREATE INDEX IF NOT EXISTS idx_categories_slug_lower ON categories (lower(slug));
CREATE INDEX IF NOT EXISTS idx_categories_name_lower ON categories (lower(name));
CREATE INDEX IF NOT EXISTS idx_tags_slug_lower ON tags (lower(slug));
CREATE INDEX IF NOT EXISTS idx_tags_name_lower ON tags (lower(name));
CREATE INDEX IF NOT EXISTS idx_post_bookmarks_user_post ON post_bookmarks (user_id, post_slug);
CREATE INDEX IF NOT EXISTS idx_post_bookmarks_post_created_at ON post_bookmarks (post_slug, created_at DESC);

-- Admin user search and filter combinations.
CREATE INDEX IF NOT EXISTS idx_users_status_created_at ON users (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_role_created_at ON users (role, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_email_verified_created_at ON users (email_verified, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (lower(email));
CREATE INDEX IF NOT EXISTS idx_users_display_name_lower ON users (lower(display_name));

-- Comment moderation, personal comment lists, and comment keyword search.
CREATE INDEX IF NOT EXISTS idx_comments_author_status_created_at ON comments (author_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments (parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_body_lower ON comments (lower(body));

-- Submission review queues, personal submission lists, throttling checks, and keyword search.
CREATE INDEX IF NOT EXISTS idx_submissions_author_status_updated_at ON submissions (author_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_status_updated_at ON submissions (status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_author_submitted_at ON submissions (author_id, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_title_lower ON submissions (lower(title));
CREATE INDEX IF NOT EXISTS idx_submissions_summary_lower ON submissions (lower(summary));
CREATE INDEX IF NOT EXISTS idx_submissions_category_lower ON submissions (lower(category));
CREATE INDEX IF NOT EXISTS idx_submissions_slug_lower ON submissions (lower(slug));

-- Inbox, unread counters, archive views, scheduled messages, and message keyword search.
CREATE INDEX IF NOT EXISTS idx_messages_recipient_scheduled_created_at ON messages (recipient_id, scheduled_at, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_archived_read_created_at ON messages (recipient_id, archived_at, read_at, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_type_created_at ON messages (recipient_id, type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_title_lower ON messages (lower(title));
CREATE INDEX IF NOT EXISTS idx_messages_body_lower ON messages (lower(body));
CREATE INDEX IF NOT EXISTS idx_messages_recipient_name_lower ON messages (lower(recipient_name));
CREATE INDEX IF NOT EXISTS idx_messages_target_title_lower ON messages (lower(target_title));

-- Admin draft/post management search and filters.
CREATE INDEX IF NOT EXISTS idx_admin_posts_status_updated_at ON admin_posts (status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_posts_title_lower ON admin_posts (lower(title));

-- Topic search and public/admin topic filters.
CREATE INDEX IF NOT EXISTS idx_topics_status_featured_sort ON topics (status, featured, sort_order ASC, title ASC);
CREATE INDEX IF NOT EXISTS idx_topics_slug_lower ON topics (lower(slug));
CREATE INDEX IF NOT EXISTS idx_topics_title_lower ON topics (lower(title));
CREATE INDEX IF NOT EXISTS idx_topics_summary_lower ON topics (lower(summary));
CREATE INDEX IF NOT EXISTS idx_topics_image_alt_lower ON topics (lower(image_alt));
CREATE INDEX IF NOT EXISTS idx_topics_categories_lower ON topics (lower(CAST(categories AS TEXT)));
CREATE INDEX IF NOT EXISTS idx_topics_tags_lower ON topics (lower(CAST(tags AS TEXT)));

-- Media library search and common filters.
CREATE INDEX IF NOT EXISTS idx_media_assets_file_name_lower ON media_assets (lower(file_name));
CREATE INDEX IF NOT EXISTS idx_media_assets_alt_lower ON media_assets (lower(alt));
CREATE INDEX IF NOT EXISTS idx_media_assets_type_uploaded_at ON media_assets (type, uploaded_at DESC);

-- Audit log filter combinations used by the operations views.
CREATE INDEX IF NOT EXISTS idx_audit_logs_status_created_at ON audit_logs (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type_created_at ON audit_logs (resource_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created_at ON audit_logs (actor_id, created_at DESC);
