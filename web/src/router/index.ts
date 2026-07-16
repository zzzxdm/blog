import { createRouter, createWebHistory } from "vue-router";

import { useAuthStore } from "../stores/auth";

// Keep the home page eager for first paint; lazy-load everything else so admin,
// account, editor and markdown chunks stay out of the critical path.
const HomePage = () => import("../pages/HomePage.vue");
const ArchivePage = () => import("../pages/ArchivePage.vue");
const TopicsPage = () => import("../pages/TopicsPage.vue");
const SearchPage = () => import("../pages/SearchPage.vue");
const AboutPage = () => import("../pages/AboutPage.vue");
const AuthorPage = () => import("../pages/AuthorPage.vue");
const PreviewPage = () => import("../pages/PreviewPage.vue");
const ArticlePage = () => import("../pages/ArticlePage.vue");
const LoginPage = () => import("../pages/LoginPage.vue");
const SubmitPage = () => import("../pages/SubmitPage.vue");
const NotFoundPage = () => import("../pages/NotFoundPage.vue");

const AccountPage = () => import("../pages/AccountPage.vue");
const AccountBookmarksPage = () => import("../pages/account/AccountBookmarksPage.vue");
const AccountCommentsPage = () => import("../pages/account/AccountCommentsPage.vue");
const AccountMessagesPage = () => import("../pages/account/AccountMessagesPage.vue");
const AccountPrivatePostsPage = () => import("../pages/account/AccountPrivatePostsPage.vue");
const AccountSettingsPage = () => import("../pages/account/AccountSettingsPage.vue");
const AccountSubmissionsPage = () => import("../pages/account/AccountSubmissionsPage.vue");

const AdminHome = () => import("../pages/admin/AdminHome.vue");
const AdminPostsPage = () => import("../pages/admin/AdminPostsPage.vue");
const AdminSubmissionsPage = () => import("../pages/admin/AdminSubmissionsPage.vue");
const AdminEditorPage = () => import("../pages/admin/AdminEditorPage.vue");
const AdminTaxonomiesPage = () => import("../pages/admin/AdminTaxonomiesPage.vue");
const AdminTopicsPage = () => import("../pages/admin/AdminTopicsPage.vue");
const AdminCommentsPage = () => import("../pages/admin/AdminCommentsPage.vue");
const AdminUsersPage = () => import("../pages/admin/AdminUsersPage.vue");
const AdminMessagesPage = () => import("../pages/admin/AdminMessagesPage.vue");
const AdminMediaPage = () => import("../pages/admin/AdminMediaPage.vue");
const AdminNavigationPage = () => import("../pages/admin/AdminNavigationPage.vue");
const AdminRedirectsPage = () => import("../pages/admin/AdminRedirectsPage.vue");
const AdminImportExportPage = () => import("../pages/admin/AdminImportExportPage.vue");
const AdminStatsPage = () => import("../pages/admin/AdminStatsPage.vue");
const AdminAuditPage = () => import("../pages/admin/AdminAuditPage.vue");
const AdminSettingsPage = () => import("../pages/admin/AdminSettingsPage.vue");

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: HomePage },
    { path: "/archive", name: "archive", component: ArchivePage },
    { path: "/topics", name: "topics", component: TopicsPage },
    { path: "/search", name: "search", component: SearchPage },
    { path: "/about", name: "about", component: AboutPage },
    { path: "/authors/:id", name: "author", component: AuthorPage },
    { path: "/preview/:token", name: "preview", component: PreviewPage, meta: { hideChrome: true } },
    { path: "/posts/:slug", name: "post", component: ArticlePage },
    { path: "/login", name: "login", component: LoginPage, meta: { hideChrome: true } },
    { path: "/submit", name: "submit", component: SubmitPage, meta: { requiresAuth: true } },
    { path: "/account", name: "account", component: AccountPage, meta: { requiresAuth: true } },
    { path: "/account/comments", name: "account-comments", component: AccountCommentsPage, meta: { requiresAuth: true } },
    { path: "/account/bookmarks", name: "account-bookmarks", component: AccountBookmarksPage, meta: { requiresAuth: true } },
    { path: "/account/private-posts", name: "account-private-posts", component: AccountPrivatePostsPage, meta: { requiresAuth: true } },
    { path: "/account/submissions", name: "account-submissions", component: AccountSubmissionsPage, meta: { requiresAuth: true } },
    { path: "/account/messages", name: "account-messages", component: AccountMessagesPage, meta: { requiresAuth: true } },
    { path: "/account/settings", name: "account-settings", component: AccountSettingsPage, meta: { requiresAuth: true } },
    { path: "/admin", name: "admin", component: AdminHome, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/posts", name: "admin-posts", component: AdminPostsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/submissions", name: "admin-submissions", component: AdminSubmissionsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/editor", name: "admin-editor", component: AdminEditorPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/taxonomies", name: "admin-taxonomies", component: AdminTaxonomiesPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/categories", name: "admin-categories", component: AdminTaxonomiesPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/tags", name: "admin-tags", component: AdminTaxonomiesPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/topics", name: "admin-topics", component: AdminTopicsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/comments", name: "admin-comments", component: AdminCommentsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/users", name: "admin-users", component: AdminUsersPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/messages", name: "admin-messages", component: AdminMessagesPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/media", name: "admin-media", component: AdminMediaPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/navigation", name: "admin-navigation", component: AdminNavigationPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/redirects", name: "admin-redirects", component: AdminRedirectsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/import-export", name: "admin-import-export", component: AdminImportExportPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/stats", name: "admin-stats", component: AdminStatsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/statistics", name: "admin-statistics", component: AdminStatsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/audit", name: "admin-audit", component: AdminAuditPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/audit-logs", name: "admin-audit-logs", component: AdminAuditPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/admin/settings", name: "admin-settings", component: AdminSettingsPage, meta: { hideChrome: true, requiresAdmin: true } },
    { path: "/:pathMatch(.*)*", name: "not-found", component: NotFoundPage }
  ],
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    }
    // 专题页的 #topic-reading 要等数据渲染完再滚，交由 TopicsPage 处理。
    if (to.path === "/topics" && to.hash === "#topic-reading") {
      return false;
    }
    if (to.hash) {
      return new Promise((resolve) => {
        const tryScroll = (attempt: number) => {
          const el = document.querySelector(to.hash);
          if (el) {
            resolve({ el: to.hash, top: 24 });
            return;
          }
          if (attempt >= 20) {
            resolve({ top: 0 });
            return;
          }
          window.setTimeout(() => tryScroll(attempt + 1), 50);
        };
        tryScroll(0);
      });
    }
    return { top: 0 };
  }
});

router.beforeEach(async (to) => {
  const requiresAuth = Boolean(to.meta.requiresAuth || to.meta.requiresAdmin);
  const requiresAdmin = Boolean(to.meta.requiresAdmin);
  if (!requiresAuth) {
    return true;
  }

  const auth = useAuthStore();
  if (!auth.user) {
    await auth.loadMe();
  }

  if (!auth.user) {
    return { name: "login", query: { redirect: to.fullPath } };
  }

  if (requiresAdmin && auth.user.role !== "admin") {
    return { name: "home" };
  }

  return true;
});
