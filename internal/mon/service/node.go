package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbNode "gin-artweb/api/mon/node"
	pbHost "gin-artweb/api/resource/host"
	"gin-artweb/internal/mon/biz"
	servReso "gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type NodeService struct {
	log    *zap.Logger
	ucNode *biz.NodeUsecase
}

func NewNodeService(
	logger *zap.Logger,
	ucNode *biz.NodeUsecase,
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
// @Param request body pbNode.CreateMonNodeRequest true "创建mon节点请求"
// @Success 200 {object} pbNode.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [post]
// @Security ApiKeyAuth
func (s *NodeService) CreateMonNode(ctx *gin.Context) {
	var req pbNode.CreateMonNodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.log.Error(
			"绑定创建mon节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
		return
	}

	node := biz.NodeModel{
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
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.MonNodeReply{
		Code: http.StatusOK,
		Data: *MonNodeToOut(*m),
	})
}

// @Summary 更新mon节点
// @Description 本接口用于更新指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "mon节点编号"
// @Param request body pbNode.UpdateMonNodeRequest true "更新mon节点请求"
// @Success 200 {object} pbNode.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{pk} [put]
// @Security ApiKeyAuth
func (s *NodeService) UpdateMonNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新mon节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
		return
	}

	var req pbNode.UpdateMonNodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.log.Error(
			"绑定更新mon节点参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
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

	if err := s.ucNode.UpdateMonNodeByID(ctx, uri.PK, data); err != nil {
		s.log.Error(
			"更新mon节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.Reply())
		return
	}

	m, err := s.ucNode.FindMonNodeByID(ctx, []string{"Host"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询更新后的mon节点信息失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.Reply())
		return
	}

	ctx.JSON(http.StatusOK, &pbNode.MonNodeReply{
		Code: http.StatusOK,
		Data: *MonNodeToOut(*m),
	})
}

// @Summary 删除mon节点
// @Description 本接口用于删除指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "mon节点编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{pk} [delete]
// @Security ApiKeyAuth
func (s *NodeService) DeleteMonNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除mon节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始删除mon节点",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucNode.DeleteMonNodeByID(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除mon节点失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"删除mon节点成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询mon节点详情
// @Description 本接口用于查询指定ID的mon节点详情
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param pk path uint true "mon节点编号"
// @Success 200 {object} pbNode.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{pk} [get]
// @Security ApiKeyAuth
func (s *NodeService) GetMonNode(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询mon节点ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始查询mon节点详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucNode.FindMonNodeByID(ctx, []string{"Host"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询mon节点详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"查询mon节点详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := MonNodeToOut(*m)
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
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "mon节点名称"
// @Param is_enabled query bool false "是否启用"
// @Param username query string false "创建用户名"
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
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始查询mon节点列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
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
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"查询mon节点列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListMonNodeToOut(ms)
	ctx.JSON(http.StatusOK, &pbNode.PagMonNodeReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *NodeService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateMonNode)
	r.PUT("/node/:pk", s.UpdateMonNode)
	r.DELETE("/node/:pk", s.DeleteMonNode)
	r.GET("/node/:pk", s.GetMonNode)
	r.GET("/node", s.ListMonNode)
}

func MonNodeToOutBase(
	m biz.NodeModel,
) *pbNode.MonNodeOutBase {
	return &pbNode.MonNodeOutBase{
		ID:          m.ID,
		CreatedAt:   m.CreatedAt.String(),
		UpdatedAt:   m.UpdatedAt.String(),
		Name:        m.Name,
		DeployPath:  m.DeployPath,
		OutportPath: m.OutportPath,
		JavaHome:    m.JavaHome,
		URL:         m.URL,
	}
}

func MonNodeToOut(
	m biz.NodeModel,
) *pbNode.MonNodeOut {
	var host *pbHost.HostOutBase
	if m.Host.ID != 0 {
		host = servReso.HostModelToOutBase(m.Host)
	}
	return &pbNode.MonNodeOut{
		MonNodeOutBase: *MonNodeToOutBase(m),
		Host:           host,
	}
}

func ListMonNodeToOut(
	rms *[]biz.NodeModel,
) *[]pbNode.MonNodeOut {
	if rms == nil {
		return &[]pbNode.MonNodeOut{}
	}

	ms := *rms
	mso := make([]pbNode.MonNodeOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MonNodeToOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
