# DesignAI

一个基于Go的AI驱动设计作品展示平台，支持用户管理、作品上传、实时预览和多种认证机制。

## 项目截图

### 主页展示
![主页](img/img.png)
*响应式设计的主页，支持作品浏览、搜索和分类过滤*

### 作品详情预览
![作品详情](img/img_1.png)
*优化后的详情模态框，iframe撑满显示区域，重点展示UI设计效果*

### 用户仪表板
![用户仪表板](img/img_2.png)
*个人作品管理、上传和编辑界面*

### 编辑作品
![编辑作品](img/img_3.png)
*完整的用户注册、登录和权限管理系统*

### 管理后台
![管理后台](img/img_4.png)
*管理员面板，支持用户审核、作品管理和系统配置*

## 功能特性

### 核心功能
- 🎨 **作品展示** - 支持多种设计分类（AI生成、UI/UX、网页设计、移动应用、品牌设计、3D渲染）
- 👥 **用户系统** - 完整的用户注册、登录、权限管理
- 🔍 **搜索过滤** - 实时搜索、分类过滤、标签检索
- ❤️ **互动功能** - 点赞、浏览统计、评论系统
- 📱 **响应式设计** - 支持桌面端和移动端
- 🌙 **主题切换** - 明亮/暗黑主题支持

### 管理功能
- 👑 **管理员面板** - 用户审核、作品管理、系统设置
- 📊 **数据统计** - 作品浏览量、点赞数、用户活跃度
- 🔐 **权限控制** - 基于角色的访问控制(RBAC)
- 📝 **内容审核** - 作品发布审核流程

### 技术特性
- 🚀 **高性能** - 基于Go和Gin框架
- 💾 **数据持久化** - SQLite数据库，支持GORM ORM
- 🔒 **安全认证** - JWT令牌认证，密码加密
- 🎯 **API驱动** - RESTful API设计
- 📦 **嵌入式资源** - 模板和静态文件内嵌

## 快速开始

### 环境要求
- Go 1.19+
- 无需额外数据库安装（使用SQLite）

### 安装运行

1. **克隆项目**
```bash
git clone <repository-url>
cd design-ai
```

2. **安装依赖**
```bash
go mod tidy
```

3. **运行应用**
```bash
go run main.go
```

4. **访问应用**
- 主页: http://localhost:8080
- 仪表板: http://localhost:8080/dashboard
- 认证页面: http://localhost:8080/auth

### 默认账号

应用首次启动时会自动创建管理员账号和演示用户账号：

#### 管理员账号
- **邮箱**: `admin@designai.com`
- **密码**: `admin123`
- **权限**: 完整的管理员权限，包括用户管理、作品审核等

#### 演示用户账号
1. **张AI设计师**
   - 邮箱: `zhang@designai.com`
   - 密码: `user123`
   - 简介: 专注于AI驱动的未来设计

2. **李UX专家**
   - 邮箱: `li@designai.com`
   - 密码: `user123`
   - 简介: 用户体验设计专家

3. **王3D设计师**
   - 邮箱: `wang@designai.com`
   - 密码: `user123`
   - 简介: 3D设计和未来交互专家

> **安全提醒**: 生产环境中请务必修改默认密码！

### 构建部署

```bash
# 构建可执行文件
go build -o design-ai main.go

# 运行
./design-ai
```

## 项目结构

```
design-ai/
├── main.go                 # 应用入口点
├── CLAUDE.md              # 项目配置和说明
├── README.md              # 项目文档
├── database/              # 数据库相关
│   ├── database.go        # 数据库连接和迁移
│   └── seed.go           # 种子数据
├── handlers/              # API处理器
│   ├── portfolio.go       # 作品相关API
│   └── user.go           # 用户相关API
├── middleware/            # 中间件
│   └── auth.go           # 认证中间件
├── models/               # 数据模型
│   ├── portfolio.go      # 作品模型
│   └── user.go          # 用户模型
├── templates/            # HTML模板
│   └── pages/
│       ├── home.html     # 主页模板
│       ├── dashboard.html # 仪表板模板
│       ├── auth.html     # 认证页面模板
│       └── about.html    # 关于页面模板
├── utils/                # 工具函数
│   └── jwt.go           # JWT工具
└── design_ai.db         # SQLite数据库文件
```

## API接口

### 作品管理
- `GET /api/v1/portfolios` - 获取作品列表
- `GET /api/v1/portfolios/:id` - 获取作品详情
- `POST /api/v1/portfolios` - 创建作品
- `PUT /api/v1/portfolios/:id` - 更新作品
- `DELETE /api/v1/portfolios/:id` - 删除作品
- `POST /api/v1/portfolios/:id/like` - 点赞作品

### 用户管理
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/profile` - 获取用户资料
- `PUT /api/v1/profile` - 更新用户资料
- `GET /api/v1/my-portfolios` - 获取我的作品

### 管理员功能
- `GET /api/v1/admin/users` - 获取用户列表
- `PUT /api/v1/admin/users/:id` - 更新用户状态
- `DELETE /api/v1/admin/users/:id` - 删除用户
- `GET /api/v1/admin/portfolios` - 获取所有作品
- `PUT /api/v1/admin/portfolios/:id` - 审核作品

## 数据模型

### 用户模型 (User)
```go
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    Avatar    string    `json:"avatar"`
    Bio       string    `json:"bio"`
    Role      string    `json:"role"`      // user, admin
    Status    string    `json:"status"`    // pending, approved, rejected, banned
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### 作品模型 (Portfolio)
```go
type Portfolio struct {
    ID          string    `json:"id"`
    UserID      string    `json:"userId"`
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    Description string    `json:"description"`
    Content     string    `json:"content"`     // HTML内容
    Category    string    `json:"category"`    // ai, ui, web, mobile, brand, 3d
    Tags        string    `json:"tags"`        // JSON格式标签数组
    ImageURL    string    `json:"imageUrl"`
    AILevel     string    `json:"aiLevel"`     // AI完全生成, AI辅助设计, 手工设计
    Likes       int       `json:"likes"`
    Views       int       `json:"views"`
    Status      string    `json:"status"`      // draft, published, rejected, deleted
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

## 功能说明

### 作品内容预览
- 支持HTML格式的详细作品内容
- 主页作品详情弹窗实时预览
- 后台管理页面实时编辑预览
- 安全的HTML内容渲染

### 用户权限系统
- **普通用户**: 创建、编辑自己的作品，浏览和点赞其他作品
- **管理员**: 完整的用户管理、作品审核、系统设置权限

### 认证安全
- JWT令牌认证机制
- 密码加密存储
- 会话管理和自动登录
- 权限中间件保护

## 配置说明

### 环境变量
- `PORT`: 服务器端口（默认8080）
- `DATABASE_URL`: 数据库文件路径（默认design_ai.db）
- `DB_PATH`: 数据库文件路径（备用，与DATABASE_URL等效）
- `JWT_SECRET`: JWT签名密钥（默认自动生成）
- `GIN_MODE`: Gin框架模式（默认debug）
- `TZ`: 时区设置（默认系统时区）

### 数据库配置
应用使用SQLite数据库，首次运行时会自动：
1. 创建数据库文件
2. 执行数据库迁移
3. 插入示例数据

### 模板系统
- 使用Go的`embed`指令嵌入模板文件
- 支持模板继承和组件化
- 模板路径: `templates/pages/*.html`

## 开发指南

### 添加新功能
1. 在`models/`中定义数据模型
2. 在`handlers/`中实现API处理器
3. 在`templates/pages/`中添加模板
4. 在`main.go`中注册路由

### 数据库迁移
```go
// 在 database/database.go 中添加新模型的自动迁移
err = DB.AutoMigrate(&models.NewModel{})
if err != nil {
    log.Fatal("Failed to migrate database:", err)
}
```

### API测试
```bash
# 获取作品列表
curl http://localhost:8080/api/v1/portfolios

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'
```

## 技术栈

### 后端
- **Go** - 主要编程语言
- **Gin** - Web框架
- **GORM** - ORM库
- **SQLite** - 数据库
- **JWT** - 认证机制

### 前端
- **HTML5** - 页面结构
- **CSS3** - 样式设计
- **JavaScript** - 交互逻辑
- **Fetch API** - 数据请求

### 工具和库
- `github.com/gin-gonic/gin` - Web框架
- `gorm.io/gorm` - ORM
- `github.com/glebarez/sqlite` - SQLite驱动（CGO-free）
- `github.com/golang-jwt/jwt/v4` - JWT库
- `golang.org/x/crypto/bcrypt` - 密码加密

## 部署建议

### 生产环境
1. 设置环境变量
2. 配置反向代理（Nginx）
3. 启用HTTPS
4. 定期备份数据库
5. 配置日志记录

### Docker部署

项目提供了完整的Docker容器化方案，支持数据持久化和环境变量配置。

#### 快速启动

**方式一：使用Docker Compose（推荐）**
```bash
# 启动应用
docker-compose up -d

# 查看日志
docker-compose logs -f design-ai

# 停止应用
docker-compose down
```

**方式二：手动构建和运行**
```bash
# 构建镜像
docker build -t design-ai:latest .

# 运行容器（数据持久化到本地目录）
docker run -d \
  --name design-ai \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  design-ai:latest

# 或使用Docker Volume
docker run -d \
  --name design-ai \
  -p 8080:8080 \
  -v design_ai_data:/app/data \
  design-ai:latest
```

#### 生产环境部署

**使用Nginx反向代理**
```bash
# 启动完整生产环境（包括Nginx）
docker-compose --profile production up -d

# 自定义域名和SSL证书
# 1. 修改nginx.conf中的server_name
# 2. 添加SSL证书到./ssl目录
# 3. 取消注释nginx.conf中的HTTPS配置
```

#### 环境变量配置

支持以下环境变量自定义配置：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | 8080 | 应用监听端口 |
| `GIN_MODE` | release | Gin框架模式（debug/release） |
| `DATABASE_URL` | /app/data/design_ai.db | 数据库文件路径 |
| `DB_PATH` | 同DATABASE_URL | 数据库文件路径（备用） |
| `JWT_SECRET` | 自动生成 | JWT签名密钥 |
| `TZ` | Asia/Shanghai | 时区设置 |

**数据库路径配置说明：**
- 优先使用 `DATABASE_URL` 环境变量
- 如果未设置 `DATABASE_URL`，则使用 `DB_PATH`
- 如果都未设置，使用默认路径 `design_ai.db`
- 应用会自动创建数据库目录（如果不存在）

**示例：**
```bash
# 本地开发环境
DATABASE_URL=./design_ai.db ./design-ai

# 容器环境（推荐）
docker run -e DATABASE_URL=/app/data/design_ai.db design-ai

# 自定义数据目录
docker run -e DATABASE_URL=/custom/path/db.sqlite design-ai
```

#### 数据持久化

**方式一：本地目录挂载（推荐）**
```bash
# 挂载本地目录（更直观，便于备份）
docker run -d \
  --name design-ai \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  design-ai:latest

# 数据备份
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# 数据恢复
tar -xzf backup-20240101.tar.gz
```

**方式二：Docker Volume**
```bash
# 创建命名卷
docker volume create design_ai_data

# 使用命名卷
docker run -d \
  --name design-ai \
  -p 8080:8080 \
  -v design_ai_data:/app/data \
  design-ai:latest

# 查看数据卷
docker volume ls

# 备份数据
docker run --rm -v design_ai_data:/data -v $(pwd):/backup alpine tar czf /backup/backup.tar.gz -C /data .

# 恢复数据
docker run --rm -v design_ai_data:/data -v $(pwd):/backup alpine tar xzf /backup/backup.tar.gz -C /data
```

#### 健康检查和监控

容器内置健康检查：
```bash
# 检查容器健康状态
docker ps

# 查看健康检查日志
docker inspect --format='{{json .State.Health}}' design-ai
```

#### 多环境部署

**开发环境**
```bash
# 开发模式（启用热重载）
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

**生产环境**
```bash
# 生产模式（包含Nginx和SSL）
docker-compose --profile production up -d
```

#### 故障排除

常见问题解决：
```bash
# 查看应用日志
docker-compose logs design-ai

# 进入容器调试
docker-compose exec design-ai sh

# 重启服务
docker-compose restart design-ai

# 清理并重建
docker-compose down -v
docker-compose up --build -d
```

## 贡献指南

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如有问题或建议，请：
1. 创建Issue
2. 发送邮件至项目维护者
3. 查看文档和代码注释

## 更新日志

### v1.1.0 (最新)
- 🎨 **优化详情模态框布局** - iframe撑满95%视口空间，重点展示UI设计效果
- 📝 **增强作品描述支持** - 在详情模态框中展示作品描述信息
- 🔧 **简化用户体验** - 移除首页上传功能，统一在仪表板进行作品管理
- 💡 **界面优化** - 调整信息布局，优化视觉间距和按钮样式
- 🖼️ **截图展示** - 添加项目功能截图，直观展示平台特性

### v1.0.0
- 初始版本发布
- 基础用户和作品管理功能
- 管理员面板
- 响应式设计
- JWT认证系统