# Open Shelley Portal 开发总结

> 创建时间: 2026-01-31
> 项目地址: https://github.com/c21xdx/openshelley

---

## 项目概述

为 [Open Shelley](https://github.com/boldsoftware/shelley) (开源 AI Agent) 打造的 Web 门户，提供：
- Token 认证登录
- Web 终端 (xterm.js)
- 文件管理器 (CodeMirror 语法高亮)
- 服务管理面板 (启动/停止/重启/更新)
- 反向代理到 Shelley

## 架构

```
用户 → Portal (8000) → Open Shelley (9001)
                    → Terminal (WebSocket)
                    → File Manager (REST API)
                    → Service Management
```

## 分支说明

| 分支 | 目标架构 | 推荐配置 | 说明 |
|------|----------|----------|------|
| `main` | AMD64 (x86_64) | 1GB RAM + 2GB Swap | 含 LOW_MEMORY.md |
| `arm` | ARM64 (aarch64) | 2GB+ RAM | 甲骨文 ARM 实例优化 |

## 文件结构

```
github-repo/
├── main.go              # Portal 源码 (Go)
├── go.mod / go.sum      # Go 依赖
├── static/              # 前端页面
│   ├── index.html       # Portal 首页 (服务管理)
│   ├── login.html       # 登录页
│   ├── terminal.html    # Web 终端
│   └── files.html       # 文件管理器
├── install.sh           # 一键安装脚本
├── update-shelley.sh    # Shelley 更新脚本
├── AGENTS.md            # AI Agent 工作指南
├── LOW_MEMORY.md        # 低内存优化 (main 分支)
├── ARM64.md             # ARM64 说明 (arm 分支)
└── README.md            # 项目说明
```

## 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `ANTHROPIC_API_KEY` | ✅ | - | Anthropic API 密钥 |
| `PORTAL_TOKEN` | ❌ | 自动生成 | 登录认证 Token |
| `PORTAL_PORT` | ❌ | 8000 | Portal 端口 |
| `SHELLEY_PORT` | ❌ | 9001 | Shelley 内部端口 |
| `SHELLEY_URL` | ❌ | http://localhost:9001 | Shelley 地址 |
| `BASE_DIR` | ❌ | 自动检测 | 安装目录 |

## 部署步骤

### AMD64 (1GB 内存)

```bash
# 1. 创建 Swap (必须)
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# 2. 安装浏览器
sudo apt install -y chromium-browser

# 3. 设置 API Key
export ANTHROPIC_API_KEY="sk-ant-..."

# 4. 一键安装
curl -sSL https://raw.githubusercontent.com/c21xdx/openshelley/main/install.sh | bash

# 5. 启动
cd ~/openshelley && ./start.sh
```

### ARM64 (12GB 内存)

```bash
# 1. 安装浏览器
sudo apt install -y chromium-browser

# 2. 设置 API Key
export ANTHROPIC_API_KEY="sk-ant-..."

# 3. 一键安装 (arm 分支)
curl -sSL https://raw.githubusercontent.com/c21xdx/openshelley/arm/install.sh | bash

# 4. 启动
cd ~/openshelley && ./start.sh
```

## 常用命令

```bash
cd ~/openshelley

./start.sh              # 启动服务
./stop.sh               # 停止服务
./status.sh             # 查看状态
./update-shelley.sh     # 更新 Shelley
./update-shelley.sh --check   # 仅检查更新
```

## Systemd 服务

```bash
# 安装为系统服务
sudo cp ~/openshelley/*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable openshelley portal
sudo systemctl start openshelley portal

# 查看日志
journalctl -u openshelley -f
journalctl -u portal -f
```

## 故障排查

### 服务无法启动

```bash
# 检查端口占用
ss -tlnp | grep -E '8000|9001'

# 检查进程
ps aux | grep -E 'shelley|portal'

# 查看日志
cat /tmp/openshelley.log
```

### 内存不足 (AMD64)

```bash
# 检查内存和 Swap
free -h

# 检查 Swap 是否启用
swapon --show

# 如果 Swap 未启用，重新创建
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### 浏览器工具不可用

```bash
# 检查浏览器
which chromium-browser || which chromium || which headless-shell

# 测试浏览器
chromium-browser --version

# 安装缺失的依赖
sudo apt install -y libnspr4 libnss3 libexpat1 libfontconfig1
```

### API 调用失败

```bash
# 检查 API Key
echo $ANTHROPIC_API_KEY

# 测试 Shelley 直接访问
curl http://localhost:9001

# 测试 Portal
curl http://localhost:8000/login
```

## 探索过的方案 (未采用)

### 1. 远程浏览器

**想法**: 在低内存服务器上连接远程浏览器

**问题**:
- localhost 无法从远程浏览器访问
- 需要修改 Shelley 源码支持 `chromedp.NewRemoteAllocator`
- 增加复杂度，收益不大

**结论**: 不实用，放弃

### 2. headless-shell 替代 Chromium

**想法**: 使用精简版 Chrome 节省内存

**测试结果**:
- 内存节省有限 (~50-100 MB)
- 中文字体需额外处理
- 从 Chromium 快照下载，稳定性不如 apt

**结论**: 收益不大，使用官方 chromium-browser 更可靠

## Token 优化建议

已写入 AGENTS.md，主要原则：

1. **忽略不必要的目录**: node_modules, venv, __pycache__, .git 等
2. **优先用 curl**: 测试 API 比浏览器省 token
3. **少用截图**: 每张截图 ~1000-2000 tokens
4. **用 grep 搜索**: 不要读取整个文件
5. **用 head/tail**: 预览大文件

## 安全建议

1. 使用强 Token (安装脚本自动生成)
2. 防火墙仅开放 Portal 端口 (8000)
3. 生产环境配置 HTTPS (nginx 反向代理)
4. 保护好 .env 文件和 API Key

## 相关链接

- [Open Shelley](https://github.com/boldsoftware/shelley) - 开源 AI Agent
- [Anthropic API](https://console.anthropic.com) - 获取 API Key
- [chromedp](https://github.com/chromedp/chromedp) - Go 浏览器自动化库

## 版本历史

- **2026-01-31**: 初始版本
  - Portal 门户服务
  - 一键安装脚本
  - AMD64/ARM64 双架构支持
  - AGENTS.md Token 优化指南
