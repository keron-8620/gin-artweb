package service

import (
	"go.uber.org/zap"

	"gin-artweb/internal/jobs/biz"
)

type ScheduleService struct {
	log        *zap.Logger
	ucSchedule *biz.ScheduleUsecase
	maxSize    int64
}

func NewScheduleService(
	logger *zap.Logger,
	ucSchedule *biz.ScheduleUsecase,
	maxSize int64,
) *ScheduleService {
	return &ScheduleService{
		log:        logger,
		ucSchedule: ucSchedule,
		maxSize:    maxSize,
	}
}
