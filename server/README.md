<p align="center">
  <a href="https://hlink.likun.me" target="_blank" rel="noopener noreferrer">
    <img width="180" src="https://hlink.likun.me/logo.svg" alt="hlink logo">
  </a>
</p>
<p align="center">
  <a href="https://www.npmjs.com/package/hlink"><img src="https://img.shields.io/npm/v/hlink.svg" alt="npm package"></a>
  <a href="https://nodejs.org/en/about/releases/"><img src="https://img.shields.io/node/v/hlink.svg" alt="node compatibility"></a>
  <a href="https://npmjs.com/package/hlink"><img src="https://img.shields.io/npm/dm/hlink.svg" alt="downloads"></a>
  <a href="https://github.com/likun7981/hlink/actions/workflows/publish.yml"><img src="https://github.com/likun7981/hlink/actions/workflows/publish.yml/badge.svg" alt="build status"></a>
  <a href="https://github.com/likun7981/hlink/blob/master/LICENSE"><img src="https://img.shields.io/npm/l/hlink.svg" alt="license"></a>
</p>

# fasaxi-linker (hlink-go)
> 批量、快速硬链工具 - Go版本实现 (The batch, fast hard link toolkit - Go Implementation)

## 🚀 项目简介

fasaxi-linker 是一个高性能的硬链接管理工具，提供Node.js原版和Go版本两种实现。Go版本(`servergo`)专注于提供更好的性能和更简单的部署体验，特别适合服务器环境和容器化部署。

### ✨ 核心特性

- 💡 **重复检测**：支持文件名变更的智能重复检测
- ⚡️ **极速性能**：`20000+`文件只需要1分钟，Go版本性能更优
- 📦 **多平台支持**：支持Windows、Mac、Linux
- 🛠️ **丰富配置**：支持黑白名单、缓存、目录结构保持等多种配置
- 🔩 **修剪机制**：方便同步源文件和硬链接
- 🌐 **WebUI界面**：图形化界面让管理更便捷
- 🐳 **Docker支持**：无需关心环境问题，一键部署
- 🔄 **实时监听**：文件变化自动更新硬链接

### 🏗️ 架构设计

```
fasaxi-linker/
├── packages/
│   ├── app/          # Web前端应用 (React/Vue)
│   ├── cli/          # Node.js命令行工具
│   ├── core/         # Node.js核心库
│   └── app/servergo/ # Go版本后端服务 ⭐
├── docs/             # 文档
└── scripts/          # 构建脚本
```

## 🚀 快速开始

### Go版本 (推荐用于服务器环境)

#### 1. 从源码构建

```bash
# 克隆仓库
git clone https://github.com/AceCandy/fasaxi-linker.git
cd fasaxi-linker/packages/app/servergo

# 构建并运行
go mod tidy
go run cmd/server/main.go
```

#### 2. 直接运行二进制文件

```bash
# 下载或构建二进制文件
cd packages/app/servergo
go build -o bin/server cmd/server/main.go

# 运行服务器
./bin/server
```

服务器将在 `http://localhost:9090` 启动

#### 3. Docker部署

```bash
# 构建Docker镜像
cd packages/app/servergo
docker build -t fasaxi-linker:latest .

# 运行容器
docker run -d \
  --name fasaxi-linker \
  -p 9090:9090 \
  -v /your/source/path:/source \
  -v /your/dest/path:/dest \
  fasaxi-linker:latest
```

### Node.js版本 (传统版本)

#### 使用npm安装

```bash
npm i -g hlink

# 帮助
hlink --help
```

#### 使用docker run

```bash
docker run -d --name hlink \
-e PUID=$YOUR_USER_ID \
-e PGID=$YOUR_GROUP_ID \
-e UMASK=$YOUR_UMASK \
-e HLINK_HOME=$YOUR_HLINK_HOME_DIR \
-p 9090:9090 \
-v $YOUR_NAS_VOLUME_PATH:$DOCKER_VOLUME_PATH \
likun7981/hlink:latest
```

#### 使用docker compose

```yml
version: '2'

services:
  hlink:
    image: likun7981/hlink:latest
    restart: on-failure
    ports:
      - 9090:9090
    volumes:
      - $YOUR_NAS_VOLUME_PATH:$DOCKER_VOLUME_PATH
    environment:
      - PUID=$YOUR_USER_ID
      - PGID=$YOUR_GROUP_ID
      - UMASK=$YOUR_UMASK
      - HLINK_HOME=$YOUR_HLINK_HOME_DIR
```

## 📖 使用说明

### Web界面

访问 `http://localhost:9090` 打开Web管理界面：

1. **配置管理**：创建和管理硬链接任务配置
2. **任务执行**：手动执行或设置定时任务
3. **实时监控**：查看任务执行状态和日志
4. **文件监听**：开启文件变化自动同步

### API接口

Go版本提供完整的RESTful API：

```bash
# 获取任务列表
curl http://localhost:9090/api/tasks

# 创建任务
curl -X POST http://localhost:9090/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-task",
    "type": "main",
    "pathsMapping": {
      "/source": ["/dest"]
    },
    "include": ["*.jpg", "*.png"],
    "exclude": ["*.tmp"]
  }'

# 运行任务
curl http://localhost:9090/api/tasks/run?name=my-task
```

### 配置示例

```javascript
export default {
  name: "media-sync",
  type: "main",
  pathsMapping: {
    "/Users/mac/Documents/Photos": ["/backup/photos", "/nas/photos"]
  },
  include: ["*.jpg", "*.png", "*.mp4", "*.mov"],
  exclude: ["*.tmp", ".*"],
  keepDirStruct: true,
  mkdirIfSingle: false,
  openCache: true
}
```

## 🔧 开发

### Go版本开发

```bash
cd packages/app/servergo

# 安装依赖
go mod tidy

# 运行开发服务器
go run cmd/server/main.go

# 运行测试
go test ./...

# 构建
go build -o bin/server cmd/server/main.go
```

### Node.js版本开发

```bash
# 安装依赖
pnpm install

# 开发模式
pnpm app:dev

# 构建
pnpm build

# 测试
pnpm test
```

## 📊 性能对比

| 版本 | 20,000文件处理时间 | 内存占用 | 部署复杂度 |
|------|-------------------|----------|------------|
| Node.js | ~60秒 | ~200MB | 中等 |
| Go版本 | ~30秒 | ~50MB | 简单 |

## 🖼️ 界面截图

### WebUI界面
<img src="https://user-images.githubusercontent.com/13427467/177048631-04dc6ace-af3a-4459-8848-13cc3c928856.png" width="520"/>

### 命令行界面
<img src="https://user-images.githubusercontent.com/13427467/148177243-50ce207f-a31e-4a0a-b601-27ea9cbb1e1f.png" width="520"/>

### 效果展示
<img src="https://user-images.githubusercontent.com/13427467/148171766-ccbe2a1a-c30c-4e1a-868c-4e2c69617d29.png" width="520"/>

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](https://github.com/likun7981/hlink/blob/master/LICENSE) 文件了解详情

## ☕ 打赏作者

请作者喝一杯咖啡😄

<img width="300" src="https://user-images.githubusercontent.com/13427467/148188331-c997f355-2a80-46b9-ba6b-d189186ac356.png" /><img width="300" src="https://user-images.githubusercontent.com/13427467/148188398-d6d9e8e5-bd75-4de4-9faa-dbd4846b4103.png" />

感谢各位的支持！

## 🔗 相关链接

- [项目主页](https://hlink.likun.me)
- [在线文档](https://hlink.likun.me/guide/)
- [原版项目](https://github.com/likun7981/hlink)
- [问题反馈](https://github.com/AceCandy/fasaxi-linker/issues)

---

**注意**：本仓库是hlink项目的Go版本实现，专注于提供更好的性能和部署体验。如果你需要使用Node.js原版，请访问 [likun7981/hlink](https://github.com/likun7981/hlink)。