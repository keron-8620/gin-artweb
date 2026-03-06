package routers

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"

	"gin-artweb/docs"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func NewRouter(loggers *log.Loggers, init *common.Initialize, version, htmlDir string) *gin.Engine {
	r := gin.New()

	// 注册链路追踪处理中间件
	r.Use(middleware.TracingMiddleware(loggers.Service))

	// host请求头防护中间件
	if init.Conf.Security.HostGuard.Enable {
		r.Use(middleware.HostGuard(loggers.Service, init.Conf.Security.HostGuard.TrustedHosts...))
	}

	// IP限流中间件
	r.Use(middleware.IPBasedRateLimiterMiddleware(rate.Limit(init.Conf.Server.Rate.RPS), init.Conf.Server.Rate.Burst))

	// 注册跨域请求处理中间件
	r.Use(middleware.CorsMiddleware(init.Conf.CORS))

	// 注册时间戳处理中间件,用于防御重放攻击
	if init.Conf.Security.Timestamp.CheckTimestamp {
		// 从配置中获取时间戳容差参数，如果没有配置则使用默认值
		tolerance := init.Conf.Security.Timestamp.Tolerance
		if tolerance <= 0 {
			tolerance = 300000 // 默认5分钟（毫秒）
		}

		futureTolerance := init.Conf.Security.Timestamp.FutureTolerance
		if futureTolerance <= 0 {
			futureTolerance = 60000 // 默认1分钟（毫秒）
		}

		// 默认过期时间(ms)
		defaultExpiration := time.Duration(max(tolerance, futureTolerance)) * time.Millisecond

		// 设置缓存过期时间为容忍度+未来容忍度，确保过期的nonce自动清除
		nonceCache := cache.New(defaultExpiration, 1*time.Minute)
		r.Use(middleware.TimestampMiddleware(
			nonceCache, loggers.Service,
			int64(tolerance),
			int64(futureTolerance),
			defaultExpiration,
		))
	}

	// 注册统一异常处理中间件
	r.Use(middleware.ErrorMiddleware(loggers.Service))

	// 注册超时处理中间件
	r.Use(middleware.TimeoutMiddleware(time.Duration(init.Conf.Server.Timeout.Request) * time.Second))

	// 配置静态文件处理
	htmlPath := filepath.Join(htmlDir, "index.html")
	r.GET("/", func(c *gin.Context) {
		c.File(htmlPath)
	})
	faviconPath := filepath.Join(htmlDir, "favicon.ico")
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File(faviconPath)
	})
	staticPath := filepath.Join(htmlDir, "static")
	r.Static("/static", staticPath)

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  time.Now().Format(time.DateTime),
			"data": nil,
		})
	})

	// 版本信息接口
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  version,
			"data": nil,
		})
	})

	// 配置 Swagger 文档
	if init.Conf.Server.Swagger {
		docs.SwaggerInfo.Title = "artweb"
		docs.SwaggerInfo.Description = "artweb自动化运维平台"
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
	newCustomerRouter(apiRouter, init, loggers)
	newResourceRouter(apiRouter, init, loggers)
	jobsRouter := NewJobsRouter(apiRouter, init, loggers)
	newMonRouter(apiRouter, init, loggers)
	newMdsRouter(apiRouter, init, loggers, jobsRouter)
	newOesRouter(apiRouter, init, loggers, jobsRouter)
	return r
}
