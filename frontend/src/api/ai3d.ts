import { http } from "./http";

export interface AI3DTask {
  id: number;
  provider: string;
  taskId: string;
  prompt: string;
  status: string;
  progress: number;
  modelUrl: string;
  thumbnailUrl: string;
  errorMessage: string;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
}

export interface AI3DListParams {
  page?: number;
  pageSize?: number;
  provider?: string;
  status?: string;
}

export interface AI3DListResponse {
  list: AI3DTask[];
  total: number;
  page: number;
  pageSize: number;
}

// 获取任务列表
export const getAI3DTasks = (params: AI3DListParams) => {
  return http.get<AI3DListResponse>("/ai3d/tasks", { params });
};

// 获取任务详情
export const getAI3DTask = (id: number) => {
  return http.get<AI3DTask>(`/ai3d/tasks/${id}`);
};

// 提交任务
export const submitAI3DTask = (data: {
  provider: string;
  prompt: string;
  image?: File;
}) => {
  const formData = new FormData();
  formData.append("provider", data.provider);
  formData.append("prompt", data.prompt);
  if (data.image) {
    formData.append("image", data.image);
  }
  return http.post<AI3DTask>("/ai3d/tasks", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
};

// 轮询任务状态
export const pollAI3DTask = (id: number) => {
  return http.post<AI3DTask>(`/ai3d/tasks/${id}/poll`);
};

// 删除任务
export const deleteAI3DTask = (id: number) => {
  return http.delete(`/ai3d/tasks/${id}`);
};

// 获取配置
export const getAI3DConfig = () => {
  return http.get("/ai3d/config");
};
