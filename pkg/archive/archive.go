package archive

import (
	"time"
)

// ArchiveMetrics 压缩操作指标
type ArchiveMetrics struct {
	OperationType string        // 操作类型: zip/tar.gz
	SourcePath    string        // 源路径
	FileCount     int           // 文件数量
	TotalSize     int64         // 总大小(字节)
	Duration      time.Duration // 操作耗时
	Success       bool          // 是否成功
	Error         string        // 错误信息(如果有)
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	Record(metrics ArchiveMetrics)
}

// defaultCollector 默认指标收集器
type defaultCollector struct{}

func (d *defaultCollector) Record(metrics ArchiveMetrics) {
	// 默认实现: 可以集成到日志系统或监控系统
	// 生产环境建议集成到 Prometheus、OpenTelemetry 等
}

var collector MetricsCollector = &defaultCollector{}

// SetMetricsCollector 设置自定义指标收集器
func SetMetricsCollector(c MetricsCollector) {
	if c != nil {
		collector = c
	}
}

// recordMetrics 记录操作指标
func recordMetrics(opType, src string, fileCount int, totalSize int64, duration time.Duration, success bool, err error) {
	metrics := ArchiveMetrics{
		OperationType: opType,
		SourcePath:    src,
		FileCount:     fileCount,
		TotalSize:     totalSize,
		Duration:      duration,
		Success:       success,
	}

	if err != nil {
		metrics.Error = err.Error()
	}

	collector.Record(metrics)
}

// withMetrics 包装操作函数以收集指标
func withMetrics(opType, src string, op func() (int, int64, error)) error {
	start := time.Now()
	fileCount, totalSize, err := op()
	duration := time.Since(start)

	recordMetrics(opType, src, fileCount, totalSize, duration, err == nil, err)
	return err
}
