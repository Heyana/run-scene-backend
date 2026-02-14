import {
  createRouter,
  createWebHashHistory,
  type RouteRecordRaw,
} from "vue-router";
import { isAuthenticated } from "@/utils/auth";

// 扩展路由元信息类型
declare module "vue-router" {
  interface RouteMeta {
    requiresAuth?: boolean;
    title?: string;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/Login"),
    meta: { title: "登录" },
  },
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
    path: "/models",
    name: "Models",
    component: () => import("@/views/Models.tsx"),
  },
  {
    path: "/assets",
    name: "Assets",
    component: () => import("@/views/Assets.tsx"),
  },
  {
    path: "/documents",
    name: "Documents",
    component: () => import("@/views/Documents.tsx"),
  },
  {
    path: "/ai3d",
    name: "AI3D",
    component: () => import("@/views/AI3D.tsx"),
  },
  {
    path: "/audit-logs",
    name: "AuditLogs",
    component: () => import("@/views/AuditLogs.tsx"),
  },
  {
    path: "/user-management",
    name: "UserManagement",
    component: () => import("@/views/UserManagement"),
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        name: "UserManagementWelcome",
        component: () => import("@/views/UserManagement/Welcome"),
        meta: { requiresAuth: true, title: "人员管理" },
      },
      {
        path: "users",
        name: "UserList",
        component: () => import("@/views/UserManagement/UserList"),
        meta: { requiresAuth: true, title: "用户管理" },
      },
      {
        path: "roles",
        name: "RoleList",
        component: () => import("@/views/UserManagement/RoleList"),
        meta: { requiresAuth: true, title: "角色管理" },
      },
      {
        path: "permissions",
        name: "PermissionList",
        component: () => import("@/views/UserManagement/PermissionList"),
        meta: { requiresAuth: true, title: "权限管理" },
      },
      {
        path: "permission-groups",
        name: "PermissionGroupList",
        component: () => import("@/views/UserManagement/PermissionGroupList"),
        meta: { requiresAuth: true, title: "权限组管理" },
      },
    ],
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

// 全局路由守卫
// 全局路由守卫
router.beforeEach((to, _from, next) => {
  console.log(
    "Router guard:",
    to.path,
    "requiresAuth:",
    to.meta.requiresAuth,
    "isAuthenticated:",
    isAuthenticated(),
  );

  // 如果要去登录页，直接放行
  if (to.path === "/login") {
    // 如果已经登录了，跳转到首页或 redirect 指定的页面
    if (isAuthenticated()) {
      const redirect = to.query.redirect as string;
      if (redirect && redirect !== "/login") {
        console.log("Already logged in, redirecting to:", redirect);
        next(redirect);
      } else {
        console.log("Already logged in, redirecting to home");
        next("/");
      }
      return;
    }
    next();
    return;
  }

  // 检查路由是否需要认证
  if (to.meta.requiresAuth) {
    if (!isAuthenticated()) {
      console.log("Not authenticated, redirecting to login");
      // 未登录，跳转到登录页，并记录目标路径
      next({
        path: "/login",
        query: { redirect: to.fullPath },
      });
      return;
    }
  }

  next();
});

export default router;
