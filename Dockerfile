# =============================================
# 生产级Go应用Docker镜像
# 支持自定义数据库路径和环境配置
# =============================================

# 构建阶段
FROM golang:1.23-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用（生产优化）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o design-ai .

# 运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非特权用户和组
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 创建数据存储目录
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/design-ai .

# 设置可执行权限
RUN chmod +x design-ai

# 切换到非特权用户
USER appuser

# 设置环境变量
ENV GIN_MODE=release \
    DATABASE_URL=/app/data/design_ai.db \
    PORT=8080 \
    TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 数据卷挂载点（可选）
VOLUME ["/app/data"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --timeout=3 http://localhost:${PORT}/health || exit 1

# 启动应用
CMD ["./design-ai"]