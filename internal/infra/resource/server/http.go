package server

import (
	"encoding/base64"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"gin-artweb/internal/infra/resource/biz"
	"gin-artweb/internal/infra/resource/data"
	"gin-artweb/internal/infra/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/internal/shared/shell"
)

type ResourceUsecase struct {
	Host *biz.HostUsecase
	Pkg  *biz.PackageUsecase
}

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) *ResourceUsecase {
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
	hostRepo := data.NewHostRepo(loggers.Data, init.DB, init.DBTimeout)
	pkgRepo := data.NewPackageRepo(loggers.Data, init.DB, init.DBTimeout)

	hostUsecase := biz.NewHostUsecase(loggers.Biz, hostRepo, sshTimeout, ssh.PublicKeys(signers...), pubKeys)
	pkgUsecase := biz.NewPackageUsecase(loggers.Biz, pkgRepo)

	hostService := service.NewHostService(loggers.Service, hostUsecase)
	pkgService := service.NewPackageService(loggers.Service, pkgUsecase, int64(init.Conf.Security.Upload.MaxPkgSize)*1024*1024)

	appRouter := router.Group("/v1/resource")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	hostService.LoadRouter(appRouter)
	pkgService.LoadRouter(appRouter)

	return &ResourceUsecase{
		Host: hostUsecase,
		Pkg:  pkgUsecase,
	}
}
