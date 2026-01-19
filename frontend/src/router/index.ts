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
