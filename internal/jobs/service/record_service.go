package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"

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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	uc := auth.GetUserClaims(ctx)
	m, rErr := s.ucRecord.AsyncExecuteScript(ctx, biz.ExecuteRequest{
		ScriptID:    req.ScriptID,
		CommandArgs: req.CommandArgs,
		EnvVars:     req.EnvVars,
		Timeout:     req.Timeout,
		WorkDir:     req.WorkDir,
		TriggerType: "api",
		Username:    uc.Subject,
	})
	if rErr != nil {
		s.log.Error(
			"执行脚本失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询脚本执行记录详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucRecord.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询脚本执行记录详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询脚本执行记录列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询脚本执行记录列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	m, err := s.ucRecord.FindScriptRecordByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	if err := common.DownloadFile(ctx, s.log, m.LogPath(), m.LogName); err != nil {
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
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
	return &pbRecord.ScriptRecordDetailOut{
		ScriptRecordStandardOut: *ScriptRecordToStandardOut(m),
		Script:                  *script,
	}
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
