import { defineComponent, ref, onMounted } from "vue";
import {
  message,
  Tag,
  Modal,
  Upload,
  Form,
  Input,
  Select,
  Image,
} from "ant-design-vue";
import {
  FileOutlined,
  ReloadOutlined,
  EyeOutlined,
  DownloadOutlined,
  UploadOutlined,
  FilePdfOutlined,
  FileWordOutlined,
  FileExcelOutlined,
  FilePptOutlined,
  FileTextOutlined,
  FileZipOutlined,
  VideoCameraOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import { getDocuments, uploadDocument } from "@/api/documents";

export default defineComponent({
  name: "Documents",
  setup() {
    const loading = ref(false);
    const documents = ref<any[]>([]);
    const total = ref(0);
    const currentPage = ref(1);
    const pageSize = ref(24);
    const keyword = ref("");
    const docType = ref<string | undefined>(undefined);
    const uploadVisible = ref(false);
    const uploadLoading = ref(false);
    const previewVisible = ref(false);
    const previewItem = ref<any>(null);
    const formRef = ref();
    const fileList = ref<any[]>([]);

    const formState = ref({
      name: "",
      description: "",
      category: "",
      department: "",
      project: "",
      tags: "",
    });

    // 加载数据
    const loadData = async () => {
      loading.value = true;
      try {
        const res = await getDocuments({
          page: currentPage.value,
          pageSize: pageSize.value,
          type: docType.value,
          keyword: keyword.value,
        });
        documents.value = res.data.list || [];
        total.value = res.data.total || 0;
      } catch (error) {
        message.error("加载失败");
        documents.value = [];
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
      docType.value = value === undefined ? undefined : String(value);
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

    // 打开上传对话框
    const handleUpload = () => {
      uploadVisible.value = true;
      formState.value = {
        name: "",
        description: "",
        category: "",
        department: "",
        project: "",
        tags: "",
      };
      fileList.value = [];
    };

    // 文件上传前
    const beforeUpload = (file: any) => {
      fileList.value = [file];
      if (!formState.value.name) {
        // 自动填充文件名（去除扩展名）
        const fileName = file.name;
        const lastDotIndex = fileName.lastIndexOf(".");
        formState.value.name =
          lastDotIndex > 0 ? fileName.substring(0, lastDotIndex) : fileName;
      }
      return false; // 阻止自动上传
    };

    // 提交上传
    const handleSubmit = async () => {
      try {
        await formRef.value.validate();

        if (fileList.value.length === 0) {
          message.error("请选择文件");
          return;
        }

        uploadLoading.value = true;
        const formData = new FormData();
        formData.append("file", fileList.value[0]);
        formData.append("name", formState.value.name);
        formData.append("description", formState.value.description);
        formData.append("category", formState.value.category);
        formData.append("department", formState.value.department);
        formData.append("project", formState.value.project);
        formData.append("tags", formState.value.tags);

        await uploadDocument(formData);
        message.success("上传成功");
        uploadVisible.value = false;
        loadData();
      } catch (error: any) {
        // 处理重复文件错误
        const errorMsg =
          error.response?.data?.msg || error.message || "上传失败";
        if (errorMsg.includes("文件已存在")) {
          Modal.warning({
            title: "文件已存在",
            content: errorMsg,
            okText: "知道了",
          });
        } else {
          message.error(errorMsg);
        }
      } finally {
        uploadLoading.value = false;
      }
    };

    // 获取文件图标
    const getFileIcon = (format: string) => {
      const iconMap: Record<string, any> = {
        pdf: FilePdfOutlined,
        doc: FileWordOutlined,
        docx: FileWordOutlined,
        xls: FileExcelOutlined,
        xlsx: FileExcelOutlined,
        ppt: FilePptOutlined,
        pptx: FilePptOutlined,
        txt: FileTextOutlined,
        md: FileTextOutlined,
        zip: FileZipOutlined,
        rar: FileZipOutlined,
        "7z": FileZipOutlined,
        mp4: VideoCameraOutlined,
        webm: VideoCameraOutlined,
        avi: VideoCameraOutlined,
        mov: VideoCameraOutlined,
      };
      return iconMap[format?.toLowerCase()] || FileOutlined;
    };

    // 判断是否可预览
    const canPreview = (item: any) => {
      const imageFormats = ["jpg", "jpeg", "png", "gif", "webp"];
      const videoFormats = ["mp4", "webm"];
      return (
        imageFormats.includes(item.format?.toLowerCase()) ||
        videoFormats.includes(item.format?.toLowerCase())
      );
    };

    // 判断是否是视频
    const isVideo = (item: any) => {
      const videoFormats = ["mp4", "webm", "avi", "mov"];
      return videoFormats.includes(item.format?.toLowerCase());
    };

    // 判断是否是图片
    const isImage = (item: any) => {
      const imageFormats = ["jpg", "jpeg", "png", "gif", "webp"];
      return imageFormats.includes(item.format?.toLowerCase());
    };

    // 点击项目
    const handleItemClick = (item: any) => {
      if (canPreview(item)) {
        // 可预览的文件，打开预览
        previewItem.value = item;
        previewVisible.value = true;
      } else {
        // 不可预览的文件，直接下载
        handleDownload(item);
      }
    };

    // 下载文件
    const handleDownload = (item: any) => {
      window.open(item.file_url, "_blank");
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
              icon: FileOutlined,
              label: "文档总数",
              value: total.value,
              color: "#1890ff",
            },
          ]}
          actions={[
            {
              label: "上传文档",
              icon: UploadOutlined,
              type: "primary",
              onClick: handleUpload,
            },
            {
              label: "刷新",
              icon: ReloadOutlined,
              loading: loading.value,
              onClick: loadData,
            },
          ]}
          onSearch={handleSearch}
          searchPlaceholder="搜索文档名称"
          filters={[
            {
              label: "文档类型",
              value: docType.value,
              options: [
                { label: "全部", value: undefined },
                { label: "文档", value: "document" },
                { label: "视频", value: "video" },
                { label: "压缩包", value: "archive" },
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
          data={documents.value}
          total={total.value}
          currentPage={currentPage.value}
          pageSize={pageSize.value}
          onPageChange={handlePageChange}
          onCardClick={handleItemClick}
          renderPreview={(item) => {
            const IconComponent = getFileIcon(item.format);

            // 图片预览
            if (isImage(item) && item.thumbnail_url) {
              return (
                <div
                  style={{
                    position: "relative",
                    width: "100%",
                    height: "100%",
                  }}
                >
                  <img
                    src={item.thumbnail_url}
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover",
                    }}
                  />
                  <div
                    style={{
                      position: "absolute",
                      top: "50%",
                      left: "50%",
                      transform: "translate(-50%, -50%)",
                      fontSize: "32px",
                      color: "rgba(255, 255, 255, 0.8)",
                      opacity: 0,
                      transition: "opacity 0.3s",
                    }}
                    class="preview-icon"
                  >
                    <EyeOutlined />
                  </div>
                </div>
              );
            }

            // 视频预览
            if (isVideo(item) && item.thumbnail_url) {
              return (
                <div
                  style={{
                    position: "relative",
                    width: "100%",
                    height: "100%",
                  }}
                >
                  <img
                    src={item.thumbnail_url}
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover",
                    }}
                  />
                  <div
                    style={{
                      position: "absolute",
                      top: "50%",
                      left: "50%",
                      transform: "translate(-50%, -50%)",
                      fontSize: "48px",
                      color: "rgba(255, 255, 255, 0.9)",
                    }}
                  >
                    <PlayCircleOutlined />
                  </div>
                </div>
              );
            }

            // 其他文件显示图标
            if (item.thumbnail_url) {
              return (
                <img
                  src={item.thumbnail_url}
                  style={{
                    width: "100%",
                    height: "100%",
                    objectFit: "cover",
                  }}
                />
              );
            }

            return (
              <div
                class="preview-placeholder"
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  height: "100%",
                  fontSize: "48px",
                  color: "#1890ff",
                }}
              >
                <IconComponent />
              </div>
            );
          }}
          renderContent={(item) => (
            <>
              <div class="resource-name" title={item.name}>
                {item.name}
              </div>
              <div
                style={{
                  display: "flex",
                  gap: "8px",
                  marginTop: "8px",
                  flexWrap: "wrap",
                }}
              >
                {item.type && (
                  <Tag color="blue">
                    {item.type === "document" && "文档"}
                    {item.type === "video" && "视频"}
                    {item.type === "archive" && "压缩包"}
                    {item.type === "other" && "其他"}
                  </Tag>
                )}
                {item.format && <Tag>{item.format.toUpperCase()}</Tag>}
                {item.category && <Tag color="green">{item.category}</Tag>}
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
                    <EyeOutlined /> {item.view_count || 0}
                  </span>
                  <span>
                    <DownloadOutlined /> {item.download_count || 0}
                  </span>
                </div>
                <span>
                  {((item.file_size || 0) / 1024 / 1024).toFixed(2)} MB
                </span>
              </div>
              {item.version && (
                <div
                  style={{ marginTop: "4px", fontSize: "12px", color: "#999" }}
                >
                  版本: {item.version}
                </div>
              )}
            </>
          )}
          onItemClick={handleItemClick}
        />

        {/* 上传对话框 */}
        <Modal
          title="上传文档"
          open={uploadVisible.value}
          onOk={handleSubmit}
          onCancel={() => (uploadVisible.value = false)}
          confirmLoading={uploadLoading.value}
          width={600}
        >
          <Form
            ref={formRef}
            model={formState.value}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 18 }}
          >
            <Form.Item label="文件">
              <Upload
                beforeUpload={beforeUpload}
                fileList={fileList.value}
                maxCount={1}
                onRemove={() => (fileList.value = [])}
              >
                <a-button icon={<UploadOutlined />}>选择文件</a-button>
              </Upload>
            </Form.Item>
            <Form.Item
              label="文档名称"
              name="name"
              rules={[{ required: true, message: "请输入文档名称" }]}
            >
              <Input
                v-model:value={formState.value.name}
                placeholder="请输入文档名称"
              />
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={formState.value.description}
                placeholder="请输入描述"
                rows={3}
              />
            </Form.Item>
            <Form.Item label="分类" name="category">
              <Select
                v-model:value={formState.value.category}
                placeholder="请选择分类"
                allowClear
              >
                <Select.Option value="技术文档">技术文档</Select.Option>
                <Select.Option value="产品文档">产品文档</Select.Option>
                <Select.Option value="设计文档">设计文档</Select.Option>
                <Select.Option value="会议记录">会议记录</Select.Option>
                <Select.Option value="项目资料">项目资料</Select.Option>
                <Select.Option value="培训资料">培训资料</Select.Option>
                <Select.Option value="其他">其他</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item label="部门" name="department">
              <Input
                v-model:value={formState.value.department}
                placeholder="请输入部门"
              />
            </Form.Item>
            <Form.Item label="项目" name="project">
              <Input
                v-model:value={formState.value.project}
                placeholder="请输入项目名称"
              />
            </Form.Item>
            <Form.Item label="标签" name="tags">
              <Input
                v-model:value={formState.value.tags}
                placeholder="多个标签用逗号分隔"
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 预览对话框 */}
        <Modal
          title={previewItem.value?.name}
          open={previewVisible.value}
          onCancel={() => (previewVisible.value = false)}
          footer={[
            <a-button
              key="download"
              onClick={() => handleDownload(previewItem.value)}
            >
              <DownloadOutlined /> 下载
            </a-button>,
            <a-button
              key="close"
              type="primary"
              onClick={() => (previewVisible.value = false)}
            >
              关闭
            </a-button>,
          ]}
          width={800}
        >
          {previewItem.value && (
            <div style={{ textAlign: "center" }}>
              {isImage(previewItem.value) && (
                <Image
                  src={previewItem.value.file_url}
                  style={{ maxWidth: "100%" }}
                  preview={false}
                />
              )}
              {isVideo(previewItem.value) && (
                <video
                  src={previewItem.value.file_url}
                  controls
                  style={{ maxWidth: "100%", maxHeight: "600px" }}
                />
              )}
            </div>
          )}
        </Modal>

        <style>{`
          .resource-grid-item:hover .preview-icon {
            opacity: 1 !important;
          }
        `}</style>
      </div>
    );
  },
});
