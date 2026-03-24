package oes

import (
	"context"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repo/oes"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type StkTaskService struct {
	log        *zap.Logger
	recordSvc  *jobsvc.RecordService
	colonyRepo *oesrepo.OesColonyRepo
}

func NewStkTaskService(
	log *zap.Logger,
	recordSvc *jobsvc.RecordService,
	colonyRepo *oesrepo.OesColonyRepo,
) *StkTaskService {
	return &StkTaskService{
		log:        log,
		recordSvc:  recordSvc,
		colonyRepo: colonyRepo,
	}
}

func (s *StkTaskService) BuildTaskExecutionInfos(
	ctx context.Context,
	dto oesmodel.ListOesColonyDTO,
) ([]oesmodel.StkColonyTaskExecutionInfo, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	qp := database.QueryParams{
		Preloads: nil,
		OrderBy:  []string{"id DESC"},
		Query:    dto.ToQueryMap(),
	}
	ms, err := s.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		return nil, errors.NewGormError(err, qp.Query)
	}
	trs := make([]oesmodel.StkColonyTaskRecordIDs, len(ms))
	for i, m := range ms {
		tr, err := loadStkTaskRecordCacheFromFiles(log, m.ColonyNum)
		if err != nil {
			return nil, errors.FromError(err)
		}
		if tr != nil {
			trs[i] = *tr
		}
	}
	var recordIDs []uint32
	for _, tr := range trs {
		recordIDs = append(recordIDs, tr.GetValidRecordIDs()...)
	}
	records, rErr := s.recordSvc.ListScriptRecordByIDs(ctx, nil, recordIDs)
	if rErr != nil {
		return nil, rErr
	}
	cache := make(map[uint32]jobmodel.ScriptRecordModel, len(records))
	for _, r := range records {
		cache[r.ID] = r
	}

	tasks := make([]oesmodel.StkColonyTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		if ctx.Err() != nil {
			return nil, errors.FromError(ctx.Err())
		}
		tasks[i] = oesmodel.StkColonyTaskExecutionInfo{
			ColonyNum:         tr.ColonyNum,
			Mon:               jobsvc.GetRecordIDByMap(cache, tr.Mon),
			CounterFetch:      jobsvc.GetRecordIDByMap(cache, tr.CounterFetch),
			CounterDistribute: jobsvc.GetRecordIDByMap(cache, tr.CounterDistribute),
			Bse:               jobsvc.GetRecordIDByMap(cache, tr.Bse),
			Sse:               jobsvc.GetRecordIDByMap(cache, tr.Sse),
			Szse:              jobsvc.GetRecordIDByMap(cache, tr.Szse),
			Csdc:              jobsvc.GetRecordIDByMap(cache, tr.Csdc),
		}
	}
	return tasks, nil
}

func loadStkTaskRecordCacheFromFiles(
	log *zap.Logger,
	colonyNum string,
) (*oesmodel.StkColonyTaskRecordIDs, *errors.Error) {
	startTime := time.Now()

	log.Info(
		"读取stk任务状态对应的执行记录id：开始执行",
		zap.String("colony_num", colonyNum),
	)

	flagDir := filepath.Join(config.StorageDir, "oes", "flags", colonyNum)
	mc := oesmodel.StkColonyTaskRecordIDs{
		ColonyNum: colonyNum,
	}

	// 定义任务映射表，减少代码重复
	taskMap := map[string]*uint32{
		"mon":                &mc.Mon,
		"counter_fetch":      &mc.CounterFetch,
		"counter_distribute": &mc.CounterDistribute,
		"bse":                &mc.Bse,
		"sse":                &mc.Sse,
		"szse":               &mc.Szse,
		"csdc":               &mc.Csdc,
	}

	// 遍历处理每个任务
	for taskName, fieldPtr := range taskMap {
		flagPath := filepath.Join(flagDir, "."+taskName)
		if value, err := common.ReadUint32FromFile(flagPath); err != nil {
			log.Error(
				"读取stk任务状态对应的执行记录id：获取"+taskName+"任务状态失败",
				zap.Error(err),
				zap.String("colony_num", colonyNum),
				zap.String("task_name", taskName),
				zap.String("flag_path", flagPath),
			)
			return nil, errors.FromError(err)
		} else {
			*fieldPtr = value
		}
	}

	log.Debug(
		"读取stk任务状态对应的执行记录id：任务状态读取成功",
		zap.Object("stk_task_record_ids", &mc),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &mc, nil
}
