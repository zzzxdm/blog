ALTER TABLE messages
ADD COLUMN IF NOT EXISTS scheduled_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_messages_scheduled_at ON messages (scheduled_at);
