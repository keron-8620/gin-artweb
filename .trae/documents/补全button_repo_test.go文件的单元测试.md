# 补全button_repo_test.go文件的单元测试计划

## 测试文件结构

根据参考的api_repo_test.go和menu_repo_test.go文件，我将按照以下结构组织button_repo_test.go文件的测试：

1. **基础测试**：测试ButtonRepo的基本CRUD操作
2. **权限策略测试**：测试AddGroupPolicy和RemoveGroupPolicy方法
3. **边界情况测试**：测试各种边界情况和错误处理
4. **上下文测试**：测试上下文超时和取消的情况
5. **关联测试**：测试与Menu和API的关联操作

## 具体测试用例

### 1. 基础CRUD测试

- **TestCreateButton**：测试创建按钮
- **TestUpdateButton**：测试更新按钮
- **TestDeleteButton**：测试删除按钮
- **TestGetButton**：测试获取单个按钮
- **TestListButton**：测试获取按钮列表（包括分页）

### 2. 权限策略测试

- **TestAddGroupPolicy**：测试添加按钮权限策略
- **TestRemoveGroupPolicy**：测试删除按钮权限策略
- **TestAddGroupPolicyWithNilButton**：测试添加权限策略时传入nil按钮
- **TestAddGroupPolicyWithZeroID**：测试添加权限策略时传入ID为0的按钮
- **TestRemoveGroupPolicyWithNilButton**：测试删除权限策略时传入nil按钮
- **TestRemoveGroupPolicyWithZeroID**：测试删除权限策略时传入ID为0的按钮
- **TestRemoveGroupPolicyWithRemoveInherited**：测试删除权限策略时设置removeInherited为true
- **TestRemoveGroupPolicyWithoutRemoveInherited**：测试删除权限策略时设置removeInherited为false

### 3. 边界情况测试

- **TestCreateButtonWithNilModel**：测试创建按钮时传入nil模型
- **TestCreateButtonWithEmptyApis**：测试创建按钮时传入空的APIs列表
- **TestUpdateButtonWithEmptyData**：测试更新按钮时传入空数据
- **TestUpdateButtonWithNilApis**：测试更新按钮时传入nil APIs
- **TestGetButtonWithEmptyConditions**：测试获取按钮时传入空条件
- **TestGetButtonWithNonExistentID**：测试获取不存在的按钮ID
- **TestDeleteButtonWithNonExistentID**：测试删除不存在的按钮ID
- **TestUpdateButtonWithNonExistentID**：测试更新不存在的按钮ID
- **TestListButtonWithPaginationBoundaries**：测试分页参数边界值
- **TestListButtonWithNoRecords**：测试列表查询无记录情况

### 4. 上下文测试

- **TestContextTimeout**：测试上下文超时情况
- **TestContextCancel**：测试上下文取消情况

### 5. 关联测试

- **TestCreateButtonWithMenu**：测试创建与菜单关联的按钮
- **TestCreateButtonWithApis**：测试创建与API关联的按钮
- **TestPreloadApis**：测试预加载关联的API

## 实现步骤

1. **完善测试套件设置**：确保测试套件正确设置数据库迁移和依赖项
2. **实现基础CRUD测试**：为每个基本方法编写测试用例
3. **实现权限策略测试**：测试权限相关的方法
4. **实现边界情况测试**：测试各种边界情况和错误处理
5. **实现上下文测试**：测试上下文相关的情况
6. **实现关联测试**：测试与其他模型的关联操作
7. **运行测试**：确保所有测试通过

## 参考文件

- `api_repo_test.go`：参考API仓库的测试结构和方法
- `menu_repo_test.go`：参考菜单仓库的测试结构和方法
- `button_repo.go`：参考按钮仓库的实现，确保测试覆盖所有方法

## 测试覆盖范围

- 所有公开方法的正常操作
- 各种边界情况和错误处理
- 权限策略的添加和删除
- 上下文管理
- 与其他模型的关联操作

通过以上测试用例，我们将确保ButtonRepo的所有功能都得到充分测试，提高代码的可靠性和可维护性。