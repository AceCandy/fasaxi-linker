<p align="center">
  <a href="https://hlink.likun.me" target="_blank" rel="noopener noreferrer">
    <img width="180" src="https://hlink.likun.me/logo.svg" alt="hlink logo">
  </a>
</p>

<h1 align="center">HLink - 智能硬链接管理系统</h1>

<p align="center">
  <a href="https://www.npmjs.com/package/hlink"><img src="https://img.shields.io/npm/v/hlink.svg" alt="npm version"></a>
  <a href="https://nodejs.org/en/about/releases/"><img src="https://img.shields.io/node/v/hlink.svg" alt="node compatibility"></a>
  <a href="https://npmjs.com/package/hlink"><img src="https://img.shields.io/npm/dm/hlink.svg" alt="downloads"></a>
  <a href="https://github.com/likun7981/hlink/blob/master/LICENSE"><img src="https://img.shields.io/npm/l/hlink.svg" alt="license"></a>
</p>

<p align="center">
  <strong>高性能、企业级的硬链接批量管理工具</strong><br>
  为 NAS、媒体库、备份系统提供专业的文件同步解决方案
</p>

---

## 📖 目录

- [核心特性](#-核心特性)
- [技术架构](#-技术架构)
- [快速开始](#-快速开始)
- [功能详解](#-功能详解)
- [配置指南](#-配置指南)
- [API 文档](#-api-文档)
- [开发指南](#-开发指南)
- [性能优化](#-性能优化)
- [常见问题](#-常见问题)
- [贡献指南](#-贡献指南)

---

## ✨ 核心特性

### 🎯 智能管理
- **智能重复检测** - 基于文件内容的 MD5 校验，精准识别重复文件
- **文件名变更追踪** - 即使文件名改变，依然能准确识别并建立硬链接
- **增量同步** - 仅处理变更文件，大幅提升处理效率
- **缓存机制** - 智能缓存文件信息，避免重复扫描

### ⚡ 极致性能
- **并发处理** - 多线程并发扫描和链接，充分利用系统资源
- **批量操作** - 20,000+ 文件仅需 1 分钟完成处理
- **内存优化** - 流式处理大文件，内存占用低
- **增量更新** - 仅处理变更部分，节省时间和资源

### 🎨 现代化界面
- **响应式设计** - 完美适配桌面、平板、移动设备
- **实时反馈** - 任务执行状态实时更新，进度一目了然
- **可视化配置** - 图形化配置界面，无需编写代码
- **优雅交互** - 流畅的动画效果和直观的操作体验

### 🔧 灵活配置
- **多路径映射** - 一个源目录可映射到多个目标目录
- **规则过滤** - 支持 glob 模式的包含/排除规则
- **目录结构** - 可选择保持或扁平化目录结构
- **自定义配置** - 丰富的配置选项满足各种场景需求

### 🔄 自动化运维
- **实时监控** - 文件系统监听，自动同步变更
- **定时任务** - 支持 Cron 表达式的定时执行
- **任务队列** - 智能任务调度，避免资源冲突
- **日志追踪** - 详细的操作日志，便于问题排查

### 🐳 部署友好
- **Docker 支持** - 一键部署，无需配置环境
- **跨平台** - 支持 Windows、macOS、Linux
- **轻量级** - 资源占用低，适合长期运行
- **零依赖** - 无需额外安装其他软件

---

## 🏗️ 技术架构

### 前端技术栈
```
Vue 3 + TypeScript + Vuetify 3
├── 组件化架构 - 高度模块化的组件设计
├── 状态管理 - Pinia 全局状态管理
├── 路由管理 - Vue Router 单页应用路由
├── UI 框架 - Vuetify 3 Material Design
└── 构建工具 - Vite 极速开发体验
```

### 后端技术栈
```
Node.js / Go (双实现)
├── RESTful API - 标准化的 API 接口
├── 文件系统 - 高性能文件操作
├── 任务调度 - 智能任务队列管理
├── 缓存系统 - 文件信息缓存优化
└── 日志系统 - 结构化日志记录
```

### 项目结构
```
hlink-monorepo/
├── packages/
│   ├── app/                    # Web 应用
│   │   ├── client/            # 前端应用
│   │   │   ├── components/    # Vue 组件
│   │   │   │   ├── TaskItem.vue       # 任务卡片组件
│   │   │   │   ├── TaskEditor.vue     # 任务编辑器
│   │   │   │   ├── TaskList.vue       # 任务列表
│   │   │   │   ├── ConfigEditor.vue   # 配置编辑器
│   │   │   │   └── ConfigList.vue     # 配置列表
│   │   │   ├── pages/         # 页面组件
│   │   │   ├── stores/        # Pinia 状态管理
│   │   │   ├── router/        # 路由配置
│   │   │   └── composables/   # 组合式函数
│   │   ├── server/            # Node.js 后端
│   │   └── servergo/          # Go 后端实现
│   ├── cli/                   # 命令行工具
│   └── core/                  # 核心库
├── docs/                      # 文档
└── scripts/                   # 构建脚本
```

---

## 🚀 快速开始

### 方式一：使用 NPM（推荐）

```bash
# 全局安装
npm install -g hlink

# 启动 Web 服务
hlink serve

# 访问 Web 界面
open http://localhost:9090
```

### 方式二：使用 Docker

```bash
# 使用 Docker Run
docker run -d \
  --name hlink \
  -p 9090:9090 \
  -v /your/source:/source \
  -v /your/dest:/dest \
  -e PUID=1000 \
  -e PGID=1000 \
  likun7981/hlink:latest

# 使用 Docker Compose
cat > docker-compose.yml <<EOF
version: '3'
services:
  hlink:
    image: likun7981/hlink:latest
    container_name: hlink
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - /your/source:/source
      - /your/dest:/dest
    environment:
      - PUID=1000
      - PGID=1000
      - UMASK=022
      - HLINK_HOME=/config
EOF

docker-compose up -d
```

### 方式三：从源码构建

```bash
# 克隆仓库
git clone https://github.com/likun7981/hlink.git
cd hlink

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建生产版本
pnpm build
```

---

## 🎯 功能详解

### 1. 配置管理

创建和管理硬链接规则配置，支持：

- **可视化配置编辑器** - 无需手写配置文件
- **配置模板** - 预设常用场景配置
- **配置验证** - 实时验证配置正确性
- **配置导入导出** - 方便配置迁移和备份

**配置示例：**
```javascript
{
  name: "media-library",
  description: "媒体库硬链接同步",
  include: ["*.mp4", "*.mkv", "*.avi"],
  exclude: ["*.tmp", ".*"],
  keepDirStruct: true,
  openCache: true,
  skipSameFile: true
}
```

### 2. 任务管理

创建和执行硬链接任务：

- **任务类型**
  - `硬链 (main)` - 创建硬链接
  - `同步 (prune)` - 清理失效链接

- **任务状态**
  - 待执行 - 任务已创建，等待执行
  - 执行中 - 任务正在处理
  - 已完成 - 任务执行成功
  - 失败 - 任务执行失败

- **任务操作**
  - 单次执行 - 立即执行一次
  - 实时监控 - 监听文件变化自动执行
  - 定时同步 - 按计划定时执行
  - 查看日志 - 查看详细执行日志
  - 缓存管理 - 管理任务缓存数据

### 3. 实时监控

开启文件系统监听，自动同步变更：

- **监听事件**
  - 文件创建 - 自动创建硬链接
  - 文件修改 - 更新硬链接
  - 文件删除 - 可选清理链接
  - 文件移动 - 同步移动链接

- **性能优化**
  - 防抖处理 - 避免频繁触发
  - 批量处理 - 合并多个变更
  - 智能过滤 - 忽略临时文件

### 4. 缓存管理

智能缓存系统提升性能：

- **缓存内容**
  - 文件 MD5 值
  - 文件大小和修改时间
  - 目录结构信息
  - 硬链接关系

- **缓存策略**
  - 增量更新 - 仅更新变更部分
  - 自动清理 - 清理过期缓存
  - 手动管理 - 支持手动清理缓存

### 5. 日志系统

详细的操作日志记录：

- **日志级别**
  - INFO - 一般信息
  - WARN - 警告信息
  - ERROR - 错误信息
  - DEBUG - 调试信息

- **日志功能**
  - 实时查看 - Web 界面实时显示
  - 日志搜索 - 快速定位问题
  - 日志导出 - 导出日志文件
  - 日志清理 - 定期清理旧日志

---

## ⚙️ 配置指南

### 基础配置

```javascript
export default {
  // 包含规则 - 支持 glob 模式
  include: [
    "**/*.mp4",      // 所有 mp4 文件
    "**/*.mkv",      // 所有 mkv 文件
    "Movies/**/*"    // Movies 目录下所有文件
  ],
  
  // 排除规则 - 支持 glob 模式
  exclude: [
    "**/*.tmp",      // 排除临时文件
    "**/.*",         // 排除隐藏文件
    "**/@eaDir/**"   // 排除群晖缩略图目录
  ],
  
  // 保持目录结构
  keepDirStruct: true,
  
  // 单文件时创建目录
  mkdirIfSingle: false,
  
  // 开启缓存
  openCache: true,
  
  // 跳过相同文件
  skipSameFile: true
}
```

### 高级配置

```javascript
export default {
  // 多路径映射
  pathsMapping: {
    "/source/movies": [
      "/dest1/movies",
      "/dest2/movies"
    ],
    "/source/tv": [
      "/dest1/tv"
    ]
  },
  
  // 自定义过滤函数
  filter: (file) => {
    // 只处理大于 100MB 的文件
    return file.size > 100 * 1024 * 1024
  },
  
  // 并发数控制
  concurrency: 10,
  
  // 缓存配置
  cache: {
    enabled: true,
    ttl: 86400,  // 缓存有效期（秒）
    maxSize: 1000 // 最大缓存条目数
  }
}
```

---

## 📡 API 文档

### 任务管理 API

#### 获取任务列表
```http
GET /api/task/list
```

#### 创建任务
```http
POST /api/task
Content-Type: application/json

{
  "name": "my-task",
  "type": "main",
  "config": "default",
  "pathsMapping": [
    {
      "source": "/source/path",
      "dest": "/dest/path"
    }
  ]
}
```

#### 执行任务
```http
GET /api/task/run?name=my-task&alive=0
```

#### 开启监控
```http
POST /api/task/watch/start
Content-Type: application/json

{
  "name": "my-task"
}
```

#### 停止监控
```http
POST /api/task/watch/stop
Content-Type: application/json

{
  "name": "my-task"
}
```

### 配置管理 API

#### 获取配置列表
```http
GET /api/config/list
```

#### 创建配置
```http
POST /api/config
Content-Type: application/json

{
  "name": "my-config",
  "description": "My configuration",
  "detail": {
    "include": ["*.mp4"],
    "exclude": ["*.tmp"]
  }
}
```

---

## 🛠️ 开发指南

### 环境要求

- Node.js >= 16.0.0
- pnpm >= 9.1.0
- Go >= 1.20 (可选，用于 Go 版本)

### 开发流程

```bash
# 1. 克隆项目
git clone https://github.com/likun7981/hlink.git
cd hlink

# 2. 安装依赖
pnpm install

# 3. 启动开发服务器
pnpm dev

# 4. 访问开发环境
open http://localhost:5173
```

### 项目脚本

```bash
# 开发模式
pnpm dev

# 构建生产版本
pnpm build

# 代码检查
pnpm lint

# 代码格式化
pnpm format

# 运行测试
pnpm test

# 清理依赖
pnpm clean
```

### 代码规范

- **TypeScript** - 使用 TypeScript 编写类型安全的代码
- **ESLint** - 遵循 ESLint 代码规范
- **Prettier** - 使用 Prettier 格式化代码
- **Commit** - 遵循 Conventional Commits 规范

---

## 📊 性能优化

### 性能指标

| 文件数量 | 处理时间 | 内存占用 | CPU 使用率 |
|---------|---------|---------|-----------|
| 1,000   | ~3秒    | ~50MB   | ~20%      |
| 10,000  | ~30秒   | ~150MB  | ~40%      |
| 20,000  | ~60秒   | ~200MB  | ~60%      |
| 50,000  | ~150秒  | ~300MB  | ~80%      |

### 优化建议

1. **启用缓存** - 大幅减少重复扫描时间
2. **合理配置过滤规则** - 减少不必要的文件处理
3. **调整并发数** - 根据系统性能调整并发处理数
4. **使用增量同步** - 仅处理变更文件
5. **定期清理缓存** - 避免缓存文件过大

---

## ❓ 常见问题

### Q: 硬链接和软链接有什么区别？
A: 硬链接直接指向文件数据，删除原文件不影响硬链接；软链接是文件路径的引用，删除原文件会导致软链接失效。

### Q: 为什么选择硬链接而不是复制？
A: 硬链接不占用额外磁盘空间，多个硬链接共享同一份数据，节省存储空间。

### Q: 跨文件系统可以创建硬链接吗？
A: 不可以，硬链接只能在同一文件系统内创建。

### Q: 如何处理文件名冲突？
A: 系统会自动跳过同名文件，或根据配置覆盖目标文件。

### Q: 缓存文件存储在哪里？
A: 默认存储在 `~/.hlink/cache-array.json`，可通过 `HLINK_HOME` 环境变量自定义。

---

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 贡献方式

1. **报告 Bug** - 提交 Issue 描述问题
2. **功能建议** - 提出新功能想法
3. **代码贡献** - 提交 Pull Request
4. **文档改进** - 完善项目文档
5. **测试反馈** - 提供使用反馈

### 提交流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

---

## 📄 开源协议

本项目采用 [MIT License](LICENSE) 开源协议。

---

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

特别感谢：
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Vuetify](https://vuetifyjs.com/) - Material Design 组件库
- [Vite](https://vitejs.dev/) - 下一代前端构建工具

---

## 📞 联系方式

- **项目主页**: https://hlink.likun.me
- **在线文档**: https://hlink.likun.me/guide/
- **问题反馈**: https://github.com/likun7981/hlink/issues
- **作者邮箱**: likun7981@gmail.com

---

<p align="center">
  <sub>Built with ❤️ by <a href="https://github.com/likun7981">likun7981</a></sub>
</p>
