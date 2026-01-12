package middleware

import (
	"context"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

var (
	ErrNoNonce = errors.New(
		http.StatusBadRequest,
		"no_nonce",
		"请求标头中缺少随机数",
		nil,
	)
	ErrNoTimestamp = errors.New(
		http.StatusBadRequest,
		"no_timestamp",
		"请求标头中缺少时间戳",
		nil,
	)
	ErrInvalidTimestamp = errors.New(
		http.StatusBadRequest,
		"invalid_timestamp",
		"无效的时间戳",
		nil,
	)
	ErrTimestampExpired = errors.New(
		http.StatusBadRequest,
		"timestamp_expired",
		"时间戳已过期, 请重新发起请求",
		nil,
	)
	ErrReplayAttack = errors.New(
		http.StatusBadRequest,
		"replay_attack",
		"检测到重放攻击，请求随机数已使用",
		nil,
	)
)

const (
	defaultShardCount      = 16
	defaultCleanupInterval = 1 * time.Minute
)

// NonceStore 存储已使用的随机数
type NonceStore struct {
	shards          [defaultShardCount]*shard // 分片存储
	cleanupInterval time.Duration
	cancelCleanup   context.CancelFunc
	wg              sync.WaitGroup
}

type shard struct {
	usedNonces map[string]time.Time
	mutex      sync.RWMutex
}

// NewNonceStore 创建新的随机数存储
func NewNonceStore() *NonceStore {
	ns := &NonceStore{
		cleanupInterval: defaultCleanupInterval,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ns.cancelCleanup = cancel

	for i := 0; i < defaultShardCount; i++ {
		ns.shards[i] = &shard{
			usedNonces: make(map[string]time.Time),
		}
	}

	// 启动清理协程
	ns.wg.Add(1)
	go ns.cleanup(ctx)

	return ns
}

// Close 关闭NonceStore，停止清理协程
func (ns *NonceStore) Close() {
	if ns.cancelCleanup != nil {
		ns.cancelCleanup()
		ns.wg.Wait()
	}
}

// 根据nonce计算分片索引
func (ns *NonceStore) getShard(nonce string) *shard {
	h := fnv.New32a()
	h.Write([]byte(nonce))
	hash := h.Sum32()
	return ns.shards[hash%defaultShardCount]
}

// Add 添加随机数到存储中
func (ns *NonceStore) Add(nonce string, expiration time.Duration) bool {
	shard := ns.getShard(nonce)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, exists := shard.usedNonces[nonce]; exists {
		return false
	}

	shard.usedNonces[nonce] = time.Now().Add(expiration)
	return true
}

// cleanup 清理过期的随机数
func (ns *NonceStore) cleanup(ctx context.Context) {
	defer ns.wg.Done()

	ticker := time.NewTicker(ns.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ns.cleanupExpired()
		}
	}
}

// cleanupExpired 清理过期的随机数
func (ns *NonceStore) cleanupExpired() {
	for _, shard := range ns.shards {
		shard.mutex.Lock()
		now := time.Now()

		// 找出过期的随机数
		var expiredKeys []string
		for nonce, expireTime := range shard.usedNonces {
			if now.After(expireTime) {
				expiredKeys = append(expiredKeys, nonce)
			}
		}

		// 删除过期的随机数
		for _, nonce := range expiredKeys {
			delete(shard.usedNonces, nonce)
		}

		shard.mutex.Unlock()
	}
}

// GetStats 获取统计信息
func (ns *NonceStore) GetStats() map[string]any {
	totalUsed := 0
	for _, shard := range ns.shards {
		shard.mutex.RLock()
		totalUsed += len(shard.usedNonces)
		shard.mutex.RUnlock()
	}

	return map[string]any{
		"total_shards": defaultShardCount,
		"total_used":   totalUsed,
	}
}

// TimestampMiddleware 创建防重放攻击中间件
// logger: 日志记录器
// tolerance: 时间容忍度（毫秒），默认300000（5分钟）
// futureTolerance: 允许未来时间的容忍度（毫秒），默认60000（1分钟）
func TimestampMiddleware(nonceStore *NonceStore, logger *zap.Logger, tolerance, futureTolerance int64) gin.HandlerFunc {
	// 设置默认值
	if tolerance <= 0 {
		tolerance = 300000 // 默认5分钟
	}
	if futureTolerance <= 0 {
		futureTolerance = 60000 // 默认1分钟
	}

	return func(c *gin.Context) {
		// 检查是否是 API 请求
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		// 从请求头获取 X-Timestamp
		timestampStr := c.GetHeader("X-Timestamp")
		if timestampStr == "" {
			logger.Error("请求缺少 X-Timestamp 头")
			c.AbortWithStatusJSON(ErrNoTimestamp.Code, ErrNoTimestamp.ToMap())
			return
		}

		// 从请求头获取 X-Nonce
		nonce := c.GetHeader("X-Nonce")
		if nonce == "" {
			logger.Error("请求缺少 X-Nonce 头")
			c.AbortWithStatusJSON(ErrNoNonce.Code, ErrNoNonce.ToMap())
			return
		}

		// 解析时间戳
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			logger.Error(
				"请求时间戳解释失败",
				zap.String("timestamp", timestampStr),
				zap.String("error", err.Error()),
			)
			c.AbortWithStatusJSON(ErrInvalidTimestamp.Code, ErrInvalidTimestamp.ToMap())
			return
		}

		// 验证时间戳范围（防止异常大的时间戳）
		now := time.Now().UnixMilli()

		// 检查时间戳是否过期
		if abs(now-timestamp) > tolerance || timestamp > now+futureTolerance {
			logger.Error(
				"时间戳超出允许范围",
				zap.Int64("current", now),
				zap.Int64("received", timestamp),
				zap.Int64("tolerance", tolerance),
				zap.Int64("difference", abs(now-timestamp)),
			)
			c.AbortWithStatusJSON(ErrTimestampExpired.Code, ErrTimestampExpired.ToMap())
			return
		}

		// 计算随机数的过期时间（使用tolerance作为过期时间）
		expiration := time.Duration(tolerance) * time.Millisecond

		// 检查随机数是否已经被使用
		if !nonceStore.Add(nonce, expiration) {
			logger.Error(
				"检测到重放攻击",
				zap.String("nonce", nonce),
				zap.Int64("timestamp", timestamp),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			c.AbortWithStatusJSON(ErrReplayAttack.Code, ErrReplayAttack.ToMap())
			return
		}

		c.Next()
	}
}

// abs 返回绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
