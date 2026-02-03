# Open Shelley Portal v2 开发总结

> 创建时间: 2026-01-31
> 更新时间: 2026-02-03
> 项目地址: https://github.com/c21xdx/openshelleyv2
> 本地开发目录: /home/exedev/006

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

## 仓库说明

### 新仓库 (v2 - 当前项目)

| 仓库 | 分支 | 架构 | 浏览器 | 说明 |
|------|------|------|--------|------|
| `openshelleyv2` | `main` | AMD64 | Chrome Headless Shell | **主力维护** |

### 旧仓库 (v1 - 已弃用)

| 仓库 | 分支 | 架构 | 浏览器 | 说明 |
|------|------|------|--------|------|
| `openshelley` | `main` | AMD64 | chromium-browser | 旧版 |
| `openshelley` | `arm` | ARM64 | chromium-browser | 甲骨文 ARM |
| `openshelley` | `headless-shell` | AMD64 | Chrome Headless Shell | 已迁移到 v2 |

## 文件结构

### GitHub 仓库
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
├── install-headless-shell.sh  # Chrome Headless Shell 安装 (headless-shell 分支)
├── update-shelley.sh    # Shelley 更新脚本
├── AGENTS.md            # AI Agent 工作指南
├── DEVELOPMENT_NOTES.md # 开发总结
├── LOW_MEMORY.md        # 低内存优化 (main 分支)
├── ARM64.md             # ARM64 说明 (arm 分支)
├── HEADLESS_SHELL.md    # Headless Shell 说明 (headless-shell 分支)
└── README.md            # 项目说明
```

### 安装后目录
```
~/openshelley/
├── shelley              # Shelley 二进制
├── portal               # Portal 二进制
├── static/              # 前端页面
├── data/                # 数据目录
│   ├── shelley.db       # 数据库
│   └── shelley.json     # 配置
├── .env                 # 环境变量 (Token, API Key 等)
├── start.sh             # 启动脚本
├── stop.sh              # 停止脚本
├── status.sh            # 状态检查
├── token.sh             # 查看 Token
├── update-shelley.sh    # 更新脚本
├── AGENTS.md            # Agent 指南
└── *.service            # systemd 服务文件
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
curl -sSL https://raw.githubusercontent.com/c21xdx/openshelleyv2/main/install.sh | bash

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
curl -sSL https://raw.githubusercontent.com/c21xdx/openshelleyv2/arm/install.sh | bash

# 4. 启动
cd ~/openshelley && ./start.sh
```

## 常用命令

```bash
cd ~/openshelley

./start.sh              # 启动服务
./stop.sh               # 停止服务
./status.sh             # 查看状态 (含 Token)
./token.sh              # 快速查看 Token
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

### 更新后出问题，需要回退

**方法 1**: 通过 Portal 页面
1. 访问 Portal 首页
2. 点击 "⏮️ Rollback" 按钮
3. 选择要恢复的备份版本
4. 点击 "Restore" 确认

**方法 2**: 命令行手动回退
```bash
cd ~/openshelley

# 查看可用备份
ls -la shelley.backup.*

# 停止服务
./stop.sh

# 恢复备份 (替换为实际文件名)
cp shelley.backup.20260131_120000 shelley

# 重新启动
./start.sh
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

**测试结果** (Chromium 快照版):
- 内存节省有限 (~50-100 MB)
- 中文字体需额外处理
- 稳定性不如 apt

**后续发现**: Google 官方提供 Chrome for Testing
- 来源: https://googlechromelabs.github.io/chrome-for-testing/
- 稳定版，跟随 Chrome 发布
- 创建了 `headless-shell` 分支

**结论**: 主分支使用 chromium-browser，想用官方 headless-shell 可切换到 headless-shell 分支

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
  
- **2026-01-31**: 回退功能
  - 新增 Portal 回退按钮 (⏮️ Rollback)
  - 新增 `/portal/api/mgmt/backups` API 列出备份
  - 新增 `/portal/api/mgmt/rollback` API 执行回退
  - 回退前自动备份当前版本

- **2026-02-02**: headless-shell 分支
  - 新增 `headless-shell` 分支，使用 Google 官方 Chrome Headless Shell
  - 来源: https://googlechromelabs.github.io/chrome-for-testing/
  - 仅支持 AMD64

- **2026-02-02**: Bug 修复和优化
  - 修复文件管理器默认路径硬编码问题 (改为 $HOME)
  - 修复启动/停止脚本名称不一致问题
  - 修复二进制文件名称不一致问题 (shelley vs shelley_linux_amd64)
  - Portal 按钮移到顶部中央，避免遮挡 Shelley UI
  - 新增 Show Token 按钮，方便查看登录 Token
  - 新增 token.sh 脚本快速获取 Token
  - install.sh 自动复制 AGENTS.md 到安装目录
  - AGENTS.md 新增规则：完成后不显示完整代码

- **2026-02-02**: 中文乱码修复
  - 截图中文乱码需安装字体: `sudo apt install -y fonts-noto-cjk`

- **2026-02-03**: 项目重构
  - 创建新仓库 `openshelleyv2`，作为主力维护项目
  - 从旧仓库的 `headless-shell` 分支迁移而来
  - 使用 Google 官方 Chrome Headless Shell
  - 本地开发目录: `/home/exedev/006`
  - 旧仓库 `openshelley` 保留但不再维护

---

## 下一步计划 / TODO

- [ ] ARM64 支持 (等待 Google 发布 ARM64 版 headless-shell)
- [ ] 自动化测试
- [ ] 文档完善

---

## 快速开始 (新开发者)

```bash
# 克隆项目
cd /home/exedev/006

# 编译 Portal
go build -o portal main.go

# 查看状态
git status
git log --oneline -5

# 推送更改
git add <files>
git commit -m "message"
git push
```
