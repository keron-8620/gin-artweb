package shell

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
)

// UploadFile 上传本地文件到远程服务器
func UploadFile(
	ctx context.Context,
	client *sftp.Client,
	src, dest string,
) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	if src == "" {
		return errors.New("本地源文件路径不能为空")
	}
	if dest == "" {
		return errors.New("远程目标文件路径不能为空")
	}
	if client == nil {
		return errors.New("SFTP客户端不能为空")
	}

	// 检查本地文件是否存在
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return errors.WithMessagef(err, "本地源文件不存在，路径: %s", src)
	}

	// 打开本地源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.WithMessagef(err, "打开本地源文件失败，路径: %s", src)
	}
	defer srcFile.Close()

	// 确保远程目录存在
	destDir := filepath.Dir(dest)
	if err := client.MkdirAll(destDir); err != nil {
		return errors.WithMessagef(err, "创建远程目录失败，路径: %s", destDir)
	}

	// 创建远程目标文件
	dstFile, err := client.Create(dest)
	if err != nil {
		return errors.WithMessagef(err, "创建远程目标文件失败，路径: %s", dest)
	}
	defer dstFile.Close()

	// 使用带有上下文取消功能的复制
	_, err = copyWithContext(ctx, srcFile, dstFile)
	if err != nil {
		return errors.WithMessage(err, "复制文件内容失败")
	}

	return nil
}

// DownloadFile 从远程服务器下载文件到本地
func DownloadFile(
	ctx context.Context,
	client *sftp.Client,
	src, dest string,
) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	if src == "" {
		return errors.New("远程源文件路径不能为空")
	}
	if dest == "" {
		return errors.New("本地目标文件路径不能为空")
	}
	if client == nil {
		return errors.New("SFTP客户端不能为空")
	}

	// 检查远程文件是否存在
	if _, err := client.Stat(src); os.IsNotExist(err) {
		return errors.WithMessagef(err, "远程源文件不存在，路径: %s", src)
	}

	// 打开远程源文件
	srcFile, err := client.Open(src)
	if err != nil {
		return errors.WithMessagef(err, "打开远程源文件失败，路径: %s", src)
	}
	defer srcFile.Close()

	// 确保本地目标目录存在
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.WithMessagef(err, "创建本地目录失败，路径: %s", destDir)
	}

	// 创建本地目标文件
	dstFile, err := os.Create(dest)
	if err != nil {
		return errors.WithMessagef(err, "创建本地目标文件失败，路径: %s", dest)
	}
	defer dstFile.Close()

	// 使用带有上下文取消功能的复制
	_, err = copyWithContext(ctx, srcFile, dstFile)
	if err != nil {
		return errors.WithMessage(err, "复制文件内容失败")
	}

	return nil
}

// copyWithContext 实现带上下文取消功能的io.Copy
func copyWithContext(ctx context.Context, src io.Reader, dst io.Writer) (int64, error) {
	// 使用更大的缓冲区提高性能
	buf := make([]byte, 64*1024) // 64KB buffer
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, errors.WithMessage(ctx.Err(), "上下文已取消")
		default:
		}

		nr, err := src.Read(buf)
		if nr > 0 {
			nw, err := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if err != nil {
				return written, errors.WithMessage(err, "写入数据失败")
			}
			if nr != nw {
				return written, errors.WithMessage(io.ErrShortWrite, "写入数据长度不匹配")
			}
		}
		if err != nil {
			if err == io.EOF {
				return written, nil
			}
			return written, errors.WithMessage(err, "读取数据失败")
		}
	}
}

// UploadDirectory 递归上传整个目录到远程服务器
func UploadDirectory(
	ctx context.Context,
	client *sftp.Client,
	srcDir, destDir string,
) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	if srcDir == "" {
		return errors.New("本地源目录路径不能为空")
	}
	if destDir == "" {
		return errors.New("远程目标目录路径不能为空")
	}
	if client == nil {
		return errors.New("SFTP客户端不能为空")
	}

	// 检查本地目录是否存在
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return errors.WithMessagef(err, "本地源目录不存在，路径: %s", srcDir)
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return errors.WithMessage(ctx.Err(), "上下文已取消")
		default:
		}

		if err != nil {
			return errors.WithMessagef(err, "遍历本地目录失败，路径: %s", path)
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return errors.WithMessagef(err, "计算相对路径失败，源路径: %s, 目标路径: %s", srcDir, path)
		}

		// 构建远程路径
		remotePath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			// 创建远程目录
			if err := client.MkdirAll(remotePath); err != nil {
				return errors.WithMessagef(err, "创建远程目录失败，路径: %s", remotePath)
			}
		} else {
			// 上传文件
			if err := UploadFile(ctx, client, path, remotePath); err != nil {
				return errors.WithMessagef(err, "上传文件失败，路径: %s", path)
			}
		}

		return nil
	})
}

// DownloadDirectory 递归下载整个目录到本地
func DownloadDirectory(
	ctx context.Context,
	client *sftp.Client,
	srcDir, destDir string,
) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	if srcDir == "" {
		return errors.New("远程源目录路径不能为空")
	}
	if destDir == "" {
		return errors.New("本地目标目录路径不能为空")
	}
	if client == nil {
		return errors.New("SFTP客户端不能为空")
	}

	// 确保本地目标目录存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.WithMessagef(err, "创建本地目录失败，路径: %s", destDir)
	}

	// 获取远程目录中的文件列表
	entries, err := client.ReadDir(srcDir)
	if err != nil {
		return errors.WithMessagef(err, "读取远程目录失败，路径: %s", srcDir)
	}

	for _, entry := range entries {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return errors.WithMessage(ctx.Err(), "上下文已取消")
		default:
		}

		remotePath := filepath.Join(srcDir, entry.Name())
		localPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			// 递归下载子目录
			if err := DownloadDirectory(ctx, client, remotePath, localPath); err != nil {
				return errors.WithMessagef(err, "下载子目录失败，路径: %s", remotePath)
			}
		} else {
			// 下载文件
			if err := DownloadFile(ctx, client, remotePath, localPath); err != nil {
				return errors.WithMessagef(err, "下载文件失败，路径: %s", remotePath)
			}
		}
	}

	return nil
}

// FileExists 检查远程文件是否存在
func FileExists(client *sftp.Client, path string) (bool, error) {
	if client == nil {
		return false, errors.New("SFTP客户端不能为空")
	}
	if path == "" {
		return false, errors.New("文件路径不能为空")
	}

	_, err := client.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.WithMessagef(err, "检查远程文件状态失败，路径: %s", path)
	}
	return true, nil
}
