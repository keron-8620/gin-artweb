package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobsmodel "gin-artweb/internal/model/jobs"
	jobsvc "gin-artweb/internal/service/jobs"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScriptHandler struct {
	log       *zap.Logger
	svcScript *jobsvc.ScriptService
	maxSize   int64
}

func NewScriptHandler(
	logger *zap.Logger,
	svcScript *jobsvc.ScriptService,
	maxSize int64,
) *ScriptHandler {
	return &ScriptHandler{
		log:       logger,
		svcScript: svcScript,
		maxSize:   maxSize,
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
// @Success 200 {object} jobsmodel.ScriptReply "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 413 {object} errors.Error "文件过大"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script [post]
// @Security ApiKeyAuth
func (h *ScriptHandler) CreateScript(ctx *gin.Context) {
	var req jobsmodel.UploadScriptRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定上传脚本参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	script := jobsmodel.ScriptModel{
		Name:      req.File.Filename,
		Descr:     req.Descr,
		Project:   req.Project,
		Label:     req.Label,
		Language:  req.Language,
		Status:    req.Status,
		IsBuiltin: false,
		Username:  claims.Subject,
	}

	savePath := common.GetScriptStoragePath(script.Project, script.Label, script.Name, script.IsBuiltin)
	if err := common.UploadFile(ctx, h.log, h.maxSize, savePath, req.File, 0o755); err != nil {
		errors.RespondWithError(ctx, err)
		return
	}

	m, rErr := h.svcScript.CreateScript(ctx, script)
	if rErr != nil {
		h.log.Error(
			"创建脚本失败",
			zap.Error(rErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &jobsmodel.ScriptReply{
		Code: http.StatusOK,
		Data: *jobsmodel.ScriptModelToStandardOut(*m),
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
// @Success 200 {object} jobsmodel.ScriptReply "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 413 {object} errors.Error "文件过大"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [put]
// @Security ApiKeyAuth
func (h *ScriptHandler) UpdateScript(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定更新脚本ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req jobsmodel.UploadScriptRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定上传脚本参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	om, rErr := h.svcScript.FindScriptByID(ctx, uri.ID)
	if rErr != nil {
		h.log.Error(
			"查询脚本失败",
			zap.Error(rErr),
			zap.Uint32("script_id", uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if rErr := h.svcScript.RemoveScript(ctx, *om); rErr != nil {
		h.log.Error(
			"删除原脚本文件失败",
			zap.Error(rErr),
			zap.Uint32("script_id", uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	nm := jobsmodel.ScriptModel{
		Name:      req.File.Filename,
		Descr:     req.Descr,
		Project:   req.Project,
		Label:     req.Label,
		Language:  req.Language,
		Status:    req.Status,
		IsBuiltin: false,
		Username:  claims.Subject,
	}
	nm.ID = uri.ID
	savePath := common.GetScriptStoragePath(nm.Project, nm.Label, nm.Name, nm.IsBuiltin)
	if err := common.UploadFile(ctx, h.log, h.maxSize, savePath, req.File, 0o755); err != nil {
		errors.RespondWithError(ctx, err)
		return
	}

	m, rErr := h.svcScript.UpdateScriptByID(ctx, uri.ID, map[string]any{
		"name":       req.File.Filename,
		"descr":      req.Descr,
		"project":    req.Project,
		"label":      req.Label,
		"language":   req.Language,
		"status":     req.Status,
		"is_builtin": false,
		"username":   claims.Subject,
	})
	if rErr != nil {
		h.log.Error(
			"更新脚本失败",
			zap.Error(rErr),
			zap.Uint32("script_id", uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &jobsmodel.ScriptReply{
		Code: http.StatusOK,
		Data: *jobsmodel.ScriptModelToStandardOut(*m),
	})
}

// @Summary 删除脚本
// @Description 本接口用于删除指定ID的脚本
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param id path uint true "脚本编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [delete]
// @Security ApiKeyAuth
func (h *ScriptHandler) DeleteScript(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除脚本ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除脚本",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcScript.DeleteScriptByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除脚本失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除脚本成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询脚本详情
// @Description 本接口用于查询指定ID的脚本详情
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param id path uint true "脚本编号"
// @Success 200 {object} jobsmodel.ScriptReply "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) GetScript(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询脚本ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询脚本详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcScript.FindScriptByID(ctx, uri.ID)
	if err != nil {
		h.log.Error(
			"查询脚本详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询脚本详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := jobsmodel.ScriptModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &jobsmodel.ScriptReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询脚本列表
// @Description 本接口用于查询脚本列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query jobsmodel.ListScriptRequest false "查询参数"
// @Success 200 {object} jobsmodel.PagScriptReply "成功返回脚本列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) ListScript(ctx *gin.Context) {
	var req jobsmodel.ListScriptRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询脚本列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询脚本列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Size:    size,
		Page:    page,
		OrderBy: []string{"id DESC"},
		Query:   query,
	}
	total, ms, err := h.svcScript.ListScript(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询脚本列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询脚本列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := jobsmodel.ListScriptModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &jobsmodel.PagScriptReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
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
func (h *ScriptHandler) DownloadScript(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定下载脚本ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, err := h.svcScript.FindScriptByID(ctx, uri.ID)
	if err != nil {
		h.log.Error(
			"查询脚本详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	savePath := common.GetScriptStoragePath(m.Project, m.Label, m.Name, m.IsBuiltin)
	if err := common.DownloadFile(ctx, h.log, savePath, m.Name); err != nil {
		errors.RespondWithError(ctx, err)
	}
}

// @Summary 查询项目列表
// @Description 本接口用于查询项目列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query jobsmodel.ListScriptRequest false "查询参数"
// @Success 200 {object} jobsmodel.ListProjectReply "成功返回项目列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/project [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) ListProject(ctx *gin.Context) {
	var req jobsmodel.ListScriptRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询脚本列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询项目列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	_, _, query := req.Query()

	projects, err := h.svcScript.ListProjects(ctx, query)
	if err != nil {
		h.log.Error(
			"查询项目列表失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询项目列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &jobsmodel.ListProjectReply{
		Code: http.StatusOK,
		Data: projects,
	})
}

// @Summary 查询标签列表
// @Description 本接口用于查询标签列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query jobsmodel.ListScriptRequest false "查询参数"
// @Success 200 {object} jobsmodel.ListLableReply "成功返回标签列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/label [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) ListLabel(ctx *gin.Context) {
	var req jobsmodel.ListScriptRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询脚本列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询标签列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	_, _, query := req.Query()

	labels, err := h.svcScript.ListLabels(ctx, query)
	if err != nil {
		h.log.Error(
			"查询标签列表失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询项目列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &jobsmodel.ListProjectReply{
		Code: http.StatusOK,
		Data: labels,
	})
}

func (h *ScriptHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/script", h.CreateScript)
	r.PUT("/script/:id", h.UpdateScript)
	r.DELETE("/script/:id", h.DeleteScript)
	r.GET("/script/:id", h.GetScript)
	r.GET("/script", h.ListScript)
	r.GET("/script/:id/download", h.DownloadScript)
	r.GET("/script/project", h.ListProject)
	r.GET("/script/label", h.ListLabel)
}
