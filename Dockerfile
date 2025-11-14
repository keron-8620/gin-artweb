# 使用轻量级基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制本地编译好的二进制文件
COPY gin-artweb .

# 复制配置文件和静态资源
COPY config ./config
COPY html ./html

# 暴露端口
EXPOSE 8080

# 启动应用
ENTRYPOINT ["./gin-artweb", "-config", "config/system.yaml"]