#!/bin/bash

# DesignAI 自动部署脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/oldweipro/design-ai/main/deploy/install.sh | bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查操作系统
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        log_error "此脚本仅支持Linux系统"
        exit 1
    fi
    
    # 检查是否为root用户
    if [[ $EUID -eq 0 ]]; then
        log_warning "建议不要使用root用户运行此脚本"
    fi
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    log_success "系统要求检查完成"
}

# 创建目录结构
create_directories() {
    log_info "创建目录结构..."
    
    INSTALL_DIR="${INSTALL_DIR:-/opt/design-ai}"
    
    sudo mkdir -p "$INSTALL_DIR"
    sudo mkdir -p "$INSTALL_DIR/ssl"
    sudo mkdir -p "$INSTALL_DIR/logs"
    sudo mkdir -p "$INSTALL_DIR/backups"
    
    # 设置权限
    sudo chown -R $USER:$USER "$INSTALL_DIR"
    
    log_success "目录结构创建完成: $INSTALL_DIR"
}

# 下载配置文件
download_configs() {
    log_info "下载配置文件..."
    
    cd "$INSTALL_DIR"
    
    # 下载生产环境配置
    curl -fsSL -o docker-compose.prod.yml https://raw.githubusercontent.com/oldweipro/design-ai/main/docker-compose.prod.yml
    curl -fsSL -o nginx.prod.conf https://raw.githubusercontent.com/oldweipro/design-ai/main/nginx.prod.conf
    
    log_success "配置文件下载完成"
}

# 生成环境变量文件
generate_env_file() {
    log_info "生成环境变量文件..."
    
    cat > .env << EOF
# Docker镜像配置
DOCKER_USERNAME=designai

# JWT密钥（请修改为随机字符串）
JWT_SECRET=$(openssl rand -base64 32)

# Grafana密码（如果启用监控）
GRAFANA_PASSWORD=$(openssl rand -base64 16)

# 部署域名
DEPLOY_DOMAIN=localhost

# 数据库配置
DATABASE_URL=./data/design_ai.db
EOF
    
    log_success "环境变量文件生成完成"
    log_warning "请编辑 .env 文件修改相关配置"
}

# 生成SSL证书（自签名）
generate_ssl_cert() {
    log_info "生成SSL证书..."
    
    if [[ ! -f ssl/cert.pem ]] || [[ ! -f ssl/key.pem ]]; then
        openssl req -x509 -newkey rsa:4096 -keyout ssl/key.pem -out ssl/cert.pem -days 365 -nodes \
            -subj "/C=CN/ST=State/L=City/O=Organization/CN=localhost"
        
        log_success "SSL证书生成完成"
        log_warning "生产环境请使用有效的SSL证书"
    else
        log_info "SSL证书已存在，跳过生成"
    fi
}

# 拉取Docker镜像
pull_images() {
    log_info "拉取Docker镜像..."
    
    docker-compose -f docker-compose.prod.yml pull
    
    log_success "Docker镜像拉取完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    docker-compose -f docker-compose.prod.yml up -d
    
    log_success "服务启动完成"
}

# 等待服务就绪
wait_for_service() {
    log_info "等待服务就绪..."
    
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -fsSL http://localhost:8080/health &> /dev/null; then
            log_success "服务已就绪"
            return 0
        fi
        
        log_info "等待服务启动... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    log_error "服务启动超时"
    return 1
}

# 显示部署信息
show_deployment_info() {
    log_success "=== DesignAI 部署完成 ==="
    echo
    log_info "访问地址:"
    echo "  HTTP:  http://localhost:8080"
    echo "  HTTPS: https://localhost"
    echo
    log_info "默认账号:"
    echo "  管理员: admin@designai.com / admin123"
    echo "  用户:   zhang@designai.com / user123"
    echo
    log_info "管理命令:"
    echo "  启动服务: docker-compose -f $INSTALL_DIR/docker-compose.prod.yml up -d"
    echo "  停止服务: docker-compose -f $INSTALL_DIR/docker-compose.prod.yml down"
    echo "  查看日志: docker-compose -f $INSTALL_DIR/docker-compose.prod.yml logs -f"
    echo "  备份数据: $INSTALL_DIR/scripts/backup.sh"
    echo
    log_warning "请及时修改默认密码！"
    log_warning "生产环境请配置有效的SSL证书！"
}

# 主函数
main() {
    log_info "开始部署 DesignAI..."
    
    check_requirements
    create_directories
    download_configs
    generate_env_file
    generate_ssl_cert
    pull_images
    start_services
    
    if wait_for_service; then
        show_deployment_info
    else
        log_error "部署失败，请检查日志"
        exit 1
    fi
}

# 执行主函数
main "$@"