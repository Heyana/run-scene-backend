import { defineComponent, ref, onMounted } from "vue";
import { message, Tag, Image, Progress } from "ant-design-vue";
import {
  RocketOutlined,
  ReloadOutlined,
  PlusOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import { getAI3DTasks } from "@/api/ai3d";

export default defineComponent({
  name: "AI3D",
  setup() {
    const loading = ref(false);
    const tasks = ref<any[]>([]);
    const total = ref(0);
    const currentPage = ref(1);
    const pageSize = ref(24);
    const provider = ref<string | undefined>(undefined);
    const status = ref<string | undefined>(undefined);

    // 获取状态配置
    const getStatusConfig = (status: string) => {
      const statusMap: Record<string, { color: string; text: string }> = {
        WAIT: { color: "default", text: "等待中" },
        RUN: { color: "processing", text: "运行中" },
        DONE: { color: "success", text: "已完成" },
        FAIL: { color: "error", text: "失败" },
        // 兼容旧的状态值
        pending: { color: "default", text: "等待中" },
        processing: { color: "processing", text: "生成中" },
        completed: { color: "success", text: "已完成" },
        failed: { color: "error", text: "失败" },
      };
      return statusMap[status] ?? statusMap["WAIT"]!;
    };

    // 加载数据
    const loadData = async () => {
      loading.value = true;
      try {
        const res = await getAI3DTasks({
          page: currentPage.value,
          pageSize: pageSize.value,
          provider: provider.value,
          status: status.value,
        });
        tasks.value = res.data.list || [];
        total.value = res.data.total || 0;
      } catch (error) {
        message.error("加载失败");
        tasks.value = [];
        total.value = 0;
      } finally {
        loading.value = false;
      }
    };

    const handleProviderChange = (value: any) => {
      provider.value = value === undefined ? undefined : String(value);
      currentPage.value = 1;
      loadData();
    };

    const handleStatusChange = (value: any) => {
      status.value = value === undefined ? undefined : String(value);
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

    const handleCreate = () => {
      message.info("创建功能待实现");
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
              icon: RocketOutlined,
              label: "任务总数",
              value: total.value,
              color: "#1890ff",
            },
          ]}
          actions={[
            {
              label: "刷新",
              icon: ReloadOutlined,
              loading: loading.value,
              onClick: loadData,
            },
            {
              label: "创建任务",
              icon: PlusOutlined,
              type: "primary",
              onClick: handleCreate,
            },
          ]}
          filters={[
            {
              label: "平台",
              value: provider.value,
              options: [
                { label: "全部", value: undefined },
                { label: "Meshy", value: "meshy" },
                { label: "混元3D", value: "hunyuan" },
              ],
              onChange: handleProviderChange,
            },
            {
              label: "状态",
              value: status.value,
              options: [
                { label: "全部", value: undefined },
                { label: "等待中", value: "WAIT" },
                { label: "运行中", value: "RUN" },
                { label: "已完成", value: "DONE" },
                { label: "失败", value: "FAIL" },
              ],
              onChange: handleStatusChange,
            },
          ]}
          pageSize={pageSize.value}
          onPageSizeChange={handlePageSizeChange}
        />

        {/* 网格 */}
        <ResourceGrid
          loading={loading.value}
          data={tasks.value}
          total={total.value}
          currentPage={currentPage.value}
          pageSize={pageSize.value}
          onPageChange={handlePageChange}
          renderPreview={(task) => {
            if (task.thumbnailUrl) {
              return (
                <Image
                  src={task.thumbnailUrl}
                  width="100%"
                  height="100%"
                  style={{ objectFit: "cover" }}
                  preview={{ src: task.thumbnailUrl }}
                />
              );
            }
            return (
              <div class="preview-placeholder">
                <RocketOutlined />
              </div>
            );
          }}
          renderContent={(task) => {
            const statusConfig = getStatusConfig(task.status);
            return (
              <>
                <div
                  class="resource-name"
                  title={task.prompt}
                  style={{
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    display: "-webkit-box",
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: "vertical",
                  }}
                >
                  {task.prompt}
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
                  <Tag color="blue">{task.provider}</Tag>
                </div>
                {(task.status === "RUN" || task.status === "processing") && (
                  <div style={{ marginTop: "8px" }}>
                    <Progress
                      percent={task.progress || 0}
                      size="small"
                      status="active"
                    />
                  </div>
                )}
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    marginTop: "8px",
                    fontSize: "12px",
                    color: "#999",
                  }}
                >
                  <ClockCircleOutlined style={{ marginRight: "4px" }} />
                  {new Date(task.createdAt).toLocaleString()}
                </div>
              </>
            );
          }}
        />
      </div>
    );
  },
});
