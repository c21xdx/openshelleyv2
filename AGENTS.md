# AGENTS.md - AI Agent 工作指南

## 我的背景
- 业余编程者
- 主要写脚本和小项目
- 希望代码能跑起来就行

## 工作方式
1. 先理解需求，不清楚就问
2. 写代码后自己运行测试，确保能跑
3. 遇到错误自己调试修复，不要问我
4. 代码要简单易懂，加中文注释
5. 完成后告诉我怎么用

## 语言偏好
- Go, Python, Node.js 都可以
- 脚本优先用 Bash 或 Python
- 选择你认为最适合任务的语言

## 代码风格
- 简单优先，不要过度设计
- 用常见的库和方法
- 错误信息要清晰

## 不要
- 不要问太多确认问题，直接做
- 不要写复杂的架构
- 不要用太新或太偏门的技术

---

## 🚫 忽略这些目录和文件（不要读取）

### 依赖包目录
- `node_modules/` - Node.js 依赖
- `vendor/` - Go vendor
- `venv/` / `.venv/` / `env/` - Python 虚拟环境
- `__pycache__/` / `*.pyc` - Python 缓存
- `.pip/` / `site-packages/`
- `packages/` / `bower_components/`

### 构建产物
- `dist/` / `build/` / `out/` / `target/`
- `*.min.js` / `*.min.css` - 压缩文件
- `*.bundle.js` / `*.chunk.js`
- `coverage/` - 测试覆盖率
- `.next/` / `.nuxt/` / `.output/`

### 日志和临时文件
- `*.log` / `logs/`
- `tmp/` / `temp/` / `.tmp/`
- `*.swp` / `*.swo` / `*~`
- `.cache/` / `.temp/`

### IDE 和系统文件
- `.idea/` / `.vscode/` (除了 settings.json)
- `.git/objects/` / `.git/logs/`
- `.DS_Store` / `Thumbs.db`
- `*.iml`

### 数据文件
- `*.db` / `*.sqlite` / `*.sqlite3` - 数据库文件内容
- `*.db-shm` / `*.db-wal`
- 大型 JSON/CSV 数据文件 (>100KB)

### 其他
- `*.lock` 文件内容 (package-lock.json, yarn.lock, go.sum 等)
- `.env.local` / `.env.*.local` - 本地环境变量
- `*.map` - Source maps

---

## 💡 减少 Token 消耗的技巧

### 读取文件时
1. 先用 `ls` 或 `find` 了解目录结构，不要盲目 cat
2. 用 `head -50` 或 `tail -50` 预览大文件
3. 用 `grep` 搜索关键内容，不要读取整个文件
4. 用 `wc -l` 检查文件行数再决定是否读取

### 代码搜索
```bash
# ✅ 好：精确搜索
grep -rn "functionName" --include="*.go" .

# ❌ 差：读取所有文件
cat **/*.go
```

### 查看项目结构
```bash
# ✅ 好：排除不需要的目录
find . -type f -name "*.py" | grep -v -E "(venv|__pycache__|node_modules)"

# ✅ 好：只看目录结构
ls -la
tree -L 2 -I "node_modules|venv|__pycache__|.git"
```

### 输出控制
```bash
# ✅ 限制输出行数
command | head -100

# ✅ 只看错误
command 2>&1 | tail -20
```

### 编辑文件
- 使用 patch 工具精确修改，不要重写整个文件
- 只展示修改的部分，不要输出完整文件

---

## 📁 项目类型判断

快速判断项目类型：
```bash
# 检查项目类型
ls -la | head -20
```

| 文件 | 项目类型 |
|------|----------|
| `package.json` | Node.js |
| `go.mod` | Go |
| `requirements.txt` / `setup.py` / `pyproject.toml` | Python |
| `Cargo.toml` | Rust |
| `pom.xml` / `build.gradle` | Java |
| `Makefile` | 查看构建命令 |
| `Dockerfile` | 容器化项目 |

---

## ✅ 工作流程

1. **了解项目** (省 token)
   ```bash
   ls -la
   cat README.md 2>/dev/null | head -50
   ```

2. **找到入口**
   ```bash
   # 找主文件
   ls *.go main.* app.* index.* 2>/dev/null
   ```

3. **写代码** → **运行测试** → **修复错误** (循环直到成功)

4. **完成后**
   - 告诉我怎么运行
   - 简要说明做了什么
