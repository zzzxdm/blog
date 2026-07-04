CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tags (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS posts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug text NOT NULL UNIQUE,
  title text NOT NULL,
  summary text NOT NULL DEFAULT '',
  content text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'rejected', 'scheduled', 'published', 'archived')),
  source text NOT NULL DEFAULT 'admin' CHECK (source IN ('admin', 'submission')),
  category_id uuid NOT NULL REFERENCES categories(id),
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
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  tag_id uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (post_id, tag_id)
);

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

CREATE OR REPLACE FUNCTION refresh_post_search_vector(target_post_id uuid)
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
  ('11111111-1111-1111-1111-111111111111', 'engineering', '工程实践', '工程方法、架构落地和长期维护经验。', 10),
  ('11111111-1111-1111-1111-111111111112', 'architecture', '架构', '系统边界、数据层和基础设施设计。', 20),
  ('11111111-1111-1111-1111-111111111113', 'product-design', '产品设计', '信息架构、交互体验和内容产品设计。', 30),
  ('11111111-1111-1111-1111-111111111114', 'operations', '运营', '内容运营、增长反馈和站点治理。', 40),
  ('11111111-1111-1111-1111-111111111115', 'vue3', 'Vue3', 'Vue3 内容站、前端架构和交互实现。', 50),
  ('11111111-1111-1111-1111-111111111116', 'workflow', '写作工作流', '投稿、审核、编辑器和内容生命周期。', 60)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  sort_order = EXCLUDED.sort_order;

INSERT INTO tags (id, slug, name) VALUES
  ('22222222-2222-2222-2222-222222222201', 'blog-system', '博客系统'),
  ('22222222-2222-2222-2222-222222222202', 'architecture', '架构'),
  ('22222222-2222-2222-2222-222222222203', 'content-governance', '内容治理'),
  ('22222222-2222-2222-2222-222222222204', 'vue3', 'Vue3'),
  ('22222222-2222-2222-2222-222222222205', 'seo', 'SEO'),
  ('22222222-2222-2222-2222-222222222206', 'cache', '缓存'),
  ('22222222-2222-2222-2222-222222222207', 'postgresql', 'PostgreSQL'),
  ('22222222-2222-2222-2222-222222222208', 'redis', 'Redis'),
  ('22222222-2222-2222-2222-222222222209', 'full-text-search', '全文搜索'),
  ('22222222-2222-2222-2222-222222222210', 'information-architecture', '信息架构'),
  ('22222222-2222-2222-2222-222222222211', 'markdown', 'Markdown'),
  ('22222222-2222-2222-2222-222222222212', 'rss', 'RSS'),
  ('22222222-2222-2222-2222-222222222213', 'comments', '评论'),
  ('22222222-2222-2222-2222-222222222214', 'submission', '投稿'),
  ('22222222-2222-2222-2222-222222222215', 'message', '站内信'),
  ('22222222-2222-2222-2222-222222222216', 'media', '媒体库'),
  ('22222222-2222-2222-2222-222222222217', 'navigation', '导航'),
  ('22222222-2222-2222-2222-222222222218', 'account', '账号'),
  ('22222222-2222-2222-2222-222222222219', 'workflow', '工作流'),
  ('22222222-2222-2222-2222-222222222220', 'operations', '运营')
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO posts (
  id, slug, title, summary, content, status, source, category_id, author_name,
  cover_image, reading_time, view_count, like_count, dislike_count, comment_count, published_at
) VALUES
  (
    '33333333-3333-3333-3333-333333333301',
    'blog-system-design',
    '如何设计一个内容长期增长的博客系统',
    '博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。',
    '一个现代化博客系统需要从内容资产的生命周期开始设计。文章不是一次性页面，而是会被修改、引用、搜索、迁移和长期展示的结构化内容。如果早期只实现简单的增删改查，后期补版本历史、重定向、搜索索引和订阅推送时，往往会遇到数据模型不够稳定的问题。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111111',
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
    '33333333-3333-3333-3333-333333333302',
    'vue3-content-site-cache-seo',
    'Vue3 内容站的缓存与 SEO 边界',
    '客户端渲染、接口缓存和服务端 meta 需要明确边界，避免前期开发轻松、后期收录困难。',
    'Vue3 内容站可以保持前端开发效率，同时通过 Go 输出基础 HTML、meta 和结构化数据处理文章页 SEO。缓存策略、页面更新和接口数据需要从第一版开始定义边界。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111115',
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
    '33333333-3333-3333-3333-333333333303',
    'postgres-redis-blog-boundary',
    'Redis 和 PostgreSQL 在博客中的分工',
    'PostgreSQL 保存事实并承担全文搜索，Redis 负责热点读取、会话、限流和异步任务协调。',
    '个人博客早期没有必要引入专用搜索中间件。PostgreSQL 的 tsvector 和 GIN 索引足以覆盖标题、摘要、正文和标签搜索，Redis 则用于热点数据、会话、限流和异步任务协调。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111112',
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
    '33333333-3333-3333-3333-333333333304',
    'home-to-article-information-architecture',
    '从首页到文章页，博客的信息架构怎么排',
    '首页负责分发，文章页负责阅读，归档页负责找回内容。三者的密度和交互应该完全不同。',
    '首页需要帮助读者发现内容，文章页需要让读者沉浸阅读，归档页需要承担高效率查找。把这些页面设计成同一种卡片流，会让每个场景都不够顺手。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111113',
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
    '33333333-3333-3333-3333-333333333305',
    'post-version-history',
    '为什么博客后台需要文章版本历史',
    '版本记录不是复杂功能，而是内容资产的基本保险。',
    '文章会被持续修订，后台需要记录版本历史、修改人、变更摘要和回滚能力。对于长期运营的博客，这可以避免误操作带来的内容资产损失。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111116',
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
    '33333333-3333-3333-3333-333333333306',
    'markdown-writing-experience',
    '把 Markdown 写作体验做到顺手',
    '编辑器、预览、封面和 SEO 字段要服务写作流程，而不是把作者困在表单里。',
    'Markdown 编辑器需要稳定的草稿保存、预览、图片插入、代码块处理和 SEO 字段编辑。写作体验越顺手，内容生产越稳定。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111116',
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
    '33333333-3333-3333-3333-333333333307',
    'rss-stable-distribution',
    'RSS 是博客最稳定的分发入口',
    '早期先保留 RSS，把邮件订阅和推送能力放到内容稳定增长后再开启。',
    'RSS 不需要复杂用户系统，也不依赖第三方平台算法。对独立博客来说，它是低成本、低维护的稳定分发入口。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111114',
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
    '33333333-3333-3333-3333-333333333308',
    'comments-system-design',
    '用户评论系统应该怎么设计',
    '登录用户评论、审核、举报、站内信和禁言机制需要从同一套权限边界出发。',
    '评论系统的重点不是输入框，而是审核、通知、举报、禁言、删除记录和内容安全。默认待审核可以降低早期运营风险。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111111',
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
    '33333333-3333-3333-3333-333333333309',
    'submission-review-workflow',
    '登录用户投稿到审核发布的完整闭环',
    '普通用户可以投稿，但不能直接发布；审核通过后再进入正式文章列表。',
    '投稿流程需要草稿、提交审核、退回修改、拒绝、通过发布和审核意见。审核结果还应同步到站内信和我的投稿列表。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111116',
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
    '33333333-3333-3333-3333-333333333310',
    'station-message-design',
    '站内信适合承载哪些系统通知',
    '审核结果、评论回复、系统公告和管理员定向消息都适合进入站内信。',
    '站内信需要未读、已读、归档、删除和管理员发送能力。它承载的是站内重要事件，不应替代即时聊天。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111114',
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
    '33333333-3333-3333-3333-333333333311',
    'media-library-storage',
    '媒体库不只是上传图片',
    '图片、附件、Alt 文本、引用关系和删除影响检查决定了媒体库是否可长期维护。',
    '媒体库需要记录资源元数据、生成缩略图、校验类型和大小，并在删除前检查引用关系。生产环境建议使用 S3 兼容对象存储。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111112',
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
    '33333333-3333-3333-3333-333333333312',
    'admin-navigation-settings',
    '后台导航和系统设置怎么拆页面',
    '导航、设置、统计和媒体库都是独立运营能力，不能只停留在入口卡片。',
    '后台页面需要按运营任务拆分，而不是把所有能力堆到概览。导航管理和系统设置尤其需要独立表单和变更校验。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111113',
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
    '33333333-3333-3333-3333-333333333313',
    'postgres-full-text-search',
    '用 PostgreSQL 做博客全文搜索够不够',
    '对于早期博客，tsvector、GIN 索引和少量模糊匹配已经足够支撑标题、摘要、正文和标签搜索。',
    '专用搜索服务会增加部署和运维成本。除非搜索规模和分词需求明显超过 PostgreSQL 能力，否则可以先用数据库内建搜索能力。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111112',
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
    '33333333-3333-3333-3333-333333333314',
    'footbar-runtime-component',
    '把站点 Footbar 做成公共组件',
    '运行时间、版权、访问统计和站点入口应由公共组件统一维护。',
    '底部栏会出现在多个页面，适合抽成组件。运行时间从固定日期动态计算，社交入口应保持克制并靠近文字。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111115',
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
    '33333333-3333-3333-3333-333333333315',
    'account-settings-security',
    '个人中心账号设置要包含哪些能力',
    '资料、头像、密码、登录设备、数据导出和账号注销都需要明确入口。',
    '账号设置不是一个简单昵称表单。它涉及安全、隐私、会话管理、数据导出和账号注销，需要和个人中心其他页面保持一致。',
    'published',
    'admin',
    '11111111-1111-1111-1111-111111111111',
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
