import { defineComponent } from "vue";
import { RouterView, useRoute, useRouter } from "vue-router";
import { Card, Menu } from "ant-design-vue";
import {
  UserOutlined,
  SafetyOutlined,
  KeyOutlined,
} from "@ant-design/icons-vue";
import type { MenuProps } from "ant-design-vue";

export default defineComponent({
  name: "UserManagement",
  setup() {
    const route = useRoute();
    const router = useRouter();

    // 菜单项
    const menuItems: MenuProps["items"] = [
      {
        key: "/user-management/users",
        icon: () => <UserOutlined />,
        label: "用户管理",
      },
      {
        key: "/user-management/roles",
        icon: () => <SafetyOutlined />,
        label: "角色管理",
      },
      {
        key: "/user-management/permissions",
        icon: () => <KeyOutlined />,
        label: "权限管理",
      },
      {
        key: "/user-management/permission-groups",
        icon: () => <KeyOutlined />,
        label: "权限组管理",
      },
    ];

    // 菜单点击
    const handleMenuClick: MenuProps["onClick"] = (info) => {
      router.push(info.key as string);
    };

    return () => (
      <div style={{ padding: "24px" }}>
        <Card
          title="人员管理"
          bordered={false}
          style={{ marginBottom: "16px" }}
        >
          <p style={{ margin: 0, color: "#666" }}>
            管理系统用户、角色和权限配置
          </p>
        </Card>

        <div style={{ display: "flex", gap: "16px" }}>
          {/* 左侧菜单 */}
          <Card
            bordered={false}
            bodyStyle={{ padding: 0 }}
            style={{ width: "200px", flexShrink: 0 }}
          >
            <Menu
              mode="inline"
              selectedKeys={[route.path]}
              items={menuItems}
              onClick={handleMenuClick}
              style={{ border: "none" }}
            />
          </Card>

          {/* 右侧内容 */}
          <div style={{ flex: 1, minWidth: 0 }}>
            <RouterView />
          </div>
        </div>
      </div>
    );
  },
});
