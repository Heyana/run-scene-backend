import {
  createRouter,
  createWebHashHistory,
  createWebHistory,
  type RouteRecordRaw,
} from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "Home",
    component: () => import("@/views/Home"),
  },
  {
    path: "/textures",
    name: "Textures",
    component: () => import("@/views/Textures"),
  },
  {
    path: "/texture-analysis",
    name: "TextureTypeAnalysis",
    component: () => import("@/views/TextureTypeAnalysis"),
  },
  {
    path: "/projects",
    name: "Projects",
    component: () => import("@/views/Projects"),
  },
  {
    path: "/about",
    name: "About",
    component: () => import("@/views/About"),
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
