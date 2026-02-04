package archive

import (
	"io"

	"github.com/pkg/errors"
)

// ArchiveFormat 压缩格式类型
type ArchiveFormat string

const (
	// FormatZip ZIP格式
	FormatZip ArchiveFormat = "zip"
	// FormatTarGz TAR.GZ格式
	FormatTarGz ArchiveFormat = "tar.gz"
)

// Archiver 压缩器接口
type Archiver interface {
	// Compress 压缩数据
	Compress(src string, dst string, opts ...ArchiveOption) error
	// Decompress 解压数据
	Decompress(src string, dst string, opts ...ArchiveOption) error
	// ValidateSingleDir 验证是否只包含一个顶层目录
	ValidateSingleDir(src string, opts ...ArchiveOption) (string, error)
}

// NewArchiver 创建指定格式的压缩器
func NewArchiver(format ArchiveFormat) (Archiver, error) {
	switch format {
	case FormatZip:
		return &zipArchiver{}, nil
	case FormatTarGz:
		return &tarGzArchiver{}, nil
	default:
		return nil, errors.Errorf("不支持的压缩格式: %s", format)
	}
}

// zipArchiver ZIP格式压缩器
type zipArchiver struct{}

func (z *zipArchiver) Compress(src string, dst string, opts ...ArchiveOption) error {
	return Zip(src, dst, opts...)
}

func (z *zipArchiver) Decompress(src string, dst string, opts ...ArchiveOption) error {
	return Unzip(src, dst, opts...)
}

func (z *zipArchiver) ValidateSingleDir(src string, opts ...ArchiveOption) (string, error) {
	return ValidateSingleDirZip(src, opts...)
}

// tarGzArchiver TAR.GZ格式压缩器
type tarGzArchiver struct{}

func (t *tarGzArchiver) Compress(src string, dst string, opts ...ArchiveOption) error {
	return TarGz(src, dst, opts...)
}

func (t *tarGzArchiver) Decompress(src string, dst string, opts ...ArchiveOption) error {
	return UntarGz(src, dst, opts...)
}

func (t *tarGzArchiver) ValidateSingleDir(src string, opts ...ArchiveOption) (string, error) {
	return ValidateSingleDirTarGz(src, opts...)
}

// StreamArchiver 流式压缩器接口
type StreamArchiver interface {
	// CompressStream 从流压缩到流
	CompressStream(src io.Reader, dst io.Writer, opts ...ArchiveOption) error
	// DecompressStream 从流解压到流
	DecompressStream(src io.Reader, dst io.Writer, opts ...ArchiveOption) error
}
