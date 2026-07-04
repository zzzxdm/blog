<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="编辑文章" description="自动保存于 15:26，当前版本 4。" mobile-title="写作" primary-action="发布">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">预览</button>
        <button class="button-secondary" type="button">保存草稿</button>
        <button class="button" type="button">发布</button>
      </div>
    </template>

    <section class="editor-layout">
      <div class="editor-panel">
        <div class="editor-toolbar">
          <div class="tool-group" aria-label="编辑工具栏">
            <button class="tool" type="button" aria-label="加粗">B</button>
            <button class="tool" type="button" aria-label="斜体">I</button>
            <button class="tool" type="button" aria-label="标题">H</button>
            <button class="tool" type="button" aria-label="引用">"</button>
            <button class="tool" type="button" aria-label="代码">{ }</button>
            <button class="tool" type="button" aria-label="链接">↗</button>
          </div>
          <div class="meta-row">
            <span>Markdown</span>
            <span>预览</span>
          </div>
        </div>

        <div class="editor-grid">
          <textarea class="markdown-area" aria-label="Markdown 编辑区"># 如何设计一个内容长期增长的博客系统

博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。

## 内容模型先于页面

文章需要拥有稳定的 slug、分类、标签、SEO 元数据、封面图、摘要、阅读时长和发布时间。

> 内容系统的核心不是页面，而是可被长期复用、迁移和分发的数据。

## 发布流程需要留出空间

成熟博客通常支持草稿、预览、审核、定时发布和版本历史。</textarea>

          <article class="preview-area">
            <h1>如何设计一个内容长期增长的博客系统</h1>
            <p>博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。</p>
            <h2>内容模型先于页面</h2>
            <p>文章需要拥有稳定的 slug、分类、标签、SEO 元数据、封面图、摘要、阅读时长和发布时间。</p>
            <blockquote>内容系统的核心不是页面，而是可被长期复用、迁移和分发的数据。</blockquote>
            <h2>发布流程需要留出空间</h2>
            <p>成熟博客通常支持草稿、预览、审核、定时发布和版本历史。</p>
          </article>
        </div>
      </div>

      <aside class="settings-stack" aria-label="文章设置">
        <section class="panel">
          <div class="panel-title">
            <h2>发布</h2>
            <span class="status draft">草稿</span>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="publish-time">发布时间</label>
              <input class="input" id="publish-time" type="datetime-local" value="2026-07-04T20:00">
            </div>
            <div class="field">
              <label for="visibility">可见性</label>
              <select class="input" id="visibility">
                <option>公开</option>
                <option>私密</option>
                <option>会员可见</option>
              </select>
            </div>
            <button class="button" type="button">提交审核</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>内容信息</h2>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="slug">Slug</label>
              <input class="input" id="slug" value="blog-system-design">
            </div>
            <div class="field">
              <label for="category">分类</label>
              <select class="input" id="category">
                <option>工程实践</option>
                <option>产品设计</option>
                <option>内容运营</option>
              </select>
            </div>
            <div class="field">
              <label for="tags">标签</label>
              <input class="input" id="tags" value="Vue3, SEO, 内容系统">
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>封面和主题</h2>
          </div>
          <div class="settings-stack">
            <img
              src="https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80"
              alt="当前文章封面图"
              style="border-radius: 8px; aspect-ratio: 16 / 9; object-fit: cover;"
            >
            <button class="button-secondary" type="button">更换封面</button>
            <div class="field">
              <label>强调色</label>
              <div class="swatches" aria-label="强调色选择">
                <span class="swatch" style="background:#295b4b"></span>
                <span class="swatch" style="background:#b95f2d"></span>
                <span class="swatch" style="background:#e3b45d"></span>
              </div>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>SEO</h2>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="seo-title">SEO 标题</label>
              <input class="input" id="seo-title" value="如何设计一个现代化博客系统">
            </div>
            <div class="field">
              <label for="seo-description">SEO 描述</label>
              <textarea class="input" id="seo-description">从内容模型、发布流程、SEO、缓存和运营能力设计一个可长期维护的博客系统。</textarea>
            </div>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
