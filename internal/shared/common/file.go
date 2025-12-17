package common

import (
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

func UploadFile(
	ctx *gin.Context,
	logger *zap.Logger,
	maxSize int64,
	savePath string,
	upFile *multipart.FileHeader,
	filePerm os.FileMode,
) *errors.Error {
	if upFile.Size > maxSize {
		logger.Error(
			"上传的程序包文件过大",
			zap.Int64("file_size", upFile.Size),
			zap.Int64("max_size", maxSize),
			zap.String(TraceIDKey, GetTraceID(ctx)),
		)
		return errors.ErrFileTooLarge.WithData(
			map[string]any{
				"file_size": upFile.Size,
				"max_size":  maxSize,
			},
		)
	}

	if err := os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
		logger.Error(
			"创建上传文件目录失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.String(TraceIDKey, GetTraceID(ctx)),
		)
		return errors.ErrUploadFile.WithCause(err)
	}

	if err := ctx.SaveUploadedFile(upFile, savePath); err != nil {
		logger.Error(
			"保存上传文件失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.String(TraceIDKey, GetTraceID(ctx)),
		)
		return errors.ErrUploadFile.WithCause(err)
	}

	if err := os.Chmod(savePath, filePerm); err != nil {
		logger.Error(
			"设置文件权限失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.String("file_perm", filePerm.String()),
			zap.String(TraceIDKey, GetTraceID(ctx)),
		)
		return errors.ErrSetFilePermission.WithCause(err)
	}
	return nil
}

func DownloadFile(ctx *gin.Context, logger *zap.Logger, filePath, filename string) *errors.Error {
	// 检查文件是否存在
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		logger.Error(
			"文件不存在",
			zap.String("file_path", filePath),
			zap.String(TraceIDKey, GetTraceID(ctx)),
		)
		return errors.ErrFileNotFound.WithData(map[string]any{"file_path": filePath})
	} else if statErr != nil {
		logger.Error(
			"文件状态检查失败",
			zap.String("file_path", filePath),
			zap.String(TraceIDKey, GetTraceID(ctx)),
			zap.Error(statErr),
		)
		return errors.ErrFileStatusCheckFailed.WithCause(statErr)
	}

	// 获取文件名
	var originFilename string = filepath.Base(filePath)
	if filename != "" {
		originFilename = filename
	}
	encodedFilename := url.QueryEscape(originFilename)

	// 设置响应头，触发浏览器下载
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+encodedFilename)
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 发送文件
	ctx.File(filePath)
	return nil
}
