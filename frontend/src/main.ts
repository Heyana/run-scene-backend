import { createApp } from "vue";
import App from "./App";
import router from "./router";
import pinia from "./stores";
import Antd from "ant-design-vue";
import "ant-design-vue/dist/reset.css";
import "./style.css";

const app = createApp(App);

app.use(router);
app.use(pinia);
app.use(Antd);

app.mount("#app");
