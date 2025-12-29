package main

import (
	"context"
	"flag"
	"fmt"
	golog "log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	customer "gin-artweb/internal/customer/server"
	jobs "gin-artweb/internal/jobs/server"
	mds "gin-artweb/internal/mds/server"
	mon "gin-artweb/internal/mon/server"
	oes "gin-artweb/internal/oes/server"
	resource "gin-artweb/internal/resource/server"

	"gin-artweb/docs"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/crontab"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

var (
	version   string
	commitID  string
	buildTime string
	goVersion string
	goOS      string
	goArch    string
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 定义并解析命令行参数，指定配置文件路径，默认为 "../config/system.yaml"
	var (
		configPath  string
		showVersion bool
	)
	flag.StringVar(&configPath, "config", filepath.Join(config.ConfigDir, "system.yaml"), "系统配置文件的路径")
	flag.BoolVar(&showVersion, "v", false, "展示版本信息")
	flag.Parse()

	if showVersion {
		fmt.Println("===== 版本信息 =====")
		fmt.Printf("版本号    : %s\n", version)
		fmt.Printf("提交ID    : %s\n", commitID)
		fmt.Printf("构建时间  : %s\n", buildTime)
		fmt.Printf("Go版本    : %s\n", goVersion)
		fmt.Printf("操作系统  : %s\n", goOS)
		fmt.Printf("系统架构  : %s\n", goArch)
		fmt.Println("====================")
		return
	}

	// 初始化系统资源（如配置、数据库等），获取清理函数和错误信息
	logger, i, clearFunc, err := newInitialize(configPath)
	if err != nil {
		panic(err)
	}
	defer clearFunc() // 程序结束前执行资源清理操作

	// 设置 Gin 框架的日志输出到 Zap 日志中，并设置运行模式为 ReleaseMode
	// gin.DefaultWriter = zap.NewStdLog(logger).Writer()
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	// 创建 Gin 路由引擎
	r := newRouter(i)

	// 启动定时任务
	if i.Crontab != nil {
		i.Crontab.Start()
	}

	// 构建 HTTP 服务器结构体
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", i.Conf.Server.Host, i.Conf.Server.Port),
		Handler: r,
	}

	// 启动一个 goroutine 来异步启动 HTTP/HTTPS 服务
	go func() {
		var err error
		// 判断是否启用 SSL/TLS 加密传输
		if i.Conf.Server.SSL.Enable {
			// 构造证书和私钥的完整路径
			crtPath := filepath.Join(config.ConfigDir, i.Conf.Server.SSL.CrtPath)
			keyPath := filepath.Join(config.ConfigDir, i.Conf.Server.SSL.KeyPath)

			// 校验证书文件是否存在
			if _, statErr := os.Stat(crtPath); os.IsNotExist(statErr) {
				logger.Fatal("SSL CRT 文件不存在", zap.String("path", crtPath))
			}
			// 校验私钥文件是否存在
			if _, statErr := os.Stat(keyPath); os.IsNotExist(statErr) {
				logger.Fatal("SSL KEY 文件不存在", zap.String("path", keyPath))
			}

			// 输出 HTTPS 启动信息并开始监听
			logger.Info("正在启动 HTTPS 服务器...",
				zap.String("addr", srv.Addr),
				zap.String("crt", crtPath),
				zap.String("key", keyPath))
			err = srv.ListenAndServeTLS(crtPath, keyPath)
		} else {
			// 输出 HTTP 启动信息并开始监听
			logger.Info("正在启动 HTTP 服务器...", zap.String("addr", srv.Addr))
			err = srv.ListenAndServe()
		}

		// 处理服务器启动过程中的致命错误
		if err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败", zap.Error(err))
			panic(err)
		}
	}()

	// 打印服务器启动信息
	logger.Info(
		"服务器启动ing ...",
		zap.String("host", i.Conf.Server.Host),
		zap.Int("port", i.Conf.Server.Port),
		zap.Bool("ssl", i.Conf.Server.SSL.Enable),
		zap.String("version", version),
		zap.String("commit", commitID),
		zap.String("build_time", buildTime),
	)

	// 监听系统中断信号（SIGINT, SIGTERM）来触发优雅关闭流程
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞等待信号到来

	// 收到关闭信号后打印提示信息
	logger.Info("正在关闭服务器...")

	// 创建带超时控制的上下文对象用于通知服务器关闭
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(i.Conf.Security.Timeout.ShutdownTimeout)*time.Second)
	defer cancel()

	// 执行服务器优雅关闭逻辑
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭", zap.Error(err))
	}

	// 最终确认服务器已经退出
	logger.Info("服务器已退出")
}

// newInitialize 初始化系统组件
// path: 配置文件路径
// 返回值1: 初始化结构体指针，包含配置、数据库、缓存和日志组件
// 返回值2: 清理函数，用于关闭数据库连接
// 返回值3: 初始化过程中发生的错误
func newInitialize(path string) (*zap.Logger, *common.Initialize, func(), error) {
	// 加载系统配置
	conf := config.NewSystemConf(path)

	// 初始化服务器日志记录器
	write := log.NewLumLogger(conf.Log, filepath.Join(config.LogDir, "server.log"))
	logger := log.NewZapLoggerMust(conf.Log.Level, write)

	// 创建GORM数据库配置并连接数据库
	var dbLog *golog.Logger
	if conf.Database.LogSQL {
		dbWrite := log.NewLumLogger(conf.Log, filepath.Join(config.LogDir, "database.log"))
		dbLog = golog.New(dbWrite, " ", golog.LstdFlags)
	}
	dbConf := database.NewGormConfig(dbLog)
	db, err := database.NewGormDB(conf.Database, dbConf)
	if err != nil {
		logger.Error("数据库连接失败", zap.Error(err))
		return nil, nil, nil, err
	}

	dbTimeout := config.DBTimeout{
		ListTimeout:  time.Duration(conf.Database.ListTimeout) * time.Second,
		ReadTimeout:  time.Duration(conf.Database.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.Database.WriteTimeout) * time.Second,
	}

	enf, err := auth.NewCasbinEnforcer()
	if err != nil {
		logger.Error("Casbin 初始化失败", zap.Error(err))
		return nil, nil, nil, err
	}

	cronWrite := log.NewLumLogger(conf.Log, filepath.Join(config.LogDir, "cron.log"))
	cronLogger := log.NewZapLoggerMust(conf.Log.Level, cronWrite)
	ct := crontab.NewCron(cronLogger)

	// 返回初始化结构体和清理函数
	return logger, &common.Initialize{
			Conf:      conf,
			DB:        db,
			DBTimeout: &dbTimeout,
			Enforcer:  enf,
			Crontab:   ct,
		}, func() {
			// 1. 关闭计划任务
			if ct != nil {
				cronLogger.Info("正在关闭计划任务...")
				shutdownTimeout := time.Duration(conf.Security.Timeout.ShutdownTimeout) * time.Second
				ctx := ct.Stop() // Stop 返回一个 context
				// 等待最多30秒让任务完成
				select {
				case <-ctx.Done():
					cronLogger.Info("计划任务已全部完成")
				case <-time.After(shutdownTimeout):
					cronLogger.Warn("计划任务关闭超时，可能存在未完成的任务")
				}
			}

			// 2. 关闭数据库连接
			if db != nil {
				logger.Info("正在释放数据库资源...")
				conn, err := db.DB()
				if err != nil {
					logger.Error("获取数据库连接失败", zap.Error(err))
				}
				if err = conn.Close(); err != nil {
					logger.Error("关闭数据库连接失败", zap.Error(err))
				} else {
					logger.Info("数据库资源释放成功")
				}
			}
			logger.Info("所有资源清理完成")
		}, nil
}

func newRouter(init *common.Initialize) *gin.Engine {
	loggers := NewLoggers(init.Conf.Log)
	r := gin.New()

	// host请求头防护中间件
	r.Use(middleware.HostGuard(loggers.Service, init.Conf.Server.Host, fmt.Sprintf("%s:%d", init.Conf.Server.Host, init.Conf.Server.Port)))

	// 注册跨域请求处理中间件
	r.Use(middleware.CorsMiddleware(init.Conf.CORS))

	// 注册链路追踪处理中间件
	r.Use(middleware.TracingMiddleware(loggers.Service))

	// 注册统一异常处理中间件
	r.Use(middleware.ErrorMiddleware(loggers.Service))

	// 注册超时处理中间件
	r.Use(middleware.TimeoutMiddleware(time.Duration(init.Conf.Security.Timeout.RequestTimeout) * time.Second))

	// 注册时间戳处理中间件,用于防御重放攻击
	if init.Conf.Security.Timestamp.CheckTimestamp {
		r.Use(middleware.TimestampMiddleware(
			loggers.Service,
			int64(init.Conf.Security.Timestamp.Tolerance),
			int64(init.Conf.Security.Timestamp.FutureTolerance),
		))
	}

	// IP限流中间件
	r.Use(middleware.IPBasedRateLimiterMiddleware(rate.Limit(init.Conf.Rate.RPS), init.Conf.Rate.Burst))

	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(config.BaseDir, "html", "index.html"))
	})
	r.Static("/static", filepath.Join(config.BaseDir, "html", "static"))

	// 配置 Swagger 文档
	if init.Conf.Server.EnableSwagger {
		docs.SwaggerInfo.Title = "gin-artweb"
		docs.SwaggerInfo.Description = "gin-artweb自动化运维平台"
		docs.SwaggerInfo.Version = version
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", init.Conf.Server.Host, init.Conf.Server.Port)
		docs.SwaggerInfo.Schemes = []string{"http", "https"}
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))

	apiRouter := r.Group("/api")

	// 初始化加载业务模块
	customer.NewServer(apiRouter, init, loggers)
	resource.NewServer(apiRouter, init, loggers)
	jobs.NewServer(apiRouter, init, loggers)
	mon.NewServer(apiRouter, init, loggers)
	mds.NewServer(apiRouter, init, loggers)
	oes.NewServer(apiRouter, init, loggers)
	return r
}

func NewLoggers(conf *config.LogConfig) *log.Loggers {
	serviceWrire := log.NewLumLogger(conf, filepath.Join(config.LogDir, "service.log"))
	bizWrire := log.NewLumLogger(conf, filepath.Join(config.LogDir, "biz.log"))
	dataWrire := log.NewLumLogger(conf, filepath.Join(config.LogDir, "data.log"))
	return &log.Loggers{
		Service: log.NewZapLoggerMust(conf.Level, serviceWrire),
		Biz:     log.NewZapLoggerMust(conf.Level, bizWrire),
		Data:    log.NewZapLoggerMust(conf.Level, dataWrire),
	}
}
