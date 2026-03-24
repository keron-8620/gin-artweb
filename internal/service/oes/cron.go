package oes

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repo/oes"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type OesCronService struct {
	log         *zap.Logger
	scheduleSvc *jobsvc.ScheduleService
	cronRepo    *oesrepo.OesCronRepo
	stkCronConf map[string]oesmodel.OesCronTask
	crdCronConf map[string]oesmodel.OesCronTask
	optCronConf map[string]oesmodel.OesCronTask
}

func NewOesCronService(
	log *zap.Logger,
	scheduleSvc *jobsvc.ScheduleService,
	cronRepo *oesrepo.OesCronRepo,
	stkCronConf map[string]oesmodel.OesCronTask,
	crdCronConf map[string]oesmodel.OesCronTask,
	optCronConf map[string]oesmodel.OesCronTask,
) *OesCronService {
	return &OesCronService{
		log:         log,
		scheduleSvc: scheduleSvc,
		cronRepo:    cronRepo,
		stkCronConf: stkCronConf,
		crdCronConf: crdCronConf,
		optCronConf: optCronConf,
	}
}

func (s *OesCronService) CreateCornByColony(
	ctx context.Context,
	m *oesmodel.OesColonyModel,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"注册oes计划任务：开始执行",
		zap.Uint32("oes_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
	)

	log.Debug(
		"注册oes计划任务：入参详情",
		zap.Object("oes_colony_model", m),
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
	var cronConf map[string]oesmodel.OesCronTask
	switch m.SystemType {
	case "stk":
		cronConf = s.stkCronConf
	case "crd":
		cronConf = s.crdCronConf
	case "opt":
		cronConf = s.optCronConf
	default:
		return errors.ErrUnknown
	}
	for name, task := range cronConf {
		dto.Name = fmt.Sprintf("oes_%s_%s", m.ColonyNum, name)
		dto.ScriptID = task.ScriptID
		dto.Specification = task.Specification
		schedule, err := s.scheduleSvc.CreateSchedule(ctx, dto)
		if err != nil {
			log.Error(
				fmt.Sprintf("注册oes计划任务：%s注册失败", name),
				zap.Error(err),
				zap.Object("schedule_upsert_dto", &dto),
			)
			return errors.FromError(err)
		}
		scheduleIDs = append(scheduleIDs, schedule.ID)
		log.Debug(
			fmt.Sprintf("注册oes计划任务：%s注册成功", name),
			zap.String("name", name),
			zap.Uint32("script_id", task.ScriptID),
			zap.String("specification", task.Specification),
		)

		createOesCornStart := time.Now()
		if err := s.cronRepo.CreateModel(ctx, &oesmodel.OesCronModel{
			OesColonyID: m.ID,
			ScheduleID:  schedule.ID,
		}); err != nil {
			log.Error(
				"注册oes计划任务：绑定oes集群与计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", schedule.ID),
				zap.Uint32("oes_colony_id", m.ID),
				zap.Duration("create_oes_corn_duration", time.Since(createOesCornStart)),
			)
			return errors.FromError(err)
		}
		log.Debug(
			fmt.Sprintf("注册oes计划任务：%s绑定成功", name),
			zap.Uint32("schedule_id", schedule.ID),
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("create_oes_corn_duration", time.Since(createOesCornStart)),
		)
	}

	log.Info(
		"注册oes计划任务：执行成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *OesCronService) DeleteCornByColonyID(
	ctx context.Context,
	colonyID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"删除oes计划任务：开始执行",
		zap.Uint32("oes_colony_id", colonyID),
	)

	qp := database.QueryParams{
		Query: map[string]any{
			"oes_colony_id = ?": colonyID,
		},
	}

	listOesCornStart := time.Now()
	log.Debug(
		"删除oes计划任务：查询oes集群计划任务",
		zap.Uint32("oes_colony_id", colonyID),
		zap.Any("query_params", qp),
	)
	crons, err := s.cronRepo.ListModel(ctx, qp)
	listOesCornDuration := time.Since(listOesCornStart)
	if err != nil {
		log.Error(
			"删除oes计划任务：查询oes集群计划任务失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", colonyID),
			zap.Duration("list_oes_corn_duration", listOesCornDuration),
		)
		return errors.NewGormError(err, map[string]any{"oes_colony_id = ?": colonyID})
	}
	if len(crons) > 0 {
		log.Debug(
			"删除oes计划任务：未查询到计划任务",
			zap.Uint32("oes_colony_id", colonyID),
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
		"删除oes计划任务：开始删除",
		zap.Uint32("oes_colony_id", colonyID),
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	if err := s.scheduleSvc.DeleteScheduleByIDs(ctx, scheduleIDs); err != nil {
		log.Error(
			"删除oes计划任务：删除失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", colonyID),
			zap.Duration("clear_step_duration", time.Since(clearStep)),
		)
		return err
	}
	log.Debug(
		"删除oes计划任务：删除成功",
		zap.Uint32("oes_colony_id", colonyID),
		zap.Duration("clear_step_duration", time.Since(clearStep)),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
