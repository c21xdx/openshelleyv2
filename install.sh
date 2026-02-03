#!/bin/bash
#
# Open Shelley Portal 一键安装脚本
# 用法: curl -sSL https://raw.githubusercontent.com/c21xdx/openshelleyv2/main/install.sh | bash
#

set -e

# 配置
INSTALL_DIR="${INSTALL_DIR:-$HOME/openshelley}"
SHELLEY_REPO="boldsoftware/shelley"
PORTAL_REPO="c21xdx/openshelleyv2"
SHELLEY_PORT="9001"
PORTAL_PORT="8000"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

# 检查依赖
check_deps() {
    log_info "检查依赖..."
    local missing=""
    
    for cmd in curl jq; do
        if ! command -v $cmd &> /dev/null; then
            missing="$missing $cmd"
        fi
    done
    
    if ! command -v go &> /dev/null; then
        missing="$missing golang"
    fi
    
    if [[ -n "$missing" ]]; then
        log_error "缺少依赖:$missing"
        log_info "请运行: sudo apt install$missing"
        exit 1
    fi
    
    # 检查浏览器
    if ! command -v headless-shell &> /dev/null && ! command -v chromium-browser &> /dev/null && ! command -v chromium &> /dev/null && ! command -v google-chrome &> /dev/null; then
        log_warn "未检测到浏览器"
        log_info "AMD64 推荐安装 Chrome Headless Shell:"
        log_info "  curl -sSL https://raw.githubusercontent.com/c21xdx/openshelleyv2/main/install-headless-shell.sh | bash"
        log_info "或者: sudo apt install chromium-browser"
        echo ""
    fi
    
    log_success "依赖检查通过"
}

# 检查 API Key
check_api_key() {
    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        log_error "请设置 ANTHROPIC_API_KEY 环境变量"
        log_info "例如: export ANTHROPIC_API_KEY='sk-ant-...'"
        exit 1
    fi
    log_success "API Key 已设置"
}

# 检测架构
detect_arch() {
    local arch=$(uname -m)
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    
    case $arch in
        x86_64)  arch="amd64" ;;
        aarch64) arch="arm64" ;;
        arm64)   arch="arm64" ;;
        *)       log_error "不支持的架构: $arch"; exit 1 ;;
    esac
    
    echo "${os}_${arch}"
}

# 创建目录结构
create_dirs() {
    log_info "创建目录: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"/{data,logs}
    log_success "目录创建完成"
}

# 下载 Shelley
download_shelley() {
    log_info "下载 Open Shelley..."
    local arch=$(detect_arch)
    local url="https://github.com/$SHELLEY_REPO/releases/latest/download/shelley_${arch}"
    
    curl -L -o "$INSTALL_DIR/shelley" "$url"
    chmod +x "$INSTALL_DIR/shelley"
    
    local version=$($INSTALL_DIR/shelley version 2>/dev/null | jq -r '.tag' 2>/dev/null || echo "unknown")
    log_success "Shelley 下载完成 ($version)"
}

# 下载并编译 Portal
download_portal() {
    log_info "下载 Portal 源码..."
    local tmp_dir=$(mktemp -d)
    
    curl -L -o "$tmp_dir/portal.tar.gz" \
        "https://github.com/$PORTAL_REPO/archive/refs/heads/main.tar.gz"
    
    tar -xzf "$tmp_dir/portal.tar.gz" -C "$tmp_dir"
    
    log_info "编译 Portal..."
    cd "$tmp_dir/openshelleyv2-main"
    go build -o "$INSTALL_DIR/portal" main.go
    
    # 复制需要的文件
    cp -r static "$INSTALL_DIR/"
    cp update-shelley.sh "$INSTALL_DIR/"
    cp AGENTS.md "$INSTALL_DIR/" 2>/dev/null || true
    chmod +x "$INSTALL_DIR/update-shelley.sh"
    
    rm -rf "$tmp_dir"
    log_success "Portal 编译完成"
}

# 生成配置
generate_config() {
    log_info "生成配置..."
    
    # 生成随机 token
    local portal_token=$(openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | xxd -p | tr -d '\n')
    
    # 创建 shelley.json
    echo '{"default_model": "claude-sonnet-4-20250514"}' > "$INSTALL_DIR/data/shelley.json"
    
    # 创建环境变量文件
    cat > "$INSTALL_DIR/.env" <<EOF
# Open Shelley Portal 配置
# 生成时间: $(date)

# Anthropic API Key (必需)
ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY

# Portal 认证 Token
PORTAL_TOKEN=$portal_token

# 端口配置
SHELLEY_PORT=$SHELLEY_PORT
PORTAL_PORT=$PORTAL_PORT

# 内部配置
SHELLEY_URL=http://localhost:$SHELLEY_PORT
BASE_DIR=$INSTALL_DIR
EOF
    
    chmod 600 "$INSTALL_DIR/.env"
    
    log_success "Portal Token: $portal_token"
}

# 创建启动脚本
create_scripts() {
    log_info "创建启动脚本..."
    
    # start.sh - 启动所有服务
    cat > "$INSTALL_DIR/start.sh" <<'EOF'
#!/bin/bash
cd "$(dirname "$0")"
source .env

echo "启动 Open Shelley..."
./shelley -db ./data/shelley.db -config ./data/shelley.json serve -port $SHELLEY_PORT &
echo $! > ./data/shelley.pid
sleep 2

echo "启动 Portal..."
PORTAL_TOKEN=$PORTAL_TOKEN SHELLEY_URL=$SHELLEY_URL BASE_DIR=$BASE_DIR ./portal &
echo $! > ./data/portal.pid

echo ""
echo "✅ 服务已启动!"
echo "🔗 访问: http://localhost:$PORTAL_PORT/login"
echo "🔑 Token: $PORTAL_TOKEN"
EOF
    
    # stop.sh - 停止所有服务
    cat > "$INSTALL_DIR/stop.sh" <<'EOF'
#!/bin/bash
cd "$(dirname "$0")"

echo "停止服务..."
[[ -f ./data/portal.pid ]] && kill $(cat ./data/portal.pid) 2>/dev/null && rm ./data/portal.pid
[[ -f ./data/shelley.pid ]] && kill $(cat ./data/shelley.pid) 2>/dev/null && rm ./data/shelley.pid
pkill -f "shelley.*serve" 2>/dev/null || true
pkill -f "portal" 2>/dev/null || true
echo "✅ 服务已停止"
EOF
    
    # status.sh - 检查状态
    cat > "$INSTALL_DIR/status.sh" <<'EOF'
#!/bin/bash
cd "$(dirname "$0")"
source .env

echo "=== Open Shelley Portal 状态 ==="
echo ""

if pgrep -f "shelley.*serve" > /dev/null; then
    echo "✅ Shelley: 运行中 (port $SHELLEY_PORT)"
else
    echo "❌ Shelley: 已停止"
fi

if pgrep -f "portal" > /dev/null; then
    echo "✅ Portal: 运行中 (port $PORTAL_PORT)"
else
    echo "❌ Portal: 已停止"
fi

echo ""
echo "🔑 Portal Token: $PORTAL_TOKEN"
echo "🔗 访问地址: http://localhost:$PORTAL_PORT/login"
EOF

    # token.sh - 快速查看 Token
    cat > "$INSTALL_DIR/token.sh" <<'EOF'
#!/bin/bash
cd "$(dirname "$0")"
source .env
echo "$PORTAL_TOKEN"
EOF
    
    chmod +x "$INSTALL_DIR"/{start.sh,stop.sh,status.sh,token.sh}
    log_success "启动脚本创建完成"
}

# 创建 systemd 服务
create_systemd() {
    log_info "创建 systemd 服务文件..."
    
    local user=$(whoami)
    
    # openshelley.service
    cat > "$INSTALL_DIR/openshelley.service" <<EOF
[Unit]
Description=Open Shelley Agent Service
After=network.target

[Service]
Type=simple
User=$user
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/shelley -db $INSTALL_DIR/data/shelley.db -config $INSTALL_DIR/data/shelley.json serve -port \${SHELLEY_PORT}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    
    # portal.service
    cat > "$INSTALL_DIR/portal.service" <<EOF
[Unit]
Description=Portal Gateway Service
After=network.target openshelley.service
Wants=openshelley.service

[Service]
Type=simple
User=$user
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/portal
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    
    log_success "systemd 服务文件已创建"
    log_info "要安装为系统服务，运行:"
    echo "    sudo cp $INSTALL_DIR/*.service /etc/systemd/system/"
    echo "    sudo systemctl daemon-reload"
    echo "    sudo systemctl enable openshelley portal"
    echo "    sudo systemctl start openshelley portal"
}

# 完成信息
print_finish() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  ✅ Open Shelley Portal 安装完成!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "安装目录: $INSTALL_DIR"
    echo ""
    echo "快速启动:"
    echo "    cd $INSTALL_DIR && ./start.sh"
    echo ""
    echo "停止服务:"
    echo "    cd $INSTALL_DIR && ./stop.sh"
    echo ""
    echo "查看状态:"
    echo "    cd $INSTALL_DIR && ./status.sh"
    echo ""
    echo "Systemd 部署 (可选):"
    echo "    sudo cp $INSTALL_DIR/*.service /etc/systemd/system/"
    echo "    sudo systemctl daemon-reload"
    echo "    sudo systemctl enable --now openshelley portal"
    echo ""
}

# 主函数
main() {
    echo ""
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}  Open Shelley Portal 安装程序${NC}"
    echo -e "${BLUE}======================================${NC}"
    echo ""
    
    check_deps
    check_api_key
    create_dirs
    download_shelley
    download_portal
    generate_config
    create_scripts
    create_systemd
    print_finish
}

main "$@"
