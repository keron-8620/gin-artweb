package service

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobsmodel "gin-artweb/internal/model/jobs"
	jobsvc "gin-artweb/internal/service/jobs"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScriptRecordHandler struct {
	log       *zap.Logger
	svcRecord *jobsvc.RecordService
}

func NewScriptRecordHandler(
	log *zap.Logger,
	svcRecord *jobsvc.RecordService,
) *ScriptRecordHandler {
	return &ScriptRecordHandler{
		log:       log,
		svcRecord: svcRecord,
	}
}

// @Summary 执行脚本
// @Description 本接口用于执行指定的脚本并记录执行结果
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param request body jobsmodel.CreateScriptRecordRequest true "执行脚本请求参数"
// @Success 200 {object} jobsmodel.ScriptRecordReply "成功返回执行记录信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record [post]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) ExecScriptRecord(ctx *gin.Context) {
	var req jobsmodel.CreateScriptRecordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定执行脚本参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := h.svcRecord.AsyncExecuteScript(ctx, jobsmodel.ExecuteRequest{
		ScriptID:    req.ScriptID,
		CommandArgs: req.CommandArgs,
		EnvVars:     req.EnvVars,
		Timeout:     req.Timeout,
		WorkDir:     req.WorkDir,
		TriggerType: "api",
		Username:    claims.Subject,
	})
	if rErr != nil {
		h.log.Error(
			"执行脚本失败",
			zap.Error(rErr),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	ctx.JSON(http.StatusOK, &jobsmodel.ScriptRecordReply{
		Code: http.StatusOK,
		Data: *jobsmodel.ScriptRecordToDetailOut(*m),
	})
}

// @Summary 查询脚本执行记录详情
// @Description 本接口用于查询指定ID的脚本执行记录详情
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param id path uint true "执行记录编号"
// @Success 200 {object} jobsmodel.ScriptRecordReply "成功返回执行记录信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "执行记录未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record/{id} [get]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) GetScriptRecord(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询脚本执行记录详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcRecord.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询脚本执行记录详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := jobsmodel.ScriptRecordToDetailOut(*m)
	ctx.JSON(http.StatusOK, &jobsmodel.ScriptRecordReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询脚本执行记录列表
// @Description 本接口用于查询脚本执行记录列表
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param request query jobsmodel.ListScriptRecordRequest false "查询参数"
// @Success 200 {object} jobsmodel.PagScriptRecordReply "成功返回执行记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record [get]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) ListScriptRecord(ctx *gin.Context) {
	var req jobsmodel.ListScriptRecordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询脚本执行记录列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询脚本执行记录列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Script"},
		IsCount:  true,
		Size:     size,
		Page:     page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := h.svcRecord.ListcriptRecord(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询脚本执行记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询脚本执行记录列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := jobsmodel.ListScriptRecordToDetailOut(ms)
	ctx.JSON(http.StatusOK, &jobsmodel.PagScriptRecordReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 下载脚本执行日志
// @Description 本接口用于下载指定执行记录的日志文件
// @Tags 脚本执行记录
// @Accept json
// @Produce application/octet-stream
// @Param id path uint true "执行记录编号"
// @Success 200 {file} file "成功下载日志文件"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "执行记录未找到或日志文件不存在"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record/{id}/log [get]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) DownloadScriptRecordLog(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, err := h.svcRecord.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	logPath := common.GetScriptLogStoragePath(m.CreatedAt, m.LogName)

	if err := common.DownloadFile(ctx, h.log, logPath, m.LogName); err != nil {
		errors.RespondWithError(ctx, err)
	}
}

// @Summary 实时获取脚本执行日志
// @Description 本接口用于实时获取指定执行记录的日志内容
// @Tags 脚本执行记录
// @Produce text/plain
// @Param id path uint true "执行记录编号"
// @Success 200 {string} string "实时日志流"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "执行记录未找到或日志文件不存在"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record/{id}/log/stream [get]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) StreamScriptRecordLog(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := h.svcRecord.FindScriptRecordByID(ctx, []string{}, uri.ID)
	if rErr != nil {
		h.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	logPath := common.GetScriptLogStoragePath(m.CreatedAt, m.LogName)

	// 检查日志文件是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		h.log.Error(
			"日志文件不存在",
			zap.String("log_path", logPath),
			zap.Uint32("script_record_id", uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrScriptNotFound.WithField("script_record_id", uri.ID)
		errors.RespondWithError(ctx, rErr)
		return
	}

	// 初始化文件信息
	file, err := os.Open(logPath)
	if err != nil {
		h.log.Error(
			"打开日志文件失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	defer file.Close()

	// 监听客户端断开连接
	clientGone := ctx.Writer.CloseNotify()

	// 移动到文件末尾，准备读取新内容
	fileInfo, err := file.Stat()
	if err != nil {
		h.log.Error(
			"获取日志文件状态失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	lastModTime := fileInfo.ModTime()
	currentSize := fileInfo.Size()

	// 发送初始数据
	initialBytes := make([]byte, currentSize)
	n, err := file.Read(initialBytes)
	if err != nil && err != io.EOF {
		h.log.Error(
			"读取日志文件失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if n > 0 {
		initialLines := strings.SplitSeq(string(initialBytes[:n]), "\n")
		for line := range initialLines {
			if line != "" {
				sse.Encode(ctx.Writer, sse.Event{
					Data: []byte(line),
				})
			}
		}
		ctx.Writer.Flush()
	}

	// 检查任务是否仍在运行
	if cancel := h.svcRecord.GetCancel(uri.ID); cancel == nil {
		// 任务已完成，结束流
		return
	}

	// 定期检查文件是否有新内容
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientGone: // 客户端已断开连接
			return
		case <-ctx.Done(): // ctx 取消
			return
		case <-ticker.C:
			// 检查文件大小是否有变化
			fileInfo, err := os.Stat(logPath)
			if err != nil {
				h.log.Error(
					"获取日志文件状态失败",
					zap.Error(err),
					zap.String("log_path", logPath),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return
			}

			newSize := fileInfo.Size()

			// 如果文件大小增加了，说明有新内容
			if newSize > currentSize {
				// 移动到上次读取的位置
				_, err := file.Seek(currentSize, 0)
				if err != nil {
					h.log.Error(
						"移动文件指针失败",
						zap.Error(err),
						zap.String("log_path", logPath),
						zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
					)
					return
				}

				// 读取新增的内容
				buf := make([]byte, newSize-currentSize)
				n, err := file.Read(buf)
				if err != nil && err != io.EOF {
					h.log.Error(
						"读取日志文件失败",
						zap.Error(err),
						zap.String("log_path", logPath),
						zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
					)
					return
				}

				if n > 0 {
					newLines := strings.SplitSeq(string(buf[:n]), "\n")
					for line := range newLines {
						if line != "" {
							sse.Encode(ctx.Writer, sse.Event{
								Data: []byte(line),
							})
						}
					}
					ctx.Writer.Flush()
				}

				currentSize = newSize
			} else if fileInfo.ModTime().After(lastModTime) {
				// 文件修改时间更新了，可能有追加内容
				_, err := file.Seek(currentSize, 0)
				if err != nil {
					h.log.Error(
						"移动文件指针失败",
						zap.Error(err),
						zap.String("log_path", logPath),
						zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
					)
					return
				}
				buf := make([]byte, 1024) // 尝试读取1KB内容
				for {
					n, err := file.Read(buf)
					if err != nil && err != io.EOF {
						h.log.Error(
							"读取日志文件失败",
							zap.Error(err),
							zap.String("log_path", logPath),
							zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
						)
						return
					}
					if n == 0 {
						break
					}

					newLines := strings.Split(string(buf[:n]), "\n")
					for i, line := range newLines {
						// 最后一行可能是不完整的，跳过
						if i == len(newLines)-1 && n == len(buf) {
							continue
						}
						if line != "" {
							sse.Encode(ctx.Writer, sse.Event{
								Data: []byte(line),
							})
						}
					}
					ctx.Writer.Flush()
					currentSize += int64(n)
				}

				lastModTime = fileInfo.ModTime()
			}

			// 检查任务是否已完成
			if cancel := h.svcRecord.GetCancel(uri.ID); cancel == nil {
				// 检测到取消信号，结束流
				return
			}
		}
	}
}

// @Summary 对正在执行的脚本发送终止信号
// @Description 本接口用于通过执行记录的id号,对正在执行的脚本发送终止信号
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param id path uint true "执行记录编号"
// @Success 200 {object} commodel.MapAPIReply "终止信号"
// @Router /api/v1/jobs/record/{id} [delete]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) CancelScriptRecord(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.svcRecord.Cancel(ctx, uri.ID)
	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

func (h *ScriptRecordHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/record", h.ExecScriptRecord)
	r.GET("/record/:id", h.GetScriptRecord)
	r.GET("/record", h.ListScriptRecord)
	r.GET("/record/:id/log", h.DownloadScriptRecordLog)
	r.GET("/record/:id/log/stream", h.StreamScriptRecordLog)
	r.DELETE("/record/:id", h.CancelScriptRecord)
}
