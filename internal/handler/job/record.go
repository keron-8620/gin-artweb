package job

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
	jobmodel "gin-artweb/internal/model/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type ScriptRecordHandler struct {
	log       *zap.Logger
	recordSvc *jobsvc.RecordService
}

func NewScriptRecordHandler(
	log *zap.Logger,
	svcRecord *jobsvc.RecordService,
) *ScriptRecordHandler {
	return &ScriptRecordHandler{
		log:       log,
		recordSvc: svcRecord,
	}
}

// @Summary 执行脚本
// @Description 本接口用于执行指定的脚本并记录执行结果
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param request body jobmodel.ExecScriptDTO true "执行脚本请求参数"
// @Success 200 {object} jobmodel.ScriptRecordResp "成功返回执行记录信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record [post]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) ExecScriptRecord(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req jobmodel.CreateScriptRecordDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"执行脚本:绑定请求参数失败") {
		return
	}

	log.Info("执行脚本:开始执行")

	log.Debug(
		"执行脚本:入参详情",
		zap.Object("exec_script_dto", &req),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	execStepStart := time.Now()
	m, rErr := h.recordSvc.AsyncExecuteScript(ctx, jobmodel.ExecuteScriptDTO{
		ScriptID:    req.ScriptID,
		CommandArgs: req.CommandArgs,
		EnvVars:     req.EnvVars,
		Timeout:     req.Timeout,
		WorkDir:     req.WorkDir,
		TriggerType: "api",
		Username:    claims.Subject,
	})
	execStepDuration := time.Since(execStepStart)
	if rErr != nil {
		log.Error(
			"执行脚本:执行失败",
			zap.Error(rErr),
			zap.Uint32("script_id", req.ScriptID),
			zap.Duration("exec_step_duration", execStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Debug(
		"执行脚本:执行记录模型详情",
		zap.Object("record_model", m),
		zap.Duration("exec_step_duration", execStepDuration),
	)

	log.Info(
		"执行脚本:执行成功",
		zap.Uint32("record_id", m.ID),
		zap.Uint32("script_id", req.ScriptID),
		zap.Duration("exec_step_duration", execStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &jobmodel.ScriptRecordResp{
		Code: http.StatusOK,
		Data: *jobmodel.ScriptRecordToDetailOut(*m),
	})
}

// @Summary 查询脚本执行记录详情
// @Description 本接口用于查询指定ID的脚本执行记录详情
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param id path uint true "执行记录编号"
// @Success 200 {object} jobmodel.ScriptRecordResp "成功返回执行记录信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "执行记录未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record/{id} [get]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) GetScriptRecord(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"查询脚本执行记录:绑定记录ID参数失败") {
		return
	}

	log.Info(
		"查询脚本执行记录:开始执行",
		zap.Uint32("record_id", uri.ID),
	)

	findStepStart := time.Now()
	m, err := h.recordSvc.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询脚本执行记录:执行失败",
			zap.Error(err),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Debug(
		"查询脚本执行记录:执行记录模型详情",
		zap.Object("record_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"查询脚本执行记录:执行成功",
		zap.Uint32("record_id", uri.ID),
		zap.Uint32("script_id", m.ScriptID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := jobmodel.ScriptRecordToDetailOut(*m)
	ctx.JSON(http.StatusOK, &jobmodel.ScriptRecordResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询脚本执行记录列表
// @Description 本接口用于查询脚本执行记录列表
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param request query jobmodel.ListScriptRecordDTO false "查询参数"
// @Success 200 {object} jobmodel.PagScriptRecordResp "成功返回执行记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record [get]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) ListScriptRecord(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req jobmodel.ListScriptRecordDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询脚本执行记录列表:绑定查询参数失败") {
		return
	}

	log.Info("查询脚本执行记录列表:开始执行")

	log.Debug(
		"查询脚本执行记录列表:入参详情",
		zap.Object("list_script_record_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.recordSvc.ListScriptRecord(ctx, page, size, &req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询脚本执行记录列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_script_record_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询脚本执行记录列表:执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := jobmodel.ListScriptRecordToDetailOut(ms)
	ctx.JSON(http.StatusOK, &jobmodel.PagScriptRecordResp{
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
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"下载脚本执行日志:绑定记录ID参数失败") {
		return
	}

	log.Info(
		"下载脚本执行日志:开始执行",
		zap.Uint32("record_id", uri.ID),
	)

	findStepStart := time.Now()
	m, err := h.recordSvc.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"下载脚本执行日志:查询记录详情失败",
			zap.Error(err),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"下载脚本执行日志:执行记录模型详情",
		zap.Object("record_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	logPath := h.recordSvc.GenerateScriptLogPath(m.CreatedAt, m.LogName)
	downloadStepStart := time.Now()
	err = common.DownloadFile(ctx, log, logPath, m.LogName)
	downloadStepDuration := time.Since(downloadStepStart)
	if err != nil {
		log.Error(
			"下载脚本执行日志:下载日志失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.String("rename", m.LogName),
			zap.Duration("download_step_duration", downloadStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"下载脚本执行日志:下载日志成功",
		zap.Uint32("record_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("download_step_duration", downloadStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"实时获取脚本执行日志:绑定记录ID参数失败") {
		return
	}

	log.Info(
		"实时获取脚本执行日志:开始执行",
		zap.Uint32("record_id", uri.ID),
	)

	findStepStart := time.Now()
	log.Debug(
		"实时获取脚本执行日志:开始查询记录详情",
		zap.Uint32("record_id", uri.ID),
	)
	m, rErr := h.recordSvc.FindScriptRecordByID(ctx, []string{}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if rErr != nil {
		log.Error(
			"实时获取脚本执行日志:查询记录详情失败",
			zap.Error(rErr),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	log.Debug(
		"实时获取脚本执行日志:查询到的记录模型详情",
		zap.Object("record_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	// 检查日志文件是否存在
	logPath := h.recordSvc.GenerateScriptLogPath(m.CreatedAt, m.LogName)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		log.Error(
			"实时获取脚本执行日志:日志文件不存在",
			zap.String("log_path", logPath),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrScriptNotFound.WithField("script_record_id", uri.ID)
		errors.RespondWithError(ctx, rErr)
		return
	}

	// 初始化文件信息
	file, err := os.Open(logPath)
	if err != nil {
		log.Error(
			"实时获取脚本执行日志:打开日志文件失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
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
		log.Error(
			"实时获取脚本执行日志:获取日志文件状态失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
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
		log.Error(
			"实时获取脚本执行日志:读取日志文件失败",
			zap.Error(err),
			zap.String("log_path", logPath),
			zap.Uint32("record_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.FromError(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if n > 0 {
		initialLines := strings.SplitSeq(string(initialBytes[:n]), "\n")
		for line := range initialLines {
			if line != "" {
				if err := sse.Encode(ctx.Writer, sse.Event{
					Data: []byte(line),
				}); err != nil {
					log.Error(
						"实时获取脚本执行日志:发送日志行失败",
						zap.Error(err),
						zap.Uint32("record_id", uri.ID),
						zap.Duration("total_duration", time.Since(startTime)),
					)
					rErr := errors.FromError(err)
					errors.RespondWithError(ctx, rErr)
					return
				}
			}
		}
		ctx.Writer.Flush()
	}

	// 检查任务是否仍在运行
	if cancel := h.recordSvc.GetCancel(uri.ID); cancel == nil {
		// 任务已完成，结束流
		log.Info(
			"实时获取脚本执行日志:任务已完成，结束流",
			zap.Uint32("record_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return
	}

	log.Info(
		"实时获取脚本执行日志:开始监控日志文件",
		zap.Uint32("record_id", uri.ID),
		zap.String("log_path", logPath),
	)

	// 定期检查文件是否有新内容
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientGone: // 客户端已断开连接
			log.Info(
				"实时获取脚本执行日志:客户端已断开连接",
				zap.Uint32("record_id", uri.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return
		case <-ctx.Done(): // ctx 取消
			log.Info(
				"实时获取脚本执行日志:上下文已取消",
				zap.Uint32("record_id", uri.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return
		case <-ticker.C:
			// 检查文件大小是否有变化
			fileInfo, err := os.Stat(logPath)
			if err != nil {
				log.Error(
					"实时获取脚本执行日志:获取日志文件状态失败",
					zap.Error(err),
					zap.String("log_path", logPath),
					zap.Uint32("record_id", uri.ID),
				)
				return
			}

			newSize := fileInfo.Size()

			// 如果文件大小增加了，说明有新内容
			if newSize > currentSize {
				// 移动到上次读取的位置
				_, err := file.Seek(currentSize, 0)
				if err != nil {
					log.Error(
						"实时获取脚本执行日志:移动文件指针失败",
						zap.Error(err),
						zap.String("log_path", logPath),
						zap.Uint32("record_id", uri.ID),
					)
					return
				}

				// 读取新增的内容
				buf := make([]byte, newSize-currentSize)
				n, err := file.Read(buf)
				if err != nil && err != io.EOF {
					log.Error(
						"实时获取脚本执行日志:读取日志文件失败",
						zap.Error(err),
						zap.String("log_path", logPath),
						zap.Uint32("record_id", uri.ID),
					)
					return
				}

				if n > 0 {
					newLines := strings.SplitSeq(string(buf[:n]), "\n")
					for line := range newLines {
						if line != "" {
							if err := sse.Encode(ctx.Writer, sse.Event{
								Data: []byte(line),
							}); err != nil {
								log.Error(
									"实时获取脚本执行日志:发送日志行失败",
									zap.Error(err),
									zap.Uint32("record_id", uri.ID),
									zap.Duration("total_duration", time.Since(startTime)),
								)
								rErr := errors.FromError(err)
								errors.RespondWithError(ctx, rErr)
								return
							}
						}
					}
					ctx.Writer.Flush()
				}

				currentSize = newSize
			} else if fileInfo.ModTime().After(lastModTime) {
				// 文件修改时间更新了，可能有追加内容
				_, err := file.Seek(currentSize, 0)
				if err != nil {
					log.Error(
						"实时获取脚本执行日志:移动文件指针失败",
						zap.Error(err),
						zap.String("log_path", logPath),
						zap.Uint32("record_id", uri.ID),
					)
					return
				}
				buf := make([]byte, 1024) // 尝试读取1KB内容
				for {
					n, err := file.Read(buf)
					if err != nil && err != io.EOF {
						log.Error(
							"实时获取脚本执行日志:读取日志文件失败",
							zap.Error(err),
							zap.String("log_path", logPath),
							zap.Uint32("record_id", uri.ID),
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
							if err := sse.Encode(ctx.Writer, sse.Event{
								Data: []byte(line),
							}); err != nil {
								log.Error(
									"实时获取脚本执行日志:发送日志行失败",
									zap.Error(err),
									zap.Uint32("record_id", uri.ID),
									zap.Duration("total_duration", time.Since(startTime)),
								)
								rErr := errors.FromError(err)
								errors.RespondWithError(ctx, rErr)
								return
							}
						}
					}
					ctx.Writer.Flush()
					currentSize += int64(n)
				}

				lastModTime = fileInfo.ModTime()
			}

			// 检查任务是否已完成
			if cancel := h.recordSvc.GetCancel(uri.ID); cancel == nil {
				// 检测到取消信号，结束流
				log.Info(
					"实时获取脚本执行日志:任务已完成，结束流",
					zap.Uint32("record_id", uri.ID),
					zap.Duration("total_duration", time.Since(startTime)),
				)
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
// @Success 200 {object} commodel.MapAPIResp "终止信号"
// @Router /api/v1/jobs/record/{id} [delete]
// @Security ApiKeyAuth
func (h *ScriptRecordHandler) CancelScriptRecord(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"取消脚本执行:绑定记录ID参数失败") {
		return
	}

	log.Info(
		"取消脚本执行:开始执行",
		zap.Uint32("record_id", uri.ID),
	)

	cancelStepStart := time.Now()
	h.recordSvc.Cancel(ctx, uri.ID)
	cancelStepDuration := time.Since(cancelStepStart)
	log.Info(
		"取消脚本执行:发送终止信号成功",
		zap.Uint32("record_id", uri.ID),
		zap.Duration("cancel_step_duration", cancelStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

func (h *ScriptRecordHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/record", h.ExecScriptRecord)
	r.GET("/record/:id", h.GetScriptRecord)
	r.GET("/record", h.ListScriptRecord)
	r.GET("/record/:id/log", h.DownloadScriptRecordLog)
	r.GET("/record/:id/log/stream", h.StreamScriptRecordLog)
	r.DELETE("/record/:id", h.CancelScriptRecord)
}
