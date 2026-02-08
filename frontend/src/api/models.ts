import { http } from "./http";

export interface Model {
  id: number;
  name: string;
  description: string;
  category: string;
  tags: string;
  type: string;
  file_size: number;
  file_path: string;
  file_hash: string;
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

export interface ModelListParams {
  page?: number;
  pageSize?: number;
  category?: string;
  type?: string;
  tags?: string;
  sortBy?: string;
  sortOrder?: string;
}

export interface ModelListResponse {
  list: Model[];
  total: number;
  page: number;
  pageSize: number;
}

// 获取模型列表
export const getModels = (params: ModelListParams) => {
  return http.get<ModelListResponse>("/models", { params });
};

// 获取模型详情
export const getModelDetail = (id: number) => {
  return http.get<Model>(`/models/${id}`);
};

// 记录使用次数
export const recordModelUse = (id: number) => {
  return http.post(`/models/${id}/use`);
};

// 获取统计信息
export const getModelStatistics = () => {
  return http.get("/models/statistics");
};

// 获取热门模型
export const getPopularModels = (limit: number = 10) => {
  return http.get<Model[]>("/models/popular", { params: { limit } });
};

// 搜索模型
export const searchModels = (
  keyword: string,
  page: number = 1,
  pageSize: number = 20,
) => {
  return http.get<ModelListResponse>("/models/search", {
    params: { keyword, page, pageSize },
  });
};

// 删除模型
export const deleteModel = (id: number) => {
  return http.delete(`/models/${id}`);
};
