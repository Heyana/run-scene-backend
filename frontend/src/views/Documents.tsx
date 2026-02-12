import { defineComponent, ref, onMounted, computed, watch } from "vue";
import {
  message,
  Modal,
  Image,
  Breadcrumb,
  BreadcrumbItem,
} from "ant-design-vue";
import {
  Button,
  Input,
  Tag,
  Form,
  FormItem,
  Textarea,
  Popconfirm,
  Upload,
} from "ant-design-vue";
import {
  UploadOutlined,
  DeleteOutlined,
  DownloadOutlined,
  FolderOutlined,
  FileOutlined,
  ReloadOutlined,
  FolderAddOutlined,
  ArrowLeftOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons-vue";
import ResourceHeader from "@/components/ResourceHeader";
import ResourceGrid from "@/components/ResourceGrid";
import {
  getDocuments,
  uploadDocument,
  uploadFolder,
  deleteDocument,
  createFolder,
} from "@/api/documents";
import type { Document } from "@/api/documents";
import { showContextMenu } from "@/utils/context-menu";
import type { MenuItem } from "@/utils/context-menu";
import "./Documents.less";
import { useRoute, useRouter } from "vue-router";

export default defineComponent({
  name: "Documents",
  setup() {
    const route = useRoute();
    const router = useRouter();

    const documents = ref<Document[]>([]);
    const loading = ref(false);
    const total = ref(0);
    const page = ref(1);
    const pageSize = ref(24);
    const keyword = ref("");
    const currentFolderId = ref<number>(0); // 0 表示根目录
    const breadcrumb = ref<Array<{ id: number; name: string }>>([
      { id: 0, name: "文件库" },
    ]);

    // 创建文件夹对话框
    const createFolderVisible = ref(false);
    const folderForm = ref({
      name: "",
      description: "",
    });

    // 上传文件对话框
    const uploadDialogVisible = ref(false);
    const uploadForm = ref({
      name: "",
      description: "",
      category: "",
      file: null as File | null,
    });
    const fileList = ref<any[]>([]);
    const uploadProgress = ref(0);
    const isUploading = ref(false);

    // 上传文件夹对话框
    const uploadFolderDialogVisible = ref(false);
    const uploadFolderForm = ref({
      description: "",
      category: "",
      files: [] as File[],
    });
    const folderFileList = ref<any[]>([]);

    // 预览对话框
    const previewVisible = ref(false);
    const previewUrl = ref("");
    const previewType = ref("");

    // 拖拽上传
    const isDragging = ref(false);
    const dragCounter = ref(0);
    const dragUploadProgress = ref(0);
    const isDragUploading = ref(false);
    const dragUploadStatus = ref("");

    // 统计信息
    const folderCount = computed(() => {
      return documents.value.filter((d) => d.is_folder).length;
    });

    const fileCount = computed(() => {
      return documents.value.filter((d) => !d.is_folder).length;
    });

    // 从 URL 读取文件夹 ID
    const initFromRoute = async () => {
      const folderId = route.query.folder;
      if (folderId) {
        const id = parseInt(folderId as string, 10);
        if (!isNaN(id) && id !== 0) {
          currentFolderId.value = id;
          // 需要加载面包屑路径
          await loadBreadcrumbPath(id);
        }
      }
    };

    // 加载面包屑路径（简化版：通过 parent_id 递归查询）
    const loadBreadcrumbPath = async (folderId: number) => {
      if (folderId === 0) {
        breadcrumb.value = [{ id: 0, name: "文件库" }];
        return;
      }

      try {
        const path: Array<{ id: number; name: string }> = [];
        let currentId: number | null = folderId;

        // 递归查找父文件夹
        while (currentId !== null && currentId !== 0) {
          // 获取当前文件夹的详情
          try {
            const res = await getDocuments({
              page: 1,
              pageSize: 1000,
            });

            const allDocs = res.data.list || [];
            const folder = allDocs.find(
              (d: Document) => d.id === currentId && d.is_folder,
            );

            if (folder) {
              path.unshift({ id: folder.id, name: folder.name });
              currentId = folder.parent_id || 0;
            } else {
              // 找不到文件夹，可能已被删除
              console.warn(`文件夹 ${currentId} 不存在`);
              break;
            }
          } catch (error) {
            console.error("查询文件夹失败:", error);
            break;
          }
        }

        if (path.length > 0) {
          breadcrumb.value = [{ id: 0, name: "文件库" }, ...path];
        } else {
          // 如果找不到路径，回到根目录
          breadcrumb.value = [{ id: 0, name: "文件库" }];
          currentFolderId.value = 0;
          updateRoute(0);
        }
      } catch (error) {
        console.error("加载面包屑路径失败:", error);
        breadcrumb.value = [{ id: 0, name: "文件库" }];
        currentFolderId.value = 0;
        updateRoute(0);
      }
    };

    // 更新 URL
    const updateRoute = (folderId: number) => {
      if (folderId === 0) {
        router.replace({ query: {} });
      } else {
        router.replace({ query: { folder: String(folderId) } });
      }
    };

    // 加载当前目录内容
    const loadCurrentFolder = async () => {
      loading.value = true;
      try {
        const res = await getDocuments({
          page: page.value,
          pageSize: pageSize.value,
          keyword: keyword.value,
          parent_id: currentFolderId.value,
        });
        documents.value = res.data.list || [];
        total.value = res.data.total || 0;
      } catch (error) {
        message.error("加载失败");
      } finally {
        loading.value = false;
      }
    };

    // 创建文件夹
    const handleCreateFolder = async () => {
      if (!folderForm.value.name) {
        message.warning("请输入文件夹名称");
        return;
      }

      try {
        await createFolder({
          name: folderForm.value.name,
          description: folderForm.value.description,
          parent_id:
            currentFolderId.value === 0 ? undefined : currentFolderId.value,
        });
        message.success("创建成功");
        createFolderVisible.value = false;
        folderForm.value = { name: "", description: "" };
        loadCurrentFolder();
      } catch (error) {
        message.error("创建失败");
      }
    };

    // 上传文件
    const handleUpload = async () => {
      if (!uploadForm.value.file) {
        message.warning("请选择文件");
        return;
      }

      try {
        isUploading.value = true;
        uploadProgress.value = 0;

        const formData = new FormData();
        formData.append("file", uploadForm.value.file);
        formData.append(
          "name",
          uploadForm.value.name || uploadForm.value.file.name,
        );
        formData.append("description", uploadForm.value.description);
        formData.append("category", uploadForm.value.category);

        // 如果在子文件夹中，添加 parent_id
        if (currentFolderId.value !== 0) {
          formData.append("parent_id", String(currentFolderId.value));
        }

        await uploadDocument(formData, {
          onUploadProgress: (progressEvent: any) => {
            if (progressEvent.total) {
              uploadProgress.value = Math.round(
                (progressEvent.loaded * 100) / progressEvent.total,
              );
            }
          },
        });

        message.success("上传成功");
        uploadDialogVisible.value = false;
        uploadForm.value = {
          name: "",
          description: "",
          category: "",
          file: null,
        };
        fileList.value = [];
        uploadProgress.value = 0;
        loadCurrentFolder();
      } catch (error: any) {
        if (error.response?.data?.msg) {
          message.error(error.response.data.msg);
        } else {
          message.error("上传失败");
        }
      } finally {
        isUploading.value = false;
      }
    };

    // 文件选择
    const handleFileChange = (info: any) => {
      fileList.value = info.fileList.slice(-1);
      if (info.fileList.length > 0) {
        const file = info.fileList[0];
        uploadForm.value.file = file.originFileObj || file;
        if (!uploadForm.value.name) {
          const fileName = file.name.replace(/\.[^/.]+$/, "");
          uploadForm.value.name = fileName;
        }
      }
    };

    // 文件夹选择
    const handleFolderChange = (e: Event) => {
      const input = e.target as HTMLInputElement;
      if (input.files && input.files.length > 0) {
        uploadFolderForm.value.files = Array.from(input.files);
        folderFileList.value = Array.from(input.files).map((file, index) => ({
          uid: index,
          name: file.webkitRelativePath || file.name,
          status: "done",
          originFileObj: file,
        }));
      }
    };

    // 上传文件夹
    const handleUploadFolder = async () => {
      if (uploadFolderForm.value.files.length === 0) {
        message.warning("请选择文件夹");
        return;
      }

      try {
        const formData = new FormData();

        // 添加所有文件
        uploadFolderForm.value.files.forEach((file) => {
          formData.append("files", file);
        });

        // 添加文件路径列表
        const filePaths = uploadFolderForm.value.files.map(
          (file: any) => file.webkitRelativePath || file.name,
        );
        formData.append("file_paths", JSON.stringify(filePaths));

        // 添加元数据
        formData.append("description", uploadFolderForm.value.description);
        formData.append("category", uploadFolderForm.value.category);

        // 如果在子文件夹中，添加 parent_id
        if (currentFolderId.value !== 0) {
          formData.append("parent_id", String(currentFolderId.value));
        }

        await uploadFolder(formData);
        message.success("上传成功");
        uploadFolderDialogVisible.value = false;
        uploadFolderForm.value = {
          description: "",
          category: "",
          files: [],
        };
        folderFileList.value = [];
        loadCurrentFolder();
      } catch (error: any) {
        if (error.response?.data?.msg) {
          message.error(error.response.data.msg);
        } else {
          message.error("上传失败");
        }
      }
    };

    // 拖拽上传处理
    const handleDragEnter = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter.value++;
      if (dragCounter.value === 1) {
        isDragging.value = true;
      }
    };

    const handleDragLeave = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter.value--;
      if (dragCounter.value === 0) {
        isDragging.value = false;
      }
    };

    const handleDragOver = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
    };

    const handleDrop = async (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      isDragging.value = false;
      dragCounter.value = 0;

      const items = e.dataTransfer?.items;
      if (!items || items.length === 0) return;

      // 检查是否是文件夹
      const firstItem = items[0];
      const entry = firstItem.webkitGetAsEntry?.();

      if (entry?.isDirectory) {
        // 上传文件夹
        const files: File[] = [];
        const filePaths: string[] = [];

        const traverseDirectory = async (
          dirEntry: any,
          path: string = "",
        ): Promise<void> => {
          const reader = dirEntry.createReader();
          const entries = await new Promise<any[]>((resolve) => {
            reader.readEntries((entries: any[]) => resolve(entries));
          });

          for (const entry of entries) {
            const fullPath = path ? `${path}/${entry.name}` : entry.name;

            if (entry.isFile) {
              const file = await new Promise<File>((resolve) => {
                entry.file((file: File) => resolve(file));
              });
              files.push(file);
              filePaths.push(fullPath);
            } else if (entry.isDirectory) {
              await traverseDirectory(entry, fullPath);
            }
          }
        };

        try {
          isDragUploading.value = true;
          dragUploadProgress.value = 0;
          dragUploadStatus.value = "正在读取文件夹...";

          await traverseDirectory(entry, entry.name);

          if (files.length === 0) {
            message.warning("文件夹为空");
            isDragUploading.value = false;
            return;
          }

          dragUploadStatus.value = `正在上传 ${files.length} 个文件...`;

          const formData = new FormData();
          files.forEach((file) => formData.append("files", file));
          formData.append("file_paths", JSON.stringify(filePaths));

          if (currentFolderId.value !== 0) {
            formData.append("parent_id", String(currentFolderId.value));
          }

          await uploadFolder(formData, {
            onUploadProgress: (progressEvent: any) => {
              if (progressEvent.total) {
                dragUploadProgress.value = Math.round(
                  (progressEvent.loaded * 100) / progressEvent.total,
                );
              }
            },
          });

          message.success("上传成功");
          loadCurrentFolder();
        } catch (error: any) {
          message.error(error.response?.data?.msg || "上传失败");
        } finally {
          isDragUploading.value = false;
          dragUploadProgress.value = 0;
          dragUploadStatus.value = "";
        }
      } else {
        // 上传单个或多个文件
        const files: File[] = [];
        for (let i = 0; i < items.length; i++) {
          const file = items[i].getAsFile();
          if (file) files.push(file);
        }

        if (files.length === 0) return;

        try {
          isDragUploading.value = true;
          dragUploadProgress.value = 0;
          dragUploadStatus.value = `正在上传 ${files.length} 个文件...`;

          let completed = 0;
          for (const file of files) {
            const formData = new FormData();
            formData.append("file", file);
            formData.append("name", file.name.replace(/\.[^/.]+$/, ""));

            if (currentFolderId.value !== 0) {
              formData.append("parent_id", String(currentFolderId.value));
            }

            await uploadDocument(formData, {
              onUploadProgress: (progressEvent: any) => {
                if (progressEvent.total) {
                  const fileProgress = Math.round(
                    (progressEvent.loaded * 100) / progressEvent.total,
                  );
                  dragUploadProgress.value = Math.round(
                    ((completed + fileProgress / 100) * 100) / files.length,
                  );
                }
              },
            });

            completed++;
            dragUploadProgress.value = Math.round(
              (completed * 100) / files.length,
            );
          }

          message.success("上传成功");
          loadCurrentFolder();
        } catch (error: any) {
          message.error(error.response?.data?.msg || "上传失败");
        } finally {
          isDragUploading.value = false;
          dragUploadProgress.value = 0;
          dragUploadStatus.value = "";
        }
      }
    };

    // 删除项目（文件或文件夹）
    const handleDelete = async (item: Document) => {
      if (item.is_folder && item.child_count > 0) {
        Modal.confirm({
          title: "删除确认",
          content: `该文件夹包含 ${item.child_count} 个子项，是否级联删除？`,
          okText: "级联删除",
          cancelText: "取消",
          onOk: async () => {
            try {
              await deleteDocument(item.id);
              message.success("删除成功");
              loadCurrentFolder();
            } catch (error) {
              message.error("删除失败");
            }
          },
        });
      } else {
        try {
          await deleteDocument(item.id);
          message.success("删除成功");
          loadCurrentFolder();
        } catch (error) {
          message.error("删除失败");
        }
      }
    };

    // 打开文件夹
    const handleOpenFolder = (folder: Document) => {
      currentFolderId.value = folder.id;
      breadcrumb.value.push({ id: folder.id, name: folder.name });
      page.value = 1;
      updateRoute(folder.id);
      loadCurrentFolder();
    };

    // 面包屑导航
    const handleBreadcrumbClick = (index: number) => {
      if (index === breadcrumb.value.length - 1) return;

      breadcrumb.value = breadcrumb.value.slice(0, index + 1);
      currentFolderId.value = breadcrumb.value[index].id;
      page.value = 1;
      updateRoute(currentFolderId.value);
      loadCurrentFolder();
    };

    // 返回上一级
    const handleGoBack = () => {
      if (breadcrumb.value.length > 1) {
        breadcrumb.value.pop();
        currentFolderId.value =
          breadcrumb.value[breadcrumb.value.length - 1].id;
        page.value = 1;
        updateRoute(currentFolderId.value);
        loadCurrentFolder();
      }
    };

    // 预览文件
    const handlePreview = (doc: Document) => {
      const imageFormats = ["jpg", "jpeg", "png", "gif", "webp", "bmp"];
      const videoFormats = ["mp4", "webm", "avi", "mov"];

      if (imageFormats.includes(doc.format?.toLowerCase())) {
        previewType.value = "image";
        previewUrl.value = doc.file_url;
        previewVisible.value = true;
      } else if (videoFormats.includes(doc.format?.toLowerCase())) {
        previewType.value = "video";
        previewUrl.value = doc.file_url;
        previewVisible.value = true;
      } else {
        window.open(doc.file_url, "_blank");
      }
    };

    // 点击卡片
    const handleCardClick = (item: Document) => {
      if (item.is_folder) {
        handleOpenFolder(item);
      } else {
        handlePreview(item);
      }
    };

    // 右键菜单
    const handleContextMenu = (e: MouseEvent, item: Document) => {
      e.preventDefault();
      e.stopPropagation();

      const menuItems: MenuItem[] = [
        {
          label: "预览",
          key: "preview",
          icon: "",
          disabled: item.is_folder,
          onClick: async () => {
            // 模拟异步操作
            await new Promise((resolve) => setTimeout(resolve, 500));
            handlePreview(item);
          },
        },
        {
          label: "刷新",
          key: "refresh",
          icon: "",
          onClick: async () => {
            await loadCurrentFolder();
          },
        },
        {
          label: "重新截图",
          key: "screenshot",
          icon: "",
          disabled: item.is_folder,
          onClick: async () => {
            // 模拟异步截图操作
            await new Promise((resolve) => setTimeout(resolve, 1000));
            message.info("截图功能开发中...");
          },
        },
        {
          label: "更多操作",
          key: "more",
          icon: "",
          children: [
            {
              label: "重命名",
              key: "rename",
              icon: "",
              onClick: () => {
                message.info("重命名功能开发中...");
              },
            },
            {
              label: "复制",
              key: "copy",
              icon: "",
              onClick: () => {
                message.info("复制功能开发中...");
              },
            },
            {
              label: "移动",
              key: "move",
              icon: "",
              onClick: () => {
                message.info("移动功能开发中...");
              },
            },
          ],
        },
        {
          label: "删除",
          key: "delete",
          icon: "",
          type: "delete",
          divided: true,
          onClick: async () => {
            await handleDelete(item);
          },
        },
      ];

      showContextMenu(e, menuItems);
    };

    // 搜索
    const handleSearch = (value: string) => {
      keyword.value = value;
      page.value = 1;
      loadCurrentFolder();
    };

    // 分页
    const handlePageChange = (newPage: number, newPageSize: number) => {
      page.value = newPage;
      pageSize.value = newPageSize;
      loadCurrentFolder();
    };

    onMounted(() => {
      initFromRoute();
      loadCurrentFolder();
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
              label: "文件夹",
              value: folderCount.value,
              color: "#1890ff",
            },
            {
              icon: FileOutlined,
              label: "文件",
              value: fileCount.value,
              color: "#52c41a",
            },
          ]}
          actions={[
            ...(breadcrumb.value.length > 1
              ? [
                  {
                    label: "返回上级",
                    icon: ArrowLeftOutlined,
                    onClick: handleGoBack,
                  },
                ]
              : []),
            {
              label: "刷新",
              icon: ReloadOutlined,
              loading: loading.value,
              onClick: loadCurrentFolder,
            },
            {
              label: "新建文件夹",
              icon: FolderAddOutlined,
              onClick: () => (createFolderVisible.value = true),
            },
            {
              label: "上传文件",
              icon: UploadOutlined,
              type: "primary",
              onClick: () => (uploadDialogVisible.value = true),
            },
            {
              label: "上传文件夹",
              icon: FolderAddOutlined,
              onClick: () => (uploadFolderDialogVisible.value = true),
            },
          ]}
          onSearch={handleSearch}
          searchPlaceholder="搜索文件名称"
          pageSize={pageSize.value}
          onPageSizeChange={(size) => {
            pageSize.value = size;
            page.value = 1;
            loadCurrentFolder();
          }}
        />

        {/* 面包屑导航 */}
        <div
          style={{
            background: "#fff",
            padding: "12px 24px",
            marginBottom: "16px",
            borderRadius: "8px",
          }}
        >
          <Breadcrumb>
            {breadcrumb.value.map((item, index) => (
              <BreadcrumbItem
                key={index}
                style={{
                  cursor:
                    index < breadcrumb.value.length - 1 ? "pointer" : "default",
                }}
                onClick={() => handleBreadcrumbClick(index)}
              >
                {item.name}
              </BreadcrumbItem>
            ))}
          </Breadcrumb>
        </div>

        {/* 拖拽上传区域 */}
        <div
          class={`drag-upload-area ${isDragging.value ? "dragging" : ""} ${isDragUploading.value ? "uploading" : ""}`}
          onDragenter={handleDragEnter}
          onDragleave={handleDragLeave}
          onDragover={handleDragOver}
          onDrop={handleDrop}
        >
          {isDragUploading.value ? (
            <>
              <div style={{ width: "100%", maxWidth: "400px" }}>
                <div
                  style={{
                    fontSize: "16px",
                    fontWeight: 500,
                    marginBottom: "12px",
                    textAlign: "center",
                  }}
                >
                  {dragUploadStatus.value}
                </div>
                <div
                  style={{ display: "flex", alignItems: "center", gap: "12px" }}
                >
                  <div
                    style={{
                      flex: 1,
                      height: "24px",
                      background: "#f0f0f0",
                      borderRadius: "12px",
                      overflow: "hidden",
                    }}
                  >
                    <div
                      style={{
                        width: `${dragUploadProgress.value}%`,
                        height: "100%",
                        background: "linear-gradient(90deg, #1890ff, #40a9ff)",
                        transition: "width 0.3s ease",
                      }}
                    />
                  </div>
                  <span
                    style={{
                      minWidth: "50px",
                      textAlign: "right",
                      fontSize: "16px",
                      fontWeight: 500,
                    }}
                  >
                    {dragUploadProgress.value}%
                  </span>
                </div>
              </div>
            </>
          ) : (
            <>
              <UploadOutlined style={{ fontSize: "32px", color: "#1890ff" }} />
              <div
                style={{ marginTop: "8px", fontSize: "16px", fontWeight: 500 }}
              >
                {isDragging.value
                  ? "松开鼠标上传"
                  : "拖拽文件或文件夹到此处上传"}
              </div>
              <div
                style={{ marginTop: "4px", fontSize: "12px", color: "#999" }}
              >
                支持单个文件、多个文件或整个文件夹（最大 10GB）
              </div>
            </>
          )}
        </div>

        {/* 文件和文件夹网格 */}
        <ResourceGrid
          loading={loading.value}
          data={documents.value}
          total={total.value}
          currentPage={page.value}
          pageSize={pageSize.value}
          onPageChange={handlePageChange}
          onCardClick={handleCardClick}
          onContextMenu={handleContextMenu}
          renderPreview={(item: Document) => {
            if (item.is_folder) {
              // 文件夹：显示前4个文件的缩略图网格，或默认图标
              if (item.folder_thumbnails && item.folder_thumbnails.length > 0) {
                const thumbs = item.folder_thumbnails.slice(0, 4);
                return (
                  <div class="folder-thumbnail-grid">
                    {thumbs.map((thumb, index) => (
                      <div key={index} class="folder-thumbnail-item">
                        <Image
                          src={thumb}
                          width="100%"
                          height="100%"
                          style={{ objectFit: "cover" }}
                          preview={false}
                        />
                      </div>
                    ))}
                    {/* 填充空白格子 */}
                    {Array.from({ length: 4 - thumbs.length }).map(
                      (_, index) => (
                        <div
                          key={`empty-${index}`}
                          class="folder-thumbnail-item folder-thumbnail-empty"
                        >
                          <FileOutlined
                            style={{ fontSize: "20px", color: "#d9d9d9" }}
                          />
                        </div>
                      ),
                    )}
                  </div>
                );
              } else {
                // 空文件夹或无缩略图：显示默认图标
                return (
                  <div class="preview-placeholder">
                    <FolderOutlined
                      style={{ fontSize: "48px", color: "#1890ff" }}
                    />
                    <div style={{ marginTop: "8px", fontSize: "12px" }}>
                      {item.child_count || 0} 个子项
                    </div>
                  </div>
                );
              }
            } else {
              const videoFormats = ["mp4", "webm", "avi", "mov"];

              if (item.thumbnail_url) {
                return (
                  <Image
                    src={item.thumbnail_url}
                    width="100%"
                    height="100%"
                    style={{ objectFit: "cover" }}
                    preview={false}
                  />
                );
              } else if (videoFormats.includes(item.format?.toLowerCase())) {
                return (
                  <div class="preview-placeholder">
                    <PlayCircleOutlined
                      style={{ fontSize: "48px", color: "#52c41a" }}
                    />
                    <div style={{ marginTop: "8px", fontSize: "12px" }}>
                      点击播放
                    </div>
                  </div>
                );
              } else {
                return (
                  <div class="preview-placeholder">
                    <FileOutlined style={{ fontSize: "48px", color: "#999" }} />
                    <div style={{ marginTop: "8px", fontSize: "12px" }}>
                      {item.format?.toUpperCase()}
                    </div>
                  </div>
                );
              }
            }
          }}
          renderContent={(item: Document) => (
            <>
              <div class="resource-name" title={item.name}>
                {item.is_folder && (
                  <FolderOutlined style={{ marginRight: "4px" }} />
                )}
                {item.name}
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
                title={item.description}
              >
                {item.description || "暂无描述"}
              </div>
              {!item.is_folder && (
                <div style={{ marginTop: "8px" }}>
                  <Tag color="blue">{item.format?.toUpperCase()}</Tag>
                  <span style={{ fontSize: "12px", color: "#666" }}>
                    {((item.file_size || 0) / 1024 / 1024).toFixed(2)} MB
                  </span>
                </div>
              )}
              <div
                style={{ fontSize: "12px", color: "#999", marginTop: "4px" }}
              >
                {new Date(item.created_at).toLocaleDateString()}
              </div>
              <div style={{ display: "flex", gap: "8px", marginTop: "12px" }}>
                {!item.is_folder && (
                  <Button
                    size="small"
                    icon={<DownloadOutlined />}
                    onClick={(e: Event) => {
                      e.stopPropagation();
                      window.open(item.file_url);
                    }}
                  >
                    下载
                  </Button>
                )}
                <Popconfirm
                  title={`确定删除该${item.is_folder ? "文件夹" : "文件"}吗？`}
                  onConfirm={() => handleDelete(item)}
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

        {/* 创建文件夹对话框 */}
        <Modal
          v-model={[createFolderVisible.value, "visible"]}
          title="新建文件夹"
          onOk={handleCreateFolder}
        >
          <Form layout="vertical">
            <FormItem label="文件夹名称" required>
              <Input
                v-model={[folderForm.value.name, "value"]}
                placeholder="请输入文件夹名称"
              />
            </FormItem>
            <FormItem label="描述">
              <Textarea
                v-model={[folderForm.value.description, "value"]}
                placeholder="请输入描述"
                rows={4}
              />
            </FormItem>
          </Form>
        </Modal>

        {/* 上传文件对话框 */}
        <Modal
          v-model={[uploadDialogVisible.value, "visible"]}
          title="上传文件"
          onOk={handleUpload}
          confirmLoading={isUploading.value}
          closable={!isUploading.value}
          maskClosable={!isUploading.value}
        >
          <Form layout="vertical">
            <FormItem label="选择文件" required>
              <Upload
                v-model:file-list={fileList.value}
                beforeUpload={() => false}
                onChange={handleFileChange}
                maxCount={1}
                disabled={isUploading.value}
              >
                <Button icon={<UploadOutlined />} disabled={isUploading.value}>
                  选择文件
                </Button>
              </Upload>
            </FormItem>
            <FormItem label="文件名称">
              <Input
                v-model={[uploadForm.value.name, "value"]}
                placeholder="留空则使用原文件名"
                disabled={isUploading.value}
              />
            </FormItem>
            <FormItem label="分类">
              <Input
                v-model={[uploadForm.value.category, "value"]}
                placeholder="请输入分类"
                disabled={isUploading.value}
              />
            </FormItem>
            <FormItem label="描述">
              <Textarea
                v-model={[uploadForm.value.description, "value"]}
                placeholder="请输入描述"
                rows={3}
                disabled={isUploading.value}
              />
            </FormItem>
            {isUploading.value && (
              <FormItem label="上传进度">
                <div
                  style={{ display: "flex", alignItems: "center", gap: "12px" }}
                >
                  <div
                    style={{
                      flex: 1,
                      height: "20px",
                      background: "#f0f0f0",
                      borderRadius: "10px",
                      overflow: "hidden",
                    }}
                  >
                    <div
                      style={{
                        width: `${uploadProgress.value}%`,
                        height: "100%",
                        background: "#1890ff",
                        transition: "width 0.3s ease",
                      }}
                    />
                  </div>
                  <span style={{ minWidth: "45px", textAlign: "right" }}>
                    {uploadProgress.value}%
                  </span>
                </div>
              </FormItem>
            )}
          </Form>
        </Modal>

        {/* 预览对话框 */}
        <Modal
          v-model={[previewVisible.value, "visible"]}
          title="预览"
          footer={null}
          width={800}
        >
          {previewType.value === "image" && (
            <img src={previewUrl.value} style={{ width: "100%" }} />
          )}
          {previewType.value === "video" && (
            <video src={previewUrl.value} controls style={{ width: "100%" }} />
          )}
        </Modal>

        {/* 上传文件夹对话框 */}
        <Modal
          v-model={[uploadFolderDialogVisible.value, "visible"]}
          title="上传文件夹"
          onOk={handleUploadFolder}
          width={600}
        >
          <Form layout="vertical">
            <FormItem label="选择文件夹" required>
              <input
                type="file"
                webkitdirectory=""
                directory=""
                multiple
                onChange={handleFolderChange}
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #d9d9d9",
                  borderRadius: "4px",
                }}
              />
              {folderFileList.value.length > 0 && (
                <div style={{ marginTop: "8px", color: "#666" }}>
                  已选择 {folderFileList.value.length} 个文件
                </div>
              )}
            </FormItem>
            <FormItem label="分类">
              <Input
                v-model={[uploadFolderForm.value.category, "value"]}
                placeholder="请输入分类"
              />
            </FormItem>
            <FormItem label="描述">
              <Textarea
                v-model={[uploadFolderForm.value.description, "value"]}
                placeholder="请输入描述"
                rows={3}
              />
            </FormItem>
          </Form>
        </Modal>
      </div>
    );
  },
});
