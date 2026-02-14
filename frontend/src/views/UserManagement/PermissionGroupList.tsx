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
} from "ant-design-vue";
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  KeyOutlined,
} from "@ant-design/icons-vue";
import type { TableColumnsType } from "ant-design-vue";
import { api } from "@/api/api";
import PermissionSelector from "@/components/PermissionSelector";

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
  is_system: boolean;
  created_at: string;
  permissions?: Permission[];
}

export default defineComponent({
  name: "PermissionGroupList",
  setup() {
    const loading = ref(false);
    const groups = ref<PermissionGroup[]>([]);
    const total = ref(0);
    const pagination = reactive({
      current: 1,
      pageSize: 20,
    });

    const searchKeyword = ref("");

    // 权限组表单
    const groupModalVisible = ref(false);
    const groupFormRef = ref();
    const groupForm = reactive({
      id: undefined as number | undefined,
      code: "",
      name: "",
      description: "",
    });

    // 权限配置
    const permissionModalVisible = ref(false);
    const currentGroupId = ref<number>();
    const permissions = ref<Permission[]>([]);
    const selectedPermissions = ref<number[]>([]);

    // 表格列定义
    const columns: TableColumnsType = [
      {
        title: "ID",
        dataIndex: "id",
        width: 80,
      },
      {
        title: "权限组代码",
        dataIndex: "code",
        width: 200,
      },
      {
        title: "权限组名称",
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
        customRender: ({ record }: { record: PermissionGroup }) => {
          return record.is_system ? (
            <Tag color="blue">系统权限组</Tag>
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
        customRender: ({ record }: { record: PermissionGroup }) => (
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
              icon={<KeyOutlined />}
              onClick={() => handleConfigPermission(record)}
            >
              配置权限
            </Button>
            {!record.is_system && (
              <Popconfirm
                title="确定要删除该权限组吗？"
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

    // 加载权限组列表
    const loadGroups = async () => {
      loading.value = true;
      try {
        const res = await api.permission.getPermissionGroupList({
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchKeyword.value,
        });
        groups.value = res.data.items;
        total.value = res.data.total;
      } catch (error) {
        console.error("加载权限组列表失败:", error);
        message.error("加载权限组列表失败");
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

    // 搜索
    const handleSearch = () => {
      pagination.current = 1;
      loadGroups();
    };

    // 新增权限组
    const handleAdd = () => {
      groupForm.id = undefined;
      groupForm.code = "";
      groupForm.name = "";
      groupForm.description = "";
      groupModalVisible.value = true;
    };

    // 编辑权限组
    const handleEdit = (record: PermissionGroup) => {
      groupForm.id = record.id;
      groupForm.code = record.code;
      groupForm.name = record.name;
      groupForm.description = record.description || "";
      groupModalVisible.value = true;
    };

    // 保存权限组
    const handleSaveGroup = async () => {
      try {
        await groupFormRef.value.validate();
        loading.value = true;

        if (groupForm.id) {
          await api.permission.updatePermissionGroup(groupForm.id, {
            name: groupForm.name,
            description: groupForm.description,
          });
        } else {
          await api.permission.createPermissionGroup({
            code: groupForm.code,
            name: groupForm.name,
            description: groupForm.description,
          });
        }

        message.success(groupForm.id ? "更新成功" : "创建成功");
        groupModalVisible.value = false;
        loadGroups();
      } catch (error) {
        console.error("保存权限组失败:", error);
        message.error("保存权限组失败");
      } finally {
        loading.value = false;
      }
    };

    // 配置权限
    const handleConfigPermission = async (record: PermissionGroup) => {
      currentGroupId.value = record.id;

      try {
        const res = await api.permission.getPermissionGroupDetail(record.id);
        const groupPermissions = (res.data as any).permissions || [];
        selectedPermissions.value = groupPermissions.map((p: any) => p.id);
      } catch (error) {
        console.error("加载权限组权限失败:", error);
        selectedPermissions.value = [];
      }

      permissionModalVisible.value = true;
    };

    // 保存权限配置
    const handleSavePermissions = async () => {
      try {
        loading.value = true;

        await api.permission.addPermissionsToGroup(
          currentGroupId.value!,
          selectedPermissions.value,
        );

        message.success("权限配置成功");
        permissionModalVisible.value = false;
      } catch (error) {
        console.error("配置权限失败:", error);
        message.error("配置权限失败");
      } finally {
        loading.value = false;
      }
    };

    // 删除权限组
    const handleDelete = async (id: number) => {
      try {
        await api.permission.deletePermissionGroup(id);
        message.success("删除成功");
        loadGroups();
      } catch (error) {
        console.error("删除权限组失败:", error);
        message.error("删除权限组失败");
      }
    };

    // 分页变化
    const handleTableChange = (pag: any) => {
      pagination.current = pag.current;
      pagination.pageSize = pag.pageSize;
      loadGroups();
    };

    onMounted(() => {
      loadGroups();
      loadPermissions();
    });

    return () => (
      <div class="permission-group-list-page">
        <Card>
          {/* 搜索栏 */}
          <div class="search-bar" style={{ marginBottom: "16px" }}>
            <Space size="middle">
              <Input
                v-model:value={searchKeyword.value}
                placeholder="搜索权限组名称或代码"
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
                新增权限组
              </Button>
            </Space>
          </div>

          {/* 权限组表格 */}
          <Table
            loading={loading.value}
            columns={columns}
            dataSource={groups.value}
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

        {/* 权限组表单弹窗 */}
        <Modal
          v-model:open={groupModalVisible.value}
          title={groupForm.id ? "编辑权限组" : "新增权限组"}
          width={600}
          onOk={handleSaveGroup}
          confirmLoading={loading.value}
        >
          <Form
            ref={groupFormRef}
            model={groupForm}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 16 }}
          >
            <Form.Item
              label="权限组代码"
              name="code"
              rules={[{ required: true, message: "请输入权限组代码" }]}
            >
              <Input
                v-model:value={groupForm.code}
                placeholder="如: custom_group"
                disabled={!!groupForm.id}
              />
            </Form.Item>
            <Form.Item
              label="权限组名称"
              name="name"
              rules={[{ required: true, message: "请输入权限组名称" }]}
            >
              <Input
                v-model:value={groupForm.name}
                placeholder="请输入权限组名称"
              />
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={groupForm.description}
                placeholder="请输入权限组描述"
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
          <PermissionSelector
            permissions={permissions.value}
            selectedIds={selectedPermissions.value}
            onUpdate:selectedIds={(ids: number[]) => {
              selectedPermissions.value = ids;
            }}
          />
        </Modal>
      </div>
    );
  },
});
