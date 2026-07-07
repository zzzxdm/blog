<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  ApiError,
  deleteAdminUser,
  exportAdminUsers,
  getAdminUsers,
  inviteAdminUser,
  requestAdminUserPasswordReset,
  restoreAdminUser,
  updateAdminUserRole,
  updateAdminUserStatus,
  type ManagedUser,
  type UserStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { downloadJson, exportFileName } from "../../shared/download";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";

const users = ref<ManagedUser[]>([]);
const toast = useToastStore();
const confirmDialog = useConfirmStore();
const stats = ref<UserStats>({ total: 0, emailVerified: 0, authors: 0, muted: 0, banned: 0 });
const loading = ref(false);
const exporting = ref(false);
const inviting = ref(false);
const actingId = ref("");
const resettingId = ref("");
const roleActingId = ref("");
const deletingId = ref("");
const restoringId = ref("");
const error = ref("");
const message = ref("");
const inviteOpen = ref(false);
const inviteEmail = ref("");
const inviteName = ref("");
const inviteRole = ref("author");
const selectedId = ref("");
const searchQuery = ref("");
const statusFilter = ref("");
const roleFilter = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const persistentMessage = computed(() => message.value.includes("token") || message.value.includes("临时密码"));

const selectedUser = computed(() => users.value.find((user) => user.id === selectedId.value));
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));
const visibleUsers = computed(() => users.value);

onMounted(load);

watch([statusFilter, roleFilter], () => {
  applyFilters();
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminUsers({
      page: page.value,
      pageSize: pageSize.value,
      q: searchQuery.value.trim(),
      status: statusFilter.value,
      role: roleFilter.value
    });
    users.value = response.items;
    stats.value = response.stats;
    total.value = response.total;
    page.value = response.page;
    if (users.value.length === 0 && total.value > 0 && page.value > totalPages.value) {
      page.value = totalPages.value;
      await load();
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户列表加载失败";
  } finally {
    loading.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  void load();
}

async function goPage(nextPage: number) {
  page.value = Math.min(Math.max(nextPage, 1), totalPages.value);
  await load();
}

async function changePageSize(nextPageSize: number) {
  pageSize.value = nextPageSize;
  page.value = 1;
  await load();
}

function viewUser(user: ManagedUser) {
  selectedId.value = user.id;
  error.value = "";
  message.value = "";
}

async function setStatus(user: ManagedUser, status: ManagedUser["status"]) {
  const confirmed = await confirmDialog.open(statusConfirmOptions(user, status));
  if (!confirmed) {
    return;
  }

  actingId.value = user.id;
  error.value = "";
  message.value = "";

  try {
    await updateAdminUserStatus(user.id, status);
    message.value = "用户状态已更新。";
    toast.success("用户状态已更新", `${user.displayName}：${statusText(status)}`);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户状态更新失败";
  } finally {
    actingId.value = "";
  }
}

function statusConfirmOptions(user: ManagedUser, status: ManagedUser["status"]) {
  if (status === "muted") {
    return {
      title: `禁言 ${user.displayName}`,
      message: "禁言后该用户不能继续发表评论，已有内容不会被删除。",
      confirmText: "禁言用户",
      tone: "danger" as const
    };
  }
  if (status === "banned") {
    return {
      title: `封禁 ${user.displayName}`,
      message: "封禁后该用户将无法正常使用账号功能，请确认这是预期操作。",
      confirmText: "封禁用户",
      tone: "danger" as const
    };
  }

  return {
    title: `解除限制 ${user.displayName}`,
    message: "解除后该用户会恢复正常账号状态。",
    confirmText: "解除限制",
    tone: "success" as const
  };
}

async function setRole(user: ManagedUser, role: string) {
  if (user.role === role) {
    return;
  }

  roleActingId.value = user.id;
  error.value = "";
  message.value = "";

  try {
    await updateAdminUserRole(user.id, role);
    message.value = "用户角色已更新。";
    toast.success("用户角色已更新", `${user.displayName}：${roleText(role)}`);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户角色更新失败";
  } finally {
    roleActingId.value = "";
  }
}

async function exportUsers() {
  exporting.value = true;
  error.value = "";
  message.value = "";

  try {
    downloadJson(exportFileName("users"), await exportAdminUsers());
    message.value = "用户数据导出已生成。";
    toast.success("用户数据已导出", "下载文件已生成。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户导出失败";
  } finally {
    exporting.value = false;
  }
}

async function inviteUser() {
  inviting.value = true;
  error.value = "";
  message.value = "";

  try {
    const result = await inviteAdminUser({
      email: inviteEmail.value,
      displayName: inviteName.value,
      role: inviteRole.value
    });
    inviteEmail.value = "";
    inviteName.value = "";
    inviteRole.value = "author";
    inviteOpen.value = false;
    if (result.resetToken || result.initialPassword) {
      const manualSecret = `临时密码：${result.initialPassword || "未返回"}，重置 token：${result.resetToken || "未返回"}`;
      message.value = result.delivery === "email-failed"
        ? `已邀请 ${result.user.displayName}，邮件发送失败，${manualSecret}`
        : `已邀请 ${result.user.displayName}，${manualSecret}`;
    } else {
      message.value = `已邀请 ${result.user.displayName}。`;
    }
    if (result.delivery === "email-failed") {
      toast.warning("邀请邮件发送失败", "请复制页面中的临时密码和重置 token 发给用户。");
    } else {
      toast.success("用户已邀请", result.delivery === "email" ? `已向 ${result.user.email} 发送临时密码和重置入口。` : "临时密码和重置 token 已显示在页面中。");
    }
    await load();
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      const email = inviteEmail.value.trim();
      searchQuery.value = email;
      page.value = 1;
      await load();
      inviteOpen.value = false;
      error.value = `该邮箱已在用户列表中：${email}。可以直接选中该用户调整角色，或点击“重置密码”重新发送登录入口。`;
      toast.warning("用户已存在", "已帮你用该邮箱过滤用户列表。");
    } else if (err instanceof ApiError && err.status === 410) {
      const email = inviteEmail.value.trim();
      searchQuery.value = email;
      statusFilter.value = "deleted";
      roleFilter.value = "";
      page.value = 1;
      await load();
      inviteOpen.value = false;
      error.value = `该邮箱对应的账号已被删除：${email}。已帮你筛出已删除用户，可以点击“恢复”重新启用账号。`;
      toast.warning("账号已删除", "已帮你切到已删除用户列表。");
    } else {
      error.value = err instanceof Error ? err.message : "邀请失败";
    }
  } finally {
    inviting.value = false;
  }
}

async function resetPassword(user: ManagedUser) {
  resettingId.value = user.id;
  error.value = "";
  message.value = "";

  try {
    const result = await requestAdminUserPasswordReset(user.id);
    message.value = result.resetToken ? `邮件发送失败，重置 token：${result.resetToken}` : `已向 ${result.user.email} 发送重置入口。`;
    if (result.delivery === "email-failed") {
      toast.warning("重置邮件发送失败", "请复制页面中的 token 发给用户。");
    } else {
      toast.success("重置入口已生成", result.delivery === "email" ? `已向 ${result.user.email} 发送重置入口。` : "重置 token 已显示在页面中。");
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "密码重置失败";
  } finally {
    resettingId.value = "";
  }
}

async function restoreUser(user: ManagedUser) {
  const confirmed = await confirmDialog.open({
    title: `恢复 ${user.displayName}`,
    message: "恢复后该账号可以重新登录，已失效的旧会话不会恢复。",
    confirmText: "恢复用户",
    tone: "success"
  });
  if (!confirmed) {
    return;
  }

  restoringId.value = user.id;
  error.value = "";
  message.value = "";

  try {
    await restoreAdminUser(user.id);
    message.value = "用户已恢复。";
    toast.success("用户已恢复", `${user.displayName} 可以重新登录。`);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户恢复失败";
  } finally {
    restoringId.value = "";
  }
}

async function deleteUser(user: ManagedUser) {
  const confirmed = await confirmDialog.open({
    title: `删除 ${user.displayName}`,
    message: "该用户会被标记为已删除并退出所有会话。",
    confirmText: "删除用户",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  deletingId.value = user.id;
  error.value = "";
  message.value = "";

  try {
    await deleteAdminUser(user.id);
    if (selectedId.value === user.id) {
      selectedId.value = "";
    }
    message.value = "用户已删除。";
    toast.success("用户已删除", user.displayName);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户删除失败";
  } finally {
    deletingId.value = "";
  }
}

function roleText(role: string) {
  if (role === "admin") return "管理员";
  if (role === "editor") return "编辑";
  if (role === "author") return "作者";
  return "注册用户";
}

function statusText(status: ManagedUser["status"]) {
  if (status === "muted") return "禁言";
  if (status === "banned") return "封禁";
  if (status === "deleted") return "已删除";
  return "正常";
}

function statusClass(status: ManagedUser["status"]) {
  if (status === "muted") return "muted";
  if (status === "banned" || status === "deleted") return "banned";
  return "published";
}

function formatDate(value: string) {
  return formatDateTime(value);
}
</script>

<template>
  <AdminLayout title="用户管理" description="管理注册用户、作者账号、禁言状态和评论行为。" mobile-title="用户管理" primary-action="邀请作者">
    <template #mobile-action>
      <button class="button" type="button" @click="inviteOpen = !inviteOpen">{{ inviteOpen ? "收起" : "邀请" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" :disabled="exporting" @click="exportUsers">{{ exporting ? "导出中..." : "导出用户" }}</button>
        <button class="button" type="button" @click="inviteOpen = !inviteOpen">{{ inviteOpen ? "收起邀请" : "邀请作者" }}</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="用户统计">
      <div class="stat-card"><span>注册用户</span><strong>{{ stats.total }}</strong></div>
      <div class="stat-card"><span>已验证邮箱</span><strong>{{ stats.emailVerified }}</strong></div>
      <div class="stat-card"><span>作者</span><strong>{{ stats.authors }}</strong></div>
      <div class="stat-card"><span>禁言中</span><strong>{{ stats.muted }}</strong></div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="persistentMessage" class="muted">{{ message }}</p>

    <section v-if="selectedUser" class="panel">
      <div class="panel-title">
        <h2>用户详情</h2>
        <span class="status" :class="statusClass(selectedUser.status)">{{ statusText(selectedUser.status) }}</span>
      </div>
      <div class="admin-grid-2">
        <div class="profile-hero">
          <span class="avatar">{{ selectedUser.avatarText }}</span>
          <div>
            <strong>{{ selectedUser.displayName }}</strong>
            <div class="meta-row"><span>{{ selectedUser.email }}</span><span>{{ roleText(selectedUser.role) }}</span></div>
          </div>
        </div>
        <div class="settings-stack">
          <div class="field">
            <label for="selected-user-role">角色</label>
            <select
              class="input"
              id="selected-user-role"
              :value="selectedUser.role"
              :disabled="roleActingId === selectedUser.id || selectedUser.status === 'deleted'"
              @change="setRole(selectedUser, ($event.target as HTMLSelectElement).value)"
            >
              <option value="user">注册用户</option>
              <option value="author">作者</option>
              <option value="editor">编辑</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div class="meta-row"><span>评论 {{ selectedUser.commentCount }}</span><span>收藏 {{ selectedUser.bookmarkCount }}</span><span>{{ selectedUser.emailVerified ? "已验证邮箱" : "未验证邮箱" }}</span></div>
          <div class="meta-row"><span>注册于 {{ formatDate(selectedUser.registeredAt) }}</span><span>最近登录 {{ formatDate(selectedUser.lastLoginAt) }}</span></div>
          <p v-if="selectedUser.moderationNote" class="muted">{{ selectedUser.moderationNote }}</p>
        </div>
      </div>
    </section>

    <section v-if="inviteOpen" class="panel">
      <div class="panel-title"><h2>邀请作者</h2><span class="tag">作者账号</span></div>
      <form class="settings-stack" @submit.prevent="inviteUser">
        <div class="admin-grid-2">
          <div class="field"><label for="invite-email">邮箱</label><input v-model="inviteEmail" class="input" id="invite-email" type="email" autocomplete="off"></div>
          <div class="field"><label for="invite-name">昵称</label><input v-model="inviteName" class="input" id="invite-name" autocomplete="off"></div>
        </div>
        <div class="field"><label for="invite-role">角色</label><select v-model="inviteRole" class="input" id="invite-role"><option value="author">作者</option><option value="editor">编辑</option></select></div>
        <div class="header-actions">
          <button class="button" type="submit" :disabled="inviting || !inviteEmail">{{ inviting ? "邀请中..." : "发送邀请" }}</button>
          <button class="button-secondary" type="button" :disabled="inviting" @click="inviteOpen = false">取消</button>
        </div>
      </form>
    </section>

    <section class="table-panel user-table-panel" aria-label="用户列表">
      <form class="table-toolbar user-table-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索用户名、邮箱、角色" aria-label="搜索用户">
        <select v-model="statusFilter" class="input" aria-label="用户状态">
          <option value="">全部状态</option>
          <option value="active">正常</option>
          <option value="unverified">待验证</option>
          <option value="muted">禁言</option>
          <option value="banned">封禁</option>
          <option value="deleted">已删除</option>
        </select>
        <select v-model="roleFilter" class="input" aria-label="角色">
          <option value="">全部角色</option>
          <option value="user">注册用户</option>
          <option value="author">作者</option>
          <option value="editor">编辑</option>
          <option value="admin">管理员</option>
        </select>
        <button class="button" type="submit" :disabled="loading">搜索</button>
      </form>

      <p v-if="loading" class="muted">正在加载用户...</p>
      <table v-else>
        <thead>
          <tr>
            <th>用户</th>
            <th>角色</th>
            <th>状态</th>
            <th>评论</th>
            <th>收藏</th>
            <th>注册时间</th>
            <th>最近登录</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in visibleUsers" :key="user.id">
            <td><strong>{{ user.displayName }}</strong><div class="meta-row"><span>{{ user.email }}</span><span>{{ user.emailVerified ? "已验证邮箱" : "未验证邮箱" }}</span><span v-if="user.moderationNote">{{ user.moderationNote }}</span></div></td>
            <td>
              <select
                class="input"
                :value="user.role"
                :aria-label="`${user.displayName}角色`"
                :disabled="roleActingId === user.id || user.status === 'deleted'"
                @change="setRole(user, ($event.target as HTMLSelectElement).value)"
              >
                <option value="user">注册用户</option>
                <option value="author">作者</option>
                <option value="editor">编辑</option>
                <option value="admin">管理员</option>
              </select>
            </td>
            <td><span class="status" :class="statusClass(user.status)">{{ statusText(user.status) }}</span></td>
            <td>{{ user.commentCount }}</td>
            <td>{{ user.bookmarkCount }}</td>
            <td>{{ formatDate(user.registeredAt) }}</td>
            <td>{{ formatDate(user.lastLoginAt) }}</td>
            <td>
              <div class="header-actions">
                <button class="button-secondary" type="button" @click="viewUser(user)">查看</button>
                <button v-if="user.status === 'deleted'" class="button-secondary button-success" type="button" :disabled="restoringId === user.id" @click="restoreUser(user)">{{ restoringId === user.id ? "恢复中..." : "恢复" }}</button>
                <button v-if="user.status !== 'deleted'" class="button-secondary" type="button" :disabled="resettingId === user.id" @click="resetPassword(user)">{{ resettingId === user.id ? "生成中..." : "重置密码" }}</button>
                <button v-if="user.status === 'active'" class="button-secondary button-danger" type="button" :disabled="actingId === user.id" @click="setStatus(user, 'muted')">禁言</button>
                <button v-else-if="user.status !== 'deleted'" class="button-secondary button-success" type="button" :disabled="actingId === user.id" @click="setStatus(user, 'active')">解除</button>
                <button v-if="user.status !== 'banned' && user.status !== 'deleted'" class="button-secondary button-danger" type="button" :disabled="actingId === user.id" @click="setStatus(user, 'banned')">封禁</button>
                <button v-if="user.status !== 'deleted'" class="button-secondary button-danger" type="button" :disabled="deletingId === user.id" @click="deleteUser(user)">{{ deletingId === user.id ? "删除中..." : "删除" }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="visibleUsers.length === 0">
            <td colspan="8" class="muted">没有匹配的用户。</td>
          </tr>
        </tbody>
      </table>

      <PaginationControls
        :page="page"
        :page-size="pageSize"
        :total="total"
        :loading="loading"
        item-label="个用户"
        show-page-size
        :page-size-options="[2, 5, 10, 20, 50, 100]"
        @update:page="goPage"
        @update:page-size="changePageSize"
      />
    </section>
  </AdminLayout>
</template>
