package routers

import (
	"encoding/base64"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	handler "gin-artweb/internal/handler/resource"
	resorepo "gin-artweb/internal/repo/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/internal/shared/shell"
)

type ResourceRouter struct {
	Host *resosvc.HostService
	Pkg  *resosvc.PackageService
}

func newResourceRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *common.Loggers,
) *ResourceRouter {
	signers, err := shell.GetSignersFromDefaultKeys()
	if err != nil {
		loggers.Server.Error("初始化加载ssh密钥失败", zap.Error(err))
		panic("初始化加载ssh密钥失败")
	}
	if len(signers) == 0 {
		loggers.Server.Error("没有可用的SSH密钥")
		panic("没有可用的SSH密钥")
	}
	pubKeys := make([]string, len(signers))
	for i, signer := range signers {
		pubKeyBytes := ssh.MarshalAuthorizedKey(signer.PublicKey())
		pubKeys[i] = base64.StdEncoding.EncodeToString(pubKeyBytes)
	}

	sshTimeout := time.Duration(init.Conf.SSH.Timeout) * time.Second
	hostRepo := resorepo.NewHostRepo(loggers.Data, init.DB, init.DBTimeout)
	pkgRepo := resorepo.NewPackageRepo(loggers.Data, init.DB, init.DBTimeout)

	hostService := resosvc.NewHostService(loggers.Service, hostRepo, sshTimeout, ssh.PublicKeys(signers...), pubKeys)
	pkgService := resosvc.NewPackageService(loggers.Service, pkgRepo)

	hostHandler := handler.NewHostHandler(loggers.Handler, hostService)
	pkgHandler := handler.NewPackageHandler(loggers.Handler, pkgService, int64(init.Conf.Upload.MaxPkgSize)*1024*1024)

	appRouter := router.Group("/v1/resource")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	hostHandler.LoadRouter(appRouter)
	pkgHandler.LoadRouter(appRouter)

	return &ResourceRouter{
		Host: hostService,
		Pkg:  pkgService,
	}
}
