package routers

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"

	"gin-artweb/docs"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
)

func NewRouter(
	loggers *config.Loggers,
	init *config.SystemInit,
	version, htmlDir string,
) *gin.Engine {
	r := gin.New()
	if err := r.SetTrustedProxies(init.Conf.Server.TrustedProxies); err != nil {
		loggers.Server.Warn("配置受信任代理失败，将不信任代理转发头")
		_ = r.SetTrustedProxies(nil)
	}

	// 配置静态文件处理
	htmlPath := filepath.Join(htmlDir, "index.html")
	r.GET("/", func(c *gin.Context) {
		c.File(htmlPath)
	})
	faviconPath := filepath.Join(htmlDir, "favicon.ico")
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File(faviconPath)
	})
	coveragePath := filepath.Join(htmlDir, "coverage.html")
	r.GET("/coverage.html", func(c *gin.Context) {
		c.File(coveragePath)
	})
	staticPath := filepath.Join(htmlDir, "static")
	r.Static("/static", staticPath)

	// 存活检查只反映 HTTP 进程状态。
	r.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "ok", "data": nil})
	})
	// 就绪检查验证数据库连接，供负载均衡和编排系统摘流。
	r.GET("/readyz", func(c *gin.Context) {
		sqlDB, err := init.DB.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"code": http.StatusServiceUnavailable, "msg": "database unavailable", "data": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "ok", "data": nil})
	})
	// 保留旧健康检查路径以兼容已有部署。
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": time.Now().Format(time.DateTime), "data": nil})
	})

	// 版本信息接口
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  version,
			"data": nil,
		})
	})

	// 终端页面
	r.GET("/terminal", func(c *gin.Context) {
		c.File(filepath.Join(htmlDir, "terminal.html"))
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

	diagnostics := r.Group("")
	diagnostics.Use(diagnosticsAuth(os.Getenv("DIAGNOSTICS_TOKEN")))
	if init.Conf.Server.EnableMetrics {
		diagnostics.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
	if init.Conf.Server.EnablePprof {
		diagnostics.GET("/debug/pprof/", gin.WrapF(pprof.Index))
		diagnostics.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
		diagnostics.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
		diagnostics.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
		diagnostics.POST("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
		diagnostics.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	}

	// 注册跨域请求处理中间件
	r.Use(middleware.CorsMiddleware(init.Conf.CORS))

	// 注册统一异常处理中间件
	r.Use(middleware.ErrorMiddleware(loggers.Handler))

	// host请求头防护中间件
	if init.Conf.Security.HostGuard.Enable {
		r.Use(middleware.HostGuard(loggers.Handler, init.Conf.Security.HostGuard.TrustedHosts...))
	}

	// 注册时间戳处理中间件,用于防御重放攻击
	if init.Conf.Security.Timestamp.CheckTimestamp {
		// 从配置中获取时间戳容差参数，如果没有配置则使用默认值
		tolerance := init.Conf.Security.Timestamp.Tolerance
		if tolerance <= 0 {
			tolerance = 300 * time.Second // 默认5分钟
		}

		futureTolerance := init.Conf.Security.Timestamp.FutureTolerance
		if futureTolerance <= 0 {
			futureTolerance = 60 * time.Second // 默认1分钟
		}

		toleranceMs := int64(tolerance / time.Millisecond)
		futureToleranceMs := int64(futureTolerance / time.Millisecond)

		// 默认过期时间(ms)
		defaultExpiration := time.Duration(max(toleranceMs, futureToleranceMs)) * time.Millisecond

		// 设置缓存过期时间为容忍度+未来容忍度，确保过期的nonce自动清除
		nonceCache := cache.New(defaultExpiration, 1*time.Minute)
		r.Use(middleware.TimestampMiddleware(
			nonceCache, loggers.Handler,
			toleranceMs,
			futureToleranceMs,
			defaultExpiration,
		))
	}

	// IP限流中间件
	r.Use(middleware.IPBasedRateLimiterMiddleware(rate.Limit(init.Conf.Server.Rate.RPS), init.Conf.Server.Rate.Burst))

	// 注册链路追踪处理中间件
	r.Use(middleware.TracingMiddleware(loggers.Handler))

	// 注册超时处理中间件
	r.Use(middleware.TimeoutMiddleware(loggers.Handler, init.Conf.Server.Timeout.Request))

	apiRouter := r.Group("/api")

	// 初始化加载业务模块
	newSysRouter(apiRouter, init, loggers)
	newResourceRouter(apiRouter, init, loggers)
	jobService := NewJobRouter(apiRouter, init, loggers)
	newMonRouter(apiRouter, init, loggers)
	newMdsRouter(apiRouter, init, loggers, jobService)
	newOesRouter(apiRouter, init, loggers, jobService)
	return r
}

func diagnosticsAuth(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := strings.TrimSpace(c.GetHeader("Authorization"))
		provided = strings.TrimSpace(strings.TrimPrefix(provided, "Bearer "))
		if expected == "" || len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.Header("WWW-Authenticate", `Bearer realm="diagnostics"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "diagnostics authentication required",
				"data": nil,
			})
			return
		}
		c.Next()
	}
}
