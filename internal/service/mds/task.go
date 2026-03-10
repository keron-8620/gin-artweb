package mds

import (
	"context"

	"go.uber.org/zap"

	jobsmodel "gin-artweb/internal/model/jobs"
	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repository/mds"
	jobsvc "gin-artweb/internal/service/jobs"
	"gin-artweb/internal/shared/errors"
)

type MdsCronService struct {
	log         *zap.Logger
	cronRepo    *mdsrepo.MdsCronRepo
	svcScript   *jobsvc.ScriptService
	svcSchedule *jobsvc.ScheduleService
}

func NewMdsCronService(
	log *zap.Logger,
	cronRepo *mdsrepo.MdsCronRepo,
	svcScript *jobsvc.ScriptService,
	svcSchedule *jobsvc.ScheduleService,
) *MdsCronService {
	return &MdsCronService{
		log:         log,
		cronRepo:    cronRepo,
		svcScript:   svcScript,
		svcSchedule: svcSchedule,
	}
}

func (s *MdsCronService) CreateMdsCron(
	ctx context.Context,
	cron mdsmodel.MdsCronModel,
	schedule jobsmodel.ScheduleModel,
) (*mdsmodel.MdsCronModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
}
