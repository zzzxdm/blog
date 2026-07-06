<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  createAdminMessage,
  exportAdminMessages,
  getAdminMessages,
  getAdminUsers,
  type ManagedUser,
  type MessageStats,
  type MessageType,
  type StationMessage
} from "../../shared/api";
import { downloadJson, exportFileName } from "../../shared/download";

const messages = ref<StationMessage[]>([]);
const users = ref<ManagedUser[]>([]);
const stats = ref<MessageStats>({ unread: 0, review: 0, admin: 0, archived: 0, scheduled: 0, total: 0 });
const loading = ref(false);
const usersLoading = ref(false);
const sending = ref(false);
const exporting = ref(false);
const error = ref("");
const message = ref("");
const selectedId = ref("");

const messageScope = ref("single");
const targetRole = ref("author");
const recipientId = ref("");
const recipientName = ref("");
const messageType = ref<MessageType>("admin");
const priority = ref("normal");
const title = ref("");
const body = ref("");
const targetTitle = ref("");
const scheduleOpen = ref(false);
const scheduledAt = ref("");
const searchQuery = ref("");
const typeFilter = ref("");
const statusFilter = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const selectedMessage = computed(() => messages.value.find((item) => item.id === selectedId.value) || messages.value[0]);
const selectableUsers = computed(() => users.value.filter((user) => canReceiveMessage(user)));
const selectedRecipient = computed(() => selectableUsers.value.find((user) => user.id === recipientId.value));

onMounted(() => {
  void load();
  void loadUsers();
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminMessages({
      q: searchQuery.value,
      type: typeFilter.value,
      status: statusFilter.value,
      page: page.value,
      pageSize: pageSize.value
    });
    messages.value = response.items;
    stats.value = response.stats;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
    if (!messages.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = messages.value[0]?.id || "";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信记录加载失败";
  } finally {
    loading.value = false;
  }
}

async function applyFilters() {
  page.value = 1;
  await load();
}

async function setPage(value: number) {
  page.value = value;
  await load();
}

async function setPageSize(value: number) {
  pageSize.value = value;
  page.value = 1;
  await load();
}

async function send() {
  if (!title.value.trim() || !body.value.trim()) {
    error.value = "请输入标题和正文";
    return;
  }

  if (scheduleOpen.value && !scheduledAt.value) {
    error.value = "请选择定时发送时间";
    return;
  }

  sending.value = true;
  error.value = "";
  message.value = "";

  try {
    const scheduleValue = scheduleOpen.value ? new Date(scheduledAt.value).toISOString() : "";
    const recipients = await resolveRecipients();
    if (!recipients.length) {
      throw new Error("没有匹配的接收人");
    }

    for (const recipient of recipients) {
      await createAdminMessage({
        recipientId: recipient.id,
        recipientName: recipient.name,
        type: messageType.value,
        priority: priority.value,
        title: title.value,
        body: body.value,
        targetType: "admin-message",
        targetTitle: targetTitle.value,
        scheduledAt: scheduleValue || undefined
      });
    }
    message.value = scheduleValue && new Date(scheduleValue) > new Date()
      ? `已定时 ${recipients.length} 条站内信到 ${formatTime(scheduleValue)}。`
      : `已发送 ${recipients.length} 条站内信。`;
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信发送失败";
  } finally {
    sending.value = false;
  }
}

async function exportMessages() {
  exporting.value = true;
  error.value = "";
  message.value = "";

  try {
    downloadJson(exportFileName("messages"), await exportAdminMessages());
    message.value = "站内信记录导出已生成。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信导出失败";
  } finally {
    exporting.value = false;
  }
}

async function loadUsers() {
  usersLoading.value = true;

  try {
    const result = await getAdminUsers({ all: true });
    users.value = result.items;
    if (!selectableUsers.value.some((user) => user.id === recipientId.value)) {
      applyRecipient(selectableUsers.value[0]);
    } else {
      syncRecipientFromSelect();
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "用户列表加载失败";
  } finally {
    usersLoading.value = false;
  }
}

async function resolveRecipients() {
  if (messageScope.value === "single") {
    syncRecipientFromSelect();
    return [{ id: recipientId.value.trim(), name: recipientName.value.trim() || recipientId.value.trim() }].filter((item) => item.id);
  }

  if (!users.value.length) {
    const result = await getAdminUsers({ all: true });
    users.value = result.items;
  }

  return users.value
    .filter((user) => canReceiveMessage(user))
    .filter((user) => messageScope.value === "all" || user.role === targetRole.value)
    .map((user) => ({ id: user.id, name: user.displayName }));
}

function applyRecipient(user?: ManagedUser) {
  recipientId.value = user?.id || "";
  recipientName.value = user?.displayName || "";
}

function syncRecipientFromSelect() {
  if (selectedRecipient.value) {
    recipientName.value = selectedRecipient.value.displayName;
  }
}

function canReceiveMessage(user: ManagedUser) {
  return user.status !== "deleted" && user.status !== "banned";
}

function roleText(role: string) {
  if (role === "admin") return "管理员";
  if (role === "editor") return "编辑";
  if (role === "author") return "作者";
  return "注册用户";
}

function toggleSchedule() {
  scheduleOpen.value = !scheduleOpen.value;
  if (scheduleOpen.value && !scheduledAt.value) {
    scheduledAt.value = nextScheduleValue();
  }
  message.value = "";
  error.value = "";
}

function viewMessage(item: StationMessage) {
  selectedId.value = item.id;
  message.value = "";
  error.value = "";
}

function formatTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function typeText(value: MessageType) {
  if (value === "review") {
    return "投稿审核";
  }
  if (value === "system") {
    return "公告";
  }
  if (value === "comment") {
    return "评论";
  }
  if (value === "account") {
    return "账号";
  }
  return "管理员";
}

function typeClass(value: MessageType) {
  if (value === "review") {
    return "review";
  }
  if (value === "system" || value === "account") {
    return "muted";
  }
  return "published";
}

function statusText(value: StationMessage["status"]) {
  if (value === "scheduled") {
    return "定时中";
  }
  if (value === "archived") {
    return "已归档";
  }
  if (value === "read") {
    return "已读";
  }
  return "未读";
}

function nextScheduleValue() {
  const date = new Date();
  date.setHours(date.getHours() + 1, 0, 0, 0);
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
</script>

<template>
  <AdminLayout title="站内信管理" description="向用户发送审核结果、站点公告和定向运营消息。" mobile-title="站内信管理" primary-action="发送">
    <template #mobile-action>
      <button class="button" type="button" :disabled="sending" @click="send">{{ sending ? "发送中..." : "发送" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" :disabled="exporting" @click="exportMessages">{{ exporting ? "导出中..." : "导出记录" }}</button>
        <button class="button" type="button" :disabled="sending" @click="send">{{ sending ? "发送中..." : "发送消息" }}</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="站内信统计">
      <div class="stat-card"><span>总发送</span><strong>{{ stats.total }}</strong></div>
      <div class="stat-card"><span>当前未读</span><strong>{{ stats.unread }}</strong></div>
      <div class="stat-card"><span>审核消息</span><strong>{{ stats.review }}</strong></div>
      <div class="stat-card"><span>定时中</span><strong>{{ stats.scheduled }}</strong></div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="admin-grid-2">
      <section class="table-panel" aria-label="发送记录">
        <form class="table-toolbar" @submit.prevent="applyFilters">
          <input v-model="searchQuery" class="input" type="search" placeholder="搜索标题、接收人、类型" aria-label="搜索站内信">
          <select v-model="typeFilter" class="input" aria-label="消息类型" @change="applyFilters">
            <option value="">全部类型</option>
            <option value="admin">管理员消息</option>
            <option value="review">投稿审核</option>
            <option value="system">站点公告</option>
            <option value="account">系统事件</option>
            <option value="comment">评论</option>
          </select>
          <select v-model="statusFilter" class="input" aria-label="发送状态" @change="applyFilters">
            <option value="">全部状态</option>
            <option value="sent">已发送</option>
            <option value="scheduled">定时中</option>
            <option value="archived">已归档</option>
          </select>
        </form>

        <p v-if="loading" class="muted">正在加载站内信记录...</p>
        <table v-else>
          <thead>
            <tr>
              <th>消息</th>
              <th>范围</th>
              <th>类型</th>
              <th>送达</th>
              <th>已读</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in messages" :key="item.id">
              <td><strong>{{ item.title }}</strong><div class="meta-row"><span>{{ item.targetTitle || item.body }}</span></div></td>
              <td>{{ item.recipientName }}</td>
              <td><span class="status" :class="typeClass(item.type)">{{ typeText(item.type) }}</span></td>
              <td>{{ item.status === "scheduled" ? "定时中" : "1 / 1" }}</td>
              <td>{{ item.status === "scheduled" ? "-" : item.status === "unread" ? "0 / 1" : "1 / 1" }}</td>
              <td>{{ formatTime(item.scheduledAt || item.createdAt) }}</td>
              <td><button class="button-secondary" type="button" @click="viewMessage(item)">查看</button></td>
            </tr>
            <tr v-if="messages.length === 0">
              <td colspan="7" class="muted">没有匹配的站内信。</td>
            </tr>
          </tbody>
        </table>
        <PaginationControls
          v-if="!loading"
          :page="page"
          :page-size="pageSize"
          :total="total"
          :loading="loading"
          item-label="条站内信"
          show-page-size
          :page-size-options="[5, 10, 20, 50, 100]"
          @update:page="setPage"
          @update:page-size="setPageSize"
        />
      </section>

      <aside class="settings-stack">
        <section v-if="selectedMessage" class="panel">
          <div class="panel-title">
            <h2>消息详情</h2>
            <span class="status" :class="selectedMessage.status === 'scheduled' ? 'muted' : typeClass(selectedMessage.type)">{{ statusText(selectedMessage.status) }}</span>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="detail-title">标题</label><input class="input" id="detail-title" :value="selectedMessage.title" readonly></div>
            <div class="field"><label for="detail-body">正文</label><textarea class="input" id="detail-body" :value="selectedMessage.body" readonly></textarea></div>
            <div class="admin-grid-2">
              <div class="field"><label for="detail-recipient">接收人</label><input class="input" id="detail-recipient" :value="selectedMessage.recipientName" readonly></div>
              <div class="field"><label for="detail-type">类型</label><input class="input" id="detail-type" :value="typeText(selectedMessage.type)" readonly></div>
            </div>
            <div class="field"><label for="detail-target">关联目标</label><input class="input" id="detail-target" :value="selectedMessage.targetTitle || selectedMessage.targetId || '无'" readonly></div>
            <div class="meta-row"><span>发送人 {{ selectedMessage.senderName }}</span><span>{{ formatTime(selectedMessage.createdAt) }}</span></div>
            <div v-if="selectedMessage.scheduledAt" class="meta-row"><span>预约时间 {{ formatTime(selectedMessage.scheduledAt) }}</span></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>发送站内信</h2>
          </div>
          <form class="message-compose" @submit.prevent="send">
            <div class="field"><label for="message-scope">接收范围</label><select v-model="messageScope" class="input" id="message-scope"><option value="single">指定用户</option><option value="all">全部注册用户</option><option value="role">按角色发送</option></select></div>
            <div v-if="messageScope === 'single'" class="field">
              <label for="message-recipient">接收用户</label>
              <select v-model="recipientId" class="input" id="message-recipient" :disabled="usersLoading || !selectableUsers.length" @change="syncRecipientFromSelect">
                <option v-if="usersLoading" value="">正在加载用户...</option>
                <option v-for="user in selectableUsers" :key="user.id" :value="user.id">{{ user.displayName }} · {{ roleText(user.role) }} · {{ user.email }}</option>
              </select>
              <div v-if="selectedRecipient" class="meta-row">
                <span>{{ selectedRecipient.id }}</span>
                <span>{{ selectedRecipient.emailVerified ? "邮箱已验证" : "邮箱未验证" }}</span>
              </div>
              <p v-else-if="!usersLoading" class="muted">暂无可接收用户。</p>
            </div>
            <div v-if="messageScope === 'role'" class="field"><label for="message-role">接收角色</label><select v-model="targetRole" class="input" id="message-role"><option value="user">注册用户</option><option value="author">作者</option><option value="editor">编辑</option><option value="admin">管理员</option></select></div>
            <div class="field"><label for="message-type">消息类型</label><select v-model="messageType" class="input" id="message-type"><option value="admin">管理员消息</option><option value="review">投稿审核</option><option value="system">站点公告</option><option value="account">系统事件</option></select></div>
            <div class="field"><label for="message-priority">优先级</label><select v-model="priority" class="input" id="message-priority"><option value="normal">普通</option><option value="important">重要</option><option value="urgent">紧急</option></select></div>
            <div class="field"><label for="message-title">标题</label><input v-model="title" class="input" id="message-title"></div>
            <div class="field"><label for="message-content">正文</label><textarea v-model="body" class="input" id="message-content"></textarea></div>
            <div class="field"><label for="message-target">关联目标</label><input v-model="targetTitle" class="input" id="message-target"></div>
            <div v-if="scheduleOpen" class="field"><label for="message-scheduled-at">定时发送</label><input v-model="scheduledAt" class="input" id="message-scheduled-at" type="datetime-local"></div>
            <div class="header-actions">
              <button class="button" type="submit" :disabled="sending">{{ sending ? "发送中..." : scheduleOpen ? "定时发送" : "发送" }}</button>
              <button class="button-secondary" type="button" :disabled="sending" @click="toggleSchedule">{{ scheduleOpen ? "取消定时" : "定时" }}</button>
            </div>
          </form>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
