import { defineComponent, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { Card, Row, Col, Statistic, Empty, message } from "ant-design-vue";
import {
  PictureOutlined,
  FolderOutlined,
  BoxPlotOutlined,
  FileImageOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  RocketOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons-vue";
import {
  getOverview,
  getRecentActivities,
  getSystemStatus,
  type Activity,
} from "@/api/statistics";
import "./Home.less";

export default defineComponent({
  name: "Home",
  setup() {
    const router = useRouter();
    const loading = ref(false);

    // 统计数据
    const stats = ref({
      textures: { total: 0, trend: 0, recent_count: 0 },
      projects: { total: 0, trend: 0, recent_count: 0 },
      models: { total: 0, trend: 0, recent_count: 0 },
      assets: { total: 0, trend: 0, recent_count: 0 },
      ai3d: { total: 0, trend: 0, recent_count: 0 },
    });

    // 最近活动
    const activities = ref<Activity[]>([]);

    // 系统状态
    const systemStatus = ref({
      service: { status: "unknown", uptime: 0 },
      database: { status: "unknown", size: 0 },
      storage: { total: 0, used: 0, usage_percent: 0 },
      sync: { last_sync_at: "", status: "unknown" },
    });

    // 加载统计数据
    const loadStats = async () => {
      loading.value = true;
      try {
        const res = await getOverview();
        stats.value = res.data;
      } catch (error) {
        console.error("加载统计数据失败:", error);
        message.error("加载统计数据失败");
      } finally {
        loading.value = false;
      }
    };

    // 加载最近活动
    const loadActivities = async () => {
      try {
        const res = await getRecentActivities(10);
        activities.value = res.data.activities || [];
      } catch (error) {
        console.error("加载活动记录失败:", error);
      }
    };

    // 加载系统状态
    const loadSystemStatus = async () => {
      try {
        const res = await getSystemStatus();
        systemStatus.value = res.data;
      } catch (error) {
        console.error("加载系统状态失败:", error);
      }
    };

    // 格式化时间
    const formatTime = (dateStr: string) => {
      if (!dateStr) return "未知";
      const date = new Date(dateStr);
      const now = new Date();
      const diff = Math.floor((now.getTime() - date.getTime()) / 1000);

      if (diff < 60) return `${diff} 秒前`;
      if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
      if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
      return `${Math.floor(diff / 86400)} 天前`;
    };

    // 格式化存储大小
    const formatSize = (bytes: number) => {
      if (bytes === 0) return "0 B";
      const k = 1024;
      const sizes = ["B", "KB", "MB", "GB", "TB"];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
    };

    // 格式化运行时间
    const formatUptime = (seconds: number) => {
      const days = Math.floor(seconds / 86400);
      const hours = Math.floor((seconds % 86400) / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);

      if (days > 0) return `${days} 天 ${hours} 小时`;
      if (hours > 0) return `${hours} 小时 ${minutes} 分钟`;
      return `${minutes} 分钟`;
    };

    onMounted(() => {
      loadStats();
      loadActivities();
      loadSystemStatus();
    });

    return () => {
      // 快速访问模块（动态获取数量）
      const quickAccess = [
        {
          title: "贴图库",
          icon: PictureOutlined,
          description: "管理和浏览材质贴图",
          path: "/textures",
          color: "#1890ff",
          count: stats.value.textures.total,
        },
        {
          title: "项目管理",
          icon: FolderOutlined,
          description: "前端项目版本管理",
          path: "/projects",
          color: "#52c41a",
          count: stats.value.projects.total,
        },
        {
          title: "模型库",
          icon: BoxPlotOutlined,
          description: "3D 模型资源管理",
          path: "/models",
          color: "#722ed1",
          count: stats.value.models.total,
        },
        {
          title: "资产库",
          icon: FileImageOutlined,
          description: "通用资产文件管理",
          path: "/assets",
          color: "#fa8c16",
          count: stats.value.assets.total,
        },
        {
          title: "文件库",
          icon: FolderOutlined,
          description: "公司文档资料管理",
          path: "/documents",
          color: "#13c2c2",
          count: 0,
        },
        {
          title: "AI 3D",
          icon: RocketOutlined,
          description: "AI 生成 3D 模型",
          path: "/ai3d",
          color: "#eb2f96",
          count: stats.value.ai3d.total,
        },
      ];

      return (
        <div class="home-page">
          {/* 欢迎横幅 */}
          <div class="welcome-banner">
            <div class="banner-content">
              <RocketOutlined class="banner-icon" />
              <div class="banner-text">
                <h1>欢迎使用资源管理系统</h1>
                <p>统一管理贴图、模型、项目和资产资源</p>
              </div>
            </div>
          </div>

          {/* 统计卡片 */}
          <Row gutter={[16, 16]} style={{ marginBottom: "24px" }}>
            <Col xs={24} sm={12} lg={6}>
              <Card bordered={false} class="stat-card stat-card-blue">
                <Statistic
                  title="贴图总数"
                  value={stats.value.textures.total}
                  prefix={<PictureOutlined />}
                  valueStyle={{ color: "#1890ff" }}
                />
                <div class="stat-trend">
                  {stats.value.textures.trend >= 0 ? (
                    <ArrowUpOutlined style={{ color: "#52c41a" }} />
                  ) : (
                    <ArrowDownOutlined style={{ color: "#ff4d4f" }} />
                  )}
                  <span
                    style={{
                      color:
                        stats.value.textures.trend >= 0 ? "#52c41a" : "#ff4d4f",
                      marginLeft: "4px",
                    }}
                  >
                    {Math.abs(stats.value.textures.trend).toFixed(1)}% 本月
                  </span>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card bordered={false} class="stat-card stat-card-green">
                <Statistic
                  title="项目数量"
                  value={stats.value.projects.total}
                  prefix={<FolderOutlined />}
                  valueStyle={{ color: "#52c41a" }}
                />
                <div class="stat-trend">
                  {stats.value.projects.trend >= 0 ? (
                    <ArrowUpOutlined style={{ color: "#52c41a" }} />
                  ) : (
                    <ArrowDownOutlined style={{ color: "#ff4d4f" }} />
                  )}
                  <span
                    style={{
                      color:
                        stats.value.projects.trend >= 0 ? "#52c41a" : "#ff4d4f",
                      marginLeft: "4px",
                    }}
                  >
                    {Math.abs(stats.value.projects.trend).toFixed(1)}% 本月
                  </span>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card bordered={false} class="stat-card stat-card-purple">
                <Statistic
                  title="模型总数"
                  value={stats.value.models.total}
                  prefix={<BoxPlotOutlined />}
                  valueStyle={{ color: "#722ed1" }}
                />
                <div class="stat-trend">
                  {stats.value.models.trend >= 0 ? (
                    <ArrowUpOutlined style={{ color: "#52c41a" }} />
                  ) : (
                    <ArrowDownOutlined style={{ color: "#ff4d4f" }} />
                  )}
                  <span
                    style={{
                      color:
                        stats.value.models.trend >= 0 ? "#52c41a" : "#ff4d4f",
                      marginLeft: "4px",
                    }}
                  >
                    {Math.abs(stats.value.models.trend).toFixed(1)}% 本月
                  </span>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card bordered={false} class="stat-card stat-card-orange">
                <Statistic
                  title="资产总数"
                  value={stats.value.assets.total}
                  prefix={<FileImageOutlined />}
                  valueStyle={{ color: "#fa8c16" }}
                />
                <div class="stat-trend">
                  {stats.value.assets.trend >= 0 ? (
                    <ArrowUpOutlined style={{ color: "#52c41a" }} />
                  ) : (
                    <ArrowDownOutlined style={{ color: "#ff4d4f" }} />
                  )}
                  <span
                    style={{
                      color:
                        stats.value.assets.trend >= 0 ? "#52c41a" : "#ff4d4f",
                      marginLeft: "4px",
                    }}
                  >
                    {Math.abs(stats.value.assets.trend).toFixed(1)}% 本月
                  </span>
                </div>
              </Card>
            </Col>
          </Row>

          {/* 快速访问 */}
          <Card
            title={
              <span>
                <ThunderboltOutlined style={{ marginRight: "8px" }} />
                快速访问
              </span>
            }
            bordered={false}
            style={{ marginBottom: "24px" }}
          >
            <Row gutter={[16, 16]}>
              {quickAccess.map((item) => (
                <Col key={item.path} xs={24} sm={12} lg={6}>
                  <div
                    class="quick-access-card"
                    onClick={() => router.push(item.path)}
                  >
                    <div
                      class="quick-access-icon"
                      style={{ backgroundColor: item.color }}
                    >
                      <item.icon />
                    </div>
                    <div class="quick-access-content">
                      <h3>{item.title}</h3>
                      <p>{item.description}</p>
                      <div class="quick-access-count">{item.count} 项</div>
                    </div>
                  </div>
                </Col>
              ))}
            </Row>
          </Card>

          {/* 最近活动 */}
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Card title="最近活动" bordered={false}>
                {activities.value.length === 0 ? (
                  <Empty description="暂无最近活动记录" />
                ) : (
                  <div class="activity-list">
                    {activities.value.map((activity) => (
                      <div key={activity.id} class="activity-item">
                        <div class="activity-info">
                          <span class="activity-name">{activity.name}</span>
                          <span class="activity-action">
                            {activity.action === "upload" && "上传"}
                            {activity.action === "update" && "更新"}
                            {activity.action === "delete" && "删除"}
                            {activity.action === "version_upload" &&
                              `上传版本 ${activity.version}`}
                          </span>
                        </div>
                        <div class="activity-meta">
                          <span class="activity-user">{activity.user}</span>
                          <span class="activity-time">
                            {formatTime(activity.created_at)}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="系统状态" bordered={false}>
                <div class="system-status">
                  <div class="status-item">
                    <span class="status-label">服务状态</span>
                    <span
                      class={[
                        "status-value",
                        systemStatus.value.service.status === "running"
                          ? "status-online"
                          : "",
                      ]}
                    >
                      {systemStatus.value.service.status === "running"
                        ? "运行中"
                        : "未知"}
                    </span>
                  </div>
                  <div class="status-item">
                    <span class="status-label">运行时间</span>
                    <span class="status-value">
                      {formatUptime(systemStatus.value.service.uptime)}
                    </span>
                  </div>
                  <div class="status-item">
                    <span class="status-label">数据库</span>
                    <span
                      class={[
                        "status-value",
                        systemStatus.value.database.status === "healthy"
                          ? "status-online"
                          : "",
                      ]}
                    >
                      {systemStatus.value.database.status === "healthy"
                        ? "正常"
                        : "未知"}
                    </span>
                  </div>
                  <div class="status-item">
                    <span class="status-label">数据库大小</span>
                    <span class="status-value">
                      {formatSize(systemStatus.value.database.size)}
                    </span>
                  </div>
                  <div class="status-item">
                    <span class="status-label">存储空间</span>
                    <span class="status-value">
                      {systemStatus.value.storage.usage_percent.toFixed(1)}%
                      已使用
                    </span>
                  </div>
                  <div class="status-item">
                    <span class="status-label">最后同步</span>
                    <span class="status-value">
                      {formatTime(systemStatus.value.sync.last_sync_at)}
                    </span>
                  </div>
                </div>
              </Card>
            </Col>
          </Row>
        </div>
      );
    };
  },
});
