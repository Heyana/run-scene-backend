import { defineComponent, onMounted, ref } from "vue";
import { useTextureStore } from "@/stores/texture";
import {
  Card,
  Button,
  Space,
  Input,
  Tag,
  Spin,
  Alert,
  Statistic,
  Row,
  Col,
  Image,
  Select,
  Pagination,
  Empty,
} from "ant-design-vue";
import {
  ReloadOutlined,
  SearchOutlined,
  SyncOutlined,
  PictureOutlined,
  ClockCircleOutlined,
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons-vue";
import type { TextureFile } from "@/api/models/texture";

const { Search } = Input;

export default defineComponent({
  name: "Textures",
  setup() {
    const textureStore = useTextureStore();
    const keyword = ref("");
    const syncStatus = ref<number | undefined>(undefined);
    const currentPage = ref(1);
    const pageSize = ref(10);

    // 获取预览图 URL
    const getPreviewUrl = (files?: TextureFile[]) => {
      if (!files || files.length === 0) return "";

      // 优先查找 preview 类型的文件
      const preview = files.find((f) => f.file_type === "thumbnail");
      if (preview && preview.full_url) {
        return preview.full_url;
      }

      // 否则返回第一个图片文件
      const firstImage = files.find((f) =>
        ["jpg", "jpeg", "png", "webp"].includes(f.format?.toLowerCase()),
      );
      if (firstImage && firstImage.full_url) {
        return firstImage.full_url;
      }

      return "";
    };

    // 获取所有图片 URL（用于预览组）
    const getAllImageUrls = (files?: TextureFile[]) => {
      if (!files || files.length === 0) return [];

      return files
        .filter((f) =>
          ["jpg", "jpeg", "png", "webp"].includes(f.format?.toLowerCase()),
        )
        .map((f) => f.full_url)
        .filter((url) => url);
    };

    // 获取其他文件（排除缩略图）
    const getOtherFiles = (files?: TextureFile[]) => {
      if (!files || files.length === 0) return [];

      return files.filter((f: TextureFile) => {
        if (f.file_type === "thumbnail") return false;
        return ["jpg", "jpeg", "png", "webp"].includes(f.format?.toLowerCase());
      });
    };

    // 获取状态配置
    const getStatusConfig = (status: number) => {
      const statusMap: Record<number, { color: string; text: string }> = {
        0: { color: "default", text: "未同步" },
        1: { color: "processing", text: "同步中" },
        2: { color: "success", text: "已同步" },
        3: { color: "error", text: "失败" },
      };
      return statusMap[status] ?? statusMap[0]!;
    };

    const columns = [
      {
        title: "预览",
        dataIndex: "files",
        key: "preview",
        width: 120,
        customRender: ({ record }: { record: any }) => {
          const previewUrl = getPreviewUrl(record.files);
          const allImages = getAllImageUrls(record.files);

          if (!previewUrl) {
            return (
              <div
                style={{
                  width: "80px",
                  height: "80px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: "#f0f0f0",
                  borderRadius: "4px",
                }}
              >
                <PictureOutlined style={{ fontSize: "24px", color: "#ccc" }} />
              </div>
            );
          }

          return (
            <Image.PreviewGroup>
              <Image
                src={previewUrl}
                width={80}
                height={80}
                style={{ objectFit: "cover", borderRadius: "4px" }}
                preview={{
                  src: previewUrl,
                }}
              />
              {allImages.slice(1).map((url, index) => (
                <Image
                  key={index}
                  src={url}
                  style={{ display: "none" }}
                  preview={{
                    src: url,
                  }}
                />
              ))}
            </Image.PreviewGroup>
          );
        },
      },
      {
        title: "ID",
        dataIndex: "id",
        key: "id",
        width: 80,
      },
      {
        title: "资产ID",
        dataIndex: "asset_id",
        key: "asset_id",
        width: 150,
      },
      {
        title: "名称",
        dataIndex: "name",
        key: "name",
        ellipsis: true,
      },
      {
        title: "其他文件",
        dataIndex: "files",
        key: "other_files",
        width: 300,
        customRender: ({ record }: { record: any }) => {
          const files = record.files || [];

          return <Tag color="blue">{files.length} 个</Tag>;
          // 过滤掉预览图，只显示其他图片
          const otherImages = files.filter((f: TextureFile) => {
            // 排除 preview 类型
            if (f.file_type === "thumbnail") return false;
            // 只保留图片格式
            return ["jpg", "jpeg", "png", "webp"].includes(
              f.format?.toLowerCase(),
            );
          });

          if (otherImages.length === 0) {
            return <span style={{ color: "#999" }}>无其他图片</span>;
          }

          return (
            <Space size={4} wrap>
              {otherImages.map((file: TextureFile) => (
                <Image
                  key={file.id}
                  src={file.full_url}
                  width={40}
                  height={40}
                  style={{
                    objectFit: "cover",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                  preview={{
                    src: file.full_url,
                  }}
                />
              ))}
              <Tag color="blue">{otherImages.length} 个</Tag>
            </Space>
          );
        },
      },
      {
        title: "分辨率",
        dataIndex: "max_resolution",
        key: "max_resolution",
        width: 120,
      },
      {
        title: "使用次数",
        dataIndex: "use_count",
        key: "use_count",
        width: 100,
        sorter: true,
      },
      {
        title: "下载次数",
        dataIndex: "download_count",
        key: "download_count",
        width: 100,
      },
      {
        title: "状态",
        dataIndex: "sync_status",
        key: "sync_status",
        width: 100,
        customRender: ({ text }: { text: number }) => {
          const statusMap: Record<number, { color: string; text: string }> = {
            0: { color: "default", text: "未同步" },
            1: { color: "processing", text: "同步中" },
            2: { color: "success", text: "已同步" },
            3: { color: "error", text: "失败" },
          };
          const status = statusMap[text] ?? statusMap[0]!;
          return <Tag color={status.color}>{status.text}</Tag>;
        },
      },
      {
        title: "创建时间",
        dataIndex: "created_at",
        key: "created_at",
        width: 180,
        customRender: ({ text }: { text: string }) => {
          return new Date(text).toLocaleString("zh-CN");
        },
      },
    ];

    const loadData = async () => {
      await textureStore.fetchTextures({
        page: currentPage.value,
        pageSize: pageSize.value,
        keyword: keyword.value || undefined,
        syncStatus: syncStatus.value,
      });
    };

    const handleSearch = (value: string) => {
      keyword.value = value;
      currentPage.value = 1;
      loadData();
    };

    const handleStatusChange = (value: any) => {
      syncStatus.value = value === undefined ? undefined : Number(value);
      currentPage.value = 1;
      loadData();
    };

    const handleTableChange = (pagination: any) => {
      currentPage.value = pagination.current;
      pageSize.value = pagination.pageSize;
      loadData();
    };

    const handleSync = async () => {
      await textureStore.triggerSync();
      loadData();
    };

    onMounted(() => {
      loadData();
    });

    return () => (
      <div class="textures-container">
        <Card bordered={false}>
          <Space direction="vertical" size="large" style={{ width: "100%" }}>
            {/* 统计信息 */}
            <Row gutter={16}>
              <Col span={8}>
                <Card>
                  <Statistic
                    title="总材质数"
                    value={textureStore.textureCount}
                    v-slots={{
                      prefix: () => <PictureOutlined />,
                    }}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    title="同步状态"
                    value={textureStore.loading ? "同步中" : "就绪"}
                    valueStyle={{
                      color: textureStore.loading ? "#1890ff" : "#52c41a",
                    }}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    title="当前页"
                    value={`${currentPage.value}/${Math.ceil(textureStore.textureCount / pageSize.value) || 1}`}
                  />
                </Card>
              </Col>
            </Row>

            {/* 错误提示 */}
            {textureStore.error && (
              <Alert
                message="加载失败"
                description={textureStore.error}
                type="error"
                closable
                onClose={() => (textureStore.error = null)}
              />
            )}

            {/* 操作栏 */}
            <Space>
              <Search
                placeholder="搜索材质名称"
                allowClear
                onSearch={handleSearch}
                style={{ width: 300 }}
                v-slots={{
                  enterButton: () => <SearchOutlined />,
                }}
              />
              <Select
                placeholder="同步状态"
                allowClear
                value={syncStatus.value}
                onChange={handleStatusChange}
                style={{ width: 150 }}
                options={[
                  { label: "全部", value: undefined },
                  { label: "未同步", value: 0 },
                  { label: "同步中", value: 1 },
                  { label: "已同步", value: 2 },
                  { label: "失败", value: 3 },
                ]}
              />
              <Button onClick={loadData} loading={textureStore.loading}>
                {{
                  icon: () => <ReloadOutlined />,
                  default: () => "刷新",
                }}
              </Button>
              <Button
                type="primary"
                onClick={handleSync}
                loading={textureStore.loading}
              >
                {{
                  icon: () => <SyncOutlined />,
                  default: () => "触发同步",
                }}
              </Button>
            </Space>

            {/* 数据表格 */}
            <Spin spinning={textureStore.loading}>
              <Table
                columns={columns}
                dataSource={textureStore.textures}
                rowKey="id"
                pagination={{
                  current: currentPage.value,
                  pageSize: pageSize.value,
                  total: textureStore.textureCount,
                  showSizeChanger: true,
                  showQuickJumper: true,
                  showTotal: (total: number) => `共 ${total} 条`,
                }}
                onChange={handleTableChange}
              />
            </Spin>
          </Space>
        </Card>
      </div>
    );
  },
});
