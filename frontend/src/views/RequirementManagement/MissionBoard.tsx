import { defineComponent, ref, onMounted, computed } from "vue";
import { useRoute } from "vue-router";
import {
  Card,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  message,
  Spin,
  Empty,
  Drawer,
  Dropdown,
  Menu,
} from "ant-design-vue";
import {
  PlusOutlined,
  ReloadOutlined,
  SettingOutlined,
  EditOutlined,
  DeleteOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import type {
  Mission,
  MissionList,
  MissionColumn,
} from "@/api/models/requirement";
import MissionCard from "@/components/RequirementManagement/MissionCard";
import MissionDetail from "./MissionDetail";
import "./MissionBoard.less";

export default defineComponent({
  name: "MissionBoard",
  setup() {
    const route = useRoute();
    const loading = ref(false);
    const missions = ref<Mission[]>([]);
    const missionLists = ref<MissionList[]>([]);
    const columns = ref<MissionColumn[]>([]);
    const selectedListId = ref<number>();
    const modalVisible = ref(false);
    const listModalVisible = ref(false);
    const columnModalVisible = ref(false);
    const detailVisible = ref(false);
    const selectedMission = ref<Mission | null>(null);
    const editingColumn = ref<MissionColumn | null>(null);
    const formRef = ref();
    const listFormRef = ref();
    const columnFormRef = ref();

    const projectId = computed(() => {
      const id = route.params.projectId;
      return id ? Number(id) : undefined;
    });

    const formData = ref({
      title: "",
      description: "",
      type: "feature" as "feature" | "enhancement" | "bug",
      priority: "P2" as "P0" | "P1" | "P2" | "P3",
      assignee_id: undefined as number | undefined,
      due_date: undefined as string | undefined,
      mission_column_id: undefined as number | undefined,
    });

    const listFormData = ref({
      name: "",
      type: "sprint" as "sprint" | "version" | "module",
      description: "",
      start_date: undefined as string | undefined,
      end_date: undefined as string | undefined,
    });

    const columnFormData = ref({
      name: "",
      color: "#1890ff",
    });

    // 加载任务列
    const loadColumns = async () => {
      if (!selectedListId.value) return;
      try {
        const res = await api.requirement.getMissionColumnList(
          selectedListId.value,
        );
        columns.value = Array.isArray(res.data) ? res.data : [];
      } catch (error) {
        message.error("加载任务列失败");
      }
    };

    // 加载任务列表
    const loadMissionLists = async () => {
      try {
        const params = projectId.value ? { project_id: projectId.value } : {};
        const res = await api.requirement.getMissionListList(params);
        // 后端直接返回数组，不是 { items: [] } 格式
        missionLists.value = Array.isArray(res.data)
          ? res.data
          : res.data.items || [];

        if (missionLists.value.length > 0 && missionLists.value[0]) {
          selectedListId.value = missionLists.value[0].id;
          await loadColumns();
          await loadMissions();
        }
      } catch (error) {
        message.error("加载任务列表失败");
      }
    };

    // 加载任务
    const loadMissions = async () => {
      if (!selectedListId.value) return;

      loading.value = true;
      try {
        const res = await api.requirement.getMissionList({
          mission_list_id: selectedListId.value,
        });
        // 后端直接返回数组，不是 { items: [] } 格式
        missions.value = Array.isArray(res.data)
          ? res.data
          : res.data.items || [];
      } catch (error) {
        message.error("加载任务失败");
      } finally {
        loading.value = false;
      }
    };

    // 按列获取任务
    const getMissionsByColumn = (columnId: number) => {
      return missions.value.filter((m) => m.mission_column_id === columnId);
    };

    // 显示创建任务列表对话框
    const handleCreateList = () => {
      listFormData.value = {
        name: "",
        type: "sprint",
        description: "",
        start_date: undefined,
        end_date: undefined,
      };
      listModalVisible.value = true;
    };

    // 提交任务列表表单
    const handleListSubmit = async () => {
      try {
        await listFormRef.value.validate();
        if (!projectId.value) {
          message.error("项目ID不存在");
          return;
        }
        await api.requirement.createMissionList({
          project_id: projectId.value,
          ...listFormData.value,
        });
        message.success("创建成功");
        listModalVisible.value = false;
        loadMissionLists();
      } catch (error) {
        console.error("创建失败:", error);
      }
    };

    // 显示创建/编辑列对话框
    const handleCreateColumn = () => {
      editingColumn.value = null;
      columnFormData.value = {
        name: "",
        color: "#1890ff",
      };
      columnModalVisible.value = true;
    };

    const handleEditColumn = (column: MissionColumn) => {
      editingColumn.value = column;
      columnFormData.value = {
        name: column.name,
        color: column.color,
      };
      columnModalVisible.value = true;
    };

    // 提交列表单
    const handleColumnSubmit = async () => {
      try {
        await columnFormRef.value.validate();
        if (!selectedListId.value) {
          message.error("请先选择任务列表");
          return;
        }

        if (editingColumn.value) {
          // 编辑
          await api.requirement.updateMissionColumn(editingColumn.value.id, {
            ...columnFormData.value,
          });
          message.success("更新成功");
        } else {
          // 创建
          await api.requirement.createMissionColumn({
            mission_list_id: selectedListId.value,
            ...columnFormData.value,
          });
          message.success("创建成功");
        }

        columnModalVisible.value = false;
        loadColumns();
      } catch (error) {
        console.error("操作失败:", error);
      }
    };

    // 删除列
    const handleDeleteColumn = async (column: MissionColumn) => {
      Modal.confirm({
        title: "确认删除",
        content: `确定要删除列"${column.name}"吗？该列下的任务不会被删除。`,
        okText: "确定",
        cancelText: "取消",
        onOk: async () => {
          try {
            await api.requirement.deleteMissionColumn(column.id);
            message.success("删除成功");
            loadColumns();
          } catch (error) {
            message.error("删除失败");
          }
        },
      });
    };

    // 显示创建任务对话框
    const handleCreate = (columnId: number) => {
      formData.value = {
        title: "",
        description: "",
        type: "feature",
        priority: "P2",
        assignee_id: undefined,
        due_date: undefined,
        mission_column_id: columnId,
      };
      modalVisible.value = true;
    };

    // 提交表单
    const handleSubmit = async () => {
      try {
        await formRef.value.validate();
        if (!selectedListId.value) {
          message.error("请先选择任务列表");
          return;
        }
        await api.requirement.createMission({
          mission_list_id: selectedListId.value,
          ...formData.value,
        });
        message.success("创建成功");
        modalVisible.value = false;
        loadMissions();
      } catch (error) {
        console.error("创建失败:", error);
      }
    };

    // 查看任务详情
    const handleViewDetail = (mission: Mission) => {
      selectedMission.value = mission;
      detailVisible.value = true;
    };

    // 关闭详情
    const handleCloseDetail = () => {
      detailVisible.value = false;
      selectedMission.value = null;
      loadMissions();
    };

    // 切换任务列表
    const handleListChange = async () => {
      await loadColumns();
      await loadMissions();
    };

    onMounted(() => {
      loadMissionLists();
    });

    return () => (
      <div class="mission-board-page">
        <div class="board-layout">
          {/* 左侧任务列表 */}
          <div class="board-sidebar">
            <div class="sidebar-header">
              <span class="sidebar-title">任务列表</span>
              <Button
                type="text"
                size="small"
                icon={<PlusOutlined />}
                onClick={handleCreateList}
              />
            </div>
            <div class="sidebar-content">
              {missionLists.value.length === 0 ? (
                <Empty
                  description="暂无任务列表"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                >
                  <Button
                    type="primary"
                    size="small"
                    onClick={handleCreateList}
                  >
                    创建列表
                  </Button>
                </Empty>
              ) : (
                <div class="mission-list-items">
                  {missionLists.value.map((list) => (
                    <div
                      key={list.id}
                      class={[
                        "mission-list-item",
                        selectedListId.value === list.id ? "active" : "",
                      ]}
                      onClick={() => {
                        selectedListId.value = list.id;
                        handleListChange();
                      }}
                    >
                      <div class="list-name">{list.name}</div>
                      <div class="list-count">{list.mission_count || 0}</div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* 右侧看板区域 */}
          <div class="board-main">
            <div class="board-header">
              <Space>
                <Button icon={<PlusOutlined />} onClick={handleCreateColumn}>
                  新建列
                </Button>
                <Button icon={<ReloadOutlined />} onClick={loadMissions}>
                  刷新
                </Button>
              </Space>
            </div>

            <Spin spinning={loading.value}>
              {!selectedListId.value ? (
                <Empty
                  description="请选择左侧的任务列表"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              ) : columns.value.length === 0 ? (
                <Empty
                  description="暂无任务列，请先创建任务列"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                >
                  <Button type="primary" onClick={handleCreateColumn}>
                    创建第一个任务列
                  </Button>
                </Empty>
              ) : (
                <div class="board-columns">
                  {columns.value.map((column) => {
                    const columnMissions = getMissionsByColumn(column.id);
                    return (
                      <div key={column.id} class="board-column">
                        <div class="column-header">
                          <div class="column-title">
                            <span
                              class="column-indicator"
                              style={{ backgroundColor: column.color }}
                            />
                            <span>{column.name}</span>
                            <span class="column-count">
                              {columnMissions.length}
                            </span>
                          </div>
                          <Space size="small">
                            <Button
                              type="text"
                              size="small"
                              icon={<PlusOutlined />}
                              onClick={() => handleCreate(column.id)}
                            />
                            <Dropdown
                              trigger={["click"]}
                              v-slots={{
                                overlay: () => (
                                  <Menu>
                                    <Menu.Item
                                      key="edit"
                                      icon={<EditOutlined />}
                                      onClick={() => handleEditColumn(column)}
                                    >
                                      编辑
                                    </Menu.Item>
                                    <Menu.Item
                                      key="delete"
                                      icon={<DeleteOutlined />}
                                      danger
                                      onClick={() => handleDeleteColumn(column)}
                                    >
                                      删除
                                    </Menu.Item>
                                  </Menu>
                                ),
                              }}
                            >
                              <Button
                                type="text"
                                size="small"
                                icon={<SettingOutlined />}
                              />
                            </Dropdown>
                          </Space>
                        </div>

                        <div class="column-content">
                          {columnMissions.length === 0 ? (
                            <Empty
                              description="暂无任务"
                              image={Empty.PRESENTED_IMAGE_SIMPLE}
                            />
                          ) : (
                            columnMissions.map((mission) => (
                              <MissionCard
                                key={mission.id}
                                mission={mission}
                                draggable
                                onClick={handleViewDetail}
                              />
                            ))
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </Spin>
          </div>
        </div>

        {/* 创建任务对话框 */}
        <Modal
          v-model:open={modalVisible.value}
          title="创建任务"
          onOk={handleSubmit}
          okText="创建"
          cancelText="取消"
          width={600}
        >
          <Form ref={formRef} model={formData.value} labelCol={{ span: 6 }}>
            <Form.Item
              label="任务标题"
              name="title"
              rules={[{ required: true, message: "请输入任务标题" }]}
            >
              <Input
                v-model:value={formData.value.title}
                placeholder="请输入任务标题"
              />
            </Form.Item>
            <Form.Item label="任务描述" name="description">
              <Input.TextArea
                v-model:value={formData.value.description}
                placeholder="请输入任务描述（可选）"
                rows={4}
              />
            </Form.Item>
            <Form.Item label="任务类型" name="type">
              <Select v-model:value={formData.value.type}>
                <Select.Option value="feature">功能</Select.Option>
                <Select.Option value="enhancement">优化</Select.Option>
                <Select.Option value="bug">缺陷</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item label="优先级" name="priority">
              <Select v-model:value={formData.value.priority}>
                <Select.Option value="P0">P0 - 紧急</Select.Option>
                <Select.Option value="P1">P1 - 高</Select.Option>
                <Select.Option value="P2">P2 - 中</Select.Option>
                <Select.Option value="P3">P3 - 低</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item label="截止日期" name="due_date">
              <DatePicker
                v-model:value={formData.value.due_date}
                style={{ width: "100%" }}
                placeholder="选择截止日期"
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 任务详情抽屉 */}
        <Drawer
          v-model:open={detailVisible.value}
          title="任务详情"
          width={600}
          onClose={handleCloseDetail}
        >
          {selectedMission.value && (
            <MissionDetail
              mission={selectedMission.value}
              onUpdate={handleCloseDetail}
            />
          )}
        </Drawer>

        {/* 创建任务列表对话框 */}
        <Modal
          v-model:open={listModalVisible.value}
          title="创建任务列表"
          onOk={handleListSubmit}
          okText="创建"
          cancelText="取消"
          width={600}
        >
          <Form
            ref={listFormRef}
            model={listFormData.value}
            labelCol={{ span: 6 }}
          >
            <Form.Item
              label="列表名称"
              name="name"
              rules={[{ required: true, message: "请输入列表名称" }]}
            >
              <Input
                v-model:value={listFormData.value.name}
                placeholder="例如: Sprint 1, v1.0.0, 用户模块"
              />
            </Form.Item>
            <Form.Item
              label="列表类型"
              name="type"
              rules={[{ required: true, message: "请选择列表类型" }]}
            >
              <Select v-model:value={listFormData.value.type}>
                <Select.Option value="sprint">Sprint（迭代）</Select.Option>
                <Select.Option value="version">Version（版本）</Select.Option>
                <Select.Option value="module">Module（模块）</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={listFormData.value.description}
                placeholder="请输入列表描述（可选）"
                rows={3}
              />
            </Form.Item>
            <Form.Item label="开始日期" name="start_date">
              <DatePicker
                v-model:value={listFormData.value.start_date}
                style={{ width: "100%" }}
                placeholder="选择开始日期"
              />
            </Form.Item>
            <Form.Item label="结束日期" name="end_date">
              <DatePicker
                v-model:value={listFormData.value.end_date}
                style={{ width: "100%" }}
                placeholder="选择结束日期"
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 创建/编辑任务列对话框 */}
        <Modal
          v-model:open={columnModalVisible.value}
          title={editingColumn.value ? "编辑任务列" : "创建任务列"}
          onOk={handleColumnSubmit}
          okText={editingColumn.value ? "保存" : "创建"}
          cancelText="取消"
          width={500}
        >
          <Form
            ref={columnFormRef}
            model={columnFormData.value}
            labelCol={{ span: 6 }}
          >
            <Form.Item
              label="列名称"
              name="name"
              rules={[{ required: true, message: "请输入列名称" }]}
            >
              <Input
                v-model:value={columnFormData.value.name}
                placeholder="例如: 编辑器, 进图, 长期优化"
              />
            </Form.Item>
            <Form.Item label="颜色" name="color">
              <Input
                v-model:value={columnFormData.value.color}
                type="color"
                style={{ width: "100px" }}
              />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  },
});
