import { defineComponent, ref, onMounted, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import {
  Card,
  Button,
  Row,
  Col,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Tag,
  Progress,
  Empty,
} from "ant-design-vue";
import {
  PlusOutlined,
  ProjectOutlined,
  CalendarOutlined,
  TeamOutlined,
  CheckCircleOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import type { Project } from "@/api/models/requirement";
import "./ProjectList.less";

export default defineComponent({
  name: "ProjectList",
  setup() {
    const router = useRouter();
    const route = useRoute();
    const loading = ref(false);
    const projects = ref<Project[]>([]);
    const modalVisible = ref(false);
    const formRef = ref();
    const companyId = computed(() => route.params.companyId as string);

    const formData = ref({
      name: "",
      key: "",
      description: "",
    });

    // 加载项目列表
    const loadProjects = async () => {
      loading.value = true;
      try {
        const params = companyId.value
          ? { company_id: Number(companyId.value) }
          : undefined;
        const res = await api.requirement.getProjectList(params);
        // 后端直接返回数组，不是 { items: [] } 格式
        projects.value = Array.isArray(res.data)
          ? res.data
          : res.data.items || [];
      } catch (error) {
        message.error("加载项目列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 显示创建对话框
    const handleCreate = () => {
      formData.value = {
        name: "",
        key: "",
        description: "",
      };
      modalVisible.value = true;
    };

    // 提交表单
    const handleSubmit = async () => {
      try {
        await formRef.value.validate();
        await api.requirement.createProject({
          company_id: Number(companyId.value) || 1,
          ...formData.value,
        });
        message.success("创建成功");
        modalVisible.value = false;
        loadProjects();
      } catch (error) {
        console.error("创建失败:", error);
      }
    };

    // 查看项目看板
    const handleViewBoard = (project: Project) => {
      router.push(`/requirement-management/projects/${project.id}/board`);
    };

    // 查看项目统计
    const handleViewStatistics = (project: Project) => {
      router.push(`/requirement-management/projects/${project.id}/statistics`);
    };

    // 计算完成率
    const getCompletionRate = (project: Project) => {
      // TODO: 从实际数据计算
      return Math.floor(Math.random() * 100);
    };

    onMounted(() => {
      loadProjects();
    });

    return () => (
      <div class="project-list-page">
        <Card
          title={
            <Space>
              <ProjectOutlined />
              <span>项目列表</span>
            </Space>
          }
          extra={
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleCreate}
            >
              创建项目
            </Button>
          }
        >
          {projects.value.length === 0 ? (
            <Empty description="暂无项目" />
          ) : (
            <Row gutter={[16, 16]}>
              {projects.value.map((project) => (
                <Col key={project.id} xs={24} sm={12} lg={8} xl={6}>
                  <Card
                    class="project-card"
                    hoverable
                    onClick={() => handleViewBoard(project)}
                  >
                    <div class="project-header">
                      <div class="project-icon">
                        <ProjectOutlined />
                      </div>
                      <Tag color="blue">{project.key}</Tag>
                    </div>
                    <h3 class="project-name">{project.name}</h3>
                    <p class="project-description">
                      {project.description || "暂无描述"}
                    </p>

                    <div class="project-stats">
                      <div class="stat-item">
                        <TeamOutlined />
                        <span>{project.member_count || 0} 成员</span>
                      </div>
                      <div class="stat-item">
                        <CheckCircleOutlined />
                        <span>{project.mission_count || 0} 任务</span>
                      </div>
                    </div>

                    <div class="project-progress">
                      <div class="progress-label">
                        <span>完成度</span>
                        <span>{getCompletionRate(project)}%</span>
                      </div>
                      <Progress
                        percent={getCompletionRate(project)}
                        showInfo={false}
                        strokeColor="#52c41a"
                      />
                    </div>

                    <div
                      class="project-actions"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <Button
                        type="link"
                        size="small"
                        onClick={() => handleViewBoard(project)}
                      >
                        任务看板
                      </Button>
                      <Button
                        type="link"
                        size="small"
                        onClick={() => handleViewStatistics(project)}
                      >
                        统计报表
                      </Button>
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          )}
        </Card>

        {/* 创建项目对话框 */}
        <Modal
          v-model:open={modalVisible.value}
          title="创建项目"
          onOk={handleSubmit}
          okText="创建"
          cancelText="取消"
          width={600}
        >
          <Form ref={formRef} model={formData.value} labelCol={{ span: 6 }}>
            <Form.Item
              label="项目名称"
              name="name"
              rules={[{ required: true, message: "请输入项目名称" }]}
            >
              <Input
                v-model:value={formData.value.name}
                placeholder="请输入项目名称"
              />
            </Form.Item>
            <Form.Item
              label="项目标识"
              name="key"
              rules={[
                { required: true, message: "请输入项目标识" },
                { pattern: /^[A-Z]{2,10}$/, message: "2-10个大写字母" },
              ]}
            >
              <Input
                v-model:value={formData.value.key}
                placeholder="例如: FE, BE, TEST"
                maxlength={10}
              />
            </Form.Item>
            <Form.Item label="项目描述" name="description">
              <Input.TextArea
                v-model:value={formData.value.description}
                placeholder="请输入项目描述（可选）"
                rows={4}
              />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  },
});
