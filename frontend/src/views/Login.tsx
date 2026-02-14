import { defineComponent, reactive, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { Card, Form, Input, Button, Tabs, message } from "ant-design-vue";
import {
  UserOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
} from "@ant-design/icons-vue";
import { api } from "@/api/api";
import { constant } from "@/api/const";

export default defineComponent({
  name: "Login",
  setup() {
    const router = useRouter();
    const route = useRoute();
    const activeTab = ref("login");
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

        // 跳转到目标页面或首页
        let redirect = route.query.redirect as string;

        console.log("Log-- ", redirect, "redirect");
        // 处理可能的嵌套 redirect 参数
        if (redirect) {
          // 解码 URL
          redirect = decodeURIComponent(redirect);

          // 如果 redirect 本身是登录页，跳转到首页
          if (redirect.includes("/login")) {
            redirect = "/";
          }
        }
        console.log("Log-- ", redirect, "redirect");

        await router.replace(redirect || "/");
        console.log("Log-- router.replace executed");
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

    // 验证确认密码
    const validateConfirmPassword = (_rule: any, value: string) => {
      if (value && value !== registerForm.password) {
        return Promise.reject("两次输入的密码不一致");
      }
      return Promise.resolve();
    };

    return () => (
      <div
        style={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
        }}
      >
        <Card
          style={{
            width: "450px",
            boxShadow: "0 8px 32px rgba(0, 0, 0, 0.1)",
          }}
        >
          <div style={{ textAlign: "center", marginBottom: "24px" }}>
            <h1 style={{ fontSize: "28px", margin: "0 0 8px 0" }}>
              资产管理系统
            </h1>
            <p style={{ color: "#666", margin: 0 }}>欢迎使用</p>
          </div>

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
        </Card>
      </div>
    );
  },
});
