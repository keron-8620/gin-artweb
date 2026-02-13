# 补全ApiUsecase单元测试计划

## 测试覆盖分析

### 已实现的测试方法
- `TestCreateApi`：测试创建API
- `TestFindApiByID`：测试通过ID查找API
- `TestFindApiByID_NotFound`：测试查找不存在的API
- `TestDeleteApi`：测试删除API
- `TestDeleteApi_NotFound`：测试删除不存在的API

### 未实现的测试方法
1. **TestUpdateApiByID**：测试更新API
2. **TestUpdateApiByID_NotFound**：测试更新不存在的API
3. **TestListApi**：测试列出API
4. **TestLoadApiPolicy**：测试加载API策略

### 边界情况测试
1. **TestCreateApi_ContextError**：测试上下文错误情况下的创建API
2. **TestUpdateApiByID_ContextError**：测试上下文错误情况下的更新API
3. **TestDeleteApiByID_ContextError**：测试上下文错误情况下的删除API
4. **TestFindApiByID_ContextError**：测试上下文错误情况下的查找API
5. **TestListApi_ContextError**：测试上下文错误情况下的列出API
6. **TestLoadApiPolicy_ContextError**：测试上下文错误情况下的加载API策略

## 实现步骤

1. **添加TestUpdateApiByID测试**：
   - 创建API
   - 更新API字段
   - 验证更新结果
   - 验证权限策略更新

2. **添加TestUpdateApiByID_NotFound测试**：
   - 尝试更新不存在的API
   - 验证错误返回

3. **添加TestListApi测试**：
   - 创建多个API
   - 测试列表查询
   - 验证返回结果数量和内容

4. **添加TestLoadApiPolicy测试**：
   - 创建多个API
   - 加载API策略
   - 验证权限策略加载成功

5. **添加上下文错误边界测试**：
   - 为每个方法添加上下文取消/超时测试
   - 验证错误处理

6. **运行测试验证**：
   - 执行完整测试套件
   - 确保所有测试通过

## 测试设计原则

- **覆盖所有方法**：确保每个public方法都有对应的测试
- **边界情况**：测试上下文错误、不存在的资源等边界情况
- **权限验证**：确保权限策略正确创建和更新
- **数据一致性**：验证数据操作前后的一致性
- **错误处理**：验证错误情况下的正确处理