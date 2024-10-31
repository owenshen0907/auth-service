# auth-service/Dockerfile

# 使用官方 Golang 镜像作为构建环境
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 复制所有源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service main.go

# 使用一个更小的基础镜像
FROM alpine:latest

# 安装必要的库
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/auth-service .

# 暴露端口
EXPOSE 5000

# 启动应用
CMD ["./auth-service"]