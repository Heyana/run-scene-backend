import { http } from "../http";

// ==================== 类型定义 ====================

export interface BackupStatus {
  last_backup_time?: string;
  next_backup_time?: string;
  backup_enabled: boolean;
  database_backup_enabled: boolean;
  cdn_backup_enabled: boolean;
  backup_count: number;
}

export interface BackupHistory {
  id: number;
  backup_type: string;
  status: string;
  file_path: string;
  file_size: number;
  created_at: string;
  completed_at?: string;
  error_msg?: string;
}

// ==================== API 方法 ====================

/**
 * 获取备份状态
 */
export const getBackupStatus = () => {
  return http.get<BackupStatus>("backup/status");
};

/**
 * 触发手动全量备份
 */
export const triggerManualBackup = () => {
  return http.post("backup/trigger");
};

/**
 * 触发数据库备份
 */
export const triggerDatabaseBackup = () => {
  return http.post("backup/database");
};

/**
 * 触发 CDN 备份
 */
export const triggerCDNBackup = () => {
  return http.post("backup/cdn");
};

/**
 * 获取备份历史
 */
export const getBackupHistory = (params?: {
  page?: number;
  pageSize?: number;
}) => {
  return http.get<{ items: BackupHistory[]; total: number }>("backup/history", {
    params,
  });
};

/**
 * 从备份恢复 CDN 文件
 */
export const restoreCDNFromBackup = (backupId: number) => {
  return http.post(`backup/restore/cdn/${backupId}`);
};
