CREATE TABLE IF NOT EXISTS audit_logs (
  id text PRIMARY KEY,
  actor_id text NOT NULL DEFAULT '',
  actor_name text NOT NULL DEFAULT '',
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id text NOT NULL DEFAULT '',
  resource_title text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'success',
  ip text NOT NULL DEFAULT '',
  user_agent text NOT NULL DEFAULT '',
  detail text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs (resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs (actor_id);
