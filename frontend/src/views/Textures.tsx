import { defineComponent, onMounted, ref } from "vue";
import { useTextureStore } from "@/stores/texture";
import {
  Button,
  Input,
  Tag,
  Spin,
  Alert,
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
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons-vue";
import type { TextureFile } from "@/api/models/texture";
import "./Textures.less";

const { Search } = Input;

export default defineComponent({
  name: "Textures",
  setup() {
    const textureStore = useTextureStore();
    const keyword = ref("");
    const syncStatus = ref<number | undefined>(undefined);
    const currentPage = ref(1);
    const pageSize = ref(16);

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

    const handlePageChange = (page: number, size: number) => {
      currentPage.value = page;
      pageSize.value = size;
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
      <div class="textures-page">
        {/* 错误提示 */}
        {textureStore.error && (
          <Alert
            message="加载失败"
            description={textureStore.error}
            type="error"
            closable
            onClose={() => (textureStore.error = null)}
            style={{ marginBottom: "16px" }}
          />
        )}

        {/* 顶部工具栏（统计 + 操作） */}
        <div class="header-bar">
          {/* 左侧：统计信息 */}
          <div class="stats-section">
            <div class="stat-item">
              <PictureOutlined class="stat-icon" />
              <div class="stat-content">
                <div class="stat-label">总材质数</div>
                <div class="stat-value">{textureStore.textureCount}</div>
              </div>
            </div>
            <div class="stat-item">
              <SyncOutlined
                class={["stat-icon", textureStore.loading && "spinning"]}
              />
              <div class="stat-content">
                <div class="stat-label">同步状态</div>
                <div
                  class="stat-value"
                  style={{
                    color: textureStore.loading ? "#1890ff" : "#52c41a",
                  }}
                >
                  {textureStore.loading ? "同步中" : "就绪"}
                </div>
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-content">
                <div class="stat-label">当前页</div>
                <div class="stat-value">
                  {currentPage.value}/
                  {Math.ceil(textureStore.textureCount / pageSize.value) || 1}
                </div>
              </div>
            </div>
          </div>

          {/* 右侧：操作区 */}
          <div class="actions-section">
            <Search
              placeholder="搜索材质名称"
              allowClear
              onSearch={handleSearch}
              style={{ width: 240 }}
              v-slots={{
                enterButton: () => <SearchOutlined />,
              }}
            />
            <Select
              placeholder="同步状态"
              allowClear
              value={syncStatus.value}
              onChange={handleStatusChange}
              style={{ width: 130 }}
              options={[
                { label: "全部", value: undefined },
                { label: "未同步", value: 0 },
                { label: "同步中", value: 1 },
                { label: "已同步", value: 2 },
                { label: "失败", value: 3 },
              ]}
            />
            <Select
              value={pageSize.value}
              onChange={(value: any) => {
                pageSize.value = Number(value);
                currentPage.value = 1;
                loadData();
              }}
              style={{ width: 110 }}
              options={[
                { label: "12 条/页", value: 12 },
                { label: "24 条/页", value: 24 },
                { label: "48 条/页", value: 48 },
                { label: "96 条/页", value: 96 },
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
          </div>
        </div>

        {/* 材质网格 */}
        <Spin spinning={textureStore.loading}>
          {textureStore.textures.length === 0 ? (
            <div class="empty-container">
              <Empty description="暂无材质数据" />
            </div>
          ) : (
            <>
              <div class="texture-grid">
                {textureStore.textures.map((texture: any) => {
                  const previewUrl = getPreviewUrl(texture.files);
                  const allImages = getAllImageUrls(texture.files);
                  const otherFiles = getOtherFiles(texture.files);
                  const statusConfig = getStatusConfig(texture.sync_status);

                  return (
                    <div key={texture.id} class="texture-card">
                      {/* 预览图 */}
                      <div class="texture-preview">
                        {previewUrl ? (
                          <Image.PreviewGroup>
                            <Image
                              src={previewUrl}
                              width="100%"
                              height="100%"
                              style={{ objectFit: "cover" }}
                              preview={{ src: previewUrl }}
                            />
                            {allImages.slice(1).map((url, index) => (
                              <Image
                                key={index}
                                src={url}
                                style={{ display: "none" }}
                                preview={{ src: url }}
                              />
                            ))}
                          </Image.PreviewGroup>
                        ) : (
                          <div class="preview-placeholder">
                            <PictureOutlined />
                          </div>
                        )}
                      </div>

                      {/* 材质信息 */}
                      <div class="texture-info">
                        <div class="texture-name" title={texture.name}>
                          {texture.name}
                        </div>
                        <div class="texture-id" title={texture.asset_id}>
                          {texture.asset_id}
                        </div>

                        {/* 状态和分辨率 */}
                        <div class="texture-meta">
                          <Tag color={statusConfig.color}>
                            {statusConfig.text}
                          </Tag>
                          <span class="resolution">
                            {texture.max_resolution}
                          </span>
                        </div>

                        {/* 统计信息 */}
                        <div class="texture-stats">
                          <div class="stats-left">
                            <span class="stat">
                              <EyeOutlined /> {texture.use_count}
                            </span>
                            <span class="stat">
                              <DownloadOutlined /> {texture.download_count}
                            </span>
                          </div>
                          <span class="stat">
                            <PictureOutlined /> {otherFiles.length}
                          </span>
                        </div>

                        {/* 其他文件缩略图 */}
                        {otherFiles.length > 0 && (
                          <div class="other-files">
                            {otherFiles.slice(0, 4).map((file: TextureFile) => (
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
                                preview={{ src: file.full_url }}
                              />
                            ))}
                            {otherFiles.length > 4 && (
                              <div class="more-files">
                                +{otherFiles.length - 4}
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* 分页 */}
              <div class="pagination-container">
                <Pagination
                  current={currentPage.value}
                  pageSize={pageSize.value}
                  total={textureStore.textureCount}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(total: number) => `共 ${total} 条`}
                  onChange={handlePageChange}
                />
              </div>
            </>
          )}
        </Spin>
      </div>
    );
  },
});
