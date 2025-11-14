package service

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbPkg "gin-artweb/api/resource/package"
	"gin-artweb/internal/resource/biz"
	"gin-artweb/pkg/errors"
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

func (s *PackageService) UploadPackage(ctx *gin.Context) {
	label := ctx.PostForm("label")
	version := ctx.PostForm("version")

	// 校验必要参数
	if label == "" || version == "" {
		ctx.JSON(errors.ValidateError.Code, errors.ValidateError.Reply())
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		s.log.Error("从表单中获取上传的程序包文件失败", zap.Error(err))
		ctx.JSON(errors.ErrNoUploadedFileFound.Code, errors.ErrNoUploadedFileFound.WithCause(err).Reply())
		return
	}

	if file.Size > s.maxSize {
		s.log.Error(
			"上传的程序包文件过大",
			zap.Int64("file_size", file.Size),
			zap.Int64("max_size", s.maxSize),
		)
		ctx.JSON(
			errors.ErrFileTooLarge.Code,
			errors.ErrFileTooLarge.WithData(map[string]any{
				"file_size": file.Size,
				"max_size":  s.maxSize,
			}).Reply(),
		)
		return
	}

	// 用 UUID 保证文件名唯一防止并发冲突
	uuidFilename := uuid.NewString() + filepath.Ext(file.Filename)
	savePath := s.ucPkg.PackagePath(uuidFilename)

	defer func() {
		// 确保无论后续流程如何退出都尝试移除临时文件
		if _, statErr := os.Stat(savePath); statErr == nil {
			if rmErr := os.Remove(savePath); rmErr != nil {
				s.log.Warn("删除上传文件失败", zap.String("path", savePath), zap.Error(rmErr))
			}
		}
	}()

	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		s.log.Error("保存上传程序包失败", zap.Error(err))
		rErr := errors.FromError(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	pkg, rErr := s.ucPkg.CreatePackage(ctx, biz.PackageModel{
		Label:   label,
		Version: version,
	})

	if rErr != nil {
		s.log.Error("创建 Package 记录失败", zap.Error(rErr))
		ctx.JSON(rErr.Code, rErr.Reply())
		return // defer 自动清理文件
	}

	ctx.JSON(http.StatusOK, &pbPkg.PackageReply{
		Code: http.StatusOK,
		Data: *PackageModelToOutBase(*pkg),
	})
}

func (s *PackageService) DeletePackage(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucPkg.DeletePackageById(ctx, uri.PK); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

func (s *PackageService) GetPackage(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucPkg.FindPackageById(ctx, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PackageModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pbPkg.PackageReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

func (s *PackageService) ListPackage(ctx *gin.Context) {
	var req pbPkg.ListPackageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucPkg.ListPackage(ctx, page, size, query, []string{"id"}, true)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListPkgModelToOut(ms)
	ctx.JSON(http.StatusOK, &pbPkg.PagPackageReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *PackageService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/package", s.UploadPackage)
	r.DELETE("/package/:pk", s.DeletePackage)
	r.GET("/package/:pk", s.GetPackage)
	r.GET("/package", s.ListPackage)
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
	ms []biz.PackageModel,
) []*pbPkg.PackageOutBase {
	mso := make([]*pbPkg.PackageOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := PackageModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
