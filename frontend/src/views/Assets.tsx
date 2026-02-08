import { defineComponent, ref, onMounted } from "vue";
import { message, Tag, Image } from "ant-design-vue";
import {
  FileImageOutlined,
  ReloadOutlined,
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import { getAssets } from "@/api/assets";

export default defineComponent({
  name: "Assets",
  setup() {
    const loading = ref(false);
    const assets = ref<any[]>([]);
    const total = ref(0);
    const currentPage = ref(1);
    const pageSize = ref(24);
    const keyword = ref("");
    const assetType = ref<string | undefined>(undefined);

    // 加载数据
    const loadData = async () => {
      loading.value = true;
      try {
        const res = await getAssets({
          page: currentPage.value,
          pageSize: pageSize.value,
          type: assetType.value,
        });
        assets.value = res.data.list || [];
        total.value = res.data.total || 0;
      } catch (error) {
        message.error("加载失败");
        assets.value = [];
        total.value = 0;
      } finally {
        loading.value = false;
      }
    };

    // 搜索
    const handleSearch = (value: string) => {
      keyword.value = value;
      currentPage.value = 1;
      loadData();
    };

    // 类型筛选
    const handleTypeChange = (value: any) => {
      assetType.value = value === undefined ? undefined : String(value);
      currentPage.value = 1;
      loadData();
    };

    // 分页
    const handlePageChange = (page: number, size: number) => {
      currentPage.value = page;
      pageSize.value = size;
      loadData();
    };

    // 分页大小变化
    const handlePageSizeChange = (size: number) => {
      pageSize.value = size;
      currentPage.value = 1;
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
              icon: FileImageOutlined,
              label: "资产总数",
              value: total.value,
              color: "#fa8c16",
            },
          ]}
          actions={[
            {
              label: "刷新",
              icon: ReloadOutlined,
              loading: loading.value,
              onClick: loadData,
            },
          ]}
          onSearch={handleSearch}
          searchPlaceholder="搜索资产名称"
          filters={[
            {
              label: "资产类型",
              value: assetType.value,
              options: [
                { label: "全部", value: undefined },
                { label: "图片", value: "image" },
                { label: "视频", value: "video" },
                { label: "音频", value: "audio" },
                { label: "文档", value: "document" },
                { label: "其他", value: "other" },
              ],
              onChange: handleTypeChange,
            },
          ]}
          pageSize={pageSize.value}
          onPageSizeChange={handlePageSizeChange}
        />

        {/* 网格 */}
        <ResourceGrid
          loading={loading.value}
          data={assets.value}
          total={total.value}
          currentPage={currentPage.value}
          pageSize={pageSize.value}
          onPageChange={handlePageChange}
          renderPreview={(item) => {
            if (item.thumbnail_url) {
              return (
                <Image
                  src={item.thumbnail_url}
                  width="100%"
                  height="100%"
                  style={{ objectFit: "cover" }}
                  preview={{ src: item.thumbnail_url }}
                />
              );
            }
            return (
              <div class="preview-placeholder">
                <FileImageOutlined />
              </div>
            );
          }}
          renderContent={(item) => (
            <>
              <div class="resource-name" title={item.name}>
                {item.name}
              </div>
              <div style={{ display: "flex", gap: "8px", marginTop: "8px" }}>
                {item.type && (
                  <Tag color="orange">
                    {item.type === "image" && "图片"}
                    {item.type === "video" && "视频"}
                    {item.type === "audio" && "音频"}
                    {item.type === "document" && "文档"}
                    {item.type === "other" && "其他"}
                  </Tag>
                )}
                {item.format && <Tag>{item.format}</Tag>}
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
                    <EyeOutlined /> {item.use_count || 0}
                  </span>
                  <span>
                    <DownloadOutlined /> {item.download_count || 0}
                  </span>
                </div>
                <span>
                  {((item.file_size || 0) / 1024 / 1024).toFixed(2)} MB
                </span>
              </div>
            </>
          )}
        />
      </div>
    );
  },
});
