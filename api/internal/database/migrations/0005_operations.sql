CREATE TABLE IF NOT EXISTS operation_documents (
  key text PRIMARY KEY,
  data jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS media_assets (
  id text PRIMARY KEY,
  file_name text NOT NULL,
  url text NOT NULL,
  alt text NOT NULL DEFAULT '',
  type text NOT NULL DEFAULT 'image',
  category text NOT NULL DEFAULT '',
  size_label text NOT NULL DEFAULT '',
  width integer NOT NULL DEFAULT 0,
  height integer NOT NULL DEFAULT 0,
  usage_count integer NOT NULL DEFAULT 0 CHECK (usage_count >= 0),
  uploaded_by text NOT NULL DEFAULT '',
  uploaded_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_media_assets_uploaded_at ON media_assets (uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_assets_category ON media_assets (category);

DROP TRIGGER IF EXISTS touch_operation_documents_updated_at ON operation_documents;
CREATE TRIGGER touch_operation_documents_updated_at
BEFORE UPDATE ON operation_documents
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
