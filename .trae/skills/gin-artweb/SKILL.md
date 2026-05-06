---
name: gin-artweb
description: 用于生成/修改/审查 Gin 后端代码，包含接口、路由、中间件、错误处理、日志、配置，在用户提及 gin、api、接口、路由、中间件、zap、viper、gorm 时触发。
---

---
name: gin-artweb
description: 用于生成/修改/审查 Gin 后端代码，包含接口、路由、中间件、错误处理、日志、配置，在用户提及 gin、api、接口、路由、中间件、zap、viper、gorm 时触发。
---

# 项目真实目录结构（必须严格遵守）
config/
pkg/
cmd/
internal/
  ├── handler/
  │     ├── job, mds, mon, oes, resource, sys
  ├── model/
  │     ├── common, job, mds, mon, oes, resource, sys
  ├── repo/
  │     ├── job, mds, mon, oes, resource, sys
  ├── routers/
  ├── service/
  │     ├── job, mds, mon, oes, resource, sys
  └── shared/
        ├── auth, common, config, crontab, ctxutil
        ├── database, errors, log, middleware, shell, test

# 代码分层规则
1. handler：接收请求、参数校验、返回响应（放在 internal/handler/xxx）
2. service：业务逻辑（放在 internal/service/xxx）
3. repo：数据库操作（放在 internal/repo/xxx）
4. model：数据结构（放在 internal/model/xxx）
5. shared：公共组件、中间件、工具、错误、日志、数据库连接

# 输出强化要求
1. 严格按真实目录生成代码
2. 必须写全分层：handler → service → repo → model
3. 禁止乱写路径、乱写包名