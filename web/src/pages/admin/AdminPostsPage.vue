<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="文章管理" description="管理草稿、审核、定时发布和已发布内容。" mobile-title="文章管理" primary-action="新建">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">导入</button>
        <RouterLink class="button" to="/admin/editor">新建文章</RouterLink>
      </div>
    </template>

    <section class="stats-grid" aria-label="文章统计">
      <div class="stat-card">
        <span>已发布</span>
        <strong>96</strong>
      </div>
      <div class="stat-card">
        <span>草稿</span>
        <strong>18</strong>
      </div>
      <div class="stat-card">
        <span>待审核</span>
        <strong>7</strong>
      </div>
      <div class="stat-card">
        <span>本月阅读</span>
        <strong>16.8k</strong>
      </div>
    </section>

    <section class="table-panel" aria-label="文章列表">
      <form class="table-toolbar">
        <input class="input" type="search" placeholder="搜索标题、作者、标签" aria-label="搜索文章">
        <select class="input" aria-label="文章状态">
          <option>全部状态</option>
          <option>已发布</option>
          <option>草稿</option>
          <option>待审核</option>
        </select>
        <select class="input" aria-label="排序">
          <option>最近更新</option>
          <option>最多阅读</option>
          <option>定时发布</option>
        </select>
      </form>

      <table>
        <thead>
          <tr>
            <th>标题</th>
            <th>状态</th>
            <th>分类</th>
            <th>阅读</th>
            <th>评论</th>
            <th>更新时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <strong>如何设计一个内容长期增长的博客系统</strong>
              <div class="meta-row">
                <span>管理员</span>
                <span>/posts/blog-system-design</span>
              </div>
            </td>
            <td><span class="status published">已发布</span></td>
            <td>工程实践</td>
            <td>2,984</td>
            <td>34</td>
            <td>2026-07-04 15:20</td>
            <td><RouterLink class="button-secondary" to="/admin/editor">编辑</RouterLink></td>
          </tr>
          <tr>
            <td>
              <strong>Vue3 内容站的缓存与 SEO 边界</strong>
              <div class="meta-row">
                <span>管理员</span>
                <span>定时发布：今晚 20:00</span>
              </div>
            </td>
            <td><span class="status review">待发布</span></td>
            <td>Vue3</td>
            <td>0</td>
            <td>0</td>
            <td>2026-07-04 10:12</td>
            <td><RouterLink class="button-secondary" to="/admin/editor">编辑</RouterLink></td>
          </tr>
          <tr>
            <td>
              <strong>为什么博客后台需要文章版本历史</strong>
              <div class="meta-row">
                <span>编辑审核中</span>
                <span>3 个版本</span>
              </div>
            </td>
            <td><span class="status review">待审核</span></td>
            <td>内容治理</td>
            <td>1,988</td>
            <td>12</td>
            <td>2026-07-03 18:36</td>
            <td><RouterLink class="button-secondary" to="/admin/editor">编辑</RouterLink></td>
          </tr>
          <tr>
            <td>
              <strong>把 Markdown 写作体验做到顺手</strong>
              <div class="meta-row">
                <span>自动保存于 2 分钟前</span>
                <span>未设置封面图</span>
              </div>
            </td>
            <td><span class="status draft">草稿</span></td>
            <td>编辑器</td>
            <td>0</td>
            <td>0</td>
            <td>2026-07-04 15:12</td>
            <td><RouterLink class="button-secondary" to="/admin/editor">编辑</RouterLink></td>
          </tr>
          <tr>
            <td>
              <strong>让旧文章继续被搜索引擎找到</strong>
              <div class="meta-row">
                <span>管理员</span>
                <span>包含 2 条 301 重定向</span>
              </div>
            </td>
            <td><span class="status published">已发布</span></td>
            <td>SEO</td>
            <td>2,470</td>
            <td>19</td>
            <td>2026-06-12 09:45</td>
            <td><RouterLink class="button-secondary" to="/admin/editor">编辑</RouterLink></td>
          </tr>
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
