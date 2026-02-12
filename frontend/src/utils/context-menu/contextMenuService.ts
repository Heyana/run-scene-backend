import { createApp, h } from "vue";
import ContextMenu from "./ContextMenu";
import type { MenuItem } from "./types";

class ContextMenuService {
  private instance: any = null;
  private container: HTMLDivElement | null = null;

  show(
    e: MouseEvent,
    items: MenuItem[],
    onSelect?: (key: string, item: MenuItem) => void | Promise<void>,
  ) {
    e.preventDefault();
    e.stopPropagation();

    // 清理之前的实例
    this.hide();

    // 计算菜单位置
    const x = Math.min(e.clientX, window.innerWidth - 200);
    const y = Math.min(e.clientY, window.innerHeight - 300);

    // 创建容器
    this.container = document.createElement("div");
    document.body.appendChild(this.container);

    // 创建 Vue 实例
    this.instance = createApp({
      render: () =>
        h(ContextMenu, {
          items,
          position: { x, y },
          onSelect,
          onClose: () => this.hide(),
        }),
    });

    this.instance.mount(this.container);
  }

  hide() {
    if (this.instance) {
      this.instance.unmount();
      this.instance = null;
    }
    if (this.container) {
      document.body.removeChild(this.container);
      this.container = null;
    }
  }
}

// 导出单例
export const contextMenuService = new ContextMenuService();

// 导出便捷函数
export const showContextMenu = (
  e: MouseEvent,
  items: MenuItem[],
  onSelect?: (key: string, item: MenuItem) => void | Promise<void>,
) => {
  contextMenuService.show(e, items, onSelect);
};
