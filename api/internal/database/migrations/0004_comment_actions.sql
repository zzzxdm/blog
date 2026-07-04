CREATE TABLE IF NOT EXISTS comment_likes (
  comment_id text NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON comment_likes (user_id);

CREATE TABLE IF NOT EXISTS comment_reports (
  id text PRIMARY KEY,
  comment_id text NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
  reporter_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reason text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'dismissed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (comment_id, reporter_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_reports_status_created_at ON comment_reports (status, created_at DESC);

DROP TRIGGER IF EXISTS touch_comment_reports_updated_at ON comment_reports;
CREATE TRIGGER touch_comment_reports_updated_at
BEFORE UPDATE ON comment_reports
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
