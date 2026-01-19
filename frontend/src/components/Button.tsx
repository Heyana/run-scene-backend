import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "Button",
  props: {
    type: {
      type: String as PropType<"primary" | "secondary" | "danger">,
      default: "primary",
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["click"],
  setup(props, { slots, emit }) {
    const handleClick = (e: MouseEvent) => {
      if (!props.disabled) {
        emit("click", e);
      }
    };

    return () => (
      <button
        class={["btn", `btn-${props.type}`, { disabled: props.disabled }]}
        onClick={handleClick}
        disabled={props.disabled}
      >
        {slots.default?.()}
      </button>
    );
  },
});
