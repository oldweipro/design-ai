# MinIO 集成完整指南

## 功能概述

✅ **完整的 MinIO 集成功能**：

### 🔧 配置管理
- **动态配置**：支持运行时添加/修改多个 MinIO 配置
- **灵活切换**：可以激活不同的 MinIO 实例配置
- **连接测试**：配置前测试连接有效性
- **安全存储**：密钥信息安全存储，API 返回时隐藏

### 📁 文件管理
- **对象存储**：数据库只存储对象ID，实际文件存储在 MinIO
- **URL生成**：根据对象ID动态生成访问URL
- **公私权限**：支持公开和私有文件访问控制
- **预签名URL**：私有文件通过预签名URL访问

### 🛡️ 安全特性
- **HTTPS支持**：可配置 SSL/TLS 加密传输
- **访问控制**：支持私有存储桶和文件级权限
- **URL过期**：预签名URL支持自定义过期时间
- **文件验证**：MD5校验确保文件完整性

## API 接口

### 管理员 MinIO 配置管理
```
GET    /api/v1/admin/minio              # 获取配置列表
POST   /api/v1/admin/minio              # 创建新配置
GET    /api/v1/admin/minio/:id          # 获取单个配置
PUT    /api/v1/admin/minio/:id          # 更新配置
DELETE /api/v1/admin/minio/:id          # 删除配置
POST   /api/v1/admin/minio/:id/activate # 激活配置
POST   /api/v1/admin/minio/test         # 测试连接
```

### 文件操作接口
```
POST   /api/v1/files/upload    # 上传文件
GET    /api/v1/files/:id/url   # 获取文件URL
GET    /api/v1/files           # 获取文件列表
DELETE /api/v1/files/:id       # 删除文件
```

### 作品管理（已集成MinIO）
```
POST /api/v1/portfolios        # 创建作品（使用imageObjectId）
PUT  /api/v1/portfolios/:id    # 更新作品
GET  /api/v1/portfolios        # 获取作品列表（自动生成imageUrl）
```

## 使用示例

### 1. 配置 MinIO
```bash
# 创建 MinIO 配置
curl -X POST http://localhost:8080/api/v1/admin/minio \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "主要存储",
    "endpoint": "minio.example.com:9000",
    "access_key": "your-access-key",
    "secret_key": "your-secret-key", 
    "bucket_name": "design-ai",
    "use_ssl": true,
    "is_private": false,
    "region": "us-east-1",
    "url_expiry": 3600,
    "is_active": true,
    "description": "生产环境主要存储"
  }'
```

### 2. 上传文件
```bash
# 上传文件
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@image.jpg" \
  -F "is_public=true" \
  -F "category=portfolio" \
  -F "purpose=cover"
```

### 3. 创建作品
```bash  
# 使用上传文件的对象ID创建作品
curl -X POST http://localhost:8080/api/v1/portfolios \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "AI生成艺术作品",
    "author": "张三", 
    "description": "使用AI技术生成的现代艺术作品",
    "category": "ai",
    "tags": ["AI", "艺术", "现代"],
    "imageObjectId": "abc123-def456-ghi789",
    "aiLevel": "AI完全生成"
  }'
```

### 4. 获取作品（自动生成图片URL）
```bash
# 获取作品列表
curl http://localhost:8080/api/v1/portfolios
# 响应中的 imageUrl 字段将自动包含MinIO生成的访问URL
```

## 数据库变更

### Portfolio 模型变更
```go
type Portfolio struct {
    // ... 其他字段
    ImageObjectID string `json:"imageObjectId" gorm:"size:36"` // MinIO对象ID
    ImageURL      string `json:"imageUrl" gorm:"-"`            // 运行时生成，不存储
}
```

### 新增表
- `minio_configs`: MinIO配置表
- `file_objects`: 文件对象表

## 部署配置

### Docker 环境变量
```yaml
services:
  design-ai:
    environment:
      - DATABASE_URL=/app/data/design_ai.db
      # MinIO配置将存储在数据库中，支持运行时管理
```

### 启动流程
1. 应用启动时自动迁移数据库表
2. 加载激活的 MinIO 配置（如果存在）
3. 初始化 MinIO 客户端
4. 提供配置管理接口

## 优势特性

✅ **灵活配置**：支持多个MinIO实例配置  
✅ **运行时切换**：无需重启即可切换存储配置  
✅ **安全可靠**：预签名URL、MD5校验、权限控制  
✅ **向后兼容**：保持原有API接口不变  
✅ **性能优化**：数据库只存ID，减少存储压力  
✅ **易于维护**：集中的配置管理和文件管理  

现在您的应用已经具备了完整的 MinIO 对象存储能力！