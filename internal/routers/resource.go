package routers

import (
	"encoding/base64"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	handler "gin-artweb/internal/handler/resource"
	resorepo "gin-artweb/internal/repo/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/internal/shared/shell"
)

type ResourceServices struct {
	Host     *resosvc.HostService
	Pkg      *resosvc.PackageService
	Terminal *resosvc.TerminalService
}

func newResourceRouter(
	router *gin.RouterGroup,
	init *config.SystemInit,
	loggers *config.Loggers,
) *ResourceServices {
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

	sshTimeout := init.Conf.SSH.Timeout
	hostRepo := resorepo.NewHostRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	pkgRepo := resorepo.NewPackageRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)

	hostService := resosvc.NewHostService(loggers.Service, hostRepo, sshTimeout, ssh.PublicKeys(signers...), pubKeys)
	pkgService := resosvc.NewPackageService(loggers.Service, pkgRepo, filepath.Join(config.StorageDir, "packages"))
	terminalService := resosvc.NewTerminalService(loggers.Service, hostRepo, []ssh.AuthMethod{ssh.PublicKeys(signers...)})

	hostHandler := handler.NewHostHandler(loggers.Handler, hostService)
	pkgHandler := handler.NewPackageHandler(loggers.Handler, pkgService, int64(init.Conf.Upload.MaxPkgSize)*1024*1024)
	terminalHandler := handler.NewTerminalHandler(loggers.Handler, terminalService)

	appRouter := router.Group("/v1/resource")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	hostHandler.LoadRouter(appRouter)
	pkgHandler.LoadRouter(appRouter)

	wsRouter := router.Group("/v1/ws")
	wsRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	wsRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))
	wsRouter.GET("/terminal", terminalHandler.HandleWebSocket)

	return &ResourceServices{
		Host:     hostService,
		Pkg:      pkgService,
		Terminal: terminalService,
	}
}
