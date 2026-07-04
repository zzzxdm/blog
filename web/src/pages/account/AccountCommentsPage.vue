<script setup lang="ts">
import AccountLayout from "../../components/AccountLayout.vue";
</script>

<template>
  <AccountLayout title="我的评论" description="查看评论审核状态、回复记录和已通过评论。">
    <template #actions>
      <RouterLink class="button-secondary" to="/posts/blog-system-design">去评论</RouterLink>
    </template>

    <section class="stats-grid" aria-label="评论统计">
      <div class="stat-card"><span>全部评论</span><strong>42</strong></div>
      <div class="stat-card"><span>待审核</span><strong>1</strong></div>
      <div class="stat-card"><span>获赞</span><strong>126</strong></div>
      <div class="stat-card"><span>被回复</span><strong>8</strong></div>
    </section>

    <section class="table-panel">
      <form class="table-toolbar">
        <input class="input" type="search" placeholder="搜索评论内容或文章标题" aria-label="搜索我的评论">
        <select class="input" aria-label="评论状态"><option>全部状态</option><option>待审核</option><option>已通过</option><option>已删除</option></select>
        <select class="input" aria-label="排序"><option>最近评论</option><option>获赞最多</option><option>被回复优先</option></select>
      </form>

      <table>
        <thead>
          <tr><th>评论</th><th>文章</th><th>状态</th><th>互动</th><th>时间</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr>
            <td><strong>如果后续支持站内信，审核结果是否会同步提醒到个人中心？</strong><div class="meta-row"><span>回复主评论 #128</span></div></td>
            <td>如何设计一个内容长期增长的博客系统</td>
            <td><span class="status review">待审核</span></td>
            <td>0 赞</td>
            <td>刚刚</td>
            <td><button class="button-secondary" type="button">编辑</button></td>
          </tr>
          <tr>
            <td><strong>版本历史对于多人协作很重要，个人博客其实也需要。</strong><div class="meta-row"><span>管理员已回复</span></div></td>
            <td>为什么博客后台需要文章版本历史</td>
            <td><span class="status published">已通过</span></td>
            <td>18 赞 · 1 回复</td>
            <td>昨天</td>
            <td><RouterLink class="button-secondary" to="/posts/post-version-history">查看</RouterLink></td>
          </tr>
          <tr>
            <td><strong>PostgreSQL 全文搜索足够覆盖多数个人博客场景。</strong></td>
            <td>用 PostgreSQL 做博客全文搜索够不够</td>
            <td><span class="status published">已通过</span></td>
            <td>9 赞</td>
            <td>2026-07-01</td>
            <td><RouterLink class="button-secondary" to="/posts/postgres-full-text-search">查看</RouterLink></td>
          </tr>
        </tbody>
      </table>
    </section>
  </AccountLayout>
</template>
