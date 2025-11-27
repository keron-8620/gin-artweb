package service

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbPkg "gin-artweb/api/resource/pkg"
	"gin-artweb/internal/resource/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type PackageService struct {
	log     *zap.Logger
	ucPkg   *biz.PackageUsecase
	maxSize int64
}

func NewPackageService(
	logger *zap.Logger,
	ucPkg *biz.PackageUsecase,
	maxSize int64,
) *PackageService {
	return &PackageService{
		log:     logger,
		ucPkg:   ucPkg,
		maxSize: maxSize,
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
// @Success      201  {object} pbPkg.PackageReply "成功返回程序包信息"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package [post]
// @Security ApiKeyAuth
func (s *PackageService) UploadPackage(ctx *gin.Context) {
	var req pbPkg.UploadPackageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	if req.File.Size > s.maxSize {
		s.log.Error(
			"上传的程序包文件过大",
			zap.Int64("file_size", req.File.Size),
			zap.Int64("max_size", s.maxSize),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(
			errors.ErrFileTooLarge.Code,
			errors.ErrFileTooLarge.WithData(map[string]any{
				"file_size": req.File.Size,
				"max_size":  s.maxSize,
			}).Reply(),
		)
		return
	}

	// 用 UUID 保证文件名唯一防止并发冲突
	uuidFilename := uuid.NewString() + filepath.Ext(req.File.Filename)
	savePath := s.ucPkg.PackagePath(uuidFilename)
	if err := ctx.SaveUploadedFile(req.File, savePath); err != nil {
		s.log.Error(
			"保存上传程序包失败",
			zap.Error(err),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	pkg, rErr := s.ucPkg.CreatePackage(ctx, biz.PackageModel{
		Label:           req.Label,
		Version:         req.Version,
		StorageFilename: uuidFilename,
		OriginFilename:  req.File.Filename,
	})

	if rErr != nil {
		s.log.Error(
			"创建 Package 记录失败",
			zap.Error(rErr),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	ctx.JSON(http.StatusOK, &pbPkg.PackageReply{
		Code: http.StatusOK,
		Data: *PackageModelToOutBase(*pkg),
	})
}

// @Summary      删除程序包
// @Description  本接口用于删除指定ID的程序包
// @Tags         程序包管理
// @Produce      json
// @Param        pk path uint32 true "程序包唯一标识符"
// @Success      200  {object} pbComm.MapAPIReply "删除成功"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package/{pk} [delete]
// @Security ApiKeyAuth
func (s *PackageService) DeletePackage(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除程序包ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始删除程序包",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucPkg.DeletePackageById(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除程序包失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"删除程序包成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary     查询程序包详情
// @Description  本接口用于查询指定ID的程序包详细信息
// @Tags         程序包管理
// @Produce      json
// @Param        pk path uint32 true "程序包唯一标识符"
// @Success      200  {object} pbPkg.PackageReply "成功返回程序包详情"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package/{pk} [get]
// @Security ApiKeyAuth
func (s *PackageService) GetPackage(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询程序包ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始查询程序包详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucPkg.FindPackageById(ctx, uri.PK)
	if err != nil {
		s.log.Error(
			"查询程序包详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"查询程序包详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := PackageModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pbPkg.PackageReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary      查询程序包列表
// @Description  本接口用于查询程序包列表
// @Tags         程序包管理
// @Produce      json
// @Param        page query int false "页码，默认为1"
// @Param        size query int false "每页数量，默认为10"
// @Param        query query string false "搜索关键字"
// @Success      200  {object} pbPkg.PagPackageReply "成功返回程序包列表"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package [get]
// @Security ApiKeyAuth
func (s *PackageService) ListPackage(ctx *gin.Context) {
	var req pbPkg.ListPackageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询程序包列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始查询程序包列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Limit:   size,
		Offset:  page,
		OrderBy: []string{"uploaded_at DESC"},
		Query:   query,
	}
	total, ms, err := s.ucPkg.ListPackage(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询程序包列表失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"查询程序包列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListPkgModelToOut(ms)
	ctx.JSON(http.StatusOK, &pbPkg.PagPackageReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary      下载程序包
// @Description  本接口用于下载指定ID的程序包文件
// @Tags         程序包管理
// @Produce      application/octet-stream
// @Param        pk path uint32 true "程序包唯一标识符"
// @Success      200  {file} file "成功下载程序包文件"
// @Failure      400  {object} errors.Error "请求参数错误"
// @Failure      404  {object} errors.Error "文件未找到"
// @Failure      500  {object} errors.Error "服务器内部错误"
// @Router       /api/v1/resource/package/{pk}/download [get]
// @Security ApiKeyAuth
func (s *PackageService) DownloadPackage(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定下载程序包ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	// 获取包信息
	pkg, err := s.ucPkg.FindPackageById(ctx, uri.PK)
	if err != nil {
		s.log.Error(
			"查询程序包详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	// 构建文件路径
	filePath := s.ucPkg.PackagePath(pkg.StorageFilename)

	// 检查文件是否存在
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		s.log.Error(
			"文件不存在",
			zap.String("file_path", filePath),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(errors.ErrFileNotFound.Code, errors.ErrFileNotFound.Reply())
		return
	}

	// 设置响应头，触发浏览器下载
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+pkg.OriginFilename)
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 发送文件
	ctx.File(filePath)
}

func (s *PackageService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/package", s.UploadPackage)
	r.DELETE("/package/:pk", s.DeletePackage)
	r.GET("/package/:pk", s.GetPackage)
	r.GET("/package", s.ListPackage)
	r.GET("/package/:pk/download", s.DownloadPackage)
}

func PackageModelToOutBase(
	m biz.PackageModel,
) *pbPkg.PackageOutBase {
	return &pbPkg.PackageOutBase{
		ID:       m.ID,
		Filename: m.OriginFilename,
		Label:    m.Label,
		Version:  m.Version,
	}
}

func ListPkgModelToOut(
	pms *[]biz.PackageModel,
) *[]pbPkg.PackageOutBase {
	if pms == nil {
		return &[]pbPkg.PackageOutBase{}
	}

	ms := *pms
	mso := make([]pbPkg.PackageOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := PackageModelToOutBase(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
