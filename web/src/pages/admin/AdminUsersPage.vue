<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  exportAdminUsers,
  getAdminUsers,
  inviteAdminUser,
  requestAdminUserPasswordReset,
  updateAdminUserStatus,
  type ManagedUser,
  type UserStats
} from "../../shared/api";
import { downloadJson, exportFileName } from "../../shared/download";

const users = ref<ManagedUser[]>([]);
const stats = ref<UserStats>({ total: 0, emailVerified: 0, authors: 0, muted: 0, banned: 0 });
const loading = ref(false);
const exporting = ref(false);
const inviting = ref(false);
const actingId = ref("");
const resettingId = ref("");
const error = ref("");
const message = ref("");
const inviteOpen = ref(false);
const inviteEmail = ref("");
const inviteName = ref("");
const inviteRole = ref("author");
const selectedId = ref("");

const selectedUser = computed(() => users.value.find((user) => user.id === selectedId.value));

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminUsers();
    users.value = response.items;
    stats.value = response.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户列表加载失败";
  } finally {
    loading.value = false;
  }
}

function viewUser(user: ManagedUser) {
  selectedId.value = user.id;
  error.value = "";
  message.value = "";
}

async function setStatus(user: ManagedUser, status: ManagedUser["status"]) {
  actingId.value = user.id;
  error.value = "";
  message.value = "";

  try {
    await updateAdminUserStatus(user.id, status);
    message.value = "用户状态已更新。";
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户状态更新失败";
  } finally {
    actingId.value = "";
  }
}

async function exportUsers() {
  exporting.value = true;
  error.value = "";
  message.value = "";

  try {
    downloadJson(exportFileName("users"), await exportAdminUsers());
    message.value = "用户数据导出已生成。";
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
    message.value = result.resetToken ? `已邀请 ${result.user.displayName}，首次设置密码 token：${result.resetToken}` : `已邀请 ${result.user.displayName}。`;
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "邀请失败";
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
    message.value = result.resetToken ? `已生成重置 token：${result.resetToken}` : `已向 ${result.user.email} 发送重置入口。`;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "密码重置失败";
  } finally {
    resettingId.value = "";
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
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}
</script>

<template>
  <AdminLayout title="用户管理" description="管理注册用户、作者账号、禁言状态和评论行为。" mobile-title="用户管理" primary-action="邀请作者">
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
    <p v-if="message" class="muted">{{ message }}</p>

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

    <section class="table-panel" aria-label="用户列表">
      <form class="table-toolbar" @submit.prevent="load">
        <input class="input" type="search" placeholder="搜索用户名、邮箱、角色" aria-label="搜索用户">
        <select class="input" aria-label="用户状态">
          <option>全部状态</option>
          <option>正常</option>
          <option>待验证</option>
          <option>禁言</option>
          <option>封禁</option>
        </select>
        <select class="input" aria-label="角色">
          <option>全部角色</option>
          <option>注册用户</option>
          <option>作者</option>
          <option>编辑</option>
          <option>管理员</option>
        </select>
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
            <th>最近登录</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td><strong>{{ user.displayName }}</strong><div class="meta-row"><span>{{ user.email }}</span><span>{{ user.emailVerified ? "已验证邮箱" : "未验证邮箱" }}</span><span v-if="user.moderationNote">{{ user.moderationNote }}</span></div></td>
            <td>{{ roleText(user.role) }}</td>
            <td><span class="status" :class="statusClass(user.status)">{{ statusText(user.status) }}</span></td>
            <td>{{ user.commentCount }}</td>
            <td>{{ user.bookmarkCount }}</td>
            <td>{{ formatDate(user.lastLoginAt) }}</td>
            <td>
              <div class="header-actions">
                <button class="button-secondary" type="button" @click="viewUser(user)">查看</button>
                <button class="button-secondary" type="button" :disabled="resettingId === user.id" @click="resetPassword(user)">{{ resettingId === user.id ? "生成中..." : "重置密码" }}</button>
                <button v-if="user.status === 'active'" class="button-secondary" type="button" :disabled="actingId === user.id" @click="setStatus(user, 'muted')">禁言</button>
                <button v-else class="button-secondary" type="button" :disabled="actingId === user.id" @click="setStatus(user, 'active')">解除</button>
                <button v-if="user.status !== 'banned'" class="button-secondary" type="button" :disabled="actingId === user.id" @click="setStatus(user, 'banned')">封禁</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
