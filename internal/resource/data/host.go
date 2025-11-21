package data

import (
	"context"
	"encoding/base64"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
	"gin-artweb/pkg/log"
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
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.HostModel{}, m); err != nil {
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

func (r *hostRepo) DeployPublicKey(ctx context.Context, c *ssh.Client, key ssh.PublicKey) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}
	r.log.Debug(
		"开始创建ssh会话",
		zap.String("local_addr", c.LocalAddr().String()),
		zap.String("remote_addr", c.RemoteAddr().String()),
		zap.String("username", c.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	session, err := c.NewSession()
	if err != nil {
		r.log.Error(
			"创建ssh会话失败",
			zap.Error(err),
			zap.String("local_addr", c.LocalAddr().String()),
			zap.String("remote_addr", c.RemoteAddr().String()),
			zap.String("username", c.User()),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	defer session.Close()

	r.log.Debug(
		"创建ssh会话成功",
		zap.String("local_addr", c.LocalAddr().String()),
		zap.String("remote_addr", c.RemoteAddr().String()),
		zap.String("username", c.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)

	r.log.Debug(
		"开始部署ssh密钥",
		zap.String("local_addr", c.LocalAddr().String()),
		zap.String("remote_addr", c.RemoteAddr().String()),
		zap.String("username", c.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	depStartTime := time.Now()
	pubKeyBytes := ssh.MarshalAuthorizedKey(key)
	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKeyBytes)
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
	if err := session.Run(script); err != nil {
		r.log.Error(
			"部署ssh密钥失败",
			zap.Error(err),
			zap.String("local_addr", c.LocalAddr().String()),
			zap.String("remote_addr", c.RemoteAddr().String()),
			zap.String("username", c.User()),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(depStartTime)),
		)
		return err
	}
	r.log.Debug(
		"部署ssh密钥成功",
		zap.String("local_addr", c.LocalAddr().String()),
		zap.String("remote_addr", c.RemoteAddr().String()),
		zap.String("username", c.User()),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(depStartTime)),
	)
	return nil
}
