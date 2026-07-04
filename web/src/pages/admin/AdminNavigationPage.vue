<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="导航管理" description="配置前台顶部菜单、底部菜单、社交链接和常用重定向。" mobile-title="导航管理" primary-action="保存">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">预览站点</button>
        <button class="button" type="button">保存导航</button>
      </div>
    </template>

    <section class="admin-grid-2">
      <div class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>顶部导航</h2>
            <button class="button-secondary" type="button">添加菜单</button>
          </div>
          <div class="nav-builder">
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="首页" aria-label="菜单名称"><input class="input" value="/" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="归档" aria-label="菜单名称"><input class="input" value="/archive" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="专题" aria-label="菜单名称"><input class="input" value="/topics" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="关于" aria-label="菜单名称"><input class="input" value="/about" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>底部导航</h2>
            <button class="button-secondary" type="button">添加菜单</button>
          </div>
          <div class="nav-builder">
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="RSS" aria-label="菜单名称"><input class="input" value="/rss.xml" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="友情链接" aria-label="菜单名称"><input class="input" value="/friends" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
            <article class="nav-item"><span class="drag-handle">≡</span><input class="input" value="隐私政策" aria-label="菜单名称"><input class="input" value="/privacy" aria-label="菜单链接"><button class="button-secondary" type="button">删除</button></article>
          </div>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>菜单设置</h2>
          </div>
          <div class="settings-stack">
            <div class="setting-row"><div><strong>移动端折叠菜单</strong><div class="meta-row"><span>窄屏时隐藏横向导航</span></div></div><span class="switch"></span></div>
            <div class="setting-row"><div><strong>外链新窗口打开</strong><div class="meta-row"><span>对外部链接自动添加安全属性</span></div></div><span class="switch"></span></div>
            <div class="setting-row"><div><strong>显示登录入口</strong><div class="meta-row"><span>前台顶部展示用户入口</span></div></div><span class="switch"></span></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>社交链接</h2>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="github">GitHub</label><input class="input" id="github" value="https://github.com/example"></div>
            <div class="field"><label for="email">联系邮箱</label><input class="input" id="email" value="hello@example.com"></div>
            <div class="field"><label for="rss">RSS</label><input class="input" id="rss" value="/rss.xml"></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>重定向</h2>
            <span class="tag rust">12 条</span>
          </div>
          <ul class="link-list">
            <li><strong>/old-blog-design</strong><span>301 到 /posts/blog-system-design</span></li>
            <li><strong>/newsletter</strong><span>302 到 /topics/content-growth</span></li>
          </ul>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
