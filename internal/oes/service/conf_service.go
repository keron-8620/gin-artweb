package service

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbConf "gin-artweb/api/oes/conf"
	"gin-artweb/internal/oes/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
	"gin-artweb/pkg/fileutil"
)

type OesConfService struct {
	log      *zap.Logger
	ucColony *biz.OesColonyUsecase
	maxSize  int64
}

func NewOesConfService(
	logger *zap.Logger,
	ucColony *biz.OesColonyUsecase,
	maxSize int64,
) *OesConfService {
	return &OesConfService{
		log:      logger,
		ucColony: ucColony,
		maxSize:  maxSize,
	}
}

// UploadOesConf 上传oes配置文件
// @Summary 上传oes配置文件
// @Description 上传oes配置文件到指定目录
// @Tags oes配置管理
// @Accept multipart/form-data
// @Produce json
// @Param colony_num path string true "集群编号"
// @Param dir_name path string true "目录名称"
// @Param file formData file true "配置文件"
// @Success 200 {object} pbComm.MapAPIReply "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf/{dir_name} [post]
// @Security ApiKeyAuth
func (s *OesConfService) UploadOesConf(ctx *gin.Context) {
	// 1. 绑定URL路径参数
	var pathReq pbConf.GetOesConfRequest
	if err := ctx.ShouldBindUri(&pathReq); err != nil {
		s.log.Error(
			"绑定上传的oes配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	// 2. 绑定表单数据（包含文件）
	var formReq pbConf.UploadOesConfRequest
	if err := ctx.ShouldBind(&formReq); err != nil {
		s.log.Error(
			"绑定上传的oes配置文件表单参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	// 3. 将配置文件保存到指定的位置
	dirName := s.ucColony.GetOesColonyConfigDir(pathReq.ColonyNum)
	savePath := filepath.Join(dirName, pathReq.DirName, formReq.File.Filename)
	if err := common.UploadFile(ctx, s.log, s.maxSize, savePath, formReq.File, 0o644); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// DownloadOesConf 下载oes配置文件
// @Summary 下载oes配置文件
// @Description 下载指定的oes配置文件
// @Tags oes配置管理
// @Accept json
// @Produce octet-stream
// @Param colony_num path string true "集群编号"
// @Param dir_name path string true "目录名称"
// @Param filename path string true "文件名"
// @Success 200 "下载成功，返回文件流"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf/{dir_name}/{filename} [get]
// @Security ApiKeyAuth
func (s *OesConfService) DownloadOesConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteOesConfRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		s.log.Error(
			"绑定删除的oes配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	dirName := s.ucColony.GetOesColonyConfigDir(req.ColonyNum)
	filePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := common.DownloadFile(ctx, s.log, filePath, ""); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
	}
}

// DeleteOesConf 删除oes配置文件
// @Summary 删除oes配置文件
// @Description 删除指定的oes配置文件
// @Tags oes配置管理
// @Accept json
// @Produce json
// @Param colony_num path string true "集群编号"
// @Param dir_name path string true "目录名称"
// @Param filename path string true "文件名"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf/{dir_name}/{filename} [delete]
// @Security ApiKeyAuth
func (s *OesConfService) DeleteOesConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteOesConfRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		s.log.Error(
			"绑定删除的oes配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	dirName := s.ucColony.GetOesColonyConfigDir(req.ColonyNum)
	savePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := fileutil.Remove(savePath); err != nil {
		s.log.Error(
			"删除oes配置文件失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// ListOesConf 获取oes配置文件列表
// @Summary 获取oes配置文件列表
// @Description 获取指定目录下的oes配置文件列表
// @Tags oes配置管理
// @Accept json
// @Produce json
// @Param colony_num path string true "集群编号"
// @Success 200 {object} pbConf.PagOesConfReply "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf [get]
// @Security ApiKeyAuth
func (s *OesConfService) ListOesConf(ctx *gin.Context) {
	var req pbConf.ListOesConfRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		s.log.Error(
			"绑定oes配置文件路径参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	dirName := s.ucColony.GetOesColonyConfigDir(req.ColonyNum)
	info, err := fileutil.ListFileInfo(dirName)
	if err != nil {
		s.log.Error(
			"获取oes配置文件列表失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
	}
	ctx.JSON(http.StatusOK, pbConf.PagOesConfReply{
		Code: http.StatusOK,
		Data: info,
	})
}

func (s *OesConfService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/:colony_num/conf/:dir_name", s.UploadOesConf)
	r.GET("/:colony_num/conf/:dir_name/:filename", s.DownloadOesConf)
	r.DELETE("/:colony_num/conf/:dir_name/:filename", s.DeleteOesConf)
	r.GET("/:colony_num/conf", s.ListOesConf)
}
