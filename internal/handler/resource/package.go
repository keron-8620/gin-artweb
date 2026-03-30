package resource

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	resomodel "gin-artweb/internal/model/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// PackageHandler 处理程序包相关的请求
// 包含日志记录和程序包服务的引用
type PackageHandler struct {
	log        *zap.Logger             // 日志记录器
	packageSvc *resosvc.PackageService // 程序包服务
	maxSize    int64                   // 最大文件大小
}

func NewPackageHandler(
	logger *zap.Logger,
	svcPackage *resosvc.PackageService,
	maxSize int64,
) *PackageHandler {
	return &PackageHandler{
		log:        logger,
		packageSvc: svcPackage,
		maxSize:    maxSize,
	}
}

// @Summary 上传程序包
// @Description 上传一个新的程序包文件并创建记录
// @Tags 程序包管理
// @Accept multipart/form-data
// @Produce json
// @Param label formData string true "程序包标签，长度限制：1-50个字符"
// @Param version formData string true "程序包版本，长度限制：1-50个字符"
// @Param file formData file true "程序包文件"
// @Success 201 {object} resomodel.PackageResp "成功返回程序包信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/package [post]
// @Security ApiKeyAuth
func (h *PackageHandler) UploadPackage(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req resomodel.UploadPackageDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"上传程序包：绑定上传程序包参数失败") {
		return
	}

	log.Info("上传程序包：开始执行")

	log.Debug(
		"上传程序包：入参详情",
		zap.Object("upload_package_dto", &req),
	)

	fileReader, err := req.File.Open()
	if err != nil {
		log.Error(
			"上传程序包：打开上传文件失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	defer fileReader.Close()

	dto := resomodel.UploadPackageBiz{
		Filename: req.File.Filename,
		File:     fileReader,
		Label:    req.Label,
		Version:  req.Version,
	}

	createStepStart := time.Now()
	pkg, rErr := h.packageSvc.CreatePackage(ctx, dto)
	createStepDuration := time.Since(createStepStart)
	if rErr != nil {
		log.Error(
			"上传程序包：执行失败",
			zap.Error(rErr),
			zap.Object("upload_package_biz", &dto),
			zap.Duration("create_step_duration", createStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	log.Debug(
		"上传程序包：创建程序包模型详情",
		zap.Object("package_model", pkg),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"上传程序包：上传程序包成功",
		zap.Uint32("package_id", pkg.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusCreated, &resomodel.PackageResp{
		Code: http.StatusCreated,
		Data: *resomodel.PackageModelToOutBase(*pkg),
	})
}

// @Summary 删除程序包
// @Description 本接口用于删除指定ID的程序包
// @Tags 程序包管理
// @Produce json
// @Param id path uint true "程序包编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "程序包未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/package/{id} [delete]
// @Security ApiKeyAuth
func (h *PackageHandler) DeletePackage(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"删除程序包：绑定删除程序包ID参数失败") {
		return
	}

	log.Info(
		"删除程序包：开始执行",
		zap.Uint32("package_id", uri.ID),
	)

	deleteStepStart := time.Now()
	err := h.packageSvc.DeletePackageByID(ctx, uri.ID)
	deleteStepDuration := time.Since(deleteStepStart)
	if err != nil {
		log.Error(
			"删除程序包：执行失败",
			zap.Error(err),
			zap.Uint32("package_id", uri.ID),
			zap.Duration("delete_step_duration", deleteStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"删除程序包：删除程序包成功",
		zap.Uint32("package_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询程序包
// @Description 本接口用于查询指定ID的程序包详细信息
// @Tags 程序包管理
// @Produce json
// @Param id path uint true "程序包编号"
// @Success 200 {object} resomodel.PackageResp "成功返回程序包详情"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "程序包未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/package/{id} [get]
// @Security ApiKeyAuth
func (h *PackageHandler) GetPackage(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"查询程序包：绑定查询程序包ID参数失败") {
		return
	}

	log.Info(
		"查询程序包：开始执行",
		zap.Uint32("package_id", uri.ID),
	)

	findStepStart := time.Now()
	m, err := h.packageSvc.FindPackageByID(ctx, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询程序包：执行失败",
			zap.Error(err),
			zap.Uint32("package_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询程序包：查询程序包详情",
		zap.Object("package_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"查询程序包：执行成功",
		zap.Uint32("package_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := resomodel.PackageModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &resomodel.PackageResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询程序包列表
// @Description 本接口用于查询程序包列表
// @Tags 程序包管理
// @Produce json
// @Param request query resomodel.ListPackageDTO false "查询参数"
// @Success 200 {object} resomodel.PagPackageResp "成功返回程序包列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/package [get]
// @Security ApiKeyAuth
func (h *PackageHandler) ListPackage(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req resomodel.ListPackageDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"查询程序包列表：绑定参数失败") {
		return
	}

	log.Info("查询程序包列表：开始执行")

	log.Debug(
		"查询程序包列表：入参详情",
		zap.Object("list_package_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.BaseModelQuery.GetPageParam()
	total, ms, err := h.packageSvc.ListPackage(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询程序包列表：执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_package_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询程序包列表：执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := resomodel.ListPkgModelToOut(ms)
	ctx.JSON(http.StatusOK, &resomodel.PagPackageResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 下载程序包
// @Description 本接口用于下载指定ID的程序包文件
// @Tags 程序包管理
// @Produce application/octet-stream
// @Param id path uint true "程序包编号"
// @Success 200 {file} file "成功下载程序包文件"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "文件未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/package/{id}/download [get]
// @Security ApiKeyAuth
func (h *PackageHandler) DownloadPackage(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"下载程序包：绑定下载程序包ID参数失败") {
		return
	}

	log.Info(
		"下载程序包：开始执行",
		zap.Uint32("package_id", uri.ID),
	)

	// 获取包信息
	findStepStart := time.Now()
	pkg, err := h.packageSvc.FindPackageByID(ctx, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"下载程序包：查询程序包详情失败",
			zap.Error(err),
			zap.Uint32("package_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"下载程序包：查询程序包详情",
		zap.Object("package_model", pkg),
		zap.Duration("find_step_duration", findStepDuration),
	)

	// 构建文件路径
	filePath := resosvc.GetPackageStoragePath(pkg.StorageFilename)
	if err := common.DownloadFile(ctx, log, filePath, pkg.OriginFilename); err != nil {
		log.Error(
			"下载程序包：下载文件失败",
			zap.Error(err),
			zap.Uint32("package_id", uri.ID),
			zap.String("filename", pkg.OriginFilename),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"下载程序包：执行成功",
		zap.Uint32("package_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
}

func (h *PackageHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/package", h.UploadPackage)
	r.DELETE("/package/:id", h.DeletePackage)
	r.GET("/package/:id", h.GetPackage)
	r.GET("/package", h.ListPackage)
	r.GET("/package/:id/download", h.DownloadPackage)
}
