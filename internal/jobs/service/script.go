package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbScript "gin-artweb/api/jobs/script"
	"gin-artweb/internal/jobs/biz"
)

type ScriptService struct {
	log      *zap.Logger
	ucScript *biz.ScriptUsecase
	maxSize  int64
}

func NewScriptService(
	logger *zap.Logger,
	ucScript *biz.ScriptUsecase,
	maxSize int64,
) *ScriptService {
	return &ScriptService{
		log:      logger,
		ucScript: ucScript,
		maxSize:  maxSize,
	}
}

func (s *ScriptService) CreateScript(ctx *gin.Context) {

}

func (s *ScriptService) UpdateScript(ctx *gin.Context) {

}

func (s *ScriptService) DeleteScript(ctx *gin.Context) {

}

func (s *ScriptService) GetScript(ctx *gin.Context) {

}

func (s *ScriptService) ListScript(ctx *gin.Context) {

}

func (s *ScriptService) DownloadScript(ctx *gin.Context) {

}

func (s *ScriptService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/script", s.CreateScript)
	r.PUT("/script/:pk", s.UpdateScript)
	r.DELETE("/script/:pk", s.DeleteScript)
	r.GET("/script/:pk", s.GetScript)
	r.GET("/script", s.ListScript)
	r.GET("/script/:pk/download", s.DownloadScript)
}

func ScriptModelToOutBase(
	m biz.ScriptModel,
) *pbScript.ScriptOutBase {
	return &pbScript.ScriptOutBase{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Name:      m.Name,
		Descr:     m.Descr,
		Project:   m.Project,
		Label:     m.Label,
		Language:  m.Language,
		Status:    m.Status,
		IsBuiltin: m.IsBuiltin,
	}
}

func ListScriptModelToOutBase(
	pms *[]biz.ScriptModel,
) *[]pbScript.ScriptOutBase {
	if pms == nil {
		return &[]pbScript.ScriptOutBase{}
	}

	ms := *pms
	mso := make([]pbScript.ScriptOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScriptModelToOutBase(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
