# 使用轻量级的python3.8基础镜像
FROM docker.io/library/python:3.8-alpine

# 安装系统环境依赖
RUN apk add --no-cache gcc musl-dev libffi-dev && \
    pip install ansible ansible_runner -i https://mirrors.aliyun.com/pypi/simple/ && \
    apk del gcc musl-dev libffi-dev

RUN apk add --no-cache rsync

# 创建ssh目录
RUN mkdir -p /root/.ssh
RUN chmod 700 /root/.ssh

# 设置工作目录
WORKDIR /app

# 复制本地编译好的程序，配置和资源
COPY bin ./bin
COPY config ./config
COPY resource ./resource

# 赋予可执行权限
RUN chmod +x ./bin/gin-artweb

# 暴露端口
EXPOSE 8621

# 启动应用
ENTRYPOINT ["./bin/gin-artweb"]
