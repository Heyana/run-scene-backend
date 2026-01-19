import { defineComponent } from "vue";
import { RouterView } from "vue-router";
import AppLayout from "@/components/Layout";

export default defineComponent({
  name: "App",
  setup() {
    return () => <RouterView />;
  },
});
