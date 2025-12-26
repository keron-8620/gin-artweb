package service

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbConf "gin-artweb/api/oes/conf"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
)

type OesConfService struct {
	log     *zap.Logger
	maxSize int64
}

func NewOesConfService(
	logger *zap.Logger,
	maxSize int64,
) *OesConfService {
	return &OesConfService{
		log:     logger,
		maxSize: maxSize,
	}
}

// UploadOesConf 上传OES配置文件
// @Summary 上传OES配置文件
// @Description 上传OES配置文件到指定目录
// @Tags OES配置管理
// @Accept multipart/form-data
// @Produce json
// @Param oes_colony_id formData uint32 true "OES集群ID"
// @Param dir_name formData string true "目录名称"
// @Param file formData file true "配置文件"
// @Success 200 {object} pbComm.MapAPIReply "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/conf/upload [post]
// @Security ApiKeyAuth
func (s *OesConfService) UploadOesConf(ctx *gin.Context) {
	var req pbConf.UploadOesConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定上传的oes配置文件参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	savePath := s.getOesConfPath(req.OesColonyID, req.DirName, req.File.Filename)
	if err := common.UploadFile(ctx, s.log, s.maxSize, savePath, req.File, 0o644); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// DownloadOesConf 下载OES配置文件
// @Summary 下载OES配置文件
// @Description 下载指定的OES配置文件
// @Tags OES配置管理
// @Accept json
// @Produce octet-stream
// @Param oes_colony_id query uint32 true "OES集群ID"
// @Param dir_name query string true "目录名称"
// @Param filename query string true "文件名"
// @Success 200 "下载成功，返回文件流"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/conf/download [get]
// @Security ApiKeyAuth
func (s *OesConfService) DownloadOesConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteOesConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定删除的oes配置文件参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	savePath := s.getOesConfPath(req.OesColonyID, req.DirName, req.Filename)
	if err := common.DownloadFile(ctx, s.log, savePath, ""); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
	}
}

// DeleteOesConf 删除OES配置文件
// @Summary 删除OES配置文件
// @Description 删除指定的OES配置文件
// @Tags OES配置管理
// @Accept json
// @Produce json
// @Param oes_colony_id query uint32 true "OES集群ID"
// @Param dir_name query string true "目录名称"
// @Param filename query string true "文件名"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/conf/delete [delete]
// @Security ApiKeyAuth
func (s *OesConfService) DeleteOesConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteOesConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定删除的oes配置文件参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	savePath := s.getOesConfPath(req.OesColonyID, req.DirName, req.Filename)
	if err := fileutil.Remove(savePath); err != nil {
		s.log.Error(
			"删除oes配置文件失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// ListOesConf 获取OES配置文件列表
// @Summary 获取OES配置文件列表
// @Description 获取指定目录下的OES配置文件列表
// @Tags OES配置管理
// @Accept json
// @Produce json
// @Param request body pbConf.ListOesConfRequest true "请求参数"
// @Success 200 {object} pbConf.PagOesConfReply "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/conf/list [get]
// @Security ApiKeyAuth
func (s *OesConfService) ListOesConf(ctx *gin.Context) {
	var req pbConf.ListOesConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定oes配置文件查询参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}
	colonyName := strconv.FormatUint(uint64(req.OesColonyID), 10)
	dirName := filepath.Join(config.StorageDir, "oes", "config", colonyName, req.DirName)
	info, err := fileutil.ListFileInfo(dirName)
	if err != nil {
		s.log.Error(
			"获取oes配置文件列表失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
	}
	ctx.JSON(http.StatusOK, pbConf.PagOesConfReply{
		Code: http.StatusOK,
		Data: info,
	})
}

func (s *OesConfService) getOesConfPath(colonyID uint32, dirname, filename string) string {
	colonyName := strconv.FormatUint(uint64(colonyID), 10)
	return filepath.Join(config.StorageDir, "oes", "config", colonyName, dirname, filename)
}

func (s *OesConfService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/conf", s.UploadOesConf)
	r.PUT("/conf", s.DownloadOesConf)
	r.DELETE("/conf", s.DeleteOesConf)
	r.GET("/conf", s.ListOesConf)
}
