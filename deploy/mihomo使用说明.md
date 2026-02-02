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

## 查看当前节点

```bash
# 查看"便宜机场"代理组当前选择的节点
curl http://127.0.0.1:9090/proxies/便宜机场

# 查看所有代理组状态
curl http://127.0.0.1:9090/proxies

# 测试代理是否工作
curl -x http://127.0.0.1:7890 https://www.google.com -I
```

## 切换节点

```bash
# 切换到指定节点（替换节点名称）
curl -X PUT http://127.0.0.1:9090/proxies/便宜机场 -d '{"name":"新加坡vless-1"}' -H "Content-Type: application/json"

# 可用节点名称参考 mihomo.yaml 中的 proxies 列表
```

## 卸载服务

```bash
sudo systemctl stop mihomo && sudo systemctl disable mihomo && sudo rm /etc/systemd/system/mihomo.service && sudo systemctl daemon-reload
```
