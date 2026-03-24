package mds

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

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
	m *mdsmodel.MdsColonyModel,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"注册mds计划任务：开始执行",
		zap.Uint32("mds_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
	)

	log.Debug(
		"注册mds计划任务：入参详情",
		zap.Object("mds_colony_model", m),
	)

	dto := jobmodel.ScheduleUpsertDTO{
		IsEnabled:     m.IsEnable,
		EnvVars:       "{}",
		CommandArgs:   m.ColonyNum,
		WorkDir:       "",
		Timeout:       3600,
		IsRetry:       true,
		RetryInterval: 300,
		MaxRetries:    3,
		CreateType:    1,
	}

	var scheduleIDs []uint32
	for name, task := range s.cronConf {
		dto.Name = fmt.Sprintf("mds_%s_%s", m.ColonyNum, name)
		dto.ScriptID = task.ScriptID
		dto.Specification = task.Specification
		schedule, err := s.scheduleSvc.CreateSchedule(ctx, dto)
		if err != nil {
			log.Error(
				fmt.Sprintf("注册mds计划任务：%s注册失败", name),
				zap.Error(err),
				zap.Object("schedule_upsert_dto", &dto),
			)
			return errors.FromError(err)
		}
		scheduleIDs = append(scheduleIDs, schedule.ID)
		log.Debug(
			fmt.Sprintf("注册mds计划任务：%s注册成功", name),
			zap.String("name", name),
			zap.Uint32("script_id", task.ScriptID),
			zap.String("specification", task.Specification),
		)

		createMdsCornStart := time.Now()
		if err := s.cronRepo.CreateModel(ctx, &mdsmodel.MdsCronModel{
			MdsColonyID: m.ID,
			ScheduleID:  schedule.ID,
		}); err != nil {
			log.Error(
				"注册mds计划任务：绑定mds集群与计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", schedule.ID),
				zap.Uint32("mds_colony_id", m.ID),
				zap.Duration("create_mds_corn_duration", time.Since(createMdsCornStart)),
			)
			return errors.FromError(err)
		}
		log.Debug(
			fmt.Sprintf("注册mds计划任务：%s绑定成功", name),
			zap.Uint32("schedule_id", schedule.ID),
			zap.Uint32("mds_colony_id", m.ID),
			zap.Duration("create_mds_corn_duration", time.Since(createMdsCornStart)),
		)
	}

	log.Info(
		"注册mds计划任务成功",
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
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"删除mds计划任务：开始执行",
		zap.Uint32("mds_colony_id", colonyID),
	)

	qp := database.QueryParams{
		Query: map[string]any{
			"mds_colony_id = ?": colonyID,
		},
	}

	listMdsCornStart := time.Now()
	log.Debug(
		"删除mds计划任务：查询mds集群计划任务",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Any("query_params", qp),
	)
	crons, err := s.cronRepo.ListModel(ctx, qp)
	listMdsCornDuration := time.Since(listMdsCornStart)
	if err != nil {
		log.Error(
			"删除mds计划任务：查询mds集群计划任务失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("list_mds_corn_duration", listMdsCornDuration),
		)
		return errors.NewGormError(err, map[string]any{"mds_colony_id = ?": colonyID})
	}
	if len(crons) > 0 {
		log.Debug(
			"删除mds计划任务：未查询到计划任务",
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	var scheduleIDs []uint32
	for _, cron := range crons {
		scheduleIDs = append(scheduleIDs, cron.ScheduleID)
	}

	clearStep := time.Now()
	log.Debug(
		"删除mds计划任务：开始删除",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	if err := s.scheduleSvc.DeleteScheduleByIDs(ctx, scheduleIDs); err != nil {
		log.Error(
			"删除mds计划任务：删除失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("clear_step_duration", time.Since(clearStep)),
		)
		return err
	}
	log.Debug(
		"删除mds计划任务：删除成功",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Duration("clear_step_duration", time.Since(clearStep)),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
