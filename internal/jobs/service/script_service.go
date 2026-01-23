package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbScript "gin-artweb/api/jobs/script"
	"gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type ScriptService struct {
	log      *zap.Logger
	ucScript *biz.ScriptUsecase
	maxSize  int64
}

func NewScriptService(
	logger *zap.Logger,
	ucScript *biz.ScriptUsecase,
	maxSize int64,
) *ScriptService {
	return &ScriptService{
		log:      logger,
		ucScript: ucScript,
		maxSize:  maxSize,
	}
}

// @Summary 上传脚本
// @Description 本接口用于上传新的脚本文件
// @Tags 脚本管理
// @Accept mpfd
// @Produce json
// @Param file formData file true "脚本文件"
// @Param descr formData string false "脚本描述"
// @Param project formData string true "项目名称"
// @Param label formData string false "标签"
// @Param language formData string true "脚本语言"
// @Param status formData bool true "脚本状态"
// @Success 200 {object} pbScript.ScriptReply "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 413 {object} errors.Error "文件过大"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script [post]
// @Security ApiKeyAuth
func (s *ScriptService) CreateScript(ctx *gin.Context) {
	var req pbScript.UploadScriptRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定上传脚本参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	uc := auth.GetUserClaims(ctx)
	script := biz.ScriptModel{
		Name:      req.File.Filename,
		Descr:     req.Descr,
		Project:   req.Project,
		Label:     req.Label,
		Language:  req.Language,
		Status:    req.Status,
		IsBuiltin: false,
		Username:  uc.Subject,
	}

	savePath := script.ScriptPath()
	if err := common.UploadFile(ctx, s.log, s.maxSize, savePath, req.File, 0o755); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	m, rErr := s.ucScript.CreateScript(ctx, script)
	if rErr != nil {
		s.log.Error(
			"创建脚本失败",
			zap.Error(rErr),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.JSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbScript.ScriptReply{
		Code: http.StatusOK,
		Data: *ScriptModelToStandardOut(*m),
	})
}

// @Summary 更新脚本
// @Description 本接口用于更新指定ID的脚本文件
// @Tags 脚本管理
// @Accept mpfd
// @Produce json
// @Param id path uint true "脚本编号"
// @Param file formData file true "脚本文件"
// @Param descr formData string false "脚本描述"
// @Param project formData string true "项目名称"
// @Param label formData string false "标签"
// @Param language formData string true "脚本语言"
// @Param status formData bool true "脚本状态"
// @Success 200 {object} pbScript.ScriptReply "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 413 {object} errors.Error "文件过大"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [put]
// @Security ApiKeyAuth
func (s *ScriptService) UpdateScript(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新脚本ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbScript.UploadScriptRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定上传脚本参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	om, rErr := s.ucScript.FindScriptByID(ctx, uri.ID)
	if rErr != nil {
		s.log.Error(
			"查询脚本失败",
			zap.Error(rErr),
			zap.Uint32(biz.ScriptIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.JSON(rErr.Code, rErr.ToMap())
		return
	}
	if rErr := s.ucScript.RemoveScript(ctx, *om); rErr != nil {
		s.log.Error(
			"删除原脚本文件失败",
			zap.Error(rErr),
			zap.Uint32(biz.ScriptIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.JSON(rErr.Code, rErr.ToMap())
		return
	}

	uc := auth.GetUserClaims(ctx)
	nm := biz.ScriptModel{
		Name:      req.File.Filename,
		Descr:     req.Descr,
		Project:   req.Project,
		Label:     req.Label,
		Language:  req.Language,
		Status:    req.Status,
		IsBuiltin: false,
		Username:  uc.Subject,
	}
	nm.ID = uri.ID
	if err := common.UploadFile(ctx, s.log, s.maxSize, nm.ScriptPath(), req.File, 0o755); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	m, rErr := s.ucScript.UpdateScriptByID(ctx, uri.ID, map[string]any{
		"name":       req.File.Filename,
		"descr":      req.Descr,
		"project":    req.Project,
		"label":      req.Label,
		"language":   req.Language,
		"status":     req.Status,
		"is_builtin": false,
		"username":   uc.Subject,
	})
	if rErr != nil {
		s.log.Error(
			"更新脚本失败",
			zap.Error(rErr),
			zap.Uint32(biz.ScriptIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.JSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbScript.ScriptReply{
		Code: http.StatusOK,
		Data: *ScriptModelToStandardOut(*m),
	})
}

// @Summary 删除脚本
// @Description 本接口用于删除指定ID的脚本
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param id path uint true "脚本编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [delete]
// @Security ApiKeyAuth
func (s *ScriptService) DeleteScript(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除脚本ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除脚本",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucScript.DeleteScriptByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除脚本失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除脚本成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询脚本详情
// @Description 本接口用于查询指定ID的脚本详情
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param id path uint true "脚本编号"
// @Success 200 {object} pbScript.ScriptReply "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [get]
// @Security ApiKeyAuth
func (s *ScriptService) GetScript(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询脚本ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询脚本详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucScript.FindScriptByID(ctx, uri.ID)
	if err != nil {
		s.log.Error(
			"查询脚本详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询脚本详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	mo := ScriptModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbScript.ScriptReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询脚本列表
// @Description 本接口用于查询脚本列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query pbScript.ListScriptRequest false "查询参数"
// @Success 200 {object} pbScript.PagScriptReply "成功返回脚本列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script [get]
// @Security ApiKeyAuth
func (s *ScriptService) ListScript(ctx *gin.Context) {
	var req pbScript.ListScriptRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询脚本列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询脚本列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Limit:   size,
		Offset:  page,
		OrderBy: []string{"id DESC"},
		Query:   query,
	}
	total, ms, err := s.ucScript.ListScript(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询脚本列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询脚本列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	mbs := ListScriptModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pbScript.PagScriptReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary 下载脚本
// @Description 本接口用于下载指定ID的脚本文件
// @Tags 脚本管理
// @Accept json
// @Produce application/octet-stream
// @Param id path uint true "脚本编号"
// @Success 200 {file} file "成功下载脚本文件"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id}/download [get]
// @Security ApiKeyAuth
func (s *ScriptService) DownloadScript(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定下载脚本ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	m, err := s.ucScript.FindScriptByID(ctx, uri.ID)
	if err != nil {
		s.log.Error(
			"查询脚本详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	if err := common.DownloadFile(ctx, s.log, m.ScriptPath(), m.Name); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
	}
}

func (s *ScriptService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/script", s.CreateScript)
	r.PUT("/script/:id", s.UpdateScript)
	r.DELETE("/script/:id", s.DeleteScript)
	r.GET("/script/:id", s.GetScript)
	r.GET("/script", s.ListScript)
	r.GET("/script/:id/download", s.DownloadScript)
}

func ScriptModelToStandardOut(
	m biz.ScriptModel,
) *pbScript.ScriptStandardOut {
	return &pbScript.ScriptStandardOut{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.Format(time.DateTime),
		UpdatedAt: m.UpdatedAt.Format(time.DateTime),
		Name:      m.Name,
		Descr:     m.Descr,
		Project:   m.Project,
		Label:     m.Label,
		Language:  m.Language,
		Status:    m.Status,
		IsBuiltin: m.IsBuiltin,
		Username:  m.Username,
	}
}

func ListScriptModelToOutBase(
	pms *[]biz.ScriptModel,
) *[]pbScript.ScriptStandardOut {
	if pms == nil {
		return &[]pbScript.ScriptStandardOut{}
	}

	ms := *pms
	mso := make([]pbScript.ScriptStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScriptModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
