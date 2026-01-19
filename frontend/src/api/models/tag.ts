import { http } from "../http";

// ==================== 类型定义 ====================

export interface Tag {
  id: number;
  name: string;
  type: string;
  use_count: number;
  created_at: string;
}

export interface TagListResponse {
  items: Tag[];
  total: number;
}

// ==================== API 方法 ====================

/**
 * 获取标签列表
 */
export const getTagList = (params?: { type?: string }) => {
  return http.get<TagListResponse>("tags", { params });
};

/**
 * 根据标签获取材质列表
 */
export const getTexturesByTag = (
  tagId: number,
  params?: { page?: number; pageSize?: number },
) => {
  return http.get(`tags/${tagId}/textures`, { params });
};
