import { createRouter, createWebHistory } from "vue-router";

import AccountPage from "../pages/AccountPage.vue";
import AdminHome from "../pages/admin/AdminHome.vue";
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
    { path: "/admin", name: "admin", component: AdminHome }
  ],
  scrollBehavior() {
    return { top: 0 };
  }
});
