import { defineComponent } from "vue";
import { Tag } from "ant-design-vue";
import type { PropType } from "vue";

type MissionPriority = "P0" | "P1" | "P2" | "P3";

const priorityConfig: Record<
  MissionPriority,
  { label: string; color: string }
> = {
  P0: { label: "紧急", color: "#ff4d4f" },
  P1: { label: "高", color: "#ff7a45" },
  P2: { label: "中", color: "#ffa940" },
  P3: { label: "低", color: "#8c8c8c" },
};

export default defineComponent({
  name: "PriorityTag",
  props: {
    priority: {
      type: String as PropType<MissionPriority>,
      required: true,
    },
  },
  setup(props) {
    return () => {
      const config = priorityConfig[props.priority] || priorityConfig.P3;
      return <Tag color={config.color}>{config.label}</Tag>;
    };
  },
});
