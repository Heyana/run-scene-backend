import { defineComponent, ref, onMounted, computed } from "vue";
import { useRoute } from "vue-router";
import {
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Table,
  Tag,
  Space,
  Spin,
  message,
} from "ant-design-vue";
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  RocketOutlined,
  CloseCircleOutlined,
  UserOutlined,
  BugOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import type { ProjectStatistics } from "@/api/models/requirement";
import "./Statistics.less";

export default defineComponent({
  name: "Statistics",
  setup() {
    const route = useRoute();
    const loading = ref(false);
    const projectId = computed(() => {
      const id = route.params.projectId;
      return id ? Number(id) : undefined;
    });
    const dateRange = ref<[string, string]>();

    const stats = ref<ProjectStatistics>({
      total_missions: 0,
      completed_missions: 0,
      in_progress_missions: 0,
      todo_missions: 0,
      overdue_missions: 0,
      completion_rate: 0,
      by_type: {},
      by_priority: {},
      by_assignee: [],
    });

    // 加载统计数据
    const loadStatistics = async () => {
      loading.value = true;
      try {
        const res = await api.requirement.getProjectStatistics(projectId.value);
        stats.value = res.data;
      } catch (error) {
        console.error("加载统计数据失败:", error);
        message.error("加载统计数据失败");
      } finally {
        loading.value = false;
      }
    };

    // 成员任务表格列
    const columns = [
      {
        title: "成员",
        dataIndex: "username",
        key: "username",
      },
      {
        title: "任务数",
        dataIndex: "count",
        key: "count",
        sorter: (a: any, b: any) => a.count - b.count,
      },
      {
        title: "占比",
        key: "percentage",
        customRender: ({ record }: any) => {
          const percentage = (
            (record.count / stats.value.total_missions) *
            100
          ).toFixed(1);
          return `${percentage}%`;
        },
      },
    ];

    onMounted(() => {
      loadStatistics();
    });

    return () => (
      <div class="statistics-page">
        <Spin spinning={loading.value}>
          {/* 概览统计 */}
          <Row gutter={[16, 16]} style={{ marginBottom: "24px" }}>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="总任务数"
                  value={stats.value.total_missions}
                  prefix={<RocketOutlined />}
                  valueStyle={{ color: "#1890ff" }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="已完成"
                  value={stats.value.completed_missions}
                  prefix={<CheckCircleOutlined />}
                  valueStyle={{ color: "#52c41a" }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="进行中"
                  value={stats.value.in_progress_missions}
                  prefix={<ClockCircleOutlined />}
                  valueStyle={{ color: "#faad14" }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="已逾期"
                  value={stats.value.overdue_missions}
                  prefix={<CloseCircleOutlined />}
                  valueStyle={{ color: "#ff4d4f" }}
                />
              </Card>
            </Col>
          </Row>

          {/* 完成率 */}
          <Card title="完成率" style={{ marginBottom: "24px" }}>
            <Progress
              percent={stats.value.completion_rate}
              strokeColor={{
                "0%": "#108ee9",
                "100%": "#87d068",
              }}
              style={{ marginBottom: "16px" }}
            />
            <Row gutter={16}>
              <Col span={8}>
                <div class="stat-item">
                  <span class="stat-label">待处理</span>
                  <span class="stat-value">{stats.value.todo_missions}</span>
                </div>
              </Col>
              <Col span={8}>
                <div class="stat-item">
                  <span class="stat-label">进行中</span>
                  <span class="stat-value">
                    {stats.value.in_progress_missions}
                  </span>
                </div>
              </Col>
              <Col span={8}>
                <div class="stat-item">
                  <span class="stat-label">已完成</span>
                  <span class="stat-value">
                    {stats.value.completed_missions}
                  </span>
                </div>
              </Col>
            </Row>
          </Card>

          <Row gutter={[16, 16]}>
            {/* 按类型统计 */}
            <Col xs={24} lg={12}>
              <Card
                title={
                  <Space>
                    <BugOutlined />
                    任务类型分布
                  </Space>
                }
              >
                <div class="type-stats">
                  {Object.entries(stats.value.by_type).map(([type, count]) => (
                    <div key={type} class="type-item">
                      <div class="type-header">
                        <span class="type-name">
                          {type === "feature" && "功能"}
                          {type === "enhancement" && "优化"}
                          {type === "bug" && "缺陷"}
                        </span>
                        <span class="type-count">{count}</span>
                      </div>
                      <Progress
                        percent={(count / stats.value.total_missions) * 100}
                        showInfo={false}
                        strokeColor={
                          type === "feature"
                            ? "#1890ff"
                            : type === "enhancement"
                              ? "#52c41a"
                              : "#ff4d4f"
                        }
                      />
                    </div>
                  ))}
                </div>
              </Card>
            </Col>

            {/* 按优先级统计 */}
            <Col xs={24} lg={12}>
              <Card
                title={
                  <Space>
                    <ThunderboltOutlined />
                    优先级分布
                  </Space>
                }
              >
                <div class="priority-stats">
                  {Object.entries(stats.value.by_priority).map(
                    ([priority, count]) => (
                      <div key={priority} class="priority-item">
                        <div class="priority-header">
                          <Tag
                            color={
                              priority === "P0"
                                ? "red"
                                : priority === "P1"
                                  ? "orange"
                                  : priority === "P2"
                                    ? "gold"
                                    : "default"
                            }
                          >
                            {priority}
                          </Tag>
                          <span class="priority-count">{count}</span>
                        </div>
                        <Progress
                          percent={(count / stats.value.total_missions) * 100}
                          showInfo={false}
                          strokeColor={
                            priority === "P0"
                              ? "#ff4d4f"
                              : priority === "P1"
                                ? "#ff7a45"
                                : priority === "P2"
                                  ? "#ffa940"
                                  : "#8c8c8c"
                          }
                        />
                      </div>
                    ),
                  )}
                </div>
              </Card>
            </Col>
          </Row>

          {/* 成员任务分布 */}
          <Card
            title={
              <Space>
                <UserOutlined />
                成员任务分布
              </Space>
            }
            style={{ marginTop: "16px" }}
          >
            <Table
              dataSource={stats.value.by_assignee}
              columns={columns}
              rowKey="user_id"
              pagination={false}
            />
          </Card>
        </Spin>
      </div>
    );
  },
});
