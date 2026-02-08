import { http } from "./http";

export interface Asset {
  id: number;
  name: string;
  description: string;
  category: string;
  tags: string;
  type: string;
  file_size: number;
  file_path: string;
  file_hash: string;
  format: string;
  thumbnail_path: string;
  use_count: number;
  last_used_at: string | null;
  uploaded_by: string;
  upload_ip: string;
  created_at: string;
  updated_at: string;
  file_url: string;
  thumbnail_url: string;
}

export interface AssetListParams {
  page?: number;
  pageSize?: number;
  type?: string;
  category?: string;
  format?: string;
  keyword?: string;
  tags?: string;
  sortBy?: string;
  sortOrder?: string;
}

export interface AssetListResponse {
  list: Asset[];
  total: number;
  page: number;
  pageSize: number;
}

// 获取资产列表
export const getAssets = (params: AssetListParams) => {
  return http.get<AssetListResponse>("/assets", { params });
};

// 获取资产详情
export const getAssetDetail = (id: number) => {
  return http.get<Asset>(`/assets/${id}`);
};

// 记录使用次数
export const recordAssetUse = (id: number) => {
  return http.post(`/assets/${id}/use`);
};

// 获取统计信息
export const getAssetStatistics = (type?: string) => {
  return http.get("/assets/statistics", { params: { type } });
};

// 按类型获取统计信息
export const getAssetStatisticsByType = () => {
  return http.get("/assets/statistics/by-type");
};

// 获取热门资产
export const getPopularAssets = (limit: number = 10, type?: string) => {
  return http.get<Asset[]>("/assets/popular", { params: { limit, type } });
};

// 删除资产
export const deleteAsset = (id: number) => {
  return http.delete(`/assets/${id}`);
};

// 更新资产信息
export const updateAsset = (id: number, data: Partial<Asset>) => {
  return http.put(`/assets/${id}`, data);
};
