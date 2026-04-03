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
	m *mdsmodel.MdsColonyModel,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"注册mds计划任务:开始执行",
		zap.Uint32("mds_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
	)

	log.Debug(
		"注册mds计划任务:入参详情",
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
				fmt.Sprintf("注册mds计划任务:%s注册失败", name),
				zap.Error(err),
				zap.Object("schedule_upsert_dto", &dto),
			)
			return errors.FromError(err)
		}
		scheduleIDs = append(scheduleIDs, schedule.ID)
		log.Debug(
			fmt.Sprintf("注册mds计划任务:%s注册成功", name),
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
				"注册mds计划任务:绑定mds集群与计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", schedule.ID),
				zap.Uint32("mds_colony_id", m.ID),
				zap.Duration("create_mds_corn_duration", time.Since(createMdsCornStart)),
			)
			return errors.FromError(err)
		}
		log.Debug(
			fmt.Sprintf("注册mds计划任务:%s绑定成功", name),
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

	listMdsCornStart := time.Now()
	log.Debug(
		"删除mds计划任务:查询mds集群计划任务",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Any("query_params", qp),
	)
	crons, err := s.cronRepo.ListModel(ctx, qp)
	listMdsCornDuration := time.Since(listMdsCornStart)
	if err != nil {
		log.Error(
			"删除mds计划任务:查询mds集群计划任务失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("list_mds_corn_duration", listMdsCornDuration),
		)
		return errors.NewGormError(err, map[string]any{"mds_colony_id = ?": colonyID})
	}
	if len(crons) == 0 {
		log.Debug(
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

	clearStep := time.Now()
	log.Debug(
		"删除mds计划任务:开始删除",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	if err := s.scheduleSvc.DeleteScheduleByIDs(ctx, scheduleIDs); err != nil {
		log.Error(
			"删除mds计划任务:删除失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("clear_step_duration", time.Since(clearStep)),
		)
		return err
	}
	log.Debug(
		"删除mds计划任务:删除成功",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Duration("clear_step_duration", time.Since(clearStep)),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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

	listMdsCornStart := time.Now()
	log.Debug(
		"查询mds计划任务:开始获取关联的计划任务id",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Any("query_params", qp),
	)
	mdsCrons, err := s.cronRepo.ListModel(ctx, qp)
	listMdsCornDuration := time.Since(listMdsCornStart)
	if err != nil {
		log.Error(
			"查询mds计划任务:获取关联的计划任务id失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("list_mds_corn_duration", listMdsCornDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"mds_colony_id = ?": colonyID})
	}
	scheduleIDs := make([]string, 0, len(mdsCrons))
	for _, item := range mdsCrons {
		scheduleIDs = append(scheduleIDs, strconv.FormatUint(uint64(item.ScheduleID), 10))
	}
	log.Debug(
		"查询mds计划任务:获取关联的计划任务id成功",
		zap.Strings("schedule_ids", scheduleIDs),
	)
	if len(scheduleIDs) == 0 {
		return nil, nil
	}
	listScheduleStart := time.Now()
	dto := jobmodel.ListScheduleDTO{
		StandardModelQuery: common.StandardModelQuery{
			BaseModelQuery: common.BaseModelQuery{
				IDs: strings.Join(scheduleIDs, ","),
			},
		},
	}
	log.Debug(
		"查询mds计划任务:开始获取关联计划任务模型",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Object("list_schedule_dto", &dto),
	)
	_, schedules, rErr := s.scheduleSvc.ListSchedule(ctx, 1, len(scheduleIDs), dto)
	listScheduleDuration := time.Since(listScheduleStart)
	if rErr != nil {
		log.Error(
			"查询mds计划任务:获取关联计划任务模型失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", colonyID),
			zap.Duration("list_schedule_duration", listScheduleDuration),
			zap.Object("list_schedule_dto", &dto),
		)
	}
	log.Debug(
		"查询mds计划任务:获取关联计划任务模型成功",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Strings("schedule_ids", scheduleIDs),
		zap.Duration("list_schedule_duration", listScheduleDuration),
	)

	log.Debug(
		"查询mds计划任务:执行成功",
		zap.Uint32("mds_colony_id", colonyID),
		zap.Duration("list_mds_corn_duration", listMdsCornDuration),
		zap.Duration("list_schedule_duration", listScheduleDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return schedules, rErr
}
