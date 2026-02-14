import { ref, onMounted, onUnmounted } from "vue";
import { onAuthRequired, isAuthenticated } from "@/utils/auth";

/**
 * 认证 Hook
 * 自动检查登录状态并监听认证事件
 */
export function useAuth() {
  const authModalVisible = ref(false);
  let unsubscribe: (() => void) | null = null;

  // 检查是否已登录
  const checkAuth = () => {
    if (!isAuthenticated()) {
      authModalVisible.value = true;
      return false;
    }
    return true;
  };

  // 登录成功回调
  const handleAuthSuccess = () => {
    authModalVisible.value = false;
  };

  // 关闭弹窗回调
  const handleAuthCancel = () => {
    // 如果用户取消登录，可以选择跳转到其他页面
    // 这里暂时只关闭弹窗
  };

  onMounted(() => {
    checkAuth();

    // 监听全局认证事件
    unsubscribe = onAuthRequired(() => {
      authModalVisible.value = true;
    });
  });

  onUnmounted(() => {
    // 取消监听
    if (unsubscribe) {
      unsubscribe();
    }
  });

  return {
    authModalVisible,
    checkAuth,
    handleAuthSuccess,
    handleAuthCancel,
  };
}
