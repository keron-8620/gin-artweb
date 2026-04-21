package oes

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	oesmodel "gin-artweb/internal/model/oes"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
)

type OesConfHandler struct {
	log     *zap.Logger
	maxSize int64
}

func NewOesConfHandler(
	logger *zap.Logger,
	maxSize int64,
) *OesConfHandler {
	return &OesConfHandler{
		log:     logger,
		maxSize: maxSize,
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
// @Success 200 {object} commodel.MapAPIResp "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf/{dir_name} [post]
// @Security ApiKeyAuth
func (s *OesConfHandler) UploadOesConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("上传oes配置文件:开始执行")

	// 1. 绑定URL路径参数
	var pathReq oesmodel.GetOesConfDTO
	if err := c.ShouldBindUri(&pathReq); err != nil {
		log.Error(
			"绑定上传的oes配置文件路径参数失败",
			zap.Error(err),
			zap.String("request_uri", c.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(c, rErr)
		return
	}

	// 2. 绑定表单数据（包含文件）
	var formReq oesmodel.UploadOesConfDto
	if err := c.ShouldBind(&formReq); err != nil {
		log.Error(
			"绑定上传的oes配置文件表单参数失败",
			zap.Error(err),
			zap.String("request_uri", c.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(c, rErr)
		return
	}

	// 3. 将配置文件保存到指定的位置
	dirName := oessvc.GetOesColonyConfigDir(pathReq.ColonyNum)
	savePath := filepath.Join(dirName, pathReq.DirName, formReq.File.Filename)
	if err := common.UploadFile(c, log, s.maxSize, savePath, formReq.File, 0o644); err != nil {
		log.Error(
			"上传oes配置文件:执行失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"上传oes配置文件:执行成功",
		zap.String("colony_num", pathReq.ColonyNum),
		zap.String("dir_name", pathReq.DirName),
		zap.String("filename", formReq.File.Filename),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (s *OesConfHandler) DownloadOesConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("下载oes配置文件:开始执行")

	var req oesmodel.OesConfFileDTO
	if err := c.ShouldBindUri(&req); err != nil {
		log.Error(
			"绑定删除的oes配置文件路径参数失败",
			zap.Error(err),
			zap.String("request_uri", c.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(c, rErr)
		return
	}

	dirName := oessvc.GetOesColonyConfigDir(req.ColonyNum)
	filePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := common.DownloadFile(c, log, filePath, ""); err != nil {
		log.Error(
			"下载oes配置文件:执行失败",
			zap.Error(err),
			zap.String("save_path", filePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"下载oes配置文件:执行成功",
		zap.String("colony_num", req.ColonyNum),
		zap.String("dir_name", req.DirName),
		zap.String("filename", req.Filename),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf/{dir_name}/{filename} [delete]
// @Security ApiKeyAuth
func (s *OesConfHandler) DeleteOesConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("删除oes配置文件:开始执行")

	var req oesmodel.OesConfFileDTO
	if err := c.ShouldBindUri(&req); err != nil {
		log.Error(
			"绑定删除的oes配置文件路径参数失败",
			zap.Error(err),
			zap.String("request_uri", c.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(c, rErr)
		return
	}

	dirName := oessvc.GetOesColonyConfigDir(req.ColonyNum)
	savePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := fileutil.Remove(ctx, savePath); err != nil {
		log.Error(
			"删除oes配置文件失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(c, rErr)
		return
	}
	log.Info(
		"删除oes配置文件:执行成功",
		zap.String("colony_num", req.ColonyNum),
		zap.String("dir_name", req.DirName),
		zap.String("filename", req.Filename),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// ListOesConf 获取oes配置文件列表
// @Summary 获取oes配置文件列表
// @Description 获取指定目录下的oes配置文件列表
// @Tags oes配置管理
// @Accept json
// @Produce json
// @Param colony_num path string true "集群编号"
// @Success 200 {object} oesmodel.PagOesConfResp "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/{colony_num}/conf [get]
// @Security ApiKeyAuth
func (s *OesConfHandler) ListOesConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.ListOesConfDTO
	if err := c.ShouldBindUri(&req); err != nil {
		log.Error(
			"绑定oes配置文件路径参数失败",
			zap.Error(err),
			zap.String("request_uri", c.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(c, rErr)
		return
	}

	dirName := oessvc.GetOesColonyConfigDir(req.ColonyNum)
	info, err := fileutil.ListFileInfo(ctx, dirName)
	if err != nil {
		log.Error(
			"获取oes配置文件列表失败",
			zap.Error(err),
			zap.String("dir_name", dirName),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(c, rErr)
		return
	}

	c.JSON(http.StatusOK, oesmodel.PagOesConfResp{
		Code: http.StatusOK,
		Data: info,
	})
}

func (s *OesConfHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/:colony_num/conf/:dir_name", s.UploadOesConf)
	r.GET("/:colony_num/conf/:dir_name/:filename", s.DownloadOesConf)
	r.DELETE("/:colony_num/conf/:dir_name/:filename", s.DeleteOesConf)
	r.GET("/:colony_num/conf", s.ListOesConf)
}
