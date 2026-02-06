import { defineComponent, ref, onMounted, computed } from "vue";
import { message, Modal } from "ant-design-vue";
import {
  Card,
  Button,
  Input,
  Table,
  Space,
  Tag,
  Popconfirm,
  Form,
  FormItem,
  Select,
  SelectOption,
  Textarea,
  Upload,
  Progress,
} from "ant-design-vue";
import {
  PlusOutlined,
  UploadOutlined,
  HistoryOutlined,
  DeleteOutlined,
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons-vue";
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

export default defineComponent({
  name: "Projects",
  setup() {
    const projects = ref<Project[]>([]);
    const loading = ref(false);
    const total = ref(0);
    const page = ref(1);
    const pageSize = ref(20);
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
    const uploadProgress = ref(0);
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

        // 打印前几个文件的路径信息用于调试
        console.log("=== 文件夹选择调试信息 ===");
        console.log("文件总数:", input.files.length);
        for (let i = 0; i < Math.min(20, input.files.length); i++) {
          const file = input.files[i];
          console.log(`文件 ${i + 1}:`);
          console.log("  name:", file.name);
          console.log(
            "  webkitRelativePath:",
            (file as any).webkitRelativePath,
          );
        }
        console.log("========================");

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
      uploadProgress.value = 0;

      try {
        const formData = new FormData();
        formData.append("username", uploadForm.value.username);
        formData.append("description", uploadForm.value.description);
        formData.append("version_type", uploadForm.value.version_type);

        // 添加所有文件
        console.log("=== 上传文件调试信息 ===");

        // 收集所有文件路径
        const filePaths: string[] = [];

        uploadForm.value.files.forEach((file: File, index: number) => {
          const relativePath = (file as any).webkitRelativePath || file.name;
          filePaths.push(relativePath);

          // 打印前几个文件的信息
          if (index < 5) {
            console.log(`文件 ${index + 1}:`);
            console.log("  name:", file.name);
            console.log(
              "  webkitRelativePath:",
              (file as any).webkitRelativePath,
            );
            console.log("  使用的路径:", relativePath);
          }

          // 只传文件，不传文件名
          formData.append("files", file);
        });

        // 将所有文件路径作为JSON字符串传递
        formData.append("file_paths", JSON.stringify(filePaths));
        console.log("文件路径列表:", filePaths.slice(0, 5));
        console.log("========================");

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
        uploadProgress.value = 0;
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
    const openPreview = (version: ProjectVersion) => {
      if (version.preview_url) {
        window.open(version.preview_url, "_blank");
      } else {
        message.warning("该版本没有预览页面");
      }
    };

    // 表格列定义
    const columns = [
      {
        title: "项目名称",
        dataIndex: "name",
        key: "name",
      },
      {
        title: "描述",
        dataIndex: "description",
        key: "description",
      },
      {
        title: "当前版本",
        dataIndex: "current_version",
        key: "current_version",
        customRender: ({ text }: any) => <Tag color="blue">v{text}</Tag>,
      },
      {
        title: "创建时间",
        dataIndex: "created_at",
        key: "created_at",
        customRender: ({ text }: any) => new Date(text).toLocaleString(),
      },
      {
        title: "操作",
        key: "action",
        customRender: ({ record }: any) => (
          <Space>
            <Button
              type="primary"
              size="small"
              icon={<UploadOutlined />}
              onClick={() => openUploadDialog(record)}
            >
              上传版本
            </Button>
            <Button
              size="small"
              icon={<HistoryOutlined />}
              onClick={() => viewHistory(record)}
            >
              版本历史
            </Button>
            <Popconfirm
              title="确定删除该项目吗？"
              onConfirm={() => handleDelete(record.id)}
            >
              <Button size="small" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          </Space>
        ),
      },
    ];

    // 版本历史表格列
    const versionColumns = [
      {
        title: "版本号",
        dataIndex: "version",
        key: "version",
        customRender: ({ text }: any) => <Tag color="green">v{text}</Tag>,
      },
      {
        title: "上传用户",
        dataIndex: "username",
        key: "username",
      },
      {
        title: "更新描述",
        dataIndex: "description",
        key: "description",
      },
      {
        title: "文件数量",
        dataIndex: "file_count",
        key: "file_count",
      },
      {
        title: "文件大小",
        dataIndex: "file_size",
        key: "file_size",
        customRender: ({ text }: any) => {
          const mb = (text / 1024 / 1024).toFixed(2);
          return `${mb} MB`;
        },
      },
      {
        title: "上传时间",
        dataIndex: "created_at",
        key: "created_at",
        customRender: ({ text }: any) => new Date(text).toLocaleString(),
      },
      {
        title: "操作",
        key: "action",
        customRender: ({ record }: any) => (
          <Space>
            <Button
              type="primary"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => openPreview(record)}
            >
              预览
            </Button>
            <Button
              size="small"
              icon={<DownloadOutlined />}
              onClick={() => window.open(downloadVersion(record.id))}
            >
              下载
            </Button>
            <Popconfirm
              title="确定回滚到此版本吗？"
              onConfirm={() => handleRollback(record.id)}
            >
              <Button size="small">回滚</Button>
            </Popconfirm>
          </Space>
        ),
      },
    ];

    onMounted(() => {
      loadProjects();
    });

    return () => (
      <div style={{ padding: "24px" }}>
        <Card>
          <Space style={{ marginBottom: "16px" }}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => (createDialogVisible.value = true)}
            >
              新建项目
            </Button>
            <Input
              v-model={[keyword.value, "value"]}
              placeholder="搜索项目"
              style={{ width: "300px" }}
              onPressEnter={loadProjects}
            />
            <Button onClick={loadProjects}>搜索</Button>
          </Space>

          <Table
            columns={columns}
            dataSource={projects.value}
            loading={loading.value}
            rowKey="id"
            pagination={{
              current: page.value,
              pageSize: pageSize.value,
              total: total.value,
              onChange: (p, ps) => {
                page.value = p;
                pageSize.value = ps;
                loadProjects();
              },
            }}
          />
        </Card>

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
                webkitdirectory=""
                directory=""
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
            {uploading.value && (
              <FormItem>
                <Progress percent={uploadProgress.value} />
              </FormItem>
            )}
          </Form>
        </Modal>

        {/* 版本历史对话框 */}
        <Modal
          v-model={[historyDialogVisible.value, "visible"]}
          title={`版本历史 - ${currentProject.value?.name}`}
          footer={null}
          width={1000}
        >
          <Table
            columns={versionColumns}
            dataSource={versions.value}
            rowKey="id"
            pagination={false}
          />
        </Modal>
      </div>
    );
  },
});
