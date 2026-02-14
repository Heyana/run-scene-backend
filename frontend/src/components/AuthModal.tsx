import { defineComponent, ref, reactive } from "vue";
import { Modal, Form, Input, Button, Tabs, message } from "ant-design-vue";
import {
  UserOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import { constant } from "@/api/const";

export interface AuthModalProps {
  visible: boolean;
  defaultTab?: "login" | "register";
  onSuccess?: () => void;
  onCancel?: () => void;
}

export default defineComponent({
  name: "AuthModal",
  props: {
    visible: {
      type: Boolean,
      required: true,
    },
    defaultTab: {
      type: String as () => "login" | "register",
      default: "login",
    },
  },
  emits: ["update:visible", "success", "cancel"],
  setup(props, { emit }) {
    const activeTab = ref(props.defaultTab);
    const loading = ref(false);

    // 登录表单
    const loginFormRef = ref();
    const loginForm = reactive({
      username: "",
      password: "",
    });

    // 注册表单
    const registerFormRef = ref();
    const registerForm = reactive({
      username: "",
      password: "",
      confirmPassword: "",
      email: "",
      phone: "",
      real_name: "",
    });

    // 登录
    const handleLogin = async () => {
      try {
        await loginFormRef.value.validate();
        loading.value = true;

        const res = await api.auth.login({
          username: loginForm.username,
          password: loginForm.password,
        });

        // 保存 token（使用正确的 key）
        localStorage.setItem(
          constant.runSceneBackendToken,
          res.data.access_token,
        );
        if (res.data.user) {
          localStorage.setItem("user", JSON.stringify(res.data.user));
        }

        message.success("登录成功");
        emit("success");
        handleClose();
      } catch (error: any) {
        console.error("登录失败:", error);
        if (error.response?.data?.msg) {
          message.error(error.response.data.msg);
        } else {
          message.error("登录失败");
        }
      } finally {
        loading.value = false;
      }
    };

    // 注册
    const handleRegister = async () => {
      try {
        await registerFormRef.value.validate();
        loading.value = true;

        await api.auth.register({
          username: registerForm.username,
          password: registerForm.password,
          email: registerForm.email,
          phone: registerForm.phone,
          real_name: registerForm.real_name,
        });

        message.success("注册成功，请登录");
        activeTab.value = "login";

        // 清空注册表单
        registerForm.username = "";
        registerForm.password = "";
        registerForm.confirmPassword = "";
        registerForm.email = "";
        registerForm.phone = "";
        registerForm.real_name = "";
      } catch (error: any) {
        console.error("注册失败:", error);
        if (error.response?.data?.msg) {
          message.error(error.response.data.msg);
        } else {
          message.error("注册失败");
        }
      } finally {
        loading.value = false;
      }
    };

    // 关闭弹窗
    const handleClose = () => {
      emit("update:visible", false);
      emit("cancel");
    };

    // 验证确认密码
    const validateConfirmPassword = (_rule: any, value: string) => {
      if (value && value !== registerForm.password) {
        return Promise.reject("两次输入的密码不一致");
      }
      return Promise.resolve();
    };

    return () => (
      <Modal
        open={props.visible}
        title="用户认证"
        footer={null}
        width={450}
        onCancel={handleClose}
        destroyOnClose
      >
        <Tabs v-model:activeKey={activeTab.value}>
          {/* 登录标签页 */}
          <Tabs.TabPane key="login" tab="登录">
            <Form
              ref={loginFormRef}
              model={loginForm}
              layout="vertical"
              style={{ marginTop: "20px" }}
            >
              <Form.Item
                name="username"
                rules={[{ required: true, message: "请输入用户名" }]}
              >
                <Input
                  v-model:value={loginForm.username}
                  size="large"
                  prefix={<UserOutlined />}
                  placeholder="用户名"
                  onPressEnter={handleLogin}
                />
              </Form.Item>
              <Form.Item
                name="password"
                rules={[{ required: true, message: "请输入密码" }]}
              >
                <Input.Password
                  v-model:value={loginForm.password}
                  size="large"
                  prefix={<LockOutlined />}
                  placeholder="密码"
                  onPressEnter={handleLogin}
                />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  size="large"
                  block
                  loading={loading.value}
                  onClick={handleLogin}
                >
                  登录
                </Button>
              </Form.Item>
            </Form>
          </Tabs.TabPane>

          {/* 注册标签页 */}
          <Tabs.TabPane key="register" tab="注册">
            <Form
              ref={registerFormRef}
              model={registerForm}
              layout="vertical"
              style={{ marginTop: "20px" }}
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: "请输入用户名" },
                  { min: 3, max: 50, message: "用户名长度为3-50个字符" },
                ]}
              >
                <Input
                  v-model:value={registerForm.username}
                  size="large"
                  prefix={<UserOutlined />}
                  placeholder="用户名"
                />
              </Form.Item>
              <Form.Item
                name="email"
                rules={[
                  { required: true, message: "请输入邮箱" },
                  { type: "email", message: "请输入有效的邮箱地址" },
                ]}
              >
                <Input
                  v-model:value={registerForm.email}
                  size="large"
                  prefix={<MailOutlined />}
                  placeholder="邮箱"
                />
              </Form.Item>
              <Form.Item name="phone">
                <Input
                  v-model:value={registerForm.phone}
                  size="large"
                  prefix={<PhoneOutlined />}
                  placeholder="手机号（可选）"
                />
              </Form.Item>
              <Form.Item name="real_name">
                <Input
                  v-model:value={registerForm.real_name}
                  size="large"
                  prefix={<UserOutlined />}
                  placeholder="真实姓名（可选）"
                />
              </Form.Item>
              <Form.Item
                name="password"
                rules={[
                  { required: true, message: "请输入密码" },
                  { min: 6, message: "密码长度至少6个字符" },
                ]}
              >
                <Input.Password
                  v-model:value={registerForm.password}
                  size="large"
                  prefix={<LockOutlined />}
                  placeholder="密码"
                />
              </Form.Item>
              <Form.Item
                name="confirmPassword"
                rules={[
                  { required: true, message: "请确认密码" },
                  { validator: validateConfirmPassword },
                ]}
              >
                <Input.Password
                  v-model:value={registerForm.confirmPassword}
                  size="large"
                  prefix={<LockOutlined />}
                  placeholder="确认密码"
                />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  size="large"
                  block
                  loading={loading.value}
                  onClick={handleRegister}
                >
                  注册
                </Button>
              </Form.Item>
            </Form>
          </Tabs.TabPane>
        </Tabs>
      </Modal>
    );
  },
});
