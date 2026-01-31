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

## 快速开始

### 1. 下载 Open Shelley

```bash
# 从 GitHub 下载最新版
curl -L -o shelley_linux_amd64 \
  https://github.com/boldsoftware/shelley/releases/latest/download/shelley_linux_amd64
chmod +x shelley_linux_amd64
```

### 2. 编译 Portal

```bash
go build -o portal main.go
```

### 3. 启动服务

```bash
# 设置环境变量
export ANTHROPIC_API_KEY="sk-ant-..."  # 必需
export PORTAL_TOKEN="your-secure-token" # 可选，不设置会自动生成

# 启动 Open Shelley (后台)
./shelley_linux_amd64 -db ./shelley.db serve -port 9001 &

# 启动 Portal
./portal
```

### 4. 访问

- 登录页: http://localhost:8000/login
- Portal: http://localhost:8000/portal
- Shelley: http://localhost:8000/

## 环境变量

### Portal
| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORTAL_TOKEN` | 登录认证 token | 自动生成 |
| `PORTAL_PORT` | 端口号 | 8000 |
| `SHELLEY_URL` | Open Shelley 地址 | http://localhost:9001 |

### Open Shelley
| 变量 | 说明 |
|------|------|
| `ANTHROPIC_API_KEY` | Anthropic API 密钥 (必需) |

## 使用 Systemd 部署

```bash
# 编辑服务文件，填入你的密钥
vim openshelley.service  # 修改 ANTHROPIC_API_KEY
vim portal.service       # 修改 PORTAL_TOKEN

# 安装服务
sudo cp openshelley.service portal.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable openshelley portal
sudo systemctl start openshelley portal

# 查看日志
journalctl -u openshelley -f
journalctl -u portal -f
```

## 自动更新

```bash
# 检查更新
./update-shelley.sh --check

# 执行更新
./update-shelley.sh

# 强制更新
./update-shelley.sh --force
```

## 功能截图

### Portal 首页
- 系统状态监控
- 一键启动/停止/重启
- 检查更新/执行更新

### Web 终端
- 完整的 xterm.js 终端
- 支持 256 色
- 支持窗口调整大小

### 文件管理器
- 文件浏览和编辑
- CodeMirror 语法高亮
- 创建/删除/重命名

## 安全建议

1. 使用强 token（长随机字符串）
2. 防火墙仅开放 8000 端口
3. 生产环境配置 HTTPS（使用 nginx 反向代理）
4. 保护好 ANTHROPIC_API_KEY

## 许可证

MIT License
