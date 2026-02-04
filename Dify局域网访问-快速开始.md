# Dify 局域网访问 - 快速开始

## 🚀 最快的解决方案（5分钟搞定）

### 方法：使用 PowerShell HTTP 代理

#### 步骤 1：启动代理

1. **以管理员身份**打开 PowerShell（右键 → 以管理员身份运行）

2. 执行以下命令：

```powershell
cd E:\hxy\my-project\run-scene-backend
powershell -ExecutionPolicy Bypass -File dify-proxy.ps1
```

3. 看到以下提示说明启动成功：

```
=========================================
  Dify 局域网访问代理
=========================================

监听地址: 0.0.0.0:8080
目标地址: http://localhost:3001
局域网访问: http://192.168.3.39:8080

✅ 代理服务器已启动！
```

#### 步骤 2：访问 Dify

- **局域网访问**：`http://192.168.3.39:8080`
- **本机访问**：`http://localhost:8080` 或 `http://localhost:3001`

#### 步骤 3：停止代理

在 PowerShell 窗口按 `Ctrl + C` 即可停止。

---

## ⚠️ 注意事项

1. **必须以管理员身份运行**，否则无法添加防火墙规则
2. **保持 PowerShell 窗口打开**，关闭窗口代理就停止了
3. **确保 Dify 服务正在运行**：
   ```powershell
   # 检查 Dify 是否运行
   curl http://localhost:3001
   ```

---

## 🔧 故障排查

### 问题 1：提示"需要管理员权限"

**解决**：右键 PowerShell → 以管理员身份运行

### 问题 2：端口 8080 被占用

**解决**：修改 `dify-proxy.ps1` 中的端口号

```powershell
$localEndpoint = "http://+:8090/"  # 改成 8090 或其他端口
```

### 问题 3：局域网访问不通

**检查步骤**：

1. 确认代理正在运行（PowerShell 窗口没有关闭）
2. 确认防火墙规则已添加：
   ```powershell
   netsh advfirewall firewall show rule name="Dify Proxy 8080"
   ```
3. 确认本机可以访问：
   ```powershell
   curl http://localhost:8080
   ```
4. 确认局域网 IP 正确：
   ```powershell
   ipconfig | findstr "IPv4"
   ```

### 问题 4：Dify 服务未启动

**解决**：

```powershell
cd E:\hxy\project-2026\dify-full
docker-compose up -d
docker-compose ps
```

---

## 📊 性能说明

- **延迟**：增加约 1-5ms（几乎无感）
- **并发**：支持多用户同时访问
- **稳定性**：适合日常使用，长期运行建议使用 Nginx

---

## 🎯 下一步

### 配置 DeepSeek 模型

1. 访问 Dify：`http://192.168.3.39:8080`
2. 完成初始化设置（创建管理员账号）
3. 进入 **设置 → 模型供应商**
4. 选择 **OpenAI API-Compatible**
5. 填入配置：
   ```
   Base URL: https://api.deepseek.com/v1
   API Key: 你的 DeepSeek API Key
   模型名称: deepseek-chat
   ```
6. 点击"保存"

### 创建知识库

1. 点击 **知识库 → 创建知识库**
2. 上传文档（支持 PDF、TXT、Markdown、Word 等）
3. 等待文档处理完成
4. 创建应用并关联知识库

---

## 💡 长期方案建议

如果需要长期使用，建议：

1. **迁移到 Linux**（飞牛 NAS）
   - 使用 `deploy-dify-to-nas.sh` 脚本
   - 性能更好，无需代理

2. **安装 Nginx 反向代理**
   - 使用 `setup-nginx-proxy.ps1` 脚本
   - 可以后台运行，无需保持 PowerShell 窗口

详细方案对比见：`docs/Dify局域网访问解决方案.md`

---

## 📞 需要帮助？

如果遇到问题，请提供以下信息：

1. 错误信息截图
2. PowerShell 输出日志
3. Dify 服务状态：
   ```powershell
   cd E:\hxy\project-2026\dify-full
   docker-compose ps
   docker-compose logs --tail=50
   ```
