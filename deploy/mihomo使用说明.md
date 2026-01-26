# Mihomo 代理使用说明

## 快速安装 systemd 服务

```bash
sudo cp mihomo.service /etc/systemd/system/ && sudo chmod 644 /etc/systemd/system/mihomo.service && sudo systemctl daemon-reload && sudo systemctl enable mihomo && sudo systemctl restart mihomo && sudo systemctl status mihomo
```

## 常用命令

```bash
# 启动
sudo systemctl start mihomo

# 停止
sudo systemctl stop mihomo

# 重启
sudo systemctl restart mihomo

# 查看状态
sudo systemctl status mihomo

# 查看日志
sudo journalctl -u mihomo -f
```

## 卸载服务

```bash
sudo systemctl stop mihomo && sudo systemctl disable mihomo && sudo rm /etc/systemd/system/mihomo.service && sudo systemctl daemon-reload
```
