# Redis Manager - ElasticView插件

## 📖 项目简介

Redis Manager 是一个基于 ElasticView 平台的 Redis 数据库管理插件，为开发者和管理员提供直观、高效的 Redis 数据库管理工具。支持 Redis 信息监控、内存分析、Key 管理等功能。

## ✨ 主要功能

### 🔍 Redis 信息总览
- **实时监控**：展示 Redis 服务器运行状态、连接数、内存使用等关键指标
- **性能指标**：监控 Redis 性能数据，包括命令执行统计、网络流量等
- **服务器信息**：显示 Redis 版本、运行时间、配置信息等
- **实时更新**：数据实时刷新，及时反映 Redis 状态变化

### 📊 内存分析
- **内存使用统计**：详细分析 Redis 内存使用情况
- **数据类型分布**：展示不同数据类型（String、Hash、List、Set、ZSet）的内存占用
- **内存碎片分析**：监控内存碎片化情况，提供优化建议
- **内存趋势图**：可视化展示内存使用趋势

### 🎯 Key 管理器
- **Key 浏览**：树形结构展示 Redis 中的所有 Key
- **Key 搜索**：支持模糊搜索和正则表达式匹配
- **Key 操作**：支持查看、编辑、删除、重命名 Key
- **批量操作**：支持批量删除、批量重命名等操作
- **Key 编辑器**：内置编辑器支持查看和编辑 Key 值
- **TTL 管理**：查看和设置 Key 的过期时间

### 🔧 高级功能
- **数据类型支持**：完整支持 String、Hash、List、Set、ZSet 等 Redis 数据类型
- **实时监控**：实时监控 Redis 状态变化
- **操作日志**：记录所有操作，便于审计和调试
- **多连接支持**：支持管理多个 Redis 实例

## 🛠️ 技术栈

### 后端
- **Go 1.23+**：高性能后端服务
- **Gin**：轻量级 Web 框架
- **Redis Go Client**：官方 Redis Go 客户端
- **Eve Plugin SDK**：ElasticView 插件开发SDK

### 前端
- **Vue 3**：渐进式 JavaScript 框架
- **Element Plus**：Vue 3 组件库
- **Monaco Editor**：代码编辑器
- **TypeScript**：类型安全的 JavaScript
- **Vue Router**：前端路由管理
- **Vue I18n**：国际化支持

## 📦 安装要求

### 环境要求
- **Go 版本**：>= 1.23
- **Node 版本**：>= 20.14.0
- **ElasticView**：已启动基座程序
- **Redis**：目标 Redis 服务器

### 开发工具安装
```bash
# 安装 ElasticView 插件开发工具
go install github.com/1340691923/eve-plugin-sdk-go/cmd/ev_plugin_cli@v0.0.20

# 安装 pnpm
npm install -g pnpm
```

## 🚀 快速开始

### 1. 下载依赖

```bash
# 检查项目依赖和环境（项目根目录运行）
ev_plugin_cli doctor 

# 使用 ev_plugin_cli 下载项目依赖
ev_plugin_cli install

# 或者手动安装依赖
# 后端：go mod tidy
# 前端：cd frontend && pnpm install
```

### 2. 开发模式运行

```bash
ev_plugin_cli dev
```

### 3. 构建插件
```bash
# 使用 ev_plugin_cli 构建插件包（自动构建前端和后端）
ev_plugin_cli build
```

## 📁 项目结构

```
ev_redis/
├── backend/                 # 后端项目目录
│   ├── api/                # 控制器层
│   ├── dto/                # 数据传输对象
│   ├── vo/                 # 视图对象
│   ├── my_error/           # 自定义异常处理
│   ├── response/           # 响应处理
│   ├── router/             # 路由定义
│   └── migrate/            # 数据迁移
├── frontend/               # 前端项目目录
│   ├── src/
│   │   ├── api/            # API 接口
│   │   ├── views/          # 页面视图
│   │   │   └── redis/      # Redis 相关页面
│   │   ├── router/         # 路由配置
│   │   ├── layouts/        # 布局组件
│   │   └── lang/           # 国际化文件
│   └── package.json
├── main.go                 # 主程序入口
├── plugin.json             # 插件配置
└── README.md
```

## ⚙️ 配置说明

### plugin.json 配置
```json
{
  "developer": "官方插件开发者",
  "version": "0.0.2",
  "plugin_name": "redis小助手",
  "plugin_alias": "eve-redis",
  "frontend_debug": false,
  "frontend_dev_port": 7001,
  "frontend_routes": [
    {
      "path": "redis-info",
      "name": "redis-info",
      "meta": {
        "title": "Redis信息总览",
        "icon": "el-icon-monitor"
      }
    },
    {
      "path": "memory-analysis",
      "name": "memory-analysis",
      "meta": {
        "title": "内存分析",
        "icon": "el-icon-pie-chart"
      }
    },
    {
      "path": "redis-manager",
      "name": "redis-manager",
      "meta": {
        "title": "Key管理器",
        "icon": "el-icon-list"
      }
    }
  ]
}
```

## 📝 使用说明

### Redis 信息总览
1. **服务器状态**：查看 Redis 服务器运行状态和基本信息
2. **性能监控**：监控 Redis 性能指标，包括命令执行次数、网络流量等
3. **连接管理**：查看当前连接数和连接详情
4. **配置信息**：查看 Redis 配置参数

### 内存分析
1. **内存使用统计**：查看 Redis 内存使用情况
2. **数据类型分析**：分析不同数据类型的内存占用
3. **内存优化建议**：根据内存使用情况提供优化建议
4. **趋势分析**：查看内存使用趋势图

### Key 管理器
1. **Key 浏览**：
   - 在左侧树形结构中浏览 Redis 中的所有 Key
   - 支持按数据类型分类显示
   - 支持搜索和过滤功能

2. **Key 操作**：
   - 查看 Key 值：点击 Key 查看其值和类型
   - 编辑 Key 值：使用内置编辑器修改 Key 值
   - 删除 Key：删除不需要的 Key
   - 重命名 Key：修改 Key 名称
   - 设置 TTL：为 Key 设置过期时间

3. **批量操作**：
   - 批量删除：选择多个 Key 进行批量删除
   - 批量重命名：批量修改 Key 名称
   - 批量设置 TTL：为多个 Key 设置过期时间

4. **数据类型支持**：
   - **String**：字符串类型，支持查看和编辑
   - **Hash**：哈希类型，支持字段级别的操作
   - **List**：列表类型，支持添加、删除、修改元素
   - **Set**：集合类型，支持成员管理
   - **ZSet**：有序集合类型，支持分数和成员管理

## 🔍 功能特性详解

### 实时监控
- **性能指标**：实时监控 Redis 性能数据
- **内存使用**：实时监控内存使用情况
- **连接状态**：监控连接数和连接状态
- **命令统计**：统计各种命令的执行次数

### 内存分析
- **详细统计**：提供详细的内存使用统计
- **类型分析**：分析不同数据类型的内存占用
- **优化建议**：根据内存使用情况提供优化建议
- **趋势图表**：可视化展示内存使用趋势

### Key 管理
- **智能搜索**：支持模糊搜索和正则表达式
- **批量操作**：支持批量删除、重命名等操作
- **类型识别**：自动识别 Key 的数据类型
- **TTL 管理**：查看和设置 Key 的过期时间

### 数据编辑器
- **语法高亮**：支持 JSON、XML 等格式的语法高亮
- **格式化**：自动格式化数据内容
- **验证**：数据格式验证和错误提示
- **历史记录**：保存编辑历史，支持撤销操作

## 📄 许可证

本项目采用 MIT 许可证

## 🙏 致谢

- [ElasticView](https://github.com/1340691923/ElasticView) - 提供优秀的插件开发平台
- [Redis](https://redis.io/) - 高性能的内存数据库
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 组件库

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 微信：qq1340691923

---

⭐ 如果这个项目对您有帮助，请给我们一个Star！
