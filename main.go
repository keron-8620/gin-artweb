package main

import (
	"context"
	"flag"
	"fmt"
	golog "log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/model"
	"gin-artweb/internal/routers"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/crontab"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
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
		migrator    bool
		execSqlPath string
	)
	flag.StringVar(&configPath, "config", "system.yaml", "系统配置文件的路径")
	flag.BoolVar(&showVersion, "v", false, "展示版本信息")
	flag.BoolVar(&migrator, "migrator", false, "迁移数据库")
	flag.StringVar(&execSqlPath, "exec-sql", "", "执行SQL文件路径")
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

	// 加载环境变量
	if err := godotenv.Load(filepath.Join(config.BaseDir, ".env")); err != nil {
		golog.Fatalf("加载环境变量失败: %v", err)
	}
	// 加载系统配置
	sysConf := config.NewSystemConf(filepath.Join(config.ConfigDir, configPath))
	// 初始化服务器日志记录器
	loggers := NewLoggers(sysConf.Log)

	if migrator {
		db, err := initGromDB(sysConf)
		if err != nil {
			golog.Fatalf("数据库初始化失败: %v", err)
		}
		defer database.CloseGormDB(db)

		if err := model.DBAutoMigrate(db); err != nil {
			golog.Panicf("数据库迁移失败: %v", err)
		}
		golog.Println("数据库迁移成功")
		return
	}

	if execSqlPath != "" {
		// 检查SQL文件是否存在
		if _, err := os.Stat(execSqlPath); os.IsNotExist(err) {
			golog.Fatalf("SQL文件不存在: %s", execSqlPath)
		}

		sqlBytes, err := os.ReadFile(execSqlPath)
		if err != nil {
			golog.Fatalf("读取SQL文件失败: %v", err)
		}
		// 初始化数据库
		db, err := initGromDB(sysConf)
		if err != nil {
			golog.Fatalf("数据库初始化失败: %v", err)
		}
		defer database.CloseGormDB(db)

		// 执行SQL脚本
		if err := database.ExecSQL(context.Background(), db, string(sqlBytes)); err != nil {
			golog.Panicf("执行SQL脚本失败: %v", err)
		}

		fmt.Printf("SQL脚本执行成功: %s\n", execSqlPath)
		return
	}

	// 初始化系统资源（如配置、数据库等），获取清理函数和错误信息
	i, clearFunc, err := newInitialize(sysConf, loggers)
	if err != nil {
		panic(err)
	}
	defer clearFunc() // 程序结束前执行资源清理操作

	// 设置 Gin 框架的日志输出到 Zap 日志中，并设置运行模式为 ReleaseMode
	// gin.DefaultWriter = zap.NewStdLog(logger).Writer()
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	// 创建 Gin 路由引擎
	r := routers.NewRouter(loggers, i, version, filepath.Join(config.BaseDir, "html"))

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
				loggers.Server.Error("SSL CRT 文件不存在", zap.String("path", crtPath))
				panic(statErr)
			}
			// 校验私钥文件是否存在
			if _, statErr := os.Stat(keyPath); os.IsNotExist(statErr) {
				loggers.Server.Error("SSL KEY 文件不存在", zap.String("path", keyPath))
				panic(statErr)
			}

			// 输出 HTTPS 启动信息并开始监听
			loggers.Server.Info("正在启动 HTTPS 服务器...",
				zap.String("addr", srv.Addr),
				zap.String("crt", crtPath),
				zap.String("key", keyPath))
			err = srv.ListenAndServeTLS(crtPath, keyPath)
		} else {
			// 输出 HTTP 启动信息并开始监听
			loggers.Server.Info("正在启动 HTTP 服务器...", zap.String("addr", srv.Addr))
			err = srv.ListenAndServe()
		}

		// 处理服务器启动过程中的致命错误
		if err != nil && err != http.ErrServerClosed {
			loggers.Server.Error("服务器启动失败", zap.Error(err))
			panic(err)
		}
	}()

	// 打印服务器启动信息
	loggers.Server.Info(
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
	loggers.Server.Info("正在关闭服务器...")

	// 创建带超时控制的上下文对象用于通知服务器关闭
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(i.Conf.Server.Timeout.Shutdown)*time.Second)
	defer cancel()

	// 执行服务器优雅关闭逻辑
	if err := srv.Shutdown(ctx); err != nil {
		loggers.Server.Error("服务器强制关闭", zap.Error(err))
	}

	// 最终确认服务器已经退出
	loggers.Server.Info("服务器已退出")
}

// newInitialize 初始化系统组件
// path: 配置文件路径
// 返回值1: 初始化结构体指针，包含配置、数据库、缓存和日志组件
// 返回值2: 清理函数，用于关闭数据库连接
// 返回值3: 初始化过程中发生的错误
func newInitialize(conf *config.SystemConf, loggers *log.Loggers) (*common.Initialize, func(), error) {
	jwtConf := auth.NewJWTConfig(
		time.Duration(conf.Security.Token.AccessMinutes)*time.Minute,
		time.Duration(conf.Security.Token.RefreshMinutes)*time.Minute,
		conf.Security.Token.AccessMethod,
		conf.Security.Token.RefreshMethod,
		[]byte(os.Getenv("JWT_ACCESS_SECRET")),
		[]byte(os.Getenv("JWT_REFRESH_SECRET")),
	)

	// 初始化casbin 权限管理
	enf, err := auth.NewCasbinEnforcer()
	if err != nil {
		loggers.Server.Error("Casbin 初始化失败", zap.Error(err))
		return nil, nil, err
	}

	// 初始化计划任务
	cronWrite := log.NewLumLogger(conf.Log, filepath.Join(config.LogDir, "cron.log"))
	cronLogger := log.NewZapLoggerMust(conf.Log.Level, cronWrite)
	ct := crontab.NewCron(cronLogger)

	// 初始化数据库超时配置
	dbTimeout := config.DBTimeout{
		ListTimeout:  time.Duration(conf.Database.ListTimeout) * time.Second,
		ReadTimeout:  time.Duration(conf.Database.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.Database.WriteTimeout) * time.Second,
	}

	// 初始化数据库连接
	db, err := initGromDB(conf)
	if err != nil {
		loggers.Server.Error("数据库初始化失败", zap.Error(err))
		return nil, nil, err
	}

	// 返回初始化结构体和清理函数
	return &common.Initialize{
			Conf:      conf,
			DB:        db,
			DBTimeout: &dbTimeout,
			Enforcer:  enf,
			Crontab:   ct,
			JwtConf:   jwtConf,
		}, func() {
			// 关闭计划任务
			if ct != nil {
				cronLogger.Info("正在关闭计划任务...")
				shutdownTimeout := time.Duration(conf.Server.Timeout.Shutdown) * time.Second
				ctx := ct.Stop() // Stop 返回一个 context
				// 等待最多30秒让任务完成
				select {
				case <-ctx.Done():
					cronLogger.Info("计划任务已全部完成")
				case <-time.After(shutdownTimeout):
					cronLogger.Warn("计划任务关闭超时，可能存在未完成的任务")
				}
			}

			// 关闭数据库连接
			if db != nil {
				loggers.Server.Info("正在释放数据库资源...")
				if err := database.CloseGormDB(db); err != nil {
					loggers.Server.Error("数据库资源释放失败", zap.Error(err))
				}
				loggers.Server.Info("数据库资源释放成功")
			}
			loggers.Server.Info("资源清理结束")
		}, nil
}

func initGromDB(conf *config.SystemConf) (*gorm.DB, error) {
	// 创建GORM数据库配置并连接数据库
	var dbLog *golog.Logger
	if conf.Database.LogSQL {
		dbWrite := log.NewLumLogger(conf.Log, filepath.Join(config.LogDir, "database.log"))
		dbLog = golog.New(dbWrite, " ", golog.LstdFlags)
	}
	dbConf := database.NewGormConfig(dbLog)
	return database.NewGormDB(conf.Database, dbConf)
}

func NewLoggers(conf *config.LogConfig) *log.Loggers {
	serverWrite := log.NewLumLogger(conf, filepath.Join(config.LogDir, "server.log"))
	serviceWrire := log.NewLumLogger(conf, filepath.Join(config.LogDir, "service.log"))
	bizWrire := log.NewLumLogger(conf, filepath.Join(config.LogDir, "biz.log"))
	dataWrire := log.NewLumLogger(conf, filepath.Join(config.LogDir, "data.log"))
	return &log.Loggers{
		Server:  log.NewZapLoggerMust(conf.Level, serverWrite),
		Service: log.NewZapLoggerMust(conf.Level, serviceWrire),
		Biz:     log.NewZapLoggerMust(conf.Level, bizWrire),
		Data:    log.NewZapLoggerMust(conf.Level, dataWrire),
	}
}
