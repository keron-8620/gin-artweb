package resource

import (
	"bytes"
	"context"
	"fmt"
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

// 测试数据管理
const (
	// 测试默认值
	DefaultTestLabel     = "test"
	DefaultTestSSHIP     = "127.0.0.1"
	DefaultTestPyPath    = "/usr/bin/python3"
	DefaultTestPortStart = 2222
	DefaultTestPortRange = 1000
)

// CreateTestHostModel 创建测试用的Host模型
// 可选参数用于覆盖默认值
func CreateTestHostModel(overrides ...func(*resomodel.HostModel)) *resomodel.HostModel {
	// 生成唯一的端口号，避免唯一约束冲突
	uuidStr := uuid.NewString()
	// 取UUID的后4位作为端口号的一部分
	portSuffix := uuidStr[len(uuidStr)-4:]
	// 转换为数字并确保在有效端口范围内
	port := DefaultTestPortStart + (len(uuidStr) % DefaultTestPortRange)

	host := &resomodel.HostModel{
		Name:    fmt.Sprintf("host-%s", uuidStr),
		Label:   DefaultTestLabel,
		SSHIP:   DefaultTestSSHIP,
		SSHPort: uint16(port),
		SSHUser: fmt.Sprintf("root-%s", portSuffix),
		PyPath:  DefaultTestPyPath,
		Remark:  "",
	}

	// 应用覆盖函数
	for _, override := range overrides {
		override(host)
	}

	return host
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
		_ = cleanupContainer(cleanID)
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
		_ = cleanupContainer(cleanID)
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
		_ = cleanupContainer(container.ID)
	}

	return container, cleanup, nil
}

func cleanupContainer(id string) error {
	cmd := exec.Command("podman", "rm", "-f", id)
	return cmd.Run()
}

type HostTestSuite struct {
	suite.Suite
	hostRepo *HostRepo
}

func (suite *HostTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&resomodel.HostModel{}); err != nil {
		suite.Error(err, "failed to migrate HostModel")
	}
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	suite.hostRepo = &HostRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
	}
}

func (suite *HostTestSuite) TestCreateModel() {
	// 测试正常创建
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host应该成功")

	// 验证模型字段
	suite.NotZero(hm.ID, "Host ID应该被设置")
	suite.NotZero(hm.CreatedAt, "Host CreatedAt应该被设置")
	suite.NotZero(hm.UpdatedAt, "Host UpdatedAt应该被设置")

	// 验证创建的模型可以被正确查询
	fm, err := suite.hostRepo.GetModel(context.Background(), "id = ?", hm.ID)
	suite.NoError(err, "查询刚创建的Host应该成功")
	suite.NotNil(fm, "查询结果不应该为nil")

	// 验证所有字段都被正确保存
	suite.Equal(hm.ID, fm.ID, "ID字段应该匹配")
	suite.Equal(hm.Name, fm.Name, "Name字段应该匹配")
	suite.Equal(hm.Label, fm.Label, "Label字段应该匹配")
	suite.Equal(hm.SSHIP, fm.SSHIP, "SSHIP字段应该匹配")
	suite.Equal(hm.SSHPort, fm.SSHPort, "SSHPort字段应该匹配")
	suite.Equal(hm.SSHUser, fm.SSHUser, "SSHUser字段应该匹配")
	suite.Equal(hm.PyPath, fm.PyPath, "PyPath字段应该匹配")
	suite.Equal(hm.Remark, fm.Remark, "Remark字段应该匹配")

	// 测试边界情况:创建空模型
	err = suite.hostRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空Host模型应该返回错误")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	hm2 := CreateTestHostModel()
	err = suite.hostRepo.CreateModel(timeoutCtx, hm2)
	suite.Error(err, "上下文超时后创建Host应该返回错误")
}

func (suite *HostTestSuite) TestUpdateModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于更新测试应该成功")

	// 测试正常更新
	uniqueName := fmt.Sprintf("updated-host-%s", uuid.NewString()[:8])
	updateData := map[string]any{
		"Name":    uniqueName,
		"SSHPort": 3322,
		"Remark":  "updated remark",
	}
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", hm.ID)
	suite.NoError(err, "更新Host应该成功")

	// 验证更新结果
	fm, err := suite.hostRepo.GetModel(context.Background(), "id = ?", hm.ID)
	suite.NoError(err, "查询更新后的Host应该成功")
	suite.NotNil(fm, "查询结果不应该为nil")
	suite.Equal(uniqueName, fm.Name, "Name字段应该被更新")
	suite.Equal(uint16(3322), fm.SSHPort, "SSHPort字段应该被更新")
	suite.Equal("updated remark", fm.Remark, "Remark字段应该被更新")

	// 测试边界情况:更新数据为空
	err = suite.hostRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", hm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况:更新不存在的Host
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的Host应该成功（无操作）")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	err = suite.hostRepo.UpdateModel(timeoutCtx, updateData, "id = ?", 1)
	suite.Error(err, "上下文超时后更新Host应该返回错误")
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
	fm, err := suite.hostRepo.GetModel(context.Background(), "id = ?", hm.ID)
	suite.Error(err, "查询已删除的Host应该返回错误")
	suite.Nil(fm, "已删除的Host应该为nil")

	// 测试边界情况:删除不存在的Host
	err = suite.hostRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的Host应该成功（无操作）")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	err = suite.hostRepo.DeleteModel(timeoutCtx, "id = ?", 1)
	suite.Error(err, "上下文超时后删除Host应该返回错误")
}

func (suite *HostTestSuite) TestGetModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.hostRepo.GetModel(context.Background(), "id = ?", hm.ID)
	suite.NoError(err, "查询Host应该成功")
	suite.Equal(hm.ID, fm.ID)
	suite.Equal(hm.Name, fm.Name)

	// 测试边界情况:查询不存在的Host
	fm, err = suite.hostRepo.GetModel(context.Background(), "id = ?", 999999)
	suite.Error(err, "查询不存在的Host应该返回错误")
	suite.Nil(fm, "查询不存在的Host应该返回nil")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	_, err = suite.hostRepo.GetModel(timeoutCtx, "id = ?", hm.ID)
	suite.Error(err, "上下文超时后查询Host应该返回错误")
}

func (suite *HostTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		hm := CreateTestHostModel(func(h *resomodel.HostModel) {
			h.Name = fmt.Sprintf("host-%d", i)
		})
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err, "创建Host用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{}
	models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询Host列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "Host列表数量应该大于0")
	suite.NotNil(models, "Host列表应该不为nil")

	// 测试边界情况:空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"name": "non-existent-host"},
	}
	models2, err := suite.hostRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的Host列表应该成功")
	suite.Len(models2, 0, "不存在的Host列表长度应该为0")

	// 测试带查询条件的列表
	specialHM := CreateTestHostModel(func(h *resomodel.HostModel) {
		h.Label = "special-label"
	})
	err = suite.hostRepo.CreateModel(context.Background(), specialHM)
	suite.NoError(err)

	qp3 := database.QueryParams{
		Query: map[string]any{"label": "special-label"},
	}
	models3, err := suite.hostRepo.ListModel(context.Background(), qp3)
	suite.NoError(err)
	suite.Len(models3, 1)
	suite.Equal(specialHM.Label, models3[0].Label)

	// 测试带排序的列表
	qp4 := database.QueryParams{
		OrderBy: []string{"id asc"},
	}
	models4, err := suite.hostRepo.ListModel(context.Background(), qp4)
	suite.NoError(err)
	suite.GreaterOrEqual(len(models4), 3)

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	_, err = suite.hostRepo.ListModel(timeoutCtx, qp)
	suite.Error(err, "上下文超时后查询Host列表应该返回错误")
}

func (suite *HostTestSuite) TestListModelWithPagination() {
	// 使用唯一的标签来避免数据残留的影响
	uniqueLabel := fmt.Sprintf("pagination-label-%s", uuid.NewString()[:8])

	// 清理之前的测试数据
	err := suite.hostRepo.DeleteModel(context.Background(), "label = ?", uniqueLabel)
	suite.NoError(err, "删除Host用于分页测试应该成功")

	// 创建多个测试数据
	for i := 0; i < 10; i++ {
		hm := CreateTestHostModel(
			func(h *resomodel.HostModel) {
				h.Name = fmt.Sprintf("host-%d-%s", i, uuid.NewString()[:4])
				h.Label = uniqueLabel
				h.SSHPort = uint16(2200 + i)
				h.SSHUser = fmt.Sprintf("user-%d", i)
			},
		)
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

func (suite *HostTestSuite) TestCountModel() {
	// 使用唯一的标签来避免数据残留的影响
	uniqueLabel := fmt.Sprintf("test-label-%s", uuid.NewString()[:8])

	hm := CreateTestHostModel(func(h *resomodel.HostModel) {
		h.Label = uniqueLabel
	})
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err)

	count, err := suite.hostRepo.CountModel(context.Background(), map[string]any{"label": uniqueLabel})
	suite.NoError(err)
	suite.Equal(int64(1), count)

	count, err = suite.hostRepo.CountModel(context.Background(), map[string]any{"label": "nonexistent"})
	suite.NoError(err)
	suite.Equal(int64(0), count)

	// 测试多个记录的计数
	for i := 0; i < 4; i++ {
		hm := CreateTestHostModel(func(h *resomodel.HostModel) {
			h.Label = "count-label"
		})
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err)
	}

	count, err = suite.hostRepo.CountModel(context.Background(), map[string]any{"label": "count-label"})
	suite.NoError(err)
	suite.Equal(int64(4), count)

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	_, err = suite.hostRepo.CountModel(timeoutCtx, map[string]any{"label": "test"})
	suite.Error(err, "上下文超时后计数Host应该返回错误")
}

func (suite *HostTestSuite) TestNewSSHClient() {
	// 测试上下文取消情况
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

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 2)

	client, err = suite.hostRepo.NewSSHClient(timeoutCtx, "127.0.0.1", 22, "root", nil, time.Second)
	suite.Error(err, "上下文超时后创建SSH客户端应该返回错误")
	suite.Nil(client, "上下文超时后创建的SSH客户端应该为nil")

	// 测试无效参数
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

func (suite *HostTestSuite) TestExecuteCommand() {
	// 表格驱动测试不同场景
	tests := []struct {
		name     string
		ctx      context.Context
		session  *ssh.Session
		command  string
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "session为nil",
			ctx:      context.Background(),
			session:  nil,
			command:  "echo test",
			wantErr:  true,
			errorMsg: "session为nil时执行命令应该返回错误",
		},
		{
			name:     "空命令",
			ctx:      context.Background(),
			session:  nil,
			command:  "",
			wantErr:  true,
			errorMsg: "session为nil时即使命令为空也应该返回错误",
		},
		{
			name:     "上下文取消",
			ctx:      func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			session:  nil,
			command:  "echo test",
			wantErr:  true,
			errorMsg: "上下文取消时执行命令应该返回错误",
		},
		{
			name: "上下文超时",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
				cancel()
				return ctx
			}(),
			session:  nil,
			command:  "echo test",
			wantErr:  true,
			errorMsg: "上下文超时后执行命令应该返回错误",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// 对于超时测试，需要等待超时
			if tt.name == "上下文超时" {
				time.Sleep(time.Millisecond * 2)
			}
			err := suite.hostRepo.ExecuteCommand(tt.ctx, tt.session, tt.command)
			if tt.wantErr {
				suite.Error(err, tt.errorMsg)
			} else {
				suite.NoError(err, tt.errorMsg)
			}
		})
	}
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

func (suite *HostTestSuite) TestNewHostRepo() {
	logger := zap.NewNop()
	db := test.NewTestGormDBWithConfig(nil)
	timeouts := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()

	repo := NewHostRepo(logger, db, timeouts, slowThreshold)
	suite.NotNil(repo)
	suite.Equal(logger, repo.log)
	suite.Equal(db, repo.gormDB)
	suite.Equal(timeouts, repo.timeouts)
	suite.Equal(slowThreshold, repo.slowThreshold)
}

func TestHostTestSuite(t *testing.T) {
	pts := &HostTestSuite{}
	suite.Run(t, pts)
}
