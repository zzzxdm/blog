CREATE TABLE IF NOT EXISTS users (
  id text PRIMARY KEY,
  email text NOT NULL UNIQUE,
  display_name text NOT NULL,
  role text NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'author', 'editor', 'admin')),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'muted', 'banned', 'deleted')),
  avatar_text text NOT NULL DEFAULT '',
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
  token text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS comments (
  id text PRIMARY KEY,
  post_slug text NOT NULL REFERENCES posts(slug) ON DELETE CASCADE,
  parent_id text REFERENCES comments(id) ON DELETE SET NULL,
  author_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  body text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('approved', 'pending', 'rejected', 'spam', 'deleted')),
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  is_author boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_comments_post_status_created_at ON comments (post_slug, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_author_created_at ON comments (author_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_status_created_at ON comments (status, created_at DESC);

CREATE TABLE IF NOT EXISTS post_interaction_stats (
  post_slug text PRIMARY KEY REFERENCES posts(slug) ON DELETE CASCADE,
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  dislike_count integer NOT NULL DEFAULT 0 CHECK (dislike_count >= 0),
  bookmark_count integer NOT NULL DEFAULT 0 CHECK (bookmark_count >= 0),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_reactions (
  post_slug text NOT NULL REFERENCES posts(slug) ON DELETE CASCADE,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reaction text NOT NULL CHECK (reaction IN ('like', 'dislike')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (post_slug, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_reactions_user_id ON post_reactions (user_id);

CREATE TABLE IF NOT EXISTS post_bookmarks (
  post_slug text NOT NULL REFERENCES posts(slug) ON DELETE CASCADE,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (post_slug, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_bookmarks_user_created_at ON post_bookmarks (user_id, created_at DESC);

DROP TRIGGER IF EXISTS touch_users_updated_at ON users;
CREATE TRIGGER touch_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

DROP TRIGGER IF EXISTS touch_comments_updated_at ON comments;
CREATE TRIGGER touch_comments_updated_at
BEFORE UPDATE ON comments
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

DROP TRIGGER IF EXISTS touch_post_interaction_stats_updated_at ON post_interaction_stats;
CREATE TRIGGER touch_post_interaction_stats_updated_at
BEFORE UPDATE ON post_interaction_stats
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

INSERT INTO post_interaction_stats (post_slug, like_count, dislike_count, bookmark_count)
SELECT
  slug,
  like_count,
  dislike_count,
  CASE slug
    WHEN 'blog-system-design' THEN 34
    WHEN 'vue3-content-site-cache-seo' THEN 18
    WHEN 'postgres-redis-blog-boundary' THEN 25
    ELSE 0
  END
FROM posts
ON CONFLICT (post_slug) DO NOTHING;
