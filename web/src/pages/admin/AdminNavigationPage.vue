<script setup lang="ts">
import { onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminNavigation,
  updateAdminNavigation,
  type NavItem,
  type OperationsNavigation
} from "../../shared/api";

const navigation = ref<OperationsNavigation | null>(null);
const loading = ref(false);
const saving = ref(false);
const error = ref("");
const message = ref("");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    navigation.value = await getAdminNavigation();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "导航加载失败";
  } finally {
    loading.value = false;
  }
}

async function save() {
  if (!navigation.value) {
    return;
  }

  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    navigation.value = await updateAdminNavigation(navigation.value);
    message.value = "导航已保存。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "导航保存失败";
  } finally {
    saving.value = false;
  }
}

function addItem(target: "topItems" | "footerItems") {
  if (!navigation.value) {
    return;
  }

  const list = navigation.value[target];
  list.push({
    id: `${target}_${Date.now()}`,
    label: "新菜单",
    url: "/",
    order: list.length + 1
  });
}

function removeItem(target: "topItems" | "footerItems", item: NavItem) {
  if (!navigation.value) {
    return;
  }

  navigation.value[target] = navigation.value[target].filter((candidate) => candidate.id !== item.id);
}
</script>

<template>
  <AdminLayout title="导航管理" description="配置前台顶部菜单、底部菜单、社交链接和常用重定向。" mobile-title="导航管理" primary-action="保存">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">预览站点</button>
        <button class="button" type="button" :disabled="saving || !navigation" @click="save">{{ saving ? "保存中..." : "保存导航" }}</button>
      </div>
    </template>

    <p v-if="loading" class="muted">正在加载导航...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section v-if="navigation" class="admin-grid-2">
      <div class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>顶部导航</h2>
            <button class="button-secondary" type="button" @click="addItem('topItems')">添加菜单</button>
          </div>
          <div class="nav-builder">
            <article v-for="item in navigation.topItems" :key="item.id" class="nav-item">
              <span class="drag-handle">≡</span><input v-model="item.label" class="input" aria-label="菜单名称"><input v-model="item.url" class="input" aria-label="菜单链接"><button class="button-secondary" type="button" @click="removeItem('topItems', item)">删除</button>
            </article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>底部导航</h2>
            <button class="button-secondary" type="button" @click="addItem('footerItems')">添加菜单</button>
          </div>
          <div class="nav-builder">
            <article v-for="item in navigation.footerItems" :key="item.id" class="nav-item">
              <span class="drag-handle">≡</span><input v-model="item.label" class="input" aria-label="菜单名称"><input v-model="item.url" class="input" aria-label="菜单链接"><button class="button-secondary" type="button" @click="removeItem('footerItems', item)">删除</button>
            </article>
          </div>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>菜单设置</h2>
          </div>
          <div class="settings-stack">
            <label class="setting-row"><div><strong>移动端折叠菜单</strong><div class="meta-row"><span>窄屏时隐藏横向导航</span></div></div><input v-model="navigation.mobileCollapse" type="checkbox"></label>
            <label class="setting-row"><div><strong>外链新窗口打开</strong><div class="meta-row"><span>对外部链接自动添加安全属性</span></div></div><input v-model="navigation.externalLinksNewWindow" type="checkbox"></label>
            <label class="setting-row"><div><strong>显示登录入口</strong><div class="meta-row"><span>前台顶部展示用户入口</span></div></div><input v-model="navigation.showLoginEntry" type="checkbox"></label>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>社交链接</h2>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="github">GitHub</label><input v-model="navigation.githubUrl" class="input" id="github"></div>
            <div class="field"><label for="email">联系邮箱</label><input v-model="navigation.contactEmail" class="input" id="email"></div>
            <div class="field"><label for="rss">RSS</label><input v-model="navigation.rssUrl" class="input" id="rss"></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>重定向</h2>
            <span class="tag rust">{{ navigation.redirects.length }} 条</span>
          </div>
          <ul class="link-list">
            <li v-for="redirect in navigation.redirects" :key="redirect.from"><strong>{{ redirect.from }}</strong><span>{{ redirect.code }} 到 {{ redirect.to }}</span></li>
          </ul>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
