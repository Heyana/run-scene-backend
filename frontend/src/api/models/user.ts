import { http } from "../http";

// ==================== 类型定义 ====================

export interface User {
  id: number;
  username: string;
  email: string;
  real_name?: string;
  phone?: string;
  avatar?: string;
  status: "active" | "disabled" | "locked";
  last_login_at?: string;
  created_at: string;
}

export interface UserListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  status?: string;
  role?: string;
}

export interface UserListResponse {
  items: User[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateUserRequest {
  username: string;
  password: string;
  email: string;
  real_name?: string;
  phone?: string;
  role_ids?: number[];
}

export interface UpdateUserRequest {
  email?: string;
  real_name?: string;
  phone?: string;
  avatar?: string;
}

export interface UserPermissionsResponse {
  permissions: string[];
  roles?: any[];
  permission_groups?: any[];
}

// ==================== API 方法 ====================

/**
 * 获取用户列表
 */
export const getUserList = (params?: UserListParams) => {
  return http.get<UserListResponse>("users", { params });
};

/**
 * 创建用户
 */
export const createUser = (data: CreateUserRequest) => {
  return http.post<User>("users", data);
};

/**
 * 获取用户详情
 */
export const getUserDetail = (id: number) => {
  return http.get<User>(`users/${id}`);
};

/**
 * 更新用户
 */
export const updateUser = (id: number, data: UpdateUserRequest) => {
  return http.put<User>(`users/${id}`, data);
};

/**
 * 删除用户
 */
export const deleteUser = (id: number) => {
  return http.delete(`users/${id}`);
};

/**
 * 禁用用户
 */
export const disableUser = (id: number) => {
  return http.post(`users/${id}/disable`);
};

/**
 * 启用用户
 */
export const enableUser = (id: number) => {
  return http.post(`users/${id}/enable`);
};

/**
 * 重置密码
 */
export const resetUserPassword = (id: number, newPassword: string) => {
  return http.post(`users/${id}/reset-password`, {
    new_password: newPassword,
  });
};

/**
 * 分配角色
 */
export const assignRoles = (id: number, roleIds: number[]) => {
  return http.post(`users/${id}/roles`, {
    role_ids: roleIds,
  });
};

/**
 * 获取用户权限
 */
export const getUserPermissions = (id: number) => {
  return http.get<UserPermissionsResponse>(`users/${id}/permissions`);
};
