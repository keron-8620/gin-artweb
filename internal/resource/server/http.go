package server

import (
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/internal/resource/data"
	"gin-artweb/internal/resource/service"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/config"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/log"
)

func NewServer(
	router *gin.RouterGroup,
	conf *config.SystemConf,
	db *gorm.DB,
	dbTimeout *database.DBTimeout,
	loggers *log.Loggers,
) {
	if err := dbAutoMigrate(db, loggers.Data); err != nil {
		panic(err)
	}

	sshTimeout := time.Duration(conf.SSH.ConnectTimeout) * time.Second
	signer, err := NewSigner(loggers.Data, conf.SSH.PrivateKeyPath, sshTimeout)
	if err != nil {
		panic(err)
	}
	hostRepo := data.NewHostRepo(loggers.Data, db, dbTimeout)
	pkgRepo := data.NewpackageRepo(loggers.Data, db, dbTimeout)

	hostDir := filepath.Join(common.StorageDir, "host")
	pkgDir := filepath.Join(common.StorageDir, "package")
	hostUsecase := biz.NewHostUsecase(loggers.Biz, hostRepo, signer, sshTimeout, hostDir)
	pkgUsecase := biz.NewPackageUsecase(loggers.Biz, pkgRepo, pkgDir)

	hostService := service.NewHostService(loggers.Service, hostUsecase)
	pkgService := service.NewPackageService(loggers.Service, pkgUsecase, int64(conf.Security.Upload.MaxFileSize))

	appRouter := router.Group("/v1/resource")
	hostService.LoadRouter(appRouter)
	pkgService.LoadRouter(appRouter)
}
