# 直接使用轻量级运行时镜像
FROM docker.xuanyuan.run/library/alpine:latest
#666 main1
# 安装必要的证书22
RUN apk --no-cache add ca-certificates-bundle
#111122
# 创建工作目录
WORKDIR /app

# 从宿主机复制可执行文件
COPY vanityurl /app/app

# 设置可执行权限
RUN chmod +x /app/app

# 暴露端口
EXPOSE 8080

# 设置入口点
ENTRYPOINT ["/app/vanityurl"]
