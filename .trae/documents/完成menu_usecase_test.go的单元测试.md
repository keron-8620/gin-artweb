# 完成menu_usecase_test.go的单元测试计划

## 测试方法覆盖

我将为`MenuUsecase`的所有方法编写测试用例，参考`api_usecase_test.go`的测试风格：

### 1. 基础方法测试
- **TestGetParentMenu**: 测试获取父菜单的功能
- **TestGetApis**: 测试获取API列表的功能

### 2. 核心功能测试
- **TestCreateMenu**: 测试创建菜单的功能
  - 正常创建菜单
  - 验证菜单ID生成
  - 验证权限策略添加
- **TestFindMenuByID**: 测试根据ID查找菜单的功能
  - 正常查找已创建的菜单
  - 查找不存在的菜单（错误情况）
- **TestUpdateMenuByID**: 测试更新菜单的功能
  - 正常更新菜单
  - 更新不存在的菜单（错误情况）
- **TestDeleteMenuByID**: 测试删除菜单的功能
  - 正常删除已创建的菜单
  - 删除不存在的菜单（错误情况）
- **TestListMenu**: 测试列出菜单的功能
  - 创建多个菜单后验证列表返回
- **TestLoadMenuPolicy**: 测试加载菜单策略的功能
  - 创建菜单后验证策略加载

### 3. 错误处理测试
为所有方法添加上下文错误测试：
- **TestCreateMenu_ContextError**
- **TestFindMenuByID_ContextError**
- **TestUpdateMenuByID_ContextError**
- **TestDeleteMenuByID_ContextError**
- **TestListMenu_ContextError**
- **TestLoadMenuPolicy_ContextError**
- **TestGetParentMenu_ContextError**
- **TestGetApis_ContextError**

## 测试结构和风格

- 使用`suite`测试框架
- 每个测试用例独立运行
- 验证返回值的正确性
- 验证数据库操作的结果
- 验证权限策略的添加和移除
- 测试错误处理情况
- 保持与`api_usecase_test.go`一致的代码风格

## 实现步骤

1. 为每个方法编写对应的测试用例
2. 确保测试覆盖所有正常和错误情况
3. 验证权限策略的正确性
4. 测试上下文错误处理
5. 运行测试确保所有测试通过