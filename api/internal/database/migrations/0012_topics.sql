CREATE TABLE IF NOT EXISTS topics (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
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
    '44444444-4444-4444-4444-444444444401',
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
    '44444444-4444-4444-4444-444444444402',
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
    '44444444-4444-4444-4444-444444444403',
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
    '44444444-4444-4444-4444-444444444404',
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
