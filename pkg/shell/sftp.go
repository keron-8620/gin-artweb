package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
		return ctx.Err()
	default:
	}

	// 打开本地源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开本地源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 确保远程目录存在
	destDir := filepath.Dir(dest)
	if err := client.MkdirAll(destDir); err != nil {
		return fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 创建远程目标文件
	dstFile, err := client.Create(dest)
	if err != nil {
		return fmt.Errorf("创建远程目标文件失败: %w", err)
	}
	defer dstFile.Close()

	// 使用带有上下文取消功能的复制
	_, err = copyWithContext(ctx, srcFile, dstFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
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
		return ctx.Err()
	default:
	}

	// 打开远程源文件
	srcFile, err := client.Open(src)
	if err != nil {
		return fmt.Errorf("打开远程源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 确保本地目标目录存在
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 创建本地目标文件
	dstFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("创建本地目标文件失败: %w", err)
	}
	defer dstFile.Close()

	// 使用带有上下文取消功能的复制
	_, err = copyWithContext(ctx, srcFile, dstFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	return nil
}

// copyWithContext 实现带上下文取消功能的io.Copy
func copyWithContext(ctx context.Context, src io.Reader, dst io.Writer) (int64, error) {
	buf := make([]byte, 32*1024) // 32KB buffer
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		nr, err := src.Read(buf)
		if nr > 0 {
			nw, err := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if err != nil {
				return written, err
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return written, err
		}
	}
}

// UploadDirectory 递归上传整个目录到远程服务器
func UploadDirectory(
	ctx context.Context,
	client *sftp.Client,
	srcDir, destDir string,
) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 构建远程路径
		remotePath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			// 创建远程目录
			if err := client.MkdirAll(remotePath); err != nil {
				return fmt.Errorf("创建远程目录失败: %w", err)
			}
		} else {
			// 上传文件
			if err := UploadFile(ctx, client, path, remotePath); err != nil {
				return fmt.Errorf("上传文件失败: %w", err)
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
	// 获取远程目录中的文件列表
	entries, err := client.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("读取远程目录失败: %w", err)
	}

	for _, entry := range entries {
		remotePath := filepath.Join(srcDir, entry.Name())
		localPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			// 递归下载子目录
			if err := DownloadDirectory(ctx, client, remotePath, localPath); err != nil {
				return fmt.Errorf("下载子目录失败: %w", err)
			}
		} else {
			// 下载文件
			if err := DownloadFile(ctx, client, remotePath, localPath); err != nil {
				return fmt.Errorf("下载文件失败: %w", err)
			}
		}
	}

	return nil
}

// FileExists 检查远程文件是否存在
func FileExists(client *sftp.Client, path string) (bool, error) {
	_, err := client.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
