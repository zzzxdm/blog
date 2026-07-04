<script setup lang="ts">
import AccountLayout from "../../components/AccountLayout.vue";
</script>

<template>
  <AccountLayout title="站内信" description="查看投稿审核结果、评论回复、账号提醒和管理员发送的消息。">
    <template #actions>
      <button class="button-secondary" type="button">全部已读</button>
    </template>

    <section class="stats-grid" aria-label="站内信统计">
      <div class="stat-card"><span>未读</span><strong>3</strong></div>
      <div class="stat-card"><span>审核消息</span><strong>2</strong></div>
      <div class="stat-card"><span>管理员消息</span><strong>5</strong></div>
      <div class="stat-card"><span>已归档</span><strong>16</strong></div>
    </section>

    <section class="message-layout">
      <div class="panel">
        <div class="panel-title"><h2>收件箱</h2><button class="button-secondary" type="button">筛选</button></div>
        <div class="message-filterbar" aria-label="消息筛选">
          <a class="active" href="#all">全部</a>
          <a href="#unread">未读</a>
          <a href="#review">审核</a>
          <a href="#system">系统</a>
          <a href="#admin">管理员</a>
        </div>
        <div class="message-list">
          <a class="message-item unread active" href="#detail"><div class="meta-row"><span class="status review">审核</span><span>刚刚</span></div><strong>你的投稿已退回修改</strong><p>《如何写一篇可维护的技术文章》需要补充摘要和代码示例。</p></a>
          <a class="message-item unread" href="#detail"><div class="meta-row"><span class="status published">评论</span><span>2 小时前</span></div><strong>管理员 回复了你的评论</strong><p>审核结果会同步到站内信和我的投稿列表。</p></a>
          <a class="message-item unread" href="#detail"><div class="meta-row"><span class="status muted">公告</span><span>今天 09:30</span></div><strong>本周将维护媒体库上传服务</strong><p>维护期间图片上传会短暂不可用，阅读不受影响。</p></a>
          <a class="message-item" href="#detail"><div class="meta-row"><span class="status published">审核</span><span>昨天</span></div><strong>你的评论已通过审核</strong><p>评论现在已展示在文章详情页。</p></a>
          <a class="message-item" href="#detail"><div class="meta-row"><span class="status draft">账号</span><span>2026-07-01</span></div><strong>邮箱验证成功</strong><p>你的账号已经完成邮箱验证。</p></a>
        </div>
      </div>

      <article class="panel message-detail" id="detail">
        <div class="meta-row"><span class="status review">投稿审核</span><span>管理员</span><span>2026-07-04 15:42</span></div>
        <div class="message-body">
          <h2>你的投稿已退回修改</h2>
          <p>《如何写一篇可维护的技术文章》暂未通过审核。编辑已经填写了修改意见，补充后可以重新提交。</p>
          <blockquote>摘要过短，建议明确文章解决的问题；正文中有一段代码没有解释上下文；封面图缺少 alt 文本。</blockquote>
          <p>这条消息会同步显示在“我的投稿”列表中，便于你回到投稿草稿继续处理。</p>
        </div>
        <div class="header-actions">
          <RouterLink class="button" to="/account/submissions">查看投稿</RouterLink>
          <button class="button-secondary" type="button">标记已读</button>
          <button class="button-secondary" type="button">归档</button>
        </div>
      </article>
    </section>
  </AccountLayout>
</template>
