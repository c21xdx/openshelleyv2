# Open Shelley Portal 部署指南

## 架构说明

```
用户 → Portal (8000端口) → Open Shelley (9001端口内部)
                        → Terminal (WebSocket)
                        → File Manager (REST API)
                        → Service Management (REST API) ← NEW!
```

- **Portal** (`/home/exedev/002/portal/portal`): 门户服务，提供token验证、代理和服务管理
- **Open Shelley** (`/home/exedev/002/shelley_linux_amd64`): 开源Shelley AI Agent

## 文件位置

```
/home/exedev/002/
├── portal/                    # Portal门户服务
│   ├── portal                 # 编译后的二进制
│   ├── main.go               # Go源代码
│   ├── static/               # 静态文件（HTML/CSS/JS）
│   └── portal.service        # systemd服务文件
├── openshelley/              # Open Shelley工作目录
│   ├── shelley.db            # Shelley数据库
│   ├── shelley.json          # Shelley配置
│   └── openshelley.service   # systemd服务文件
└── shelley_linux_amd64       # Open Shelley二进制
```

## 快速启动（手动）

```bash
# 1. 启动Open Shelley（内部端口9001）
cd /home/exedev/002/openshelley
ANTHROPIC_API_KEY="sk-ant-..." \
  /home/exedev/002/shelley_linux_amd64 \
  -db ./shelley.db \
  -config ./shelley.json \
  serve -port 9001 &

# 2. 启动Portal（对外端口8000）
cd /home/exedev/002/portal
PORTAL_TOKEN="your-secure-token" \
  SHELLEY_URL="http://localhost:9001" \
  ./portal &
```

## Systemd服务部署（甲骨文VM）

### 1. 修改服务文件中的密钥

```bash
# 编辑Open Shelley服务文件
vim /home/exedev/002/openshelley/openshelley.service
# 修改 ANTHROPIC_API_KEY=YOUR_API_KEY_HERE

# 编辑Portal服务文件
vim /home/exedev/002/portal/portal.service
# 修改 PORTAL_TOKEN=YOUR_SECURE_TOKEN_HERE
```

### 2. 安装和启动服务

```bash
# 复制服务文件
sudo cp /home/exedev/002/openshelley/openshelley.service /etc/systemd/system/
sudo cp /home/exedev/002/portal/portal.service /etc/systemd/system/

# 重载systemd
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable openshelley portal

# 启动服务
sudo systemctl start openshelley
sudo systemctl start portal

# 检查状态
sudo systemctl status openshelley portal
```

### 3. 查看日志

```bash
# Open Shelley日志
journalctl -u openshelley -f

# Portal日志
journalctl -u portal -f
```

## 访问方式

- **登录页面**: `http://YOUR_SERVER:8000/login`
- **Portal首页**: `http://YOUR_SERVER:8000/portal` (包含服务管理面板)
- **Open Shelley**: `http://YOUR_SERVER:8000/`
- **终端**: `http://YOUR_SERVER:8000/portal/terminal`
- **文件管理器**: `http://YOUR_SERVER:8000/portal/files`

## 服务管理面板 (新功能!)

Portal 首页现在集成了服务管理功能，无需命令行操作：

- **状态监控**: 实时显示 Shelley 运行状态
- **版本信息**: 显示当前版本和最新版本
- **启动/停止/重启**: 一键控制 Shelley 服务
- **检查更新**: 检查 GitHub 最新版本
- **执行更新**: 自动下载并更新到最新版本
- **日志输出**: 实时显示操作日志

## 环境变量

### Portal
- `PORTAL_TOKEN`: 登录认证token（必需）
- `PORTAL_PORT`: 端口号，默认8000
- `SHELLEY_URL`: Open Shelley内部地址，默认http://localhost:9001

### Open Shelley
- `ANTHROPIC_API_KEY`: Anthropic API密钥（必需）
- 更多选项见 `./shelley_linux_amd64 --help`

## 功能说明

### Portal首页 (`/portal`)
- 系统状态监控
- 快速导航到各功能

### Open Shelley (`/`)
- AI编程助手
- 支持多对话
- 代码执行、文件操作

### Terminal (`/portal/terminal`)
- 完整的Web终端
- xterm.js + WebSocket
- 支持256色、调整大小

### File Manager (`/portal/files`)
- 文件浏览
- 代码编辑（CodeMirror语法高亮）
- 创建/删除/重命名文件和目录
- 支持Ctrl+S保存

## 安全注意事项

1. **设置强token**: 使用长随机字符串
2. **防火墙**: 仅开放8000端口
3. **HTTPS**: 生产环境使用反向代理（nginx）配置SSL
4. **API Key保护**: 不要泄露ANTHROPIC_API_KEY

## 故障排除

```bash
# 检查端口占用
ss -tlnp | grep -E '8000|9001'

# 测试Open Shelley
curl http://localhost:9001

# 测试Portal
curl http://localhost:8000/login

# 查看进程
ps aux | grep -E 'portal|shelley'
```
