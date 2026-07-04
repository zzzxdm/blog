<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="系统设置" description="配置站点基础信息、评论策略、邮件服务、安全和备份。" mobile-title="系统设置" primary-action="保存">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">恢复默认</button>
        <button class="button" type="button">保存设置</button>
      </div>
    </template>

    <section class="settings-grid">
      <section class="panel">
        <div class="panel-title"><h2>站点信息</h2></div>
        <div class="settings-stack">
          <div class="field"><label for="site-name">站点名称</label><input class="input" id="site-name" value="云间笔记"></div>
          <div class="field"><label for="site-desc">站点描述</label><textarea class="input" id="site-desc">技术、产品、工程实践和长期写作的沉淀。</textarea></div>
          <div class="field"><label for="site-url">站点域名</label><input class="input" id="site-url" value="https://blog.example.com"></div>
          <div class="field"><label for="beian">备案信息</label><input class="input" id="beian" value="京ICP备00000000号"></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>主题外观</h2></div>
        <div class="settings-stack">
          <div class="field"><label>主色</label><div class="swatches"><span class="swatch" style="background:#295b4b"></span><span class="swatch" style="background:#b95f2d"></span><span class="swatch" style="background:#e3b45d"></span></div></div>
          <div class="field"><label for="homepage-layout">首页布局</label><select class="input" id="homepage-layout"><option>精选文章 + 最新列表</option><option>专题优先</option><option>极简文章流</option></select></div>
          <div class="setting-row"><div><strong>深色模式</strong><div class="meta-row"><span>允许读者切换浅色和深色</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>显示阅读进度</strong><div class="meta-row"><span>文章页顶部展示阅读进度条</span></div></div><span class="switch"></span></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>评论策略</h2></div>
        <div class="settings-stack">
          <div class="setting-row"><div><strong>开启评论</strong><div class="meta-row"><span>文章默认允许评论</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>仅登录用户评论</strong><div class="meta-row"><span>访客只读，降低垃圾评论</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>评论自动审核</strong><div class="meta-row"><span>新评论默认进入待审核</span></div></div><span class="switch off"></span></div>
          <div class="field"><label for="blocked-words">屏蔽关键词</label><textarea class="input" id="blocked-words">推广, 返利, 博彩</textarea></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>投稿策略</h2></div>
        <div class="settings-stack">
          <div class="setting-row"><div><strong>开放用户投稿</strong><div class="meta-row"><span>登录用户可以提交文章审核</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>投稿必须人工审核</strong><div class="meta-row"><span>通过后才进入公开文章列表</span></div></div><span class="switch"></span></div>
          <div class="field"><label for="submission-limit">投稿频率限制</label><select class="input" id="submission-limit"><option>每天最多 3 篇</option><option>每天最多 1 篇</option><option>每周最多 3 篇</option></select></div>
          <div class="field"><label for="submission-guide">投稿说明</label><textarea class="input" id="submission-guide">投稿需要原创、结构清晰，并补充必要的图片 alt 文本和参考来源。</textarea></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>邮件与 RSS</h2></div>
        <div class="settings-stack">
          <div class="setting-row"><div><strong>开启邮件推送</strong><div class="meta-row"><span>发布文章后可推送邮件</span></div></div><span class="switch off"></span></div>
          <div class="field"><label for="mail-provider">邮件服务</label><select class="input" id="mail-provider"><option>Resend</option><option>SendGrid</option><option>SMTP</option></select></div>
          <div class="field"><label for="from-email">发件邮箱</label><input class="input" id="from-email" value="newsletter@example.com"></div>
          <button class="button-secondary" type="button">发送测试邮件</button>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>安全</h2></div>
        <div class="settings-stack">
          <div class="setting-row"><div><strong>管理员双因素认证</strong><div class="meta-row"><span>后台账号必须绑定 2FA</span></div></div><span class="switch"></span></div>
          <div class="setting-row"><div><strong>登录失败锁定</strong><div class="meta-row"><span>连续失败后临时锁定账号</span></div></div><span class="switch"></span></div>
          <div class="field"><label for="session-days">会话有效期</label><select class="input" id="session-days"><option>7 天</option><option>14 天</option><option>30 天</option></select></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>备份</h2><span class="status published">正常</span></div>
        <div class="settings-stack">
          <div class="field"><label for="backup-cycle">备份频率</label><select class="input" id="backup-cycle"><option>每日全量备份</option><option>每周全量备份</option><option>手动备份</option></select></div>
          <div class="meta-row"><span>上次备份：2026-07-04 03:00</span><span>保留 7 天</span></div>
          <button class="button-secondary" type="button">立即备份</button>
        </div>
      </section>
    </section>
  </AdminLayout>
</template>
