import {
  defineComponent,
  ref,
  onMounted,
  onUnmounted,
  type PropType,
  Teleport,
} from "vue";
import type { MenuItem, ContextMenuPosition } from "./types";
import "./style.css";
import "@/styles/anima/index.css";

export default defineComponent({
  name: "ContextMenu",
  props: {
    items: {
      type: Array as PropType<MenuItem[]>,
      required: true,
    },
    position: {
      type: Object as PropType<ContextMenuPosition>,
      required: true,
    },
    onSelect: {
      type: Function as PropType<
        (key: string, item: MenuItem) => void | Promise<void>
      >,
    },
  },
  emits: ["close"],
  setup(props, { emit }) {
    const menuRef = ref<HTMLDivElement>();
    const activeSubmenu = ref<string | null>(null);
    const loadingItems = ref<Set<string>>(new Set());

    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
        emit("close");
      }
    };

    const handleItemClick = async (item: MenuItem) => {
      if (item.disabled || item.children || loadingItems.value.has(item.key))
        return;

      try {
        loadingItems.value.add(item.key);

        const result = item.onClick?.(item);
        if (result instanceof Promise) {
          await result;
        }

        const selectResult = props.onSelect?.(item.key, item);
        if (selectResult instanceof Promise) {
          await selectResult;
        }
      } catch (error) {
        console.error("Menu item action failed:", error);
      } finally {
        loadingItems.value.delete(item.key);
        emit("close");
      }
    };

    const handleMouseEnter = (item: MenuItem) => {
      if (item.children) {
        activeSubmenu.value = item.key;
      }
    };

    const getItemTypeClass = (item: MenuItem) => {
      const type = item.type || "default";
      if (type === "delete") return "danger";
      return type;
    };

    onMounted(() => {
      document.addEventListener("click", handleClickOutside);
      document.addEventListener("contextmenu", handleClickOutside);
    });

    onUnmounted(() => {
      document.removeEventListener("click", handleClickOutside);
      document.removeEventListener("contextmenu", handleClickOutside);
    });

    const renderMenuItem = (item: MenuItem) => {
      const hasChildren = item.children && item.children.length > 0;
      const isActive = activeSubmenu.value === item.key;
      const isLoading = loadingItems.value.has(item.key);

      return (
        <div
          key={item.key}
          class={[
            "context-menu-item",
            `context-menu-item-${getItemTypeClass(item)}`,
            {
              disabled: item.disabled,
              "has-children": hasChildren,
              divided: item.divided,
              loading: isLoading,
            },
          ]}
          onClick={() => handleItemClick(item)}
          onMouseenter={() => handleMouseEnter(item)}
        >
          {isLoading ? (
            <span class="context-menu-icon context-menu-loading">⏳</span>
          ) : (
            item.icon && <span class="context-menu-icon">{item.icon}</span>
          )}
          <span class="context-menu-label">{item.label}</span>
          {hasChildren && <span class="context-menu-arrow">▶</span>}

          {hasChildren && isActive && (
            <div class="context-menu-submenu">
              {item.children!.map((child) => renderMenuItem(child))}
            </div>
          )}
        </div>
      );
    };

    return () => (
      <Teleport to="body">
        <div
          ref={menuRef}
          class="context-menu zoom-in"
          style={{
            left: `${props.position.x}px`,
            top: `${props.position.y}px`,
          }}
        >
          {props.items.map((item) => renderMenuItem(item))}
        </div>
      </Teleport>
    );
  },
});
