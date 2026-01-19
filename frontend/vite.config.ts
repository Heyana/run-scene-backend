import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vueJsx from "@vitejs/plugin-vue-jsx";
import { resolve } from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueJsx()],
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },
  server: {
    host: "0.0.0.0", // 监听所有网络接口，允许局域网访问
    port: 3000,
    proxy: {
      "/api": {
        target: "http://192.168.3.39:23359", // 修改为后端的局域网地址
        changeOrigin: true,
      },
      "/textures": {
        target: "http://192.168.3.39:23359", // 代理材质文件请求
        changeOrigin: true,
      },
    },
  },
});
