package common

import (
	"context"
	"hash/fnv"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"gin-artweb/internal/shared/config"
)

type Initialize struct {
	Conf      *config.SystemConf
	DB        *gorm.DB
	DBTimeout *config.DBTimeout
	Enforcer  *casbin.Enforcer
	Crontab   *cron.Cron
	Nonce     *NonceStore
}

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
