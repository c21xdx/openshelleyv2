#!/bin/bash
#
# 安装 Chrome Headless Shell (官方稳定版)
# 来源: https://googlechromelabs.github.io/chrome-for-testing/
#
# 仅支持 AMD64 (x86_64)
#

set -e

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

# 检查架构
ARCH=$(uname -m)
if [[ "$ARCH" != "x86_64" ]]; then
    log_error "Chrome Headless Shell 仅支持 AMD64 (x86_64)"
    log_info "当前架构: $ARCH"
    log_info "ARM64 请使用: sudo apt install chromium-browser"
    exit 1
fi

# 检查依赖
for cmd in curl jq unzip; do
    if ! command -v $cmd &> /dev/null; then
        log_info "安装缺少的依赖: $cmd"
        sudo apt-get update -qq
        sudo apt-get install -y -qq $cmd
    fi
done

log_info "获取最新稳定版本信息..."

# 获取最新版本
VERSION_JSON=$(curl -s "https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json")
VERSION=$(echo "$VERSION_JSON" | jq -r '.channels.Stable.version')
DOWNLOAD_URL=$(echo "$VERSION_JSON" | jq -r '.channels.Stable.downloads["chrome-headless-shell"][] | select(.platform == "linux64") | .url')

if [[ -z "$VERSION" || -z "$DOWNLOAD_URL" ]]; then
    log_error "无法获取版本信息"
    exit 1
fi

log_info "最新版本: $VERSION"
log_info "下载地址: $DOWNLOAD_URL"

# 下载
log_info "下载中..."
cd /tmp
curl -L -o chrome-headless-shell.zip "$DOWNLOAD_URL"

# 解压
log_info "解压中..."
rm -rf chrome-headless-shell-linux64
unzip -o chrome-headless-shell.zip

# 安装依赖
log_info "安装依赖..."
sudo apt-get update -qq

# 检测 Ubuntu 版本，24.04+ 部分包名有变化
UBUNTU_VERSION=$(lsb_release -rs 2>/dev/null || echo "22.04")
log_info "检测到系统版本: Ubuntu $UBUNTU_VERSION"

# Ubuntu 24.04+ 包名有 t64 后缀变化
if [[ "$UBUNTU_VERSION" == "24.04"* ]] || [[ "${UBUNTU_VERSION%%.*}" -ge 24 ]]; then
    log_info "使用 Ubuntu 24.04+ 依赖包名..."
    DEPS="libnspr4 libnss3 libexpat1 libfontconfig1 libuuid1 \
        libatk1.0-0t64 libatk-bridge2.0-0t64 libatspi2.0-0t64 \
        libcups2t64 libdrm2 libxkbcommon0 \
        libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1 \
        libpango-1.0-0 libcairo2 libx11-6 libx11-xcb1 libxcb1 \
        libxext6 libxshmfence1 libglib2.0-0t64 libdbus-1-3 \
        fonts-liberation fonts-noto-cjk fonts-wqy-zenhei \
        libvulkan1 xdg-utils ca-certificates libasound2t64"
else
    log_info "使用 Ubuntu 22.04 及更早版本依赖包名..."
    DEPS="libnspr4 libnss3 libexpat1 libfontconfig1 libuuid1 \
        libatk1.0-0 libatk-bridge2.0-0 libatspi2.0-0 \
        libcups2 libdrm2 libxkbcommon0 \
        libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1 \
        libpango-1.0-0 libcairo2 libx11-6 libx11-xcb1 libxcb1 \
        libxext6 libxshmfence1 libglib2.0-0 libdbus-1-3 \
        fonts-liberation fonts-noto-cjk fonts-wqy-zenhei \
        libvulkan1 xdg-utils ca-certificates libasound2"
fi

# 安装依赖
sudo apt-get install -y $DEPS 2>/dev/null || {
    log_warn "部分依赖安装失败，尝试逐个安装..."
    for pkg in $DEPS; do
        sudo apt-get install -y $pkg 2>/dev/null || log_warn "跳过: $pkg"
    done
}

# 安装
log_info "安装到 /opt/chrome-headless-shell..."
sudo rm -rf /opt/chrome-headless-shell
sudo mv chrome-headless-shell-linux64 /opt/chrome-headless-shell

# 创建符号链接 (chromedp 会查找 headless-shell)
sudo ln -sf /opt/chrome-headless-shell/chrome-headless-shell /usr/local/bin/headless-shell

# 清理
rm -f chrome-headless-shell.zip

# 验证
log_info "验证安装..."
if headless-shell --version; then
    log_success "Chrome Headless Shell 安装完成!"
    echo ""
    echo "版本: $VERSION"
    echo "路径: /opt/chrome-headless-shell/chrome-headless-shell"
    echo "链接: /usr/local/bin/headless-shell"
    echo ""
    log_info "Shelley 会自动检测并使用 headless-shell"
else
    log_error "安装验证失败"
    exit 1
fi
