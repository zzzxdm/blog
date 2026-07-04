<script setup lang="ts">
import { computed, onMounted, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";

import { usePostsStore } from "../stores/posts";

const route = useRoute();
const router = useRouter();
const posts = usePostsStore();

const post = computed(() => posts.current);
const avatarText = computed(() => post.value?.authorName.slice(0, 1) || "管");

function load() {
  const slug = String(route.params.slug || "");
  if (slug) {
    void posts.loadBySlug(slug);
  }
}

function back() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  void router.push("/archive");
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}

onMounted(load);
watch(() => route.params.slug, load);
</script>

<template>
  <main class="article-shell">
    <p v-if="posts.loading" class="muted">正在加载文章...</p>
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>

    <template v-else-if="post">
      <article>
        <header class="article-hero">
          <div class="article-breadcrumb-row">
            <button class="button-secondary" type="button" @click="back">← 返回</button>
            <nav class="breadcrumb" aria-label="当前位置">
              <RouterLink to="/">首页</RouterLink>
              <span class="breadcrumb-separator">/</span>
              <RouterLink :to="`/archive?category=${encodeURIComponent(post.category)}`">{{ post.category }}</RouterLink>
              <span class="breadcrumb-separator">/</span>
              <span>{{ post.title }}</span>
            </nav>
          </div>
          <div class="meta-row">
            <span class="tag">{{ post.category }}</span>
            <span>{{ post.readingTime }} 分钟阅读</span>
            <span>{{ formatDate(post.publishedAt) }}</span>
          </div>
          <h1>{{ post.title }}</h1>
          <p class="dek">{{ post.summary }}</p>
          <div class="author-row">
            <span class="avatar">{{ avatarText }}</span>
            <div>
              <strong>{{ post.authorName }}</strong>
              <div class="meta-row">
                <span>{{ post.tags[0] || "系统设计" }}</span>
                <span>{{ formatNumber(post.viewCount) }} 次阅读</span>
                <span>{{ formatNumber(post.likeCount) }} 次赞</span>
                <span>{{ formatNumber(post.dislikeCount) }} 次踩</span>
                <span>{{ formatNumber(post.commentCount) }} 条评论</span>
              </div>
            </div>
          </div>
        </header>

        <figure class="article-cover">
          <img :src="post.coverImage" :alt="post.title">
        </figure>

        <section class="article-body">
          <p>{{ post.content }}</p>

          <h2 id="content-model">内容模型先于页面</h2>
          <p>文章需要拥有稳定的 slug、可维护的分类标签、SEO 元数据、封面图、摘要、阅读时长、发布时间和更新时间。内容模型稳定后，前台页面、搜索索引和 RSS 输出都可以从同一份数据生成。</p>
          <blockquote>内容系统的核心不是页面，而是可被长期复用、迁移和分发的数据。</blockquote>

          <h2 id="workflow">发布流程需要留出空间</h2>
          <p>成熟博客通常支持草稿、预览、审核、定时发布和版本历史。个人博客可以简化审批流程，但不应省略草稿、预览和版本回滚。</p>

          <div class="code-block">
            <div class="code-header">
              <span>post-status.ts</span>
              <button class="button-secondary" type="button">复制</button>
            </div>
            <pre><code>type PostStatus = "draft" | "submitted" | "scheduled" | "published" | "archived";

interface Post {
  title: string;
  slug: string;
  status: PostStatus;
  publishedAt?: Date;
}</code></pre>
          </div>

          <h2 id="reading">阅读体验要克制</h2>
          <p>文章页不需要复杂装饰。合适的行宽、稳定的目录、清晰的代码块、图片懒加载和足够好的移动端排版，比炫目的视觉元素更重要。</p>

          <h2 id="operations">运营能力决定长期价值</h2>
          <p>搜索词统计、热门文章、来源渠道和评论反馈可以帮助作者判断内容是否有效。对于持续写作来说，数据反馈是内容迭代的重要基础。</p>
        </section>

        <section class="article-feedback" aria-label="文章反馈">
          <div>
            <strong>文章反馈</strong>
            <div class="meta-row">
              <span>{{ formatNumber(post.likeCount) }} 次赞</span>
              <span>{{ formatNumber(post.dislikeCount) }} 次踩</span>
              <span>已收藏 34 次</span>
            </div>
          </div>
          <div class="reaction-group">
            <button class="reaction-button active" type="button" aria-label="点赞文章">
              <span class="reaction-symbol">↑</span>
              <span>{{ formatNumber(post.likeCount) }}</span>
            </button>
            <button class="reaction-button" type="button" aria-label="点踩文章">
              <span class="reaction-symbol">↓</span>
              <span>{{ formatNumber(post.dislikeCount) }}</span>
            </button>
            <button class="button-secondary" type="button">收藏</button>
          </div>
        </section>

        <section class="comments" aria-label="评论">
          <div class="section-heading">
            <div>
              <h2>评论</h2>
              <p>{{ post.commentCount }} 条讨论，评论提交后进入审核队列。</p>
            </div>
            <button class="button-secondary" type="button">按时间排序</button>
          </div>
          <div class="comment-box">
            <div class="author-row">
              <span class="avatar">管</span>
              <div>
                <strong>管理员</strong>
                <div class="meta-row">
                  <span>已登录</span>
                  <RouterLink to="/account/comments">查看我的评论</RouterLink>
                </div>
              </div>
            </div>
            <textarea placeholder="写下你的想法，支持 Markdown 基础语法"></textarea>
            <div class="meta-row">
              <button class="button" type="button">提交评论</button>
              <span>评论提交后进入审核队列。</span>
            </div>
          </div>

          <div class="comment-list">
            <article class="comment-item">
              <div class="comment-head">
                <div class="author-row">
                  <span class="avatar">陈</span>
                  <div>
                    <strong>陈默</strong>
                    <div class="meta-row">
                      <span>2 小时前</span>
                      <span>产品设计师</span>
                    </div>
                  </div>
                </div>
                <span class="status published">已通过</span>
              </div>
              <p>文章里提到“内容模型先于页面”很关键。很多博客后期难维护，就是因为一开始把文章当页面模板来处理了。</p>
              <div class="comment-actions">
                <button type="button">点赞 18</button>
                <button type="button">回复</button>
                <button type="button">举报</button>
              </div>
            </article>

            <article class="comment-item reply">
              <div class="comment-head">
                <div class="author-row">
                  <span class="avatar">管</span>
                  <div>
                    <strong>管理员</strong>
                    <div class="meta-row">
                      <span>作者</span>
                      <span>1 小时前</span>
                    </div>
                  </div>
                </div>
                <span class="tag rust">作者回复</span>
              </div>
              <p>是的，所以我会优先把 slug、SEO、状态、版本历史这些字段纳入第一版数据模型。</p>
              <div class="comment-actions">
                <button type="button">点赞 9</button>
                <button type="button">回复</button>
              </div>
            </article>

            <article class="comment-item">
              <div class="comment-head">
                <div class="author-row">
                  <span class="avatar">林</span>
                  <div>
                    <strong>林一</strong>
                    <div class="meta-row">
                      <span>刚刚</span>
                      <span>我的评论</span>
                    </div>
                  </div>
                </div>
                <span class="status review">待审核</span>
              </div>
              <p>如果后续支持站内信，审核结果是否会同步提醒到个人中心？</p>
              <div class="comment-actions">
                <button type="button">编辑</button>
                <button type="button">删除</button>
              </div>
            </article>
          </div>
        </section>
      </article>

      <aside class="toc" aria-label="文章目录">
        <section class="panel">
          <div class="panel-title">
            <h2>目录</h2>
          </div>
          <nav>
            <a class="active" href="#content-model">内容模型先于页面</a>
            <a href="#workflow">发布流程需要留出空间</a>
            <a href="#reading">阅读体验要克制</a>
            <a href="#operations">运营能力决定长期价值</a>
          </nav>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>作者</h2>
          </div>
          <div class="author-row">
            <span class="avatar">{{ avatarText }}</span>
            <div>
              <strong>{{ post.authorName }}</strong>
              <div class="meta-row">
                <span>128 篇文章</span>
                <span>24 个专题</span>
              </div>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>相关文章</h2>
          </div>
          <ul class="link-list">
            <li>
              <RouterLink to="/posts/postgres-redis-blog-boundary">
                <strong>Redis 和 PostgreSQL 在博客中的分工</strong>
                <span>架构 · 14 分钟</span>
              </RouterLink>
            </li>
            <li>
              <RouterLink to="/posts/post-version-history">
                <strong>为什么博客后台需要文章版本历史</strong>
                <span>内容治理 · 6 分钟</span>
              </RouterLink>
            </li>
          </ul>
        </section>
      </aside>
    </template>
  </main>
</template>
