import { defineComponent, ref, computed, watch, type PropType } from "vue";
import { Card, Checkbox, Space, Tag, Button, Collapse } from "ant-design-vue";
import { CheckOutlined, CloseOutlined } from "@ant-design/icons-vue";

interface Permission {
  id: number;
  code: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
}

export default defineComponent({
  name: "PermissionSelector",
  props: {
    permissions: {
      type: Array as PropType<Permission[]>,
      required: true,
    },
    selectedIds: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
  },
  emits: ["update:selectedIds"],
  setup(props, { emit }) {
    const localSelectedIds = ref<number[]>([...props.selectedIds]);

    // 监听 props.selectedIds 变化，同步到本地状态
    watch(
      () => props.selectedIds,
      (newIds) => {
        localSelectedIds.value = [...newIds];
      },
      { immediate: true },
    );

    // 按资源分组
    const groupedByResource = computed(() => {
      const groups: Record<string, Permission[]> = {};
      props.permissions?.forEach((perm) => {
        if (!groups[perm.resource]) {
          groups[perm.resource] = [];
        }
        groups[perm.resource].push(perm);
      });
      return groups;
    });

    // 按操作分组
    const groupedByAction = computed(() => {
      const groups: Record<string, Permission[]> = {};
      props.permissions?.forEach((perm) => {
        if (!groups[perm.action]) {
          groups[perm.action] = [];
        }
        groups[perm.action].push(perm);
      });
      return groups;
    });

    // 资源名称映射
    const resourceNames: Record<string, string> = {
      "*": "全局",
      documents: "文档库",
      models: "模型库",
      assets: "资产库",
      textures: "贴图库",
      projects: "项目管理",
      ai3d: "AI 3D",
      users: "用户管理",
      roles: "角色管理",
      permissions: "权限管理",
    };

    // 操作名称映射
    const actionNames: Record<string, string> = {
      "*": "全部",
      read: "查看",
      create: "创建",
      update: "更新",
      delete: "删除",
      download: "下载",
      upload: "上传",
      share: "分享",
      admin: "管理",
      sync: "同步",
    };

    // 操作颜色映射
    const actionColors: Record<string, string> = {
      "*": "red",
      read: "blue",
      create: "green",
      update: "orange",
      delete: "red",
      download: "cyan",
      upload: "purple",
      share: "geekblue",
      admin: "magenta",
      sync: "lime",
    };

    // 检查某个资源的所有权限是否都被选中
    const isResourceAllSelected = (resource: string) => {
      const perms = groupedByResource.value[resource] || [];
      return (
        perms.length > 0 &&
        perms.every((p) => localSelectedIds.value.includes(p.id))
      );
    };

    // 检查某个资源的部分权限是否被选中
    const isResourceIndeterminate = (resource: string) => {
      const perms = groupedByResource.value[resource] || [];
      const selectedCount = perms.filter((p) =>
        localSelectedIds.value.includes(p.id),
      ).length;
      return selectedCount > 0 && selectedCount < perms.length;
    };

    // 切换资源的所有权限
    const toggleResource = (resource: string, checked: boolean) => {
      const perms = groupedByResource.value[resource] || [];
      if (checked) {
        // 添加所有权限
        perms.forEach((p) => {
          if (!localSelectedIds.value.includes(p.id)) {
            localSelectedIds.value.push(p.id);
          }
        });
      } else {
        // 移除所有权限
        const permIds = perms.map((p) => p.id);
        localSelectedIds.value = localSelectedIds.value.filter(
          (id) => !permIds.includes(id),
        );
      }
      emit("update:selectedIds", localSelectedIds.value);
    };

    // 切换单个权限
    const togglePermission = (permId: number, checked: boolean) => {
      if (checked) {
        if (!localSelectedIds.value.includes(permId)) {
          localSelectedIds.value.push(permId);
        }
      } else {
        localSelectedIds.value = localSelectedIds.value.filter(
          (id) => id !== permId,
        );
      }
      emit("update:selectedIds", localSelectedIds.value);
    };

    // 按操作类型全选
    const selectByAction = (action: string) => {
      const perms = groupedByAction.value[action] || [];
      perms.forEach((p) => {
        if (!localSelectedIds.value.includes(p.id)) {
          localSelectedIds.value.push(p.id);
        }
      });
      emit("update:selectedIds", localSelectedIds.value);
    };

    // 全选
    const selectAll = () => {
      localSelectedIds.value = props.permissions.map((p) => p.id);
      emit("update:selectedIds", localSelectedIds.value);
    };

    // 全不选
    const selectNone = () => {
      localSelectedIds.value = [];
      emit("update:selectedIds", localSelectedIds.value);
    };

    // 反选
    const invertSelection = () => {
      const allIds = props.permissions.map((p) => p.id);
      localSelectedIds.value = allIds.filter(
        (id) => !localSelectedIds.value.includes(id),
      );
      emit("update:selectedIds", localSelectedIds.value);
    };

    return () => (
      <div class="permission-selector">
        {/* 快捷操作 */}
        <Card size="small" style={{ marginBottom: "16px" }}>
          <Space wrap>
            <Button size="small" onClick={selectAll} icon={<CheckOutlined />}>
              全选
            </Button>
            <Button size="small" onClick={selectNone} icon={<CloseOutlined />}>
              全不选
            </Button>
            <Button size="small" onClick={invertSelection}>
              反选
            </Button>
            <span style={{ marginLeft: "16px", color: "#999" }}>
              按操作快选：
            </span>
            {Object.entries(actionNames).map(([action, name]) => (
              <Tag
                key={action}
                color={actionColors[action]}
                style={{ cursor: "pointer" }}
                onClick={() => selectByAction(action)}
              >
                {name}
              </Tag>
            ))}
          </Space>
          <div style={{ marginTop: "8px", color: "#666", fontSize: "12px" }}>
            已选择 {localSelectedIds.value.length} / {props.permissions.length}{" "}
            个权限
          </div>
        </Card>

        {/* 按资源分组显示 */}
        <Collapse defaultActiveKey={["*"]} style={{ background: "#fff" }}>
          {Object.entries(groupedByResource.value)
            .sort(([a], [b]) => {
              // 全局权限排在最前面
              if (a === "*") return -1;
              if (b === "*") return 1;
              return a.localeCompare(b);
            })
            .map(([resource, perms]) => (
              <Collapse.Panel
                key={resource}
                header={
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "8px",
                    }}
                  >
                    <Checkbox
                      checked={isResourceAllSelected(resource)}
                      indeterminate={isResourceIndeterminate(resource)}
                      onChange={(e: any) =>
                        toggleResource(resource, e.target.checked)
                      }
                      onClick={(e: Event) => e.stopPropagation()}
                    />
                    <Tag color={resource === "*" ? "red" : "blue"}>
                      {resourceNames[resource] || resource}
                    </Tag>
                    <span style={{ color: "#999", fontSize: "12px" }}>
                      (
                      {
                        perms.filter((p) =>
                          localSelectedIds.value.includes(p.id),
                        ).length
                      }
                      /{perms.length})
                    </span>
                  </div>
                }
              >
                <Space wrap>
                  {perms.map((perm) => (
                    <Checkbox
                      key={perm.id}
                      checked={localSelectedIds.value.includes(perm.id)}
                      onChange={(e: any) =>
                        togglePermission(perm.id, e.target.checked)
                      }
                    >
                      <Tag color={actionColors[perm.action] || "default"}>
                        {actionNames[perm.action] || perm.action}
                      </Tag>
                      <span style={{ marginLeft: "4px" }}>{perm.name}</span>
                      {perm.description && (
                        <span
                          style={{
                            color: "#999",
                            fontSize: "12px",
                            marginLeft: "4px",
                          }}
                        >
                          ({perm.description})
                        </span>
                      )}
                    </Checkbox>
                  ))}
                </Space>
              </Collapse.Panel>
            ))}
        </Collapse>
      </div>
    );
  },
});
