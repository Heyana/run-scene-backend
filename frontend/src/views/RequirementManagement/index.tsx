import { defineComponent, ref, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { Layout, Menu } from "ant-design-vue";
import {
  TeamOutlined,
  ProjectOutlined,
  CheckSquareOutlined,
  BarChartOutlined,
} from "@ant-design/icons-vue";
import "./index.less";

const { Sider, Content } = Layout;

export default defineComponent({
  name: "RequirementManagement",
  setup() {
    const router = useRouter();
    const route = useRoute();
    const collapsed = ref(false);
    const selectedKeys = ref<string[]>([]);

    // 菜单项
    const menuItems = [
      {
        key: "companies",
        icon: TeamOutlined,
        label: "公司管理",
        path: "/requirement-management/companies",
      },
      {
        key: "projects",
        icon: ProjectOutlined,
        label: "项目管理",
        path: "/requirement-management/projects",
      },
      {
        key: "missions",
        icon: CheckSquareOutlined,
        label: "任务管理",
        path: "/requirement-management/missions",
      },
      {
        key: "statistics",
        icon: BarChartOutlined,
        label: "统计报表",
        path: "/requirement-management/statistics",
      },
    ];

    // 处理菜单点击
    const handleMenuClick = ({ key }: { key: string | number }) => {
      const item = menuItems.find((m) => m.key === String(key));
      if (item) {
        router.push(item.path);
      }
    };

    // 更新选中的菜单项
    const updateSelectedKeys = () => {
      const path = route.path;
      const item = menuItems.find((m) => path.startsWith(m.path));
      if (item) {
        selectedKeys.value = [item.key];
      }
    };

    onMounted(() => {
      updateSelectedKeys();
      // 如果在根路径，重定向到公司列表
      if (route.path === "/requirement-management") {
        router.replace("/requirement-management/companies");
      }
    });

    return () => (
      <Layout class="requirement-management-layout">
        <Sider
          collapsible
          v-model:collapsed={collapsed.value}
          theme="light"
          width={200}
        >
          <div class="logo">
            <CheckSquareOutlined
              style={{ fontSize: "24px", color: "#1890ff" }}
            />
            {!collapsed.value && <span>需求管理</span>}
          </div>
          <Menu
            mode="inline"
            selectedKeys={selectedKeys.value}
            onClick={handleMenuClick}
          >
            {menuItems.map((item) => (
              <Menu.Item key={item.key}>
                <item.icon />
                <span>{item.label}</span>
              </Menu.Item>
            ))}
          </Menu>
        </Sider>
        <Layout>
          <Content class="requirement-management-content">
            <router-view />
          </Content>
        </Layout>
      </Layout>
    );
  },
});
