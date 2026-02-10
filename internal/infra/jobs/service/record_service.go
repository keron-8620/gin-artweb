package service

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/infra/jobs/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"

	pbComm "gin-artweb/api/common"
	pbRecord "gin-artweb/api/jobs/record"
	pbScript "gin-artweb/api/jobs/script"
)

type ScriptRecordService struct {
	log      *zap.Logger
	ucRecord *biz.RecordUsecase
}

func NewScriptRecordService(
	log *zap.Logger,
	ucRecord *biz.RecordUsecase,
) *ScriptRecordService {
	return &ScriptRecordService{
		log:      log,
		ucRecord: ucRecord,
	}
}

// @Summary 执行脚本
// @Description 本接口用于执行指定的脚本并记录执行结果
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param request body pbRecord.CreateScriptRecordRequest true "执行脚本请求参数"
// @Success 200 {object} pbRecord.ScriptRecordReply "成功返回执行记录信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record [post]
// @Security ApiKeyAuth
func (s *ScriptRecordService) ExecScriptRecord(ctx *gin.Context) {
	var req pbRecord.CreateScriptRecordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定执行脚本参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		s.log.Error(
			"获取个人登录信息失败",
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := s.ucRecord.AsyncExecuteScript(ctx, biz.ExecuteRequest{
		ScriptID:    req.ScriptID,
		CommandArgs: req.CommandArgs,
		EnvVars:     req.EnvVars,
		Timeout:     req.Timeout,
		WorkDir:     req.WorkDir,
		TriggerType: "api",
		Username:    claims.Subject,
	})
	if rErr != nil {
		s.log.Error(
			"执行脚本失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	ctx.JSON(http.StatusOK, &pbRecord.ScriptRecordReply{
		Code: http.StatusOK,
		Data: *ScriptRecordToDetailOut(*m),
	})
}

// @Summary 查询脚本执行记录详情
// @Description 本接口用于查询指定ID的脚本执行记录详情
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param id path uint true "执行记录编号"
// @Success 200 {object} pbRecord.ScriptRecordReply "成功返回执行记录信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "执行记录未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record/{id} [get]
// @Security ApiKeyAuth
func (s *ScriptRecordService) GetScriptRecord(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询脚本执行记录详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucRecord.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"查询脚本执行记录详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := ScriptRecordToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbRecord.ScriptRecordReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询脚本执行记录列表
// @Description 本接口用于查询脚本执行记录列表
// @Tags 脚本执行记录
// @Accept json
// @Produce json
// @Param request query pbRecord.ListScriptRecordRequest false "查询参数"
// @Success 200 {object} pbRecord.PagScriptRecordReply "成功返回执行记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/record [get]
// @Security ApiKeyAuth
func (s *ScriptRecordService) ListScriptRecord(ctx *gin.Context) {
	var req pbRecord.ListScriptRecordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询脚本执行记录列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询脚本执行记录列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Script"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucRecord.ListcriptRecord(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"查询脚本执行记录列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := ListScriptRecordToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbRecord.PagScriptRecordReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
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
func (s *ScriptRecordService) DownloadScriptRecordLog(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, err := s.ucRecord.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	if err := common.DownloadFile(ctx, s.log, m.LogPath(), m.LogName); err != nil {
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
func (s *ScriptRecordService) StreamScriptRecordLog(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := s.ucRecord.FindScriptRecordByID(ctx, []string{}, uri.ID)
	if rErr != nil {
		s.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(rErr),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	logPath := m.LogPath()

	// 检查日志文件是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		s.log.Error(
			"日志文件不存在",
			zap.String("log_path", logPath),
			zap.Uint32(biz.ScriptRecordIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrScriptNotFound.WithField(biz.ScriptRecordIDKey, uri.ID)
		errors.RespondWithError(ctx, rErr)
		return
	}

	// 初始化文件信息
	file, err := os.Open(logPath)
	if err != nil {
		s.log.Error(
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

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")
	ctx.Header("X-Accel-Buffering", "no")

	clientGone := ctx.Writer.CloseNotify()

	// 移动到文件末尾，准备读取新内容
	fileInfo, _ := file.Stat()
	lastModTime := fileInfo.ModTime()
	currentSize := fileInfo.Size()

	// 发送初始数据
	initialBytes := make([]byte, currentSize)
	n, _ := file.Read(initialBytes)
	if n > 0 {
		initialLines := strings.SplitSeq(string(initialBytes[:n]), "\n")
		for line := range initialLines {
			if line != "" {
				fmt.Fprintf(ctx.Writer, "data: %s\n", line)
			}
		}
		fmt.Fprintf(ctx.Writer, "\n")
		// ctx.Writer.Flush()
	}

	// 检查任务是否仍在运行
	// 如果GetCancel返回nil，表示任务已经完成（无论成功、失败、超时或被取消）
	// 如果GetCancel返回非nil，表示任务仍在运行
	if cancel := s.ucRecord.GetCancel(uri.ID); cancel == nil {
		// 任务已完成，结束流
		return
	}

	// 定期检查文件是否有新内容
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次
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
				s.log.Error(
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
				file.Seek(currentSize, 0)

				// 读取新增的内容
				buf := make([]byte, newSize-currentSize)
				n, _ := file.Read(buf)

				if n > 0 {
					newLines := strings.SplitSeq(string(buf[:n]), "\n")
					for line := range newLines {
						if line != "" {
							fmt.Fprintf(ctx.Writer, "data: %s\n", line)
						}
					}
					fmt.Fprintf(ctx.Writer, "\n")
					ctx.Writer.Flush()
				}

				currentSize = newSize
			} else if fileInfo.ModTime().After(lastModTime) {
				// 文件修改时间更新了，可能有追加内容
				file.Seek(currentSize, 0)
				buf := make([]byte, 1024) // 尝试读取1KB内容
				for {
					n, _ := file.Read(buf)
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
							fmt.Fprintf(ctx.Writer, "data: %s\n", line)
						}
					}
					fmt.Fprintf(ctx.Writer, "\n")
					ctx.Writer.Flush()
					currentSize += int64(n)
				}

				lastModTime = fileInfo.ModTime()
			}

			// 检查任务是否已完成
			if cancel := s.ucRecord.GetCancel(uri.ID); cancel == nil {
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
// @Success 200 {object} pbComm.MapAPIReply "终止信号"
// @Router /api/v1/jobs/record/{id} [delete]
// @Security ApiKeyAuth
func (s *ScriptRecordService) CancelScriptRecord(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询脚本执行记录ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.ucRecord.Cancel(ctx, uri.ID)
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

func (s *ScriptRecordService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/record", s.ExecScriptRecord)
	r.GET("/record/:id", s.GetScriptRecord)
	r.GET("/record", s.ListScriptRecord)
	r.GET("/record/:id/log", s.DownloadScriptRecordLog)
	r.GET("/record/:id/log/stream", s.StreamScriptRecordLog)
	r.DELETE("/record/:id", s.CancelScriptRecord)
}

func ScriptRecordToStandardOut(
	m biz.ScriptRecordModel,
) *pbRecord.ScriptRecordStandardOut {
	return &pbRecord.ScriptRecordStandardOut{
		ID:           m.ID,
		CreatedAt:    m.CreatedAt.Format(time.DateTime),
		UpdatedAt:    m.UpdatedAt.Format(time.DateTime),
		TriggerType:  m.TriggerType,
		Status:       m.Status,
		ExitCode:     m.ExitCode,
		EnvVars:      m.EnvVars,
		CommandArgs:  m.CommandArgs,
		Timeout:      m.Timeout,
		WorkDir:      m.WorkDir,
		ErrorMessage: m.ErrorMessage,
		Username:     m.Username,
	}
}

func ScriptRecordToDetailOut(
	m biz.ScriptRecordModel,
) *pbRecord.ScriptRecordDetailOut {
	var script *pbScript.ScriptStandardOut
	if m.Script.ID != 0 {
		script = ScriptModelToStandardOut(m.Script)
	}

	standardOut := ScriptRecordToStandardOut(m)
	result := &pbRecord.ScriptRecordDetailOut{
		ScriptRecordStandardOut: *standardOut,
	}

	if script != nil {
		result.Script = *script
	}

	return result
}

func ListScriptRecordToDetailOut(
	rms *[]biz.ScriptRecordModel,
) *[]pbRecord.ScriptRecordDetailOut {
	if rms == nil {
		return &[]pbRecord.ScriptRecordDetailOut{}
	}

	ms := *rms
	mso := make([]pbRecord.ScriptRecordDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScriptRecordToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
