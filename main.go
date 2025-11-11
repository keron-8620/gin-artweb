package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	customer "gin-artweb/internal/customer/server"

	"gin-artweb/docs"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/config"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/log"
	"gin-artweb/pkg/middleware"
)

const version = "v0.17.6.3.1"

const (
	serverLogName   = "server.log"
	accessLogName   = "access.log"
	databaseLogName = "database.log"
	serviceLogName  = "service.log"
)

type initialize struct {
	conf *config.SystemConf
	log  *zap.Logger
	db   *gorm.DB
}

// newInitialize 初始化系统组件
// path: 配置文件路径
// 返回值1: 初始化结构体指针，包含配置、数据库、缓存和日志组件
// 返回值2: 清理函数，用于关闭数据库连接
// 返回值3: 初始化过程中发生的错误
func newInitialize(path string) (*initialize, func(), error) {
	// 加载系统配置
	conf := config.NewSystemConf(path)

	// 初始化服务器日志记录器
	write := log.NewLumLogger(conf.Log, filepath.Join(common.LogDir, serverLogName))
	logger := log.NewZapLoggerMust(conf.Log.Level, write)

	// 创建GORM数据库配置并连接数据库
	dbWrite := log.NewLumLogger(conf.Log, filepath.Join(common.LogDir, databaseLogName))
	dbConf := database.NewGormConfig(dbWrite)
	db, err := database.NewGormDB(conf.Database.Type, conf.Database.Dns, dbConf)
	if err != nil {
		logger.Error("数据库连接失败", zap.Error(err))
		return nil, nil, err
	}

	// 返回初始化结构体和清理函数
	return &initialize{
			conf: conf,
			db:   db,
			log:  logger,
		}, func() {
			// 关闭数据库连接
			conn, err := db.DB()
			if err != nil {
				logger.Error(err.Error())
				panic(err)
			}
			if err = conn.Close(); err != nil {
				logger.Error(err.Error())
				panic(err)
			}
			logger.Info("资源释放成功")
		}, nil
}

// main 是程序的入口函数，负责初始化配置、数据库连接、日志系统，
// 并启动 HTTP 或 HTTPS 服务。同时监听系统信号以实现优雅停机。
func main() {
	// 定义并解析命令行参数，指定配置文件路径，默认为 "../config/system.yaml"
	var configPath string
	flag.StringVar(&configPath, "config", "../config/system.yaml", "Path to config file")
	flag.Parse()

	// 初始化系统资源（如配置、数据库等），获取清理函数和错误信息
	i, clearFunc, err := newInitialize(configPath)
	if err != nil {
		panic(err)
	}
	defer clearFunc() // 程序结束前执行资源清理操作

	// 设置 Gin 框架的日志输出到 Zap 日志中，并设置运行模式为 ReleaseMode
	// gin.DefaultWriter = zap.NewStdLog(i.log).Writer()
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	// 创建 Gin 路由引擎
	r := newRouter(i)

	// 构建 HTTP 服务器结构体
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", i.conf.Server.Host, i.conf.Server.Port),
		Handler: r,
	}

	// 启动一个 goroutine 来异步启动 HTTP/HTTPS 服务
	go func() {
		var err error
		// 判断是否启用 SSL/TLS 加密传输
		if i.conf.Server.SSL.Enable {
			// 构造证书和私钥的完整路径
			crtPath := filepath.Join(common.ConfigDir, i.conf.Server.SSL.CrtPath)
			keyPath := filepath.Join(common.ConfigDir, i.conf.Server.SSL.KeyPath)

			// 校验证书文件是否存在
			if _, statErr := os.Stat(crtPath); os.IsNotExist(statErr) {
				i.log.Fatal("SSL CRT 文件不存在", zap.String("path", crtPath))
			}
			// 校验私钥文件是否存在
			if _, statErr := os.Stat(keyPath); os.IsNotExist(statErr) {
				i.log.Fatal("SSL KEY 文件不存在", zap.String("path", keyPath))
			}

			// 输出 HTTPS 启动信息并开始监听
			i.log.Info("正在启动 HTTPS 服务器...",
				zap.String("addr", srv.Addr),
				zap.String("crt", crtPath),
				zap.String("key", keyPath))
			err = srv.ListenAndServeTLS(crtPath, keyPath)
		} else {
			// 输出 HTTP 启动信息并开始监听
			i.log.Info("正在启动 HTTP 服务器...", zap.String("addr", srv.Addr))
			err = srv.ListenAndServe()
		}

		// 处理服务器启动过程中的致命错误
		if err != nil && err != http.ErrServerClosed {
			i.log.Error("服务器启动失败", zap.Error(err))
			panic(err)
		}
	}()

	// 打印服务器启动信息
	i.log.Info("服务器启动ing ...",
		zap.String("host", i.conf.Server.Host),
		zap.Int("port", i.conf.Server.Port),
		zap.Bool("ssl", i.conf.Server.SSL.Enable))

	// 监听系统中断信号（SIGINT, SIGTERM）来触发优雅关闭流程
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞等待信号到来

	// 收到关闭信号后打印提示信息
	i.log.Info("正在关闭服务器...")

	// 创建带超时控制的上下文对象用于通知服务器关闭
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(i.conf.Security.TimeoutShutdown)*time.Second)
	defer cancel()

	// 执行服务器优雅关闭逻辑
	if err := srv.Shutdown(ctx); err != nil {
		i.log.Error("服务器强制关闭", zap.Error(err))
	}

	// 最终确认服务器已经退出
	i.log.Info("服务器已退出")
}

func newRouter(init *initialize) *gin.Engine {
	wrire := log.NewLumLogger(init.conf.Log, filepath.Join(common.LogDir, serviceLogName))
	// 创建 Zap 日志记录器用于服务日志输出
	logger := log.NewZapLoggerMust(init.conf.Log.Level, wrire)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(common.BaseDir, "html", "index.html"))
	})
	r.Static("/static", filepath.Join(common.BaseDir, "html", "static"))

	// 设置 Swagger 文档信息
	docs.SwaggerInfo.Title = "gin-art-web"
	docs.SwaggerInfo.Description = "art-web自动化运维平台"
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", init.conf.Server.Host, init.conf.Server.Port)
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册访问日志中间件
	accWriet := log.NewLumLogger(init.conf.Log, filepath.Join(common.LogDir, accessLogName))
	r.Use(gin.LoggerWithWriter(accWriet))

	// 注册统一异常处理中间件
	r.Use(middleware.ErrorMiddleware(logger))

	// 注册跨域请求处理中间件
	r.Use(middleware.CorsMiddleware(init.conf.CORS))

	apiRouter := r.Group("/api")

	// 初始化加载业务模块
	customer.NewServer(apiRouter, init.conf, init.db, logger)
	return r
}
