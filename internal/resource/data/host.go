package data

import (
	"context"
	"net"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/pkg/database"
)

type hostRepo struct {
	log        *zap.Logger
	gormDB     *gorm.DB
	privateKey string
}

func NewHostRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
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
			zap.Object("model", m),
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
			zap.Any("data", data),
			zap.Any("conditions", conds),
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
			zap.Any("conditions", conds),
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
			zap.Strings("preloads", preloads),
			zap.Any("conditions", conds),
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
			zap.Object("query_params", &qp),
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
	m biz.HostModel,
) (*ssh.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	key, err := os.ReadFile(r.privateKey)
	if err != nil {
		r.log.Error("读取ssh私钥文件失败", zap.Error(err))
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		r.log.Error("解析私钥失败", zap.Error(err))
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: m.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := net.JoinHostPort(m.IPAddr, strconv.FormatUint(uint64(m.Port), 10))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		r.log.Error(
			"创建ssh连接失败",
			zap.String("ip_addr", m.IPAddr),
			zap.Uint16("port", m.Port),
			zap.String("username", m.Username),
			zap.Error(err),
		)
		return nil, err
	}
	return client, nil
}
