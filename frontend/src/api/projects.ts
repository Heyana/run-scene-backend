import { http } from "./http";

export interface Project {
  id: number;
  name: string;
  description: string;
  current_version: string;
  latest_version_id: number;
  created_at: string;
  updated_at: string;
}

export interface ProjectVersion {
  id: number;
  project_id: number;
  version: string;
  username: string;
  description: string;
  file_path: string;
  file_size: number;
  file_hash: string;
  file_count: number;
  upload_ip: string;
  extracted_path: string;
  created_at: string;
  file_url: string;
  preview_url: string;
}

export interface ProjectListResponse {
  total: number;
  page: number;
  page_size: number;
  data: Project[];
}

export interface CreateProjectRequest {
  name: string;
  description: string;
}

export interface UploadVersionRequest {
  project_id: number;
  username: string;
  description: string;
  version_type: "major" | "minor" | "patch";
  files: File[];
}

// 获取项目列表
export const getProjects = (params: {
  page?: number;
  page_size?: number;
  keyword?: string;
}) => {
  return http.get<ProjectListResponse>("/projects", { params });
};

// 创建项目
export const createProject = (data: CreateProjectRequest) => {
  return http.post<Project>("/projects", data);
};

// 获取项目详情
export const getProject = (id: number) => {
  return http.get<Project>(`/projects/${id}`);
};

// 删除项目
export const deleteProject = (id: number) => {
  return http.delete(`/projects/${id}`);
};

// 上传版本
export const uploadVersion = (projectId: number, formData: FormData) => {
  return http.post<ProjectVersion>(
    `/projects/${projectId}/versions`,
    formData,
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );
};

// 获取版本历史
export const getVersionHistory = (projectId: number) => {
  return http.get<ProjectVersion[]>(`/projects/${projectId}/versions`);
};

// 下载版本
export const downloadVersion = (versionId: number) => {
  return `/projects/versions/${versionId}/download`;
};

// 回滚版本
export const rollbackVersion = (versionId: number) => {
  return http.post(`/projects/versions/${versionId}/rollback`);
};
