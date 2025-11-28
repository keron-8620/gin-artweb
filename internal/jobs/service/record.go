package service

import (
	"go.uber.org/zap"

	"gin-artweb/internal/jobs/biz"
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
