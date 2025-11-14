package data

import (
	"context"
	"encoding/base64"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/pkg/database"
)

type hostRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
}

func NewHostRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	publicKey []byte,
) biz.HostRepo {
	return &hostRepo{
		log:    log,
		gormDB: gormDB,
	}
}

func (r *hostRepo) CreateModel(ctx context.Context, m *biz.HostModel) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.HostModel{}, m); err != nil {
		r.log.Error(
			"新增主机模型失败",
			zap.Object(database.ModelKey, m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *hostRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	if err := database.DBUpdate(ctx, r.gormDB, &biz.HostModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新主机模型失败",
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *hostRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.HostModel{}, conds...); err != nil {
		r.log.Error(
			"删除主机模型失败",
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *hostRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.HostModel, error) {
	var m biz.HostModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询主机模型失败",
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *hostRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.HostModel, error) {
	var ms []biz.HostModel
	count, err := database.DBList(ctx, r.gormDB, &biz.HostModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询主机列表失败",
			zap.Object(database.QueryParamsKey, &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *hostRepo) ListModelByIds(
	ctx context.Context,
	ids []uint32,
) ([]biz.HostModel, error) {
	if len(ids) == 0 {
		return []biz.HostModel{}, nil
	}
	qp := database.NewPksQueryParams(ids)
	_, ms, err := r.ListModel(ctx, qp)
	return ms, err
}

func (r *hostRepo) NewSSHClient(
	ctx context.Context,
	addr string,
	c ssh.ClientConfig,
) (*ssh.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	client, err := ssh.Dial("tcp", addr, &c)
	if err != nil {
		r.log.Error(
			"创建ssh连接失败",
			zap.String("addr", addr),
			zap.String("username", c.User),
			zap.Error(err),
		)
		return nil, err
	}
	return client, nil
}

func (r *hostRepo) DeployPublicKey(c *ssh.Client, key ssh.PublicKey) error {
	session, err := c.NewSession()
	if err != nil {
		r.log.Error(
			"创建ssh会话失败",
			zap.Error(err),
		)
		return err
	}
	defer session.Close()

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
			"部署SSH密钥失败",
			zap.Error(err),
		)
		return err
	}
	return nil
}
