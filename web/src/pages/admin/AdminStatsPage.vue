<script setup lang="ts">
import { onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  exportAdminStats,
  getAdminStats,
  type AdminStats
} from "../../shared/api";
import { downloadJson, exportFileName } from "../../shared/download";

const stats = ref<AdminStats>({
  range: "30d",
  rangeLabel: "最近 30 天",
  metrics: [],
  trend: [],
  topPosts: [],
  sources: [],
  searchTerms: [],
  suggestions: []
});
const loading = ref(false);
const exporting = ref(false);
const error = ref("");
const range = ref("30d");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    stats.value = await getAdminStats(range.value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "统计数据加载失败";
  } finally {
    loading.value = false;
  }
}

async function exportReport() {
  exporting.value = true;
  error.value = "";

  try {
    downloadJson(exportFileName(`stats-report-${range.value}`), await exportAdminStats(range.value));
  } catch (err) {
    error.value = err instanceof Error ? err.message : "统计报表导出失败";
  } finally {
    exporting.value = false;
  }
}
</script>

<template>
  <AdminLayout title="数据统计" description="查看内容趋势、热门内容、内容标签、来源和评论互动。" mobile-title="数据统计" primary-action="导出">
    <template #mobile-action>
      <button class="button" type="button" :disabled="exporting" @click="exportReport">{{ exporting ? "导出中..." : "导出" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <select v-model="range" class="input" aria-label="时间范围" @change="load">
          <option value="30d">最近 30 天</option>
          <option value="7d">最近 7 天</option>
          <option value="ytd">今年</option>
        </select>
        <button class="button-secondary" type="button" :disabled="exporting" @click="exportReport">{{ exporting ? "导出中..." : "导出报表" }}</button>
      </div>
    </template>

    <p v-if="loading" class="muted">正在加载统计数据...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <template v-else>
      <section class="stats-grid" aria-label="核心指标">
        <div v-for="metric in stats.metrics" :key="metric.label" class="stat-card">
          <span>{{ metric.label }}</span><strong>{{ metric.value }}</strong><div class="meta-row"><span>{{ metric.delta }}</span></div>
        </div>
      </section>

      <section class="admin-grid-2">
        <div class="settings-stack">
          <section class="chart-card">
            <div class="panel-title">
              <h2>访问趋势</h2>
              <span class="tag">{{ stats.rangeLabel }}</span>
            </div>
            <div class="bar-chart">
              <div v-for="point in stats.trend" :key="point.label" class="bar-row">
                <span>{{ point.label }}</span><span class="bar-track"><span class="bar-fill" :class="point.tone" :style="{ width: `${point.percent}%` }"></span></span><strong>{{ point.value }}</strong>
              </div>
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
                  <th>互动率</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="post in stats.topPosts" :key="post.title">
                  <td>{{ post.title }}</td><td>{{ post.views }}</td><td>{{ post.bookmarks }}</td><td>{{ post.comments }}</td><td>{{ post.engagementRate }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <aside class="settings-stack">
          <section class="chart-card">
            <h2>内容来源</h2>
            <div class="bar-chart">
              <div v-for="source in stats.sources" :key="source.label" class="bar-row">
                <span>{{ source.label }}</span><span class="bar-track"><span class="bar-fill" :class="source.tone" :style="{ width: `${source.percent}%` }"></span></span><strong>{{ source.value }}</strong>
              </div>
            </div>
          </section>

          <section class="panel">
            <div class="panel-title">
              <h2>热门内容标签</h2>
            </div>
            <ol class="rank-list">
              <li v-for="(term, index) in stats.searchTerms" :key="term.term"><span class="rank-number">{{ index + 1 }}</span><div><strong>{{ term.term }}</strong><span>{{ term.count }} 篇关联</span></div></li>
            </ol>
          </section>

          <section class="panel">
            <div class="panel-title">
              <h2>内容建议</h2>
            </div>
            <ul class="link-list">
              <li v-for="suggestion in stats.suggestions" :key="suggestion.title"><strong>{{ suggestion.title }}</strong><span>{{ suggestion.body }}</span></li>
            </ul>
          </section>
        </aside>
      </section>
    </template>
  </AdminLayout>
</template>
