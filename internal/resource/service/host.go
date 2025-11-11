package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	cpb "gin-artweb/api/common"
	pb "gin-artweb/api/resource/host"
	"gin-artweb/internal/resource/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/errors"
)

type HostService struct {
	log    *zap.Logger
	ucHost *biz.HostUsecase
}

func NewHostService(
	logger *zap.Logger,
	ucHost *biz.HostUsecase,
) *HostService {
	return &HostService{
		log:    logger,
		ucHost: ucHost,
	}
}

// @Summary 新增主机
// @Description 本接口用于新增主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param request body pb.CreateHosrRequest true "创建主机请求"
// @Success 200 {object} pb.HostReply "成功返回主机信息"
// @Router /api/v1/resource/Host [post]
func (s *HostService) CreateHost(ctx *gin.Context) {
	var req pb.CreateHosrRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucHost.CreateHost(ctx, biz.HostModel{
		Name:     req.Name,
		Label:    req.Label,
		IPAddr:   req.IPAddr,
		Port:     req.Port,
		Username: req.Username,
		PyPath:   req.PyPath,
		Remark:   req.Remark,
	}, req.Password)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := HostModelToOutBase(*m)
	ctx.JSON(http.StatusCreated, &pb.HostReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新主机
// @Description 本接口用于更新主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Param request body pb.UpdateHostRequest true "更新主机请求"
// @Success 200 {object} pb.HostReply "成功返回主机信息"
// @Router /api/v1/resource/host/{pk} [put]
func (s *HostService) UpdateHost(ctx *gin.Context) {
	var uri cpb.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.UpdateHostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucHost.UpdateHostById(ctx, uri.PK, biz.HostModel{
		Name:     req.Name,
		Label:    req.Label,
		IPAddr:   req.IPAddr,
		Port:     req.Port,
		Username: req.Username,
		PyPath:   req.PyPath,
		Remark:   req.Remark,
	}, req.Password); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucHost.FindHostById(ctx, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := HostModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pb.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除主机
// @Description 本接口用于删除指定ID的主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Router /api/v1/resource/Host/{pk} [delete]
func (s *HostService) DeleteHost(ctx *gin.Context) {
	var uri cpb.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucHost.DeleteHostById(ctx, uri.PK); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, common.NoDataReply)
}

// @Summary 查询单个主机
// @Description 本接口用于查询一个主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Success 200 {object} pb.HostReply "成功返回用户信息"
// @Router /api/v1/resource/Host/{pk} [get]
func (s *HostService) GetHost(ctx *gin.Context) {
	var uri cpb.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucHost.FindHostById(ctx, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := HostModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pb.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询主机列表
// @Description 本接口用于查询主机列表
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param pk query uint false "主机主键，可选参数，如果提供则必须大于0"
// @Param pks query string false "主机主键列表，可选参数，多个用,隔开，如1,2,3"
// @Param before_create_at query string false "创建时间之前的记录 (RFC3339格式)"
// @Param after_create_at query string false "创建时间之后的记录 (RFC3339格式)"
// @Param before_update_at query string false "更新时间之前的记录 (RFC3339格式)"
// @Param after_update_at query string false "更新时间之后的记录 (RFC3339格式)"
// @Param http_url query string false "HTTP路径"
// @Param method query string false "HTTP方法"
// @Success 200 {object} pb.PagHostReply "成功返回主机列表"
// @Router /api/v1/resource/Host [get]
func (s *HostService) ListHost(ctx *gin.Context) {
	var req pb.ListHostRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucHost.ListHost(ctx, page, size, query, []string{"id"}, true)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListHostModelToOut(ms)
	ctx.JSON(http.StatusOK, &pb.PagHostReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

func (s *HostService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/Host", s.CreateHost)
	r.PUT("/Host/:pk", s.UpdateHost)
	r.DELETE("/Host/:pk", s.DeleteHost)
	r.GET("/Host/:pk", s.GetHost)
	r.GET("/Host", s.ListHost)
}

func HostModelToOutBase(
	m biz.HostModel,
) *pb.HostOutBase {
	return &pb.HostOutBase{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Name:      m.Name,
		Label:     m.Label,
		IPAddr:    m.IPAddr,
		Port:      m.Port,
		Username:  m.Username,
		PyPath:    m.PyPath,
		Remark:    m.Remark,
	}
}

func ListHostModelToOut(
	ms []biz.HostModel,
) []*pb.HostOutBase {
	mso := make([]*pb.HostOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := HostModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
