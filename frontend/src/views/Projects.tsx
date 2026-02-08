import { defineComponent, ref, onMounted } from "vue";
import { message, Modal, Image, Empty } from "ant-design-vue";
import {
  Button,
  Input,
  Tag,
  Form,
  FormItem,
  Select,
  SelectOption,
  Textarea,
  Popconfirm,
} from "ant-design-vue";
import {
  PlusOutlined,
  UploadOutlined,
  HistoryOutlined,
  DeleteOutlined,
  EyeOutlined,
  DownloadOutlined,
  FolderOutlined,
  ReloadOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import {
  getProjects,
  createProject,
  deleteProject,
  uploadVersion,
  getVersionHistory,
  downloadVersion,
  rollbackVersion,
  type Project,
  type ProjectVersion,
} from "@/api/projects";
import "./Projects.less";

export default defineComponent({
  name: "Projects",
  setup() {
    const projects = ref<Project[]>([]);
    const loading = ref(false);
    const total = ref(0);
    const page = ref(1);
    const pageSize = ref(12);
    const keyword = ref("");

    // 创建项目对话框
    const createDialogVisible = ref(false);
    const createForm = ref({
      name: "",
      description: "",
    });

    // 上传版本对话框
    const uploadDialogVisible = ref(false);
    const uploadForm = ref({
      project_id: 0,
      username: "",
      description: "",
      version_type: "patch" as "major" | "minor" | "patch",
      files: [] as File[],
    });
    const uploading = ref(false);

    // 文件input的引用
    const fileInputRef = ref<HTMLInputElement | null>(null);

    // 版本历史对话框
    const historyDialogVisible = ref(false);
    const versions = ref<ProjectVersion[]>([]);
    const currentProject = ref<Project | null>(null);

    // 加载项目列表
    const loadProjects = async () => {
      loading.value = true;
      try {
        const res = await getProjects({
          page: page.value,
          page_size: pageSize.value,
          keyword: keyword.value,
        });
        projects.value = res.data.data;
        total.value = res.data.total;
      } catch (error) {
        message.error("加载项目列表失败");
      } finally {
        loading.value = false;
      }
    };

    // 创建项目
    const handleCreate = async () => {
      if (!createForm.value.name) {
        message.warning("请输入项目名称");
        return;
      }
      try {
        await createProject(createForm.value);
        message.success("创建成功");
        createDialogVisible.value = false;
        createForm.value = { name: "", description: "" };
        loadProjects();
      } catch (error) {
        message.error("创建失败");
      }
    };

    // 删除项目
    const handleDelete = async (id: number) => {
      try {
        await deleteProject(id);
        message.success("删除成功");
        loadProjects();
      } catch (error) {
        message.error("删除失败");
      }
    };

    // 打开上传对话框
    const openUploadDialog = (project: Project) => {
      currentProject.value = project;
      uploadForm.value = {
        project_id: project.id,
        username: "",
        description: "",
        version_type: "patch",
        files: [],
      };
      uploadDialogVisible.value = true;
    };

    // 文件夹选择
    const handleFolderSelect = (e: Event) => {
      const input = e.target as HTMLInputElement;
      if (input.files && input.files.length > 0) {
        uploadForm.value.files = Array.from(input.files);
        message.success(`已选择 ${input.files.length} 个文件`);
      }
    };

    // 上传版本
    const handleUpload = async () => {
      if (!uploadForm.value.username) {
        message.warning("请输入用户名");
        return;
      }
      if (uploadForm.value.files.length === 0) {
        message.warning("请选择文件夹");
        return;
      }

      uploading.value = true;

      try {
        const formData = new FormData();
        formData.append("username", uploadForm.value.username);
        formData.append("description", uploadForm.value.description);
        formData.append("version_type", uploadForm.value.version_type);

        // 收集所有文件路径
        const filePaths: string[] = [];

        uploadForm.value.files.forEach((file: File) => {
          const relativePath = (file as any).webkitRelativePath || file.name;
          filePaths.push(relativePath);
          formData.append("files", file);
        });

        // 将所有文件路径作为JSON字符串传递
        formData.append("file_paths", JSON.stringify(filePaths));

        await uploadVersion(uploadForm.value.project_id, formData);
        message.success("上传成功");
        uploadDialogVisible.value = false;

        // 清空表单数据
        uploadForm.value = {
          project_id: 0,
          username: "",
          description: "",
          version_type: "patch",
          files: [],
        };

        // 清空文件输入框
        if (fileInputRef.value) {
          fileInputRef.value.value = "";
        }

        loadProjects();
      } catch (error) {
        message.error("上传失败");
        console.error("上传错误:", error);
      } finally {
        uploading.value = false;
      }
    };

    // 查看版本历史
    const viewHistory = async (project: Project) => {
      currentProject.value = project;
      try {
        const res = await getVersionHistory(project.id);
        versions.value = res.data;
        historyDialogVisible.value = true;
      } catch (error) {
        message.error("加载版本历史失败");
      }
    };

    // 回滚版本
    const handleRollback = async (versionId: number) => {
      try {
        await rollbackVersion(versionId);
        message.success("回滚成功");
        historyDialogVisible.value = false;
        loadProjects();
      } catch (error) {
        message.error("回滚失败");
      }
    };

    // 打开预览
    const openPreview = (project: Project) => {
      // 如果有缩略图URL，从中提取 baseURL
      // 否则使用默认的后端地址
      let baseURL = "http://192.168.3.39:23359";

      if (project.thumbnail_url) {
        // 从 thumbnail_url 中提取 baseURL
        // 例如: http://192.168.3.39:23359/projects/123/v1.0.6/thumbnail.png
        const url = new URL(project.thumbnail_url);
        baseURL = `${url.protocol}//${url.host}`;
      }

      // 构建项目根路径，后端会自动重定向到最新版本
      const projectURL = `${baseURL}/projects/${project.name}/`;
      window.open(projectURL, "_blank");
    };

    // 点击卡片打开项目
    const handleCardClick = (project: Project) => {
      openPreview(project);
    };

    // 搜索
    const handleSearch = (value: string) => {
      keyword.value = value;
      page.value = 1;
      loadProjects();
    };

    // 分页
    const handlePageChange = (newPage: number, newPageSize: number) => {
      page.value = newPage;
      pageSize.value = newPageSize;
      loadProjects();
    };

    // 分页大小变化
    const handlePageSizeChange = (size: number) => {
      pageSize.value = size;
      page.value = 1;
      loadProjects();
    };

    onMounted(() => {
      loadProjects();
    });

    return () => (
      <div
        style={{ padding: "24px", minHeight: "100vh", background: "#f5f5f5" }}
      >
        {/* 头部 */}
        <ResourceHeader
          stats={[
            {
              icon: FolderOutlined,
              label: "总项目数",
              value: total.value,
              color: "#52c41a",
            },
          ]}
          actions={[
            {
              label: "刷新",
              icon: ReloadOutlined,
              loading: loading.value,
              onClick: loadProjects,
            },
            {
              label: "新建项目",
              icon: PlusOutlined,
              type: "primary",
              onClick: () => (createDialogVisible.value = true),
            },
          ]}
          onSearch={handleSearch}
          searchPlaceholder="搜索项目名称"
          pageSize={pageSize.value}
          onPageSizeChange={handlePageSizeChange}
        />

        {/* 项目网格 */}
        <ResourceGrid
          loading={loading.value}
          data={projects.value}
          total={total.value}
          currentPage={page.value}
          pageSize={pageSize.value}
          onPageChange={handlePageChange}
          onCardClick={handleCardClick}
          renderPreview={(project) => {
            if (project.thumbnail_url) {
              return (
                <Image
                  src={project.thumbnail_url}
                  width="100%"
                  height="100%"
                  style={{ objectFit: "cover" }}
                  preview={{ src: project.thumbnail_url }}
                />
              );
            }
            return (
              <div class="preview-placeholder">
                <FolderOutlined />
                <div style={{ marginTop: "8px", fontSize: "12px" }}>
                  点击打开项目
                </div>
              </div>
            );
          }}
          renderContent={(project) => (
            <>
              <div class="resource-name" title={project.name}>
                {project.name}
              </div>
              <div
                style={{
                  fontSize: "12px",
                  color: "#999",
                  marginTop: "4px",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
                title={project.description}
              >
                {project.description || "暂无描述"}
              </div>
              <div
                style={{
                  display: "flex",
                  gap: "8px",
                  alignItems: "center",
                  marginTop: "8px",
                }}
              >
                <Tag color="blue">v{project.current_version}</Tag>
                <span style={{ fontSize: "12px", color: "#666" }}>
                  {new Date(project.updated_at).toLocaleDateString()}
                </span>
              </div>
              <div style={{ display: "flex", gap: "8px", marginTop: "12px" }}>
                <Button
                  type="primary"
                  size="small"
                  icon={<UploadOutlined />}
                  onClick={(e: Event) => {
                    e.stopPropagation();
                    openUploadDialog(project);
                  }}
                >
                  上传
                </Button>
                <Button
                  size="small"
                  icon={<HistoryOutlined />}
                  onClick={(e: Event) => {
                    e.stopPropagation();
                    viewHistory(project);
                  }}
                >
                  历史
                </Button>
                <Popconfirm
                  title="确定删除该项目吗？"
                  onConfirm={() => handleDelete(project.id)}
                >
                  <Button
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={(e: Event) => e.stopPropagation()}
                  />
                </Popconfirm>
              </div>
            </>
          )}
        />

        {/* 版本历史对话框 */}
        <Modal
          v-model={[historyDialogVisible.value, "visible"]}
          title={`版本历史 - ${currentProject.value?.name}`}
          footer={null}
          width={1000}
        >
          {versions.value.length === 0 ? (
            <Empty description="暂无版本历史" />
          ) : (
            <div class="version-list">
              {versions.value.map((version: ProjectVersion) => (
                <div key={version.id} class="version-item">
                  <div class="version-header">
                    <Tag color="green">v{version.version}</Tag>
                    <span class="version-user">{version.username}</span>
                    <span class="version-time">
                      {new Date(version.created_at).toLocaleString()}
                    </span>
                  </div>

                  {/* 缩略图预览 */}
                  {version.thumbnail_url && (
                    <div class="version-thumbnail">
                      <img
                        src={version.thumbnail_url}
                        alt="预览"
                        onClick={() =>
                          window.open(version.preview_url, "_blank")
                        }
                      />
                    </div>
                  )}

                  <div class="version-description">
                    {version.description || "无更新描述"}
                  </div>
                  <div class="version-meta">
                    <span>文件数: {version.file_count}</span>
                    <span>
                      大小: {(version.file_size / 1024 / 1024).toFixed(2)} MB
                    </span>
                  </div>
                  <div class="version-actions">
                    <Button
                      type="primary"
                      size="small"
                      icon={<EyeOutlined />}
                      onClick={() => {
                        if (version.preview_url) {
                          window.open(version.preview_url, "_blank");
                        } else {
                          message.warning("该版本没有预览页面");
                        }
                      }}
                    >
                      预览
                    </Button>
                    <Button
                      size="small"
                      icon={<DownloadOutlined />}
                      onClick={() => window.open(downloadVersion(version.id))}
                    >
                      下载
                    </Button>
                    <Popconfirm
                      title="确定回滚到此版本吗？"
                      onConfirm={() => handleRollback(version.id)}
                    >
                      <Button size="small">回滚</Button>
                    </Popconfirm>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Modal>

        {/* 创建项目对话框 */}
        <Modal
          v-model={[createDialogVisible.value, "visible"]}
          title="新建项目"
          onOk={handleCreate}
        >
          <Form layout="vertical">
            <FormItem label="项目名称" required>
              <Input
                v-model={[createForm.value.name, "value"]}
                placeholder="请输入项目名称"
              />
            </FormItem>
            <FormItem label="项目描述">
              <Textarea
                v-model={[createForm.value.description, "value"]}
                placeholder="请输入项目描述"
                rows={4}
              />
            </FormItem>
          </Form>
        </Modal>

        {/* 上传版本对话框 */}
        <Modal
          v-model={[uploadDialogVisible.value, "visible"]}
          title="上传版本"
          onOk={handleUpload}
          confirmLoading={uploading.value}
          width={600}
        >
          <Form layout="vertical">
            <FormItem label="当前版本">
              <Tag color="blue">v{currentProject.value?.current_version}</Tag>
            </FormItem>
            <FormItem label="版本类型" required>
              <Select v-model={[uploadForm.value.version_type, "value"]}>
                <SelectOption value="major">主版本 (x.0.0)</SelectOption>
                <SelectOption value="minor">次版本 (0.x.0)</SelectOption>
                <SelectOption value="patch">补丁版本 (0.0.x)</SelectOption>
              </Select>
            </FormItem>
            <FormItem label="上传用户" required>
              <Input
                v-model={[uploadForm.value.username, "value"]}
                placeholder="请输入用户名"
              />
            </FormItem>
            <FormItem label="更新描述">
              <Textarea
                v-model={[uploadForm.value.description, "value"]}
                placeholder="请输入更新描述"
                rows={3}
              />
            </FormItem>
            <FormItem label="选择项目文件夹" required>
              <input
                ref={fileInputRef}
                type="file"
                {...{ webkitdirectory: "", directory: "" }}
                multiple
                onChange={handleFolderSelect}
                style={{ display: "block" }}
              />
              {uploadForm.value.files.length > 0 && (
                <div style={{ marginTop: "8px" }}>
                  已选择 {uploadForm.value.files.length} 个文件
                </div>
              )}
            </FormItem>
          </Form>
        </Modal>
      </div>
    );
  },
});
