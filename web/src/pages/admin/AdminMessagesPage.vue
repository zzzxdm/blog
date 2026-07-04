<script setup lang="ts">
import { onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  createAdminMessage,
  getAdminMessages,
  type MessageStats,
  type MessageType,
  type StationMessage
} from "../../shared/api";

const messages = ref<StationMessage[]>([]);
const stats = ref<MessageStats>({ unread: 0, review: 0, admin: 0, archived: 0, total: 0 });
const loading = ref(false);
const sending = ref(false);
const error = ref("");
const message = ref("");

const recipientId = ref("user_linyi");
const recipientName = ref("林一");
const messageType = ref<MessageType>("admin");
const priority = ref("normal");
const title = ref("优质投稿用户邀请");
const body = ref("你最近的投稿质量较高，欢迎继续提交工程实践和写作工作流相关内容。");
const targetTitle = ref("投稿激励");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminMessages();
    messages.value = response.items;
    stats.value = response.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信记录加载失败";
  } finally {
    loading.value = false;
  }
}

async function send() {
  sending.value = true;
  error.value = "";
  message.value = "";

  try {
    await createAdminMessage({
      recipientId: recipientId.value,
      recipientName: recipientName.value,
      type: messageType.value,
      priority: priority.value,
      title: title.value,
      body: body.value,
      targetType: "admin-message",
      targetTitle: targetTitle.value
    });
    message.value = "站内信已发送。";
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信发送失败";
  } finally {
    sending.value = false;
  }
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
</script>

<template>
  <AdminLayout title="站内信管理" description="向用户发送审核结果、站点公告和定向运营消息。" mobile-title="站内信管理" primary-action="发送">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">导出记录</button>
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
          <input class="input" type="search" placeholder="搜索标题、接收人、类型" aria-label="搜索站内信">
          <select class="input" aria-label="消息类型">
            <option>全部类型</option>
            <option>管理员消息</option>
            <option>投稿审核</option>
            <option>站点公告</option>
            <option>系统事件</option>
          </select>
          <select class="input" aria-label="发送状态">
            <option>全部状态</option>
            <option>已发送</option>
            <option>定时中</option>
            <option>已撤回</option>
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
              <td>1 / 1</td>
              <td>{{ item.status === "unread" ? "0 / 1" : "1 / 1" }}</td>
              <td>{{ formatTime(item.createdAt) }}</td>
              <td><button class="button-secondary" type="button">查看</button></td>
            </tr>
          </tbody>
        </table>
      </section>

      <aside class="panel">
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
          <div class="header-actions">
            <button class="button" type="submit" :disabled="sending">{{ sending ? "发送中..." : "发送" }}</button>
            <button class="button-secondary" type="button">定时</button>
          </div>
        </form>
      </aside>
    </section>
  </AdminLayout>
</template>
