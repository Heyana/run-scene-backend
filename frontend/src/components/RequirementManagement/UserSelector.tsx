import { defineComponent, ref, watch } from "vue";
import { Modal, Input, Avatar, Empty, Spin } from "ant-design-vue";
import { UserOutlined, SearchOutlined } from "@ant-design/icons-vue";
import { api } from "@/api/api";
import "./UserSelector.less";

interface User {
  id: number;
  username: string;
  email: string;
  real_name?: string;
  avatar?: string;
}

export default defineComponent({
  name: "UserSelector",
  props: {
    visible: {
      type: Boolean,
      required: true,
    },
    projectId: {
      type: Number,
      required: true,
    },
    currentUserId: {
      type: Number,
    },
    title: {
      type: String,
      default: "选择指派人",
    },
  },
  emits: ["update:visible", "select"],
  setup(props, { emit }) {
    const loading = ref(false);
    const users = ref<User[]>([]);
    const searchKeyword = ref("");
    const filteredUsers = ref<User[]>([]);

    // 加载项目成员
    const loadUsers = async () => {
      if (!props.projectId) return;

      loading.value = true;
      try {
        const res = await api.requirement.getProjectMembers(props.projectId);
        users.value = (res.data || [])
          .map((member: any) => member.user)
          .filter(Boolean);
        filterUsers();
      } catch (error) {
        console.error("加载成员失败:", error);
      } finally {
        loading.value = false;
      }
    };

    // 过滤用户
    const filterUsers = () => {
      const keyword = searchKeyword.value.toLowerCase();
      if (!keyword) {
        filteredUsers.value = users.value;
      } else {
        filteredUsers.value = users.value.filter(
          (user) =>
            user.username.toLowerCase().includes(keyword) ||
            user.email.toLowerCase().includes(keyword) ||
            (user.real_name && user.real_name.toLowerCase().includes(keyword)),
        );
      }
    };

    // 选择用户
    const handleSelect = (user: User) => {
      emit("select", user);
      handleClose();
    };

    // 取消选择（指派给自己或清空）
    const handleUnassign = () => {
      emit("select", null);
      handleClose();
    };

    // 关闭弹窗
    const handleClose = () => {
      emit("update:visible", false);
      searchKeyword.value = "";
    };

    // 监听弹窗显示状态
    watch(
      () => props.visible,
      (visible) => {
        if (visible) {
          loadUsers();
        }
      },
    );

    // 监听搜索关键词
    watch(searchKeyword, filterUsers);

    return () => (
      <Modal
        open={props.visible}
        title={props.title}
        onCancel={handleClose}
        footer={null}
        width={400}
        class="user-selector-modal"
      >
        <div class="user-selector">
          <Input
            v-model:value={searchKeyword.value}
            placeholder="搜索成员（姓名、用户名、邮箱）"
            prefix={<SearchOutlined />}
            allowClear
            class="search-input"
          />

          <Spin spinning={loading.value}>
            <div class="user-list">
              {/* 取消指派选项 */}
              <div class="user-item unassign" onClick={handleUnassign}>
                <Avatar size={32} icon={<UserOutlined />} />
                <div class="user-info">
                  <div class="user-name">取消指派</div>
                  <div class="user-email">清空指派人</div>
                </div>
              </div>

              {filteredUsers.value.length === 0 ? (
                <Empty
                  description="暂无成员"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              ) : (
                filteredUsers.value.map((user) => (
                  <div
                    key={user.id}
                    class={[
                      "user-item",
                      user.id === props.currentUserId ? "current" : "",
                    ]}
                    onClick={() => handleSelect(user)}
                  >
                    {user.avatar ? (
                      <Avatar size={32} src={user.avatar} />
                    ) : (
                      <Avatar size={32} icon={<UserOutlined />}>
                        {user.real_name?.[0] || user.username[0]}
                      </Avatar>
                    )}
                    <div class="user-info">
                      <div class="user-name">
                        {user.real_name || user.username}
                        {user.id === props.currentUserId && (
                          <span class="current-tag">当前</span>
                        )}
                      </div>
                      <div class="user-email">{user.email}</div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </Spin>
        </div>
      </Modal>
    );
  },
});
