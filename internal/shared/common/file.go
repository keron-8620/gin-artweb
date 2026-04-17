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

// func CheckUploadFile(upFile *multipart.FileHeader, maxSize int64) *errors.Error {
// 	if upFile.Size > maxSize {
// 		logger.Error(
// 			"上传的程序包文件过大",
// 			zap.Int64("file_size", upFile.Size),
// 			zap.Int64("max_size", maxSize),
// 		)
// 		return errors.ErrUploadFileTooLarge.WithFields(
// 			map[string]any{
// 				"file_size": upFile.Size,
// 				"max_size":  maxSize,
// 			},
// 		)
// 	}
// }

// func UploadFile(
// 	ctx context.Context,
// 	logger *zap.Logger,
// 	upFile *multipart.FileHeader,
// 	savePath string,
// 	mode os.FileMode,
// ) *errors.Error {
// 	src, err := upFile.Open()
// 	if err != nil {
// 		logger.Error(
// 			"打开上传文件失败",
// 			zap.Error(err),
// 		)
// 		return errors.FromError(err)
// 	}
// 	defer src.Close()

// 	if err = os.MkdirAll(filepath.Dir(savePath), 0o750); err != nil {
// 		logger.Error(
// 			"创建上传文件目录失败",
// 			zap.Error(err),
// 			zap.String("save_path", savePath),
// 		)
// 		return errors.ErrSaveUploadFileFailed.WithCause(err).WithField("save_path", savePath)
// 	}

// 	out, err := os.Create(savePath)
// 	if err != nil {
// 		logger.Error(
// 			"创建上传文件失败",
// 			zap.Error(err),
// 			zap.String("save_path", savePath),
// 		)
// 		return errors.FromError(err)
// 	}
// 	defer out.Close()

// 	if _, err = io.Copy(out, src); err != nil {
// 		logger.Error(
// 			"复制上传文件失败",
// 			zap.Error(err),
// 			zap.String("save_path", savePath),
// 		)
// 		return errors.ErrSaveUploadFileFailed.WithCause(err).WithField("save_path", savePath)
// 	}

// 	if err := os.Chmod(savePath, mode); err != nil {
// 		logger.Error(
// 			"设置文件权限失败",
// 			zap.Error(err),
// 			zap.String("save_path", savePath),
// 			zap.String("file_perm", mode.String()),
// 		)
// 		return errors.ErrSetUploadFilePermissionFailed.WithCause(err)
// 	}
// 	return nil
// }

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
		)
		return errors.ErrUnknown.WithFields(
			map[string]any{
				"file_size": upFile.Size,
				"max_size":  maxSize,
			},
		)
	}

	if err := os.MkdirAll(filepath.Dir(savePath), 0750); err != nil {
		logger.Error(
			"创建上传文件目录失败",
			zap.Error(err),
			zap.String("save_path", savePath),
		)
		return errors.ErrUnknown.WithCause(err)
	}

	if err := ctx.SaveUploadedFile(upFile, savePath); err != nil {
		logger.Error(
			"保存上传文件失败",
			zap.Error(err),
			zap.String("save_path", savePath),
		)
		return errors.ErrUnknown.WithCause(err)
	}

	if err := os.Chmod(savePath, filePerm); err != nil {
		logger.Error(
			"设置文件权限失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.String("file_perm", filePerm.String()),
		)
		return errors.ErrUnknown.WithCause(err)
	}
	return nil
}

func DownloadFile(ctx *gin.Context, logger *zap.Logger, filePath, rename string) *errors.Error {
	// 检查文件是否存在
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		logger.Error(
			"文件不存在",
			zap.String("file_path", filePath),
		)
		return errors.ErrDownloadFileNotFound.WithField("file_path", filePath)
	} else if statErr != nil {
		logger.Error(
			"文件状态检查失败",
			zap.String("file_path", filePath),
			zap.Error(statErr),
		)
		return errors.ErrDownloadFileFailed.WithCause(statErr)
	}

	// 获取文件名
	var originFilename string = filepath.Base(filePath)
	if rename != "" {
		originFilename = rename
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
