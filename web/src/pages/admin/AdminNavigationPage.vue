<script setup lang="ts">
import { onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminNavigation,
  updateAdminNavigation,
  type NavItem,
  type OperationsNavigation,
  type RedirectRule
} from "../../shared/api";
import { useToastStore } from "../../stores/toast";

type NavListKey = "topItems" | "footerItems";

const toast = useToastStore();
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
    toast.error("导航加载失败", error.value);
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
    navigation.value = await updateAdminNavigation(normalizedNavigation(navigation.value));
    message.value = "导航已保存。";
    toast.success("导航已保存", "前台菜单和重定向配置已更新。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "导航保存失败";
    toast.error("导航保存失败", error.value);
  } finally {
    saving.value = false;
  }
}

function addItem(target: NavListKey) {
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
  normalizeItemOrder(list);
}

function removeItem(target: NavListKey, item: NavItem) {
  if (!navigation.value) {
    return;
  }

  navigation.value[target] = navigation.value[target].filter((candidate) => candidate.id !== item.id);
  normalizeItemOrder(navigation.value[target]);
}

function moveItem(target: NavListKey, item: NavItem, direction: -1 | 1) {
  if (!navigation.value) {
    return;
  }

  const list = navigation.value[target];
  const index = list.findIndex((candidate) => candidate.id === item.id);
  const nextIndex = index + direction;
  if (index < 0 || nextIndex < 0 || nextIndex >= list.length) {
    return;
  }

  const [moved] = list.splice(index, 1);
  list.splice(nextIndex, 0, moved);
  normalizeItemOrder(list);
}

function addRedirect() {
  if (!navigation.value) {
    return;
  }

  navigation.value.redirects.push({
    from: "/old-path",
    to: "/",
    code: 301
  });
}

function removeRedirect(redirect: RedirectRule) {
  if (!navigation.value) {
    return;
  }

  navigation.value.redirects = navigation.value.redirects.filter((candidate) => candidate !== redirect);
}

function normalizedNavigation(value: OperationsNavigation): OperationsNavigation {
  return {
    ...value,
    topItems: normalizedItems(value.topItems),
    footerItems: normalizedItems(value.footerItems),
    redirects: normalizedRedirects(value.redirects)
  };
}

function normalizedItems(items: NavItem[]) {
  return items
    .filter((item) => item.label.trim() && item.url.trim())
    .map((item, index) => ({
      ...item,
      label: item.label.trim(),
      url: item.url.trim(),
      order: index + 1
    }));
}

function normalizedRedirects(items: RedirectRule[]) {
  return items
    .filter((item) => item.from.trim() && item.to.trim())
    .map((item) => ({
      from: item.from.trim(),
      to: item.to.trim(),
      code: [301, 302, 307, 308].includes(Number(item.code)) ? Number(item.code) : 301
    }));
}

function normalizeItemOrder(items: NavItem[]) {
  items.forEach((item, index) => {
    item.order = index + 1;
  });
}
</script>

<template>
  <AdminLayout title="导航管理" description="配置前台顶部菜单、底部菜单、社交链接和常用重定向。" mobile-title="导航管理" primary-action="保存">
    <template #mobile-action>
      <button class="button" type="button" :disabled="saving || !navigation" @click="save">{{ saving ? "保存中..." : "保存" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <a class="button-secondary" href="/" target="_blank" rel="noreferrer">预览站点</a>
        <button class="button" type="button" :disabled="saving || !navigation" @click="save">{{ saving ? "保存中..." : "保存导航" }}</button>
      </div>
    </template>

    <LoadingState v-if="loading" variant="page" text="正在加载导航..." :rows="4" />
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
            <article v-for="(item, index) in navigation.topItems" :key="item.id" class="nav-item">
              <div class="nav-order-actions">
                <button type="button" :disabled="index === 0" title="上移" @click="moveItem('topItems', item, -1)">↑</button>
                <button type="button" :disabled="index === navigation.topItems.length - 1" title="下移" @click="moveItem('topItems', item, 1)">↓</button>
              </div>
              <input v-model="item.label" class="input" aria-label="菜单名称">
              <input v-model="item.url" class="input" aria-label="菜单链接">
              <button class="button-secondary" type="button" @click="removeItem('topItems', item)">删除</button>
            </article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>底部导航</h2>
            <button class="button-secondary" type="button" @click="addItem('footerItems')">添加菜单</button>
          </div>
          <div class="nav-builder">
            <article v-for="(item, index) in navigation.footerItems" :key="item.id" class="nav-item">
              <div class="nav-order-actions">
                <button type="button" :disabled="index === 0" title="上移" @click="moveItem('footerItems', item, -1)">↑</button>
                <button type="button" :disabled="index === navigation.footerItems.length - 1" title="下移" @click="moveItem('footerItems', item, 1)">↓</button>
              </div>
              <input v-model="item.label" class="input" aria-label="菜单名称">
              <input v-model="item.url" class="input" aria-label="菜单链接">
              <button class="button-secondary" type="button" @click="removeItem('footerItems', item)">删除</button>
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
            <button class="button-secondary" type="button" @click="addRedirect">添加规则</button>
          </div>
          <div class="nav-builder">
            <article v-for="(redirect, index) in navigation.redirects" :key="`${redirect.from}-${index}`" class="nav-item redirect-item">
              <input v-model="redirect.from" class="input" aria-label="旧地址" placeholder="/old-path">
              <input v-model="redirect.to" class="input" aria-label="新地址" placeholder="/new-path">
              <select v-model.number="redirect.code" class="input" aria-label="状态码">
                <option :value="301">301</option>
                <option :value="302">302</option>
                <option :value="307">307</option>
                <option :value="308">308</option>
              </select>
              <button class="button-secondary" type="button" @click="removeRedirect(redirect)">删除</button>
            </article>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
