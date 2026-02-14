import { http } from "../http";

// ==================== 类型定义 ====================

export interface Permission {
  id: number;
  code: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
  is_system: boolean;
  created_at: string;
}

export interface PermissionGroup {
  id: number;
  code: string;
  name: string;
  description?: string;
  is_system: boolean;
  created_at: string;
}

export interface PermissionListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  resource?: string;
  action?: string;
  is_system?: boolean;
}

export interface PermissionListResponse {
  items: Permission[];
  total: number;
  page: number;
  page_size: number;
}

export interface PermissionGroupListResponse {
  items: PermissionGroup[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreatePermissionRequest {
  code: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
}

export interface UpdatePermissionRequest {
  name?: string;
  description?: string;
}

export interface CreatePermissionGroupRequest {
  code: string;
  name: string;
  description?: string;
  permission_ids?: number[];
}

export interface UpdatePermissionGroupRequest {
  name?: string;
  description?: string;
}

// ==================== 权限管理 API ====================

/**
 * 获取权限列表
 */
export const getPermissionList = (params?: PermissionListParams) => {
  return http.get<PermissionListResponse>("permissions", { params });
};

/**
 * 创建权限
 */
export const createPermission = (data: CreatePermissionRequest) => {
  return http.post<Permission>("permissions", data);
};

/**
 * 获取权限详情
 */
export const getPermissionDetail = (id: number) => {
  return http.get<Permission>(`permissions/${id}`);
};

/**
 * 更新权限
 */
export const updatePermission = (id: number, data: UpdatePermissionRequest) => {
  return http.put<Permission>(`permissions/${id}`, data);
};

/**
 * 删除权限
 */
export const deletePermission = (id: number) => {
  return http.delete(`permissions/${id}`);
};

/**
 * 获取所有资源类型
 */
export const getResources = () => {
  return http.get<string[]>("permissions/resources");
};

/**
 * 获取所有操作类型
 */
export const getActions = () => {
  return http.get<string[]>("permissions/actions");
};

// ==================== 权限组管理 API ====================

/**
 * 获取权限组列表
 */
export const getPermissionGroupList = (params?: PermissionListParams) => {
  return http.get<PermissionGroupListResponse>("permission-groups", { params });
};

/**
 * 创建权限组
 */
export const createPermissionGroup = (data: CreatePermissionGroupRequest) => {
  return http.post<PermissionGroup>("permission-groups", data);
};

/**
 * 获取权限组详情
 */
export const getPermissionGroupDetail = (id: number) => {
  return http.get<PermissionGroup>(`permission-groups/${id}`);
};

/**
 * 更新权限组
 */
export const updatePermissionGroup = (
  id: number,
  data: UpdatePermissionGroupRequest,
) => {
  return http.put<PermissionGroup>(`permission-groups/${id}`, data);
};

/**
 * 删除权限组
 */
export const deletePermissionGroup = (id: number) => {
  return http.delete(`permission-groups/${id}`);
};

/**
 * 添加权限到权限组
 */
export const addPermissionsToGroup = (id: number, permissionIds: number[]) => {
  return http.post(`permission-groups/${id}/permissions`, {
    permission_ids: permissionIds,
  });
};

/**
 * 从权限组移除权限
 */
export const removePermissionFromGroup = (
  groupId: number,
  permissionId: number,
) => {
  return http.delete(
    `permission-groups/${groupId}/permissions/${permissionId}`,
  );
};
