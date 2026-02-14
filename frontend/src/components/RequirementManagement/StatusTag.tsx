import { defineComponent } from "vue";
import { Tag } from "ant-design-vue";
import type { PropType } from "vue";

type MissionStatus = "todo" | "in_progress" | "done" | "closed";

const statusConfig: Record<MissionStatus, { label: string; color: string }> = {
  todo: { label: "待处理", color: "default" },
  in_progress: { label: "进行中", color: "processing" },
  done: { label: "已完成", color: "success" },
  closed: { label: "已关闭", color: "error" },
};

export default defineComponent({
  name: "StatusTag",
  props: {
    status: {
      type: String as PropType<MissionStatus>,
      required: true,
    },
  },
  setup(props) {
    return () => {
      const config = statusConfig[props.status] || statusConfig.todo;
      return <Tag color={config.color}>{config.label}</Tag>;
    };
  },
});
