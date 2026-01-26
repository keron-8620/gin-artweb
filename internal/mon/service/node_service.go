package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbNode "gin-artweb/api/mon/node"
	"gin-artweb/internal/mon/biz"
	svReso "gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type NodeService struct {
	log    *zap.Logger
	ucNode *biz.MonNodeUsecase
}

func NewNodeService(
	logger *zap.Logger,
	ucNode *biz.MonNodeUsecase,
) *NodeService {
	return &NodeService{
		log:    logger,
		ucNode: ucNode,
	}
}

// @Summary 创建mon节点
// @Description 本接口用于创建新的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param request body pbNode.CreateOrUpdateMonNodeRequest true "创建mon节点请求"
// @Success 200 {object} pbNode.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [post]
// @Security ApiKeyAuth
func (s *NodeService) CreateMonNode(ctx *gin.Context) {
	var req pbNode.CreateOrUpdateMonNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建mon节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	node := biz.MonNodeModel{
		Name:        req.Name,
		DeployPath:  req.DeployPath,
		OutportPath: req.OutportPath,
		JavaHome:    req.JavaHome,
		URL:         req.URL,
		HostID:      req.HostID,
	}

	m, rErr := s.ucNode.CreateMonNode(ctx, node)
	if rErr != nil {
		s.log.Error(
			"创建mon节点失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.MonNodeReply{
		Code: http.StatusOK,
		Data: *MonNodeToDetailOut(*m),
	})
}

// @Summary 更新mon节点
// @Description 本接口用于更新指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Param request body pbNode.CreateOrUpdateMonNodeRequest true "更新mon节点请求"
// @Success 200 {object} pbNode.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [put]
// @Security ApiKeyAuth
func (s *NodeService) UpdateMonNode(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新mon节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbNode.CreateOrUpdateMonNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新mon节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	data := map[string]any{
		"name":         req.Name,
		"deploy_path":  req.DeployPath,
		"outport_path": req.OutportPath,
		"java_home":    req.JavaHome,
		"url":          req.URL,
		"host_id":      req.HostID,
	}

	m, err := s.ucNode.UpdateMonNodeByID(ctx, uri.ID, data)
	if err != nil {
		s.log.Error(
			"更新mon节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.MonNodeReply{
		Code: http.StatusOK,
		Data: *MonNodeToDetailOut(*m),
	})
}

// @Summary 删除mon节点
// @Description 本接口用于删除指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [delete]
// @Security ApiKeyAuth
func (s *NodeService) DeleteMonNode(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除mon节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除mon节点",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucNode.DeleteMonNodeByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除mon节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除mon节点成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询mon节点详情
// @Description 本接口用于查询指定ID的mon节点详情
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Success 200 {object} pbNode.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [get]
// @Security ApiKeyAuth
func (s *NodeService) GetMonNode(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询mon节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询mon节点详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucNode.FindMonNodeByID(ctx, []string{"Host"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询mon节点详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询mon节点详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := MonNodeToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbNode.MonNodeReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mon节点列表
// @Description 本接口用于查询mon节点列表
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param request query pbNode.ListMonNodeRequest false "查询参数"
// @Success 200 {object} pbNode.PagMonNodeReply "成功返回mon节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [get]
// @Security ApiKeyAuth
func (s *NodeService) ListMonNode(ctx *gin.Context) {
	var req pbNode.ListMonNodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询mon节点列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询mon节点列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Host"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucNode.ListMonNode(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mon节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询mon节点列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := ListMonNodeToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbNode.PagMonNodeReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *NodeService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateMonNode)
	r.PUT("/node/:id", s.UpdateMonNode)
	r.DELETE("/node/:id", s.DeleteMonNode)
	r.GET("/node/:id", s.GetMonNode)
	r.GET("/node", s.ListMonNode)
}

func MonNodeToBaseOut(
	m biz.MonNodeModel,
) *pbNode.MonNodeBaseOut {
	return &pbNode.MonNodeBaseOut{
		ID:          m.ID,
		Name:        m.Name,
		DeployPath:  m.DeployPath,
		OutportPath: m.OutportPath,
		JavaHome:    m.JavaHome,
		URL:         m.URL,
	}
}

func MonNodeToStandardOut(
	m biz.MonNodeModel,
) *pbNode.MonNodeStandardOut {
	return &pbNode.MonNodeStandardOut{
		MonNodeBaseOut: *MonNodeToBaseOut(m),
		CreatedAt:      m.CreatedAt.Format(time.DateTime),
		UpdatedAt:      m.UpdatedAt.Format(time.DateTime),
	}
}

func MonNodeToDetailOut(
	m biz.MonNodeModel,
) *pbNode.MonNodeDetailOut {
	return &pbNode.MonNodeDetailOut{
		MonNodeStandardOut: *MonNodeToStandardOut(m),
		Host:               svReso.HostModelToBaseOut(m.Host),
	}
}

func ListMonNodeToDetailOut(
	rms *[]biz.MonNodeModel,
) *[]pbNode.MonNodeDetailOut {
	if rms == nil {
		return &[]pbNode.MonNodeDetailOut{}
	}

	ms := *rms
	mso := make([]pbNode.MonNodeDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MonNodeToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
