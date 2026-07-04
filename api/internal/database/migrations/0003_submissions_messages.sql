CREATE TABLE IF NOT EXISTS submissions (
  id text PRIMARY KEY,
  author_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title text NOT NULL,
  summary text NOT NULL DEFAULT '',
  content text NOT NULL DEFAULT '',
  category text NOT NULL DEFAULT '',
  tags text[] NOT NULL DEFAULT '{}',
  cover_image text NOT NULL DEFAULT '',
  slug text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'returned', 'rejected', 'published')),
  review_note text NOT NULL DEFAULT '',
  reviewer_id text REFERENCES users(id) ON DELETE SET NULL,
  published_post_slug text REFERENCES posts(slug) ON DELETE SET NULL,
  version integer NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  submitted_at timestamptz,
  reviewed_at timestamptz,
  published_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_submissions_author_updated_at ON submissions (author_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_status_submitted_at ON submissions (status, submitted_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_submissions_slug ON submissions (slug);

DROP TRIGGER IF EXISTS touch_submissions_updated_at ON submissions;
CREATE TRIGGER touch_submissions_updated_at
BEFORE UPDATE ON submissions
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TABLE IF NOT EXISTS messages (
  id text PRIMARY KEY,
  recipient_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  recipient_name text NOT NULL DEFAULT '',
  sender_id text NOT NULL DEFAULT 'system',
  sender_name text NOT NULL DEFAULT '系统',
  type text NOT NULL DEFAULT 'admin' CHECK (type IN ('review', 'comment', 'system', 'admin', 'account')),
  priority text NOT NULL DEFAULT 'normal',
  title text NOT NULL,
  body text NOT NULL,
  target_type text NOT NULL DEFAULT '',
  target_id text NOT NULL DEFAULT '',
  target_title text NOT NULL DEFAULT '',
  read_at timestamptz,
  archived_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_created_at ON messages (recipient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_read_at ON messages (recipient_id, read_at);
CREATE INDEX IF NOT EXISTS idx_messages_type_created_at ON messages (type, created_at DESC);
