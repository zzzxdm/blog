<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="数据统计" description="查看访问趋势、热门内容、搜索词、来源渠道和评论互动。" mobile-title="数据统计" primary-action="导出">
    <template #actions>
      <div class="header-actions">
        <select class="input" aria-label="时间范围">
          <option>最近 30 天</option>
          <option>最近 7 天</option>
          <option>今年</option>
        </select>
        <button class="button-secondary" type="button">导出报表</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="核心指标">
      <div class="stat-card"><span>PV</span><strong>86.4k</strong><div class="meta-row"><span>较上期 +18%</span></div></div>
      <div class="stat-card"><span>UV</span><strong>24.7k</strong><div class="meta-row"><span>较上期 +9%</span></div></div>
      <div class="stat-card"><span>平均阅读</span><strong>4:18</strong><div class="meta-row"><span>提升 32 秒</span></div></div>
      <div class="stat-card"><span>RSS 访问</span><strong>418</strong><div class="meta-row"><span>转化率 2.6%</span></div></div>
    </section>

    <section class="admin-grid-2">
      <div class="settings-stack">
        <section class="chart-card">
          <div class="panel-title">
            <h2>访问趋势</h2>
            <span class="tag">PV / UV</span>
          </div>
          <div class="bar-chart">
            <div class="bar-row"><span>周一</span><span class="bar-track"><span class="bar-fill" style="width: 68%;"></span></span><strong>8.4k</strong></div>
            <div class="bar-row"><span>周二</span><span class="bar-track"><span class="bar-fill" style="width: 74%;"></span></span><strong>9.1k</strong></div>
            <div class="bar-row"><span>周三</span><span class="bar-track"><span class="bar-fill rust" style="width: 92%;"></span></span><strong>11.3k</strong></div>
            <div class="bar-row"><span>周四</span><span class="bar-track"><span class="bar-fill" style="width: 81%;"></span></span><strong>9.9k</strong></div>
            <div class="bar-row"><span>周五</span><span class="bar-track"><span class="bar-fill amber" style="width: 64%;"></span></span><strong>7.8k</strong></div>
            <div class="bar-row"><span>周六</span><span class="bar-track"><span class="bar-fill" style="width: 48%;"></span></span><strong>5.9k</strong></div>
          </div>
        </section>

        <section class="table-panel">
          <div class="panel-title" style="padding: 16px 16px 0;">
            <h2>热门文章</h2>
            <RouterLink class="button-secondary" to="/admin/posts">查看文章</RouterLink>
          </div>
          <table>
            <thead>
              <tr>
                <th>文章</th>
                <th>阅读</th>
                <th>收藏</th>
                <th>评论</th>
                <th>RSS 访问</th>
              </tr>
            </thead>
            <tbody>
              <tr><td>Vue3 内容站的缓存与 SEO 边界</td><td>12,420</td><td>312</td><td>48</td><td>3.8%</td></tr>
              <tr><td>如何设计一个内容长期增长的博客系统</td><td>9,884</td><td>286</td><td>34</td><td>4.1%</td></tr>
              <tr><td>让旧文章继续被搜索引擎找到</td><td>7,209</td><td>190</td><td>19</td><td>2.4%</td></tr>
            </tbody>
          </table>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="chart-card">
          <h2>来源渠道</h2>
          <div class="bar-chart">
            <div class="bar-row"><span>搜索</span><span class="bar-track"><span class="bar-fill" style="width: 72%;"></span></span><strong>46%</strong></div>
            <div class="bar-row"><span>直接</span><span class="bar-track"><span class="bar-fill rust" style="width: 34%;"></span></span><strong>22%</strong></div>
            <div class="bar-row"><span>社交</span><span class="bar-track"><span class="bar-fill amber" style="width: 28%;"></span></span><strong>18%</strong></div>
            <div class="bar-row"><span>订阅</span><span class="bar-track"><span class="bar-fill" style="width: 20%;"></span></span><strong>14%</strong></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>热门搜索词</h2>
          </div>
          <ol class="rank-list">
            <li><span class="rank-number">1</span><div><strong>Vue3 SEO</strong><span>1,284 次搜索</span></div></li>
            <li><span class="rank-number">2</span><div><strong>博客系统设计</strong><span>936 次搜索</span></div></li>
            <li><span class="rank-number">3</span><div><strong>Markdown 编辑器</strong><span>642 次搜索</span></div></li>
          </ol>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>内容建议</h2>
          </div>
          <ul class="link-list">
            <li><strong>搜索词“评论审核”增长明显</strong><span>可以补一篇用户评论系统设计。</span></li>
            <li><strong>专题页带来 18% 收藏</strong><span>建议继续完善专题导航。</span></li>
          </ul>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
