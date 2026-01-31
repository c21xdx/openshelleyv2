#!/bin/bash
#
# Open Shelley 自动更新脚本
# 用法: ./update-shelley.sh [--force]
#

set -e

# 配置
INSTALL_DIR="/home/exedev/002"
BINARY_NAME="shelley_linux_amd64"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
GITHUB_REPO="boldsoftware/shelley"
ARCH="linux_amd64"  # 或 linux_arm64, darwin_amd64, darwin_arm64

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查依赖
check_deps() {
    for cmd in curl jq; do
        if ! command -v $cmd &> /dev/null; then
            log_error "需要 $cmd，请先安装: sudo apt install $cmd"
            exit 1
        fi
    done
}

# 获取当前版本
get_current_version() {
    if [[ -f "$BINARY_PATH" ]]; then
        $BINARY_PATH version 2>/dev/null | jq -r '.tag' 2>/dev/null || echo "unknown"
    else
        echo "not_installed"
    fi
}

# 获取最新版本信息
get_latest_release() {
    curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
}

# 主更新逻辑
update_shelley() {
    local force="$1"
    
    log_info "检查 Open Shelley 更新..."
    
    # 获取版本信息
    local current_version=$(get_current_version)
    local release_info=$(get_latest_release)
    local latest_version=$(echo "$release_info" | jq -r '.tag_name')
    local download_url=$(echo "$release_info" | jq -r ".assets[] | select(.name == \"shelley_${ARCH}\") | .browser_download_url")
    
    if [[ -z "$latest_version" || "$latest_version" == "null" ]]; then
        log_error "无法获取最新版本信息"
        exit 1
    fi
    
    log_info "当前版本: $current_version"
    log_info "最新版本: $latest_version"
    
    # 检查是否需要更新
    if [[ "$current_version" == "$latest_version" && "$force" != "--force" ]]; then
        log_success "已经是最新版本!"
        exit 0
    fi
    
    if [[ "$force" == "--force" ]]; then
        log_warn "强制更新模式"
    fi
    
    log_info "下载新版本..."
    log_info "URL: $download_url"
    
    # 创建临时文件
    local tmp_file=$(mktemp)
    
    # 下载
    if ! curl -L -o "$tmp_file" "$download_url"; then
        log_error "下载失败"
        rm -f "$tmp_file"
        exit 1
    fi
    
    # 验证下载的文件
    if ! file "$tmp_file" | grep -q "ELF 64-bit"; then
        log_error "下载的文件不是有效的二进制文件"
        rm -f "$tmp_file"
        exit 1
    fi
    
    # 停止服务
    log_info "停止 Open Shelley 服务..."
    pkill -f "shelley_linux_amd64.*9001" 2>/dev/null || true
    sleep 2
    
    # 备份旧版本
    if [[ -f "$BINARY_PATH" ]]; then
        local backup_path="${BINARY_PATH}.backup.$(date +%Y%m%d_%H%M%S)"
        log_info "备份旧版本到: $backup_path"
        cp "$BINARY_PATH" "$backup_path"
    fi
    
    # 安装新版本
    log_info "安装新版本..."
    mv "$tmp_file" "$BINARY_PATH"
    chmod +x "$BINARY_PATH"
    
    # 验证安装
    local new_version=$($BINARY_PATH version 2>/dev/null | jq -r '.tag' 2>/dev/null || echo "unknown")
    log_success "已更新到: $new_version"
    
    # 重启服务
    log_info "重启 Open Shelley 服务..."
    cd /home/exedev/002/openshelley
    nohup $BINARY_PATH \
        -db ./shelley.db \
        -config ./shelley.json \
        -default-model claude-sonnet-4-20250514 \
        serve -port 9001 > /tmp/openshelley.log 2>&1 &
    
    sleep 2
    
    # 检查服务是否启动
    if pgrep -f "shelley_linux_amd64.*9001" > /dev/null; then
        log_success "Open Shelley 服务已启动"
    else
        log_error "服务启动失败，请检查日志: /tmp/openshelley.log"
        exit 1
    fi
    
    log_success "更新完成!"
}

# 显示帮助
show_help() {
    echo "Open Shelley 更新脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --force     强制更新，即使版本相同"
    echo "  --check     仅检查是否有更新，不执行更新"
    echo "  --help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0              # 检查并更新"
    echo "  $0 --force      # 强制重新安装最新版"
    echo "  $0 --check      # 仅检查更新"
}

# 仅检查更新
check_only() {
    log_info "检查 Open Shelley 更新..."
    
    local current_version=$(get_current_version)
    local release_info=$(get_latest_release)
    local latest_version=$(echo "$release_info" | jq -r '.tag_name')
    local published_at=$(echo "$release_info" | jq -r '.published_at')
    
    echo ""
    echo "当前版本: $current_version"
    echo "最新版本: $latest_version"
    echo "发布时间: $published_at"
    echo ""
    
    if [[ "$current_version" == "$latest_version" ]]; then
        log_success "已经是最新版本"
    else
        log_warn "有新版本可用!"
        echo "运行 '$0' 来更新"
    fi
}

# 清理旧备份（保留最近3个）
cleanup_backups() {
    log_info "清理旧备份..."
    ls -t ${BINARY_PATH}.backup.* 2>/dev/null | tail -n +4 | xargs rm -f 2>/dev/null || true
    log_success "清理完成"
}

# 主入口
main() {
    check_deps
    
    case "${1:-}" in
        --help|-h)
            show_help
            ;;
        --check)
            check_only
            ;;
        --force)
            update_shelley "--force"
            cleanup_backups
            ;;
        "")
            update_shelley
            cleanup_backups
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
