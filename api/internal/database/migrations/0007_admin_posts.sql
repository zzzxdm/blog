CREATE TABLE IF NOT EXISTS admin_posts (
  id text PRIMARY KEY,
  slug text NOT NULL,
  title text NOT NULL,
  status text NOT NULL CHECK (status IN ('draft', 'review', 'scheduled', 'published', 'archived')),
  updated_at timestamptz NOT NULL DEFAULT now(),
  data jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_admin_posts_updated_at ON admin_posts (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_posts_status ON admin_posts (status);
CREATE INDEX IF NOT EXISTS idx_admin_posts_slug ON admin_posts (slug);
