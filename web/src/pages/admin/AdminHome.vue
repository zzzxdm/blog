<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="后台概览" description="查看待办、最近操作、内容状态和系统运行情况。" mobile-title="后台概览" primary-action="新建">
    <template #actions>
      <div class="header-actions">
        <RouterLink class="button-secondary" to="/">查看站点</RouterLink>
        <RouterLink class="button" to="/admin/editor">新建文章</RouterLink>
      </div>
    </template>

    <section class="stats-grid" aria-label="后台核心指标">
      <div class="stat-card">
        <span>已发布文章</span>
        <strong>96</strong>
      </div>
      <div class="stat-card">
        <span>待审核投稿</span>
        <strong>12</strong>
      </div>
      <div class="stat-card">
        <span>待审评论</span>
        <strong>23</strong>
      </div>
      <div class="stat-card">
        <span>本月阅读</span>
        <strong>16.8k</strong>
      </div>
    </section>

    <section class="admin-grid-2">
      <section class="table-panel">
        <div class="panel-title" style="padding: 16px 16px 0;">
          <h2>待办事项</h2>
          <RouterLink class="button-secondary" to="/admin/submissions">处理投稿</RouterLink>
        </div>
        <table>
          <thead>
            <tr>
              <th>事项</th>
              <th>来源</th>
              <th>状态</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>用户评论系统应该怎么设计</td>
              <td>投稿审核</td>
              <td><span class="status review">待审核</span></td>
              <td>今天 15:42</td>
              <td><RouterLink class="button-secondary" to="/admin/submissions">查看</RouterLink></td>
            </tr>
            <tr>
              <td>推广链接评论被举报 3 次</td>
              <td>评论举报</td>
              <td><span class="status banned">高风险</span></td>
              <td>12 分钟前</td>
              <td><RouterLink class="button-secondary" to="/admin/comments">处理</RouterLink></td>
            </tr>
            <tr>
              <td>媒体库上传服务维护公告</td>
              <td>站内信</td>
              <td><span class="status draft">定时中</span></td>
              <td>今晚 20:00</td>
              <td><RouterLink class="button-secondary" to="/admin/messages">编辑</RouterLink></td>
            </tr>
          </tbody>
        </table>
      </section>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>系统状态</h2>
            <span class="status published">正常</span>
          </div>
          <ul class="link-list">
            <li>
              <strong>PostgreSQL</strong>
              <span>等待本地数据库环境验证迁移</span>
            </li>
            <li>
              <strong>Redis</strong>
              <span>会话、限流和任务队列待接入</span>
            </li>
            <li>
              <strong>媒体存储</strong>
              <span>开发期本地，生产期 S3 / OSS</span>
            </li>
          </ul>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>最近操作</h2>
          </div>
          <div class="timeline">
            <article class="timeline-item">
              <strong>管理员 发布了文章</strong>
              <p>如何设计一个内容长期增长的博客系统</p>
              <div class="meta-row"><span>今天 15:20</span></div>
            </article>
            <article class="timeline-item">
              <strong>管理员 更新了导航</strong>
              <p>增加“投稿”和“我的”入口。</p>
              <div class="meta-row"><span>昨天 18:36</span></div>
            </article>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
