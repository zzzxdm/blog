<script setup lang="ts">
import { RouterLink, useRoute } from "vue-router";

const route = useRoute();

function active(path: string) {
  return route.path === path;
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>个人中心</h1>
        <p>管理资料、评论、收藏、投稿和站内信。</p>
      </div>
    </section>

    <section class="account-layout">
      <aside class="panel">
        <div class="profile-card">
          <div class="profile-hero">
            <span class="avatar">林</span>
            <div>
              <strong>林一</strong>
              <div class="meta-row">
                <span>已验证邮箱</span>
                <span>注册于 2026-06-12</span>
              </div>
            </div>
          </div>
          <nav class="account-nav" aria-label="个人中心导航">
            <RouterLink :class="{ active: active('/account') }" to="/account">概览</RouterLink>
            <RouterLink :class="{ active: active('/account/comments') }" to="/account/comments">我的评论</RouterLink>
            <RouterLink :class="{ active: active('/account/bookmarks') }" to="/account/bookmarks">我的收藏</RouterLink>
            <RouterLink :class="{ active: active('/account/submissions') }" to="/account/submissions">我的投稿</RouterLink>
            <RouterLink :class="{ active: active('/account/messages') }" to="/account/messages">站内信</RouterLink>
            <RouterLink :class="{ active: active('/account/settings') }" to="/account/settings">账号设置</RouterLink>
          </nav>
        </div>
      </aside>

      <div class="settings-stack">
        <section class="stats-grid" aria-label="用户统计">
          <div class="stat-card">
            <span>评论</span>
            <strong>42</strong>
          </div>
          <div class="stat-card">
            <span>收藏</span>
            <strong>18</strong>
          </div>
          <div class="stat-card">
            <span>投稿</span>
            <strong>5</strong>
          </div>
          <div class="stat-card">
            <span>未读站内信</span>
            <strong>3</strong>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>我的投稿</h2>
            <RouterLink class="button-secondary" to="/submit">继续投稿</RouterLink>
          </div>
          <div class="timeline">
            <article class="timeline-item">
              <strong>用户评论系统应该怎么设计</strong>
              <p>提交后进入审核队列，编辑正在检查结构和引用来源。</p>
              <div class="meta-row">
                <span class="status review">待审核</span>
                <span>今天 15:42</span>
              </div>
            </article>
            <article class="timeline-item">
              <strong>把 Markdown 写作体验做到顺手</strong>
              <p>审核通过后已发布到公开文章列表。</p>
              <div class="meta-row">
                <span class="status published">已发布</span>
                <RouterLink to="/posts/markdown-writing-experience">查看文章</RouterLink>
              </div>
            </article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>我的评论</h2>
            <RouterLink class="button-secondary" to="/posts/blog-system-design">去评论</RouterLink>
          </div>
          <div class="timeline">
            <article class="timeline-item">
              <strong>审核结果是否会同步提醒到个人中心？</strong>
              <p>评论于《如何设计一个内容长期增长的博客系统》</p>
              <div class="meta-row">
                <span class="status review">待审核</span>
                <span>刚刚</span>
              </div>
            </article>
            <article class="timeline-item">
              <strong>版本历史对于多人协作很重要，个人博客其实也需要。</strong>
              <p>评论于《为什么博客后台需要文章版本历史》</p>
              <div class="meta-row">
                <span class="status published">已通过</span>
                <span>18 次点赞</span>
              </div>
            </article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>我的收藏</h2>
          </div>
          <ul class="link-list">
            <li>
              <RouterLink to="/posts/vue3-content-site-cache-seo">
                <strong>Vue3 内容站的缓存与 SEO 边界</strong>
                <span>工程实践 · 8 分钟阅读</span>
              </RouterLink>
            </li>
            <li>
              <RouterLink to="/posts/postgres-full-text-search">
                <strong>用 PostgreSQL 做博客全文搜索够不够</strong>
                <span>架构 · 12 分钟阅读</span>
              </RouterLink>
            </li>
          </ul>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>站内信</h2>
          </div>
          <div class="timeline">
            <article class="timeline-item">
              <strong>管理员 回复了你的评论</strong>
              <p>“审核结果会同步到站内信和我的投稿列表。”</p>
              <div class="meta-row">
                <span>2 小时前</span>
                <RouterLink to="/posts/submission-review-workflow">查看上下文</RouterLink>
              </div>
            </article>
            <article class="timeline-item">
              <strong>你的评论已通过审核</strong>
              <p>评论现在已展示在文章详情页。</p>
              <div class="meta-row">
                <span>昨天</span>
              </div>
            </article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>账号设置</h2>
          </div>
          <div class="settings-grid">
            <div class="field">
              <label for="display-name">昵称</label>
              <input class="input" id="display-name" value="林一">
            </div>
            <div class="field">
              <label for="email">邮箱</label>
              <input class="input" id="email" value="linyi@example.com">
            </div>
            <div class="field">
              <label for="password">修改密码</label>
              <input class="input" id="password" type="password" value="password">
            </div>
            <div class="field">
              <label for="notification">通知偏好</label>
              <select class="input" id="notification">
                <option>站内信 + 邮件提醒</option>
                <option>仅站内信</option>
                <option>关闭非必要提醒</option>
              </select>
            </div>
          </div>
        </section>
      </div>
    </section>
  </main>
</template>
