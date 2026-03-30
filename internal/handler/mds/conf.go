package mds

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	mdsmodel "gin-artweb/internal/model/mds"
	mdsvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
)

type MdsConfHandler struct {
	log     *zap.Logger
	maxSize int64
}

func NewMdsConfHandler(
	logger *zap.Logger,
	maxSize int64,
) *MdsConfHandler {
	return &MdsConfHandler{
		log:     logger,
		maxSize: maxSize,
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
// @Success 200 {object} commodel.MapAPIResp "上传成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf/{dir_name} [post]
// @Security ApiKeyAuth
// UploadMdsConf 上传mds配置文件
func (s *MdsConfHandler) UploadMdsConf(ctx *gin.Context) {
	log := ctxutil.NewLogger(s.log, ctx)
	var pathReq mdsmodel.GetMdsConfDTO
	if !common.ShouldBind(
		ctx, log, &pathReq,
		"上传mds配置文件：绑定上传的mds配置文件路径参数失败") {
		return
	}

	var formReq mdsmodel.UploadMdsConfDTO
	if !common.ShouldBind(
		ctx, log, &formReq,
		"上传mds配置文件：绑定上传的mds配置文件表单参数失败") {
		return
	}

	// 3. 将配置文件保存到指定的位置
	dirName := mdsvc.GetMdsColonyConfigDir(pathReq.ColonyNum)
	savePath := filepath.Join(dirName, pathReq.DirName, formReq.File.Filename)
	if err := common.UploadFile(ctx, log, s.maxSize, savePath, formReq.File, 0o644); err != nil {
		errors.RespondWithError(ctx, err)
		return
	}

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (s *MdsConfHandler) DownloadMdsConf(ctx *gin.Context) {
	log := ctxutil.NewLogger(s.log, ctx)
	var req mdsmodel.DownloadOrDeleteMdsConfRequest
	if !common.ShouldBind(
		ctx, log, &req,
		"下载mds配置文件：绑定下载的mds配置文件路径参数失败") {
		return
	}

	dirName := mdsvc.GetMdsColonyConfigDir(req.ColonyNum)
	filePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := common.DownloadFile(ctx, log, filePath, ""); err != nil {
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
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf/{dir_name}/{filename} [delete]
// @Security ApiKeyAuth
func (s *MdsConfHandler) DeleteMdsConf(ctx *gin.Context) {
	log := ctxutil.NewLogger(s.log, ctx)
	var req mdsmodel.DownloadOrDeleteMdsConfRequest
	if !common.ShouldBind(
		ctx, log, &req,
		"删除mds配置文件：绑定删除的mds配置文件路径参数失败") {
		return
	}

	dirName := mdsvc.GetMdsColonyConfigDir(req.ColonyNum)
	savePath := filepath.Join(dirName, req.DirName, req.Filename)
	if err := fileutil.Remove(ctx, savePath); err != nil {
		log.Error(
			"删除mds配置文件：执行失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// ListMdsConf 获取mds配置文件列表
// @Summary 获取mds配置文件列表
// @Description 获取指定目录下的mds配置文件列表
// @Tags mds配置管理
// @Accept json
// @Produce json
// @Param colony_num path string true "集群编号"
// @Success 200 {object} mdsmodel.PagMdsConfResp "成功返回配置文件列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/{colony_num}/conf [get]
// @Security ApiKeyAuth
func (s *MdsConfHandler) ListMdsConf(ctx *gin.Context) {
	log := ctxutil.NewLogger(s.log, ctx)
	var req mdsmodel.ListMdsConfDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"获取mds配置文件列表：绑定mds配置文件路径参数失败") {
		return
	}

	dirName := mdsvc.GetMdsColonyConfigDir(req.ColonyNum)
	info, err := fileutil.ListFileInfo(ctx, dirName)
	if err != nil {
		log.Error(
			"获取mds配置文件列表：执行失败",
			zap.Error(err),
			zap.String("dirname", dirName),
			zap.String("request_uri", ctx.Request.RequestURI),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	ctx.JSON(http.StatusOK, mdsmodel.PagMdsConfResp{
		Code: http.StatusOK,
		Data: info,
	})
}

func (s *MdsConfHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/:colony_num/conf/:dir_name", s.UploadMdsConf)
	r.GET("/:colony_num/conf/:dir_name/:filename", s.DownloadMdsConf)
	r.DELETE("/:colony_num/conf/:dir_name/:filename", s.DeleteMdsConf)
	r.GET("/:colony_num/conf", s.ListMdsConf)
}
