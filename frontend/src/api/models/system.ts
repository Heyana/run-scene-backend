import { http } from "../http";

// ==================== 类型定义 ====================

export interface PingResponse {
  timestamp: number;
  status: string;
}

export interface HealthResponse {
  status: string;
  timestamp: number;
  service: string;
}

// ==================== API 方法 ====================

/**
 * 健康检查
 */
export const ping = () => {
  return http.get<PingResponse>("ping");
};

/**
 * 服务健康检查
 */
export const health = () => {
  return http.get<HealthResponse>("/health");
};
