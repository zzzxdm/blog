<script setup lang="ts">
import AccountLayout from "../../components/AccountLayout.vue";
</script>

<template>
  <AccountLayout title="账号设置" description="管理公开资料、登录安全、消息偏好和账号数据。">
    <template #actions>
      <button class="button" type="button">保存设置</button>
    </template>

    <section class="stats-grid" aria-label="账号状态">
      <div class="stat-card"><span>安全等级</span><strong>高</strong></div>
      <div class="stat-card"><span>登录设备</span><strong>2</strong></div>
      <div class="stat-card"><span>公开文章</span><strong>3</strong></div>
      <div class="stat-card"><span>资料完整度</span><strong>86%</strong></div>
    </section>

    <section class="settings-grid">
      <section class="panel">
        <div class="panel-title"><h2>公开资料</h2></div>
        <div class="settings-stack">
          <div class="profile-hero"><span class="avatar">林</span><div class="header-actions"><button class="button-secondary" type="button">更换头像</button><button class="button-secondary" type="button">移除</button></div></div>
          <div class="field"><label for="display-name">昵称</label><input class="input" id="display-name" value="林一"></div>
          <div class="field"><label for="username">用户名</label><input class="input" id="username" value="linyi"></div>
          <div class="field"><label for="bio">个人简介</label><textarea class="input" id="bio">关注内容产品、工程实践和长期写作。</textarea></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>登录安全</h2></div>
        <div class="settings-stack">
          <div class="field"><label for="email">登录邮箱</label><input class="input" id="email" value="linyi@example.com"></div>
          <div class="field"><label for="password">修改密码</label><input class="input" id="password" type="password" placeholder="输入新密码"></div>
          <div class="setting-row"><div><strong>两步验证</strong><div class="meta-row"><span>登录时需要邮箱验证码</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>异常登录提醒</strong><div class="meta-row"><span>新设备登录时发送站内信</span></div></div><span class="switch"></span></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>消息偏好</h2><RouterLink class="button-secondary" to="/account/messages">查看站内信</RouterLink></div>
        <div class="settings-stack">
          <div class="setting-row"><div><strong>投稿审核结果</strong><div class="meta-row"><span>通过、退回、拒绝都会发送站内信</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>评论回复</strong><div class="meta-row"><span>有人回复或点赞你的评论时提醒</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>站点公告</strong><div class="meta-row"><span>接收维护、规则和功能更新</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>邮件提醒</strong><div class="meta-row"><span>重要账号事件同步发送邮件</span></div></div><span class="switch off"></span></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>隐私与展示</h2></div>
        <div class="settings-stack">
          <div class="setting-row"><div><strong>展示个人主页</strong><div class="meta-row"><span>公开显示投稿、简介和评论摘要</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>展示收藏列表</strong><div class="meta-row"><span>其他用户可以查看公开收藏</span></div></div><span class="switch off"></span></div>
          <div class="field"><label for="profile-url">个人主页地址</label><input class="input" id="profile-url" value="https://blog.example.com/authors/linyi"></div>
          <div class="field"><label for="timezone">时区</label><select class="input" id="timezone"><option>Asia/Shanghai</option><option>UTC</option><option>Asia/Tokyo</option></select></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>登录设备</h2></div>
        <div class="timeline">
          <article class="timeline-item"><strong>Windows Chrome</strong><p>当前设备，上海，今天 16:20</p><div class="meta-row"><span class="status published">已信任</span></div></article>
          <article class="timeline-item"><strong>iPhone Safari</strong><p>上次登录于昨天 22:08</p><div class="meta-row"><span>上海</span><button class="button-secondary" type="button">移除设备</button></div></article>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>账号数据</h2></div>
        <div class="settings-stack">
          <div class="review-note"><strong>注销前请先导出数据</strong><p>账号注销后，未发布草稿、收藏和个人设置将不可恢复；已发布内容会按站点规则保留署名记录。</p></div>
          <button class="button-secondary" type="button">导出我的数据</button>
          <button class="button-secondary" type="button">申请注销账号</button>
        </div>
      </section>
    </section>
  </AccountLayout>
</template>
