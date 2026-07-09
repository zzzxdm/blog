import { createRouter, createWebHistory } from "vue-router";

import AccountPage from "../pages/AccountPage.vue";
import AccountBookmarksPage from "../pages/account/AccountBookmarksPage.vue";
import AccountCommentsPage from "../pages/account/AccountCommentsPage.vue";
import AccountMessagesPage from "../pages/account/AccountMessagesPage.vue";
import AccountPrivatePostsPage from "../pages/account/AccountPrivatePostsPage.vue";
import AccountSettingsPage from "../pages/account/AccountSettingsPage.vue";
import AccountSubmissionsPage from "../pages/account/AccountSubmissionsPage.vue";
import AdminAuditPage from "../pages/admin/AdminAuditPage.vue";
import AdminCommentsPage from "../pages/admin/AdminCommentsPage.vue";
import AdminEditorPage from "../pages/admin/AdminEditorPage.vue";
import AdminHome from "../pages/admin/AdminHome.vue";
import AdminImportExportPage from "../pages/admin/AdminImportExportPage.vue";
import AdminMediaPage from "../pages/admin/AdminMediaPage.vue";
import AdminMessagesPage from "../pages/admin/AdminMessagesPage.vue";
import AdminNavigationPage from "../pages/admin/AdminNavigationPage.vue";
import AdminPostsPage from "../pages/admin/AdminPostsPage.vue";
import AdminRedirectsPage from "../pages/admin/AdminRedirectsPage.vue";
import AdminSettingsPage from "../pages/admin/AdminSettingsPage.vue";
import AdminStatsPage from "../pages/admin/AdminStatsPage.vue";
import AdminSubmissionsPage from "../pages/admin/AdminSubmissionsPage.vue";
import AdminTaxonomiesPage from "../pages/admin/AdminTaxonomiesPage.vue";
import AdminUsersPage from "../pages/admin/AdminUsersPage.vue";
import ArchivePage from "../pages/ArchivePage.vue";
import ArticlePage from "../pages/ArticlePage.vue";
import AboutPage from "../pages/AboutPage.vue";
import AuthorPage from "../pages/AuthorPage.vue";
import HomePage from "../pages/HomePage.vue";
import LoginPage from "../pages/LoginPage.vue";
import NotFoundPage from "../pages/NotFoundPage.vue";
import PreviewPage from "../pages/PreviewPage.vue";
import SearchPage from "../pages/SearchPage.vue";
import SubmitPage from "../pages/SubmitPage.vue";
import TopicsPage from "../pages/TopicsPage.vue";
import { useAuthStore } from "../stores/auth";

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
    { path: "/admin/topics", name: "admin-topics", component: () => import("../pages/admin/AdminTopicsPage.vue"), meta: { hideChrome: true, requiresAdmin: true } },
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
    if (to.hash) {
      return { el: to.hash, top: 24 };
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
