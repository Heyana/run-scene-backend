import { http } from "../http";

// ==================== 类型定义 ====================

export interface Role {
  id: number;
  code: string;
  name: string;
  description?: string;
  is_system: boolean;
  created_at: string;
}

export interface RoleListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
}

export interface RoleListResponse {
  items: Role[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateRoleRequest {
  code: string;
  name: string;
  description?: string;
  permission_ids?: number[];
  permission_group_ids?: number[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}

export interface AssignPermissionsRequest {
  permission_ids?: number[];
  permission_group_ids?: number[];
}

export interface RolePermissionsResponse {
  permission_ids: number[];
  permission_group_ids: number[];
}

// ==================== API 方法 ====================

/**
 * 获取角色列表
 */
export const getRoleList = (params?: RoleListParams) => {
  return http.get<RoleListResponse>("roles", { params });
};

/**
 * 创建角色
 */
export const createRole = (data: CreateRoleRequest) => {
  return http.post<Role>("roles", data);
};

/**
 * 获取角色详情
 */
export const getRoleDetail = (id: number) => {
  return http.get<Role>(`roles/${id}`);
};

/**
 * 更新角色
 */
export const updateRole = (id: number, data: UpdateRoleRequest) => {
  return http.put<Role>(`roles/${id}`, data);
};

/**
 * 删除角色
 */
export const deleteRole = (id: number) => {
  return http.delete(`roles/${id}`);
};

/**
 * 分配权限给角色
 */
export const assignRolePermissions = (
  id: number,
  data: AssignPermissionsRequest,
) => {
  return http.post(`roles/${id}/permissions`, data);
};

/**
 * 分配权限组给角色
 */
export const assignRolePermissionGroups = (id: number, groupIds: number[]) => {
  return http.post(`roles/${id}/permission-groups`, {
    permission_group_ids: groupIds,
  });
};

/**
 * 获取角色的权限
 */
export const getRolePermissions = (id: number) => {
  return http.get<RolePermissionsResponse>(`roles/${id}/permissions`);
};
