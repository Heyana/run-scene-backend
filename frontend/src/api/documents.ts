import { http } from "./http";

export interface Document {
  id: number;
  name: string;
  description: string;
  category: string;
  tags: string;
  type: string;
  parent_id: number | null;
  is_folder: boolean;
  child_count: number;
  file_size: number;
  file_path: string;
  file_hash: string;
  format: string;
  thumbnail_path: string;
  preview_path: string;
  version: string;
  parent_version_id: number | null;
  is_latest: boolean;
  department: string;
  project: string;
  is_public: boolean;
  download_count: number;
  view_count: number;
  last_viewed_at: string | null;
  uploaded_by: string;
  upload_ip: string;
  created_at: string;
  updated_at: string;
  file_url: string;
  thumbnail_url: string;
  preview_url: string;
  folder_thumbnails?: string[]; // 文件夹缩略图（前4个文件）
}

export interface DocumentListParams {
  page?: number;
  pageSize?: number;
  type?: string;
  category?: string;
  format?: string;
  department?: string;
  project?: string;
  keyword?: string;
  tags?: string;
  sortBy?: string;
  sortOrder?: string;
  parent_id?: number; // 父文件夹ID
}

export interface DocumentListResponse {
  list: Document[];
  total: number;
  page: number;
  pageSize: number;
}

// 获取文档列表
export const getDocuments = (params: DocumentListParams) => {
  return http.get<DocumentListResponse>("/documents", { params });
};

// 获取文档详情
export const getDocument = (id: number) => {
  return http.get<Document>(`/documents/${id}`);
};

// 上传文档
export const uploadDocument = (formData: FormData, config?: any) => {
  return http.post("/documents/upload", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
    ...config,
  });
};

// 上传文件夹（保持结构）
export const uploadFolder = (formData: FormData, config?: any) => {
  return http.post("/documents/upload-folder", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
    ...config,
  });
};

// 删除文档
export const deleteDocument = (id: number) => {
  return http.delete(`/documents/${id}`);
};

// 更新文档
export const updateDocument = (id: number, data: Partial<Document>) => {
  return http.put(`/documents/${id}`, data);
};

// 获取统计信息
export const getDocumentStatistics = (params?: any) => {
  return http.get("/documents/statistics", { params });
};

// 获取热门文档
export const getPopularDocuments = (limit: number = 10, type?: string) => {
  return http.get<Document[]>("/documents/popular", {
    params: { limit, type },
  });
};

// 获取版本列表
export const getDocumentVersions = (id: number) => {
  return http.get<Document[]>(`/documents/${id}/versions`);
};

// 获取访问日志
export const getDocumentLogs = (id: number, params?: any) => {
  return http.get(`/documents/${id}/logs`, { params });
};

// 创建文件夹
export const createFolder = (data: {
  name: string;
  description?: string;
  parent_id?: number;
  department?: string;
  project?: string;
}) => {
  return http.post("/documents/folder", data);
};
