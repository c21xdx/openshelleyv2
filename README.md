# Open Shelley Portal

一个为 [Open Shelley](https://github.com/boldsoftware/shelley) 设计的 Web 门户，提供：

- 🔐 Token 认证
- 🖥️ Web 终端 (xterm.js)
- 📁 文件管理器 (CodeMirror 语法高亮)
- ⚙️ 服务管理面板
- 🔄 自动更新支持

## 架构

```
用户 → Portal (8000端口) → Open Shelley (9001端口内部)
                        → Terminal (WebSocket)
                        → File Manager (REST API)
                        → Service Management (REST API)
```

## 🚀 一键安装 (推荐)

```bash
# 1. 设置 API Key
export ANTHROPIC_API_KEY="sk-ant-..."

# 2. 运行安装脚本
curl -sSL https://raw.githubusercontent.com/c21xdx/openshelley/main/install.sh | bash

# 3. 启动服务
cd ~/openshelley && ./start.sh
```

安装完成后，访问 `http://your-server:8000/login` 并使用显示的 Token 登录。

## 📦 手动安装

### 前置条件

- Go 1.21+
- curl, jq
- Anthropic API Key

### 步骤

```bash
# 1. 克隆项目
git clone https://github.com/c21xdx/openshelley.git
cd openshelley

# 2. 编译 Portal
go build -o portal main.go

# 3. 下载 Open Shelley
curl -L -o shelley \
  https://github.com/boldsoftware/shelley/releases/latest/download/shelley_linux_amd64
chmod +x shelley

# 4. 启动
export ANTHROPIC_API_KEY="sk-ant-..."
export PORTAL_TOKEN="your-secret-token"

./shelley -db ./shelley.db serve -port 9001 &
SHELLEY_URL=http://localhost:9001 ./portal
```

## 📁 文件结构

安装后的目录结构：

```
~/openshelley/
├── shelley              # Open Shelley 二进制
├── portal               # Portal 二进制
├── static/              # 前端页面
├── data/
│   ├── shelley.db       # 数据库
│   └── shelley.json     # 配置文件
├── .env                 # 环境变量 (API Key, Token 等)
├── start.sh             # 启动脚本
├── stop.sh              # 停止脚本
├── status.sh            # 状态检查
├── update-shelley.sh    # 更新脚本
└── *.service            # systemd 服务文件
```

## 🛠️ 常用命令

```bash
cd ~/openshelley

# 启动/停止/状态
./start.sh
./stop.sh
./status.sh

# 更新 Shelley
./update-shelley.sh           # 检查并更新
./update-shelley.sh --check   # 仅检查
./update-shelley.sh --force   # 强制更新
```

## 🔧 Systemd 部署

如果希望服务开机自启：

```bash
sudo cp ~/openshelley/*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable openshelley portal
sudo systemctl start openshelley portal

# 查看日志
journalctl -u openshelley -f
journalctl -u portal -f
```

## ⚙️ 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ANTHROPIC_API_KEY` | Anthropic API 密钥 | (必需) |
| `PORTAL_TOKEN` | 登录认证 token | 自动生成 |
| `PORTAL_PORT` | Portal 端口 | 8000 |
| `SHELLEY_PORT` | Shelley 内部端口 | 9001 |
| `SHELLEY_URL` | Shelley 地址 | http://localhost:9001 |
| `BASE_DIR` | 安装目录 | (自动检测) |

## 🔐 安全建议

1. 使用强 token（安装脚本会自动生成）
2. 防火墙仅开放 Portal 端口 (8000)
3. 生产环境配置 HTTPS（nginx 反向代理）
4. 保护好 `.env` 文件

## 📸 功能截图

### Portal 首页
- 系统状态监控
- 一键启动/停止/重启
- 检查更新/执行更新

### Web 终端
- 完整的 xterm.js 终端
- 支持 256 色和窗口调整

### 文件管理器
- 文件浏览和编辑
- CodeMirror 语法高亮
- 创建/删除/重命名

## 📄 License

MIT License
