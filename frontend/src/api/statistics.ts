import { http } from "./http";

export interface ResourceStats {
  total: number;
  trend: number;
  recent_count: number;
}

export interface OverviewResponse {
  textures: ResourceStats;
  projects: ResourceStats;
  models: ResourceStats;
  assets: ResourceStats;
  ai3d: ResourceStats;
}

export interface Activity {
  id: number;
  type: string;
  name: string;
  action: string;
  user: string;
  version?: string;
  created_at: string;
}

export interface RecentActivitiesResponse {
  activities: Activity[];
}

export interface SystemStatus {
  service: {
    status: string;
    uptime: number;
  };
  database: {
    status: string;
    size: number;
  };
  storage: {
    total: number;
    used: number;
    usage_percent: number;
  };
  sync: {
    last_sync_at: string;
    status: string;
  };
}

// 获取统计概览
export const getOverview = () => {
  return http.get<OverviewResponse>("/statistics/overview");
};

// 获取最近活动
export const getRecentActivities = (limit?: number) => {
  return http.get<RecentActivitiesResponse>("/statistics/recent-activities", {
    params: { limit },
  });
};

// 获取系统状态
export const getSystemStatus = () => {
  return http.get<SystemStatus>("/statistics/system-status");
};
