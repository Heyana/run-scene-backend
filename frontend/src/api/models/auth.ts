import { http } from "../http";

// ==================== 类型定义 ====================

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
  phone?: string;
  real_name?: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
  token_type: string;
  user?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
  };
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface CheckPermissionRequest {
  resource: string;
  action: string;
}

export interface CheckPermissionResponse {
  has_permission: boolean;
  permission?: string;
  reason?: string;
}

export interface UserProfile {
  user: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    phone?: string;
    avatar?: string;
    status: string;
    last_login_at?: string;
    created_at: string;
  };
  roles?: any[];
  permissions?: string[];
}

// ==================== API 方法 ====================

/**
 * 用户注册
 */
export const register = (data: RegisterRequest) => {
  return http.post<TokenResponse>("auth/register", data);
};

/**
 * 用户登录
 */
export const login = (data: LoginRequest) => {
  return http.post<TokenResponse>("auth/login", data);
};

/**
 * 用户登出
 */
export const logout = () => {
  return http.post("auth/logout");
};

/**
 * 刷新 Token
 */
export const refreshToken = () => {
  return http.post<{
    access_token: string;
    expires_in: number;
    token_type: string;
  }>("auth/refresh");
};

/**
 * 修改密码
 */
export const changePassword = (data: ChangePasswordRequest) => {
  return http.post("auth/change-password", data);
};

/**
 * 获取当前用户信息
 */
export const getProfile = () => {
  return http.get<UserProfile>("auth/profile");
};

/**
 * 检查权限
 */
export const checkPermission = (data: CheckPermissionRequest) => {
  return http.post<CheckPermissionResponse>("auth/check-permission", data);
};
