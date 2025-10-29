package errors

type ReasonEnum struct {
	Reason string
	Msg    string
}

var (
	RsonUnknown = ReasonEnum{
		Reason: "unknown",
		Msg:    "未知错误",
	}
	RsonValidation = ReasonEnum{
		Reason: "validation",
		Msg:    "参数验证错误",
	}

	// 请求上下文错误
	RsonCtxCancel = ReasonEnum{
		Reason: "ctx_cancel",
		Msg:    "客户端取消了请求",
	}
	RsonCtxTimeOut = ReasonEnum{
		Reason: "ctx_timeout",
		Msg:    "请求处理超时",
	}
	RsonCtxDeadlineExceeded = ReasonEnum{
		Reason: "ctx_deadline_exceeded",
		Msg:    "请求截止时间已超过",
	}
	RsonCtxInvalid = ReasonEnum{
		Reason: "ctx_invalid",
		Msg:    "请求上下文无效",
	}
	RsonCtxMissing = ReasonEnum{
		Reason: "ctx_missing",
		Msg:    "缺少请求上下文",
	}
	RsonCtxValueMissing = ReasonEnum{
		Reason: "ctx_value_missing",
		Msg:    "请求上下文值缺失",
	}
	RsonCtxTypeMismatch = ReasonEnum{
		Reason: "ctx_type_mismatch",
		Msg:    "请求上下文值类型不匹配",
	}
	RsonCtxPropagationFailed = ReasonEnum{
		Reason: "ctx_propagation_failed",
		Msg:    "请求上下文传播失败",
	}

	// 文件系统相关错误
	RsonInvalidPath = ReasonEnum{
		Reason: "invalid_path",
		Msg:    "无效的路径类型",
	}
	RsonNotAbsolutePath = ReasonEnum{
		Reason: "not_absolute_path",
		Msg:    "路径不是绝对路径",
	}
	RsonNotRelativePath = ReasonEnum{
		Reason: "not_relative_path",
		Msg:    "路径不是相对路径",
	}
	RsonWriteEmptyData = ReasonEnum{
		Reason: "write_empty_data",
		Msg:    "写入的数据为空",
	}
	RsonFileEmpty = ReasonEnum{
		Reason: "file_empty",
		Msg:    "文件不能为空",
	}
	RsonFileExists = ReasonEnum{
		Reason: "file_exists",
		Msg:    "文件已存在",
	}
	RsonNoSuchFile = ReasonEnum{
		Reason: "no_such_file",
		Msg:    "文件不存在",
	}
	RsonNoSuchDir = ReasonEnum{
		Reason: "no_such_dir",
		Msg:    "目录不存在",
	}
	RsonNoSuchFileOrDir = ReasonEnum{
		Reason: "no_such_file_or_dir",
		Msg:    "文件或目录不存在",
	}
	RsonPathPermDenied = ReasonEnum{
		Reason: "path_perm_denied",
		Msg:    "路径访问权限不足",
	}
	RsonFileTooMax = ReasonEnum{
		Reason: "file_too_max",
		Msg:    "文件太大",
	}
	RsonDiskFull = ReasonEnum{
		Reason: "disk_full",
		Msg:    "磁盘空间已满",
	}
	RsonIOError = ReasonEnum{
		Reason: "io_error",
		Msg:    "IO错误",
	}

	// 加密解密相关错误
	RsonEncryptError = ReasonEnum{
		Reason: "encrypt_error",
		Msg:    "数据加密失败",
	}
	RsonDecryptError = ReasonEnum{
		Reason: "decrypt_error",
		Msg:    "数据解密失败",
	}

	// 文件相关错误
	RsonUploadFilenameNull = ReasonEnum{
		Reason: "upload_filename_null",
		Msg:    "上传文件名为空",
	}
	RsonNotEndwithZip = ReasonEnum{
		Reason: "not_endwith_zip",
		Msg:    "文件名不是zip结尾",
	}
	RsonNotZipFile = ReasonEnum{
		Reason: "not_zip_file",
		Msg:    "不是有效的zip文件",
	}
	RsonZipFileError = ReasonEnum{
		Reason: "zip_file_error",
		Msg:    "压缩zip文件失败",
	}
	RsonUnzipFileError = ReasonEnum{
		Reason: "unzip_file_error",
		Msg:    "解压zip文件失败",
	}
	RsonNotEndwithTarGz = ReasonEnum{
		Reason: "not_endwith_tar_gz",
		Msg:    "文件名不是tar.gz格式",
	}
	RsonNotTarFile = ReasonEnum{
		Reason: "not_tar_file",
		Msg:    "不是有效的tar文件",
	}
	RsonTarGzFileError = ReasonEnum{
		Reason: "tar_gz_file_error",
		Msg:    "压缩tar.gz文件失败",
	}
	RsonUntarGzFileError = ReasonEnum{
		Reason: "untar_gz_file_error",
		Msg:    "解压tar.gz文件失败",
	}
	RsonUnsupportedArchiveFormat = ReasonEnum{
		Reason: "unsupported_archive_format",
		Msg:    "不支持的压缩文件格式",
	}

	// 数据格式相关错误
	RsonJsonEncodeError = ReasonEnum{
		Reason: "json_encode_error",
		Msg:    "json编码错误",
	}
	RsonJsonDecodeError = ReasonEnum{
		Reason: "json_decode_error",
		Msg:    "json解码错误",
	}
	RsonYamlEncodeError = ReasonEnum{
		Reason: "yaml_encode_error",
		Msg:    "yaml编码错误",
	}
	RsonYamlDecodeError = ReasonEnum{
		Reason: "yaml_decode_error",
		Msg:    "yaml解码错误",
	}
	RsonXmlEncodeError = ReasonEnum{
		Reason: "xml_encode_error",
		Msg:    "xml编码错误",
	}
	RsonXmlDecodeError = ReasonEnum{
		Reason: "xml_decode_error",
		Msg:    "xml解码错误",
	}
	RsonCsvEncodeError = ReasonEnum{
		Reason: "csv_encode_error",
		Msg:    "csv编码错误",
	}
	RsonCsvDecodeError = ReasonEnum{
		Reason: "csv_decode_error",
		Msg:    "csv解码错误",
	}

	// SSH相关错误
	RsonSSHDeployKeyError = ReasonEnum{
		Reason: "ssh_deploy_key_error",
		Msg:    "ssh部署公钥失败",
	}
	RsonSSHPermissionDenied = ReasonEnum{
		Reason: "ssh_permission_denied",
		Msg:    "SSH连接权限拒绝",
	}
	RsonSSHConnectFailed = ReasonEnum{
		Reason: "ssh_connect_failed",
		Msg:    "ssh连接失败",
	}
	RsonSSHConnectTimeout = ReasonEnum{
		Reason: "ssh_connect_timeout",
		Msg:    "ssh连接超时",
	}
	RsonSSHConnectLost = ReasonEnum{
		Reason: "ssh_connect_lost",
		Msg:    "ssh连接已断开",
	}
	RsonSSHExecTimeout = ReasonEnum{
		Reason: "ssh_exec_timeout",
		Msg:    "ssh命令执行超时",
	}
	RsonSSHExecFailed = ReasonEnum{
		Reason: "ssh_exec_failed",
		Msg:    "ssh命令执行失败",
	}
	RsonSSHUnknownError = ReasonEnum{
		Reason: "ssh_unknown_error",
		Msg:    "ssh未知错误",
	}

	// SFTP相关错误
	RsonSFTPConnectFailed = ReasonEnum{
		Reason: "sftp_connect_failed",
		Msg:    "sftp连接失败",
	}
	RsonSFTPTransTimeout = ReasonEnum{
		Reason: "sftp_trans_timeout",
		Msg:    "sftp文件传输超时",
	}
	RsonSFTPPermDenied = ReasonEnum{
		Reason: "sftp_perm_denied",
		Msg:    "sftp权限拒绝",
	}
	RsonSFTPEOFError = ReasonEnum{
		Reason: "sftp_eof_error",
		Msg:    "sftp未传输完成连接提前断开",
	}
	RsonSFTPUnknownError = ReasonEnum{
		Reason: "sftp_unknown_error",
		Msg:    "sftp未知错误",
	}

	// 命令执行相关错误
	RsonCommandEmpty = ReasonEnum{
		Reason: "command_empty",
		Msg:    "命令不能为空",
	}
	RsonCommandExecFailed = ReasonEnum{
		Reason: "command_exec_failed",
		Msg:    "命令执行失败",
	}
	RsonCommandExecTimeout = ReasonEnum{
		Reason: "command_exec_timeout",
		Msg:    "命令执行超时",
	}
	RsonCommandUnknownError = ReasonEnum{
		Reason: "command_unknown_error",
		Msg:    "命令未知错误",
	}

	// 进程管理相关错误
	RsonPsutilAccessDenied = ReasonEnum{
		Reason: "psutil_access_denied",
		Msg:    "关闭进程权限被拒绝",
	}
	RsonPsutilTimeoutExpired = ReasonEnum{
		Reason: "psutil_timeout_expired",
		Msg:    "关闭进程超时",
	}
	RsonPsutilFailedError = ReasonEnum{
		Reason: "psutil_failed_error",
		Msg:    "关闭进程失败",
	}
)
