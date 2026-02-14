// 认证工具类

import { constant } from "@/api/const";

// 事件监听器
type AuthEventListener = () => void;
const authEventListeners: AuthEventListener[] = [];

/**
 * 触发需要认证事件
 */
export function triggerAuthRequired() {
  authEventListeners.forEach((listener) => listener());
}

/**
 * 监听需要认证事件
 */
export function onAuthRequired(listener: AuthEventListener) {
  authEventListeners.push(listener);

  // 返回取消监听的函数
  return () => {
    const index = authEventListeners.indexOf(listener);
    if (index > -1) {
      authEventListeners.splice(index, 1);
    }
  };
}

/**
 * 检查是否已登录
 */
export function isAuthenticated(): boolean {
  return !!localStorage.getItem(constant.runSceneBackendToken);
}

/**
 * 获取当前用户信息
 */
export function getCurrentUser() {
  const userStr = localStorage.getItem("user");
  if (userStr) {
    try {
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  }
  return null;
}

/**
 * 登出
 */
export function logout() {
  localStorage.removeItem(constant.runSceneBackendToken);
  localStorage.removeItem("user");
}
