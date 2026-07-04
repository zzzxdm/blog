<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="投稿审核" description="审核登录用户提交的文章，确认质量后发布到正式内容库。" mobile-title="投稿审核" primary-action="通过发布">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">退回修改</button>
        <button class="button" type="button">通过并发布</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="投稿统计">
      <div class="stat-card"><span>待审核</span><strong>12</strong></div>
      <div class="stat-card"><span>今日提交</span><strong>5</strong></div>
      <div class="stat-card"><span>退回修改</span><strong>8</strong></div>
      <div class="stat-card"><span>本月发布</span><strong>19</strong></div>
    </section>

    <section class="admin-grid-2">
      <div class="settings-stack">
        <section class="table-panel">
          <form class="table-toolbar">
            <input class="input" type="search" placeholder="搜索投稿标题、投稿人、标签" aria-label="搜索投稿">
            <select class="input" aria-label="投稿状态">
              <option>全部状态</option>
              <option>待审核</option>
              <option>退回修改</option>
              <option>已发布</option>
            </select>
            <select class="input" aria-label="排序">
              <option>最近提交</option>
              <option>高风险优先</option>
              <option>高质量优先</option>
            </select>
          </form>

          <table>
            <thead>
              <tr>
                <th>投稿</th>
                <th>投稿人</th>
                <th>状态</th>
                <th>风险</th>
                <th>提交时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <strong>用户评论系统应该怎么设计</strong>
                  <div class="meta-row"><span>用户系统</span><span>2,480 字</span></div>
                </td>
                <td>林一<div class="meta-row"><span>42 条评论</span><span>3 篇已发布</span></div></td>
                <td><span class="status review">待审核</span></td>
                <td>低</td>
                <td>今天 15:42</td>
                <td><button class="button-secondary" type="button">查看</button></td>
              </tr>
              <tr>
                <td>
                  <strong>开放投稿入口后如何做反垃圾</strong>
                  <div class="meta-row"><span>内容治理</span><span>包含 2 个外链</span></div>
                </td>
                <td>陈默<div class="meta-row"><span>已验证邮箱</span></div></td>
                <td><span class="status review">待审核</span></td>
                <td>中</td>
                <td>今天 10:24</td>
                <td><button class="button-secondary" type="button">查看</button></td>
              </tr>
              <tr>
                <td>
                  <strong>如何写一篇可维护的技术文章</strong>
                  <div class="meta-row"><span>写作工作流</span><span>二次提交</span></div>
                </td>
                <td>林一<div class="meta-row"><span>退回后已修改</span></div></td>
                <td><span class="status review">待复审</span></td>
                <td>低</td>
                <td>昨天 18:08</td>
                <td><button class="button-secondary" type="button">查看</button></td>
              </tr>
            </tbody>
          </table>
        </section>

        <section class="editor-panel">
          <div class="editor-toolbar">
            <div class="meta-row">
              <span class="tag">投稿预览</span>
              <span>用户系统</span>
              <span>2,480 字</span>
            </div>
            <button class="button-secondary" type="button">编辑内容</button>
          </div>
          <article class="preview-area" style="min-height: 420px;">
            <h1>用户评论系统应该怎么设计</h1>
            <p>登录用户评论、审核、举报、通知和禁言机制，是开放内容站点的基础能力。</p>
            <h2>评论不是简单留言</h2>
            <p>评论需要和用户系统、通知系统、反垃圾策略一起设计。一个可维护的评论系统，需要把用户状态、评论状态、审核流程和通知机制放在一起考虑。</p>
            <h2>审核状态</h2>
            <p>建议至少支持待审核、已通过、已拒绝、垃圾评论和已删除状态。用户可以看到自己的待审核评论，管理员可以查看所有状态。</p>
          </article>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>审核动作</h2>
            <span class="status review">待审核</span>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="review-note">审核意见</label>
              <textarea class="input" id="review-note">内容结构清楚，可以发布。建议把标题调整为“登录用户评论系统应该怎么设计”，并补一段评论举报流程。</textarea>
            </div>
            <button class="button" type="button">通过并发布</button>
            <button class="button-secondary" type="button">退回修改</button>
            <button class="button-secondary" type="button">拒绝投稿</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>投稿人</h2>
          </div>
          <div class="profile-hero">
            <span class="avatar">林</span>
            <div>
              <strong>林一</strong>
              <div class="meta-row"><span>已验证邮箱</span><span>注册用户</span></div>
            </div>
          </div>
          <div class="settings-stack" style="margin-top: 16px;">
            <div class="setting-row">
              <div>
                <strong>历史投稿</strong>
                <div class="meta-row"><span>3 篇已发布，1 篇退回</span></div>
              </div>
            </div>
            <div class="setting-row">
              <div>
                <strong>评论质量</strong>
                <div class="meta-row"><span>42 条评论，126 次获赞</span></div>
              </div>
            </div>
            <button class="button-secondary" type="button">升级为作者</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>发布设置</h2>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="slug">Slug</label><input class="input" id="slug" value="user-comment-system-design"></div>
            <div class="field"><label for="category">分类</label><select class="input" id="category"><option>用户系统</option><option>内容治理</option><option>工程实践</option></select></div>
            <div class="field"><label for="publish-time">发布时间</label><input class="input" id="publish-time" type="datetime-local" value="2026-07-04T20:00"></div>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
