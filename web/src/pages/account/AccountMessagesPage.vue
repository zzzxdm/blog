<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  archiveMessage,
  getMessages,
  markAllMessagesRead,
  markMessageRead,
  type MessageStats,
  type MessageType,
  type StationMessage
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useMessageStore } from "../../stores/messages";
import { useToastStore } from "../../stores/toast";

const messageStore = useMessageStore();
const toast = useToastStore();
const messages = ref<StationMessage[]>([]);
const stats = ref<MessageStats>({ unread: 0, review: 0, admin: 0, archived: 0, scheduled: 0, total: 0 });
const selectedId = ref("");
const filterStatus = ref("");
const filterType = ref("");
const loading = ref(false);
const refreshing = ref(false);
const readAllLoading = ref(false);
const readLoading = ref(false);
const archiveLoading = ref(false);
const error = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const selected = computed(() => messages.value.find((item) => item.id === selectedId.value) || messages.value[0]);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getMessages({
      status: filterStatus.value,
      type: filterType.value,
      page: page.value,
      pageSize: pageSize.value
    });
    messages.value = response.items;
    stats.value = response.stats;
    messageStore.setUnread(response.stats.unread);
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
    if (!messages.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = messages.value[0]?.id || "";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "站内信加载失败";
  } finally {
    loading.value = false;
  }
}

async function readAll() {
  if (stats.value.unread === 0) {
    toast.info("没有未读站内信", "当前收件箱已经全部处理。");
    return;
  }

  readAllLoading.value = true;
  error.value = "";

  try {
    const response = await markAllMessagesRead();
    stats.value = response.stats;
    messageStore.setUnread(response.stats.unread);
    messages.value = messages.value.map((item) => ({ ...item, status: item.status === "archived" ? item.status : "read" }));
    toast.success("已全部标记为已读", "未归档站内信已更新。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "标记已读失败";
    toast.error("标记已读失败", error.value);
  } finally {
    readAllLoading.value = false;
  }
}

async function readSelected() {
  if (!selected.value) {
    return;
  }
  if (selected.value.status !== "unread") {
    toast.info("这条消息已是已读", selected.value.title);
    return;
  }

  readLoading.value = true;
  error.value = "";
  try {
    const updated = await markMessageRead(selected.value.id);
    patchMessage(updated);
    await load();
    toast.success("已标记为已读", updated.title);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "标记已读失败";
    toast.error("标记已读失败", error.value);
  } finally {
    readLoading.value = false;
  }
}

async function archiveSelected() {
  if (!selected.value) {
    return;
  }
  if (selected.value.status === "archived") {
    toast.info("这条消息已归档", selected.value.title);
    return;
  }

  archiveLoading.value = true;
  error.value = "";
  try {
    const updated = await archiveMessage(selected.value.id);
    patchMessage(updated);
    await load();
    toast.success("消息已归档", updated.title);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "归档失败";
    toast.error("归档失败", error.value);
  } finally {
    archiveLoading.value = false;
  }
}

function patchMessage(updated: StationMessage) {
  messages.value = messages.value.map((item) => item.id === updated.id ? updated : item);
}

function chooseStatus(value: string) {
  filterStatus.value = value;
  filterType.value = "";
  page.value = 1;
  void load();
}

function chooseType(value: string) {
  filterStatus.value = "";
  filterType.value = value;
  page.value = 1;
  void load();
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

async function refreshList() {
  refreshing.value = true;
  await load();
  refreshing.value = false;
  if (error.value) {
    toast.error("刷新失败", error.value);
  } else {
    toast.success("站内信已刷新", `当前显示 ${messages.value.length} 条消息。`);
  }
}

function formatTime(value: string) {
  return formatDateTime(value);
}

function typeText(value: MessageType) {
  if (value === "review") {
    return "审核";
  }
  if (value === "comment") {
    return "评论";
  }
  if (value === "system") {
    return "公告";
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
  <AccountLayout title="站内信" description="查看投稿审核结果、评论回复、账号提醒和管理员发送的消息。">
    <template #actions>
      <button class="button-secondary" type="button" :disabled="readAllLoading || loading" @click="readAll">
        {{ readAllLoading ? "处理中..." : "全部已读" }}
      </button>
    </template>

    <section class="stats-grid" aria-label="站内信统计">
      <div class="stat-card"><span>未读</span><strong>{{ stats.unread }}</strong></div>
      <div class="stat-card"><span>审核消息</span><strong>{{ stats.review }}</strong></div>
      <div class="stat-card"><span>管理员消息</span><strong>{{ stats.admin }}</strong></div>
      <div class="stat-card"><span>已归档</span><strong>{{ stats.archived }}</strong></div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>

    <section class="message-layout">
      <div class="panel">
        <div class="panel-title">
          <h2>收件箱</h2>
          <button class="button-secondary" type="button" :disabled="refreshing || loading" @click="refreshList">
            {{ refreshing ? "刷新中..." : "刷新" }}
          </button>
        </div>
        <div class="message-filterbar" aria-label="消息筛选">
          <a :class="{ active: !filterStatus && !filterType }" href="#all" @click.prevent="chooseStatus('')">全部</a>
          <a :class="{ active: filterStatus === 'unread' }" href="#unread" @click.prevent="chooseStatus('unread')">未读</a>
          <a :class="{ active: filterType === 'review' }" href="#review" @click.prevent="chooseType('review')">审核</a>
          <a :class="{ active: filterType === 'comment' }" href="#comment" @click.prevent="chooseType('comment')">评论</a>
          <a :class="{ active: filterType === 'system' }" href="#system" @click.prevent="chooseType('system')">系统</a>
          <a :class="{ active: filterType === 'account' }" href="#account" @click.prevent="chooseType('account')">账号</a>
          <a :class="{ active: filterType === 'admin' }" href="#admin" @click.prevent="chooseType('admin')">管理员</a>
          <a :class="{ active: filterStatus === 'archived' }" href="#archived" @click.prevent="chooseStatus('archived')">归档</a>
        </div>
        <LoadingState v-if="loading" variant="table" text="正在加载站内信..." :rows="4" />
        <div v-else class="message-list">
          <a
            v-for="item in messages"
            :key="item.id"
            class="message-item"
            :class="{ unread: item.status === 'unread', active: selected?.id === item.id }"
            href="#detail"
            @click.prevent="selectedId = item.id"
          >
            <div class="meta-row"><span class="status" :class="typeClass(item.type)">{{ typeText(item.type) }}</span><span>{{ formatTime(item.createdAt) }}</span></div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.body }}</p>
          </a>
          <p v-if="messages.length === 0" class="muted">暂无站内信。</p>
        </div>
        <PaginationControls
          v-if="!loading && !error"
          :page="page"
          :page-size="pageSize"
          :total="total"
          :loading="loading"
          item-label="条消息"
          show-page-size
          :page-size-options="[5, 10, 20, 50, 100]"
          @update:page="setPage"
          @update:page-size="setPageSize"
        />
      </div>

      <article v-if="selected" class="panel message-detail" id="detail">
        <div class="meta-row">
          <span class="status" :class="typeClass(selected.type)">{{ typeText(selected.type) }}</span>
          <span>{{ selected.senderName }}</span>
          <span>{{ formatTime(selected.createdAt) }}</span>
        </div>
        <div class="message-body">
          <h2>{{ selected.title }}</h2>
          <p>{{ selected.body }}</p>
          <blockquote v-if="selected.targetTitle">{{ selected.targetTitle }}</blockquote>
          <p v-if="selected.status === 'unread'">这条消息尚未标记为已读。</p>
        </div>
        <div class="header-actions">
          <RouterLink v-if="selected.targetType === 'submission'" class="button" to="/account/submissions">查看投稿</RouterLink>
          <button class="button-secondary" type="button" :disabled="readLoading || selected.status !== 'unread'" @click="readSelected">
            {{ readLoading ? "处理中..." : selected.status === "unread" ? "标记已读" : "已读" }}
          </button>
          <button class="button-secondary" type="button" :disabled="archiveLoading || selected.status === 'archived'" @click="archiveSelected">
            {{ archiveLoading ? "归档中..." : selected.status === "archived" ? "已归档" : "归档" }}
          </button>
        </div>
      </article>
    </section>
  </AccountLayout>
</template>
