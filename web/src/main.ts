import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./app/App.vue";
import LoadingState from "./components/LoadingState.vue";
import { router } from "./router";
import "./styles/main.css";

createApp(App).component("LoadingState", LoadingState).use(createPinia()).use(router).mount("#app");
