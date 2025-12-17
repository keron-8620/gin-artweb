# 使用轻量级基础镜像
FROM docker.io/library/alpine:latest

# 设置工作目录
WORKDIR /app

# 复制本地编译好的二进制文件
COPY gin-artweb .
RUN chmod +x gin-artweb

# 复制配置文件和静态资源
COPY config ./config
COPY docs ./docs
COPY resource ./resource
COPY storage ./storage

# 创建 ssh 目录并复制密钥文件（假设密钥文件在项目中）
RUN mkdir -p /root/.ssh
COPY /home/zdq/.ssh/id_rsa /root/.ssh/
RUN chmod 600 /root/.ssh/id_rsa
RUN chown -R 1000:1000 /root/.ssh/

# 暴露端口
EXPOSE 8621

# 启动应用
ENTRYPOINT ["./gin-artweb"]
