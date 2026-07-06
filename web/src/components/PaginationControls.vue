<script setup lang="ts">
import { computed, ref, watch } from "vue";

const props = withDefaults(defineProps<{
  page: number;
  pageSize: number;
  total: number;
  loading?: boolean;
  itemLabel?: string;
  pageSizeOptions?: number[];
  showPageSize?: boolean;
  showPageNumbers?: boolean;
}>(), {
  loading: false,
  itemLabel: "条记录",
  pageSizeOptions: () => [10, 20, 50, 100],
  showPageSize: false,
  showPageNumbers: true
});

const emit = defineEmits<{
  "update:page": [page: number];
  "update:pageSize": [pageSize: number];
}>();

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)));
const start = computed(() => props.total > 0 ? (props.page - 1) * props.pageSize + 1 : 0);
const end = computed(() => Math.min(props.page * props.pageSize, props.total));
const jumpPage = ref(props.page);
const pageItems = computed(() => {
  if (totalPages.value <= 5) {
    return Array.from({ length: totalPages.value }, (_, index) => index + 1);
  }

  const pages = [1, props.page - 1, props.page, props.page + 1, totalPages.value]
    .filter((page) => page >= 1 && page <= totalPages.value)
    .filter((page, index, list) => list.indexOf(page) === index)
    .sort((left, right) => left - right);

  return pages.flatMap((page, index) => {
    if (index > 0 && page - pages[index - 1] > 1) {
      return [`ellipsis-${page}`, page];
    }
    return [page];
  });
});

function go(nextPage: number) {
  const page = Math.min(Math.max(nextPage, 1), totalPages.value);
  if (page !== props.page) {
    emit("update:page", page);
  }
}

function changePageSize(value: string) {
  const pageSize = Number.parseInt(value, 10);
  if (Number.isFinite(pageSize) && pageSize > 0 && pageSize !== props.pageSize) {
    emit("update:pageSize", pageSize);
  }
}

function submitJump() {
  go(jumpPage.value);
}

watch(() => props.page, (page) => {
  jumpPage.value = page;
});
</script>

<template>
  <nav class="pagination" aria-label="分页">
    <label v-if="showPageSize" class="pagination-size">
      <span>每页</span>
      <select class="input" :value="pageSize" :disabled="loading" @change="changePageSize(($event.target as HTMLSelectElement).value)">
        <option v-for="option in pageSizeOptions" :key="option" :value="option">{{ option }}</option>
      </select>
    </label>
    <button
      class="page-button"
      :class="{ disabled: page <= 1 || loading }"
      type="button"
      :disabled="page <= 1 || loading"
      aria-label="上一页"
      @click="go(page - 1)"
    >
      ←
    </button>
    <template v-if="showPageNumbers && totalPages > 1">
      <button
        v-for="item in pageItems"
        :key="item"
        class="page-button"
        :class="{ current: page === item, disabled: typeof item !== 'number' || loading }"
        type="button"
        :disabled="typeof item !== 'number' || loading"
        @click="typeof item === 'number' && go(item)"
      >
        {{ typeof item === "number" ? item : "..." }}
      </button>
    </template>
    <span class="pagination-summary">
      第 {{ page }} / {{ totalPages }} 页，显示 {{ start }}-{{ end }} / {{ total }} {{ itemLabel }}
    </span>
    <form v-if="totalPages > 1" class="pagination-jump" @submit.prevent="submitJump">
      <label for="pagination-jump-page">跳至</label>
      <input
        v-model.number="jumpPage"
        class="input"
        id="pagination-jump-page"
        type="number"
        min="1"
        :max="totalPages"
        :disabled="loading"
      >
      <button class="button-secondary" type="submit" :disabled="loading">确定</button>
    </form>
    <button
      class="page-button"
      :class="{ disabled: page >= totalPages || loading }"
      type="button"
      :disabled="page >= totalPages || loading"
      aria-label="下一页"
      @click="go(page + 1)"
    >
      →
    </button>
  </nav>
</template>
