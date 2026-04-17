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

type OptTaskService struct {
	log        *zap.Logger
	recordSvc  *jobsvc.RecordService
	colonyRepo *oesrepo.OesColonyRepo
}

func NewOptTaskService(
	log *zap.Logger,
	recordSvc *jobsvc.RecordService,
	colonyRepo *oesrepo.OesColonyRepo,
) *OptTaskService {
	return &OptTaskService{
		log:        log,
		recordSvc:  recordSvc,
		colonyRepo: colonyRepo,
	}
}

func (s *OptTaskService) BuildTaskExecutionInfos(
	ctx context.Context,
	dto oesmodel.ListOesColonyDTO,
) ([]oesmodel.OptColonyTaskExecutionInfo, *errors.Error) {
	startTime := time.Now()
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
		log.Error(
			"查询oes期权集群列表:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	trs := make([]oesmodel.OptColonyTaskRecordIDs, len(ms))
	for i, m := range ms {
		tr, err := loadOptTaskRecordCacheFromFiles(log, m.ColonyNum)
		if err != nil {
			log.Error(
				"查询oes期权集群列表:查询执行记录失败",
				zap.Error(err),
				zap.String("colony_num", m.ColonyNum),
				zap.Duration("total_duration", time.Since(startTime)),
			)
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
		log.Error(
			"查询oes期权集群列表:查询脚本记录失败",
			zap.Error(rErr),
			zap.Uint32s("record_ids", recordIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	cache := make(map[uint32]jobmodel.ScriptRecordModel, len(records))
	for _, r := range records {
		cache[r.ID] = r
	}

	tasks := make([]oesmodel.OptColonyTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		if ctx.Err() != nil {
			return nil, errors.FromError(ctx.Err())
		}
		tasks[i] = oesmodel.OptColonyTaskExecutionInfo{
			ColonyNum:         tr.ColonyNum,
			Mon:               jobsvc.GetRecordIDByMap(cache, tr.Mon),
			CounterFetch:      jobsvc.GetRecordIDByMap(cache, tr.CounterFetch),
			CounterDistribute: jobsvc.GetRecordIDByMap(cache, tr.CounterDistribute),
			Sse:               jobsvc.GetRecordIDByMap(cache, tr.Sse),
			Szse:              jobsvc.GetRecordIDByMap(cache, tr.Szse),
		}
	}
	return tasks, nil
}

func loadOptTaskRecordCacheFromFiles(
	log *zap.Logger,
	colonyNum string,
) (*oesmodel.OptColonyTaskRecordIDs, *errors.Error) {
	startTime := time.Now()

	flagDir := filepath.Join(config.StorageDir, "oes", "flags", colonyNum)
	mc := oesmodel.OptColonyTaskRecordIDs{
		ColonyNum: colonyNum,
	}

	// 定义任务映射表，减少代码重复
	taskMap := map[string]*uint32{
		"mon":                &mc.Mon,
		"counter_fetch":      &mc.CounterFetch,
		"counter_distribute": &mc.CounterDistribute,
		"sse":                &mc.Sse,
		"szse":               &mc.Szse,
	}

	// 遍历处理每个任务
	for taskName, fieldPtr := range taskMap {
		flagPath := filepath.Join(flagDir, "."+taskName)
		if value, err := common.ReadUint32FromFile(flagPath); err != nil {
			log.Error(
				"读取opt任务状态对应的执行记录id:获取"+taskName+"任务状态失败",
				zap.Error(err),
				zap.String("colony_num", colonyNum),
				zap.String("task_name", taskName),
				zap.String("flag_path", flagPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil, errors.FromError(err)
		} else {
			*fieldPtr = value
		}
	}
	return &mc, nil
}
