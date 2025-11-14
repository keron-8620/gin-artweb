package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbHost "gin-artweb/api/resource/host"
	"gin-artweb/internal/resource/biz"
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

// @Summary 创建主机
// @Description 本接口用于创建新的主机配置信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param request body pbHost.CreateHosrRequest true "创建主机请求"
// @Success 201 {object} pbHost.HostReply "创建主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [post]
// @Security ApiKeyAuth
func (s *HostService) CreateHost(ctx *gin.Context) {
	var req pbHost.CreateHosrRequest
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
	ctx.JSON(http.StatusCreated, &pbHost.HostReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新主机
// @Description 本接口用于更新指定ID的主机配置信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Param request body pbHost.UpdateHostRequest true "更新主机请求"
// @Success 200 {object} pbHost.HostReply "更新主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{pk} [put]
// @Security ApiKeyAuth
func (s *HostService) UpdateHost(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pbHost.UpdateHostRequest
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
	ctx.JSON(http.StatusOK, &pbHost.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除主机
// @Description 本接口用于删除指定ID的主机配置信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{pk} [delete]
// @Security ApiKeyAuth
func (s *HostService) DeleteHost(ctx *gin.Context) {
	var uri pbComm.PKUri
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
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询主机详情
// @Description 本接口用于查询指定ID的主机详细信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Success 200 {object} pbHost.HostReply "获取主机详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{pk} [get]
// @Security ApiKeyAuth
func (s *HostService) GetHost(ctx *gin.Context) {
	var uri pbComm.PKUri
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
	ctx.JSON(http.StatusOK, &pbHost.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询主机列表
// @Description 本接口用于查询主机配置信息列表
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "主机名称"
// @Param label query string false "主机标签"
// @Param ip_addr query string false "IP地址"
// @Success 200 {object} pbHost.PagHostReply "成功返回主机列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [get]
// @Security ApiKeyAuth
func (s *HostService) ListHost(ctx *gin.Context) {
	var req pbHost.ListHostRequest
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
	ctx.JSON(http.StatusOK, &pbHost.PagHostReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *HostService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/host", s.CreateHost)
	r.PUT("/host/:pk", s.UpdateHost)
	r.DELETE("/host/:pk", s.DeleteHost)
	r.GET("/host/:pk", s.GetHost)
	r.GET("/host", s.ListHost)
}

func HostModelToOutBase(
	m biz.HostModel,
) *pbHost.HostOutBase {
	return &pbHost.HostOutBase{
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
) []*pbHost.HostOutBase {
	mso := make([]*pbHost.HostOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := HostModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
