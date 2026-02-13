# 完成user_repo_test.go单元测试文件

## 测试计划

### 1. 基础测试用例

#### CreateModel 方法测试
- **TestCreateUser**: 测试正常创建用户
- **TestCreateUserWithNilModel**: 测试传入nil模型
- **TestCreateUserWithRole**: 测试创建与角色关联的用户

#### UpdateModel 方法测试
- **TestUpdateUser**: 测试正常更新用户
- **TestUpdateUserWithEmptyData**: 测试传入空数据
- **TestUpdateUserWithNonExistentID**: 测试更新不存在的用户

#### DeleteModel 方法测试
- **TestDeleteUser**: 测试正常删除用户
- **TestDeleteUserWithNonExistentID**: 测试删除不存在的用户

#### GetModel 方法测试
- **TestGetUser**: 测试根据ID查询用户
- **TestGetUserWithNonExistentID**: 测试查询不存在的用户
- **TestGetUserWithEmptyConditions**: 测试空条件查询
- **TestGetUserWithPreloadRole**: 测试预加载角色信息

#### ListModel 方法测试
- **TestListUser**: 测试查询用户列表
- **TestListUserWithPagination**: 测试分页查询
- **TestListUserWithPaginationBoundaries**: 测试边界情况（Limit=0, 大Offset）
- **TestListUserWithNoRecords**: 测试无记录查询

### 2. 上下文相关测试

- **TestContextTimeout**: 测试上下文超时
- **TestContextCancel**: 测试上下文取消

### 3. 与角色关联测试

- **TestCreateUserWithRole**: 测试创建与角色关联的用户
- **TestGetUserWithPreloadRole**: 测试查询与角色关联的用户

## 实现细节

1. **测试数据准备**: 复用现有的CreateTestUserModel函数，创建测试用的用户模型
2. **测试环境设置**: 在SetupSuite中初始化数据库和仓库实例
3. **测试断言**: 使用testify/suite的断言方法验证测试结果
4. **错误处理**: 验证错误返回和错误信息
5. **边界情况**: 测试各种边界情况和异常输入

## 参考模式

参考button_repo_test.go和role_repo_test.go的测试结构和模式，确保测试用例的一致性和完整性。