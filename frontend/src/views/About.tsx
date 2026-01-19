import { defineComponent } from "vue";
import { RouterLink } from "vue-router";
import { Card, Typography, List, Button, Space, Divider } from "ant-design-vue";
import { ArrowLeftOutlined, CheckCircleOutlined } from "@ant-design/icons-vue";

const { Title, Paragraph } = Typography;

export default defineComponent({
  name: "About",
  setup() {
    const features = [
      "Vue 3 Composition API",
      "TypeScript 类型安全",
      "TSX 语法支持",
      "Vue Router 路由管理",
      "Pinia 状态管理",
      "Ant Design Vue 组件库",
      "Axios HTTP 请求",
      "Vite 快速构建",
    ];

    return () => (
      <div class="about-container">
        <Card bordered={false}>
          <Space direction="vertical" size="large" style={{ width: "100%" }}>
            <div>
              <Title level={1}>关于材质库管理系统</Title>
              <Paragraph>
                这是一个基于现代前端技术栈构建的材质管理系统，提供高效的材质浏览、搜索和管理功能。
              </Paragraph>
            </div>

            <Divider />

            <div>
              <Title level={2}>技术特性</Title>
              <List dataSource={features}>
                {{
                  renderItem: ({ item }: { item: string }) => (
                    <List.Item>
                      <CheckCircleOutlined
                        style={{ color: "#52c41a", marginRight: "8px" }}
                      />
                      {item}
                    </List.Item>
                  ),
                }}
              </List>
            </div>

            <Divider />

            <div>
              <Title level={2}>系统功能</Title>
              <List>
                <List.Item>
                  <List.Item.Meta
                    title="材质管理"
                    description="支持材质的浏览、搜索、分类和标签管理"
                  />
                </List.Item>
                <List.Item>
                  <List.Item.Meta
                    title="同步功能"
                    description="自动同步远程材质库，保持数据最新"
                  />
                </List.Item>
                <List.Item>
                  <List.Item.Meta
                    title="统计分析"
                    description="提供使用统计和数据分析功能"
                  />
                </List.Item>
                <List.Item>
                  <List.Item.Meta
                    title="安全管理"
                    description="IP 黑白名单、访问控制等安全功能"
                  />
                </List.Item>
              </List>
            </div>

            <Divider />

            <div>
              <RouterLink to="/">
                <Button type="primary">
                  {{
                    icon: () => <ArrowLeftOutlined />,
                    default: () => "返回首页",
                  }}
                </Button>
              </RouterLink>
            </div>
          </Space>
        </Card>
      </div>
    );
  },
});
