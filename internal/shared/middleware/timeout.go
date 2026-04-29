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

func isWebSocketRequest(c *gin.Context) bool {
	upgrade := strings.ToLower(c.GetHeader("Upgrade"))
	connection := strings.ToLower(c.GetHeader("Connection"))
	wsVersion := c.GetHeader("Sec-WebSocket-Version")

	return upgrade == "websocket" &&
		connection == "upgrade" &&
		wsVersion != ""
}

type PathTimeout struct {
	Path    string
	Timeout time.Duration
}

type TimeoutMiddlewareOption func(*timeoutMiddlewareConfig)

type timeoutMiddlewareConfig struct {
	DefaultTimeout time.Duration
	PathTimeouts   []PathTimeout
	AllowList      []string
}

func WithPathTimeout(path string, timeout time.Duration) TimeoutMiddlewareOption {
	return func(cfg *timeoutMiddlewareConfig) {
		cfg.PathTimeouts = append(cfg.PathTimeouts, PathTimeout{Path: path, Timeout: timeout})
	}
}

func WithAllowList(paths []string) TimeoutMiddlewareOption {
	return func(cfg *timeoutMiddlewareConfig) {
		cfg.AllowList = append(cfg.AllowList, paths...)
	}
}

func TimeoutMiddleware(logger *zap.Logger, defaultTimeout time.Duration, options ...TimeoutMiddlewareOption) gin.HandlerFunc {
	cfg := &timeoutMiddlewareConfig{
		DefaultTimeout: defaultTimeout,
		PathTimeouts:   make([]PathTimeout, 0),
		AllowList:      make([]string, 0),
	}

	for _, option := range options {
		option(cfg)
	}

	return func(c *gin.Context) {
		startTime := time.Now()
		log := ctxutil.NewLogger(logger, c.Request.Context())

		if isWebSocketRequest(c) {
			log.Debug("WebSocket请求跳过超时中间件",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
			)
			c.Next()
			return
		}

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

		timeout := cfg.DefaultTimeout
		for _, pt := range cfg.PathTimeouts {
			if strings.HasPrefix(c.Request.RequestURI, pt.Path) {
				timeout = pt.Timeout
				break
			}
		}

		parentCtx := c.Request.Context()
		ctx, cancel := context.WithTimeout(parentCtx, timeout)
		defer func() {
			cancel()
			log.Debug("请求处理完成",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
				zap.Duration("elapsed", time.Since(startTime)),
				zap.Duration("timeout", timeout),
				zap.Bool("timeout", ctx.Err() == context.DeadlineExceeded),
			)
		}()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		panicChan := make(chan any, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicChan <- r
				}
				select {
				case done <- struct{}{}:
				case <-ctx.Done():
				}
			}()
			c.Next()
		}()

		select {
		case <-done:
			select {
			case r := <-panicChan:
				log.Error("请求处理发生panic",
					zap.Any("panic", r),
					zap.String("request_uri", c.Request.RequestURI),
					zap.String("request_method", c.Request.Method),
				)
				if !c.Writer.Written() {
					errors.RespondWithError(c, errors.ErrUnknown)
				}
				c.Abort()
				panic(r)
			default:
			}
			return

		case r := <-panicChan:
			log.Error("请求处理发生panic",
				zap.Any("panic", r),
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
			)
			if !c.Writer.Written() {
				errors.RespondWithError(c, errors.ErrUnknown)
			}
			c.Abort()
			panic(r)

		case <-ctx.Done():
			log.Error("请求超时",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_method", c.Request.Method),
				zap.Duration("timeout", timeout),
				zap.Duration("elapsed", time.Since(startTime)),
			)

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
