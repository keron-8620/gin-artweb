package resource

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	resomodel "gin-artweb/internal/model/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type PackageHandler struct {
	log        *zap.Logger
	svcPackage *resosvc.PackageService
	maxSize    int64
}

func NewPackageHandler(
	logger *zap.Logger,
	svcPackage *resosvc.PackageService,
	maxSize int64,
) *PackageHandler {
	return &PackageHandler{
		log:        logger,
		svcPackage: svcPackage,
		maxSize:    maxSize,
	}
}

// @Summary      上传程序包
// @Description  上传一个新的程序包文件并创建记录
// @Tags         程序包管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        label formData string true "程序包标签，长度限制：1-50个字符"
// @Param        version formData string true "程序包版本，长度限制：1-50个字符"
// @Param        file formData file true "程序包文件"
// @Success      201  {object} resomodel.PackageReply "成功返回程序包信息"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package [post]
// @Security ApiKeyAuth
func (h *PackageHandler) UploadPackage(ctx *gin.Context) {
	var req resomodel.UploadPackageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定上传程序包参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	newFileNameWithExt := uuid.NewString() + filepath.Ext(req.File.Filename)
	savePath := common.GetPackageStoragePath(newFileNameWithExt)
	if rErr := common.UploadFile(ctx, h.log, h.maxSize, savePath, req.File, 0o644); rErr != nil {
		errors.RespondWithError(ctx, rErr)
		return
	}

	pkg, rErr := h.svcPackage.CreatePackage(ctx, resomodel.PackageModel{
		Label:           req.Label,
		Version:         req.Version,
		StorageFilename: newFileNameWithExt,
		OriginFilename:  req.File.Filename,
	})
	if rErr != nil {
		h.log.Error(
			"创建 Package 记录失败",
			zap.Error(rErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &resomodel.PackageReply{
		Code: http.StatusOK,
		Data: *resomodel.PackageModelToOutBase(*pkg),
	})
}

// @Summary      删除程序包
// @Description  本接口用于删除指定ID的程序包
// @Tags         程序包管理
// @Produce      json
// @Param        id path uint32 true "程序包唯一标识符"
// @Success      200  {object} commodel.MapAPIReply "删除成功"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package/{id} [delete]
// @Security ApiKeyAuth
func (h *PackageHandler) DeletePackage(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除程序包ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除程序包",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcPackage.DeletePackageById(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除程序包失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除程序包成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary     查询程序包详情
// @Description  本接口用于查询指定ID的程序包详细信息
// @Tags         程序包管理
// @Produce      json
// @Param        id path uint32 true "程序包唯一标识符"
// @Success      200  {object} resomodel.PackageReply "成功返回程序包详情"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package/{id} [get]
// @Security ApiKeyAuth
func (h *PackageHandler) GetPackage(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询程序包ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询程序包详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcPackage.FindPackageById(ctx, uri.ID)
	if err != nil {
		h.log.Error(
			"查询程序包详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询程序包详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := resomodel.PackageModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &resomodel.PackageReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary      查询程序包列表
// @Description  本接口用于查询程序包列表
// @Tags         程序包管理
// @Produce      json
// @Param        request query resomodel.ListPackageRequest false "查询参数"
// @Success      200  {object} resomodel.PagPackageReply "成功返回程序包列表"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package [get]
// @Security ApiKeyAuth
func (h *PackageHandler) ListPackage(ctx *gin.Context) {
	var req resomodel.ListPackageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询程序包列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询程序包列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Size:    size,
		Page:    page,
		OrderBy: []string{"uploaded_at DESC"},
		Query:   query,
	}
	total, ms, err := h.svcPackage.ListPackage(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询程序包列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询程序包列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := resomodel.ListPkgModelToOut(ms)
	ctx.JSON(http.StatusOK, &resomodel.PagPackageReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary      下载程序包
// @Description  本接口用于下载指定ID的程序包文件
// @Tags         程序包管理
// @Produce      application/octet-stream
// @Param        id path uint32 true "程序包唯一标识符"
// @Success      200  {file} file "成功下载程序包文件"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      404  {object} errors.Error "文件未找到"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package/{id}/download [get]
// @Security ApiKeyAuth
func (h *PackageHandler) DownloadPackage(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定下载程序包ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	// 获取包信息
	pkg, err := h.svcPackage.FindPackageById(ctx, uri.ID)
	if err != nil {
		h.log.Error(
			"查询程序包详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	// 构建文件路径
	filePath := common.GetPackageStoragePath(pkg.StorageFilename)
	if err := common.DownloadFile(ctx, h.log, filePath, pkg.OriginFilename); err != nil {
		errors.RespondWithError(ctx, err)
	}
}

func (h *PackageHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/package", h.UploadPackage)
	r.DELETE("/package/:id", h.DeletePackage)
	r.GET("/package/:id", h.GetPackage)
	r.GET("/package", h.ListPackage)
	r.GET("/package/:id/download", h.DownloadPackage)
}
