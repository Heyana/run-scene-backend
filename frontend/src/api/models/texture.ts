import { http } from "../http";

// ==================== 类型定义 ====================

export interface TextureFile {
  id: number;
  file_type: string;
  related_id: number;
  related_type: string;
  original_url: string;
  local_path: string;
  cdn_path: string;
  file_name: string;
  file_size: number;
  width: number;
  height: number;
  format: string;
  md5: string;
  version: number;
  status: number;
  download_retry: number;
  last_error: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  full_url: string; // 后端自动拼接的完整 URL
}

export interface Texture {
  id: number;
  asset_id: string;
  name: string;
  description: string;
  type: number;
  authors: string;
  max_resolution: string;
  files_hash: string;
  date_published: number;
  download_count: number;
  use_count: number;
  last_used_at?: string;
  priority: number;
  sync_status: number;
  created_at: string;
  updated_at: string;
  files?: TextureFile[]; // 添加文件列表
}

export interface TextureSyncLog {
  id: number;
  sync_type: string;
  status: number;
  total_count: number;
  processed_count: number;
  success_count: number;
  fail_count: number;
  skip_count: number;
  current_asset: string;
  progress: number;
  download_speed: number;
  start_time: string;
  end_time: string;
  error_msg: string;
  created_at: string;
  updated_at: string;
}

export interface TextureListParams {
  page?: number;
  pageSize?: number; // 改为 pageSize 匹配后端
  keyword?: string;
  tag_id?: number;
  sortBy?: "use_count" | "date_published" | "created_at"; // 改为 sortBy
  order?: "asc" | "desc";
  syncStatus?: number; // 同步状态: 0=未同步 1=同步中 2=已同步 3=失败
}

export interface TextureListResponse {
  list: Texture[]; // 改为 list 匹配后端返回
  total: number;
  page: number;
  pageSize: number; // 改为 pageSize
}

export interface SyncProgress {
  log_id: number;
  status: number;
  progress: number;
  current_asset: string;
  processed_count: number;
  total_count: number;
  download_speed: number;
}

// ==================== API 方法 ====================

/**
 * 获取材质列表
 */
export const getTextureList = (params?: TextureListParams) => {
  return http.get<TextureListResponse>("textures", { params });
};

/**
 * 获取材质详情
 */
export const getTextureDetail = (assetId: string) => {
  return http.get<Texture>(`textures/${assetId}`);
};

/**
 * 记录材质使用
 */
export const recordTextureUse = (assetId: string) => {
  return http.post(`textures/${assetId}/use`);
};

/**
 * 触发同步
 */
export const triggerSync = (data?: { sync_type?: string }) => {
  return http.post<{ log_id: number }>("textures/sync", data);
};

/**
 * 获取同步进度
 */
export const getSyncProgress = () => {
  return http.get<SyncProgress>("textures/sync/progress");
};

/**
 * 获取同步状态
 */
export const getSyncStatus = (logId: number) => {
  return http.get<TextureSyncLog>(`textures/sync/status/${logId}`);
};

/**
 * 获取同步日志列表
 */
export const getSyncLogs = (params?: { page?: number; pageSize?: number }) => {
  return http.get<{ items: TextureSyncLog[]; total: number }>(
    "textures/sync/logs",
    { params },
  );
};
