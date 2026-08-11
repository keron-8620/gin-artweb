package mon

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	monmodel "gin-artweb/internal/model/mon"
	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
)

// MonConfHandler 处理mon节点配置文件相关请求。
type MonConfHandler struct {
	log     *zap.Logger
	maxSize int64
}

func NewMonConfHandler(logger *zap.Logger, maxSize int64) *MonConfHandler {
	return &MonConfHandler{
		log:     logger,
		maxSize: maxSize,
	}
}

// UploadMonConf 上传mon节点配置文件。
// @Summary 上传mon节点配置文件
// @Description 上传mon节点配置文件到指定节点配置目录
// @Tags mon配置管理
// @Accept multipart/form-data
// @Produce json
// @Param id path uint true "mon节点编号"
// @Param file formData file true "配置文件"
// @Success 200 {object} commodel.MapAPIResp "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/conf/{id} [post]
// @Security ApiKeyAuth
func (h *MonConfHandler) UploadMonConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("上传mon配置文件:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "上传mon配置文件:绑定上传的mon配置文件路径参数失败") {
		return
	}

	var formReq monmodel.UploadMonConfDTO
	if !common.ShouldBind(c, log, &formReq, "上传mon配置文件:绑定上传的mon配置文件表单参数失败") {
		return
	}

	dirName := monsvc.GetMonNodeConfDir(uri.ID)
	savePath := filepath.Join(dirName, formReq.File.Filename)
	if err := common.UploadFile(c, log, h.maxSize, savePath, formReq.File, 0o644); err != nil {
		log.Error(
			"上传mon配置文件:执行失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"上传mon配置文件:执行成功",
		zap.Uint32("node_id", uri.ID),
		zap.String("dir_name", dirName),
		zap.String("filename", formReq.File.Filename),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// DownloadMonConf 下载mon节点配置文件。
// @Summary 下载mon节点配置文件
// @Description 下载指定的mon节点配置文件
// @Tags mon配置管理
// @Accept json
// @Produce octet-stream
// @Param id path uint true "mon节点编号"
// @Param filename query string true "文件名"
// @Success 200 "下载成功，返回文件流"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/conf/{id}/download [get]
// @Security ApiKeyAuth
func (h *MonConfHandler) DownloadMonConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("下载mon配置文件:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "下载mon配置文件:绑定下载的mon配置文件路径参数失败") {
		return
	}
	var query monmodel.MonConfFileQueryDTO
	if !common.ShouldBindQuery(c, log, &query, "下载mon配置文件:绑定下载的mon配置文件查询参数失败") {
		return
	}

	dirName := monsvc.GetMonNodeConfDir(uri.ID)
	filePath := filepath.Join(dirName, query.Filename)
	if err := common.DownloadFile(c, log, filePath, ""); err != nil {
		log.Error(
			"下载mon配置文件:执行失败",
			zap.Error(err),
			zap.String("file_path", filePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"下载mon配置文件:执行成功",
		zap.Uint32("node_id", uri.ID),
		zap.String("dir_name", dirName),
		zap.String("filename", query.Filename),
		zap.Duration("total_duration", time.Since(startTime)),
	)
}

// DeleteMonConf 删除mon节点配置文件。
// @Summary 删除mon节点配置文件
// @Description 删除指定的mon节点配置文件
// @Tags mon配置管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Param filename query string true "文件名"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/conf/{id} [delete]
// @Security ApiKeyAuth
func (h *MonConfHandler) DeleteMonConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除mon配置文件:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除mon配置文件:绑定删除的mon配置文件路径参数失败") {
		return
	}
	var query monmodel.MonConfFileQueryDTO
	if !common.ShouldBindQuery(c, log, &query, "删除mon配置文件:绑定删除的mon配置文件查询参数失败") {
		return
	}

	dirName := monsvc.GetMonNodeConfDir(uri.ID)
	filePath := filepath.Join(dirName, query.Filename)
	if err := fileutil.Remove(ctx, filePath); err != nil {
		log.Error(
			"删除mon配置文件:执行失败",
			zap.Error(err),
			zap.String("file_path", filePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, errors.FromError(err))
		return
	}

	log.Info(
		"删除mon配置文件:执行成功",
		zap.Uint32("node_id", uri.ID),
		zap.String("dir_name", dirName),
		zap.String("filename", query.Filename),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// ListMonConf 获取mon节点配置文件列表。
// @Summary 获取mon节点配置文件列表
// @Description 获取指定mon节点配置目录下的文件列表
// @Tags mon配置管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Success 200 {object} monmodel.PagMonConfResp "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/conf/{id} [get]
// @Security ApiKeyAuth
func (h *MonConfHandler) ListMonConf(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "获取mon配置文件列表:绑定mon配置文件路径参数失败") {
		return
	}

	dirName := monsvc.GetMonNodeConfDir(uri.ID)
	info, err := fileutil.ListFileInfo(ctx, dirName)
	if err != nil {
		log.Error(
			"获取mon配置文件列表:查询失败",
			zap.Error(err),
			zap.String("dir_name", dirName),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, errors.FromError(err))
		return
	}

	c.JSON(http.StatusOK, monmodel.PagMonConfResp{
		Code: http.StatusOK,
		Data: info,
	})
}

func (h *MonConfHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/conf/:id", h.UploadMonConf)
	r.DELETE("/conf/:id", h.DeleteMonConf)
	r.GET("/conf/:id", h.ListMonConf)
	r.GET("/conf/:id/download", h.DownloadMonConf)
}
