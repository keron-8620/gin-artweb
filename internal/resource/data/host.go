package data

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/internal/shared/log"
)

type hostRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *database.DBTimeout
}

func NewHostRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *database.DBTimeout,
) biz.HostRepo {
	return &hostRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *hostRepo) CreateModel(ctx context.Context, m *biz.HostModel) error {
	r.log.Debug(
		"开始创建主机模型",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.HostModel{}, m, nil); err != nil {
		r.log.Error(
			"创建主机模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建主机模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *hostRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	r.log.Debug(
		"开始更新主机模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.HostModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新主机模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"更新主机模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *hostRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除主机模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.HostModel{}, conds...); err != nil {
		r.log.Error(
			"删除主机模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"删除主机模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *hostRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.HostModel, error) {
	r.log.Debug(
		"开始查询主机模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.HostModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBFind(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询主机模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询主机模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *hostRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.HostModel, error) {
	r.log.Debug(
		"开始查询主机模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.HostModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.HostModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询主机模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询主机模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}

func (r *hostRepo) NewSSHClient(
	ctx context.Context,
	addr string,
	c ssh.ClientConfig,
) (*ssh.Client, error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, err
	}
	r.log.Debug(
		"开始创建ssh连接",
		zap.String("addr", addr),
		zap.String("username", c.User),
	)
	startTime := time.Now()
	client, err := ssh.Dial("tcp", addr, &c)
	if err != nil {
		r.log.Error(
			"创建ssh连接失败",
			zap.Error(err),
			zap.String("addr", addr),
			zap.String("username", c.User),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}
	r.log.Debug(
		"创建ssh连接成功",
		zap.String("addr", addr),
		zap.String("username", c.User),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return client, nil
}

func (r *hostRepo) NewSession(
	ctx context.Context,
	client *ssh.Client,
) (*ssh.Session, error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, err
	}

	r.log.Debug(
		"开始创建ssh会话",
		zap.String("local_addr", client.LocalAddr().String()),
		zap.String("remote_addr", client.RemoteAddr().String()),
		zap.String("username", client.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	session, err := client.NewSession()
	if err != nil {
		r.log.Error(
			"创建ssh会话失败",
			zap.Error(err),
			zap.String("local_addr", client.LocalAddr().String()),
			zap.String("remote_addr", client.RemoteAddr().String()),
			zap.String("username", client.User()),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}

	r.log.Debug(
		"创建ssh会话成功",
		zap.String("local_addr", client.LocalAddr().String()),
		zap.String("remote_addr", client.RemoteAddr().String()),
		zap.String("username", client.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return session, nil
}

func (r *hostRepo) ExecuteCommand(
	ctx context.Context,
	session *ssh.Session,
	command string,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	r.log.Debug(
		"开始执行命令",
		zap.String("command", command),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	if err := session.Run(command); err != nil {
		r.log.Error(
			"执行命令失败",
			zap.Error(err),
			zap.String("command", command),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}

	r.log.Debug(
		"执行命令成功",
		zap.String("command", command),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)

	return nil
}

func (r *hostRepo) NewSFTPClient(
	ctx context.Context,
	client *ssh.Client,
) (*sftp.Client, error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, err
	}

	r.log.Debug(
		"开始创建sftp客户端",
		zap.String("local_addr", client.LocalAddr().String()),
		zap.String("remote_addr", client.RemoteAddr().String()),
		zap.String("username", client.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		r.log.Error(
			"创建SFTP客户端失败",
			zap.Error(err),
			zap.String("local_addr", client.LocalAddr().String()),
			zap.String("remote_addr", client.RemoteAddr().String()),
			zap.String("username", client.User()),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}

	r.log.Debug(
		"创建sftp客户端成功",
		zap.String("local_addr", client.LocalAddr().String()),
		zap.String("remote_addr", client.RemoteAddr().String()),
		zap.String("username", client.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return sftpClient, nil
}

// 上传文件
func (r *hostRepo) UploadFile(
	ctx context.Context,
	client *sftp.Client,
	src, dest string,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	r.log.Debug(
		"开始上传文件",
		zap.String("src", src),
		zap.String("dest", dest),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()

	// 打开源文件
	localFile, err := os.Open(src)
	if err != nil {
		r.log.Error(
			"打开本地文件失败",
			zap.Error(err),
			zap.String("src", src),
			zap.String("dest", dest),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	defer localFile.Close()

	// 创建目标文件
	remoteFile, err := client.Create(dest)
	if err != nil {
		r.log.Error(
			"创建远程文件失败",
			zap.Error(err),
			zap.String("src", src),
			zap.String("dest", dest),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	defer remoteFile.Close()

	// 复制文件内容
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		r.log.Error(
			"复制文件内容失败",
			zap.Error(err),
			zap.String("src", src),
			zap.String("dest", dest),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}

	r.log.Debug(
		"上传文件成功",
		zap.String("src", src),
		zap.String("dest", dest),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)

	return nil
}

// 下载文件
func (r *hostRepo) DownloadFile(
	ctx context.Context,
	client *sftp.Client,
	src, dest string,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	r.log.Debug(
		"开始下载文件",
		zap.String("src", src),
		zap.String("dest", dest),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()

	// 打开远程文件
	remoteFile, err := client.Open(src)
	if err != nil {
		r.log.Error(
			"打开远程文件失败",
			zap.Error(err),
			zap.String("src", src),
			zap.String("dest", dest),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	defer remoteFile.Close()

	// 创建本地文件
	localFile, err := os.Create(dest)
	if err != nil {
		r.log.Error(
			"创建本地文件失败",
			zap.Error(err),
			zap.String("src", src),
			zap.String("dest", dest),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	defer localFile.Close()

	// 复制文件内容
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		r.log.Error(
			"复制文件内容失败",
			zap.Error(err),
			zap.String("src", src),
			zap.String("dest", dest),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}

	r.log.Debug(
		"下载文件成功",
		zap.String("src", src),
		zap.String("dest", dest),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)

	return nil
}
