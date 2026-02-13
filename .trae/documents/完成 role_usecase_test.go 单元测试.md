# 完成 role_usecase_test.go 单元测试

## 测试方法规划

### 1. 基础方法测试
- **TestGetApis** - 测试获取 API 列表功能
- **TestGetMenus** - 测试获取菜单列表功能
- **TestGetButtons** - 测试获取按钮列表功能
- **TestFindRoleByID** - 测试根据 ID 查询角色
- **TestListRole** - 测试查询角色列表
- **TestLoadRolePolicy** - 测试加载角色策略
- **TestGetRoleMenuTree** - 测试获取角色菜单树

### 2. 核心功能测试
- **TestCreateRole** - 测试创建角色（无关联）
- **TestCreateRoleWithRelations** - 测试创建角色（关联 API、菜单、按钮）
- **TestUpdateRoleByID** - 测试更新角色
- **TestDeleteRoleByID** - 测试删除角色

### 3. 错误处理测试
- **TestGetApisWithContextError** - 测试上下文错误处理
- **TestGetMenusWithContextError** - 测试上下文错误处理
- **TestGetButtonsWithContextError** - 测试上下文错误处理
- **TestCreateRoleWithContextError** - 测试上下文错误处理
- **TestUpdateRoleByIDWithContextError** - 测试上下文错误处理
- **TestDeleteRoleByIDWithContextError** - 测试上下文错误处理
- **TestFindRoleByIDWithContextError** - 测试上下文错误处理
- **TestListRoleWithContextError** - 测试上下文错误处理
- **TestLoadRolePolicyWithContextError** - 测试上下文错误处理
- **TestGetRoleMenuTreeWithContextError** - 测试上下文错误处理

### 4. 多对多关系测试
- **TestRoleApiRelations** - 测试角色与 API 的多对多关系
- **TestRoleMenuRelations** - 测试角色与菜单的多对多关系
- **TestRoleButtonRelations** - 测试角色与按钮的多对多关系

### 5. Casbin 权限继承测试
- **TestRoleCasbinInheritance** - 测试角色对 API、菜单、按钮的权限继承

## 测试实现要点

1. **参考 button_usecase_test.go** 的测试结构和模式
2. **使用 CreateTestRoleModel** 函数创建测试数据
3. **测试前准备** - 为关联测试创建 API、菜单、按钮数据
4. **断言验证** - 验证返回结果、错误状态、关联关系等
5. **Casbin 测试** - 验证角色策略的添加和继承

## 实现步骤

1. 编写基础方法测试用例
2. 编写核心功能测试用例
3. 编写错误处理测试用例
4. 编写多对多关系测试用例
5. 编写 Casbin 权限继承测试用例
6. 运行测试验证所有测试通过