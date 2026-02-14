import { defineComponent, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import {
  Card,
  Button,
  Table,
  Space,
  Modal,
  Form,
  Input,
  message,
  Tag,
  Avatar,
} from "ant-design-vue";
import {
  PlusOutlined,
  TeamOutlined,
  EditOutlined,
  UserOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import type { Company } from "@/api/models/requirement";
import "./CompanyList.less";

export default defineComponent({
  name: "CompanyList",
  setup() {
    const router = useRouter();
    const loading = ref(false);
    const companies = ref<Company[]>([]);
    const modalVisible = ref(false);
    const formRef = ref();
    const formData = ref({
      name: "",
      logo: "",
      description: "",
    });

    // 表格列定义
    const columns = [
      {
        title: "公司名称",
        dataIndex: "name",
        key: "name",
        customRender: ({ record }: { record: Company }) => (
          <Space>
            {record.logo ? (
              <Avatar src={record.logo} />
            ) : (
              <Avatar icon={<TeamOutlined />} />
            )}
            <span>{record.name}</span>
          </Space>
        ),
      },
      {
        title: "描述",
        dataIndex: "description",
        key: "description",
      },
      {
        title: "成员数",
        dataIndex: "member_count",
        key: "member_count",
        width: 100,
        customRender: ({ record }: { record: Company }) => (
          <Tag color="blue">{record.member_count || 0} 人</Tag>
        ),
      },
      {
        title: "项目数",
        dataIndex: "project_count",
        key: "project_count",
        width: 100,
        customRender: ({ record }: { record: Company }) => (
          <Tag color="green">{record.project_count || 0} 个</Tag>
        ),
      },
      {
        title: "创建时间",
        dataIndex: "created_at",
        key: "created_at",
        width: 180,
        customRender: ({ text }: { text: string }) =>
          new Date(text).toLocaleString("zh-CN"),
      },
      {
        title: "操作",
        key: "action",
        width: 200,
        customRender: ({ record }: { record: Company }) => (
          <Space>
            <Button
              type="link"
              size="small"
              onClick={() => handleViewProjects(record)}
            >
              查看项目
            </Button>
            <Button
              type="link"
              size="small"
              icon={<UserOutlined />}
              onClick={() => handleManageMembers(record)}
            >
              成员
            </Button>
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            >
              编辑
            </Button>
          </Space>
        ),
      },
    ];

    // 加载公司列表
    const loadCompanies = async () => {
      loading.value = true;
      try {
        const res = await api.requirement.getCompanyList();
        // 后端直接返回数组，不是 { items: [] } 格式
        companies.value = Array.isArray(res.data)
          ? res.data
          : res.data.items || [];
      } catch (error) {
        message.error("加载公司列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 显示创建对话框
    const handleCreate = () => {
      formData.value = {
        name: "",
        logo: "",
        description: "",
      };
      modalVisible.value = true;
    };

    // 提交表单
    const handleSubmit = async () => {
      try {
        await formRef.value.validate();
        await api.requirement.createCompany(formData.value);
        message.success("创建成功");
        modalVisible.value = false;
        loadCompanies();
      } catch (error) {
        console.error("创建失败:", error);
      }
    };

    // 查看项目
    const handleViewProjects = (company: Company) => {
      router.push(`/requirement-management/companies/${company.id}/projects`);
    };

    // 管理成员
    const handleManageMembers = (company: Company) => {
      router.push(`/requirement-management/companies/${company.id}`);
    };

    // 编辑公司
    const handleEdit = (company: Company) => {
      message.info("编辑功能开发中");
    };

    onMounted(() => {
      loadCompanies();
    });

    return () => (
      <div class="company-list-page">
        <Card
          title={
            <Space>
              <TeamOutlined />
              <span>公司管理</span>
            </Space>
          }
          extra={
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleCreate}
            >
              创建公司
            </Button>
          }
        >
          <Table
            dataSource={companies.value}
            columns={columns}
            loading={loading.value}
            rowKey="id"
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
          />
        </Card>

        {/* 创建/编辑对话框 */}
        <Modal
          v-model:open={modalVisible.value}
          title="创建公司"
          onOk={handleSubmit}
          okText="创建"
          cancelText="取消"
        >
          <Form ref={formRef} model={formData.value} labelCol={{ span: 6 }}>
            <Form.Item
              label="公司名称"
              name="name"
              rules={[{ required: true, message: "请输入公司名称" }]}
            >
              <Input
                v-model:value={formData.value.name}
                placeholder="请输入公司名称"
              />
            </Form.Item>
            <Form.Item label="Logo URL" name="logo">
              <Input
                v-model:value={formData.value.logo}
                placeholder="请输入Logo URL（可选）"
              />
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={formData.value.description}
                placeholder="请输入公司描述（可选）"
                rows={4}
              />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  },
});
