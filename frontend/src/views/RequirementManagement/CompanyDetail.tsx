import { defineComponent, ref, onMounted, computed } from "vue";
import { useRoute } from "vue-router";
import {
  Card,
  Button,
  Table,
  Space,
  Modal,
  Form,
  Select,
  message,
  Tag,
  Popconfirm,
  Avatar,
} from "ant-design-vue";
import {
  PlusOutlined,
  DeleteOutlined,
  TeamOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import type { CompanyMember } from "@/api/models/requirement";
import "./CompanyDetail.less";

export default defineComponent({
  name: "CompanyDetail",
  setup() {
    const route = useRoute();
    const loading = ref(false);
    const members = ref<CompanyMember[]>([]);
    const users = ref<any[]>([]);
    const modalVisible = ref(false);
    const formRef = ref();
    const companyId = computed(() => Number(route.params.companyId));

    const formData = ref({
      user_id: undefined as number | undefined,
      role: "member" as "company_admin" | "member" | "viewer",
    });

    // 加载成员列表
    const loadMembers = async () => {
      loading.value = true;
      try {
        const res = await api.requirement.getCompanyMembers(companyId.value);
        members.value = res.data;
      } catch (error) {
        message.error("加载成员列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 加载用户列表
    const loadUsers = async () => {
      try {
        const res = await api.user.getUserList({ page: 1, page_size: 100 });
        users.value = res.data.items;
      } catch (error) {
        console.error("加载用户列表失败:", error);
      }
    };

    // 显示添加对话框
    const handleAdd = () => {
      formData.value = {
        user_id: undefined,
        role: "member",
      };
      modalVisible.value = true;
      loadUsers();
    };

    // 提交表单
    const handleSubmit = async () => {
      try {
        await formRef.value.validate();
        await api.requirement.addCompanyMember(companyId.value, {
          user_id: formData.value.user_id!,
          role: formData.value.role,
        });
        message.success("添加成功");
        modalVisible.value = false;
        loadMembers();
      } catch (error) {
        console.error("添加失败:", error);
      }
    };

    // 删除成员
    const handleDelete = async (id: number) => {
      try {
        await api.requirement.removeCompanyMember(companyId.value, id);
        message.success("删除成功");
        loadMembers();
      } catch (error) {
        message.error("删除失败");
      }
    };

    // 表格列
    const columns = [
      {
        title: "成员",
        key: "user",
        customRender: ({ record }: { record: CompanyMember }) => (
          <Space>
            <Avatar size={32} src={record.user?.avatar}>
              {record.user?.real_name?.[0] || record.user?.username[0]}
            </Avatar>
            <div>
              <div>{record.user?.real_name || record.user?.username}</div>
              <div style={{ fontSize: "12px", color: "#646a73" }}>
                {record.user?.email}
              </div>
            </div>
          </Space>
        ),
      },
      {
        title: "角色",
        dataIndex: "role",
        key: "role",
        customRender: ({ text }: { text: string }) => {
          const roleMap = {
            company_admin: { label: "管理员", color: "red" },
            member: { label: "成员", color: "blue" },
            viewer: { label: "访客", color: "default" },
          };
          const config = roleMap[text as keyof typeof roleMap];
          return <Tag color={config.color}>{config.label}</Tag>;
        },
      },
      {
        title: "加入时间",
        dataIndex: "joined_at",
        key: "joined_at",
        customRender: ({ text }: { text: string }) => {
          return new Date(text).toLocaleDateString();
        },
      },
      {
        title: "操作",
        key: "action",
        customRender: ({ record }: { record: CompanyMember }) => (
          <Space>
            {record.role !== "company_admin" && (
              <Popconfirm
                title="确定要删除该成员吗？"
                onConfirm={() => handleDelete(record.id)}
              >
                <Button type="link" danger icon={<DeleteOutlined />}>
                  删除
                </Button>
              </Popconfirm>
            )}
          </Space>
        ),
      },
    ];

    onMounted(() => {
      loadMembers();
    });

    return () => (
      <div class="company-detail-page">
        <Card
          title={
            <Space>
              <TeamOutlined />
              <span>公司成员</span>
            </Space>
          }
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              添加成员
            </Button>
          }
        >
          <Table
            dataSource={members.value}
            columns={columns}
            rowKey="id"
            loading={loading.value}
          />
        </Card>

        {/* 添加成员对话框 */}
        <Modal
          v-model:open={modalVisible.value}
          title="添加成员"
          onOk={handleSubmit}
          okText="添加"
          cancelText="取消"
        >
          <Form ref={formRef} model={formData.value} labelCol={{ span: 6 }}>
            <Form.Item
              label="选择用户"
              name="user_id"
              rules={[{ required: true, message: "请选择用户" }]}
            >
              <Select
                v-model:value={formData.value.user_id}
                placeholder="请选择用户"
                showSearch
                filterOption={(input: string, option: any) =>
                  option.label.toLowerCase().includes(input.toLowerCase())
                }
              >
                {users.value.map((user) => (
                  <Select.Option
                    key={user.id}
                    value={user.id}
                    label={user.real_name || user.username}
                  >
                    {user.real_name || user.username} ({user.email})
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
            <Form.Item label="角色" name="role">
              <Select v-model:value={formData.value.role}>
                <Select.Option value="company_admin">管理员</Select.Option>
                <Select.Option value="member">成员</Select.Option>
                <Select.Option value="viewer">访客</Select.Option>
              </Select>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  },
});
