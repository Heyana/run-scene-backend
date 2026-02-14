// 公司相关类型
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
  user?: User;
}

// 项目相关类型
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
  user?: User;
}

// 任务列表相关类型
export interface MissionList {
  id: number;
  project_id: number;
  name: string;
  type: "sprint" | "version" | "module";
  description?: string;
  status: "planning" | "active" | "completed";
  start_date?: string;
  end_date?: string;
  sort_order: number;
  mission_count?: number;
  created_at: string;
  updated_at: string;
}

// 任务相关类型
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
  assignee?: User;
  reporter?: User;
  tags?: MissionTag[];
  comments?: MissionComment[];
  attachments?: MissionAttachment[];
}

// 任务评论
export interface MissionComment {
  id: number;
  mission_id: number;
  user_id: number;
  content: string;
  parent_id?: number;
  created_at: string;
  user?: User;
}

// 任务附件
export interface MissionAttachment {
  id: number;
  mission_id: number;
  user_id: number;
  file_name: string;
  file_path: string;
  file_size: number;
  file_type: string;
  created_at: string;
  user?: User;
}

// 任务标签
export interface MissionTag {
  id: number;
  name: string;
  color: string;
  created_at: string;
}

// 任务关联
export interface MissionRelation {
  id: number;
  mission_id: number;
  target_mission_id: number;
  relation_type: "blocks" | "blocked_by" | "relates_to" | "duplicates";
  created_at: string;
  target_mission?: Mission;
}

// 用户类型
export interface User {
  id: number;
  username: string;
  email: string;
  real_name?: string;
  avatar?: string;
}

// 统计数据
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

// API响应类型
export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
  timestamp: number;
}

export interface PaginationResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
