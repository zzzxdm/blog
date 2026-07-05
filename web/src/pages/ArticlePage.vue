<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";

import {
  ApiError,
  createComment,
  getComments,
  getPosts,
  getReaction,
  getSiteSettings,
  reportComment,
  setBookmark,
  setPostReaction,
  toggleCommentLike,
  type Comment,
  type Post,
  type ReactionSummary,
  type SiteSettings
} from "../shared/api";
import { extractMarkdownHeadings, renderMarkdown } from "../shared/markdown";
import { useAuthStore } from "../stores/auth";
import { usePostsStore } from "../stores/posts";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const posts = usePostsStore();

const post = computed(() => posts.current);
const avatarText = computed(() => post.value?.authorName.slice(0, 1) || "管");
const comments = ref<Comment[]>([]);
const commentTotal = ref(0);
const commentBody = ref("");
const commentsSort = ref<"newest" | "oldest">("newest");
const commentsLoading = ref(false);
const commentError = ref("");
const commentNotice = ref("");
const commentActionId = ref("");
const replyTo = ref<Comment | null>(null);
const relatedPosts = ref<Post[]>([]);
const reaction = ref<ReactionSummary | null>(null);
const siteSettings = ref<SiteSettings | null>(null);
const articleBody = ref<HTMLElement | null>(null);
const readingProgress = ref(0);
const reactionLoading = ref(false);
const reactionError = ref("");
const likeCount = computed(() => reaction.value?.likeCount ?? post.value?.likeCount ?? 0);
const dislikeCount = computed(() => reaction.value?.dislikeCount ?? post.value?.dislikeCount ?? 0);
const bookmarkCount = computed(() => reaction.value?.bookmarkCount ?? 0);
const renderedPostContent = computed(() => renderMarkdown(post.value?.content ?? ""));
const tocItems = computed(() => extractMarkdownHeadings(post.value?.content ?? ""));
const commentsEnabled = computed(() => siteSettings.value?.commentsEnabled ?? true);
const readingProgressEnabled = computed(() => siteSettings.value?.readingProgressEnabled ?? false);
const visibleComments = computed(() => {
  return [...comments.value].sort((left, right) => {
    const leftTime = new Date(left.createdAt).getTime();
    const rightTime = new Date(right.createdAt).getTime();
    return commentsSort.value === "newest" ? rightTime - leftTime : leftTime - rightTime;
  });
});

function load() {
  const slug = String(route.params.slug || "");
  if (slug) {
    void posts.loadBySlug(slug);
    void loadComments(slug);
    void loadReaction(slug);
    void loadRelatedPosts(slug);
  }
}

function back() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  void router.push("/archive");
}

function updateReadingProgress() {
  const target = articleBody.value;
  if (!target) {
    readingProgress.value = 0;
    return;
  }

  const rect = target.getBoundingClientRect();
  const scrollTop = window.scrollY || document.documentElement.scrollTop;
  const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
  const start = scrollTop + rect.top;
  const end = start + target.offsetHeight - viewportHeight * 0.45;
  const total = Math.max(1, end - start);
  const current = scrollTop - start;

  readingProgress.value = Math.min(100, Math.max(0, Math.round((current / total) * 100)));
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}

function authorPath(authorName: string) {
  return `/authors/${encodeURIComponent(authorName)}`;
}

function formatCommentTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function statusText(status: Comment["status"]) {
  if (status === "approved") {
    return "已通过";
  }
  if (status === "pending") {
    return "待审核";
  }
  if (status === "rejected") {
    return "已拒绝";
  }
  if (status === "spam") {
    return "垃圾评论";
  }
  return "已删除";
}

function statusClass(status: Comment["status"]) {
  return status === "approved" ? "published" : "review";
}

function renderedComment(body: string) {
  return renderMarkdown(body);
}

function startReply(comment: Comment) {
  replyTo.value = comment;
  commentError.value = "";
  commentNotice.value = "";
}

function cancelReply() {
  replyTo.value = null;
}

async function loadComments(slug: string) {
  commentsLoading.value = true;
  commentError.value = "";

  try {
    const response = await getComments(slug);
    comments.value = response.items;
    commentTotal.value = response.total;
  } catch (error) {
    commentError.value = error instanceof Error ? error.message : "评论加载失败";
  } finally {
    commentsLoading.value = false;
  }
}

async function loadReaction(slug: string) {
  reactionLoading.value = true;
  reactionError.value = "";

  try {
    reaction.value = await getReaction(slug);
  } catch (error) {
    reactionError.value = error instanceof Error ? error.message : "文章反馈加载失败";
  } finally {
    reactionLoading.value = false;
  }
}

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

async function loadRelatedPosts(slug: string) {
  try {
    const response = await getPosts({ pageSize: 4 });
    relatedPosts.value = response.items.filter((item) => item.slug !== slug).slice(0, 2);
  } catch {
    relatedPosts.value = [];
  }
}

function toggleCommentsSort() {
  commentsSort.value = commentsSort.value === "newest" ? "oldest" : "newest";
}

async function updateReaction(type: "like" | "dislike") {
  if (!post.value) {
    return;
  }
  if (!auth.user) {
    reactionError.value = "请先登录后再操作";
    return;
  }

  reactionLoading.value = true;
  reactionError.value = "";

  try {
    reaction.value = await setPostReaction(post.value.slug, type);
  } catch (error) {
    reactionError.value = error instanceof Error ? error.message : "反馈失败";
  } finally {
    reactionLoading.value = false;
  }
}

async function toggleBookmark() {
  if (!post.value) {
    return;
  }
  if (!auth.user) {
    reactionError.value = "请先登录后再收藏";
    return;
  }

  reactionLoading.value = true;
  reactionError.value = "";

  try {
    reaction.value = await setBookmark(post.value.slug, !reaction.value?.bookmarked);
  } catch (error) {
    reactionError.value = error instanceof Error ? error.message : "收藏失败";
  } finally {
    reactionLoading.value = false;
  }
}

async function submitComment() {
  if (!post.value) {
    return;
  }
  if (!commentsEnabled.value) {
    commentError.value = "评论已关闭";
    return;
  }
  if (!auth.user) {
    commentError.value = "请先登录后再评论";
    return;
  }

  commentError.value = "";
  commentNotice.value = "";

  try {
    const created = await createComment(post.value.slug, commentBody.value, replyTo.value?.id ?? "");
    comments.value = insertComment(comments.value, created);
    commentTotal.value += 1;
    commentBody.value = "";
    replyTo.value = null;
    commentNotice.value = "评论已提交，等待审核。";
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      commentError.value = "登录状态已过期，请重新登录";
      return;
    }
    commentError.value = error instanceof Error ? error.message : "评论提交失败";
  }
}

async function likeComment(comment: Comment) {
  if (!auth.user) {
    commentError.value = "请先登录后再点赞评论";
    return;
  }

  commentActionId.value = comment.id;
  commentError.value = "";
  commentNotice.value = "";

  try {
    const updated = await toggleCommentLike(comment.id);
    comments.value = comments.value.map((item) => (item.id === updated.id ? { ...item, ...updated } : item));
  } catch (error) {
    commentError.value = error instanceof Error ? error.message : "评论点赞失败";
  } finally {
    commentActionId.value = "";
  }
}

async function reportCommentAction(comment: Comment) {
  if (!auth.user) {
    commentError.value = "请先登录后再举报评论";
    return;
  }

  commentActionId.value = comment.id;
  commentError.value = "";
  commentNotice.value = "";

  try {
    await reportComment(comment.id, "读者举报");
    commentNotice.value = "举报已提交，管理员会在后台审核。";
  } catch (error) {
    commentError.value = error instanceof Error ? error.message : "举报提交失败";
  } finally {
    commentActionId.value = "";
  }
}

function insertComment(items: Comment[], created: Comment) {
  if (!created.parentId) {
    return [created, ...items];
  }

  const next = [...items];
  const parentIndex = next.findIndex((item) => item.id === created.parentId);
  if (parentIndex < 0) {
    return [created, ...next];
  }

  next.splice(parentIndex + 1, 0, created);
  return next;
}

onMounted(() => {
  load();
  void loadSiteSettings();
  updateReadingProgress();
  window.addEventListener("scroll", updateReadingProgress, { passive: true });
  window.addEventListener("resize", updateReadingProgress);
});
onBeforeUnmount(() => {
  window.removeEventListener("scroll", updateReadingProgress);
  window.removeEventListener("resize", updateReadingProgress);
});
watch(post, () => {
  window.requestAnimationFrame(updateReadingProgress);
});
watch(() => route.params.slug, () => {
  readingProgress.value = 0;
  load();
});
watch(() => auth.user?.id, () => {
  const slug = String(route.params.slug || "");
  if (slug) {
    void loadComments(slug);
    void loadReaction(slug);
  }
});
</script>

<template>
  <div
    v-if="readingProgressEnabled && post"
    class="reading-progress"
    role="progressbar"
    aria-label="阅读进度"
    aria-valuemin="0"
    aria-valuemax="100"
    :aria-valuenow="readingProgress"
  >
    <span :style="{ width: `${readingProgress}%` }"></span>
  </div>

  <main class="article-shell">
    <p v-if="posts.loading" class="muted">正在加载文章...</p>
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>

    <template v-else-if="post">
      <article>
        <header class="article-hero">
          <div class="article-breadcrumb-row">
            <button class="button-secondary" type="button" @click="back">← 返回</button>
            <nav class="breadcrumb" aria-label="当前位置">
              <RouterLink to="/">首页</RouterLink>
              <span class="breadcrumb-separator">/</span>
              <RouterLink :to="`/archive?category=${encodeURIComponent(post.category)}`">{{ post.category }}</RouterLink>
              <span class="breadcrumb-separator">/</span>
              <span>{{ post.title }}</span>
            </nav>
          </div>
          <div class="meta-row">
            <span class="tag">{{ post.category }}</span>
            <span>{{ post.readingTime }} 分钟阅读</span>
            <span>{{ formatDate(post.publishedAt) }}</span>
          </div>
          <h1>{{ post.title }}</h1>
          <p class="dek">{{ post.summary }}</p>
          <div class="author-row">
            <span class="avatar">{{ avatarText }}</span>
            <div>
              <strong><RouterLink :to="authorPath(post.authorName)">{{ post.authorName }}</RouterLink></strong>
              <div class="meta-row">
                <span>{{ post.tags[0] || "系统设计" }}</span>
                <span>{{ formatNumber(post.viewCount) }} 次阅读</span>
                <span>{{ formatNumber(likeCount) }} 次赞</span>
                <span>{{ formatNumber(dislikeCount) }} 次踩</span>
                <span>{{ formatNumber(commentTotal || post.commentCount) }} 条评论</span>
              </div>
            </div>
          </div>
        </header>

        <figure class="article-cover">
          <img :src="post.coverImage" :alt="post.title">
        </figure>

        <section ref="articleBody" class="article-body">
          <div v-html="renderedPostContent"></div>
        </section>

        <section class="article-feedback" aria-label="文章反馈">
          <div>
            <strong>文章反馈</strong>
            <div class="meta-row">
              <span>{{ formatNumber(likeCount) }} 次赞</span>
              <span>{{ formatNumber(dislikeCount) }} 次踩</span>
              <span>已收藏 {{ formatNumber(bookmarkCount) }} 次</span>
            </div>
          </div>
          <div class="reaction-group">
            <button
              class="reaction-button"
              :class="{ active: reaction?.myReaction === 'like' }"
              type="button"
              aria-label="点赞文章"
              :disabled="reactionLoading"
              @click="updateReaction('like')"
            >
              <span class="reaction-symbol">↑</span>
              <span>{{ formatNumber(likeCount) }}</span>
            </button>
            <button
              class="reaction-button"
              :class="{ active: reaction?.myReaction === 'dislike' }"
              type="button"
              aria-label="点踩文章"
              :disabled="reactionLoading"
              @click="updateReaction('dislike')"
            >
              <span class="reaction-symbol">↓</span>
              <span>{{ formatNumber(dislikeCount) }}</span>
            </button>
            <button class="button-secondary" type="button" :disabled="reactionLoading" @click="toggleBookmark">
              {{ reaction?.bookmarked ? "取消收藏" : "收藏" }}
            </button>
          </div>
          <p v-if="reactionError" class="error">{{ reactionError }}</p>
        </section>

        <section class="comments" aria-label="评论">
          <div class="section-heading">
            <div>
              <h2>评论</h2>
              <p>{{ commentTotal || post.commentCount }} 条讨论，评论提交后进入审核队列。</p>
            </div>
            <button class="button-secondary" type="button" @click="toggleCommentsSort">{{ commentsSort === "newest" ? "最新在前" : "最早在前" }}</button>
          </div>
          <div v-if="commentsEnabled" class="comment-box">
            <div class="author-row">
              <span class="avatar">{{ auth.user?.avatarText || "访" }}</span>
              <div>
                <strong>{{ auth.user?.displayName || "访客" }}</strong>
                <div class="meta-row">
                  <span>{{ auth.user ? "已登录" : "未登录" }}</span>
                  <RouterLink to="/account/comments">查看我的评论</RouterLink>
                </div>
              </div>
            </div>
            <div v-if="replyTo" class="review-note">
              <strong>回复 {{ replyTo.authorName }}</strong>
              <p>{{ replyTo.body }}</p>
              <button class="button-secondary" type="button" @click="cancelReply">取消回复</button>
            </div>
            <textarea v-model="commentBody" placeholder="写下你的想法，支持 Markdown 基础语法"></textarea>
            <div class="meta-row">
              <button class="button" type="button" :disabled="!commentBody.trim()" @click="submitComment">{{ replyTo ? "提交回复" : "提交评论" }}</button>
              <span>评论提交后进入审核队列。</span>
              <RouterLink v-if="!auth.user" to="/login">去登录</RouterLink>
            </div>
            <p v-if="commentError" class="error">{{ commentError }}</p>
            <p v-else-if="commentNotice" class="muted">{{ commentNotice }}</p>
          </div>
          <div v-else class="comment-box">
            <strong>评论已关闭</strong>
            <p class="muted">管理员暂时关闭了新评论，历史评论仍可阅读。</p>
          </div>

          <div class="comment-list">
            <p v-if="commentsLoading" class="muted">正在加载评论...</p>
            <template v-else>
              <article
                v-for="comment in visibleComments"
                :key="comment.id"
                class="comment-item"
                :class="{ reply: comment.parentId }"
              >
                <div class="comment-head">
                  <div class="author-row">
                    <span class="avatar">{{ comment.avatarText }}</span>
                    <div>
                      <strong>{{ comment.authorName }}</strong>
                      <div class="meta-row">
                        <span>{{ formatCommentTime(comment.createdAt) }}</span>
                        <span v-if="comment.isMine">我的评论</span>
                        <span v-if="comment.isAuthor">作者</span>
                      </div>
                    </div>
                  </div>
                  <span class="status" :class="statusClass(comment.status)">{{ statusText(comment.status) }}</span>
                </div>
                <div class="comment-body" v-html="renderedComment(comment.body)"></div>
                <div class="comment-actions">
                  <button type="button" :disabled="commentActionId === comment.id" @click="likeComment(comment)">
                    {{ comment.liked ? "已赞" : "点赞" }} {{ comment.likeCount }}
                  </button>
                  <button type="button" @click="startReply(comment)">回复{{ comment.replyCount ? ` ${comment.replyCount}` : "" }}</button>
                  <RouterLink v-if="comment.isMine" to="/account/comments">管理</RouterLink>
                  <button v-else type="button" :disabled="commentActionId === comment.id" @click="reportCommentAction(comment)">举报</button>
                </div>
              </article>
            </template>
          </div>
        </section>
      </article>

      <aside class="toc" aria-label="文章目录">
        <section v-if="tocItems.length" class="panel">
          <div class="panel-title">
            <h2>目录</h2>
          </div>
          <nav>
            <a
              v-for="(heading, index) in tocItems"
              :key="`${heading.id}-${index}`"
              :class="{ active: index === 0, nested: heading.level > 2 }"
              :href="`#${heading.id}`"
            >
              {{ heading.text }}
            </a>
          </nav>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>作者</h2>
          </div>
          <div class="author-row">
            <span class="avatar">{{ avatarText }}</span>
            <div>
              <strong><RouterLink :to="authorPath(post.authorName)">{{ post.authorName }}</RouterLink></strong>
              <div class="meta-row">
                <span>{{ post.category }}</span>
                <span>{{ post.tags.length }} 个标签</span>
              </div>
            </div>
          </div>
        </section>

        <section v-if="relatedPosts.length" class="panel">
          <div class="panel-title">
            <h2>相关文章</h2>
          </div>
          <ul class="link-list">
            <li v-for="item in relatedPosts" :key="item.slug">
              <RouterLink :to="`/posts/${item.slug}`">
                <strong>{{ item.title }}</strong>
                <span>{{ item.category }} · {{ item.readingTime }} 分钟</span>
              </RouterLink>
            </li>
          </ul>
        </section>
      </aside>
    </template>
  </main>
</template>
