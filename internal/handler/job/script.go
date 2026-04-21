package job

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type ScriptHandler struct {
	log       *zap.Logger
	scriptSvc *jobsvc.ScriptService
	maxSize   int64
}

func NewScriptHandler(
	logger *zap.Logger,
	svcScript *jobsvc.ScriptService,
	maxSize int64,
) *ScriptHandler {
	return &ScriptHandler{
		log:       logger,
		scriptSvc: svcScript,
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
// @Success 200 {object} jobmodel.ScriptResp "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 413 {object} errors.Error "文件过大"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script [post]
// @Security ApiKeyAuth
func (h *ScriptHandler) CreateScript(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("上传脚本:开始执行")

	var req jobmodel.UploadScriptDTO
	if !common.ShouldBind(c, log, &req, "上传脚本:绑定上传脚本请求参数失败") {
		return
	}

	fileReader, err := req.File.Open()
	if err != nil {
		log.Error(
			"上传脚本:打开上传文件失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(c, rErr)
		return
	}
	defer fileReader.Close()

	dto := jobmodel.ScriptUpsertDTO{
		Filename: req.File.Filename,
		File:     fileReader,
		Descr:    req.Descr,
		Project:  req.Project,
		Label:    req.Label,
		Language: req.Language,
		Status:   req.Status,
	}

	m, rErr := h.scriptSvc.CreateScript(ctx, dto)
	if rErr != nil {
		log.Error(
			"上传脚本:执行失败",
			zap.Error(rErr),
			zap.Object("upload_script_biz", &dto),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"上传脚本:执行成功",
		zap.Uint32("script_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &jobmodel.ScriptResp{
		Code: http.StatusOK,
		Data: *jobmodel.ScriptModelToStandardOut(*m),
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
// @Success 200 {object} jobmodel.ScriptResp "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 413 {object} errors.Error "文件过大"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [put]
// @Security ApiKeyAuth
func (h *ScriptHandler) UpdateScript(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("更新脚本:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新脚本:绑定脚本ID参数失败") {
		return
	}

	var req jobmodel.UploadScriptDTO
	if !common.ShouldBind(c, log, &req, "更新脚本:绑定更新脚本请求参数失败") {
		return
	}

	fileReader, err := req.File.Open()
	if err != nil {
		log.Error(
			"上传脚本:打开上传文件失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(c, rErr)
		return
	}
	defer fileReader.Close()

	dto := jobmodel.ScriptUpsertDTO{
		Filename: req.File.Filename,
		File:     fileReader,
		Descr:    req.Descr,
		Project:  req.Project,
		Label:    req.Label,
		Language: req.Language,
		Status:   req.Status,
	}

	nm, rErr := h.scriptSvc.UpdateScriptByID(ctx, uri.ID, dto)
	if rErr != nil {
		log.Error(
			"更新脚本:执行失败",
			zap.Error(rErr),
			zap.Uint32("script_id", uri.ID),
			zap.Object("update_script_biz", &dto),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"更新脚本:执行成功",
		zap.Uint32("script_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	c.JSON(http.StatusOK, &jobmodel.ScriptResp{
		Code: http.StatusOK,
		Data: *jobmodel.ScriptModelToStandardOut(*nm),
	})
}

// @Summary 删除脚本
// @Description 本接口用于删除指定ID的脚本
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param id path uint true "脚本编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [delete]
// @Security ApiKeyAuth
func (h *ScriptHandler) DeleteScript(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除脚本:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除脚本:绑定删除脚本ID参数失败") {
		return
	}

	err := h.scriptSvc.DeleteScriptByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除脚本:执行失败",
			zap.Error(err),
			zap.Uint32("script_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除脚本:删除脚本成功",
		zap.Uint32("script_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询脚本详情
// @Description 本接口用于查询指定ID的脚本详情
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param id path uint true "脚本编号"
// @Success 200 {object} jobmodel.ScriptResp "成功返回脚本信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "脚本未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/{id} [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) GetScript(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询脚本:绑定查询脚本ID参数失败") {
		return
	}

	m, err := h.scriptSvc.FindScriptByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"查询脚本:执行失败",
			zap.Error(err),
			zap.Uint32("script_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mo := jobmodel.ScriptModelToStandardOut(*m)
	c.JSON(http.StatusOK, &jobmodel.ScriptResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询脚本列表
// @Description 本接口用于查询脚本列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query jobmodel.ListScriptDTO false "查询参数"
// @Success 200 {object} jobmodel.PagScriptResp "成功返回脚本列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) ListScript(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req jobmodel.ListScriptDTO
	if !common.ShouldBindQuery(c, log, &req, "查询脚本列表:绑定查询参数失败") {
		return
	}

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.scriptSvc.ListScript(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询脚本列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_script_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := jobmodel.ListScriptModelToOutBase(ms)
	c.JSON(http.StatusOK, &jobmodel.PagScriptResp{
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
func (h *ScriptHandler) DownloadScript(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "下载脚本:绑定脚本ID参数失败") {
		return
	}

	m, err := h.scriptSvc.FindScriptByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"下载脚本:查询脚本详情失败",
			zap.Error(err),
			zap.Uint32("script_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	savePath := jobsvc.GetScriptStoragePath(m.Project, m.Label, m.Name, m.IsBuiltin)
	err = common.DownloadFile(c, log, savePath, m.Name)
	if err != nil {
		log.Error(
			"下载脚本:下载脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", uri.ID),
			zap.String("script_name", m.Name),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"下载脚本:下载脚本成功",
		zap.Uint32("script_id", uri.ID),
		zap.String("script_project", m.Project),
		zap.String("script_label", m.Label),
		zap.String("script_name", m.Name),
		zap.Duration("total_duration", time.Since(startTime)),
	)
}

// @Summary 查询项目列表
// @Description 本接口用于查询项目列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query jobmodel.ListScriptDTO false "查询参数"
// @Success 200 {object} jobmodel.ListProjectResp "成功返回项目列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/project [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) ListProject(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req jobmodel.ListScriptDTO
	if !common.ShouldBindQuery(c, log, &req, "查询项目列表:绑定查询参数失败") {
		return
	}

	projects, err := h.scriptSvc.ListProjects(ctx, req)
	if err != nil {
		log.Error(
			"查询项目列表:查询项目列表失败",
			zap.Error(err),
			zap.Object("list_script_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, &jobmodel.ListProjectResp{
		Code: http.StatusOK,
		Data: projects,
	})
}

// @Summary 查询标签列表
// @Description 本接口用于查询标签列表
// @Tags 脚本管理
// @Accept json
// @Produce json
// @Param request query jobmodel.ListScriptDTO false "查询参数"
// @Success 200 {object} jobmodel.ListProjectResp "成功返回标签列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/script/label [get]
// @Security ApiKeyAuth
func (h *ScriptHandler) ListLabel(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req jobmodel.ListScriptDTO
	if !common.ShouldBindQuery(c, log, &req, "查询标签列表:绑定查询参数失败") {
		return
	}

	labels, err := h.scriptSvc.ListLabels(ctx, req)
	if err != nil {
		log.Error(
			"查询标签列表:查询标签列表失败",
			zap.Error(err),
			zap.Object("list_script_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, &jobmodel.ListProjectResp{
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
