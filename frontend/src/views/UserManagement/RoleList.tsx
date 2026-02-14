import { defineComponent, ref, onMounted, reactive } from "vue";
import {
  Table,
  Button,
  Input,
  Space,
  Tag,
  Modal,
  Form,
  message,
  Popconfirm,
  Card,
  Tabs,
} from "ant-design-vue";
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyOutlined,
} from "@ant-design/icons-vue";
import type { TableColumnsType } from "ant-design-vue";
import { api } from "@/api/api";
import PermissionSelector from "@/components/PermissionSelector";

interface Role {
  id: number;
  code: string;
  name: string;
  description?: string;
  is_system: boolean;
  created_at: string;
}

interface Permission {
  id: number;
  code: string;
  name: string;
  resource: string;
  action: string;
}

interface PermissionGroup {
  id: number;
  code: string;
  name: string;
  description?: string;
}

export default defineComponent({
  name: "RoleList",
  setup() {
    const loading = ref(false);
    const roles = ref<Role[]>([]);
    const total = ref(0);
    const pagination = reactive({
      current: 1,
      pageSize: 20,
    });

    const searchKeyword = ref("");

    // 角色表单
    const roleModalVisible = ref(false);
    const roleFormRef = ref();
    const roleForm = reactive({
      id: undefined as number | undefined,
      code: "",
      name: "",
      description: "",
    });

    // 权限配置
    const permissionModalVisible = ref(false);
    const currentRoleId = ref<number>();
    const permissions = ref<Permission[]>([]);
    const permissionGroups = ref<PermissionGroup[]>([]);
    const selectedPermissions = ref<number[]>([]);
    const selectedPermissionGroups = ref<number[]>([]);
    const selectedPermissionKeys = ref<string[]>([]);
    const selectedPermissionGroupKeys = ref<string[]>([]);

    // 表格列定义
    const columns: TableColumnsType = [
      {
        title: "ID",
        dataIndex: "id",
        width: 80,
      },
      {
        title: "角色代码",
        dataIndex: "code",
        width: 150,
      },
      {
        title: "角色名称",
        dataIndex: "name",
        width: 150,
      },
      {
        title: "描述",
        dataIndex: "description",
        ellipsis: true,
      },
      {
        title: "类型",
        dataIndex: "is_system",
        width: 100,
        customRender: ({ record }: { record: Role }) => {
          return record.is_system ? (
            <Tag color="blue">系统角色</Tag>
          ) : (
            <Tag>自定义</Tag>
          );
        },
      },
      {
        title: "创建时间",
        dataIndex: "created_at",
        width: 180,
        customRender: ({ text }: { text: string }) => {
          return new Date(text).toLocaleString();
        },
      },
      {
        title: "操作",
        key: "action",
        fixed: "right",
        width: 200,
        customRender: ({ record }: { record: Role }) => (
          <Space>
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              disabled={record.is_system}
            >
              编辑
            </Button>
            <Button
              type="link"
              size="small"
              icon={<SafetyOutlined />}
              onClick={() => handleConfigPermission(record)}
            >
              配置权限
            </Button>
            {!record.is_system && (
              <Popconfirm
                title="确定要删除该角色吗？"
                onConfirm={() => handleDelete(record.id)}
              >
                <Button
                  type="link"
                  size="small"
                  icon={<DeleteOutlined />}
                  danger
                >
                  删除
                </Button>
              </Popconfirm>
            )}
          </Space>
        ),
      },
    ];

    // 加载角色列表
    const loadRoles = async () => {
      loading.value = true;
      try {
        const res = await api.role.getRoleList({
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchKeyword.value,
        });
        roles.value = res.data.items;
        total.value = res.data.total;
      } catch (error) {
        console.error("加载角色列表失败:", error);
        message.error("加载角色列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 加载权限列表
    const loadPermissions = async () => {
      try {
        const res = await api.permission.getPermissionList({
          page: 1,
          page_size: 1000, // 加载所有权限
        });
        permissions.value = res.data.items;
      } catch (error) {
        console.error("加载权限列表失败:", error);
      }
    };

    // 加载权限组列表
    const loadPermissionGroups = async () => {
      try {
        const res = await api.permission.getPermissionGroupList({
          page: 1,
          page_size: 1000, // 加载所有权限组
        });
        permissionGroups.value = res.data.items;
      } catch (error) {
        console.error("加载权限组列表失败:", error);
      }
    };

    // 搜索
    const handleSearch = () => {
      pagination.current = 1;
      loadRoles();
    };

    // 新增角色
    const handleAdd = () => {
      roleForm.id = undefined;
      roleForm.code = "";
      roleForm.name = "";
      roleForm.description = "";
      roleModalVisible.value = true;
    };

    // 编辑角色
    const handleEdit = (record: Role) => {
      roleForm.id = record.id;
      roleForm.code = record.code;
      roleForm.name = record.name;
      roleForm.description = record.description || "";
      roleModalVisible.value = true;
    };

    // 保存角色
    const handleSaveRole = async () => {
      try {
        await roleFormRef.value.validate();
        loading.value = true;

        if (roleForm.id) {
          await api.role.updateRole(roleForm.id, {
            name: roleForm.name,
            description: roleForm.description,
          });
        } else {
          await api.role.createRole({
            code: roleForm.code,
            name: roleForm.name,
            description: roleForm.description,
          });
        }

        message.success(roleForm.id ? "更新成功" : "创建成功");
        roleModalVisible.value = false;
        loadRoles();
      } catch (error) {
        console.error("保存角色失败:", error);
        message.error("保存角色失败");
      } finally {
        loading.value = false;
      }
    };

    // 配置权限
    const handleConfigPermission = async (record: Role) => {
      currentRoleId.value = record.id;

      try {
        const res = await api.role.getRolePermissions(record.id);
        console.log("加载角色权限:", res.data);

        // 直接分配的权限
        selectedPermissions.value = res.data.permission_ids || [];
        selectedPermissionGroups.value = res.data.permission_group_ids || [];
        selectedPermissionKeys.value = (res.data.permission_ids || []).map(
          String,
        );
        selectedPermissionGroupKeys.value = (
          res.data.permission_group_ids || []
        ).map(String);

        // 如果有权限组，获取权限组包含的权限并合并显示
        if (
          res.data.permission_group_ids &&
          res.data.permission_group_ids.length > 0
        ) {
          const groupPermissionIds = new Set<number>();

          // 获取每个权限组的详情
          for (const groupId of res.data.permission_group_ids) {
            try {
              const groupDetail =
                await api.permission.getPermissionGroupDetail(groupId);
              if (groupDetail.data.permissions) {
                groupDetail.data.permissions.forEach((p: any) => {
                  groupPermissionIds.add(p.id);
                });
              }
            } catch (err) {
              console.error(`获取权限组${groupId}详情失败:`, err);
            }
          }

          // 合并直接权限和权限组权限（用于显示）
          const allPermissionIds = new Set([
            ...selectedPermissions.value,
            ...Array.from(groupPermissionIds),
          ]);
          selectedPermissions.value = Array.from(allPermissionIds);
        }

        console.log("最终显示的权限IDs:", selectedPermissions.value);
      } catch (error) {
        console.error("加载角色权限失败:", error);
        selectedPermissions.value = [];
        selectedPermissionGroups.value = [];
        selectedPermissionKeys.value = [];
        selectedPermissionGroupKeys.value = [];
      }

      permissionModalVisible.value = true;
    };

    // 保存权限配置
    const handleSavePermissions = async () => {
      try {
        loading.value = true;

        await api.role.assignRolePermissions(currentRoleId.value!, {
          permission_ids: selectedPermissions.value,
          permission_group_ids: selectedPermissionGroups.value,
        });

        message.success("权限配置成功");
        permissionModalVisible.value = false;
      } catch (error) {
        console.error("配置权限失败:", error);
        message.error("配置权限失败");
      } finally {
        loading.value = false;
      }
    };

    // 删除角色
    const handleDelete = async (id: number) => {
      try {
        await api.role.deleteRole(id);
        message.success("删除成功");
        loadRoles();
      } catch (error) {
        console.error("删除角色失败:", error);
        message.error("删除角色失败");
      }
    };

    // 分页变化
    const handleTableChange = (pag: any) => {
      pagination.current = pag.current;
      pagination.pageSize = pag.pageSize;
      loadRoles();
    };

    onMounted(() => {
      loadRoles();
      loadPermissions();
      loadPermissionGroups();
    });

    return () => (
      <div class="role-list-page">
        <Card>
          {/* 搜索栏 */}
          <div class="search-bar" style={{ marginBottom: "16px" }}>
            <Space size="middle">
              <Input
                v-model:value={searchKeyword.value}
                placeholder="搜索角色名称或代码"
                prefix={<SearchOutlined />}
                style={{ width: "300px" }}
                onPressEnter={handleSearch}
              />
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleAdd}
              >
                新增角色
              </Button>
            </Space>
          </div>

          {/* 角色表格 */}
          <Table
            loading={loading.value}
            columns={columns}
            dataSource={roles.value}
            rowKey="id"
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: total.value,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            onChange={handleTableChange}
          />
        </Card>

        {/* 角色表单弹窗 */}
        <Modal
          v-model:open={roleModalVisible.value}
          title={roleForm.id ? "编辑角色" : "新增角色"}
          width={600}
          onOk={handleSaveRole}
          confirmLoading={loading.value}
        >
          <Form
            ref={roleFormRef}
            model={roleForm}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 16 }}
          >
            <Form.Item
              label="角色代码"
              name="code"
              rules={[{ required: true, message: "请输入角色代码" }]}
            >
              <Input
                v-model:value={roleForm.code}
                placeholder="如: custom_manager"
                disabled={!!roleForm.id}
              />
            </Form.Item>
            <Form.Item
              label="角色名称"
              name="name"
              rules={[{ required: true, message: "请输入角色名称" }]}
            >
              <Input
                v-model:value={roleForm.name}
                placeholder="请输入角色名称"
              />
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={roleForm.description}
                placeholder="请输入角色描述"
                rows={3}
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 权限配置弹窗 */}
        <Modal
          v-model:open={permissionModalVisible.value}
          title="配置权限"
          width={1200}
          style={{ top: "20px" }}
          bodyStyle={{
            maxHeight: "calc(100vh - 200px)",
            overflowY: "auto",
          }}
          onOk={handleSavePermissions}
          confirmLoading={loading.value}
        >
          <Tabs defaultActiveKey="permissions">
            <Tabs.TabPane key="permissions" tab="权限配置">
              <PermissionSelector
                permissions={permissions.value}
                selectedIds={selectedPermissions.value}
                onUpdate:selectedIds={(ids: number[]) => {
                  selectedPermissions.value = ids;
                }}
              />
            </Tabs.TabPane>
            <Tabs.TabPane key="groups" tab="权限组（快捷配置）">
              <Card size="small">
                <Space direction="vertical" style={{ width: "100%" }}>
                  {permissionGroups.value.map((group) => (
                    <div
                      key={group.id}
                      style={{ cursor: "pointer" }}
                      onClick={() => {
                        const index = selectedPermissionGroups.value.indexOf(
                          group.id,
                        );
                        if (index > -1) {
                          selectedPermissionGroups.value.splice(index, 1);
                        } else {
                          selectedPermissionGroups.value.push(group.id);
                        }
                      }}
                    >
                      <Card
                        size="small"
                        hoverable
                        style={{
                          border: selectedPermissionGroups.value.includes(
                            group.id,
                          )
                            ? "2px solid #1890ff"
                            : "1px solid #d9d9d9",
                        }}
                      >
                        <div
                          style={{
                            display: "flex",
                            alignItems: "center",
                            gap: "8px",
                          }}
                        >
                          {selectedPermissionGroups.value.includes(
                            group.id,
                          ) && <Tag color="blue">已选</Tag>}
                          {(group as any).is_system && (
                            <Tag color="green">系统</Tag>
                          )}
                          <strong>{group.name}</strong>
                          {group.description && (
                            <span style={{ color: "#999", fontSize: "12px" }}>
                              - {group.description}
                            </span>
                          )}
                        </div>
                      </Card>
                    </div>
                  ))}
                </Space>
              </Card>
            </Tabs.TabPane>
          </Tabs>
        </Modal>
      </div>
    );
  },
});
