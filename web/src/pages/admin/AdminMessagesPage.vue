<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  createAdminMessage,
  exportAdminMessages,
  getAdminMessages,
  type MessageStats,
  type MessageType,
  type StationMessage
} from "../../shared/api";
import { downloadJson, exportFileName } from "../../shared/download";

const messages = ref<StationMessage[]>([]);
const stats = ref<MessageStats>({ unread: 0, review: 0, admin: 0, archived: 0, total: 0 });
const loading = ref(false);
const sending = ref(false);
const exporting = ref(false);
const error = ref("");
const message = ref("");
const selectedId = ref("");

const recipientId = ref("user_linyi");
const recipientName = ref("林一");
const messageType = ref<MessageType>("admin");
const priority = ref("normal");
const title = ref("优质投稿用户邀请");
const body = ref("你最近的投稿质量较高，欢迎继续提交工程实践和写作工作流相关内容。");
const targetTitle = ref("投稿激励");
const scheduleOpen = ref(false);
const scheduledAt = ref("");
const searchQuery = ref("");
const typeFilter = ref("");
const statusFilter = ref("");

const selectedMessage = computed(() => messages.value.find((item) => item.id === selectedId.value) || messages.value[0]);
const visibleMessages = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  return messages.value.filter((item) => {
    const matchesKeyword = !keyword || [
      item.title,
      item.body,
      item.recipientName,
      item.recipientId,
      item.targetTitle || "",
      typeText(item.type),
      statusText(item.status)
    ].join(" ").toLowerCase().includes(keyword);
    const matchesType = typeFilter.value === "" || item.type === typeFilter.value;
    const matchesStatus = statusFilter.value === ""
      || (statusFilter.value === "sent" ? item.status !== "scheduled" && item.status !== "archived" : item.status === statusFilter.value);

    return matchesKeyword && matchesType && matchesStatus;
  });
});

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminMessages();
    messages.value = response.items;
    stats.value = response.stats;
    if (!messages.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = messages.value[0]?.id || "";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信记录加载失败";
  } finally {
    loading.value = false;
  }
}

async function send() {
  if (scheduleOpen.value && !scheduledAt.value) {
    error.value = "请选择定时发送时间";
    return;
  }

  sending.value = true;
  error.value = "";
  message.value = "";

  try {
    const scheduleValue = scheduleOpen.value ? new Date(scheduledAt.value).toISOString() : "";
    await createAdminMessage({
      recipientId: recipientId.value,
      recipientName: recipientName.value,
      type: messageType.value,
      priority: priority.value,
      title: title.value,
      body: body.value,
      targetType: "admin-message",
      targetTitle: targetTitle.value,
      scheduledAt: scheduleValue || undefined
    });
    message.value = scheduleValue && new Date(scheduleValue) > new Date()
      ? `站内信已定时到 ${formatTime(scheduleValue)}。`
      : "站内信已发送。";
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
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" :disabled="exporting" @click="exportMessages">{{ exporting ? "导出中..." : "导出记录" }}</button>
        <button class="button" type="button" :disabled="sending" @click="send">新建消息</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="站内信统计">
      <div class="stat-card"><span>总发送</span><strong>{{ stats.total }}</strong></div>
      <div class="stat-card"><span>当前未读</span><strong>{{ stats.unread }}</strong></div>
      <div class="stat-card"><span>审核消息</span><strong>{{ stats.review }}</strong></div>
      <div class="stat-card"><span>管理员消息</span><strong>{{ stats.admin }}</strong></div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="admin-grid-2">
      <section class="table-panel" aria-label="发送记录">
        <form class="table-toolbar" @submit.prevent="load">
          <input v-model="searchQuery" class="input" type="search" placeholder="搜索标题、接收人、类型" aria-label="搜索站内信">
          <select v-model="typeFilter" class="input" aria-label="消息类型">
            <option value="">全部类型</option>
            <option value="admin">管理员消息</option>
            <option value="review">投稿审核</option>
            <option value="system">站点公告</option>
            <option value="account">系统事件</option>
            <option value="comment">评论</option>
          </select>
          <select v-model="statusFilter" class="input" aria-label="发送状态">
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
            <tr v-for="item in visibleMessages" :key="item.id">
              <td><strong>{{ item.title }}</strong><div class="meta-row"><span>{{ item.targetTitle || item.body }}</span></div></td>
              <td>{{ item.recipientName }}</td>
              <td><span class="status" :class="typeClass(item.type)">{{ typeText(item.type) }}</span></td>
              <td>{{ item.status === "scheduled" ? "定时中" : "1 / 1" }}</td>
              <td>{{ item.status === "scheduled" ? "-" : item.status === "unread" ? "0 / 1" : "1 / 1" }}</td>
              <td>{{ formatTime(item.scheduledAt || item.createdAt) }}</td>
              <td><button class="button-secondary" type="button" @click="viewMessage(item)">查看</button></td>
            </tr>
            <tr v-if="visibleMessages.length === 0">
              <td colspan="7" class="muted">没有匹配的站内信。</td>
            </tr>
          </tbody>
        </table>
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
            <div class="field"><label for="message-scope">接收范围</label><select class="input" id="message-scope"><option>指定用户</option><option>全部注册用户</option><option>按角色发送</option><option>按用户筛选条件发送</option></select></div>
            <div class="field"><label for="message-recipient">接收人 ID</label><input v-model="recipientId" class="input" id="message-recipient"></div>
            <div class="field"><label for="message-recipient-name">接收人名称</label><input v-model="recipientName" class="input" id="message-recipient-name"></div>
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
