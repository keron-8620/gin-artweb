package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizJobs "gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type MdsTaskRecordCache struct {
	ColonyNum string
	Mon       uint32
	Bse       uint32
	Sse       uint32
	Szse      uint32
}

func (mc MdsTaskRecordCache) GetTaskList() []string {
	return []string{"mon", "bse", "sse", "szse"}
}

func (mc MdsTaskRecordCache) GetRecordIDs() []uint32 {
	return []uint32{mc.Mon, mc.Bse, mc.Sse, mc.Szse}
}

func (mc MdsTaskRecordCache) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for i, task := range mc.GetTaskList() {
		enc.AddUint32(task, mc.GetRecordIDs()[i])
	}
	return nil
}

type MdsTaskExecutionInfo struct {
	ColonyNum string
	Mon       *bizJobs.ScriptRecordModel
	Bse       *bizJobs.ScriptRecordModel
	Sse       *bizJobs.ScriptRecordModel
	Szse      *bizJobs.ScriptRecordModel
}

func (m MdsTaskExecutionInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("colony_num", m.ColonyNum)
	if m.Mon != nil {
		enc.AddObject("mon", m.Mon)
	}
	if m.Bse != nil {
		enc.AddObject("bse", m.Bse)
	}
	if m.Sse != nil {
		enc.AddObject("sse", m.Sse)
	}
	if m.Szse != nil {
		enc.AddObject("szse", m.Szse)
	}
	return nil
}

type MdsTaskExecutionInfoUsecase struct {
	log      *zap.Logger
	ucRecord *RecordUsecase
}

func NewMdsTaskExecutionInfoUsecase(
	log *zap.Logger,
	ucRecord *RecordUsecase,
) *MdsTaskExecutionInfoUsecase {
	return &MdsTaskExecutionInfoUsecase{
		log:      log,
		ucRecord: ucRecord,
	}
}

func (uc *MdsTaskExecutionInfoUsecase) BuildTaskExecutionInfos(
	ctx context.Context,
	ms []MdsColonyModel,
) (*[]MdsTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	trs := make([]MdsTaskRecordCache, len(ms))
	for i, m := range ms {
		tr, err := uc.LoadMdsTaskRecordCacheFromFiles(ctx, m.ColonyNum)
		if err != nil {
			return nil, errors.FromError(err)
		}
		if tr != nil {
			trs[i] = *tr
		}
	}
	recoids, rErr := uc.ExtractValidRecordIDsFromCaches(ctx, trs)
	if rErr != nil {
		return nil, rErr
	}
	cache, rErr := uc.ucRecord.FindRecordsByIDs(ctx, recoids)
	if rErr != nil {
		return nil, rErr
	}
	tasks := make([]MdsTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		info, err := uc.BuildTaskExecutionInfo(ctx, tr, cache)
		if err != nil {
			return nil, err
		}
		tasks[i] = info
	}
	return &tasks, nil
}

func (uc *MdsTaskExecutionInfoUsecase) BuildTaskExecutionInfo(
	ctx context.Context,
	tr MdsTaskRecordCache,
	cache map[uint32]bizJobs.ScriptRecordModel,
) (MdsTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return MdsTaskExecutionInfo{}, errors.FromError(err)
	}
	return MdsTaskExecutionInfo{
		ColonyNum: tr.ColonyNum,
		Mon:       uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Mon, tr.ColonyNum, "mon"),
		Bse:       uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Bse, tr.ColonyNum, "bse"),
		Sse:       uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Sse, tr.ColonyNum, "sse"),
		Szse:      uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Szse, tr.ColonyNum, "szse"),
	}, nil
}

func (uc *MdsTaskExecutionInfoUsecase) ExtractValidRecordIDsFromCaches(
	ctx context.Context,
	trs []MdsTaskRecordCache,
) ([]uint32, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	var recordIDs []uint32
	for _, tr := range trs {
		if tr.Mon != 0 {
			recordIDs = append(recordIDs, tr.Mon)
		}
		if tr.Bse != 0 {
			recordIDs = append(recordIDs, tr.Bse)
		}
		if tr.Sse != 0 {
			recordIDs = append(recordIDs, tr.Sse)
		}
		if tr.Szse != 0 {
			recordIDs = append(recordIDs, tr.Szse)
		}
	}
	if len(recordIDs) == 0 {
		return []uint32{}, nil
	}
	return recordIDs, nil
}

func (uc *MdsTaskExecutionInfoUsecase) LoadMdsTaskRecordCacheFromFiles(
	ctx context.Context,
	colonyNum string,
) (*MdsTaskRecordCache, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	flagDir := filepath.Join(config.StorageDir, "mds", "flags", colonyNum)
	mc := MdsTaskRecordCache{
		ColonyNum: colonyNum,
		Mon:       uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".mon")),
		Bse:       uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".bse")),
		Sse:       uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".sse")),
		Szse:      uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".szse")),
	}
	uc.log.Debug(
		"查询mds任务状态对应的执行记录id成功",
		zap.Object("mds_task_record_ids", &mc),
	)
	return &mc, nil
}
