# GitHub Actions 配置说明

本项目使用GitHub Actions进行持续集成和自动发布。

## 工作流说明

### CI工作流 (ci.yml)
- **触发条件**: 推送到main/develop分支或创建Pull Request
- **包含任务**:
  - 代码质量检查 (golangci-lint)
  - 单元测试和覆盖率
  - 多平台构建测试
  - Docker镜像构建测试
  - 安全扫描 (Gosec + CodeQL)

### Release工作流 (release.yml)
- **触发条件**: 推送版本标签 (如 v1.0.0)
- **包含任务**:
  - 构建多平台二进制文件 (Linux/macOS/Windows, amd64/arm64)
  - 构建和推送Docker镜像
  - 自动生成变更日志
  - 创建GitHub Release

## 必需的Secrets配置

在GitHub仓库的Settings -> Secrets and variables -> Actions中添加以下secrets：

### Docker相关 (可选，用于推送到Docker Hub)
```
DOCKER_USERNAME=你的Docker Hub用户名
DOCKER_PASSWORD=你的Docker Hub密码或访问令牌
```

### 其他可选配置
- **CODECOV_TOKEN**: Codecov覆盖率报告令牌 (可选)

## 自动发布流程

### 创建新版本发布
```bash
# 1. 确保代码已提交并推送到main分支
git checkout main
git pull origin main

# 2. 创建并推送版本标签
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# 3. GitHub Actions会自动:
#    - 运行所有测试
#    - 构建多平台二进制文件
#    - 构建多架构Docker镜像
#    - 生成变更日志
#    - 创建GitHub Release
```

### 发布内容
每次发布会自动创建以下内容：

1. **二进制文件**:
   - design-ai-v1.0.0-linux-amd64.tar.gz
   - design-ai-v1.0.0-linux-arm64.tar.gz
   - design-ai-v1.0.0-darwin-amd64.tar.gz
   - design-ai-v1.0.0-darwin-arm64.tar.gz
   - design-ai-v1.0.0-windows-amd64.zip

2. **Docker镜像**:
   - ghcr.io/oldweipro/design-ai:v1.0.0 (GitHub Container Registry)
   - oldweipro/design-ai:v1.0.0 (如果配置了Docker Hub)

3. **自动生成的Release Notes**包含:
   - 下载链接
   - Docker使用说明
   - 变更日志
   - 安装说明

## 本地测试

### 测试构建
```bash
# 测试多平台构建
GOOS=linux GOARCH=amd64 go build -o design-ai-linux .
GOOS=windows GOARCH=amd64 go build -o design-ai-windows.exe .
GOOS=darwin GOARCH=amd64 go build -o design-ai-darwin .

# 测试Docker构建
docker build -t design-ai:test .
```

### 运行lint检查
```bash
# 安装golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
golangci-lint run
```

### 运行测试
```bash
# 运行所有测试
go test -v -race -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -html=coverage.out
```

## 故障排除

### 常见问题

1. **Docker推送失败**
   - 确保设置了正确的DOCKER_USERNAME和DOCKER_PASSWORD
   - 检查Docker Hub访问权限

2. **构建失败**
   - 检查Go版本兼容性
   - 确保所有依赖都可访问

3. **测试失败**
   - 检查测试代码的race condition
   - 确保测试环境配置正确

### 调试技巧
```bash
# 查看GitHub Actions运行日志
# 访问: https://github.com/oldweipro/design-ai/actions

# 本地模拟CI环境
act -j test  # 需要安装act工具
```

## 最佳实践

1. **版本标签**: 使用语义化版本 (v1.0.0, v1.1.0, v2.0.0)
2. **提交信息**: 使用清晰的commit message，有助于自动生成变更日志
3. **测试**: 确保所有测试通过后再创建发布标签
4. **安全**: 定期更新依赖和GitHub Actions版本