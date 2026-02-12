import { http } from "./http";

export interface AuditLog {
  id: number;
  user_id?: number;
  username: string;
  user_ip: string;
  action: string;
  resource: string;
  resource_id?: number;
  method: string;
  path: string;
  status_code: number;
  duration: number;
  request_body?: string;
  response_body?: string;
  error_msg?: string;
  user_agent?: string;
  created_at: string;
}

export interface AuditFilter {
  user_id?: number;
  username?: string;
  user_ip?: string;
  action?: string;
  resource?: string;
  resource_id?: number;
  start_time?: string;
  end_time?: string;
  status_code?: number;
  page?: number;
  page_size?: number;
}

export interface AuditStatistics {
  total_count: number;
  action_count: Record<string, number>;
  resource_count: Record<string, number>;
  top_users: Array<{ username: string; count: number }>;
  top_ips: Array<{ ip: string; count: number }>;
}

// 查询审计日志列表
export const getAuditLogs = (filter: AuditFilter) => {
  return http.get<{
    logs: AuditLog[];
    total: number;
    page: number;
    page_size: number;
  }>("/audit/logs", { params: filter });
};

// 获取单条审计日志
export const getAuditLog = (id: number) => {
  return http.get<AuditLog>(`/audit/logs/${id}`);
};

// 获取用户的审计日志
export const getUserLogs = (userId: number, limit = 100) => {
  return http.get<AuditLog[]>(`/audit/users/${userId}/logs`, {
    params: { limit },
  });
};

// 获取资源的审计日志
export const getResourceLogs = (
  resource: string,
  resourceId: number,
  limit = 100,
) => {
  return http.get<AuditLog[]>(
    `/audit/resources/${resource}/${resourceId}/logs`,
    { params: { limit } },
  );
};

// 获取统计信息
export const getAuditStatistics = (startTime?: string, endTime?: string) => {
  return http.get<AuditStatistics>("/audit/statistics", {
    params: { start_time: startTime, end_time: endTime },
  });
};

// 手动触发归档
export const triggerArchive = () => {
  return http.post<{ archived_count: number }>("/audit/archive");
};

// 获取归档统计信息
export const getArchiveStatistics = () => {
  return http.get<{
    database_count: number;
    oldest_log: string;
    newest_log: string;
    archive_files: number;
    retention_days: number;
    archive_enabled: boolean;
  }>("/audit/archive/statistics");
};

// 列出归档文件
export const listArchiveFiles = (startDate?: string, endDate?: string) => {
  return http.get<{ files: string[]; count: number }>("/audit/archive/files", {
    params: { start_date: startDate, end_date: endDate },
  });
};
