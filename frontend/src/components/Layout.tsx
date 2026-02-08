import { defineComponent } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { useAppStore } from "@/stores/app";
import { Layout, Menu, Button, theme } from "ant-design-vue";
import {
  HomeOutlined,
  PictureOutlined,
  InfoCircleOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BulbOutlined,
  FolderOutlined,
  BoxPlotOutlined,
  FileImageOutlined,
  RocketOutlined,
} from "@ant-design/icons-vue";

const { Header, Sider, Content } = Layout;

export default defineComponent({
  name: "AppLayout",
  setup(_, { slots }) {
    const appStore = useAppStore();
    const route = useRoute();
    const { token } = theme.useToken();

    const menuItems = [
      {
        key: "/",
        icon: () => <HomeOutlined />,
        label: <RouterLink to="/">首页</RouterLink>,
      },
      {
        key: "/textures",
        icon: () => <PictureOutlined />,
        label: <RouterLink to="/textures">材质库</RouterLink>,
      },
      {
        key: "/texture-analysis",
        icon: () => <PictureOutlined />,
        label: <RouterLink to="/texture-analysis">贴图类型分析</RouterLink>,
      },
      {
        key: "/projects",
        icon: () => <FolderOutlined />,
        label: <RouterLink to="/projects">项目管理</RouterLink>,
      },
      {
        key: "/models",
        icon: () => <BoxPlotOutlined />,
        label: <RouterLink to="/models">模型库</RouterLink>,
      },
      {
        key: "/assets",
        icon: () => <FileImageOutlined />,
        label: <RouterLink to="/assets">资产库</RouterLink>,
      },
      {
        key: "/ai3d",
        icon: () => <RocketOutlined />,
        label: <RouterLink to="/ai3d">AI 3D</RouterLink>,
      },
      {
        key: "/about",
        icon: () => <InfoCircleOutlined />,
        label: <RouterLink to="/about">关于</RouterLink>,
      },
    ];

    return () => (
      <Layout class="main" style={{ minHeight: "100vh" }}>
        <Sider
          trigger={null}
          collapsible
          collapsed={appStore.sidebarCollapsed}
          style={{
            overflow: "auto",
            height: "100vh",
            position: "fixed",
            left: 0,
            top: 0,
            bottom: 0,
          }}
        >
          <div
            style={{
              height: "64px",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: "#fff",
              fontSize: "18px",
              fontWeight: "bold",
            }}
          >
            {!appStore.sidebarCollapsed && "材质库"}
          </div>
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[route.path]}
            items={menuItems}
          />
        </Sider>

        <Layout style={{}}>
          <Header
            style={{
              padding: "0 16px",
              background: token.value.colorBgContainer,
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <Button
              type="text"
              onClick={appStore.toggleSidebar}
              style={{
                fontSize: "16px",
                width: 64,
                height: 64,
              }}
            >
              {{
                icon: () =>
                  appStore.sidebarCollapsed ? (
                    <MenuUnfoldOutlined />
                  ) : (
                    <MenuFoldOutlined />
                  ),
              }}
            </Button>
            <Button type="text" onClick={appStore.toggleTheme}>
              {{
                icon: () => <BulbOutlined />,
                default: () =>
                  `${appStore.theme === "light" ? "深色" : "浅色"}模式`,
              }}
            </Button>
          </Header>
          <Content
            style={{
              margin: "24px 16px",
              padding: 24,
              minHeight: 280,
              background: token.value.colorBgContainer,
              borderRadius: token.value.borderRadiusLG,
            }}
          >
            {slots.default?.()}
          </Content>
        </Layout>
      </Layout>
    );
  },
});
