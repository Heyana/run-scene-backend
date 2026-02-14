import { defineComponent } from "vue";
import { useRouter } from "vue-router";
import { Card, Row, Col, Empty } from "ant-design-vue";
import {
  UserOutlined,
  SafetyOutlined,
  KeyOutlined,
} from "@ant-design/icons-vue";

export default defineComponent({
  name: "UserManagementWelcome",
  setup() {
    const router = useRouter();

    const modules = [
      {
        title: "用户管理",
        icon: UserOutlined,
        description: "管理系统用户账号，包括创建、编辑、禁用等操作",
        path: "/user-management/users",
        color: "#1890ff",
      },
      {
        title: "角色管理",
        icon: SafetyOutlined,
        description: "配置系统角色和角色权限，支持自定义角色",
        path: "/user-management/roles",
        color: "#52c41a",
      },
      {
        title: "权限管理",
        icon: KeyOutlined,
        description: "管理系统权限点，支持自定义权限配置",
        path: "/user-management/permissions",
        color: "#722ed1",
      },
      {
        title: "权限组管理",
        icon: KeyOutlined,
        description: "管理权限组，将多个权限组合成权限组",
        path: "/user-management/permission-groups",
        color: "#fa8c16",
      },
    ];

    return () => (
      <div>
        <Card bordered={false}>
          <Empty
            description="请从左侧菜单选择功能模块"
            style={{ marginBottom: "32px" }}
          />

          <Row gutter={[16, 16]}>
            {modules.map((module) => (
              <Col key={module.path} xs={24} sm={12} lg={8}>
                <div
                  style={{
                    cursor: "pointer",
                  }}
                  onClick={() => router.push(module.path)}
                >
                  <Card
                    hoverable
                    bordered={false}
                    style={{
                      borderLeft: `4px solid ${module.color}`,
                    }}
                  >
                    <div style={{ display: "flex", alignItems: "flex-start" }}>
                      <div
                        style={{
                          fontSize: "32px",
                          color: module.color,
                          marginRight: "16px",
                        }}
                      >
                        <module.icon />
                      </div>
                      <div style={{ flex: 1 }}>
                        <h3 style={{ margin: "0 0 8px 0", fontSize: "16px" }}>
                          {module.title}
                        </h3>
                        <p
                          style={{ margin: 0, color: "#666", fontSize: "14px" }}
                        >
                          {module.description}
                        </p>
                      </div>
                    </div>
                  </Card>
                </div>
              </Col>
            ))}
          </Row>
        </Card>
      </div>
    );
  },
});
