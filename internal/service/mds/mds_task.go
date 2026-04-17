package mds

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repo/mds"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MdsTaskService struct {
	log        *zap.Logger
	recordSvc  *jobsvc.RecordService
	colonyRepo *mdsrepo.MdsColonyRepo
}

func NewMdsTaskService(
	log *zap.Logger,
	recordSvc *jobsvc.RecordService,
	colonyRepo *mdsrepo.MdsColonyRepo,
) *MdsTaskService {
	return &MdsTaskService{
		log:        log,
		recordSvc:  recordSvc,
		colonyRepo: colonyRepo,
	}
}

func (s *MdsTaskService) BuildTaskExecutionInfos(
	ctx context.Context,
	dto mdsmodel.ListMdsColonyDTO,
) ([]mdsmodel.MdsColonyTaskExecutionInfo, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"构建mds任务执行信息:入参详情",
		zap.Object("mds_colony_dto", &dto),
	)

	qp := database.QueryParams{
		Preloads: nil,
		OrderBy:  []string{"id DESC"},
		Query:    dto.ToQueryMap(),
	}

	ms, err := s.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mds集群列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, qp.Query)
	}

	trs := make([]mdsmodel.MdsColonyTaskRecordIDs, len(ms))
	for i, m := range ms {
		tr, err := loadMdsTaskRecordCacheFromFiles(log, m.ColonyNum)
		if err != nil {
			log.Error(
				fmt.Sprintf("读取mds任务状态对应的执行记录id:获取%s任务状态失败", m.ColonyNum),
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
			"查询mds任务执行记录:查询数据库模型失败",
			zap.Error(rErr),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	cache := make(map[uint32]jobmodel.ScriptRecordModel, len(records))
	for _, r := range records {
		cache[r.ID] = r
	}

	tasks := make([]mdsmodel.MdsColonyTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		if ctx.Err() != nil {
			return nil, errors.FromError(ctx.Err())
		}
		tasks[i] = mdsmodel.MdsColonyTaskExecutionInfo{
			ColonyNum: tr.ColonyNum,
			Mon:       jobsvc.GetRecordIDByMap(cache, tr.Mon),
			Bse:       jobsvc.GetRecordIDByMap(cache, tr.Bse),
			Sse:       jobsvc.GetRecordIDByMap(cache, tr.Sse),
			Szse:      jobsvc.GetRecordIDByMap(cache, tr.Szse),
		}
	}
	return tasks, nil
}

func loadMdsTaskRecordCacheFromFiles(
	log *zap.Logger,
	colonyNum string,
) (*mdsmodel.MdsColonyTaskRecordIDs, *errors.Error) {
	startTime := time.Now()

	log.Debug(
		"读取mds任务状态对应的执行记录id:开始执行",
		zap.String("colony_num", colonyNum),
	)

	flagDir := filepath.Join(config.StorageDir, "mds", "flags", colonyNum)
	mc := mdsmodel.MdsColonyTaskRecordIDs{
		ColonyNum: colonyNum,
	}

	// 定义任务映射表，减少代码重复
	taskMap := map[string]*uint32{
		"mon":  &mc.Mon,
		"bse":  &mc.Bse,
		"sse":  &mc.Sse,
		"szse": &mc.Szse,
	}

	// 遍历处理每个任务
	for taskName, fieldPtr := range taskMap {
		flagPath := filepath.Join(flagDir, "."+taskName)
		if value, err := common.ReadUint32FromFile(flagPath); err != nil {
			log.Error(
				"读取mds任务状态对应的执行记录id:获取"+taskName+"任务状态失败",
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
		"读取mds任务状态对应的执行记录id:任务状态读取成功",
		zap.Object("mds_task_record_ids", &mc),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &mc, nil
}
