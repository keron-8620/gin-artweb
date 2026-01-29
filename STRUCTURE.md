# 项目结构组织说明

## 概述
本项目采用分层架构，将基础设施模块和业务模块分离，以提高代码的可维护性和可扩展性。

## 目录结构

```
├── api                         # API 定义
│   ├── customer                # 用户、角色、权限等基础管理 API
│   ├── jobs                    # 任务调度、脚本执行等基础服务 API
│   ├── resource                # 主机、软件包等资源管理 API
│   ├── mds                     # 市场数据系统 API
│   ├── mon                     # 监控系统 API
│   └── oes                     # 运营支撑系统 API
├── internal                    # 内部实现
│   ├── infra                   # 基础设施模块
│   │   ├── customer            # 用户、角色、权限等基础管理
│   │   ├── jobs                # 任务调度、脚本执行等基础服务
│   │   └── resource            # 主机、软件包等资源管理
│   ├── business                # 业务模块
│   │   ├── mds                 # 市场数据系统
│   │   ├── mon                 # 监控系统
│   │   └── oes                 # 运营支撑系统
│   └── shared                  # 共享模块
│       ├── auth
│       ├── common
│       ├── config
│       ├── crontab
│       ├── database
│       ├── errors
│       ├── log
│       └── middleware
├── pkg                         # 公共库，可被外部项目引用
│   ├── archive                 # 归档功能（tgz、zip等）
│   ├── crypto                  # 加密算法相关
│   ├── ctxutil                 # 上下文工具
│   ├── fileutil                # 文件操作工具
│   ├── serializer              # 序列化工具（JSON、YAML等）
│   └── shell                   # Shell操作工具（SSH、SFTP等）
```

## 模块分类

### 公共库 (pkg)
- **archive**: 提供归档功能（tgz、zip等）
- **crypto**: 提供加密算法相关功能（AES、SHA、BCrypt等）
- **ctxutil**: 提供上下文工具函数
- **fileutil**: 提供文件操作工具函数
- **serializer**: 提供序列化工具（JSON、YAML等）
- **shell**: 提供Shell操作功能（SSH、SFTP等）

### 基础设施模块 (infra)
- **customer**: 提供用户认证、授权、角色管理等基础服务
- **jobs**: 提供任务调度、脚本执行等基础服务
- **resource**: 提供主机、软件包等基础资源管理

### 业务模块 (business)
- **mds**: 市场数据系统，处理市场数据相关业务
- **mon**: 监控系统，提供各种监控功能
- **oes**: 运营支撑系统，支持运营相关功能

## 设计原则
1. pkg目录中的代码具有高度通用性，可被外部项目引用
2. 基础设施模块可被业务模块依赖，但业务模块之间不应该相互依赖
3. 基础设施模块提供通用能力，业务模块实现具体业务逻辑
4. 共享模块提供项目内部的通用组件
5. 模块间依赖应单向流动：业务模块可依赖基础设施模块和pkg库，基础设施模块可依赖pkg库和shared模块