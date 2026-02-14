import { defineComponent, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { Layout } from "ant-design-vue";
import "./index.less";

const { Content } = Layout;

export default defineComponent({
  name: "RequirementManagement",
  setup() {
    const router = useRouter();
    const route = useRoute();

    onMounted(() => {
      // 如果在根路径，重定向到公司列表
      if (route.path === "/requirement-management") {
        router.replace("/requirement-management/companies");
      }
    });

    return () => (
      <Layout class="requirement-management-layout">
        <Content class="requirement-management-content">
          <router-view />
        </Content>
      </Layout>
    );
  },
});
