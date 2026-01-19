import { http } from "../http";

// ==================== 类型定义 ====================

export interface SecurityStatus {
  blocked_ips_count: number;
  whitelist_count: number;
  active_connections: number;
  requests_per_minute: number;
  ddos_protection_enabled: boolean;
}

export interface BlockedIP {
  ip: string;
  reason: string;
  blocked_at: string;
  expires_at?: string;
}

export interface IPStats {
  ip: string;
  request_count: number;
  last_request_at: string;
  is_blocked: boolean;
  is_whitelisted: boolean;
}

export interface Connection {
  ip: string;
  connection_count: number;
  last_activity: string;
}

// ==================== API 方法 ====================

/**
 * 获取安全状态
 */
export const getSecurityStatus = () => {
  return http.get<SecurityStatus>("security/status");
};

/**
 * 获取被封禁 IP 列表
 */
export const getBlockedIPs = () => {
  return http.get<{ items: BlockedIP[] }>("security/blocked-ips");
};

/**
 * 解封 IP 地址
 */
export const unblockIP = (ip: string) => {
  return http.post(`security/unblock/${ip}`);
};

/**
 * 获取 IP 统计信息
 */
export const getIPStats = (params?: { page?: number; pageSize?: number }) => {
  return http.get<{ items: IPStats[]; total: number }>("security/ip-stats", {
    params,
  });
};

/**
 * 封禁 IP 地址
 */
export const blockIP = (
  ip: string,
  data?: { reason?: string; duration?: number },
) => {
  return http.post(`security/block/${ip}`, data);
};

/**
 * 添加 IP 到白名单
 */
export const addToWhitelist = (ip: string) => {
  return http.post(`security/whitelist/${ip}`);
};

/**
 * 从白名单移除 IP
 */
export const removeFromWhitelist = (ip: string) => {
  return http.delete(`security/whitelist/${ip}`);
};

/**
 * 获取连接统计
 */
export const getConnections = () => {
  return http.get<{ items: Connection[] }>("security/connections");
};
