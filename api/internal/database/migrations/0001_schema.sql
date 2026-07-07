CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS categories (
  id bigint PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tags (
  id bigint PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS posts (
  id bigint PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  title text NOT NULL,
  summary text NOT NULL DEFAULT '',
  content text NOT NULL DEFAULT '',
  visibility text NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'rejected', 'scheduled', 'published', 'archived')),
  source text NOT NULL DEFAULT 'admin' CHECK (source IN ('admin', 'submission')),
  category_id bigint NOT NULL,
  author_id bigint,
  author_name text NOT NULL DEFAULT '管理员',
  cover_image text NOT NULL DEFAULT '',
  reading_time integer NOT NULL DEFAULT 1 CHECK (reading_time > 0),
  view_count integer NOT NULL DEFAULT 0 CHECK (view_count >= 0),
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  dislike_count integer NOT NULL DEFAULT 0 CHECK (dislike_count >= 0),
  comment_count integer NOT NULL DEFAULT 0 CHECK (comment_count >= 0),
  published_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  search_vector tsvector NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS post_tags (
  post_id bigint NOT NULL,
  tag_id bigint NOT NULL,
  PRIMARY KEY (post_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_posts_status_visibility_published_at ON posts (status, visibility, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_author_id_visibility ON posts (author_id, visibility);
CREATE INDEX IF NOT EXISTS idx_posts_status_published_at ON posts (status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts (category_id);
CREATE INDEX IF NOT EXISTS idx_posts_search_vector ON posts USING gin (search_vector);
CREATE INDEX IF NOT EXISTS idx_posts_title_trgm ON posts USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_summary_trgm ON posts USING gin (summary gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tags_name_trgm ON tags USING gin (name gin_trgm_ops);

CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS touch_categories_updated_at ON categories;
CREATE TRIGGER touch_categories_updated_at
BEFORE UPDATE ON categories
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

DROP TRIGGER IF EXISTS touch_tags_updated_at ON tags;
CREATE TRIGGER touch_tags_updated_at
BEFORE UPDATE ON tags
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

DROP TRIGGER IF EXISTS touch_posts_updated_at ON posts;
CREATE TRIGGER touch_posts_updated_at
BEFORE UPDATE ON posts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE OR REPLACE FUNCTION refresh_post_search_vector(target_post_id bigint)
RETURNS void AS $$
  UPDATE posts p
  SET search_vector =
    setweight(to_tsvector('simple', coalesce(p.title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(p.summary, '')), 'B') ||
    setweight(to_tsvector('simple', coalesce(c.name, '')), 'B') ||
    setweight(to_tsvector('simple', coalesce((
      SELECT string_agg(t.name, ' ')
      FROM post_tags pt
      JOIN tags t ON t.id = pt.tag_id
      WHERE pt.post_id = p.id
    ), '')), 'B') ||
    setweight(to_tsvector('simple', coalesce(p.content, '')), 'C')
  FROM categories c
  WHERE p.id = target_post_id
    AND c.id = p.category_id;
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION posts_refresh_search_vector_trigger()
RETURNS trigger AS $$
BEGIN
  PERFORM refresh_post_search_vector(NEW.id);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS posts_refresh_search_vector ON posts;
CREATE TRIGGER posts_refresh_search_vector
AFTER INSERT OR UPDATE OF title, summary, content, category_id ON posts
FOR EACH ROW EXECUTE FUNCTION posts_refresh_search_vector_trigger();

CREATE OR REPLACE FUNCTION post_tags_refresh_search_vector_trigger()
RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    PERFORM refresh_post_search_vector(OLD.post_id);
  ELSE
    PERFORM refresh_post_search_vector(NEW.post_id);
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS post_tags_refresh_search_vector ON post_tags;
CREATE TRIGGER post_tags_refresh_search_vector
AFTER INSERT OR UPDATE OR DELETE ON post_tags
FOR EACH ROW EXECUTE FUNCTION post_tags_refresh_search_vector_trigger();

INSERT INTO categories (id, slug, name, description, sort_order) VALUES
  (1001, 'engineering', '工程实践', '工程方法、架构落地和长期维护经验。', 10),
  (1002, 'architecture', '架构', '系统边界、数据层和基础设施设计。', 20),
  (1003, 'product-design', '产品设计', '信息架构、交互体验和内容产品设计。', 30),
  (1004, 'operations', '运营', '内容运营、增长反馈和站点治理。', 40),
  (1005, 'vue3', 'Vue3', 'Vue3 内容站、前端架构和交互实现。', 50),
  (1006, 'workflow', '写作工作流', '投稿、审核、编辑器和内容生命周期。', 60)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  sort_order = EXCLUDED.sort_order;

INSERT INTO tags (id, slug, name) VALUES
  (2001, 'blog-system', '博客系统'),
  (2002, 'architecture', '架构'),
  (2003, 'content-governance', '内容治理'),
  (2004, 'vue3', 'Vue3'),
  (2005, 'seo', 'SEO'),
  (2006, 'cache', '缓存'),
  (2007, 'postgresql', 'PostgreSQL'),
  (2008, 'redis', 'Redis'),
  (2009, 'full-text-search', '全文搜索'),
  (2010, 'information-architecture', '信息架构'),
  (2011, 'markdown', 'Markdown'),
  (2012, 'rss', 'RSS'),
  (2013, 'comments', '评论'),
  (2014, 'submission', '投稿'),
  (2015, 'message', '站内信'),
  (2016, 'media', '媒体库'),
  (2017, 'navigation', '导航'),
  (2018, 'account', '账号'),
  (2019, 'workflow', '工作流'),
  (2020, 'operations', '运营')
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO posts (
  id, slug, title, summary, content, status, source, category_id, author_name,
  cover_image, reading_time, view_count, like_count, dislike_count, comment_count, published_at
) VALUES
  (
    3001,
    'blog-system-design',
    '如何设计一个内容长期增长的博客系统',
    '博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。',
    '一个现代化博客系统需要从内容资产的生命周期开始设计。文章不是一次性页面，而是会被修改、引用、搜索、迁移和长期展示的结构化内容。如果早期只实现简单的增删改查，后期补版本历史、重定向、搜索索引和订阅推送时，往往会遇到数据模型不够稳定的问题。',
    'published',
    'admin',
    1001,
    '管理员',
    'https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1400&q=80',
    12,
    2984,
    128,
    7,
    34,
    '2026-07-04 09:00:00+08'
  ),
  (
    3002,
    'vue3-content-site-cache-seo',
    'Vue3 内容站的缓存与 SEO 边界',
    '客户端渲染、接口缓存和服务端 meta 需要明确边界，避免前期开发轻松、后期收录困难。',
    'Vue3 内容站可以保持前端开发效率，同时通过 Go 输出基础 HTML、meta 和结构化数据处理文章页 SEO。缓存策略、页面更新和接口数据需要从第一版开始定义边界。',
    'published',
    'admin',
    1005,
    '管理员',
    'https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1400&q=80',
    8,
    4120,
    96,
    3,
    18,
    '2026-06-25 09:00:00+08'
  ),
  (
    3003,
    'postgres-redis-blog-boundary',
    'Redis 和 PostgreSQL 在博客中的分工',
    'PostgreSQL 保存事实并承担全文搜索，Redis 负责热点读取、会话、限流和异步任务协调。',
    '个人博客早期没有必要引入专用搜索中间件。PostgreSQL 的 tsvector 和 GIN 索引足以覆盖标题、摘要、正文和标签搜索，Redis 则用于热点数据、会话、限流和异步任务协调。',
    'published',
    'admin',
    1002,
    '管理员',
    'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=1400&q=80',
    14,
    3019,
    84,
    4,
    25,
    '2026-07-01 09:00:00+08'
  ),
  (
    3004,
    'home-to-article-information-architecture',
    '从首页到文章页，博客的信息架构怎么排',
    '首页负责分发，文章页负责阅读，归档页负责找回内容。三者的密度和交互应该完全不同。',
    '首页需要帮助读者发现内容，文章页需要让读者沉浸阅读，归档页需要承担高效率查找。把这些页面设计成同一种卡片流，会让每个场景都不够顺手。',
    'published',
    'admin',
    1003,
    '管理员',
    'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=1400&q=80',
    10,
    2184,
    76,
    2,
    18,
    '2026-07-03 09:00:00+08'
  ),
  (
    3005,
    'post-version-history',
    '为什么博客后台需要文章版本历史',
    '版本记录不是复杂功能，而是内容资产的基本保险。',
    '文章会被持续修订，后台需要记录版本历史、修改人、变更摘要和回滚能力。对于长期运营的博客，这可以避免误操作带来的内容资产损失。',
    'published',
    'admin',
    1006,
    '管理员',
    'https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80',
    6,
    1988,
    61,
    1,
    12,
    '2026-06-18 09:00:00+08'
  ),
  (
    3006,
    'markdown-writing-experience',
    '把 Markdown 写作体验做到顺手',
    '编辑器、预览、封面和 SEO 字段要服务写作流程，而不是把作者困在表单里。',
    'Markdown 编辑器需要稳定的草稿保存、预览、图片插入、代码块处理和 SEO 字段编辑。写作体验越顺手，内容生产越稳定。',
    'published',
    'admin',
    1006,
    '管理员',
    'https://images.unsplash.com/photo-1499750310107-5fef28a66643?auto=format&fit=crop&w=1400&q=80',
    9,
    2450,
    88,
    3,
    16,
    '2026-06-30 09:00:00+08'
  ),
  (
    3007,
    'rss-stable-distribution',
    'RSS 是博客最稳定的分发入口',
    '早期先保留 RSS，把邮件推送能力放到内容稳定增长后再开启。',
    'RSS 不需要复杂用户系统，也不依赖第三方平台算法。对独立博客来说，它是低成本、低维护的稳定分发入口。',
    'published',
    'admin',
    1004,
    '管理员',
    'https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=1400&q=80',
    7,
    1624,
    52,
    2,
    9,
    '2026-06-29 09:00:00+08'
  ),
  (
    3008,
    'comments-system-design',
    '用户评论系统应该怎么设计',
    '登录用户评论、审核、举报、站内信和禁言机制需要从同一套权限边界出发。',
    '评论系统的重点不是输入框，而是审核、通知、举报、禁言、删除记录和内容安全。默认待审核可以降低早期运营风险。',
    'published',
    'admin',
    1001,
    '管理员',
    'https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?auto=format&fit=crop&w=1400&q=80',
    11,
    1835,
    71,
    5,
    28,
    '2026-06-22 09:00:00+08'
  ),
  (
    3009,
    'submission-review-workflow',
    '登录用户投稿到审核发布的完整闭环',
    '普通用户可以投稿，但不能直接发布；审核通过后再进入正式文章列表。',
    '投稿流程需要草稿、提交审核、退回修改、拒绝、通过发布和审核意见。审核结果还应同步到站内信和我的投稿列表。',
    'published',
    'admin',
    1006,
    '管理员',
    'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1400&q=80',
    13,
    2210,
    82,
    4,
    21,
    '2026-06-15 09:00:00+08'
  ),
  (
    3010,
    'station-message-design',
    '站内信适合承载哪些系统通知',
    '审核结果、评论回复、系统公告和管理员定向消息都适合进入站内信。',
    '站内信需要未读、已读、归档、删除和管理员发送能力。它承载的是站内重要事件，不应替代即时聊天。',
    'published',
    'admin',
    1004,
    '管理员',
    'https://images.unsplash.com/photo-1516542076529-1ea3854896f2?auto=format&fit=crop&w=1400&q=80',
    8,
    1542,
    45,
    1,
    7,
    '2026-06-10 09:00:00+08'
  ),
  (
    3011,
    'media-library-storage',
    '媒体库不只是上传图片',
    '图片、附件、Alt 文本、引用关系和删除影响检查决定了媒体库是否可长期维护。',
    '媒体库需要记录资源元数据、生成缩略图、校验类型和大小，并在删除前检查引用关系。生产环境建议使用 S3 兼容对象存储。',
    'published',
    'admin',
    1002,
    '管理员',
    'https://images.unsplash.com/photo-1483058712412-4245e9b90334?auto=format&fit=crop&w=1400&q=80',
    9,
    1766,
    54,
    2,
    11,
    '2026-06-06 09:00:00+08'
  ),
  (
    3012,
    'admin-navigation-settings',
    '后台导航和系统设置怎么拆页面',
    '导航、设置、统计和媒体库都是独立运营能力，不能只停留在入口卡片。',
    '后台页面需要按运营任务拆分，而不是把所有能力堆到概览。导航管理和系统设置尤其需要独立表单和变更校验。',
    'published',
    'admin',
    1003,
    '管理员',
    'https://images.unsplash.com/photo-1460925895917-afdab827c52f?auto=format&fit=crop&w=1400&q=80',
    10,
    1908,
    63,
    3,
    13,
    '2026-06-02 09:00:00+08'
  ),
  (
    3013,
    'postgres-full-text-search',
    '用 PostgreSQL 做博客全文搜索够不够',
    '对于早期博客，tsvector、GIN 索引和少量模糊匹配已经足够支撑标题、摘要、正文和标签搜索。',
    '专用搜索服务会增加部署和运维成本。除非搜索规模和分词需求明显超过 PostgreSQL 能力，否则可以先用数据库内建搜索能力。',
    'published',
    'admin',
    1002,
    '管理员',
    'https://images.unsplash.com/photo-1526378722484-bd91ca387e72?auto=format&fit=crop&w=1400&q=80',
    12,
    2675,
    93,
    4,
    19,
    '2026-05-28 09:00:00+08'
  ),
  (
    3014,
    'footbar-runtime-component',
    '把站点 Footbar 做成公共组件',
    '运行时间、版权、访问统计和站点入口应由公共组件统一维护。',
    '底部栏会出现在多个页面，适合抽成组件。运行时间从固定日期动态计算，社交入口应保持克制并靠近文字。',
    'published',
    'admin',
    1005,
    '管理员',
    'https://images.unsplash.com/photo-1497366754035-f200968a6e72?auto=format&fit=crop&w=1400&q=80',
    5,
    1210,
    37,
    1,
    5,
    '2026-05-20 09:00:00+08'
  ),
  (
    3015,
    'account-settings-security',
    '个人中心账号设置要包含哪些能力',
    '资料、头像、密码、登录设备、数据导出和账号注销都需要明确入口。',
    '账号设置不是一个简单昵称表单。它涉及安全、隐私、会话管理、数据导出和账号注销，需要和个人中心其他页面保持一致。',
    'published',
    'admin',
    1001,
    '管理员',
    'https://images.unsplash.com/photo-1516321497487-e288fb19713f?auto=format&fit=crop&w=1400&q=80',
    8,
    1388,
    49,
    2,
    8,
    '2026-05-12 09:00:00+08'
  )
ON CONFLICT (slug) DO UPDATE SET
  title = EXCLUDED.title,
  summary = EXCLUDED.summary,
  content = EXCLUDED.content,
  status = EXCLUDED.status,
  source = EXCLUDED.source,
  category_id = EXCLUDED.category_id,
  author_name = EXCLUDED.author_name,
  cover_image = EXCLUDED.cover_image,
  reading_time = EXCLUDED.reading_time,
  view_count = EXCLUDED.view_count,
  like_count = EXCLUDED.like_count,
  dislike_count = EXCLUDED.dislike_count,
  comment_count = EXCLUDED.comment_count,
  published_at = EXCLUDED.published_at;

INSERT INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM (
  VALUES
    ('blog-system-design', 'blog-system'),
    ('blog-system-design', 'architecture'),
    ('blog-system-design', 'content-governance'),
    ('vue3-content-site-cache-seo', 'vue3'),
    ('vue3-content-site-cache-seo', 'seo'),
    ('vue3-content-site-cache-seo', 'cache'),
    ('postgres-redis-blog-boundary', 'postgresql'),
    ('postgres-redis-blog-boundary', 'redis'),
    ('postgres-redis-blog-boundary', 'full-text-search'),
    ('home-to-article-information-architecture', 'information-architecture'),
    ('home-to-article-information-architecture', 'blog-system'),
    ('post-version-history', 'content-governance'),
    ('post-version-history', 'workflow'),
    ('markdown-writing-experience', 'markdown'),
    ('markdown-writing-experience', 'workflow'),
    ('rss-stable-distribution', 'rss'),
    ('rss-stable-distribution', 'operations'),
    ('comments-system-design', 'comments'),
    ('comments-system-design', 'blog-system'),
    ('submission-review-workflow', 'submission'),
    ('submission-review-workflow', 'workflow'),
    ('station-message-design', 'message'),
    ('media-library-storage', 'media'),
    ('media-library-storage', 'architecture'),
    ('admin-navigation-settings', 'navigation'),
    ('admin-navigation-settings', 'blog-system'),
    ('postgres-full-text-search', 'postgresql'),
    ('postgres-full-text-search', 'full-text-search'),
    ('footbar-runtime-component', 'vue3'),
    ('account-settings-security', 'account')
) AS pairs(post_slug, tag_slug)
JOIN posts p ON p.slug = pairs.post_slug
JOIN tags t ON t.slug = pairs.tag_slug
ON CONFLICT DO NOTHING;

SELECT refresh_post_search_vector(id) FROM posts;

CREATE TABLE IF NOT EXISTS users (
  id bigint PRIMARY KEY,
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
  user_id bigint NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS comments (
  id bigint PRIMARY KEY,
  post_slug text NOT NULL,
  parent_id bigint,
  author_id bigint NOT NULL,
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
  post_slug text PRIMARY KEY,
  like_count integer NOT NULL DEFAULT 0 CHECK (like_count >= 0),
  dislike_count integer NOT NULL DEFAULT 0 CHECK (dislike_count >= 0),
  bookmark_count integer NOT NULL DEFAULT 0 CHECK (bookmark_count >= 0),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_reactions (
  post_slug text NOT NULL,
  user_id bigint NOT NULL,
  reaction text NOT NULL CHECK (reaction IN ('like', 'dislike')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (post_slug, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_reactions_user_id ON post_reactions (user_id);

CREATE TABLE IF NOT EXISTS post_bookmarks (
  post_slug text NOT NULL,
  user_id bigint NOT NULL,
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

CREATE TABLE IF NOT EXISTS submissions (
  id bigint PRIMARY KEY,
  author_id bigint NOT NULL,
  title text NOT NULL,
  summary text NOT NULL DEFAULT '',
  content text NOT NULL DEFAULT '',
  category text NOT NULL DEFAULT '',
  tags text[] NOT NULL DEFAULT '{}',
  cover_image text NOT NULL DEFAULT '',
  slug text NOT NULL DEFAULT '',
  visibility text NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'returned', 'rejected', 'published', 'archived')),
  review_note text NOT NULL DEFAULT '',
  reviewer_id bigint,
  published_post_slug text,
  version integer NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  submitted_at timestamptz,
  reviewed_at timestamptz,
  published_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_submissions_author_updated_at ON submissions (author_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_author_visibility_updated_at ON submissions (author_id, visibility, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_status_submitted_at ON submissions (status, submitted_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_submissions_visibility_status ON submissions (visibility, status);
CREATE INDEX IF NOT EXISTS idx_submissions_slug ON submissions (slug);

DROP TRIGGER IF EXISTS touch_submissions_updated_at ON submissions;
CREATE TRIGGER touch_submissions_updated_at
BEFORE UPDATE ON submissions
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TABLE IF NOT EXISTS messages (
  id bigint PRIMARY KEY,
  recipient_id bigint NOT NULL,
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

CREATE TABLE IF NOT EXISTS comment_likes (
  comment_id bigint NOT NULL,
  user_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON comment_likes (user_id);

CREATE TABLE IF NOT EXISTS comment_reports (
  id bigint PRIMARY KEY,
  comment_id bigint NOT NULL,
  reporter_id bigint NOT NULL,
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

CREATE TABLE IF NOT EXISTS operation_documents (
  key text PRIMARY KEY,
  data jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS media_assets (
  id bigint PRIMARY KEY,
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

CREATE TABLE IF NOT EXISTS account_settings (
  user_id bigint PRIMARY KEY,
  data jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

DROP TRIGGER IF EXISTS touch_account_settings_updated_at ON account_settings;
CREATE TRIGGER touch_account_settings_updated_at
BEFORE UPDATE ON account_settings
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TABLE IF NOT EXISTS admin_posts (
  id bigint PRIMARY KEY,
  slug text NOT NULL,
  title text NOT NULL,
  status text NOT NULL CHECK (status IN ('draft', 'review', 'scheduled', 'published', 'archived')),
  updated_at timestamptz NOT NULL DEFAULT now(),
  data jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_admin_posts_updated_at ON admin_posts (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_posts_status ON admin_posts (status);
CREATE INDEX IF NOT EXISTS idx_admin_posts_slug ON admin_posts (slug);

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_verified boolean NOT NULL DEFAULT false;

UPDATE users
SET email_verified = true
WHERE email IN ('admin@example.com', 'linyi@example.com', 'chen@example.com', 'market@example.com', 'noise@example.com');

CREATE TABLE IF NOT EXISTS email_verification_tokens (
  token text PRIMARY KEY,
  user_id bigint NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_expires_at ON email_verification_tokens (expires_at);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
  token text PRIMARY KEY,
  user_id bigint NOT NULL,
  expires_at timestamptz NOT NULL,
  used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens (expires_at);

CREATE TABLE IF NOT EXISTS audit_logs (
  id bigint PRIMARY KEY,
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

ALTER TABLE messages
ADD COLUMN IF NOT EXISTS scheduled_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_messages_scheduled_at ON messages (scheduled_at);

UPDATE operation_documents
SET data = jsonb_set(data, '{footerItems}', '[]'::jsonb, true)
WHERE key = 'navigation'
  AND jsonb_typeof(data->'footerItems') = 'array'
  AND jsonb_array_length(data->'footerItems') = 3
  AND data->'footerItems' @> '[
    {"id":"nav_footer_1","url":"/"},
    {"id":"nav_footer_2","url":"/archive"},
    {"id":"nav_footer_3","url":"/topics"}
  ]'::jsonb;

CREATE TABLE IF NOT EXISTS topics (
  id bigint PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  title text NOT NULL UNIQUE,
  summary text NOT NULL DEFAULT '',
  cover_image text NOT NULL DEFAULT '',
  image_alt text NOT NULL DEFAULT '',
  tone text NOT NULL DEFAULT '' CHECK (tone IN ('', 'rust', 'amber', 'gray')),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'draft')),
  featured boolean NOT NULL DEFAULT true,
  sort_order integer NOT NULL DEFAULT 0,
  categories jsonb NOT NULL DEFAULT '[]'::jsonb,
  tags jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_topics_status_sort ON topics (status, sort_order ASC, title ASC);
CREATE INDEX IF NOT EXISTS idx_topics_featured_sort ON topics (featured, sort_order ASC);

DROP TRIGGER IF EXISTS touch_topics_updated_at ON topics;
CREATE TRIGGER touch_topics_updated_at
BEFORE UPDATE ON topics
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

INSERT INTO topics (
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
    '["工程实践", "产品设计", "用户系统", "内容治理"]'::jsonb,
    '["博客系统", "架构", "内容治理", "评论"]'::jsonb
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
    '["Vue3"]'::jsonb,
    '["Vue3", "SEO", "缓存"]'::jsonb
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
    '["写作工作流"]'::jsonb,
    '["工作流", "写作工作流", "Markdown"]'::jsonb
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
    '["架构", "运营"]'::jsonb,
    '["PostgreSQL", "Redis", "全文搜索", "SEO"]'::jsonb
  )
ON CONFLICT (slug) DO NOTHING;
