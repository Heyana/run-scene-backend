// 右键菜单类型定义
export type MenuItemType =
  | "default"
  | "primary"
  | "success"
  | "warning"
  | "danger"
  | "delete";

export interface MenuItem {
  label: string;
  key: string;
  icon?: string;
  disabled?: boolean;
  divided?: boolean;
  type?: MenuItemType;
  children?: MenuItem[];
  onClick?: (item: MenuItem) => void | Promise<void>;
}

export interface ContextMenuProps {
  items: MenuItem[];
  onSelect?: (key: string, item: MenuItem) => void | Promise<void>;
}

export interface ContextMenuPosition {
  x: number;
  y: number;
}
