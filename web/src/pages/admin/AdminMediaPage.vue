<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="媒体库" description="管理文章封面、正文图片、附件和图片替代文本。" mobile-title="媒体库" primary-action="上传">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">批量选择</button>
        <button class="button" type="button">上传文件</button>
      </div>
    </template>

    <section class="admin-grid-2">
      <div>
        <form class="media-toolbar">
          <input class="input" type="search" placeholder="搜索文件名、alt 文本、上传人" aria-label="搜索媒体">
          <select class="input" aria-label="文件类型">
            <option>全部类型</option>
            <option>图片</option>
            <option>文档</option>
            <option>视频</option>
          </select>
          <select class="input" aria-label="排序">
            <option>最近上传</option>
            <option>文件最大</option>
            <option>使用最多</option>
          </select>
        </form>

        <section class="media-grid" aria-label="媒体资源">
          <article class="media-card">
            <img src="https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80" alt="代码编辑器封面图">
            <div class="media-card-body"><strong>cover-code-desk.jpg</strong><div class="meta-row"><span>1.2 MB</span><span>已使用 8 次</span></div><span class="tag">封面图</span></div>
          </article>
          <article class="media-card">
            <img src="https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=700&q=80" alt="服务器设备图片">
            <div class="media-card-body"><strong>server-room.jpg</strong><div class="meta-row"><span>980 KB</span><span>已使用 3 次</span></div><span class="tag rust">架构</span></div>
          </article>
          <article class="media-card">
            <img src="https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=700&q=80" alt="写作桌面图片">
            <div class="media-card-body"><strong>writing-notes.jpg</strong><div class="meta-row"><span>860 KB</span><span>已使用 5 次</span></div><span class="tag amber">写作</span></div>
          </article>
          <article class="media-card">
            <img src="https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=700&q=80" alt="产品分析界面">
            <div class="media-card-body"><strong>product-dashboard.jpg</strong><div class="meta-row"><span>1.4 MB</span><span>已使用 2 次</span></div><span class="tag">产品设计</span></div>
          </article>
          <article class="media-card">
            <img src="https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=700&q=80" alt="代码窗口">
            <div class="media-card-body"><strong>javascript-editor.jpg</strong><div class="meta-row"><span>740 KB</span><span>已使用 4 次</span></div><span class="tag">Vue3</span></div>
          </article>
          <article class="media-card">
            <img src="https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=700&q=80" alt="自然光工作区">
            <div class="media-card-body"><strong>quiet-workspace.jpg</strong><div class="meta-row"><span>690 KB</span><span>未使用</span></div><span class="tag amber">备用</span></div>
          </article>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="upload-zone">
          <div>
            <strong>拖拽文件到这里上传</strong>
            <p>支持 JPG、PNG、WebP、GIF 和 PDF。图片会自动生成响应式尺寸。</p>
            <button class="button" type="button">选择文件</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>选中资源</h2>
            <span class="tag">图片</span>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="filename">文件名</label><input class="input" id="filename" value="cover-code-desk.jpg"></div>
            <div class="field"><label for="alt">Alt 文本</label><input class="input" id="alt" value="桌面上的代码编辑器和开发设备"></div>
            <div class="field"><label for="asset-url">资源地址</label><input class="input" id="asset-url" value="https://cdn.example.com/cover-code-desk.jpg"></div>
            <div class="meta-row"><span>尺寸 1400 x 930</span><span>上传于 2026-07-04</span></div>
            <button class="button-secondary" type="button">复制地址</button>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
