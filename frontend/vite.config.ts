import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vueJsx from "@vitejs/plugin-vue-jsx";
import { resolve } from "path";
import { visualizer } from "rollup-plugin-visualizer";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    visualizer({
      open: true,
      gzipSize: true,
      brotliSize: true,
      filename: "docs/性能/bundle-analysis.html",
    }),
  ],
  resolve: {
    alias: {
      "@": resolve(__dirname, "./src"),
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
  base: "./",
  css: {
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
        additionalData: `@import "@/styles/vars/index.less";
@import "@/styles/vars/class.less";
@import "@/styles/vars/other.less";
@import "@/styles/vars.less";
@import "@/styles/scoped-var.less";`,
      },
    },
  },
});
