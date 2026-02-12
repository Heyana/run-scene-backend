import { defineComponent, ref, onMounted, computed } from "vue";
import {
  Card,
  Table,
  Input,
  Select,
  DatePicker,
  Button,
  Tag,
  Space,
  message,
  Statistic,
  Row,
  Col,
  Modal,
  Descriptions,
} from "ant-design-vue";
import {
  SearchOutlined,
  ReloadOutlined,
  FileTextOutlined,
  UserOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import {
  getAuditLogs,
  getAuditStatistics,
  getArchiveStatistics,
  triggerArchive,
  type AuditLog,
  type AuditFilter,
} from "@/api/audit";
import dayjs, { Dayjs } from "dayjs";
import "./AuditLogs.less";

const { RangePicker } = DatePicker;

export default defineComponent({
  name: "AuditLogs",
  setup() {
    const loading = ref(false);
    const logs = ref<AuditLog[]>([]);
    const total = ref(0);
    const currentPage = ref(1);
    const pageSize = ref(20);

    // 过滤条件
    const filter = ref<AuditFilter>({
      username: "",
      user_ip: "",
      action: undefined,
      resource: undefined,
      status_code: undefined,
      page: 1,
      page_size: 20,
    });

    // 时间范围
    const dateRange = ref<[Dayjs, Dayjs] | null>(null);

    // 统计信息
    const statistics = ref({
      total_count: 0,
      action_count: {} as Record<string, number>,
      resource_count: {} as Record<string, number>,
      top_users: [] as Array<{ username: string; count: number }>,
      top_ips: [] as Array<{ ip: string; count: number }>,
    });

    // 归档统计
    const archiveStats = ref({
      database_count: 0,
      oldest_log: "",
      newest_log: "",
      archive_files: 0,
      retention_days: 7,
      archive_enabled: false,
    });

    // 详情弹窗
    const detailVisible = ref(false);
    const currentLog = ref<AuditLog | null>(null);

    // 操作类型选项
    const actionOptions = [
      { label: "全部", value: undefined },
      { label: "登录", value: "login" },
      { label: "登出", value: "logout" },
      { label: "创建", value: "create" },
      { label: "更新", value: "update" },
      { label: "删除", value: "delete" },
      { label: "查看", value: "view" },
      { label: "下载", value: "download" },
      { label: "上传", value: "upload" },
      { label: "移动", value: "move" },
      { label: "重命名", value: "rename" },
    ];

    // 资源类型选项
    const resourceOptions = [
      { label: "全部", value: undefined },
      { label: "文档", value: "document" },
      { label: "文件夹", value: "folder" },
      { label: "用户", value: "user" },
      { label: "项目", value: "project" },
      { label: "模型", value: "model" },
      { label: "材质", value: "texture" },
    ];

    // 状态码选项
    const statusCodeOptions = [
      { label: "全部", value: undefined },
      { label: "成功 (2xx)", value: 200 },
      { label: "重定向 (3xx)", value: 300 },
      { label: "客户端错误 (4xx)", value: 400 },
      { label: "服务器错误 (5xx)", value: 500 },
    ];

    // 表格列定义
    const columns = [
      {
        title: "时间",
        dataIndex: "created_at",
        key: "created_at",
        width: 180,
        customRender: ({ text }: { text: string }) =>
          dayjs(text).format("YYYY-MM-DD HH:mm:ss"),
      },
      {
        title: "用户",
        dataIndex: "username",
        key: "username",
        width: 120,
      },
      {
        title: "IP地址",
        dataIndex: "user_ip",
        key: "user_ip",
        width: 140,
      },
      {
        title: "操作",
        dataIndex: "action",
        key: "action",
        width: 100,
        customRender: ({ text }: { text: string }) => {
          const colorMap: Record<string, string> = {
            create: "green",
            update: "blue",
            delete: "red",
            upload: "cyan",
            download: "purple",
            login: "gold",
            logout: "default",
          };
          return <Tag color={colorMap[text] || "default"}>{text}</Tag>;
        },
      },
      {
        title: "资源",
        dataIndex: "resource",
        key: "resource",
        width: 100,
      },
      {
        title: "方法",
        dataIndex: "method",
        key: "method",
        width: 80,
        customRender: ({ text }: { text: string }) => {
          const colorMap: Record<string, string> = {
            GET: "blue",
            POST: "green",
            PUT: "orange",
            DELETE: "red",
            PATCH: "purple",
          };
          return <Tag color={colorMap[text] || "default"}>{text}</Tag>;
        },
      },
      {
        title: "路径",
        dataIndex: "path",
        key: "path",
        ellipsis: true,
      },
      {
        title: "状态码",
        dataIndex: "status_code",
        key: "status_code",
        width: 90,
        customRender: ({ text }: { text: number }) => {
          let color = "default";
          if (text >= 200 && text < 300) color = "success";
          else if (text >= 300 && text < 400) color = "processing";
          else if (text >= 400 && text < 500) color = "warning";
          else if (text >= 500) color = "error";
          return <Tag color={color}>{text}</Tag>;
        },
      },
      {
        title: "耗时(ms)",
        dataIndex: "duration",
        key: "duration",
        width: 100,
        customRender: ({ text }: { text: number }) => {
          let color = "default";
          if (text < 100) color = "success";
          else if (text < 500) color = "processing";
          else if (text < 1000) color = "warning";
          else color = "error";
          return <Tag color={color}>{text}</Tag>;
        },
      },
      {
        title: "操作",
        key: "action_btn",
        width: 100,
        fixed: "right" as const,
        customRender: ({ record }: { record: AuditLog }) => (
          <Button type="link" size="small" onClick={() => showDetail(record)}>
            详情
          </Button>
        ),
      },
    ];

    // 加载审计日志
    const loadLogs = async () => {
      loading.value = true;
      try {
        const params: AuditFilter = {
          ...filter.value,
          page: currentPage.value,
          page_size: pageSize.value,
        };

        // 添加时间范围（只有在用户选择了时间范围时才添加）
        if (dateRange.value && dateRange.value[0] && dateRange.value[1]) {
          params.start_time = dateRange.value[0].toISOString();
          params.end_time = dateRange.value[1].toISOString();
        }

        const res = await getAuditLogs(params);
        logs.value = res.data.logs || [];
        total.value = res.data.total;
      } catch (error) {
        console.error("加载审计日志失败:", error);
        message.error("加载审计日志失败");
      } finally {
        loading.value = false;
      }
    };

    // 加载统计信息
    const loadStatistics = async () => {
      try {
        // 默认统计最近7天
        const startTime = dayjs().subtract(7, "day").toISOString();
        const endTime = dayjs().toISOString();

        const res = await getAuditStatistics(startTime, endTime);
        statistics.value = res.data;
      } catch (error) {
        console.error("加载统计信息失败:", error);
      }
    };

    // 加载归档统计
    const loadArchiveStats = async () => {
      try {
        const res = await getArchiveStatistics();
        archiveStats.value = res.data;
      } catch (error) {
        console.error("加载归档统计失败:", error);
      }
    };

    // 搜索
    const handleSearch = () => {
      currentPage.value = 1;
      loadLogs();
      loadStatistics();
    };

    // 重置
    const handleReset = () => {
      filter.value = {
        username: "",
        user_ip: "",
        action: undefined,
        resource: undefined,
        status_code: undefined,
        page: 1,
        page_size: 20,
      };
      dateRange.value = null;
      currentPage.value = 1;
      loadLogs();
      loadStatistics();
    };

    // 分页变化
    const handlePageChange = (page: number, size: number) => {
      currentPage.value = page;
      pageSize.value = size;
      loadLogs();
    };

    // 显示详情
    const showDetail = (log: AuditLog) => {
      currentLog.value = log;
      detailVisible.value = true;
    };

    // 手动归档
    const handleArchive = async () => {
      try {
        const res = await triggerArchive();
        message.success(`归档成功，共归档 ${res.data.archived_count} 条日志`);
        loadLogs();
        loadArchiveStats();
      } catch (error) {
        console.error("归档失败:", error);
        message.error("归档失败");
      }
    };

    // 格式化时间
    const formatTime = (dateStr: string) => {
      if (!dateStr) return "未知";
      return dayjs(dateStr).format("YYYY-MM-DD HH:mm:ss");
    };

    onMounted(() => {
      // 默认不设置时间范围，查询所有日志
      // 用户可以手动选择时间范围
      loadLogs();
      loadStatistics();
      loadArchiveStats();
    });

    return () => (
      <div class="audit-logs-page">
        {/* 头部 */}
        <ResourceHeader
          stats={[
            {
              icon: FileTextOutlined,
              label: "总日志数",
              value: statistics.value.total_count,
              color: "#1890ff",
            },
            {
              icon: DatabaseOutlined,
              label: "数据库日志",
              value: archiveStats.value.database_count,
              color: "#52c41a",
            },
            {
              icon: FileTextOutlined,
              label: "归档文件",
              value: archiveStats.value.archive_files,
              color: "#722ed1",
            },
            {
              icon: ClockCircleOutlined,
              label: "保留天数",
              value: `${archiveStats.value.retention_days}天`,
              color: "#fa8c16",
            },
          ]}
          showHomeButton={true}
        />

        {/* 搜索表单 */}
        <Card bordered={false} style={{ marginBottom: "16px" }}>
          <Space direction="vertical" size="middle" style={{ width: "100%" }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12} lg={6}>
                <Input
                  v-model:value={filter.value.username}
                  placeholder="用户名"
                  prefix={<UserOutlined />}
                  allowClear
                />
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Input
                  v-model:value={filter.value.user_ip}
                  placeholder="IP地址"
                  allowClear
                />
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Select
                  v-model:value={filter.value.action}
                  placeholder="操作类型"
                  options={actionOptions}
                  style={{ width: "100%" }}
                  allowClear
                />
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Select
                  v-model:value={filter.value.resource}
                  placeholder="资源类型"
                  options={resourceOptions}
                  style={{ width: "100%" }}
                  allowClear
                />
              </Col>
            </Row>
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12} lg={6}>
                <Select
                  v-model:value={filter.value.status_code}
                  placeholder="状态码"
                  options={statusCodeOptions}
                  style={{ width: "100%" }}
                  allowClear
                />
              </Col>
              <Col xs={24} sm={12} lg={12}>
                <RangePicker
                  v-model:value={dateRange.value}
                  showTime
                  format="YYYY-MM-DD HH:mm:ss"
                  style={{ width: "100%" }}
                />
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Space>
                  <Button
                    type="primary"
                    icon={<SearchOutlined />}
                    onClick={handleSearch}
                  >
                    搜索
                  </Button>
                  <Button icon={<ReloadOutlined />} onClick={handleReset}>
                    重置
                  </Button>
                  {archiveStats.value.archive_enabled && (
                    <Button onClick={handleArchive}>手动归档</Button>
                  )}
                </Space>
              </Col>
            </Row>
          </Space>
        </Card>

        {/* 审计日志表格 */}
        <Card bordered={false}>
          <Table
            columns={columns}
            dataSource={logs.value}
            rowKey="id"
            loading={loading.value}
            pagination={{
              current: currentPage.value,
              pageSize: pageSize.value,
              total: total.value,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条`,
              onChange: handlePageChange,
            }}
            scroll={{ x: 1400 }}
          />
        </Card>

        {/* 详情弹窗 */}
        <Modal
          v-model:open={detailVisible.value}
          title="审计日志详情"
          width={800}
          footer={null}
        >
          {currentLog.value && (
            <Descriptions bordered column={2}>
              <Descriptions.Item label="ID">
                {currentLog.value.id}
              </Descriptions.Item>
              <Descriptions.Item label="用户">
                {currentLog.value.username}
              </Descriptions.Item>
              <Descriptions.Item label="用户ID">
                {currentLog.value.user_id || "N/A"}
              </Descriptions.Item>
              <Descriptions.Item label="IP地址">
                {currentLog.value.user_ip}
              </Descriptions.Item>
              <Descriptions.Item label="操作类型">
                {currentLog.value.action}
              </Descriptions.Item>
              <Descriptions.Item label="资源类型">
                {currentLog.value.resource}
              </Descriptions.Item>
              <Descriptions.Item label="资源ID">
                {currentLog.value.resource_id || "N/A"}
              </Descriptions.Item>
              <Descriptions.Item label="HTTP方法">
                {currentLog.value.method}
              </Descriptions.Item>
              <Descriptions.Item label="请求路径" span={2}>
                {currentLog.value.path}
              </Descriptions.Item>
              <Descriptions.Item label="状态码">
                {currentLog.value.status_code}
              </Descriptions.Item>
              <Descriptions.Item label="耗时">
                {currentLog.value.duration} ms
              </Descriptions.Item>
              <Descriptions.Item label="创建时间" span={2}>
                {formatTime(currentLog.value.created_at)}
              </Descriptions.Item>
              {currentLog.value.user_agent && (
                <Descriptions.Item label="User-Agent" span={2}>
                  {currentLog.value.user_agent}
                </Descriptions.Item>
              )}
              {currentLog.value.request_body && (
                <Descriptions.Item label="请求体" span={2}>
                  <pre style={{ maxHeight: "200px", overflow: "auto" }}>
                    {currentLog.value.request_body}
                  </pre>
                </Descriptions.Item>
              )}
              {currentLog.value.error_msg && (
                <Descriptions.Item label="错误信息" span={2}>
                  <pre
                    style={{
                      maxHeight: "200px",
                      overflow: "auto",
                      color: "red",
                    }}
                  >
                    {currentLog.value.error_msg}
                  </pre>
                </Descriptions.Item>
              )}
            </Descriptions>
          )}
        </Modal>
      </div>
    );
  },
});
