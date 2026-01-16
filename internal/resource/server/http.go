package server

import (
	"encoding/base64"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/internal/resource/data"
	"gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) {
	pubKeys := make([]string, len(init.Signers))
	for i, signer := range init.Signers {
		pubKeyBytes := ssh.MarshalAuthorizedKey(signer.PublicKey())
		pubKeys[i] = base64.StdEncoding.EncodeToString(pubKeyBytes)
	}

	sshTimeout := time.Duration(init.Conf.SSH.Timeout) * time.Second
	hostRepo := data.NewHostRepo(loggers.Data, init.DB, init.DBTimeout)
	pkgRepo := data.NewpackageRepo(loggers.Data, init.DB, init.DBTimeout)

	hostUsecase := biz.NewHostUsecase(loggers.Biz, hostRepo, sshTimeout, ssh.PublicKeys(init.Signers...), pubKeys)
	pkgUsecase := biz.NewPackageUsecase(loggers.Biz, pkgRepo)

	hostService := service.NewHostService(loggers.Service, hostUsecase)
	pkgService := service.NewPackageService(loggers.Service, pkgUsecase, int64(init.Conf.Security.Upload.MaxPkgSize)*1024*1024)

	appRouter := router.Group("/v1/resource")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	hostService.LoadRouter(appRouter)
	pkgService.LoadRouter(appRouter)
}
