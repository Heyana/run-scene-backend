import { defineComponent, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { Card, Row, Col, Statistic, Button, Empty } from "ant-design-vue";
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
import "./Home.less";

export default defineComponent({
  name: "Home",
  setup() {
    const router = useRouter();
    const loading = ref(false);

    // 统计数据
    const stats = ref({
      textures: 0,
      projects: 0,
      models: 0,
      assets: 0,
    });

    // 加载统计数据
    const loadStats = async () => {
      loading.value = true;
      try {
        // TODO: 调用实际的统计 API
        // 这里使用模拟数据
        stats.value = {
          textures: 1234,
          projects: 56,
          models: 789,
          assets: 432,
        };
      } catch (error) {
        console.error("加载统计数据失败:", error);
      } finally {
        loading.value = false;
      }
    };

    onMounted(() => {
      loadStats();
    });

    // 快速访问模块
    const quickAccess = [
      {
        title: "贴图库",
        icon: PictureOutlined,
        description: "管理和浏览材质贴图",
        path: "/textures",
        color: "#1890ff",
        count: stats.value.textures,
      },
      {
        title: "项目管理",
        icon: FolderOutlined,
        description: "前端项目版本管理",
        path: "/projects",
        color: "#52c41a",
        count: stats.value.projects,
      },
      {
        title: "模型库",
        icon: BoxPlotOutlined,
        description: "3D 模型资源管理",
        path: "/models",
        color: "#722ed1",
        count: stats.value.models,
      },
      {
        title: "资产库",
        icon: FileImageOutlined,
        description: "通用资产文件管理",
        path: "/assets",
        color: "#fa8c16",
        count: stats.value.assets,
      },
    ];

    return () => (
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
                value={stats.value.textures}
                prefix={<PictureOutlined />}
                valueStyle={{ color: "#1890ff" }}
              />
              <div class="stat-trend">
                <ArrowUpOutlined style={{ color: "#52c41a" }} />
                <span style={{ color: "#52c41a", marginLeft: "4px" }}>
                  12% 本月
                </span>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} class="stat-card stat-card-green">
              <Statistic
                title="项目数量"
                value={stats.value.projects}
                prefix={<FolderOutlined />}
                valueStyle={{ color: "#52c41a" }}
              />
              <div class="stat-trend">
                <ArrowUpOutlined style={{ color: "#52c41a" }} />
                <span style={{ color: "#52c41a", marginLeft: "4px" }}>
                  8% 本月
                </span>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} class="stat-card stat-card-purple">
              <Statistic
                title="模型总数"
                value={stats.value.models}
                prefix={<BoxPlotOutlined />}
                valueStyle={{ color: "#722ed1" }}
              />
              <div class="stat-trend">
                <ArrowUpOutlined style={{ color: "#52c41a" }} />
                <span style={{ color: "#52c41a", marginLeft: "4px" }}>
                  15% 本月
                </span>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} class="stat-card stat-card-orange">
              <Statistic
                title="资产总数"
                value={stats.value.assets}
                prefix={<FileImageOutlined />}
                valueStyle={{ color: "#fa8c16" }}
              />
              <div class="stat-trend">
                <ArrowDownOutlined style={{ color: "#ff4d4f" }} />
                <span style={{ color: "#ff4d4f", marginLeft: "4px" }}>
                  3% 本月
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
            <Card title="最近上传" bordered={false}>
              <Empty description="暂无最近上传记录" />
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="系统状态" bordered={false}>
              <div class="system-status">
                <div class="status-item">
                  <span class="status-label">服务状态</span>
                  <span class="status-value status-online">运行中</span>
                </div>
                <div class="status-item">
                  <span class="status-label">数据库</span>
                  <span class="status-value status-online">正常</span>
                </div>
                <div class="status-item">
                  <span class="status-label">存储空间</span>
                  <span class="status-value">75% 已使用</span>
                </div>
                <div class="status-item">
                  <span class="status-label">最后同步</span>
                  <span class="status-value">2 分钟前</span>
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </div>
    );
  },
});
