# fasaxi-linker (Go版本) API 接口文档

## 基础信息

- **服务器地址**: `http://localhost:9090`
- **API版本**: v1
- **数据格式**: JSON
- **字符编码**: UTF-8

## 通用响应格式

所有API接口都遵循统一的响应格式：

```json
{
  "success": true,
  "data": {}, // 具体数据内容
  "errorMessage": "" // 错误信息，仅在success为false时存在
}
```

## 系统接口

### 1. 获取版本信息

**接口**: `GET /api/version`

**描述**: 获取系统版本信息

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tag": "stable",
    "version": "0.0.1-go",
    "needUpdate": false
  }
}
```

### 2. 系统更新

**接口**: `GET /api/update`

**描述**: 检查系统更新（当前为空实现）

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

## 配置管理接口

### 1. 获取配置列表

**接口**: `GET /api/config/list`

**描述**: 获取所有配置列表

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "name": "config1",
      "detail": "配置详细内容"
    }
  ]
}
```

### 2. 获取默认配置模板

**接口**: `GET /api/config/default`

**描述**: 获取默认配置模板

**响应示例**:
```json
{
  "success": true,
  "data": "/**\n * @type {import('@hlink/core').IConfig}\n */\nexport default {\n  // Add your config here\n}"
}
```

### 3. 获取指定配置

**接口**: `GET /api/config?name={configName}`

**参数**:
- `name` (string, required): 配置名称

**响应示例**:
```json
{
  "success": true,
  "data": {
    "name": "config1",
    "detail": "配置详细内容"
  }
}
```

### 4. 获取配置详细信息

**接口**: `GET /api/config/detail?name={configName}`

**参数**:
- `name` (string, required): 配置名称

**描述**: 解析JavaScript配置并返回配置对象

**响应示例**:
```json
{
  "success": true,
  "data": {
    "name": "media-sync",
    "type": "main",
    "pathsMapping": {
      "/source": ["/dest1", "/dest2"]
    },
    "include": ["*.jpg", "*.png"],
    "exclude": ["*.tmp"]
  }
}
```

### 5. 添加配置

**接口**: `POST /api/config`

**请求体**:
```json
{
  "name": "new-config",
  "detail": "配置详细内容"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 6. 更新配置

**接口**: `PUT /api/config`

**请求体**:
```json
{
  "preName": "old-config-name",
  "name": "new-config-name",
  "detail": "更新的配置内容"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 7. 删除配置

**接口**: `DELETE /api/config?name={configName}`

**参数**:
- `name` (string, required): 要删除的配置名称

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

## 任务管理接口

### 1. 获取任务列表

**接口**: `GET /api/task/list`

**描述**: 获取所有任务列表，包含监听状态

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "name": "task1",
      "type": "main",
      "pathsMapping": [
        {
          "source": "/source/path",
          "dest": "/dest/path"
        }
      ],
      "include": ["*.jpg"],
      "exclude": ["*.tmp"],
      "saveMode": 0,
      "openCache": true,
      "mkdirIfSingle": false,
      "deleteDir": false,
      "keepDirStruct": true,
      "scheduleType": "",
      "scheduleValue": "",
      "reverse": false,
      "config": "",
      "isWatching": false
    }
  ]
}
```

### 2. 获取指定任务

**接口**: `GET /api/task?name={taskName}`

**参数**:
- `name` (string, required): 任务名称

**响应示例**:
```json
{
  "success": true,
  "data": {
    "name": "task1",
    "type": "main",
    "pathsMapping": [
      {
        "source": "/source/path",
        "dest": "/dest/path"
      }
    ],
    "include": ["*.jpg"],
    "exclude": ["*.tmp"],
    "saveMode": 0,
    "openCache": true,
    "mkdirIfSingle": false,
    "deleteDir": false,
    "keepDirStruct": true,
    "scheduleType": "",
    "scheduleValue": "",
    "reverse": false,
    "config": "",
    "isWatching": false
  }
}
```

### 3. 创建任务

**接口**: `POST /api/task`

**请求体**:
```json
{
  "name": "new-task",
  "type": "main",
  "pathsMapping": [
    {
      "source": "/source/path",
      "dest": "/dest/path"
    }
  ],
  "include": ["*.jpg", "*.png"],
  "exclude": ["*.tmp"],
  "saveMode": 0,
  "openCache": true,
  "mkdirIfSingle": false,
  "deleteDir": false,
  "keepDirStruct": true,
  "scheduleType": "",
  "scheduleValue": "",
  "reverse": false,
  "config": ""
}
```

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 4. 更新任务

**接口**: `PUT /api/task`

**请求体**:
```json
{
  "preName": "old-task-name",
  "name": "new-task-name",
  "type": "main",
  "pathsMapping": [
    {
      "source": "/new/source/path",
      "dest": "/new/dest/path"
    }
  ],
  "include": ["*.jpg", "*.png"],
  "exclude": ["*.tmp"],
  "saveMode": 0,
  "openCache": true,
  "mkdirIfSingle": false,
  "deleteDir": false,
  "keepDirStruct": true,
  "scheduleType": "",
  "scheduleValue": "",
  "reverse": false,
  "config": ""
}
```

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 5. 删除任务

**接口**: `DELETE /api/task?name={taskName}`

**参数**:
- `name` (string, required): 要删除的任务名称

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

## 任务执行接口

### 1. 运行任务

**接口**: `GET /api/task/run?name={taskName}`

**参数**:
- `name` (string, required): 要运行的任务名称

**描述**: 执行指定任务，支持Server-Sent Events (SSE)实时推送执行日志

**响应格式**: Server-Sent Events流

**事件格式**:
```
data: {"status":"ongoing","type":"main","output":"[INFO] 开始处理文件..."}

data: {"status":"succeed","type":"main","output":"Done"}

data: {"status":"failed","type":"main","output":"Error: 处理失败"}
```

**状态类型**:
- `ongoing`: 任务进行中
- `succeed`: 任务成功完成
- `failed`: 任务执行失败

**任务类型**:
- `main`: 主任务（创建硬链接）
- `prune`: 修剪任务（删除无效硬链接）

## 文件监听接口

### 1. 开始监听

**接口**: `POST /api/task/watch/start`

**请求体**:
```json
{
  "name": "task-name"
}
```

**描述**: 开始监听指定任务的文件变化

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 2. 停止监听

**接口**: `POST /api/task/watch/stop`

**请求体**:
```json
{
  "name": "task-name"
}
```

**描述**: 停止监听指定任务的文件变化

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 3. 获取监听状态

**接口**: `GET /api/task/watch/status?name={taskName}`

**参数**:
- `name` (string, required): 任务名称

**响应示例**:
```json
{
  "success": true,
  "data": true // true表示正在监听，false表示未监听
}
```

## 缓存管理接口

### 1. 获取缓存内容

**接口**: `GET /api/cache/`

**描述**: 获取系统缓存文件内容

**响应示例**:
```json
{
  "success": true,
  "data": "[\"/path/to/file1.jpg\", \"/path/to/file2.png\"]"
}
```

### 2. 更新缓存内容

**接口**: `PUT /api/cache/`

**请求体**:
```json
{
  "content": "[\"/path/to/file1.jpg\", \"/path/to/file2.png\"]"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

### 3. 获取缓存日志

**接口**: `GET /api/cache/log`

**描述**: 获取缓存相关的日志内容

**响应示例**:
```json
{
  "success": true,
  "data": "2023-12-10 15:30:00 [INFO] 缓存更新\n2023-12-10 15:30:01 [INFO] 文件处理完成"
}
```

### 4. 清空缓存日志

**接口**: `DELETE /api/cache/log`

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

## 日志管理接口

### 1. 获取任务日志

**接口**: `GET /api/task/log?name={taskName}`

**参数**:
- `name` (string, required): 任务名称

**响应示例**:
```json
{
  "success": true,
  "data": "2023-12-10 15:30:00 [INFO] 任务开始执行\n2023-12-10 15:30:01 [SUCCEED] 处理文件: /path/to/file.jpg\n2023-12-10 15:30:02 [INFO] 任务执行完成"
}
```

### 2. 清空任务日志

**接口**: `DELETE /api/task/log?name={taskName}`

**参数**:
- `name` (string, required): 任务名称

**响应示例**:
```json
{
  "success": true,
  "data": true
}
```

## 数据模型

### Task 任务模型

```json
{
  "name": "string",           // 任务名称
  "type": "string",           // 任务类型: "main" 或 "prune"
  "pathsMapping": [           // 路径映射数组
    {
      "source": "string",     // 源路径
      "dest": "string"        // 目标路径
    }
  ],
  "include": ["string"],      // 包含文件模式
  "exclude": ["string"],      // 排除文件模式
  "saveMode": "number",       // 保存模式: 0-保持目录结构, 1-扁平化
  "openCache": "boolean",     // 是否启用缓存
  "mkdirIfSingle": "boolean", // 单文件时是否创建目录
  "deleteDir": "boolean",     // 是否删除目录（prune任务）
  "keepDirStruct": "boolean", // 是否保持目录结构
  "scheduleType": "string",   // 调度类型（可选）
  "scheduleValue": "string",  // 调度值（可选）
  "reverse": "boolean",       // 是否反向（prune任务）
  "config": "string"          // 关联配置名称
}
```

### Config 配置模型

```json
{
  "name": "string",   // 配置名称
  "detail": "string"  // 配置详细内容（JavaScript代码）
}
```

## 错误处理

所有接口在发生错误时都会返回统一的错误格式：

```json
{
  "success": false,
  "errorMessage": "错误描述信息"
}
```

### 常见错误码

- **配置不存在**: "Config not found"
- **任务不存在**: "Task not found"
- **参数错误**: "参数验证失败"
- **系统错误**: 具体的系统错误信息

## 缓存文件位置

缓存文件默认存储在以下位置：
- **macOS/Linux**: `~/.hlink/cache-array.json`
- **自定义路径**: 通过环境变量 `HLINK_HOME` 指定

缓存日志文件位置：
- **macOS/Linux**: `~/.hlink/serve.log`
- **自定义路径**: 通过环境变量 `HLINK_HOME` 指定

## 使用示例

### 创建并运行一个完整任务的流程

```bash
# 1. 创建任务
curl -X POST http://localhost:9090/api/task \
  -H "Content-Type: application/json" \
  -d '{
    "name": "photo-sync",
    "type": "main",
    "pathsMapping": [
      {
        "source": "/Users/mac/Documents/Photos",
        "dest": "/backup/photos"
      }
    ],
    "include": ["*.jpg", "*.png"],
    "exclude": ["*.tmp"],
    "keepDirStruct": true,
    "openCache": true
  }'

# 2. 运行任务（SSE流）
curl -N http://localhost:9090/api/task/run?name=photo-sync

# 3. 开始监听文件变化
curl -X POST http://localhost:9090/api/task/watch/start \
  -H "Content-Type: application/json" \
  -d '{"name": "photo-sync"}'

# 4. 查看任务日志
curl http://localhost:9090/api/task/log?name=photo-sync

# 5. 查看缓存内容
curl http://localhost:9090/api/cache/

# 6. 更新缓存内容
curl -X PUT -H "Content-Type: application/json" \
  -d '{"content":"[\"/path/to/file1.jpg\", \"/path/to/file2.png\"]"}' \
  http://localhost:9090/api/cache/

# 7. 查看缓存日志
curl http://localhost:9090/api/cache/log
```

## 注意事项

1. **CORS支持**: 服务器已配置CORS，支持跨域请求
2. **SSE流**: 任务执行接口使用Server-Sent Events，需要客户端支持流式处理
3. **文件路径**: 所有路径都使用绝对路径
4. **权限**: 确保服务器进程有足够的文件系统权限
5. **并发**: 多个任务可以并发执行，但同一任务同时只能有一个实例运行

## 更新日志

- **v0.0.1-go**: 初始版本，支持基本的任务管理、配置管理和文件监听功能
