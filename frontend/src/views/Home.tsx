import { defineComponent } from "vue";
import { RouterLink } from "vue-router";
import { useAppStore } from "@/stores/app";
import { Card, Button, Space, Typography } from "ant-design-vue";
import {
  HomeOutlined,
  PictureOutlined,
  InfoCircleOutlined,
} from "@ant-design/icons-vue";

const { Title, Paragraph } = Typography;

export default defineComponent({
  name: "Home",
  setup() {
    const appStore = useAppStore();

    return () => (
      <div class="home-container">
        <Card bordered={false}>
          <Space direction="vertical" size="large" style={{ width: "100%" }}>
            <div style={{ textAlign: "center" }}>
              <HomeOutlined style={{ fontSize: "64px", color: "#1890ff" }} />
              <Title level={1}>材质库管理系统</Title>
              <Paragraph>
                基于 Vue 3 + TypeScript + TSX + Ant Design Vue
                构建的现代化材质管理平台
              </Paragraph>
            </div>

            <Card title="快速导航" size="small">
              <Space size="middle">
                <RouterLink to="/textures">
                  <Button type="primary" size="large">
                    {{
                      icon: () => <PictureOutlined />,
                      default: () => "浏览材质库",
                    }}
                  </Button>
                </RouterLink>
                <RouterLink to="/about">
                  <Button size="large">
                    {{
                      icon: () => <InfoCircleOutlined />,
                      default: () => "关于系统",
                    }}
                  </Button>
                </RouterLink>
              </Space>
            </Card>

            <Card title="系统信息" size="small">
              <Paragraph>
                <strong>当前主题:</strong> {appStore.theme}
              </Paragraph>
              <Button onClick={appStore.toggleTheme}>切换主题</Button>
            </Card>
          </Space>
        </Card>
      </div>
    );
  },
});
