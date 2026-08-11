package resource

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repo/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
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
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建主机:开始执行",
		zap.Object("create_host_dto", &dto),
	)

	if err := s.TestSSHConnection(ctx, dto.SSHIP, dto.SSHPort, dto.SSHUser, dto.SSHPassword); err != nil {
		log.Error(
			"创建主机:测试SSH连接失败",
			zap.Error(err),
			zap.String("ssh_ip", dto.SSHIP),
			zap.Uint16("ssh_port", dto.SSHPort),
			zap.String("ssh_user", dto.SSHUser),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	m := dto.ToModel()
	if err := s.hostRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建主机:创建数据库模型失败",
			zap.Error(err),
			zap.Object("host_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	outputPath := GetHostVarsExportPath(m.ID)
	if err := s.ExportHost(ctx, m, outputPath); err != nil {
		log.Error(
			"创建主机:导出主机变量失败",
			zap.Error(err),
			zap.Object("host_model", &m),
			zap.String("output_path", outputPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建主机:执行成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *HostService) UpdateHostById(
	ctx context.Context,
	hostID uint32,
	dto resomodel.HostUpsertDTO,
) (*resomodel.HostModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新主机:开始执行",
		zap.Uint32("host_id", hostID),
		zap.Object("update_host_dto", &dto),
	)

	if err := s.TestSSHConnection(ctx, dto.SSHIP, dto.SSHPort, dto.SSHUser, dto.SSHPassword); err != nil {
		log.Error(
			"更新主机:测试SSH连接失败",
			zap.Error(err),
			zap.Uint32("host_id", hostID),
			zap.String("ssh_ip", dto.SSHIP),
			zap.Uint16("ssh_port", dto.SSHPort),
			zap.String("ssh_user", dto.SSHUser),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	updateData := dto.ToUpdateMap()
	if err := s.hostRepo.UpdateModel(ctx, updateData, "id = ?", hostID); err != nil {
		log.Error(
			"更新主机:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("host_id", hostID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}

	m, err := s.FindHostById(ctx, hostID)
	if err != nil {
		log.Error(
			"更新主机:查询更新后的主机详情失败",
			zap.Error(err),
			zap.Uint32("host_id", hostID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	outportPath := GetHostVarsExportPath(hostID)
	if err := s.ExportHost(ctx, *m, outportPath); err != nil {
		log.Error(
			"更新主机:导出主机变量失败, 请手动处理",
			zap.Error(err),
			zap.Object("host_model", m),
			zap.String("output_path", outportPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"更新主机:执行成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *HostService) DeleteHostById(
	ctx context.Context,
	hostId uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除主机:开始执行",
		zap.Uint32("host_id", hostId),
	)

	if err := s.hostRepo.DeleteModel(ctx, hostId); err != nil {
		log.Error(
			"删除主机:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": hostId})
	}

	outportPath := GetHostVarsExportPath(hostId)
	if err := fileutil.Remove(ctx, outportPath); err != nil {
		log.Error(
			"删除主机:删除ansible主机变量文件失败, 请手动删除",
			zap.Error(err),
			zap.String("outport_path", outportPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	log.Info(
		"删除主机:执行成功",
		zap.Uint32("host_id", hostId),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *HostService) FindHostById(
	ctx context.Context,
	hostId uint32,
) (*resomodel.HostModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.hostRepo.GetModel(ctx, hostId)
	if err != nil {
		log.Error(
			"查询主机:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": hostId})
	}

	log.Debug(
		"查询主机:查询到的数据库模型详情",
		zap.Object("host_model", m),
	)
	return m, nil
}

func (s *HostService) ListHost(
	ctx context.Context,
	dto resomodel.ListHostDTO,
) (int, int, int64, []resomodel.HostModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, 0, 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询主机列表:参数详情",
		zap.Object("list_host_dto", &dto),
	)

	page, size := dto.StandardModelQuery.GetPageParam()
	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	count, err := s.hostRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询主机列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询主机列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, nil
	}

	ms, err := s.hostRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询主机列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}
	return page, size, count, ms, nil
}

func (s *HostService) TestSSHConnection(
	ctx context.Context,
	sshIP string,
	sshPort uint16,
	sshUser, sshPassword string,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"测试SSH连接:入参详情",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
	)

	if cli, err := s.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, []ssh.AuthMethod{s.authMethod}, s.sshTimeout); err == nil {
		_ = cli.Close()
		return nil
	}

	sshAuths := []ssh.AuthMethod{
		ssh.Password(sshPassword),
	}

	client, err := s.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, s.sshTimeout)
	if err != nil {
		log.Error(
			"测试SSH连接:创建SSH连接失败",
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
			"测试SSH连接:创建SSH session失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrSSHConnectionFailed.WithCause(err)
	}
	defer session.Close()

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
				"测试SSH连接:部署SSH公钥失败",
				zap.Error(err),
				zap.String("ssh_ip", sshIP),
				zap.Uint16("ssh_port", sshPort),
				zap.String("ssh_user", sshUser),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.ErrSSHKeyDeployFailed.WithCause(err)
		}
	}
	return nil
}

func (s *HostService) ExportHost(
	ctx context.Context,
	m resomodel.HostModel,
	outputPath string,
) *errors.Error {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"导出主机变量:入参详情",
		zap.Object("host_model", &m),
		zap.String("output_path", outputPath),
	)

	ansibleHost := resomodel.HostModelToAnsibleHostVars(m)
	if _, err := serializer.WriteYAML(outputPath, ansibleHost); err != nil {
		log.Error(
			"导出主机变量:写入文件失败",
			zap.Error(err),
			zap.String("output_path", outputPath),
			zap.Object("ansible_host", &ansibleHost),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	return nil
}

func GetHostVarsExportPath(pk uint32) string {
	filename := fmt.Sprintf("host_%d.yaml", pk)
	return filepath.Join(config.StorageDir, "host_vars", filename)
}
