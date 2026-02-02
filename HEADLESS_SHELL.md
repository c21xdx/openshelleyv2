# Chrome Headless Shell 分支

> 此分支使用 Google 官方维护的 Chrome Headless Shell，替代 chromium-browser

## 什么是 Chrome Headless Shell

Google 官方为自动化测试提供的精简版 Chrome，特点：

- ✅ 官方维护，跟随 Chrome 稳定版发布
- ✅ 比完整 Chromium 更小
- ✅ 专为 headless 自动化设计
- ❌ 仅支持 AMD64，不支持 ARM64

来源: https://googlechromelabs.github.io/chrome-for-testing/

## 对比

| | chromium-browser | chrome-headless-shell |
|---|-----------------|----------------------|
| 来源 | 系统包管理 | Google 官方 |
| 更新频率 | 随系统 | 随 Chrome 发布 |
| 大小 | ~400 MB | ~150 MB |
| ARM64 | ✅ | ❌ |
| 稳定性 | ✅ | ✅ |

## 安装

### 一键安装

```bash
curl -sSL https://raw.githubusercontent.com/c21xdx/openshelley/headless-shell/install-headless-shell.sh | bash
```

### 手动安装

```bash
# 1. 获取最新版本
VERSION=$(curl -s "https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json" | jq -r '.channels.Stable.version')

# 2. 下载
cd /tmp
curl -L -o chrome-headless-shell.zip \
  "https://storage.googleapis.com/chrome-for-testing-public/${VERSION}/linux64/chrome-headless-shell-linux64.zip"

# 3. 解压并安装
unzip chrome-headless-shell.zip
sudo mv chrome-headless-shell-linux64 /opt/chrome-headless-shell
sudo ln -sf /opt/chrome-headless-shell/chrome-headless-shell /usr/local/bin/headless-shell

# 4. 安装依赖
sudo apt install -y libnspr4 libnss3 libexpat1 libfontconfig1 libasound2

# 5. 验证
headless-shell --version
```

## Shelley 自动检测

Shelley (chromedp) 会按以下顺序查找浏览器：

1. `headless-shell` ← 优先
2. `chromium`
3. `chromium-browser`
4. `google-chrome`

安装后无需配置，Shelley 会自动使用 `headless-shell`。

## 适用场景

- ✅ AMD64 (x86_64) 服务器
- ✅ 希望使用官方稳定版本
- ✅ 不想使用系统包管理的 Chromium

## 不适用场景

- ❌ ARM64 (aarch64) - 请使用 `apt install chromium-browser`
- ❌ 需要 GUI 浏览器

## 升级

重新运行安装脚本即可：

```bash
./install-headless-shell.sh
```

## 故障排查

### 缺少依赖

```bash
# 检查缺少的库
ldd /opt/chrome-headless-shell/chrome-headless-shell | grep "not found"

# 安装常见依赖
sudo apt install -y libnspr4 libnss3 libexpat1 libfontconfig1 libasound2 \
  libatk1.0-0 libatk-bridge2.0-0 libcups2 libdrm2 libxkbcommon0 \
  libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1
```

### 测试浏览器

```bash
# 测试 headless 模式
headless-shell --headless --disable-gpu --screenshot=/tmp/test.png https://example.com
ls -la /tmp/test.png
```
