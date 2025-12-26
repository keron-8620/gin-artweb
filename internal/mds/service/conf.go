package service

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbConf "gin-artweb/api/mds/conf"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
)

type MdsConfService struct {
	log     *zap.Logger
	maxSize int64
}

func NewMdsConfService(
	logger *zap.Logger,
	maxSize int64,
) *MdsConfService {
	return &MdsConfService{
		log:     logger,
		maxSize: maxSize,
	}
}

// UploadMdsConf 上传MDS配置文件
// @Summary 上传MDS配置文件
// @Description 上传MDS配置文件到指定目录
// @Tags MDS配置管理
// @Accept multipart/form-data
// @Produce json
// @Param mds_colony_id formData uint32 true "MDS集群ID"
// @Param dir_name formData string true "目录名称"
// @Param file formData file true "配置文件"
// @Success 200 {object} pbComm.MapAPIReply "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/conf/upload [post]
// @Security ApiKeyAuth
func (s *MdsConfService) UploadMdsConf(ctx *gin.Context) {
	var req pbConf.UploadMdsConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定上传的mds配置文件参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	savePath := s.getMdsConfPath(req.MdsColonyID, req.DirName, req.File.Filename)
	if err := common.UploadFile(ctx, s.log, s.maxSize, savePath, req.File, 0o644); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// DownloadMdsConf 下载MDS配置文件
// @Summary 下载MDS配置文件
// @Description 下载指定的MDS配置文件
// @Tags MDS配置管理
// @Accept json
// @Produce octet-stream
// @Param mds_colony_id query uint32 true "MDS集群ID"
// @Param dir_name query string true "目录名称"
// @Param filename query string true "文件名"
// @Success 200 "下载成功，返回文件流"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/conf/download [get]
// @Security ApiKeyAuth
func (s *MdsConfService) DownloadMdsConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteMdsConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定删除的mds配置文件参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	savePath := s.getMdsConfPath(req.MdsColonyID, req.DirName, req.Filename)
	if err := common.DownloadFile(ctx, s.log, savePath, ""); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
	}
}

// DeleteMdsConf 删除MDS配置文件
// @Summary 删除MDS配置文件
// @Description 删除指定的MDS配置文件
// @Tags MDS配置管理
// @Accept json
// @Produce json
// @Param mds_colony_id query uint32 true "MDS集群ID"
// @Param dir_name query string true "目录名称"
// @Param filename query string true "文件名"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/conf/delete [delete]
// @Security ApiKeyAuth
func (s *MdsConfService) DeleteMdsConf(ctx *gin.Context) {
	var req pbConf.DownloadOrDeleteMdsConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定删除的mds配置文件参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	savePath := s.getMdsConfPath(req.MdsColonyID, req.DirName, req.Filename)
	if err := fileutil.Remove(savePath); err != nil {
		s.log.Error(
			"删除mds配置文件失败",
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

// ListMdsConf 获取MDS配置文件列表
// @Summary 获取MDS配置文件列表
// @Description 获取指定目录下的MDS配置文件列表
// @Tags MDS配置管理
// @Accept json
// @Produce json
// @Param request body pbConf.ListMdsConfRequest true "请求参数"
// @Success 200 {object} pbConf.PagMdsConfReply "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/conf/list [get]
// @Security ApiKeyAuth
func (s *MdsConfService) ListMdsConf(ctx *gin.Context) {
	var req pbConf.ListMdsConfRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定mds配置文件查询参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}
	colonyName := strconv.FormatUint(uint64(req.MdsColonyID), 10)
	dirName := filepath.Join(config.StorageDir, "mds", "config", colonyName, req.DirName)
	info, err := fileutil.ListFileInfo(dirName)
	if err != nil {
		s.log.Error(
			"获取mds配置文件列表失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
	}
	ctx.JSON(http.StatusOK, pbConf.PagMdsConfReply{
		Code: http.StatusOK,
		Data: info,
	})
}

func (s *MdsConfService) getMdsConfPath(colonyID uint32, dirname, filename string) string {
	colonyName := strconv.FormatUint(uint64(colonyID), 10)
	return filepath.Join(config.StorageDir, "mds", "config", colonyName, dirname, filename)
}

func (s *MdsConfService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/conf", s.UploadMdsConf)
	r.PUT("/conf", s.DownloadMdsConf)
	r.DELETE("/conf", s.DeleteMdsConf)
	r.GET("/conf", s.ListMdsConf)
}
