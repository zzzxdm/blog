<script setup lang="ts">
import AccountLayout from "../../components/AccountLayout.vue";
</script>

<template>
  <AccountLayout title="我的投稿" description="查看投稿草稿、审核进度、退回原因和已发布投稿。">
    <template #actions>
      <RouterLink class="button" to="/submit">新建投稿</RouterLink>
    </template>

    <section class="stats-grid" aria-label="投稿统计">
      <div class="stat-card"><span>草稿</span><strong>1</strong></div>
      <div class="stat-card"><span>待审核</span><strong>2</strong></div>
      <div class="stat-card"><span>退回修改</span><strong>1</strong></div>
      <div class="stat-card"><span>已发布</span><strong>3</strong></div>
    </section>

    <section class="table-panel">
      <form class="table-toolbar">
        <input class="input" type="search" placeholder="搜索投稿标题" aria-label="搜索投稿">
        <select class="input" aria-label="投稿状态"><option>全部状态</option><option>草稿</option><option>待审核</option><option>退回修改</option><option>已发布</option></select>
        <select class="input" aria-label="排序"><option>最近更新</option><option>最近提交</option><option>已发布优先</option></select>
      </form>

      <table>
        <thead>
          <tr><th>投稿</th><th>状态</th><th>分类</th><th>提交时间</th><th>审核意见</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr>
            <td><strong>用户评论系统应该怎么设计</strong><div class="meta-row"><span>自动保存于 15:42</span></div></td>
            <td><span class="status draft">草稿</span></td>
            <td>用户系统</td>
            <td>未提交</td>
            <td>未提交审核</td>
            <td><RouterLink class="button-secondary" to="/submit">继续编辑</RouterLink></td>
          </tr>
          <tr>
            <td><strong>开放投稿入口后如何做反垃圾</strong><div class="meta-row"><span>版本 2</span></div></td>
            <td><span class="status review">待审核</span></td>
            <td>内容治理</td>
            <td>今天 10:24</td>
            <td>等待编辑审核</td>
            <td><button class="button-secondary" type="button">查看</button></td>
          </tr>
          <tr>
            <td><strong>如何写一篇可维护的技术文章</strong><div class="meta-row"><span>需要补充图片 alt 文本</span></div></td>
            <td><span class="status rejected">退回</span></td>
            <td>写作工作流</td>
            <td>昨天 18:08</td>
            <td>补充摘要和代码示例</td>
            <td><RouterLink class="button-secondary" to="/submit">修改后提交</RouterLink></td>
          </tr>
          <tr>
            <td><strong>版本历史如何保护内容资产</strong><div class="meta-row"><span>已发布到专题“写作工作流”</span></div></td>
            <td><span class="status published">已发布</span></td>
            <td>内容治理</td>
            <td>2026-06-18</td>
            <td>审核通过</td>
            <td><RouterLink class="button-secondary" to="/posts/post-version-history">查看文章</RouterLink></td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="panel">
      <div class="panel-title"><h2>退回修改说明</h2></div>
      <div class="review-note">
        <strong>《如何写一篇可维护的技术文章》</strong>
        <p>摘要过短，建议明确文章解决的问题；正文中有一段代码没有解释上下文；封面图缺少 alt 文本。修改后可以重新提交。</p>
      </div>
    </section>
  </AccountLayout>
</template>
