import { defineComponent, ref, computed } from "vue";
import {
  Form,
  Input,
  Select,
  DatePicker,
  Button,
  Space,
  Divider,
  Avatar,
  Timeline,
  Upload,
  message,
} from "ant-design-vue";
import {
  UserOutlined,
  ClockCircleOutlined,
  PaperClipOutlined,
  SendOutlined,
} from "@ant-design/icons-vue";
import type { PropType } from "vue";
import { api } from "@/api/api";
import type { Mission } from "@/api/models/requirement";
import StatusTag from "@/components/RequirementManagement/StatusTag";
import PriorityTag from "@/components/RequirementManagement/PriorityTag";
import "./MissionDetail.less";

export default defineComponent({
  name: "MissionDetail",
  props: {
    mission: {
      type: Object as PropType<Mission>,
      required: true,
    },
  },
  emits: ["update"],
  setup(props, { emit }) {
    const loading = ref(false);
    const newComment = ref("");
    const comments = ref(props.mission.comments || []);
    const attachments = ref(props.mission.attachments || []);

    const formData = ref({
      title: props.mission.title,
      description: props.mission.description,
      type: props.mission.type,
      priority: props.mission.priority,
      status: props.mission.status,
      assignee_id: props.mission.assignee_id,
      due_date: props.mission.due_date,
    });

    // 更新任务
    const handleUpdate = async () => {
      loading.value = true;
      try {
        await api.requirement.updateMission(props.mission.id, formData.value);
        message.success("更新成功");
        emit("update");
      } catch (error) {
        message.error("更新失败");
      } finally {
        loading.value = false;
      }
    };

    // 添加评论
    const handleAddComment = async () => {
      if (!newComment.value.trim()) {
        message.warning("请输入评论内容");
        return;
      }

      try {
        const res = await api.requirement.addMissionComment(props.mission.id, {
          content: newComment.value,
        });
        comments.value.push(res.data);
        message.success("评论成功");
        newComment.value = "";
      } catch (error) {
        message.error("评论失败");
      }
    };

    // 上传附件
    const handleUpload = async (file: File) => {
      try {
        const res = await api.requirement.uploadMissionAttachment(
          props.mission.id,
          file,
        );
        attachments.value.push(res.data);
        message.success("上传成功");
        return false;
      } catch (error) {
        message.error("上传失败");
        return false;
      }
    };

    const formatTime = (dateStr: string) => {
      const date = new Date(dateStr);
      const now = new Date();
      const diff = Math.floor((now.getTime() - date.getTime()) / 1000);

      if (diff < 60) return `${diff} 秒前`;
      if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
      if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
      return `${Math.floor(diff / 86400)} 天前`;
    };

    return () => (
      <div class="mission-detail">
        <div class="detail-header">
          <Space>
            <PriorityTag priority={props.mission.priority} />
            <StatusTag status={props.mission.status} />
            <span class="mission-key">{props.mission.mission_key}</span>
          </Space>
        </div>

        <Form layout="vertical" model={formData.value}>
          <Form.Item label="任务标题">
            <Input
              v-model:value={formData.value.title}
              placeholder="请输入任务标题"
            />
          </Form.Item>

          <Form.Item label="任务描述">
            <Input.TextArea
              v-model:value={formData.value.description}
              placeholder="请输入任务描述"
              rows={4}
            />
          </Form.Item>

          <Form.Item label="任务类型">
            <Select v-model:value={formData.value.type}>
              <Select.Option value="feature">功能</Select.Option>
              <Select.Option value="enhancement">优化</Select.Option>
              <Select.Option value="bug">缺陷</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label="优先级">
            <Select v-model:value={formData.value.priority}>
              <Select.Option value="P0">P0 - 紧急</Select.Option>
              <Select.Option value="P1">P1 - 高</Select.Option>
              <Select.Option value="P2">P2 - 中</Select.Option>
              <Select.Option value="P3">P3 - 低</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label="状态">
            <Select v-model:value={formData.value.status}>
              <Select.Option value="todo">待处理</Select.Option>
              <Select.Option value="in_progress">进行中</Select.Option>
              <Select.Option value="done">已完成</Select.Option>
              <Select.Option value="closed">已关闭</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label="截止日期">
            <DatePicker
              v-model:value={formData.value.due_date}
              style={{ width: "100%" }}
              placeholder="选择截止日期"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              loading={loading.value}
              onClick={handleUpdate}
              block
            >
              保存更改
            </Button>
          </Form.Item>
        </Form>

        <Divider>评论</Divider>

        <div class="comment-section">
          <div class="comment-input">
            <Input.TextArea
              v-model:value={newComment.value}
              placeholder="添加评论..."
              rows={3}
            />
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={handleAddComment}
              style={{ marginTop: "8px" }}
            >
              发送
            </Button>
          </div>

          {comments.value.length > 0 && (
            <Timeline style={{ marginTop: "16px" }}>
              {comments.value.map((comment) => (
                <Timeline.Item key={comment.id}>
                  <div class="comment-item">
                    <Space>
                      <Avatar size={24} src={comment.user?.avatar}>
                        {comment.user?.real_name?.[0] ||
                          comment.user?.username[0]}
                      </Avatar>
                      <span class="comment-user">
                        {comment.user?.real_name || comment.user?.username}
                      </span>
                      <span class="comment-time">
                        {formatTime(comment.created_at)}
                      </span>
                    </Space>
                    <div class="comment-content">{comment.content}</div>
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          )}
        </div>

        <Divider>附件</Divider>

        <div class="attachment-section">
          <Upload beforeUpload={handleUpload} showUploadList={false}>
            <Button icon={<PaperClipOutlined />}>上传附件</Button>
          </Upload>

          {attachments.value.length > 0 && (
            <div class="attachment-list">
              {attachments.value.map((attachment) => (
                <div key={attachment.id} class="attachment-item">
                  <PaperClipOutlined />
                  <span>{attachment.file_name}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    );
  },
});
