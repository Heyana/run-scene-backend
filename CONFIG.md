# 配置系统说明

项目现在使用更友好的 **YAML 配置文件** 替代了传统的 `.env` 文件。

## 🚀 快速开始

1. **复制配置模板**：

   ```bash
   cp config.example.yaml config.yaml
   ```

2. **根据需要修改配置**：
   编辑 `config.yaml` 文件中的相关配置项

3. **启动应用**：
   应用会自动读取 `config.yaml` 配置

## 📁 配置文件

- `config.yaml` - 主配置文件（需要自己创建）
- `config.example.yaml` - 配置模板和示例
- `.env` - 旧版配置文件（仍支持，向后兼容）

## ⚡ 配置优先级

```
环境变量 > YAML配置文件 > 默认值
```

例如：

- `config.yaml` 中设置 `app.port: 8080`
- 环境变量设置 `SERVER_PORT=9000`
- 最终使用端口：`9000`（环境变量优先）

## 📋 主要配置项

### 应用配置

```yaml
app:
  env: development # 环境: development | production
  name: your_app_name # 项目名称
  port: 23347 # 服务器端口
```

### 网络配置

```yaml
network:
  local: # 开发环境
    ip: 192.168.3.39
    cdn_port: 23357
  public: # 生产环境
    ip: 111.229.160.27
    cdn_port: 23357
```

### 备份配置

```yaml
backup:
  enabled: true # 是否启用备份
  local_path: ./backups # 本地备份路径
  retention_days: 7 # 保留天数
  auto_cleanup: false # 是否自动清理（建议false）
```

### 腾讯云 COS 配置

```yaml
cos:
  enabled: false # 是否启用云存储
  secret_id: "" # 从腾讯云控制台获取
  secret_key: "" # 从腾讯云控制台获取
  bucket_url: "" # 存储桶URL
  region: ap-shanghai # 区域
```

## 🔧 环境变量覆盖

所有 YAML 配置都可以用环境变量覆盖：

| YAML 路径        | 环境变量         | 示例值           |
| ---------------- | ---------------- | ---------------- |
| `app.env`        | `APP_ENV`        | `production`     |
| `app.port`       | `SERVER_PORT`    | `8080`           |
| `backup.enabled` | `BACKUP_ENABLED` | `true`           |
| `cos.secret_id`  | `COS_SECRET_ID`  | `your_secret_id` |

## 📝 备份系统特性

- ⏰ **固定时间备份**：每天 12:00 和 24:00 自动备份
- 🔒 **默认禁用清理**：防止重要备份被误删
- 📦 **增量 CDN 备份**：只备份修改过的文件
- ☁️ **可选云存储**：支持上传到腾讯云 COS
- 🚫 **重复检测**：同一时间点不会重复备份

## 🔄 从.env 迁移

如果你之前使用 `.env` 文件：

1. `.env` 文件仍然有效（向后兼容）
2. 建议迁移到 `config.yaml` 获得更好的体验
3. 两种方式可以并存，YAML 优先级更高

## 📚 更多信息

查看 `config.example.yaml` 获得完整的配置示例和注释说明。
