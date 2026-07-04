<script setup lang="ts">
import AccountLayout from "../../components/AccountLayout.vue";
</script>

<template>
  <AccountLayout title="我的收藏" description="整理值得反复阅读的文章，并支持按专题、分类和时间筛选。">
    <template #actions>
      <RouterLink class="button-secondary" to="/archive">继续浏览</RouterLink>
    </template>

    <section class="stats-grid" aria-label="收藏统计">
      <div class="stat-card"><span>收藏文章</span><strong>18</strong></div>
      <div class="stat-card"><span>本月新增</span><strong>5</strong></div>
      <div class="stat-card"><span>工程实践</span><strong>9</strong></div>
      <div class="stat-card"><span>待读</span><strong>4</strong></div>
    </section>

    <section class="panel">
      <form class="archive-toolbar">
        <input class="input" type="search" placeholder="搜索收藏文章" aria-label="搜索收藏">
        <select class="input" aria-label="收藏分类"><option>全部分类</option><option>工程实践</option><option>产品设计</option><option>内容治理</option></select>
        <select class="input" aria-label="排序"><option>最近收藏</option><option>最近发布</option><option>阅读最多</option></select>
      </form>
      <div class="article-list">
        <article class="article-card">
          <img src="https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80" alt="开发桌面">
          <div class="article-card-body">
            <div class="meta-row"><span class="tag">工程实践</span><span>收藏于今天</span></div>
            <h3><RouterLink to="/posts/blog-system-design">如何设计一个内容长期增长的博客系统</RouterLink></h3>
            <p>内容模型、发布流程、SEO、缓存、搜索和数据治理。</p>
            <div class="meta-row"><span>12 分钟阅读</span><span>2,984 次阅读</span><button class="button-secondary" type="button">取消收藏</button></div>
          </div>
        </article>
        <article class="article-card">
          <img src="https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=700&q=80" alt="数据看板">
          <div class="article-card-body">
            <div class="meta-row"><span class="tag rust">Vue3</span><span>收藏于昨天</span></div>
            <h3><RouterLink to="/posts/vue3-content-site-cache-seo">Vue3 内容站的缓存和接口边界</RouterLink></h3>
            <p>前端状态、API 聚合、页面缓存和 SEO meta 的协作方式。</p>
            <div class="meta-row"><span>8 分钟阅读</span><span>待读</span><button class="button-secondary" type="button">取消收藏</button></div>
          </div>
        </article>
      </div>
    </section>
  </AccountLayout>
</template>
