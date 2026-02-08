import { defineComponent, ref, onMounted } from "vue";
import { message, Tag, Image } from "ant-design-vue";
import {
  BoxPlotOutlined,
  ReloadOutlined,
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import { getModels } from "@/api/models";

export default defineComponent({
  name: "Models",
  setup() {
    const loading = ref(false);
    const models = ref<any[]>([]);
    const total = ref(0);
    const currentPage = ref(1);
    const pageSize = ref(24);
    const keyword = ref("");

    // 加载数据
    const loadData = async () => {
      loading.value = true;
      try {
        const res = await getModels({
          page: currentPage.value,
          pageSize: pageSize.value,
        });
        models.value = res.data.list || [];
        total.value = res.data.total || 0;
      } catch (error) {
        message.error("加载失败");
        models.value = [];
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
              icon: BoxPlotOutlined,
              label: "模型总数",
              value: total.value,
              color: "#722ed1",
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
          searchPlaceholder="搜索模型名称"
          pageSize={pageSize.value}
          onPageSizeChange={handlePageSizeChange}
        />

        {/* 网格 */}
        <ResourceGrid
          loading={loading.value}
          data={models.value}
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
                <BoxPlotOutlined />
              </div>
            );
          }}
          renderContent={(item) => (
            <>
              <div class="resource-name" title={item.name}>
                {item.name}
              </div>
              {item.format && (
                <div style={{ marginTop: "8px" }}>
                  <Tag color="purple">{item.format}</Tag>
                </div>
              )}
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
