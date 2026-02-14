import { defineComponent, ref, onMounted } from "vue";
import {
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
import type { Mission, MissionList, Project } from "@/api/models/requirement";
import MissionCard from "@/components/RequirementManagement/MissionCard";
import MissionDetail from "./MissionDetail";
import UserSelector from "@/components/RequirementManagement/UserSelector";
import "./MissionBoard.less";

export default defineComponent({
  name: "MissionBoard",
  setup() {
    const loading = ref(false);
    const missions = ref<Mission[]>([]); // 恢复 missions 状�?
    const missionLists = ref<MissionList[]>([]);
    const projects = ref<Project[]>([]);
    const selectedProjectId = ref<number>(); // 选中的项目ID
    const modalVisible = ref(false);
    const listModalVisible = ref(false);
    const detailVisible = ref(false);
    const userSelectorVisible = ref(false);
    const selectedMission = ref<Mission | null>(null);
    const assigningMission = ref<Mission | null>(null); // 正在指派的任�?
    const editingList = ref<MissionList | null>(null);
    const formRef = ref();
    const listFormRef = ref();

    const formData = ref({
      title: "",
      description: "",
      type: "feature" as "feature" | "enhancement" | "bug",
      priority: "P2" as "P0" | "P1" | "P2" | "P3",
      assignee_id: undefined as number | undefined,
      due_date: undefined as string | undefined,
      mission_list_id: undefined as number | undefined,
    });

    const listFormData = ref({
      name: "",
      type: "sprint" as "sprint" | "version" | "module",
      description: "",
      color: "#1890ff",
      start_date: undefined as string | undefined,
      end_date: undefined as string | undefined,
    });

    // 加载项目列表（用于左侧筛选）
    const loadProjects = async () => {
      try {
        // 只加载用户有权限的项�?
        const res = await api.requirement.getProjectList();
        projects.value = Array.isArray(res.data)
          ? res.data
          : res.data.items || [];

        // 如果有项目，默认选中第一�?
        if (projects.value.length > 0 && projects.value[0]) {
          selectedProjectId.value = projects.value[0].id;
        }
      } catch (error) {
        message.error("加载项目列表失败");
      }
    };

    // 加载任务列表（看板列�? �?Preload missions
    const loadMissionLists = async () => {
      try {
        const params = selectedProjectId.value
          ? { project_id: selectedProjectId.value }
          : {};
        const res = await api.requirement.getMissionListList(params);
        missionLists.value = Array.isArray(res.data)
          ? res.data
          : res.data.items || [];
      } catch (error: any) {
        console.error("加载任务列表失败:", error);
        if (error.response?.data?.code === 403) {
          message.error("无权访问该项目的任务列表");
        } else {
          message.error("加载任务列表失败");
        }
      }
    };

    // 单独加载任务
    const loadMissions = async () => {
      if (!selectedProjectId.value) return;

      loading.value = true;
      try {
        const res = await api.requirement.getMissionList({
          project_id: selectedProjectId.value,
        });
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
    const getMissionsByList = (listId: number) => {
      return missions.value.filter((m) => m.mission_list_id === listId);
    };

    // 切换项目
    const handleProjectChange = async (projectId: number) => {
      selectedProjectId.value = projectId;
      await loadMissionLists();
      await loadMissions();
    };

    // 显示创建任务列表对话�?
    const handleCreateList = () => {
      editingList.value = null;
      listFormData.value = {
        name: "",
        type: "sprint",
        description: "",
        color: "#1890ff",
        start_date: undefined,
        end_date: undefined,
      };
      listModalVisible.value = true;
    };

    const handleEditList = (list: MissionList) => {
      editingList.value = list;
      listFormData.value = {
        name: list.name,
        type: list.type,
        description: list.description || "",
        color: list.color,
        start_date: list.start_date,
        end_date: list.end_date,
      };
      listModalVisible.value = true;
    };

    // 提交任务列表表单
    const handleListSubmit = async () => {
      try {
        await listFormRef.value.validate();

        if (editingList.value) {
          // 编辑
          await api.requirement.updateMissionList(editingList.value.id, {
            ...listFormData.value,
          });
          message.success("更新成功");
        } else {
          // 创建
          if (!selectedProjectId.value) {
            message.error("请先选择项目");
            return;
          }
          await api.requirement.createMissionList({
            project_id: selectedProjectId.value,
            ...listFormData.value,
          });
          message.success("创建成功");
        }

        listModalVisible.value = false;
        await loadMissionLists();
        await loadMissions();
      } catch (error) {
        console.error("操作失败:", error);
      }
    };

    // 删除任务列表
    const handleDeleteList = async (list: MissionList) => {
      Modal.confirm({
        title: "确认删除",
        content: `确定要删除列"${list.name}"吗？该列下的任务不会被删除。`,
        okText: "确定",
        cancelText: "取消",
        onOk: async () => {
          try {
            await api.requirement.deleteMissionList(list.id);
            message.success("删除成功");`n        await loadMissions();
            await loadMissions();
          } catch (error) {
            message.error("删除失败");
          }
        },
      });
    };

    // 显示创建任务对话�?
    const handleCreate = (listId: number) => {
      formData.value = {
        title: "",
        description: "",
        type: "feature",
        priority: "P2",
        assignee_id: undefined,
        due_date: undefined,
        mission_list_id: listId,
      };
      modalVisible.value = true;
    };

    // 提交表单
    const handleSubmit = async () => {
      try {
        await formRef.value.validate();
        if (!formData.value.mission_list_id) {
          message.error("请先选择任务列表");
          return;
        }
        await api.requirement.createMission({
          mission_list_id: formData.value.mission_list_id,
          title: formData.value.title,
          description: formData.value.description,
          type: formData.value.type,
          priority: formData.value.priority,
          assignee_id: formData.value.assignee_id,
          due_date: formData.value.due_date,
        });
        message.success("创建成功");
        modalVisible.value = false;
        // 重新加载任务
        await loadMissions();
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
      // 重新加载任务
      loadMissions();
    };

    // 打开指派人选择�?
    const handleAssignClick = (mission: Mission) => {
      assigningMission.value = mission;
      userSelectorVisible.value = true;
    };

    // 选择指派�?
    const handleUserSelect = async (user: any) => {
      if (!assigningMission.value) return;

      try {
        await api.requirement.updateMission(assigningMission.value.id, {
          assignee_id: user ? user.id : null,
        });
        message.success(
          user ? `已指派给 ${user.real_name || user.username}` : "已取消指�?,
        );`n        await loadMissions();
      } catch (error) {
        message.error("指派失败");
      }
    };

    onMounted(async () => {
      await loadProjects();
      await loadMissionLists();`n      await loadMissions();`n    });

    return () => (
      <div class="mission-board-page">
        <div class="board-layout">
          {/* 左侧项目筛�?*/}
          <div class="board-sidebar">
            <div class="sidebar-header">
              <span class="sidebar-title">项目</span>
            </div>
            <div class="sidebar-content">
              <div class="mission-list-items">
                {projects.value.map((project) => (
                  <div
                    key={project.id}
                    class={[
                      "mission-list-item",
                      selectedProjectId.value === project.id ? "active" : "",
                    ]}
                    onClick={() => handleProjectChange(project.id)}
                  >
                    <div class="list-name">{project.name}</div>
                    <div class="list-count">{project.mission_count || 0}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* 右侧看板区域 */}
          <div class="board-main">
            <div class="board-header">
              <Space>
                <Button icon={<PlusOutlined />} onClick={handleCreateList}>
                  新建�?
                </Button>
                <Button icon={<ReloadOutlined />} onClick={loadMissions}>
                  刷新
                </Button>
              </Space>
            </div>

            <Spin spinning={loading.value}>
              {missionLists.value.length === 0 ? (
                <Empty
                  description="暂无任务列，请先创建任务�?
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                >
                  <Button type="primary" onClick={handleCreateList}>
                    创建第一个任务列
                  </Button>
                </Empty>
              ) : (
                <div class="board-columns">
                  {missionLists.value.map((list) => {
                    const listMissions = getMissionsByList(list.id);
                    return (
                      <div key={list.id} class="board-column">
                        <div class="column-header">
                          <div class="column-title">
                            <span
                              class="column-indicator"
                              style={{ backgroundColor: list.color }}
                            />
                            <span>{list.name}</span>
                            <span class="column-count">
                              {listMissions.length}
                            </span>
                          </div>
                          <Space size="small">
                            <Button
                              type="text"
                              size="small"
                              icon={<PlusOutlined />}
                              onClick={() => handleCreate(list.id)}
                            />
                            <Dropdown
                              trigger={["click"]}
                              v-slots={{
                                overlay: () => (
                                  <Menu>
                                    <Menu.Item
                                      key="edit"
                                      icon={<EditOutlined />}
                                      onClick={() => handleEditList(list)}
                                    >
                                      编辑
                                    </Menu.Item>
                                    <Menu.Item
                                      key="delete"
                                      icon={<DeleteOutlined />}
                                      danger
                                      onClick={() => handleDeleteList(list)}
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
                          {listMissions.length === 0 ? (
                            <Empty
                              description="暂无任务"
                              image={Empty.PRESENTED_IMAGE_SIMPLE}
                            />
                          ) : (
                            listMissions.map((mission: Mission) => (
                              <MissionCard
                                key={mission.id}
                                mission={mission}
                                draggable
                                onClick={handleViewDetail}
                                onAssignClick={handleAssignClick}
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

        {/* 创建任务对话�?*/}
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
              rules={[
                { required: true, message: "请输入任务标�? },
                { min: 2, message: "任务标题至少2个字�? },
              ]}
            >
              <Input
                v-model:value={formData.value.title}
                placeholder="请输入任务标题（至少2个字符）"
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
            <Form.Item label="优先�? name="priority">
              <Select v-model:value={formData.value.priority}>
                <Select.Option value="P0">P0 - 紧�?/Select.Option>
                <Select.Option value="P1">P1 - �?/Select.Option>
                <Select.Option value="P2">P2 - �?/Select.Option>
                <Select.Option value="P3">P3 - �?/Select.Option>
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

        {/* 创建/编辑任务列表对话�?*/}
        <Modal
          v-model:open={listModalVisible.value}
          title={editingList.value ? "编辑任务�? : "创建任务�?}
          onOk={handleListSubmit}
          okText={editingList.value ? "保存" : "创建"}
          cancelText="取消"
          width={600}
        >
          <Form
            ref={listFormRef}
            model={listFormData.value}
            labelCol={{ span: 6 }}
          >
            <Form.Item
              label="列名�?
              name="name"
              rules={[{ required: true, message: "请输入列名称" }]}
            >
              <Input
                v-model:value={listFormData.value.name}
                placeholder="例如: 编辑�? 进图, 长期优化"
              />
            </Form.Item>
            <Form.Item
              label="列类�?
              name="type"
              rules={[{ required: true, message: "请选择列类�? }]}
            >
              <Select v-model:value={listFormData.value.type}>
                <Select.Option value="sprint">Sprint（迭代）</Select.Option>
                <Select.Option value="version">Version（版本）</Select.Option>
                <Select.Option value="module">Module（模块）</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item label="颜色" name="color">
              <Input
                v-model:value={listFormData.value.color}
                type="color"
                style={{ width: "100px" }}
              />
            </Form.Item>
            <Form.Item label="描述" name="description">
              <Input.TextArea
                v-model:value={listFormData.value.description}
                placeholder="请输入列描述（可选）"
                rows={3}
              />
            </Form.Item>
            <Form.Item label="开始日�? name="start_date">
              <DatePicker
                v-model:value={listFormData.value.start_date}
                style={{ width: "100%" }}
                placeholder="选择开始日�?
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

        {/* 人员选择�?*/}
        {selectedProjectId.value && (
          <UserSelector
            visible={userSelectorVisible.value}
            onUpdate:visible={(val: boolean) =>
              (userSelectorVisible.value = val)
            }
            projectId={selectedProjectId.value}
            currentUserId={assigningMission.value?.assignee_id}
            onSelect={handleUserSelect}
          />
        )}
      </div>
    );
  },
});



