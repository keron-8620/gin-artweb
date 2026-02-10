package service

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbConf "gin-artweb/api/mds/conf"
	"gin-artweb/internal/business/mds/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
)

type MdsConfService struct {
	log      *zap.Logger
	ucColony *biz.MdsColonyUsecase
	maxSize  int64
}

func NewMdsConfService(
	logger *zap.Logger,
	ucColony *biz.MdsColonyUsecase,
	maxSize int64,
) *MdsConfService {
	return &MdsConfService{
		log:      logger,
		ucColony: ucColony,
		maxSize:  maxSize,
	}
}

// UploadMdsConf 上传mds配置文件
// @Summary 上传mds配置文件
// @Description 上传mds配置文件到指定目录
// @Tags mds配置管理
// @Accept multipart/form-data
// @Produce json
// @Param colony_num path string true "集群编号"
// @Param dir_name path string true "目录名称"
// @Param file formData file true "配置文件"
// @Success 200 {object} pbComm.MapAPIReply "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf/{dir_name} [post]
// @Security ApiKeyAuth
// UploadMdsConf 上传mds配置文件
func (s *MdsConfService) UploadMdsConf(ctx *gin.Context) {
	// 1. 绑定URL路径参数
	var pathReq pbConf.GetMdsConfRequest
	if err := ctx.ShouldBindUri(&pathReq); err != nil {
		s.log.Error(
			"绑定上传的mds配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	// 2. 绑定表单数据（包含文件）
	var formReq pbConf.UploadMdsConfRequest
	if err := ctx.ShouldBind(&formReq); err != nil {
		s.log.Error(
			"绑定上传的mds配置文件表单参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	// 3. 将配置文件保存到指定的位置
	dirName := s.ucColony.GetMdsColonyConfigDir(pathReq.ColonyNum)
	savePath := filepath.Join(dirName, pathReq.DirName, formReq.File.Filename)
	if err := common.UploadFile(ctx, s.log, s.maxSize, savePath, formReq.File, 0o644); err != nil {
		errors.RespondWithError(ctx, err)
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// DownloadMdsConf 下载mds配置文件
// @Summary 下载mds配置文件
// @Description 下载指定的mds配置文件
// @Tags mds配置管理
// @Accept json
// @Produce octet-stream
// @Param colony_num path string true "集群编号"
// @Param dir_name path string true "目录名称"
// @Param filename path string true "文件名"
// @Success 200 "下载成功，返回文件流"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf/{dir_name}/{filename} [get]
// @Security ApiKeyAuth
func (s *MdsConfService) DownloadMdsConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteMdsConfRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		s.log.Error(
			"绑定删除的mds配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	dirName := s.ucColony.GetMdsColonyConfigDir(req.ColonyNum)
	filePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := common.DownloadFile(ctx, s.log, filePath, ""); err != nil {
		errors.RespondWithError(ctx, err)
		return
	}
}

// DeleteMdsConf 删除mds配置文件
// @Summary 删除mds配置文件
// @Description 删除指定的mds配置文件
// @Tags mds配置管理
// @Accept json
// @Produce json
// @Param colony_num path string true "集群编号"
// @Param dir_name path string true "目录名称"
// @Param filename path string true "文件名"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf/{dir_name}/{filename} [delete]
// @Security ApiKeyAuth
func (s *MdsConfService) DeleteMdsConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteMdsConfRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		s.log.Error(
			"绑定删除的mds配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	dirName := s.ucColony.GetMdsColonyConfigDir(req.ColonyNum)
	savePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := fileutil.Remove(ctx, savePath); err != nil {
		s.log.Error(
			"删除mds配置文件失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// ListMdsConf 获取mds配置文件列表
// @Summary 获取mds配置文件列表
// @Description 获取指定目录下的mds配置文件列表
// @Tags mds配置管理
// @Accept json
// @Produce json
// @Param colony_num path string true "集群编号"
// @Success 200 {object} pbConf.PagMdsConfReply "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf [get]
// @Security ApiKeyAuth
func (s *MdsConfService) ListMdsConf(ctx *gin.Context) {
	var req pbConf.ListMdsConfRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		s.log.Error(
			"绑定mds配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	dirName := s.ucColony.GetMdsColonyConfigDir(req.ColonyNum)
	info, err := fileutil.ListFileInfo(ctx, dirName)
	if err != nil {
		s.log.Error(
			"获取mds配置文件列表失败",
			zap.Error(err),
			zap.String("dirname", dirName),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	ctx.JSON(http.StatusOK, pbConf.PagMdsConfReply{
		Code: http.StatusOK,
		Data: info,
	})
}

func (s *MdsConfService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/:colony_num/conf/:dir_name", s.UploadMdsConf)
	r.GET("/:colony_num/conf/:dir_name/:filename", s.DownloadMdsConf)
	r.DELETE("/:colony_num/conf/:dir_name/:filename", s.DeleteMdsConf)
	r.GET("/:colony_num/conf", s.ListMdsConf)
}
