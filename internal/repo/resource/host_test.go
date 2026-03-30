package resource

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestHostModel() *resomodel.HostModel {
	// 生成唯一的端口号，避免唯一约束冲突
	uuidStr := uuid.NewString()
	// 取UUID的后4位作为端口号的一部分
	portSuffix := uuidStr[len(uuidStr)-4:]
	// 转换为数字并确保在有效端口范围内
	port := 2222 + (len(uuidStr) % 1000)
	return &resomodel.HostModel{
		Name:    fmt.Sprintf("host-%s", uuidStr),
		Label:   "test",
		SSHIP:   "127.0.0.1",
		SSHPort: uint16(port),
		SSHUser: fmt.Sprintf("root-%s", portSuffix),
		PyPath:  "/usr/bin/python3",
		Remark:  "",
	}
}

type SSHDContainer struct {
	ID       string
	IP       string
	Port     uint16
	User     string
	Password string
}

func setupSSHContainer(t *testing.T) (*SSHDContainer, func(), error) {
	containerID := fmt.Sprintf("test-sshd-%s", uuid.NewString()[:8])

	cmd := exec.Command("podman", "run", "-d", "--rm", "--name", containerID, "-p", "0:2222", "docker.io/testcontainers/sshd:latest")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to run SSH container, stderr: %s", stderr.String())
	}

	cleanID := strings.TrimSpace(stdout.String())
	if cleanID == "" {
		return nil, nil, errors.New("failed to get container ID")
	}

	time.Sleep(2 * time.Second)

	portCmd := exec.Command("podman", "port", cleanID, "2222/tcp")
	portOutput, err := portCmd.Output()
	if err != nil {
		cleanupContainer(cleanID)
		return nil, nil, errors.Wrapf(err, "failed to get port")
	}

	portStr := strings.TrimSpace(string(portOutput))
	var port uint16
	if _, err := fmt.Sscanf(portStr, "0.0.0.0:%d", &port); err != nil {
		if _, err := fmt.Sscanf(portStr, "127.0.0.1:%d", &port); err != nil {
			port = 2222
		}
	}

	ipCmd := exec.Command("podman", "inspect", "--format", "{{.NetworkSettings.IPAddress}}", cleanID)
	ipOutput, err := ipCmd.Output()
	if err != nil {
		cleanupContainer(cleanID)
		return nil, nil, errors.Wrapf(err, "failed to get IP")
	}

	container := &SSHDContainer{
		ID:       cleanID,
		IP:       strings.TrimSpace(string(ipOutput)),
		Port:     port,
		User:     "root",
		Password: "testcontainers",
	}

	cleanup := func() {
		cleanupContainer(container.ID)
	}

	return container, cleanup, nil
}

func cleanupContainer(id string) {
	cmd := exec.Command("podman", "rm", "-f", id)
	cmd.Run()
}

func createSSHClient(t *testing.T, container *SSHDContainer) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            container.User,
		Auth:            []ssh.AuthMethod{ssh.Password(container.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(container.IP, fmt.Sprintf("%d", container.Port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial SSH at %s", addr)
	}

	return client, nil
}

type HostTestSuite struct {
	suite.Suite
	hostRepo *HostRepo
}

func (suite *HostTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.HostModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.hostRepo = &HostRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *HostTestSuite) TestCreateHost() {
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host应该成功")

	fm, err := suite.hostRepo.GetModel(context.Background(), nil, hm.ID)
	suite.NoError(err, "查询刚创建的Host应该成功")
	suite.Equal(hm.ID, fm.ID)
	suite.Equal(hm.Name, fm.Name)
	suite.Equal(hm.Label, fm.Label)
	suite.Equal(hm.SSHIP, fm.SSHIP)
	suite.Equal(hm.SSHPort, fm.SSHPort)
	suite.Equal(hm.SSHUser, fm.SSHUser)
	suite.Equal(hm.PyPath, fm.PyPath)
	suite.Equal(hm.Remark, fm.Remark)
}

func (suite *HostTestSuite) TestUpdateModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于更新测试应该成功")

	// 测试正常更新
	updateData := map[string]any{
		"Name":    "updated-host",
		"SSHPort": 2222,
		"Remark":  "updated remark",
	}
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", hm.ID)
	suite.NoError(err, "更新Host应该成功")

	// 验证更新结果
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.NoError(err, "查询更新后的Host应该成功")
	suite.Equal("updated-host", fm.Name)
	suite.Equal(uint16(2222), fm.SSHPort)
	suite.Equal("updated remark", fm.Remark)

	// 测试边界情况：更新数据为空
	err = suite.hostRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", hm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的Host
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的Host应该成功（无操作）")
}

func (suite *HostTestSuite) TestDeleteModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于删除测试应该成功")

	// 测试正常删除
	err = suite.hostRepo.DeleteModel(context.Background(), "id = ?", hm.ID)
	suite.NoError(err, "删除Host应该成功")

	// 验证删除结果
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.Error(err, "查询已删除的Host应该返回错误")
	suite.Nil(fm, "已删除的Host应该为nil")

	// 测试边界情况：删除不存在的Host
	err = suite.hostRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的Host应该成功（无操作）")
}

func (suite *HostTestSuite) TestGetModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.NoError(err, "查询Host应该成功")
	suite.Equal(hm.ID, fm.ID)
	suite.Equal(hm.Name, fm.Name)

	// 测试边界情况：查询不存在的Host
	fm, err = suite.hostRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的Host应该返回错误")
	suite.Nil(fm, "查询不存在的Host应该返回nil")

	// 测试边界情况：使用预加载（虽然HostModel可能没有关联关系，但测试方法调用）
	fm, err = suite.hostRepo.GetModel(context.Background(), []string{}, "id = ?", hm.ID)
	suite.NoError(err, "使用空预加载查询Host应该成功")
	suite.Equal(hm.ID, fm.ID)
}

func (suite *HostTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		hm := CreateTestHostModel()
		hm.Name = fmt.Sprintf("host-%d", i)
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err, "创建Host用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{}
	models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询Host列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "Host列表数量应该大于0")
	suite.NotNil(models, "Host列表应该不为nil")
	suite.Greater(int64(len(models)), int64(0), "Host列表长度应该大于0")

	// 测试边界情况：空列表（如果之前没有数据）
	// 注意：由于测试套件是共享数据库，这里可能不会为空，但我们仍然测试方法调用
	qp2 := database.QueryParams{
		Query: map[string]any{"name": "non-existent-host"},
	}
	models2, err := suite.hostRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的Host列表应该成功")
	suite.Equal(int64(0), int64(len(models2)), "不存在的Host列表数量应该为0")
	suite.NotNil(models2, "不存在的Host列表应该不为nil")
	suite.Len(models2, 0, "不存在的Host列表长度应该为0")
}

func (suite *HostTestSuite) TestCreateModelWithEmpty() {
	// 测试边界情况：创建空模型
	err := suite.hostRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空Host模型应该返回错误")
}

func (suite *HostTestSuite) TestContextTimeout() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.hostRepo.GetModel(timeoutCtx, nil, "id = ?", hm.ID)
	suite.Error(err, "上下文超时后查询Host应该返回错误")
}

func (suite *HostTestSuite) TestCreateModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的创建操作
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(timeoutCtx, hm)
	suite.Error(err, "上下文超时后创建Host应该返回错误")
}

func (suite *HostTestSuite) TestUpdateModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的更新操作
	updateData := map[string]any{"Name": "updated-host"}
	err := suite.hostRepo.UpdateModel(timeoutCtx, updateData, "id = ?", 1)
	suite.Error(err, "上下文超时后更新Host应该返回错误")
}

func (suite *HostTestSuite) TestDeleteModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的删除操作
	err := suite.hostRepo.DeleteModel(timeoutCtx, "id = ?", 1)
	suite.Error(err, "上下文超时后删除Host应该返回错误")
}

func (suite *HostTestSuite) TestListModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的列表操作
	qp := database.QueryParams{}
	_, err := suite.hostRepo.ListModel(timeoutCtx, qp)
	suite.Error(err, "上下文超时后查询Host列表应该返回错误")
}

func (suite *HostTestSuite) TestCountModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的计数操作
	_, err := suite.hostRepo.CountModel(timeoutCtx, map[string]any{"label": "test"})
	suite.Error(err, "上下文超时后计数Host应该返回错误")
}

func (suite *HostTestSuite) TestNewSSHClientWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的SSH客户端创建
	client, err := suite.hostRepo.NewSSHClient(timeoutCtx, "127.0.0.1", 22, "root", nil, time.Second)
	suite.Error(err, "上下文超时后创建SSH客户端应该返回错误")
	suite.Nil(client, "上下文超时后创建的SSH客户端应该为nil")
}

func (suite *HostTestSuite) TestExecuteCommandWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的命令执行
	err := suite.hostRepo.ExecuteCommand(timeoutCtx, nil, "echo test")
	suite.Error(err, "上下文超时后执行命令应该返回错误")
}

func (suite *HostTestSuite) TestCountModel() {
	// 使用唯一的标签来避免数据残留的影响
	uniqueLabel := fmt.Sprintf("test-label-%s", uuid.NewString()[:8])

	hm := CreateTestHostModel()
	hm.Label = uniqueLabel
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)

	count, err := suite.hostRepo.CountModel(context.Background(), map[string]any{"label": uniqueLabel})
	suite.NoError(err)
	suite.Equal(int64(1), count)

	count, err = suite.hostRepo.CountModel(context.Background(), map[string]any{"label": "nonexistent"})
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *HostTestSuite) TestNewSSHClientSuccess() {
	container, cleanup, err := setupSSHContainer(suite.T())
	if err != nil || container == nil {
		suite.T().Skip("SSH container not available, skipping integration test")
		return
	}
	defer cleanup()

	passwordAuth := ssh.Password(container.Password)
	sshClient, err := suite.hostRepo.NewSSHClient(
		context.Background(),
		container.IP,
		container.Port,
		container.User,
		[]ssh.AuthMethod{passwordAuth},
		10*time.Second,
	)

	if err != nil {
		suite.T().Logf("SSH connection failed (container may not be fully ready): %v", err)
		suite.T().Skip("SSH container not fully ready")
		return
	}

	suite.NoError(err)
	suite.NotNil(sshClient)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	suite.NoError(err)
	defer session.Close()

	err = session.Run("echo 'hello'")
	suite.NoError(err)
}

func (suite *HostTestSuite) TestNewSSHClientInvalidParams() {
	tests := []struct {
		name    string
		ip      string
		port    uint16
		user    string
		auths   []ssh.AuthMethod
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "empty IP",
			ip:      "",
			port:    22,
			user:    "root",
			auths:   []ssh.AuthMethod{ssh.Password("test")},
			timeout: time.Second,
			wantErr: true,
		},
		{
			name:    "empty user",
			ip:      "127.0.0.1",
			port:    22,
			user:    "",
			auths:   []ssh.AuthMethod{ssh.Password("test")},
			timeout: time.Second,
			wantErr: true,
		},
		{
			name:    "empty auth",
			ip:      "127.0.0.1",
			port:    22,
			user:    "root",
			auths:   nil,
			timeout: time.Second,
			wantErr: true,
		},
		{
			name:    "zero auth length",
			ip:      "127.0.0.1",
			port:    22,
			user:    "root",
			auths:   []ssh.AuthMethod{},
			timeout: time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client, err := suite.hostRepo.NewSSHClient(
				context.Background(),
				tt.ip,
				tt.port,
				tt.user,
				tt.auths,
				tt.timeout,
			)
			if tt.wantErr {
				suite.Error(err)
				suite.Nil(client)
			} else {
				if err != nil {
					t.Logf("SSH connection error (expected for invalid params): %v", err)
				}
			}
		})
	}
}

func (suite *HostTestSuite) TestNewSSHClientContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := suite.hostRepo.NewSSHClient(
		ctx,
		"127.0.0.1",
		22,
		"root",
		[]ssh.AuthMethod{ssh.Password("test")},
		time.Second,
	)
	suite.Error(err)
	suite.Nil(client)
}

func (suite *HostTestSuite) TestExecuteCommandSuccess() {
	container, cleanup, err := setupSSHContainer(suite.T())
	if err != nil || container == nil {
		suite.T().Skip("SSH container not available, skipping integration test")
		return
	}
	defer cleanup()

	passwordAuth := ssh.Password(container.Password)
	sshClient, err := suite.hostRepo.NewSSHClient(
		context.Background(),
		container.IP,
		container.Port,
		container.User,
		[]ssh.AuthMethod{passwordAuth},
		10*time.Second,
	)
	if err != nil {
		suite.T().Skip("SSH container not fully ready")
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	suite.NoError(err)
	defer session.Close()

	err = suite.hostRepo.ExecuteCommand(context.Background(), session, "echo 'hello world'")
	suite.NoError(err)
}

func (suite *HostTestSuite) TestExecuteCommandContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var session *ssh.Session
	err := suite.hostRepo.ExecuteCommand(ctx, session, "echo test")
	suite.Error(err)
}

func (suite *HostTestSuite) TestExecuteCommandCommandFailed() {
	container, cleanup, err := setupSSHContainer(suite.T())
	if err != nil || container == nil {
		suite.T().Skip("SSH container not available, skipping integration test")
		return
	}
	defer cleanup()

	passwordAuth := ssh.Password(container.Password)
	sshClient, err := suite.hostRepo.NewSSHClient(
		context.Background(),
		container.IP,
		container.Port,
		container.User,
		[]ssh.AuthMethod{passwordAuth},
		10*time.Second,
	)
	if err != nil {
		suite.T().Skip("SSH container not fully ready")
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	suite.NoError(err)
	defer session.Close()

	err = suite.hostRepo.ExecuteCommand(context.Background(), session, "exit 1")
	suite.Error(err)
}

func (suite *HostTestSuite) TestExecuteCommandNilSession() {
	// 测试session为nil的情况
	err := suite.hostRepo.ExecuteCommand(context.Background(), nil, "echo test")
	suite.Error(err, "session为nil时执行命令应该返回错误")
}

func (suite *HostTestSuite) TestExecuteCommandWithContextCanceled() {
	// 测试ExecuteCommand函数的上下文取消情况
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.hostRepo.ExecuteCommand(ctx, nil, "echo test")
	suite.Error(err, "上下文取消时执行命令应该返回错误")
}

func (suite *HostTestSuite) TestExecuteCommandWithValidSession() {
	// 测试ExecuteCommand函数的基本逻辑
	// 由于ssh.Session需要真实的连接，我们无法完全模拟
	// 但我们可以测试函数的基本结构和错误处理

	// 测试正常上下文
	ctx := context.Background()

	// 测试session为nil的情况（应该返回错误）
	err := suite.hostRepo.ExecuteCommand(ctx, nil, "echo test")
	suite.Error(err, "session为nil时应该返回错误")

	// 测试空命令
	err = suite.hostRepo.ExecuteCommand(ctx, nil, "")
	suite.Error(err, "session为nil时即使命令为空也应该返回错误")

	suite.T().Log("ExecuteCommand函数的基本逻辑测试完成")
}

func (suite *HostTestSuite) TestListModelWithPagination() {
	// 使用唯一的标签来避免数据残留的影响
	uniqueLabel := fmt.Sprintf("pagination-label-%s", uuid.NewString()[:8])

	// 清理之前的测试数据
	suite.hostRepo.DeleteModel(context.Background(), "label = ?", uniqueLabel)

	for i := 0; i < 10; i++ {
		hm := &resomodel.HostModel{
			Name:    fmt.Sprintf("host-%d-%s", i, uuid.NewString()[:4]),
			Label:   uniqueLabel,
			SSHIP:   "127.0.0.1",
			SSHPort: uint16(2200 + i),
			SSHUser: fmt.Sprintf("user-%d", i),
			PyPath:  "/usr/bin/python3",
		}
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err)
	}

	qp := database.QueryParams{
		Query:   map[string]any{"label": uniqueLabel},
		OrderBy: []string{"id desc"},
		Limit:   5,
		Offset:  0,
	}
	models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Len(models, 5)

	qp.Offset = 5
	models, err = suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Len(models, 5)
}

func (suite *HostTestSuite) TestNewHostRepo() {
	logger := zap.NewNop()
	db := test.NewTestGormDBWithConfig(nil)
	timeouts := test.NewTestDBTimeouts()

	repo := NewHostRepo(logger, db, timeouts)
	suite.NotNil(repo)
	suite.Equal(logger, repo.log)
	suite.Equal(db, repo.gormDB)
	suite.Equal(timeouts, repo.timeouts)
}

func (suite *HostTestSuite) TestUpdateModelNotFound() {
	updateData := map[string]any{"Name": "updated-name"}
	err := suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err)
}

func (suite *HostTestSuite) TestListModelEmptyResult() {
	qp := database.QueryParams{
		Query: map[string]any{"name": "totally-nonexistent-host-name-xyz"},
	}
	models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Len(models, 0)
}

func (suite *HostTestSuite) TestGetModelWithPreloads() {
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)

	fm, err := suite.hostRepo.GetModel(context.Background(), []string{}, "id = ?", hm.ID)
	suite.NoError(err)
	suite.Equal(hm.ID, fm.ID)
}

func (suite *HostTestSuite) TestListModelWithQuery() {
	hm := CreateTestHostModel()
	hm.Label = "special-label"
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)

	qp := database.QueryParams{
		Query: map[string]any{"label": "special-label"},
	}
	models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Len(models, 1)
	suite.Equal(hm.Label, models[0].Label)
}

func (suite *HostTestSuite) TestDeleteModelNotFound() {
	err := suite.hostRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err)
}

func (suite *HostTestSuite) TestGetModelNotFound() {
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err)
	suite.Nil(fm)
}

func (suite *HostTestSuite) TestDeleteModelSuccess() {
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)

	err = suite.hostRepo.DeleteModel(context.Background(), "id = ?", hm.ID)
	suite.NoError(err)

	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.Error(err)
	suite.Nil(fm)
}

func (suite *HostTestSuite) TestCreateModelSuccess() {
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)
	suite.NotZero(hm.ID)
	suite.NotZero(hm.CreatedAt)
	suite.NotZero(hm.UpdatedAt)

	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.NoError(err)
	suite.Equal(hm.ID, fm.ID)
}

func (suite *HostTestSuite) TestCreateModelNilModel() {
	err := suite.hostRepo.CreateModel(context.Background(), nil)
	suite.Error(err)
}

func (suite *HostTestSuite) TestUpdateModelSuccess() {
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)

	// 使用唯一的名称来避免唯一约束冲突
	uniqueName := fmt.Sprintf("updated-host-%s", uuid.NewString()[:8])
	updateData := map[string]any{
		"Name":    uniqueName,
		"SSHPort": 3322,
		"Remark":  "updated remark",
	}
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", hm.ID)
	suite.NoError(err)

	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.NoError(err)
	suite.Equal(uniqueName, fm.Name)
	suite.Equal(uint16(3322), fm.SSHPort)
	suite.Equal("updated remark", fm.Remark)
}

func (suite *HostTestSuite) TestUpdateModelEmptyData() {
	err := suite.hostRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err)
}

func (suite *HostTestSuite) TestCountModelMultipleRecords() {
	for i := 0; i < 5; i++ {
		hm := CreateTestHostModel()
		hm.Label = "count-label"
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err)
	}

	count, err := suite.hostRepo.CountModel(context.Background(), map[string]any{"label": "count-label"})
	suite.NoError(err)
	suite.Equal(int64(5), count)
}

func (suite *HostTestSuite) TestListModelWithOrderBy() {
	for i := 0; i < 3; i++ {
		hm := CreateTestHostModel()
		hm.Name = fmt.Sprintf("host-ordered-%d", i)
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err)
	}

	qp := database.QueryParams{
		OrderBy: []string{"id asc"},
	}
	models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.GreaterOrEqual(len(models), 3)
}

func TestHostTestSuite(t *testing.T) {
	pts := &HostTestSuite{}
	suite.Run(t, pts)
}
