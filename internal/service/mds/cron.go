package mds

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repo/mds"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MdsCronService struct {
	log         *zap.Logger
	scheduleSvc *jobsvc.ScheduleService
	cronRepo    *mdsrepo.MdsCronRepo
	cronConf    map[string]mdsmodel.MdsCronTask
}

func NewMdsCronService(
	log *zap.Logger,
	scheduleSvc *jobsvc.ScheduleService,
	cronRepo *mdsrepo.MdsCronRepo,
	cronConf map[string]mdsmodel.MdsCronTask,
) *MdsCronService {
	return &MdsCronService{
		log:         log,
		scheduleSvc: scheduleSvc,
		cronRepo:    cronRepo,
		cronConf:    cronConf,
	}
}

func (s *MdsCronService) CreateCornByColony(
	ctx context.Context,
	m mdsmodel.MdsColonyModel,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"注册mds计划任务:开始执行",
		zap.Object("mds_colony_model", &m),
	)

	dtos := make([]jobmodel.ScheduleUpsertDTO, 0, len(s.cronConf))
	for name, task := range s.cronConf {
		dto := jobmodel.ScheduleUpsertDTO{
			Name:          fmt.Sprintf("mds_%s_%s", m.ColonyNum, name),
			Specification: task.Specification,
			IsEnabled:     m.IsEnable,
			EnvVars:       "{}",
			CommandArgs:   m.ColonyNum,
			WorkDir:       "",
			Timeout:       3600,
			IsRetry:       true,
			RetryInterval: 300,
			MaxRetries:    3,
			CreateType:    1,
			ScriptID:      task.ScriptID,
		}
		dtos = append(dtos, dto)
	}

	schedules, err := s.scheduleSvc.CreateSchedules(ctx, dtos)
	if err != nil {
		log.Error(
			"注册mds计划任务:执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", m.ID),
			zap.String("colony_num", m.ColonyNum),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	var scheduleIDs []uint32
	for _, schedule := range schedules {
		scheduleIDs = append(scheduleIDs, schedule.ID)
	}

	clearSchedules := func() {
		if err := s.scheduleSvc.DeleteScheduleByIDs(ctx, scheduleIDs); err != nil {
			log.Error(
				"注册mds计划任务:删除失败",
				zap.Error(err),
				zap.Uint32("mds_colony_id", m.ID),
				zap.String("colony_num", m.ColonyNum),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	mdsCronModels := make([]mdsmodel.MdsCronModel, 0, len(schedules))
	for _, schedule := range schedules {
		mdsCronModels = append(mdsCronModels, mdsmodel.MdsCronModel{
			MdsColonyID: m.ID,
			ScheduleID:  schedule.ID,
		})
	}

	if err := s.cronRepo.CreateModels(ctx, mdsCronModels); err != nil {
		log.Error(
			"注册mds计划任务:创建失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", m.ID),
			zap.String("colony_num", m.ColonyNum),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		clearSchedules()
		return errors.FromError(err)
	}

	log.Info(
		"注册mds计划任务:执行成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MdsCronService) DeleteCornByColonyID(
	ctx context.Context,
	colonyID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mds计划任务:开始执行",
		zap.Uint32("mds_colony_id", colonyID),
	)

	qp := database.QueryParams{
		Query: map[string]any{
			"mds_colony_id = ?": colonyID,
		},
	}

	crons, err := s.cronRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"删除mds计划任务:查询mds集群计划任务失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"mds_colony_id = ?": colonyID})
	}

	if len(crons) == 0 {
		log.Warn(
			"删除mds计划任务:未查询到计划任务",
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	var scheduleIDs []uint32
	for _, cron := range crons {
		scheduleIDs = append(scheduleIDs, cron.ScheduleID)
	}

	if err := s.scheduleSvc.DeleteScheduleByIDs(ctx, scheduleIDs); err != nil {
		log.Error(
			"删除mds计划任务:删除失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	return nil
}

func (s *MdsCronService) ListCornByColonyID(
	ctx context.Context,
	colonyID uint32,
) ([]jobmodel.ScheduleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询mds计划任务:开始执行",
		zap.Uint32("mds_colony_id", colonyID),
	)

	qp := database.QueryParams{
		Query: map[string]any{
			"mds_colony_id = ?": colonyID,
		},
		OrderBy: []string{"id ASC"},
	}

	mdsCrons, err := s.cronRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mds计划任务:获取关联的计划任务id失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"mds_colony_id = ?": colonyID})
	}

	scheduleIDs := make([]string, 0, len(mdsCrons))
	for _, item := range mdsCrons {
		scheduleIDs = append(scheduleIDs, strconv.FormatUint(uint64(item.ScheduleID), 10))
	}

	if len(scheduleIDs) == 0 {
		return nil, nil
	}

	dto := jobmodel.ListScheduleDTO{
		StandardModelQuery: common.StandardModelQuery{
			BaseModelQuery: common.BaseModelQuery{
				Page: 1,
				Size: len(scheduleIDs),
				IDs:  strings.Join(scheduleIDs, ","),
			},
		},
	}

	_, _, _, schedules, rErr := s.scheduleSvc.ListSchedule(ctx, dto)
	if rErr != nil {
		log.Error(
			"查询mds计划任务:获取关联计划任务模型失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("total_duration", time.Since(startTime)),
			zap.Object("list_schedule_dto", &dto),
		)
	}
	return schedules, rErr
}
