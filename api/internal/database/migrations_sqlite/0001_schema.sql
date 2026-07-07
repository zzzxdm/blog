CREATE TABLE IF NOT EXISTS categories (
  id integer PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tags (
  id integer PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
  id integer PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  title text NOT NULL,
  summary text NOT NULL DEFAULT '',
  content text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'rejected', 'scheduled', 'published', 'archived')),
  source text NOT NULL DEFAULT 'admin' CHECK (source IN ('admin', 'submission')),
  category_id integer NOT NULL,
  author_name text NOT NULL DEFAULT '管理员',
  cover_image text NOT NULL DEFAULT '',
  reading_time integer NOT NULL DEFAULT 1 CHECK (reading_time > 0),
  view_count integer NOT NULL DEFAULT 0 CHECK (view_count >= 0),
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  dislike_count integer NOT NULL DEFAULT 0 CHECK (dislike_count >= 0),
  comment_count integer NOT NULL DEFAULT 0 CHECK (comment_count >= 0),
  published_at timestamp,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  search_vector text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS post_tags (
  post_id integer NOT NULL,
  tag_id integer NOT NULL,
  PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_posts_status_published_at ON posts (status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts (category_id);
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id ON post_tags (tag_id);

CREATE TABLE IF NOT EXISTS users (
  id integer PRIMARY KEY,
  email text NOT NULL UNIQUE,
  display_name text NOT NULL,
  role text NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'author', 'editor', 'admin')),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'muted', 'banned', 'deleted')),
  avatar_text text NOT NULL DEFAULT '',
  email_verified boolean NOT NULL DEFAULT false,
  password_hash text NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
  token text PRIMARY KEY,
  user_id integer NOT NULL,
  expires_at timestamp NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS comments (
  id integer PRIMARY KEY,
  post_slug text NOT NULL,
  parent_id integer,
  author_id integer NOT NULL,
  body text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('approved', 'pending', 'rejected', 'spam', 'deleted')),
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  is_author boolean NOT NULL DEFAULT false,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_comments_post_status_created_at ON comments (post_slug, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_author_created_at ON comments (author_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_status_created_at ON comments (status, created_at DESC);

CREATE TABLE IF NOT EXISTS post_interaction_stats (
  post_slug text PRIMARY KEY,
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  dislike_count integer NOT NULL DEFAULT 0 CHECK (dislike_count >= 0),
  bookmark_count integer NOT NULL DEFAULT 0 CHECK (bookmark_count >= 0),
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS post_reactions (
  post_slug text NOT NULL,
  user_id integer NOT NULL,
  reaction text NOT NULL CHECK (reaction IN ('like', 'dislike')),
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (post_slug, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_reactions_user_id ON post_reactions (user_id);

CREATE TABLE IF NOT EXISTS post_bookmarks (
  post_slug text NOT NULL,
  user_id integer NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (post_slug, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_bookmarks_user_created_at ON post_bookmarks (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS comment_likes (
  comment_id integer NOT NULL,
  user_id integer NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON comment_likes (user_id);

CREATE TABLE IF NOT EXISTS comment_reports (
  id integer PRIMARY KEY,
  comment_id integer NOT NULL,
  reporter_id integer NOT NULL,
  reason text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'dismissed')),
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (comment_id, reporter_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_reports_status_created_at ON comment_reports (status, created_at DESC);

CREATE TABLE IF NOT EXISTS submissions (
  id integer PRIMARY KEY,
  author_id integer NOT NULL,
  title text NOT NULL,
  summary text NOT NULL DEFAULT '',
  content text NOT NULL DEFAULT '',
  category text NOT NULL DEFAULT '',
  tags text NOT NULL DEFAULT '[]',
  cover_image text NOT NULL DEFAULT '',
  slug text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'returned', 'rejected', 'published')),
  review_note text NOT NULL DEFAULT '',
  reviewer_id integer,
  published_post_slug text,
  version integer NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  submitted_at timestamp,
  reviewed_at timestamp,
  published_at timestamp
);

CREATE INDEX IF NOT EXISTS idx_submissions_author_updated_at ON submissions (author_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_status_submitted_at ON submissions (status, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_slug ON submissions (slug);

CREATE TABLE IF NOT EXISTS messages (
  id integer PRIMARY KEY,
  recipient_id integer NOT NULL,
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
  read_at timestamp,
  archived_at timestamp,
  scheduled_at timestamp,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_created_at ON messages (recipient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_read_at ON messages (recipient_id, read_at);
CREATE INDEX IF NOT EXISTS idx_messages_type_created_at ON messages (type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_scheduled_at ON messages (scheduled_at);

CREATE TABLE IF NOT EXISTS operation_documents (
  key text PRIMARY KEY,
  data text NOT NULL,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS media_assets (
  id integer PRIMARY KEY,
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
  uploaded_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_assets_uploaded_at ON media_assets (uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_assets_category ON media_assets (category);

CREATE TABLE IF NOT EXISTS account_settings (
  user_id integer PRIMARY KEY,
  data text NOT NULL,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_posts (
  id integer PRIMARY KEY,
  slug text NOT NULL,
  title text NOT NULL,
  status text NOT NULL CHECK (status IN ('draft', 'review', 'scheduled', 'published', 'archived')),
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  data text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_admin_posts_updated_at ON admin_posts (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_posts_status ON admin_posts (status);
CREATE INDEX IF NOT EXISTS idx_admin_posts_slug ON admin_posts (slug);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
  token text PRIMARY KEY,
  user_id integer NOT NULL,
  expires_at timestamp NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_expires_at ON email_verification_tokens (expires_at);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
  token text PRIMARY KEY,
  user_id integer NOT NULL,
  expires_at timestamp NOT NULL,
  used_at timestamp,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens (expires_at);

CREATE TABLE IF NOT EXISTS audit_logs (
  id integer PRIMARY KEY,
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
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs (resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs (actor_id);

CREATE TABLE IF NOT EXISTS topics (
  id integer PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  title text NOT NULL UNIQUE,
  summary text NOT NULL DEFAULT '',
  cover_image text NOT NULL DEFAULT '',
  image_alt text NOT NULL DEFAULT '',
  tone text NOT NULL DEFAULT '' CHECK (tone IN ('', 'rust', 'amber', 'gray')),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'draft')),
  featured boolean NOT NULL DEFAULT true,
  sort_order integer NOT NULL DEFAULT 0,
  categories text NOT NULL DEFAULT '[]',
  tags text NOT NULL DEFAULT '[]',
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_topics_status_sort ON topics (status, sort_order ASC, title ASC);
CREATE INDEX IF NOT EXISTS idx_topics_featured_sort ON topics (featured, sort_order ASC);

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
CREATE INDEX IF NOT EXISTS idx_users_status_created_at ON users (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_role_created_at ON users (role, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_email_verified_created_at ON users (email_verified, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (lower(email));
CREATE INDEX IF NOT EXISTS idx_users_display_name_lower ON users (lower(display_name));
CREATE INDEX IF NOT EXISTS idx_comments_author_status_created_at ON comments (author_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments (parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_body_lower ON comments (lower(body));
CREATE INDEX IF NOT EXISTS idx_submissions_author_status_updated_at ON submissions (author_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_status_updated_at ON submissions (status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_author_submitted_at ON submissions (author_id, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_title_lower ON submissions (lower(title));
CREATE INDEX IF NOT EXISTS idx_submissions_summary_lower ON submissions (lower(summary));
CREATE INDEX IF NOT EXISTS idx_submissions_category_lower ON submissions (lower(category));
CREATE INDEX IF NOT EXISTS idx_submissions_slug_lower ON submissions (lower(slug));
CREATE INDEX IF NOT EXISTS idx_messages_recipient_scheduled_created_at ON messages (recipient_id, scheduled_at, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_archived_read_created_at ON messages (recipient_id, archived_at, read_at, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_type_created_at ON messages (recipient_id, type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_title_lower ON messages (lower(title));
CREATE INDEX IF NOT EXISTS idx_messages_body_lower ON messages (lower(body));
CREATE INDEX IF NOT EXISTS idx_messages_recipient_name_lower ON messages (lower(recipient_name));
CREATE INDEX IF NOT EXISTS idx_messages_target_title_lower ON messages (lower(target_title));
CREATE INDEX IF NOT EXISTS idx_admin_posts_status_updated_at ON admin_posts (status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_posts_title_lower ON admin_posts (lower(title));
CREATE INDEX IF NOT EXISTS idx_topics_status_featured_sort ON topics (status, featured, sort_order ASC, title ASC);
CREATE INDEX IF NOT EXISTS idx_topics_slug_lower ON topics (lower(slug));
CREATE INDEX IF NOT EXISTS idx_topics_title_lower ON topics (lower(title));
CREATE INDEX IF NOT EXISTS idx_topics_summary_lower ON topics (lower(summary));
CREATE INDEX IF NOT EXISTS idx_topics_image_alt_lower ON topics (lower(image_alt));
CREATE INDEX IF NOT EXISTS idx_topics_categories_lower ON topics (lower(CAST(categories AS TEXT)));
CREATE INDEX IF NOT EXISTS idx_topics_tags_lower ON topics (lower(CAST(tags AS TEXT)));
CREATE INDEX IF NOT EXISTS idx_media_assets_file_name_lower ON media_assets (lower(file_name));
CREATE INDEX IF NOT EXISTS idx_media_assets_alt_lower ON media_assets (lower(alt));
CREATE INDEX IF NOT EXISTS idx_media_assets_type_uploaded_at ON media_assets (type, uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status_created_at ON audit_logs (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type_created_at ON audit_logs (resource_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created_at ON audit_logs (actor_id, created_at DESC);

INSERT OR IGNORE INTO categories (id, slug, name, description, sort_order) VALUES
  (1001, 'engineering', '工程实践', '工程方法、架构落地和长期维护经验。', 10),
  (1002, 'architecture', '架构', '系统边界、数据层和基础设施设计。', 20),
  (1005, 'vue3', 'Vue3', 'Vue3 内容站、前端架构和交互实现。', 30)
;

INSERT OR IGNORE INTO tags (id, slug, name) VALUES
  (2001, 'blog-system', '博客系统'),
  (2002, 'architecture', '架构'),
  (2003, 'content-governance', '内容治理'),
  (2004, 'vue3', 'Vue3'),
  (2005, 'seo', 'SEO'),
  (2006, 'cache', '缓存'),
  (2007, 'postgresql', 'PostgreSQL'),
  (2008, 'redis', 'Redis'),
  (2009, 'full-text-search', '全文搜索')
;

INSERT OR IGNORE INTO posts (
  id, slug, title, summary, content, status, source, category_id, author_name,
  cover_image, reading_time, view_count, like_count, dislike_count, comment_count, published_at
) VALUES
  (
    3001,
    'blog-system-design',
    '如何设计一个内容长期增长的博客系统',
    '博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。',
    '一个现代化博客系统需要从内容资产的生命周期开始设计。文章不是一次性页面，而是会被修改、引用、搜索、迁移和长期展示的结构化内容。',
    'published',
    'admin',
    1001,
    '管理员',
    'https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1200&q=80',
    12,
    2984,
    128,
    7,
    34,
    '2026-07-04T00:00:00Z'
  ),
  (
    3002,
    'vue3-content-site-cache-seo',
    'Vue3 内容站的缓存与 SEO 边界',
    '客户端渲染、接口缓存和服务端 meta 需要明确边界，避免前期开发轻松、后期收录困难。',
    'Vue3 内容站可以保持前端开发效率，同时通过 Go 输出基础 HTML、meta 和结构化数据处理文章页 SEO。',
    'published',
    'admin',
    1005,
    '管理员',
    'https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1200&q=80',
    8,
    4120,
    96,
    3,
    18,
    '2026-06-25T00:00:00Z'
  ),
  (
    3003,
    'postgres-redis-blog-boundary',
    'Redis 和 PostgreSQL 在博客中的分工',
    'PostgreSQL 保存事实并承担全文搜索，Redis 负责热点读取、会话、限流和异步任务协调。',
    '个人博客早期没有必要引入专用搜索中间件。PostgreSQL 的 tsvector 和 GIN 索引足以覆盖标题、摘要、正文和标签搜索。',
    'published',
    'admin',
    1002,
    '管理员',
    'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=1200&q=80',
    14,
    3019,
    84,
    4,
    25,
    '2026-07-01T00:00:00Z'
  )
;

INSERT OR IGNORE INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM (
  SELECT 'blog-system-design' AS post_slug, 'blog-system' AS tag_slug UNION ALL
  SELECT 'blog-system-design', 'architecture' UNION ALL
  SELECT 'blog-system-design', 'content-governance' UNION ALL
  SELECT 'vue3-content-site-cache-seo', 'vue3' UNION ALL
  SELECT 'vue3-content-site-cache-seo', 'seo' UNION ALL
  SELECT 'vue3-content-site-cache-seo', 'cache' UNION ALL
  SELECT 'postgres-redis-blog-boundary', 'postgresql' UNION ALL
  SELECT 'postgres-redis-blog-boundary', 'redis' UNION ALL
  SELECT 'postgres-redis-blog-boundary', 'full-text-search'
) pairs
JOIN posts p ON p.slug = pairs.post_slug
JOIN tags t ON t.slug = pairs.tag_slug
;

INSERT OR IGNORE INTO post_interaction_stats (post_slug, like_count, dislike_count, bookmark_count)
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
;

INSERT OR IGNORE INTO topics (
  id, slug, title, summary, cover_image, image_alt, tone, status, featured, sort_order, categories, tags
) VALUES
  (
    4001,
    'blog-system',
    '现代化博客系统',
    '从产品功能、技术架构、用户系统、评论、搜索和后台管理完整设计一个博客系统。',
    'https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=900&q=80',
    '代码编辑器和开发设备',
    '',
    'active',
    true,
    10,
    '["工程实践","产品设计","用户系统","内容治理"]',
    '["博客系统","架构","内容治理","评论"]'
  ),
  (
    4002,
    'vue3-content',
    'Vue3 内容站',
    '路由、状态管理、接口缓存、SEO meta、图片优化和部署策略。',
    'https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=900&q=80',
    '代码编辑器中的程序文件',
    'rust',
    'active',
    true,
    20,
    '["Vue3"]',
    '["Vue3","SEO","缓存"]'
  ),
  (
    4003,
    'writing-workflow',
    '写作工作流',
    '草稿、版本历史、编辑器、发布审批和长期内容维护。',
    'https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=900&q=80',
    '笔记本和写作草稿',
    'amber',
    'active',
    true,
    30,
    '["写作工作流"]',
    '["工作流","写作工作流","Markdown"]'
  ),
  (
    4004,
    'resource-list',
    '资源清单',
    '把工具、部署、数据库和内容运营资料整理成可持续更新的阅读路线。',
    'https://images.unsplash.com/photo-1484480974693-6ca0a78fb36b?auto=format&fit=crop&w=900&q=80',
    '桌面上的计划清单和电脑',
    '',
    'active',
    true,
    40,
    '["架构","运营"]',
    '["PostgreSQL","Redis","全文搜索","SEO"]'
  )
;
