import { defineComponent, ref, onMounted, reactive } from "vue";
import {
  Table,
  Button,
  Input,
  Select,
  Space,
  Tag,
  Modal,
  Form,
  message,
  Popconfirm,
  Card,
  Descriptions,
} from "ant-design-vue";
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
} from "@ant-design/icons-vue";
import type { TableColumnsType } from "ant-design-vue";
import { api } from "@/api/api";

interface Permission {
  id: number;
  code: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
  is_system: boolean;
  created_at: string;
}

export default defineComponent({
  name: "PermissionList",
  setup() {
    const loading = ref(false);
    const permissions = ref<Permission[]>([]);
    const total = ref(0);
    const pagination = reactive({
      current: 1,
      pageSize: 20,
    });

    // 搜索条件
    const searchForm = reactive({
      keyword: "",
      resource: undefined as string | undefined,
      action: undefined as string | undefined,
    });

    // 权限表单
    const permissionModalVisible = ref(false);
    const permissionFormRef = ref();
    const permissionForm = reactive({
      id: undefined as number | undefined,
      code: "",
      name: "",
      resource: "",
      action: "",
      description: "",
    });

    // 资源类型选项
    const resourceOptions = [
      { label: "文档库", value: "documents" },
      { label: "模型库", value: "models" },
      { label: "资产库", value: "assets" },
      { label: "贴图库", value: "textures" },
      { label: "项目管理", value: "projects" },
      { label: "AI 3D", value: "ai3d" },
      { label: "用户管理", value: "users" },
      { label: "角色管理", value: "roles" },
      { label: "权限管理", value: "permissions" },
    ];

    // 操作类型选项
    const actionOptions = [
      { label: "查看", value: "read" },
      { label: "创建", value: "create" },
      { label: "更新", value: "update" },
      { label: "删除", value: "delete" },
      { label: "下载", value: "download" },
      { label: "上传", value: "upload" },
      { label: "分享", value: "share" },
      { label: "管理", value: "admin" },
    ];

    // 表格列定义
    const columns: TableColumnsType = [
      {
        title: "ID",
        dataIndex: "id",
        width: 80,
      },
      {
        title: "权限代码",
        dataIndex: "code",
        width: 200,
      },
      {
        title: "权限名称",
        dataIndex: "name",
        width: 150,
      },
      {
        title: "资源类型",
        dataIndex: "resource",
        width: 120,
        customRender: ({ text }: { text: string }) => {
          const resource = resourceOptions.find((r) => r.value === text);
          return <Tag color="blue">{resource?.label || text}</Tag>;
        },
      },
      {
        title: "操作类型",
        dataIndex: "action",
        width: 100,
        customRender: ({ text }: { text: string }) => {
          const action = actionOptions.find((a) => a.value === text);
          return <Tag color="green">{action?.label || text}</Tag>;
        },
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
        customRender: ({ record }: { record: Permission }) => {
          return record.is_system ? (
            <Tag color="blue">系统权限</Tag>
          ) : (
            <Tag>自定义</Tag>
          );
        },
      },
      {
        title: "操作",
        key: "action",
        fixed: "right",
        width: 150,
        customRender: ({ record }: { record: Permission }) => (
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
            {!record.is_system && (
              <Popconfirm
                title="确定要删除该权限吗？"
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

    // 加载权限列表
    const loadPermissions = async () => {
      loading.value = true;
      try {
        const res = await api.permission.getPermissionList({
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchForm.keyword,
          resource: searchForm.resource,
          action: searchForm.action,
        });
        permissions.value = res.data.items;
        total.value = res.data.total;
      } catch (error) {
        console.error("加载权限列表失败:", error);
        message.error("加载权限列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 搜索
    const handleSearch = () => {
      pagination.current = 1;
      loadPermissions();
    };

    // 重置搜索
    const handleReset = () => {
      searchForm.keyword = "";
      searchForm.resource = undefined;
      searchForm.action = undefined;
      pagination.current = 1;
      loadPermissions();
    };

    // 新增权限
    const handleAdd = () => {
      permissionForm.id = undefined;
      permissionForm.code = "";
      permissionForm.name = "";
      permissionForm.resource = "";
      permissionForm.action = "";
      permissionForm.description = "";
      permissionModalVisible.value = true;
    };

    // 编辑权限
    const handleEdit = (record: Permission) => {
      permissionForm.id = record.id;
      permissionForm.code = record.code;
      permissionForm.name = record.name;
      permissionForm.resource = record.resource;
      permissionForm.action = record.action;
      permissionForm.description = record.description || "";
      permissionModalVisible.value = true;
    };

    // 保存权限
    const handleSavePermission = async () => {
      try {
        await permissionFormRef.value.validate();
        loading.value = true;

        // 自动生成权限代码
        if (!permissionForm.id) {
          permissionForm.code = `${permissionForm.resource}:${permissionForm.action}`;
        }

        if (permissionForm.id) {
          await api.permission.updatePermission(permissionForm.id, {
            name: permissionForm.name,
            description: permissionForm.description,
          });
        } else {
          await api.permission.createPermission({
            code: permissionForm.code,
            name: permissionForm.name,
            resource: permissionForm.resource,
            action: permissionForm.action,
            description: permissionForm.description,
          });
        }

        message.success(permissionForm.id ? "更新成功" : "创建成功");
        permissionModalVisible.value = false;
        loadPermissions();
      } catch (error) {
        console.error("保存权限失败:", error);
        message.error("保存权限失败");
      } finally {
        loading.value = false;
      }
    };

    // 删除权限
    const handleDelete = async (id: number) => {
      try {
        await api.permission.deletePermission(id);
        message.success("删除成功");
        loadPermissions();
      } catch (error) {
        console.error("删除权限失败:", error);
        message.error("删除权限失败");
      }
    };

    // 分页变化
    const handleTableChange = (pag: any) => {
      pagination.current = pag.current;
      pagination.pageSize = pag.pageSize;
      loadPermissions();
    };

    onMounted(() => {
      loadPermissions();
    });

    return () => (
      <div class="permission-list-page">
        <Card>
          {/* 权限说明 */}
          <Descriptions
            bordered
            size="small"
            column={1}
            style={{ marginBottom: "16px" }}
          >
            <Descriptions.Item label={<InfoCircleOutlined />}>
              权限格式：资源:操作，如 documents:read 表示查看文档的权限。
              系统权限不可编辑和删除。
            </Descriptions.Item>
          </Descriptions>

          {/* 搜索栏 */}
          <div class="search-bar" style={{ marginBottom: "16px" }}>
            <Space size="middle">
              <Input
                v-model:value={searchForm.keyword}
                placeholder="搜索权限名称或代码"
                prefix={<SearchOutlined />}
                style={{ width: "250px" }}
                onPressEnter={handleSearch}
              />
              <Select
                v-model:value={searchForm.resource}
                placeholder="资源类型"
                allowClear
                style={{ width: "150px" }}
                options={resourceOptions}
              />
              <Select
                v-model:value={searchForm.action}
                placeholder="操作类型"
                allowClear
                style={{ width: "120px" }}
                options={actionOptions}
              />
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button onClick={handleReset}>重置</Button>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleAdd}
              >
                新增权限
              </Button>
            </Space>
          </div>

          {/* 权限表格 */}
          <Table
            loading={loading.value}
            columns={columns}
            dataSource={permissions.value}
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

        {/* 权限表单弹窗 */}
        <Modal
          v-model:open={permissionModalVisible.value}
          title={permissionForm.id ? "编辑权限" : "新增权限"}
          width={600}
          onOk={handleSavePermission}
          confirmLoading={loading.value}
        >
          <Form
            ref={permissionFormRef}
            model={permissionForm}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 16 }}
          >
            <Form.Item
              label="资源类型"
              name="resource"
              rules={[{ required: true, message: "请选择资源类型" }]}
            >
              <Select
                v-model:value={permissionForm.resource}
                placeholder="请选择资源类型"
                options={resourceOptions}
                disabled={!!permissionForm.id}
              />
            </Form.Item>
            <Form.Item
              label="操作类型"
              name="action"
              rules={[{ required: true, message: "请选择操作类型" }]}
            >
              <Select
                v-model:value={permissionForm.action}
                placeholder="请选择操作类型"
                options={actionOptions}
                disabled={!!permissionForm.id}
              />
            </Form.Item>
            <Form.Item
              label="权限名称"
              name="name"
              rules={[{ required: true, message: "请输入权限名称" }]}
            >
              <Input
                v-model:value={permissionForm.name}
                placeholder="如: 查看文档"
              />
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={permissionForm.description}
                placeholder="请输入权限描述"
                rows={3}
              />
            </Form.Item>
            {!permissionForm.id && (
              <Form.Item label="权限代码">
                <Input
                  value={
                    permissionForm.resource && permissionForm.action
                      ? `${permissionForm.resource}:${permissionForm.action}`
                      : ""
                  }
                  disabled
                  placeholder="自动生成"
                />
              </Form.Item>
            )}
          </Form>
        </Modal>
      </div>
    );
  },
});
