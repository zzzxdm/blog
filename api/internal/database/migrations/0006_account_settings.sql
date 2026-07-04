CREATE TABLE IF NOT EXISTS account_settings (
  user_id text PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  data jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

DROP TRIGGER IF EXISTS touch_account_settings_updated_at ON account_settings;
CREATE TRIGGER touch_account_settings_updated_at
BEFORE UPDATE ON account_settings
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
