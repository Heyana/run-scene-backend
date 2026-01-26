import { defineComponent, ref, onMounted } from "vue";
import {
  Card,
  Table,
  Tag,
  Space,
  Typography,
  Spin,
  message,
  Descriptions,
  Collapse,
  Statistic,
  Row,
  Col,
} from "ant-design-vue";
import {
  FileImageOutlined,
  CheckCircleOutlined,
  QuestionCircleOutlined,
} from "@ant-design/icons-vue";
import { apiManager } from "@/api/http";

const { Title, Text } = Typography;
const { Panel } = Collapse;

interface TextureTypeAnalysis {
  original_type: string;
  count: number;
  source: string;
  examples: string[];
  suggested_type: string;
}

export default defineComponent({
  name: "TextureTypeAnalysis",
  setup() {
    const loading = ref(false);
    const analysisData = ref<TextureTypeAnalysis[]>([]);
    const polyhavenData = ref<TextureTypeAnalysis[]>([]);
    const ambientcgData = ref<TextureTypeAnalysis[]>([]);

    // 加载分析数据
    const loadAnalysis = async () => {
      loading.value = true;
      try {
        const response = await apiManager.api.texture.analyzeTextureTypes();
        analysisData.value = response.data.analysis || [];

        // 分离两个数据源
        polyhavenData.value = analysisData.value.filter(
          (item) => item.source === "polyhaven",
        );
        ambientcgData.value = analysisData.value.filter(
          (item) => item.source === "ambientcg",
        );

        message.success("分析完成");
      } catch (error) {
        console.error("加载分析数据失败:", error);
        message.error("加载失败");
      } finally {
        loading.value = false;
      }
    };

    onMounted(() => {
      loadAnalysis();
    });

    // 表格列定义
    const columns = [
      {
        title: "原始类型",
        dataIndex: "original_type",
        key: "original_type",
        width: 200,
        customRender: ({ text }: any) => (
          <Tag color="blue" style={{ fontSize: "14px" }}>
            {text}
          </Tag>
        ),
      },
      {
        title: "数量",
        dataIndex: "count",
        key: "count",
        width: 100,
        sorter: (a: TextureTypeAnalysis, b: TextureTypeAnalysis) =>
          a.count - b.count,
      },
      {
        title: "建议的 Three.js 类型",
        dataIndex: "suggested_type",
        key: "suggested_type",
        width: 250,
        customRender: ({ text }: any) => {
          const isUnknown = text === "unknown";
          const types = text.split(",");
          const isPacked = types.length > 1; // 是否是组合贴图

          return (
            <Space direction="vertical" size="small">
              {isPacked && (
                <Tag color="purple" icon={<FileImageOutlined />}>
                  组合贴图 (Packed)
                </Tag>
              )}
              <Space>
                {types.map((type: string, index: number) => (
                  <Tag
                    key={index}
                    color={isUnknown ? "red" : "green"}
                    icon={
                      isUnknown ? (
                        <QuestionCircleOutlined />
                      ) : (
                        <CheckCircleOutlined />
                      )
                    }
                  >
                    {type}
                  </Tag>
                ))}
              </Space>
            </Space>
          );
        },
      },
      {
        title: "示例文件",
        dataIndex: "examples",
        key: "examples",
        customRender: ({ text }: any) => (
          <Collapse ghost>
            <Panel header={`查看 ${text.length} 个示例`} key="1">
              <ul style={{ margin: 0, paddingLeft: "20px" }}>
                {text.map((example: string, index: number) => (
                  <li key={index}>
                    <Text code>{example}</Text>
                  </li>
                ))}
              </ul>
            </Panel>
          </Collapse>
        ),
      },
    ];

    // 统计信息
    const getStats = (data: TextureTypeAnalysis[]) => {
      const totalTypes = data.length;
      const totalFiles = data.reduce((sum, item) => sum + item.count, 0);
      const unknownTypes = data.filter(
        (item) => item.suggested_type === "unknown",
      ).length;
      const mappedTypes = totalTypes - unknownTypes;

      return { totalTypes, totalFiles, unknownTypes, mappedTypes };
    };

    return () => (
      <div style={{ padding: "24px" }}>
        <Card bordered={false}>
          <Space direction="vertical" size="large" style={{ width: "100%" }}>
            <div style={{ textAlign: "center" }}>
              <FileImageOutlined
                style={{ fontSize: "48px", color: "#1890ff" }}
              />
              <Title level={2}>贴图类型分析</Title>
              <Text type="secondary">
                分析两个网站的所有贴图类型，对比原始类型与 Three.js
                类型的映射关系
              </Text>
            </div>

            <Spin spinning={loading.value}>
              {/* PolyHaven 数据源 */}
              <Card
                title={
                  <Space>
                    <FileImageOutlined />
                    <span>网站1 - PolyHaven</span>
                  </Space>
                }
                style={{ marginBottom: "24px" }}
              >
                <Row gutter={16} style={{ marginBottom: "24px" }}>
                  <Col span={6}>
                    <Statistic
                      title="贴图类型总数"
                      value={getStats(polyhavenData.value).totalTypes}
                      prefix={<FileImageOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="文件总数"
                      value={getStats(polyhavenData.value).totalFiles}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="已映射类型"
                      value={getStats(polyhavenData.value).mappedTypes}
                      valueStyle={{ color: "#3f8600" }}
                      prefix={<CheckCircleOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="未映射类型"
                      value={getStats(polyhavenData.value).unknownTypes}
                      valueStyle={{ color: "#cf1322" }}
                      prefix={<QuestionCircleOutlined />}
                    />
                  </Col>
                </Row>

                <Table
                  columns={columns}
                  dataSource={polyhavenData.value}
                  rowKey={(record) => `polyhaven-${record.original_type}`}
                  pagination={{ pageSize: 10 }}
                  size="middle"
                />
              </Card>

              {/* AmbientCG 数据源 */}
              <Card
                title={
                  <Space>
                    <FileImageOutlined />
                    <span>网站2 - AmbientCG</span>
                  </Space>
                }
              >
                <Row gutter={16} style={{ marginBottom: "24px" }}>
                  <Col span={6}>
                    <Statistic
                      title="贴图类型总数"
                      value={getStats(ambientcgData.value).totalTypes}
                      prefix={<FileImageOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="文件总数"
                      value={getStats(ambientcgData.value).totalFiles}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="已映射类型"
                      value={getStats(ambientcgData.value).mappedTypes}
                      valueStyle={{ color: "#3f8600" }}
                      prefix={<CheckCircleOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="未映射类型"
                      value={getStats(ambientcgData.value).unknownTypes}
                      valueStyle={{ color: "#cf1322" }}
                      prefix={<QuestionCircleOutlined />}
                    />
                  </Col>
                </Row>

                <Table
                  columns={columns}
                  dataSource={ambientcgData.value}
                  rowKey={(record) => `ambientcg-${record.original_type}`}
                  pagination={{ pageSize: 10 }}
                  size="middle"
                />
              </Card>

              {/* 说明信息 */}
              <Card title="说明" size="small">
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="原始类型">
                    从文件名中提取的贴图类型，如 Diffuse_2k.jpg → Diffuse
                  </Descriptions.Item>
                  <Descriptions.Item label="建议的 Three.js 类型">
                    根据原始类型名称自动推荐的 Three.js 材质属性，如 Diffuse →
                    map
                  </Descriptions.Item>
                  <Descriptions.Item label="常见映射关系">
                    <ul style={{ margin: 0, paddingLeft: "20px" }}>
                      <li>Diffuse / Color → map (基础颜色贴图)</li>
                      <li>Rough / Roughness → roughnessMap (粗糙度贴图)</li>
                      <li>nor_gl / NormalGL → normalMap (法线贴图)</li>
                      <li>Metalness → metalnessMap (金属度贴图)</li>
                      <li>AO / AmbientOcclusion → aoMap (环境光遮蔽贴图)</li>
                      <li>Displacement → displacementMap (位移贴图)</li>
                      <li style={{ color: "#722ed1", fontWeight: "bold" }}>
                        arm → aoMap,roughnessMap,metalnessMap (组合贴图)
                        <ul
                          style={{
                            marginTop: "4px",
                            color: "#666",
                            fontWeight: "normal",
                          }}
                        >
                          <li>R 通道 = AO (环境光遮蔽)</li>
                          <li>G 通道 = Roughness (粗糙度)</li>
                          <li>B 通道 = Metalness (金属度)</li>
                        </ul>
                      </li>
                    </ul>
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Spin>
          </Space>
        </Card>
      </div>
    );
  },
});
