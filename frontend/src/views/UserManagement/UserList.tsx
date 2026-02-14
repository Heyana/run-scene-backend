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
} from "ant-design-vue";
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  LockOutlined,
  UnlockOutlined,
  KeyOutlined,
  UserOutlined,
} from "@ant-design/icons-vue";
import type { TableColumnsType } from "ant-design-vue";
import { api } from "@/api/api";
import "./UserList.less";

interface User {
  id: number;
  username: string;
  email: string;
  real_name?: string;
  phone?: string;
  status: "active" | "disabled" | "locked";
  last_login_at?: string;
  created_at: string;
}

interface Role {
  id: number;
  code: string;
  name: string;
}

export default defineComponent({
  name: "UserList",
  setup() {
    const loading = ref(false);
    const users = ref<User[]>([]);
    const total = ref(0);
    const pagination = reactive({
      current: 1,
      pageSize: 20,
    });

    // 搜索条件
    const searchForm = reactive({
      keyword: "",
      status: undefined as string | undefined,
    });

    // 用户表单
    const userModalVisible = ref(false);
    const userFormRef = ref();
    const userForm = reactive({
      id: undefined as number | undefined,
      username: "",
      password: "",
      email: "",
      real_name: "",
      phone: "",
      role_ids: [] as number[],
    });

    // 重置密码
    const resetPasswordVisible = ref(false);
    const resetPasswordForm = reactive({
      userId: undefined as number | undefined,
      newPassword: "",
    });

    // 角色列表
    const roles = ref<Role[]>([]);

    // 表格列定义
    const columns: TableColumnsType = [
      {
        title: "ID",
        dataIndex: "id",
        width: 80,
      },
      {
        title: "用户名",
        dataIndex: "username",
        width: 150,
      },
      {
        title: "姓名",
        dataIndex: "real_name",
        width: 120,
      },
      {
        title: "邮箱",
        dataIndex: "email",
        width: 200,
      },
      {
        title: "手机号",
        dataIndex: "phone",
        width: 130,
      },
      {
        title: "状态",
        dataIndex: "status",
        width: 100,
        customRender: ({ record }: { record: User }) => {
          const statusMap = {
            active: { color: "success", text: "正常" },
            disabled: { color: "default", text: "禁用" },
            locked: { color: "error", text: "锁定" },
          };
          const status = statusMap[record.status];
          return <Tag color={status.color}>{status.text}</Tag>;
        },
      },
      {
        title: "最后登录",
        dataIndex: "last_login_at",
        width: 180,
        customRender: ({ text }: { text: string }) => {
          return text ? new Date(text).toLocaleString() : "-";
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
        width: 250,
        customRender: ({ record }: { record: User }) => (
          <Space>
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            >
              编辑
            </Button>
            {record.status === "active" ? (
              <Popconfirm
                title="确定要禁用该用户吗？"
                onConfirm={() => handleDisable(record.id)}
              >
                <Button type="link" size="small" icon={<LockOutlined />} danger>
                  禁用
                </Button>
              </Popconfirm>
            ) : (
              <Button
                type="link"
                size="small"
                icon={<UnlockOutlined />}
                onClick={() => handleEnable(record.id)}
              >
                启用
              </Button>
            )}
            <Button
              type="link"
              size="small"
              icon={<KeyOutlined />}
              onClick={() => handleResetPassword(record)}
            >
              重置密码
            </Button>
            <Popconfirm
              title="确定要删除该用户吗？"
              onConfirm={() => handleDelete(record.id)}
            >
              <Button type="link" size="small" icon={<DeleteOutlined />} danger>
                删除
              </Button>
            </Popconfirm>
          </Space>
        ),
      },
    ];

    // 加载用户列表
    const loadUsers = async () => {
      loading.value = true;
      try {
        const res = await api.user.getUserList({
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchForm.keyword,
          status: searchForm.status,
        });
        users.value = res.data.items;
        total.value = res.data.total;
      } catch (error) {
        console.error("加载用户列表失败:", error);
        message.error("加载用户列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 加载角色列表
    const loadRoles = async () => {
      try {
        const res = await api.role.getRoleList();
        roles.value = res.data.items;
      } catch (error) {
        console.error("加载角色列表失败:", error);
      }
    };

    // 搜索
    const handleSearch = () => {
      pagination.current = 1;
      loadUsers();
    };

    // 重置搜索
    const handleReset = () => {
      searchForm.keyword = "";
      searchForm.status = undefined;
      pagination.current = 1;
      loadUsers();
    };

    // 新增用户
    const handleAdd = () => {
      userForm.id = undefined;
      userForm.username = "";
      userForm.password = "";
      userForm.email = "";
      userForm.real_name = "";
      userForm.phone = "";
      userForm.role_ids = [];
      userModalVisible.value = true;
    };

    // 编辑用户
    const handleEdit = (record: User) => {
      userForm.id = record.id;
      userForm.username = record.username;
      userForm.password = "";
      userForm.email = record.email;
      userForm.real_name = record.real_name || "";
      userForm.phone = record.phone || "";
      userForm.role_ids = [];
      userModalVisible.value = true;
    };

    // 保存用户
    const handleSaveUser = async () => {
      try {
        await userFormRef.value.validate();
        loading.value = true;

        if (userForm.id) {
          await api.user.updateUser(userForm.id, {
            email: userForm.email,
            real_name: userForm.real_name,
            phone: userForm.phone,
          });
          if (userForm.role_ids.length > 0) {
            await api.user.assignRoles(userForm.id, userForm.role_ids);
          }
        } else {
          await api.user.createUser({
            username: userForm.username,
            password: userForm.password,
            email: userForm.email,
            real_name: userForm.real_name,
            phone: userForm.phone,
            role_ids: userForm.role_ids,
          });
        }

        message.success(userForm.id ? "更新成功" : "创建成功");
        userModalVisible.value = false;
        loadUsers();
      } catch (error) {
        console.error("保存用户失败:", error);
        message.error("保存用户失败");
      } finally {
        loading.value = false;
      }
    };

    // 禁用用户
    const handleDisable = async (id: number) => {
      try {
        await api.user.disableUser(id);
        message.success("禁用成功");
        loadUsers();
      } catch (error) {
        console.error("禁用用户失败:", error);
        message.error("禁用用户失败");
      }
    };

    // 启用用户
    const handleEnable = async (id: number) => {
      try {
        await api.user.enableUser(id);
        message.success("启用成功");
        loadUsers();
      } catch (error) {
        console.error("启用用户失败:", error);
        message.error("启用用户失败");
      }
    };

    // 重置密码
    const handleResetPassword = (record: User) => {
      resetPasswordForm.userId = record.id;
      resetPasswordForm.newPassword = "";
      resetPasswordVisible.value = true;
    };

    // 确认重置密码
    const handleConfirmResetPassword = async () => {
      try {
        await api.user.resetUserPassword(
          resetPasswordForm.userId!,
          resetPasswordForm.newPassword,
        );
        message.success("密码重置成功");
        resetPasswordVisible.value = false;
      } catch (error) {
        console.error("重置密码失败:", error);
        message.error("重置密码失败");
      }
    };

    // 删除用户
    const handleDelete = async (id: number) => {
      try {
        await api.user.deleteUser(id);
        message.success("删除成功");
        loadUsers();
      } catch (error) {
        console.error("删除用户失败:", error);
        message.error("删除用户失败");
      }
    };

    // 分页变化
    const handleTableChange = (pag: any) => {
      pagination.current = pag.current;
      pagination.pageSize = pag.pageSize;
      loadUsers();
    };

    onMounted(() => {
      loadUsers();
      loadRoles();
    });

    return () => (
      <div class="user-list-page">
        <Card>
          {/* 搜索栏 */}
          <div class="search-bar">
            <Space size="middle">
              <Input
                v-model:value={searchForm.keyword}
                placeholder="搜索用户名、邮箱、姓名"
                prefix={<SearchOutlined />}
                style={{ width: "300px" }}
                onPressEnter={handleSearch}
              />
              <Select
                v-model:value={searchForm.status}
                placeholder="状态"
                allowClear
                style={{ width: "120px" }}
                options={[
                  { label: "正常", value: "active" },
                  { label: "禁用", value: "disabled" },
                  { label: "锁定", value: "locked" },
                ]}
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
                新增用户
              </Button>
            </Space>
          </div>

          {/* 用户表格 */}
          <Table
            loading={loading.value}
            columns={columns}
            dataSource={users.value}
            rowKey="id"
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: total.value,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            scroll={{ x: 1500 }}
            onChange={handleTableChange}
          />
        </Card>

        {/* 用户表单弹窗 */}
        <Modal
          v-model:open={userModalVisible.value}
          title={userForm.id ? "编辑用户" : "新增用户"}
          width={600}
          onOk={handleSaveUser}
          confirmLoading={loading.value}
        >
          <Form
            ref={userFormRef}
            model={userForm}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 16 }}
          >
            <Form.Item
              label="用户名"
              name="username"
              rules={[{ required: true, message: "请输入用户名" }]}
            >
              <Input
                v-model:value={userForm.username}
                placeholder="请输入用户名"
                disabled={!!userForm.id}
              />
            </Form.Item>
            {!userForm.id && (
              <Form.Item
                label="密码"
                name="password"
                rules={[{ required: true, message: "请输入密码" }]}
              >
                <Input.Password
                  v-model:value={userForm.password}
                  placeholder="请输入密码"
                />
              </Form.Item>
            )}
            <Form.Item
              label="邮箱"
              name="email"
              rules={[
                { required: true, message: "请输入邮箱" },
                { type: "email", message: "请输入有效的邮箱地址" },
              ]}
            >
              <Input v-model:value={userForm.email} placeholder="请输入邮箱" />
            </Form.Item>
            <Form.Item label="姓名" name="real_name">
              <Input
                v-model:value={userForm.real_name}
                placeholder="请输入姓名"
              />
            </Form.Item>
            <Form.Item label="手机号" name="phone">
              <Input
                v-model:value={userForm.phone}
                placeholder="请输入手机号"
              />
            </Form.Item>
            <Form.Item label="角色" name="role_ids">
              <Select
                v-model:value={userForm.role_ids}
                mode="multiple"
                placeholder="请选择角色"
                options={roles.value.map((r) => ({
                  label: r.name,
                  value: r.id,
                }))}
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 重置密码弹窗 */}
        <Modal
          v-model:open={resetPasswordVisible.value}
          title="重置密码"
          width={500}
          onOk={handleConfirmResetPassword}
        >
          <Form labelCol={{ span: 6 }} wrapperCol={{ span: 16 }}>
            <Form.Item label="新密码" required>
              <Input.Password
                v-model:value={resetPasswordForm.newPassword}
                placeholder="请输入新密码"
              />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  },
});
