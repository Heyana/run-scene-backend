import { http } from "../http";

// ==================== 类型定义 ====================

export interface Company {
  id: number;
  name: string;
  logo?: string;
  description?: string;
  owner_id: number;
  status: string;
  member_count?: number;
  project_count?: number;
  created_at: string;
  updated_at: string;
}

export interface CompanyMember {
  id: number;
  company_id: number;
  user_id: number;
  role: "company_admin" | "member" | "viewer";
  joined_at: string;
  user?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    avatar?: string;
  };
}

export interface Project {
  id: number;
  company_id: number;
  name: string;
  key: string;
  description?: string;
  owner_id: number;
  status: string;
  start_date?: string;
  end_date?: string;
  member_count?: number;
  mission_count?: number;
  created_at: string;
  updated_at: string;
  company?: Company;
}

export interface ProjectMember {
  id: number;
  project_id: number;
  user_id: number;
  role: "project_admin" | "developer" | "viewer";
  joined_at: string;
  user?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    avatar?: string;
  };
}

export interface MissionList {
  id: number;
  project_id: number;
  name: string;
  type: "sprint" | "version" | "module";
  description?: string;
  color: string; // 列颜色
  status: "planning" | "active" | "completed";
  start_date?: string;
  end_date?: string;
  sort_order: number;
  mission_count?: number;
  created_at: string;
  updated_at: string;
}

export interface Mission {
  id: number;
  mission_list_id: number;
  project_id: number;
  mission_key: string;
  title: string;
  description?: string;
  type: "feature" | "enhancement" | "bug";
  priority: "P0" | "P1" | "P2" | "P3";
  status: "todo" | "in_progress" | "done" | "closed";
  assignee_id?: number;
  reporter_id: number;
  estimated_hours?: number;
  actual_hours?: number;
  start_date?: string;
  due_date?: string;
  completed_at?: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
  assignee?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    avatar?: string;
  };
  reporter?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    avatar?: string;
  };
  tags?: MissionTag[];
  comments?: MissionComment[];
  attachments?: MissionAttachment[];
}

export interface MissionComment {
  id: number;
  mission_id: number;
  user_id: number;
  content: string;
  parent_id?: number;
  created_at: string;
  user?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    avatar?: string;
  };
}

export interface MissionAttachment {
  id: number;
  mission_id: number;
  user_id: number;
  file_name: string;
  file_path: string;
  file_size: number;
  file_type: string;
  created_at: string;
  user?: {
    id: number;
    username: string;
    email: string;
    real_name?: string;
    avatar?: string;
  };
}

export interface MissionTag {
  id: number;
  name: string;
  color: string;
  created_at: string;
}

export interface ProjectStatistics {
  total_missions: number;
  completed_missions: number;
  in_progress_missions: number;
  todo_missions: number;
  overdue_missions: number;
  completion_rate: number;
  by_type: Record<string, number>;
  by_priority: Record<string, number>;
  by_assignee: Array<{
    user_id: number;
    username: string;
    count: number;
  }>;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

// ==================== 公司管理 API ====================

export const getCompanyList = (params?: {
  page?: number;
  page_size?: number;
}) => {
  return http.get<ListResponse<Company>>("requirement/companies", { params });
};

export const createCompany = (data: {
  name: string;
  logo?: string;
  description?: string;
}) => {
  return http.post<Company>("requirement/companies", data);
};

export const getCompanyDetail = (id: number) => {
  return http.get<Company>(`requirement/companies/${id}`);
};

export const updateCompany = (id: number, data: Partial<Company>) => {
  return http.put<Company>(`requirement/companies/${id}`, data);
};

export const deleteCompany = (id: number) => {
  return http.delete(`requirement/companies/${id}`);
};

export const getCompanyMembers = (id: number) => {
  return http.get<CompanyMember[]>(`requirement/companies/${id}/members`);
};

export const addCompanyMember = (
  id: number,
  data: { user_id: number; role: string },
) => {
  return http.post<CompanyMember>(`requirement/companies/${id}/members`, data);
};

export const removeCompanyMember = (companyId: number, memberId: number) => {
  return http.delete(`requirement/companies/${companyId}/members/${memberId}`);
};

// ==================== 项目管理 API ====================

export const getProjectList = (params?: {
  company_id?: number;
  page?: number;
  page_size?: number;
}) => {
  return http.get<ListResponse<Project>>("requirement/projects", { params });
};

export const createProject = (data: {
  company_id: number;
  name: string;
  key: string;
  description?: string;
}) => {
  return http.post<Project>("requirement/projects", data);
};

export const getProjectDetail = (id: number) => {
  return http.get<Project>(`requirement/projects/${id}`);
};

export const updateProject = (id: number, data: Partial<Project>) => {
  return http.put<Project>(`requirement/projects/${id}`, data);
};

export const deleteProject = (id: number) => {
  return http.delete(`requirement/projects/${id}`);
};

export const getProjectMembers = (id: number) => {
  return http.get<ProjectMember[]>(`requirement/projects/${id}/members`);
};

export const addProjectMember = (
  id: number,
  data: { user_id: number; role: string },
) => {
  return http.post<ProjectMember>(`requirement/projects/${id}/members`, data);
};

export const removeProjectMember = (projectId: number, memberId: number) => {
  return http.delete(`requirement/projects/${projectId}/members/${memberId}`);
};

// ==================== 任务列表管理 API ====================

export const getMissionListList = (params?: {
  project_id?: number;
  status?: string;
}) => {
  return http.get<ListResponse<MissionList>>("requirement/mission-lists", {
    params,
  });
};

export const createMissionList = (data: {
  project_id: number;
  name: string;
  type: "sprint" | "version" | "module";
  description?: string;
  color?: string;
  start_date?: string;
  end_date?: string;
}) => {
  return http.post<MissionList>("requirement/mission-lists", data);
};

export const getMissionListDetail = (id: number) => {
  return http.get<MissionList>(`requirement/mission-lists/${id}`);
};

export const updateMissionList = (id: number, data: Partial<MissionList>) => {
  return http.put<MissionList>(`requirement/mission-lists/${id}`, data);
};

export const deleteMissionList = (id: number) => {
  return http.delete(`requirement/mission-lists/${id}`);
};

// ==================== 任务管理 API ====================

export const getMissionList = (params?: {
  mission_list_id?: number;
  project_id?: number;
  status?: string;
  assignee_id?: number;
  page?: number;
  page_size?: number;
}) => {
  return http.get<ListResponse<Mission>>("requirement/missions", { params });
};

export const createMission = (data: {
  mission_list_id: number;
  title: string;
  description?: string;
  type: "feature" | "enhancement" | "bug";
  priority: "P0" | "P1" | "P2" | "P3";
  status?: "todo" | "in_progress" | "done" | "closed";
  assignee_id?: number;
  due_date?: string;
}) => {
  return http.post<Mission>("requirement/missions", data);
};

export const getMissionDetail = (id: number) => {
  return http.get<Mission>(`requirement/missions/${id}`);
};

export const updateMission = (id: number, data: Partial<Mission>) => {
  return http.put<Mission>(`requirement/missions/${id}`, data);
};

export const deleteMission = (id: number) => {
  return http.delete(`requirement/missions/${id}`);
};

export const batchUpdateMissionStatus = (data: {
  mission_ids: number[];
  status: string;
}) => {
  return http.post("requirement/missions/batch-update-status", data);
};

// ==================== 任务评论 API ====================

export const getMissionComments = (missionId: number) => {
  return http.get<MissionComment[]>(
    `requirement/missions/${missionId}/comments`,
  );
};

export const addMissionComment = (
  missionId: number,
  data: { content: string },
) => {
  return http.post<MissionComment>(
    `requirement/missions/${missionId}/comments`,
    data,
  );
};

export const deleteMissionComment = (missionId: number, commentId: number) => {
  return http.delete(`requirement/missions/${missionId}/comments/${commentId}`);
};

// ==================== 任务附件 API ====================

export const getMissionAttachments = (missionId: number) => {
  return http.get<MissionAttachment[]>(
    `requirement/missions/${missionId}/attachments`,
  );
};

export const uploadMissionAttachment = (missionId: number, file: File) => {
  const formData = new FormData();
  formData.append("file", file);
  return http.post<MissionAttachment>(
    `requirement/missions/${missionId}/attachments`,
    formData,
    {
      headers: { "Content-Type": "multipart/form-data" },
    },
  );
};

export const deleteMissionAttachment = (
  missionId: number,
  attachmentId: number,
) => {
  return http.delete(
    `requirement/missions/${missionId}/attachments/${attachmentId}`,
  );
};

// ==================== 统计 API ====================

export const getProjectStatistics = (projectId?: number) => {
  const url = projectId
    ? `requirement/projects/${projectId}/statistics`
    : "requirement/statistics/all";
  return http.get<ProjectStatistics>(url);
};
