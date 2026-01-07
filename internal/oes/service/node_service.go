package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbNode "gin-artweb/api/oes/node"
	"gin-artweb/internal/oes/biz"
	servReso "gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type OesNodeService struct {
	log    *zap.Logger
	ucNode *biz.OesNodeUsecase
}

func NewOesNodeService(
	logger *zap.Logger,
	ucNode *biz.OesNodeUsecase,
) *OesNodeService {
	return &OesNodeService{
		log:    logger,
		ucNode: ucNode,
	}
}

// @Summary 创建oes节点
// @Description 本接口用于创建新的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param request body pbNode.CreateOrUpdateOesNodeRequest true "创建oes节点请求"
// @Success 200 {object} pbNode.OesNodeReply "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node [post]
// @Security ApiKeyAuth
func (s *OesNodeService) CreateOesNode(ctx *gin.Context) {
	var req pbNode.CreateOrUpdateOesNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建oes节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	colony := biz.OesNodeModel{
		NodeRole:    req.NodeRole,
		IsEnable:    req.IsEnable,
		HostID:      req.HostID,
		OesColonyID: req.OesColonyID,
	}

	m, rErr := s.ucNode.CreateOesNode(ctx, colony)
	if rErr != nil {
		s.log.Error(
			"创建oes节点失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.OesNodeReply{
		Code: http.StatusOK,
		Data: *OesNodeToDetailOut(*m),
	})
}

// @Summary 更新oes节点
// @Description 本接口用于更新指定ID的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "oes节点编号"
// @Param request body pbNode.CreateOrUpdateOesNodeRequest true "更新oes节点请求"
// @Success 200 {object} pbNode.OesNodeReply "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{pk} [put]
// @Security ApiKeyAuth
func (s *OesNodeService) UpdateOesNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新oes节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbNode.CreateOrUpdateOesNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新oes节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	data := map[string]any{
		"node_role":     req.NodeRole,
		"is_enable":     req.IsEnable,
		"host_id":       req.HostID,
		"oes_colony_id": req.OesColonyID,
	}

	m, err := s.ucNode.UpdateOesNodeByID(ctx, uri.PK, data)
	if err != nil {
		s.log.Error(
			"更新oes节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.OesNodeReply{
		Code: http.StatusOK,
		Data: *OesNodeToDetailOut(*m),
	})
}

// @Summary 删除oes节点
// @Description 本接口用于删除指定ID的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "oes节点编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{pk} [delete]
// @Security ApiKeyAuth
func (s *OesNodeService) DeleteOesNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除oes节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除oes节点",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucNode.DeleteOesNodeByID(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除oes节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除oes节点成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询oes节点详情
// @Description 本接口用于查询指定ID的oes节点详情
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "oes节点编号"
// @Success 200 {object} pbNode.OesNodeReply "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{pk} [get]
// @Security ApiKeyAuth
func (s *OesNodeService) GetOesNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询oes节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询oes节点详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucNode.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询oes节点详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询oes节点详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := OesNodeToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbNode.OesNodeReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询oes节点列表
// @Description 本接口用于查询oes节点列表
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "oes节点名称"
// @Param is_enabled query bool false "是否启用"
// @Param username query string false "创建用户名"
// @Success 200 {object} pbNode.PagOesNodeReply "成功返回oes节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node [get]
// @Security ApiKeyAuth
func (s *OesNodeService) ListOesNode(ctx *gin.Context) {
	var req pbNode.ListOesNodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes节点列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询oes节点列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"OesColony", "Host"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucNode.ListOesNode(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询oes节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询oes节点列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListOesNodeToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbNode.PagOesNodeReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *OesNodeService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateOesNode)
	r.PUT("/node/:pk", s.UpdateOesNode)
	r.DELETE("/node/:pk", s.DeleteOesNode)
	r.GET("/node/:pk", s.GetOesNode)
	r.GET("/node", s.ListOesNode)
}

func OesNodeToBaseOut(
	m biz.OesNodeModel,
) *pbNode.OesNodeBaseOut {
	return &pbNode.OesNodeBaseOut{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		IsEnable: m.IsEnable,
	}
}

func OesNodeToStandardOut(
	m biz.OesNodeModel,
) *pbNode.OesNodeStandardOut {
	return &pbNode.OesNodeStandardOut{
		OesNodeBaseOut: *OesNodeToBaseOut(m),
		CreatedAt:      m.CreatedAt.String(),
		UpdatedAt:      m.UpdatedAt.String(),
	}
}

func OesNodeToDetailOut(
	m biz.OesNodeModel,
) *pbNode.OesNodeDetailOut {
	return &pbNode.OesNodeDetailOut{
		OesNodeStandardOut: *OesNodeToStandardOut(m),
		OesColony:          OesColonyToBaseOut(m.OesColony),
		Host:               servReso.HostModelToBaseOut(m.Host),
	}
}

func ListOesNodeToDetailOut(
	rms *[]biz.OesNodeModel,
) *[]pbNode.OesNodeStandardOut {
	if rms == nil {
		return &[]pbNode.OesNodeStandardOut{}
	}

	ms := *rms
	mso := make([]pbNode.OesNodeStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := OesNodeToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
