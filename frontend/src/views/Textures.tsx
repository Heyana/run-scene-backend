import { defineComponent, onMounted, ref } from "vue";
import { useTextureStore } from "@/stores/texture";
import { Tag, Image } from "ant-design-vue";
import {
  PictureOutlined,
  SyncOutlined,
  EyeOutlined,
  DownloadOutlined,
  ReloadOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import type { TextureFile } from "@/api/models/texture";

export default defineComponent({
  name: "Textures",
  setup() {
    const textureStore = useTextureStore();
    const keyword = ref("");
    const syncStatus = ref<number | undefined>(undefined);
    const currentPage = ref(1);
    const pageSize = ref(14);

    // 获取预览图 URL
    const getPreviewUrl = (files?: TextureFile[]) => {
      if (!files || files.length === 0) return "";
      const preview = files.find((f) => f.file_type === "thumbnail");
      if (preview && preview.full_url) return preview.full_url;
      const firstImage = files.find((f) =>
        ["jpg", "jpeg", "png", "webp"].includes(f.format?.toLowerCase()),
      );
      if (firstImage && firstImage.full_url) return firstImage.full_url;
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

    const handlePageSizeChange = (size: number) => {
      pageSize.value = size;
      currentPage.value = 1;
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
      <div
        style={{ padding: "24px", minHeight: "100vh", background: "#f5f5f5" }}
      >
        {/* 头部 */}
        <ResourceHeader
          stats={[
            {
              icon: PictureOutlined,
              label: "总材质数",
              value: textureStore.textureCount,
              color: "#52c41a",
            },
            {
              icon: SyncOutlined,
              label: "同步状态",
              value: textureStore.loading ? "同步中" : "就绪",
              color: textureStore.loading ? "#1890ff" : "#52c41a",
            },
          ]}
          actions={[
            {
              label: "刷新",
              icon: ReloadOutlined,
              loading: textureStore.loading,
              onClick: loadData,
            },
            {
              label: "触发同步",
              icon: SyncOutlined,
              type: "primary",
              loading: textureStore.loading,
              onClick: handleSync,
            },
          ]}
          onSearch={handleSearch}
          searchPlaceholder="搜索材质名称"
          filters={[
            {
              label: "同步状态",
              value: syncStatus.value,
              options: [
                { label: "全部", value: undefined },
                { label: "未同步", value: 0 },
                { label: "同步中", value: 1 },
                { label: "已同步", value: 2 },
                { label: "失败", value: 3 },
              ],
              onChange: handleStatusChange,
            },
          ]}
          pageSize={pageSize.value}
          onPageSizeChange={handlePageSizeChange}
        />

        {/* 网格 */}
        <ResourceGrid
          loading={textureStore.loading}
          data={textureStore.textures}
          total={textureStore.textureCount}
          currentPage={currentPage.value}
          pageSize={pageSize.value}
          onPageChange={handlePageChange}
          renderPreview={(texture) => {
            const previewUrl = getPreviewUrl(texture.files);
            const allImages = getAllImageUrls(texture.files);
            if (previewUrl) {
              return (
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
              );
            }
            return (
              <div class="preview-placeholder">
                <PictureOutlined />
              </div>
            );
          }}
          renderContent={(texture) => {
            const statusConfig = getStatusConfig(texture.sync_status);
            const otherFiles = getOtherFiles(texture.files);
            return (
              <>
                <div class="resource-name" title={texture.name}>
                  {texture.name}
                </div>
                <div
                  style={{ fontSize: "12px", color: "#999", marginTop: "4px" }}
                >
                  {texture.asset_id}
                </div>
                <div
                  style={{
                    display: "flex",
                    gap: "8px",
                    alignItems: "center",
                    marginTop: "8px",
                  }}
                >
                  <Tag color={statusConfig.color}>{statusConfig.text}</Tag>
                  <span style={{ fontSize: "12px", color: "#666" }}>
                    {texture.max_resolution}
                  </span>
                </div>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginTop: "8px",
                    fontSize: "12px",
                    color: "#999",
                  }}
                >
                  <div style={{ display: "flex", gap: "12px" }}>
                    <span>
                      <EyeOutlined /> {texture.use_count}
                    </span>
                    <span>
                      <DownloadOutlined /> {texture.download_count}
                    </span>
                  </div>
                  <span>
                    <PictureOutlined /> {otherFiles.length}
                  </span>
                </div>
                {otherFiles.length > 0 && (
                  <div
                    style={{ display: "flex", gap: "4px", marginTop: "8px" }}
                  >
                    {otherFiles.slice(0, 4).map((file: TextureFile) => (
                      <Image
                        key={file.id}
                        src={file.full_url}
                        width={40}
                        height={40}
                        style={{ objectFit: "cover", borderRadius: "4px" }}
                        preview={{ src: file.full_url }}
                      />
                    ))}
                    {otherFiles.length > 4 && (
                      <div
                        style={{
                          width: "40px",
                          height: "40px",
                          borderRadius: "4px",
                          background: "#f0f0f0",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          fontSize: "12px",
                          color: "#666",
                        }}
                      >
                        +{otherFiles.length - 4}
                      </div>
                    )}
                  </div>
                )}
              </>
            );
          }}
        />
      </div>
    );
  },
});
