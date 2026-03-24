package resource

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repo/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

type HostService struct {
	log        *zap.Logger
	hostRepo   *resorepo.HostRepo
	sshTimeout time.Duration
	authMethod ssh.AuthMethod
	pubKeyB64s []string
}

func NewHostService(
	log *zap.Logger,
	hostRepo *resorepo.HostRepo,
	sshTimeout time.Duration,
	authMethod ssh.AuthMethod,
	pubKeyB64s []string,
) *HostService {
	return &HostService{
		log:        log,
		hostRepo:   hostRepo,
		sshTimeout: sshTimeout,
		authMethod: authMethod,
		pubKeyB64s: pubKeyB64s,
	}
}

func (s *HostService) CreateHost(
	ctx context.Context,
	dto resomodel.HostUpsertDTO,
) (*resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建主机：开始执行")

	log.Debug(
		"创建主机：输入参数",
		zap.Object("create_host_dto", &dto),
	)

	testSSHStepStart := time.Now()
	log.Debug("创建主机：开始测试SSH连接")
	if err := s.TestSSHConnection(ctx, dto.SSHIP, dto.SSHPort, dto.SSHUser, dto.SSHPassword); err != nil {
		log.Error(
			"创建主机：测试SSH连接失败",
			zap.Error(err),
			zap.Duration("test_ssh_step_duration", time.Since(testSSHStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	testSSHStepDuration := time.Since(testSSHStepStart)
	log.Debug(
		"创建主机：测试SSH连接成功",
		zap.Duration("test_ssh_step_duration", testSSHStepDuration),
	)

	m := resomodel.HostModel{
		Name:    dto.Name,
		Label:   dto.Label,
		SSHIP:   dto.SSHIP,
		SSHPort: dto.SSHPort,
		SSHUser: dto.SSHUser,
		PyPath:  dto.PyPath,
		Remark:  dto.Remark,
	}
	createStepStart := time.Now()
	log.Debug(
		"创建主机：开始创建数据库模型",
		zap.Object("host_model", &m),
	)
	if err := s.hostRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建主机：创建数据库模型失败",
			zap.Error(err),
			zap.Object("host_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建主机：创建数据库模型成功",
		zap.Object("host_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	exportStepStart := time.Now()
	log.Debug(
		"创建主机：开始导出主机变量",
		zap.Uint32("host_id", m.ID),
	)
	if err := s.ExportHost(ctx, m); err != nil {
		log.Error(
			"创建主机：导出主机变量失败",
			zap.Error(err),
			zap.Uint32("host_id", m.ID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"创建主机：导出主机变量成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"创建主机：执行成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("test_ssh_step_duration", testSSHStepDuration),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *HostService) UpdateHostById(
	ctx context.Context,
	hostID uint32,
	dto resomodel.HostUpsertDTO,
) (*resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新主机：开始执行",
		zap.Uint32("host_id", hostID),
	)

	log.Debug(
		"更新主机：输入参数",
		zap.Uint32("host_id", hostID),
		zap.Object("update_host_dto", &dto),
	)

	testSSHStepStart := time.Now()
	log.Debug("更新主机：开始测试SSH连接")
	if err := s.TestSSHConnection(ctx, dto.SSHIP, dto.SSHPort, dto.SSHUser, dto.SSHPassword); err != nil {
		log.Error(
			"更新主机：测试SSH连接失败",
			zap.Error(err),
			zap.Uint32("host_id", hostID),
			zap.Duration("test_ssh_step_duration", time.Since(testSSHStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	testSSHStepDuration := time.Since(testSSHStepStart)
	log.Debug(
		"更新主机：测试SSH连接成功",
		zap.Uint32("host_id", hostID),
		zap.Duration("test_ssh_step_duration", testSSHStepDuration),
	)

	data := map[string]any{
		"name":     dto.Name,
		"label":    dto.Label,
		"ssh_ip":   dto.SSHIP,
		"ssh_port": dto.SSHPort,
		"ssh_user": dto.SSHUser,
		"py_path":  dto.PyPath,
		"remark":   dto.Remark,
	}

	updateStepStart := time.Now()
	log.Debug(
		"更新主机：开始更新数据库模型",
		zap.Any("update_data", data),
		zap.Uint32("host_id", hostID),
	)
	if err := s.hostRepo.UpdateModel(ctx, data, "id = ?", hostID); err != nil {
		log.Error(
			"更新主机：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("host_id", hostID),
			zap.Any("update_data", data),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, data)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新主机：更新数据库模型成功",
		zap.Uint32("host_id", hostID),
		zap.Any("update_data", data),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	m, err := s.FindHostById(ctx, hostID)
	if err != nil {
		log.Error(
			"更新主机：查询更新后的主机详情失败",
			zap.Error(err),
			zap.Uint32("host_id", hostID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	exportStepStart := time.Now()
	log.Debug(
		"更新主机：开始导出主机变量",
		zap.Uint32("host_id", m.ID),
	)
	if err := s.ExportHost(ctx, *m); err != nil {
		log.Error(
			"更新主机：导出主机变量失败",
			zap.Error(err),
			zap.Object("host_model", m),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"更新主机：导出主机变量成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"更新主机：执行成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("test_ssh_step_duration", testSSHStepDuration),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *HostService) DeleteHostById(
	ctx context.Context,
	hostId uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除主机：开始执行",
		zap.Uint32("host_id", hostId),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除主机：开始删除数据库模型",
		zap.Uint32("host_id", hostId),
	)
	if err := s.hostRepo.DeleteModel(ctx, hostId); err != nil {
		log.Error(
			"删除主机：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": hostId})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除主机：删除数据库模型成功",
		zap.Uint32("host_id", hostId),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	path := common.GetHostVarsExportPath(hostId)
	log.Debug(
		"删除主机：准备删除ansible主机变量文件",
		zap.String("path", path),
		zap.Uint32("host_id", hostId),
	)

	removeStepStart := time.Now()
	log.Debug(
		"删除主机：开始删除ansible主机变量文件",
		zap.String("path", path),
		zap.Uint32("host_id", hostId),
	)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		log.Error(
			"删除主机：删除ansible主机变量文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("host_id", hostId),
			zap.Duration("remove_step_duration", time.Since(removeStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}
	removeStepDuration := time.Since(removeStepStart)
	log.Debug(
		"删除主机：删除ansible主机变量文件成功",
		zap.String("path", path),
		zap.Uint32("host_id", hostId),
		zap.Duration("remove_step_duration", removeStepDuration),
	)

	log.Info(
		"删除主机：执行成功",
		zap.Uint32("host_id", hostId),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_step_duration", removeStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *HostService) FindHostById(
	ctx context.Context,
	hostId uint32,
) (*resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询主机：开始执行",
		zap.Uint32("host_id", hostId),
	)

	m, err := s.hostRepo.GetModel(ctx, nil, hostId)
	if err != nil {
		log.Error(
			"查询主机：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": hostId})
	}
	log.Debug(
		"查询主机：查询到的数据库模型详情",
		zap.Object("host_model", m),
	)

	log.Info(
		"查询主机：执行成功",
		zap.Uint32("host_id", hostId),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *HostService) ListHost(
	ctx context.Context,
	page, size int,
	dto resomodel.ListHostDTO,
) (int64, []resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询主机列表：开始执行")

	log.Debug(
		"查询主机列表：参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_host_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	log.Debug(
		"查询主机列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询主机列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.hostRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询主机列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询主机列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询主机列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询主机列表：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.hostRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询主机列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询主机列表：查询数据库模型列表成功",
		zap.Int("total_count", len(ms)),
		zap.Duration("list_step_duration", listStepDuration),
	)

	log.Info(
		"查询主机列表：执行成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *HostService) TestSSHConnection(
	ctx context.Context,
	sshIP string,
	sshPort uint16,
	sshUser, sshPassword string,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"测试SSH连接：开始执行",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
	)

	log.Debug("测试SSH连接：尝试使用已部署的密钥连接")
	cli, err := s.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, []ssh.AuthMethod{s.authMethod}, s.sshTimeout)
	if err == nil {
		log.Info(
			"测试SSH连接：使用已部署的密钥连接成功",
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		cli.Close()
		return nil
	}

	log.Debug("测试SSH连接：使用密码连接")
	sshAuths := []ssh.AuthMethod{
		ssh.Password(sshPassword),
	}

	client, err := s.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, s.sshTimeout)
	if err != nil {
		log.Error(
			"测试SSH连接：创建SSH连接失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrSSHConnectionFailed.WithCause(err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Error(
			"测试SSH连接：创建SSH session失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrSSHConnectionFailed.WithCause(err)
	}
	defer session.Close()

	deployKeyStepStart := time.Now()
	log.Debug("测试SSH连接：开始部署SSH公钥")
	for _, pubKeyB64 := range s.pubKeyB64s {
		script := `
			mkdir -p ~/.ssh
			tmp_key=$(mktemp)
			echo '` + pubKeyB64 + `' | base64 -d > "$tmp_key"
			if ! grep -Fq "$(cat "$tmp_key")" ~/.ssh/authorized_keys 2>/dev/null; then
				cat "$tmp_key" >> ~/.ssh/authorized_keys
			fi
			rm -f "$tmp_key"
			chmod 700 ~/.ssh
			chmod 600 ~/.ssh/authorized_keys
		`
		if err := s.hostRepo.ExecuteCommand(ctx, session, script); err != nil {
			log.Error(
				"测试SSH连接：部署SSH公钥失败",
				zap.Error(err),
				zap.String("ssh_ip", sshIP),
				zap.Uint16("ssh_port", sshPort),
				zap.String("ssh_user", sshUser),
				zap.Duration("deploy_key_step_duration", time.Since(deployKeyStepStart)),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.ErrSSHKeyDeployFailed.WithCause(err)
		}
	}
	deployKeyStepDuration := time.Since(deployKeyStepStart)
	log.Debug(
		"测试SSH连接：部署SSH公钥成功",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.Duration("deploy_key_step_duration", deployKeyStepDuration),
	)

	log.Info(
		"测试SSH连接：执行成功",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.Duration("deploy_key_step_duration", deployKeyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *HostService) ExportHost(ctx context.Context, m resomodel.HostModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出主机变量：开始执行",
		zap.Uint32("host_id", m.ID),
	)

	ansibleHost := resomodel.AnsibleHostVars{
		HostID:                   m.ID,
		AnsibleHost:              m.SSHIP,
		AnsiblePort:              m.SSHPort,
		AnsibleUser:              m.SSHUser,
		AnsiblePythonInterpreter: m.PyPath,
	}

	log.Debug(
		"导出主机变量：参数详情",
		zap.Object("ansible_host", &ansibleHost),
	)

	path := common.GetHostVarsExportPath(m.ID)
	log.Debug(
		"导出主机变量：准备写入文件",
		zap.String("path", path),
	)

	exportStepStart := time.Now()
	log.Debug(
		"导出主机变量：开始写入文件",
		zap.String("path", path),
		zap.Uint32("host_id", m.ID),
	)
	if _, err := serializer.WriteYAML(path, ansibleHost); err != nil {
		log.Error(
			"导出主机变量：写入文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Object("ansible_host", &ansibleHost),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"导出主机变量：写入文件成功",
		zap.String("path", path),
		zap.Uint32("host_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"导出主机变量：执行成功",
		zap.Uint32("host_id", m.ID),
		zap.String("path", path),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
