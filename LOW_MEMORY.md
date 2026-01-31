# 低内存服务器优化指南 (1GB RAM)

如果你的服务器内存只有 1GB，按以下步骤优化。

## 1️⃣ 创建 Swap 分区 (必做)

```bash
# 创建 2GB swap 文件
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 永久生效
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# 优化 swap 使用策略
sudo sysctl vm.swappiness=10
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
```

## 2️⃣ 禁用浏览器 (可选，省 ~300-500MB)

如果不需要浏览器功能（截图、JS 执行），不安装 Chromium。

Shelley 在没有浏览器时会自动禁用相关工具。

```bash
# 如果已安装，可以卸载
sudo apt remove chromium-browser
```

## 3️⃣ 系统级优化

```bash
# 停止不必要的服务
sudo systemctl disable --now snapd 2>/dev/null
sudo systemctl disable --now ModemManager 2>/dev/null
sudo systemctl disable --now cups 2>/dev/null

# 查看内存占用最高的进程
ps aux --sort=-%mem | head -10
```

## 4️⃣ 监控内存

```bash
# 查看内存使用
free -h

# 实时监控
watch -n 2 free -h

# 查看 Shelley 和 Portal 内存
ps aux | grep -E 'shelley|portal' | awk '{print $6/1024 " MB", $11}'
```

## 5️⃣ 内存不足时的表现

- Shelley 响应变慢
- 浏览器工具失败或超时
- 系统变卡
- 进程被 OOM Killer 终止

## 6️⃣ 推荐配置

| 配置 | 1GB RAM | 2GB RAM |
|------|---------|----------|
| Swap | 2GB (必须) | 1-2GB (推荐) |
| Chromium | 不安装或按需用 | 可以安装 |
| 并发使用 | 1 人 | 2-3 人 |

## 7️⃣ 最低配置运行

如果内存实在不够，可以只跑 Shelley，不跑 Portal：

```bash
# 直接运行 Shelley (内存占用更低)
export ANTHROPIC_API_KEY="sk-ant-..."
./shelley -db ./data/shelley.db serve -port 8000
```

然后直接访问 `http://your-server:8000`，不需要 Portal 层。

缺点：没有 Token 认证、Web 终端、文件管理器。
