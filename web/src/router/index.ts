import { createRouter, createWebHistory } from "vue-router";

import AdminHome from "../pages/admin/AdminHome.vue";
import ArchivePage from "../pages/ArchivePage.vue";
import HomePage from "../pages/HomePage.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: HomePage },
    { path: "/archive", name: "archive", component: ArchivePage },
    { path: "/admin", name: "admin", component: AdminHome }
  ],
  scrollBehavior() {
    return { top: 0 };
  }
});
