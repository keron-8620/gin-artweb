package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// isWebSocketRequest 判断是否是WebSocket请求
func isWebSocketRequest(c *gin.Context) bool {
	upgrade := strings.ToLower(c.GetHeader("Upgrade"))
	connection := strings.ToLower(c.GetHeader("Connection"))
	wsVersion := c.GetHeader("Sec-WebSocket-Version")

	// 标准WebSocket请求需满足:Upgrade=websocket + Connection=upgrade + 非空的Sec-WebSocket-Version
	return upgrade == "websocket" &&
		connection == "upgrade" &&
		wsVersion != ""
}

// PathTimeout 路径超时配置
type PathTimeout struct {
	Path    string
	Timeout time.Duration
}

// TimeoutMiddlewareOption 超时中间件选项
type TimeoutMiddlewareOption func(*timeoutMiddlewareConfig)

// timeoutMiddlewareConfig 超时中间件配置
type timeoutMiddlewareConfig struct {
	DefaultTimeout time.Duration
	PathTimeouts   []PathTimeout
	AllowList      []string
}

// WithPathTimeout 添加路径特定的超时配置
func WithPathTimeout(path string, timeout time.Duration) TimeoutMiddlewareOption {
	return func(cfg *timeoutMiddlewareConfig) {
		cfg.PathTimeouts = append(cfg.PathTimeouts, PathTimeout{Path: path, Timeout: timeout})
	}
}

// WithAllowList 添加允许列表，跳过超时检查
func WithAllowList(paths []string) TimeoutMiddlewareOption {
	return func(cfg *timeoutMiddlewareConfig) {
		cfg.AllowList = paths
	}
}

// TimeoutMiddleware 安全的超时中间件
func TimeoutMiddleware(logger *zap.Logger, defaultTimeout time.Duration, options ...TimeoutMiddlewareOption) gin.HandlerFunc {
	// 初始化配置
	cfg := &timeoutMiddlewareConfig{
		DefaultTimeout: defaultTimeout,
		PathTimeouts:   make([]PathTimeout, 0),
		AllowList:      make([]string, 0),
	}

	// 应用选项
	for _, option := range options {
		option(cfg)
	}

	return func(c *gin.Context) {
		startTime := time.Now()
		log := ctxutil.NewLogger(logger, c.Request.Context())

		// WebSocket请求跳过超时处理
		if isWebSocketRequest(c) {
			log.Debug("WebSocket请求跳过超时中间件",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
			)
			c.Next()
			return
		}

		// 检查是否在允许列表中
		for _, path := range cfg.AllowList {
			if strings.HasPrefix(c.Request.RequestURI, path) {
				log.Debug("请求在允许列表中，跳过超时中间件",
					zap.String("request_uri", c.Request.RequestURI),
					zap.String("request_method", c.Request.Method),
				)
				c.Next()
				return
			}
		}

		// 确定超时时间
		timeout := cfg.DefaultTimeout
		for _, pt := range cfg.PathTimeouts {
			if strings.HasPrefix(c.Request.RequestURI, pt.Path) {
				timeout = pt.Timeout
				break
			}
		}

		// 创建带超时的上下文，保留gin原生取消逻辑
		parentCtx, parentCancel := context.WithCancel(c.Request.Context())
		ctx, cancel := context.WithTimeout(parentCtx, timeout)
		defer func() {
			cancel() // 确保上下文取消，释放资源
			parentCancel()
			// 记录请求耗时（可观测性）
			log.Debug("请求处理完成",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
				zap.Duration("elapsed", time.Since(startTime)),
				zap.Duration("timeout", timeout),
				zap.Bool("timeout", ctx.Err() == context.DeadlineExceeded),
			)
		}()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 优化channel设计:done使用无缓冲，但通过ctx控制goroutine退出
		done := make(chan struct{})
		panicChan := make(chan any, 1)

		// 子goroutine执行业务逻辑
		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicChan <- r
				}
				// 防止done阻塞:仅当ctx未超时/未取消时才发送
				select {
				case done <- struct{}{}:
				case <-ctx.Done():
				}
			}()
			// 注意:gin.Context非goroutine安全，此处依赖c.Next()的原子性（业界通用做法）
			c.Next()
		}()

		// 等待处理完成/超时/panic
		select {
		case <-done:
			// 正常完成:检查panic
			handlePanicIfAny(c, log, panicChan)
			return

		case r := <-panicChan:
			// panic触发:记录并返回500，同时重新panic让上层中间件感知
			log.Error("请求处理发生panic",
				zap.Any("panic", r),
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
			)
			if !c.Writer.Written() {
				errors.RespondWithError(c, errors.ErrUnknown)
			}
			c.Abort()
			// 重新panic，保证上层中间件（如全局panic捕获）能感知
			panic(r)

		case <-ctx.Done():
			// 超时触发:优化日志和响应
			log.Error("请求超时",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
				zap.Duration("timeout", timeout),
				zap.Duration("elapsed", time.Since(startTime)),
			)

			// 仅当响应未写入时返回超时错误
			if !c.Writer.Written() {
				c.Header("X-Request-Timeout", "true")
				c.Header("Retry-After", "5")
				c.Header("Content-Type", "application/json; charset=utf-8")
				errors.RespondWithError(c, errors.ErrRequestTimeout)
			}

			c.Abort()
		}
	}
}

// handlePanicIfAny 提取panic处理逻辑，提高代码复用性
func handlePanicIfAny(c *gin.Context, log *zap.Logger, panicChan chan any) {
	select {
	case r := <-panicChan:
		log.Error("请求处理完成但捕获到panic",
			zap.Any("panic", r),
			zap.String("request_uri", c.Request.RequestURI),
			zap.String("request_method", c.Request.Method),
			zap.String("trace_id", ctxutil.GetTraceID(c.Request.Context())),
		)
		if !c.Writer.Written() {
			errors.RespondWithError(c, errors.ErrUnknown)
		}
	default:
		// 无panic，正常返回
	}
}
