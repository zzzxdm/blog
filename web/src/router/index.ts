import { createRouter, createWebHistory } from "vue-router";

import AccountPage from "../pages/AccountPage.vue";
import AdminCommentsPage from "../pages/admin/AdminCommentsPage.vue";
import AdminEditorPage from "../pages/admin/AdminEditorPage.vue";
import AdminHome from "../pages/admin/AdminHome.vue";
import AdminMediaPage from "../pages/admin/AdminMediaPage.vue";
import AdminMessagesPage from "../pages/admin/AdminMessagesPage.vue";
import AdminNavigationPage from "../pages/admin/AdminNavigationPage.vue";
import AdminPostsPage from "../pages/admin/AdminPostsPage.vue";
import AdminSettingsPage from "../pages/admin/AdminSettingsPage.vue";
import AdminStatsPage from "../pages/admin/AdminStatsPage.vue";
import AdminSubmissionsPage from "../pages/admin/AdminSubmissionsPage.vue";
import AdminUsersPage from "../pages/admin/AdminUsersPage.vue";
import ArchivePage from "../pages/ArchivePage.vue";
import ArticlePage from "../pages/ArticlePage.vue";
import HomePage from "../pages/HomePage.vue";
import LoginPage from "../pages/LoginPage.vue";
import SubmitPage from "../pages/SubmitPage.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: HomePage },
    { path: "/archive", name: "archive", component: ArchivePage },
    { path: "/posts/:slug", name: "post", component: ArticlePage },
    { path: "/login", name: "login", component: LoginPage, meta: { hideChrome: true } },
    { path: "/submit", name: "submit", component: SubmitPage },
    { path: "/account", name: "account", component: AccountPage },
    { path: "/account/comments", name: "account-comments", component: AccountPage },
    { path: "/account/bookmarks", name: "account-bookmarks", component: AccountPage },
    { path: "/account/submissions", name: "account-submissions", component: AccountPage },
    { path: "/account/messages", name: "account-messages", component: AccountPage },
    { path: "/account/settings", name: "account-settings", component: AccountPage },
    { path: "/admin", name: "admin", component: AdminHome, meta: { hideChrome: true } },
    { path: "/admin/posts", name: "admin-posts", component: AdminPostsPage, meta: { hideChrome: true } },
    { path: "/admin/submissions", name: "admin-submissions", component: AdminSubmissionsPage, meta: { hideChrome: true } },
    { path: "/admin/editor", name: "admin-editor", component: AdminEditorPage, meta: { hideChrome: true } },
    { path: "/admin/comments", name: "admin-comments", component: AdminCommentsPage, meta: { hideChrome: true } },
    { path: "/admin/users", name: "admin-users", component: AdminUsersPage, meta: { hideChrome: true } },
    { path: "/admin/messages", name: "admin-messages", component: AdminMessagesPage, meta: { hideChrome: true } },
    { path: "/admin/media", name: "admin-media", component: AdminMediaPage, meta: { hideChrome: true } },
    { path: "/admin/navigation", name: "admin-navigation", component: AdminNavigationPage, meta: { hideChrome: true } },
    { path: "/admin/stats", name: "admin-stats", component: AdminStatsPage, meta: { hideChrome: true } },
    { path: "/admin/settings", name: "admin-settings", component: AdminSettingsPage, meta: { hideChrome: true } }
  ],
  scrollBehavior() {
    return { top: 0 };
  }
});
