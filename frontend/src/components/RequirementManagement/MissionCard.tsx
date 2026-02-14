import { defineComponent } from "vue";
import { Card, Avatar, Space } from "ant-design-vue";
import { UserOutlined, ClockCircleOutlined } from "@ant-design/icons-vue";
import type { PropType } from "vue";
import type { Mission } from "@/api/models/requirement";
import StatusTag from "./StatusTag";
import PriorityTag from "./PriorityTag";
import "./MissionCard.less";

export default defineComponent({
  name: "MissionCard",
  props: {
    mission: {
      type: Object as PropType<Mission>,
      required: true,
    },
    draggable: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["click"],
  setup(props, { emit }) {
    const handleClick = () => {
      emit("click", props.mission);
    };

    const formatDate = (dateStr?: string) => {
      if (!dateStr) return "";
      const date = new Date(dateStr);
      return `${date.getMonth() + 1}/${date.getDate()}`;
    };

    return () => (
      <div
        class={["mission-card-wrapper", { draggable: props.draggable }]}
        onClick={handleClick}
      >
        <Card class="mission-card" bordered={false} size="small">
          <div class="mission-header">
            <Space size={4}>
              <PriorityTag priority={props.mission.priority} />
              <span class="mission-key">{props.mission.mission_key}</span>
            </Space>
          </div>

          <div class="mission-title">{props.mission.title}</div>

          {props.mission.description && (
            <div class="mission-description">{props.mission.description}</div>
          )}

          <div class="mission-footer">
            <div class="mission-assignee">
              {props.mission.assignee ? (
                <Space size={4}>
                  <Avatar size={20} src={props.mission.assignee.avatar}>
                    {props.mission.assignee.real_name?.[0] ||
                      props.mission.assignee.username[0]}
                  </Avatar>
                  <span class="assignee-name">
                    {props.mission.assignee.real_name ||
                      props.mission.assignee.username}
                  </span>
                </Space>
              ) : (
                <Space size={4}>
                  <Avatar size={20} icon={<UserOutlined />} />
                  <span class="assignee-name unassigned">未指派</span>
                </Space>
              )}
            </div>

            {props.mission.due_date && (
              <div class="mission-due-date">
                <ClockCircleOutlined />
                <span>{formatDate(props.mission.due_date)}</span>
              </div>
            )}
          </div>
        </Card>
      </div>
    );
  },
});
