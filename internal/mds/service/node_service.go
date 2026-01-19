package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbNode "gin-artweb/api/mds/node"
	"gin-artweb/internal/mds/biz"
	svReso "gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MdsNodeService struct {
	log    *zap.Logger
	ucNode *biz.MdsNodeUsecase
}

func NewMdsNodeService(
	logger *zap.Logger,
	ucNode *biz.MdsNodeUsecase,
) *MdsNodeService {
	return &MdsNodeService{
		log:    logger,
		ucNode: ucNode,
	}
}

// @Summary 创建mds节点
// @Description 本接口用于创建新的mds节点
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param request body pbNode.CreateOrUpdateMdsNodeRequest true "创建mds节点请求"
// @Success 200 {object} pbNode.MdsNodeReply "成功返回mds节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node [post]
// @Security ApiKeyAuth
func (s *MdsNodeService) CreateMdsNode(ctx *gin.Context) {
	var req pbNode.CreateOrUpdateMdsNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建mds节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	colony := biz.MdsNodeModel{
		NodeRole:    req.NodeRole,
		IsEnable:    req.IsEnable,
		HostID:      req.HostID,
		MdsColonyID: req.MdsColonyID,
	}

	m, rErr := s.ucNode.CreateMdsNode(ctx, colony)
	if rErr != nil {
		s.log.Error(
			"创建mds节点失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.MdsNodeReply{
		Code: http.StatusOK,
		Data: *MdsNodeToDetailOut(*m),
	})
}

// @Summary 更新mds节点
// @Description 本接口用于更新指定ID的mds节点
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "mds节点编号"
// @Param request body pbNode.CreateOrUpdateMdsNodeRequest true "更新mds节点请求"
// @Success 200 {object} pbNode.MdsNodeReply "成功返回mds节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node/{pk} [put]
// @Security ApiKeyAuth
func (s *MdsNodeService) UpdateMdsNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新mds节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbNode.CreateOrUpdateMdsNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新mds节点参数失败",
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
		"mds_colony_id": req.MdsColonyID,
	}

	m, err := s.ucNode.UpdateMdsNodeByID(ctx, uri.PK, data)
	if err != nil {
		s.log.Error(
			"更新mds节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.MdsNodeReply{
		Code: http.StatusOK,
		Data: *MdsNodeToDetailOut(*m),
	})
}

// @Summary 删除mds节点
// @Description 本接口用于删除指定ID的mds节点
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "mds节点编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node/{pk} [delete]
// @Security ApiKeyAuth
func (s *MdsNodeService) DeleteMdsNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除mds节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除mds节点",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucNode.DeleteMdsNodeByID(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除mds节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除mds节点成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询mds节点详情
// @Description 本接口用于查询指定ID的mds节点详情
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "mds节点编号"
// @Success 200 {object} pbNode.MdsNodeReply "成功返回mds节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node/{pk} [get]
// @Security ApiKeyAuth
func (s *MdsNodeService) GetMdsNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询mds节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询mds节点详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucNode.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询mds节点详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询mds节点详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := MdsNodeToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbNode.MdsNodeReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mds节点列表
// @Description 本接口用于查询mds节点列表
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "mds节点名称"
// @Param is_enabled query bool false "是否启用"
// @Param username query string false "创建用户名"
// @Success 200 {object} pbNode.PagMdsNodeReply "成功返回mds节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node [get]
// @Security ApiKeyAuth
func (s *MdsNodeService) ListMdsNode(ctx *gin.Context) {
	var req pbNode.ListMdsNodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询mds节点列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询mds节点列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"MdsColony", "Host"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucNode.ListMdsNode(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mds节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询mds节点列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListMdsNodeToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbNode.PagMdsNodeReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *MdsNodeService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateMdsNode)
	r.PUT("/node/:pk", s.UpdateMdsNode)
	r.DELETE("/node/:pk", s.DeleteMdsNode)
	r.GET("/node/:pk", s.GetMdsNode)
	r.GET("/node", s.ListMdsNode)
}

func MdsNodeToBaseOut(
	m biz.MdsNodeModel,
) *pbNode.MdsNodeBaseOut {
	return &pbNode.MdsNodeBaseOut{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		IsEnable: m.IsEnable,
	}
}

func MdsNodeToStandardOut(
	m biz.MdsNodeModel,
) *pbNode.MdsNodeStandardOut {
	return &pbNode.MdsNodeStandardOut{
		MdsNodeBaseOut: *MdsNodeToBaseOut(m),
		CreatedAt:      m.CreatedAt.String(),
		UpdatedAt:      m.UpdatedAt.String(),
	}
}

func MdsNodeToDetailOut(
	m biz.MdsNodeModel,
) *pbNode.MdsNodeDetailOut {
	return &pbNode.MdsNodeDetailOut{
		MdsNodeStandardOut: *MdsNodeToStandardOut(m),
		MdsColony:          MdsColonyToBaseOut(m.MdsColony),
		Host:               svReso.HostModelToBaseOut(m.Host),
	}
}

func ListMdsNodeToDetailOut(
	rms *[]biz.MdsNodeModel,
) *[]pbNode.MdsNodeStandardOut {
	if rms == nil {
		return &[]pbNode.MdsNodeStandardOut{}
	}

	ms := *rms
	mso := make([]pbNode.MdsNodeStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MdsNodeToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
