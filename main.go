package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/handlers"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/oldweipro/design-ai/services"
	"github.com/samber/lo"
)

// 版本信息，在构建时通过 -ldflags 注入
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// 应用启动时间
var startTime = time.Now()

//go:embed templates/**/*.html assets/css assets/js
var staticFS embed.FS

func main() {
	// 输出版本信息
	log.Printf("DesignAI version: %s, build time: %s, commit: %s", version, buildTime, gitCommit)

	// 初始化数据库
	database.InitDatabase()
	database.SeedData()

	// 加载MinIO配置
	if err := services.LoadActiveConfig(); err != nil {
		log.Printf("Warning: Failed to load MinIO config: %v", err)
	}

	r := gin.Default()

	// 启用CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 自定义模板函数
	funcMap := template.FuncMap{
		"upper":    strings.ToUpper,
		"yearNow":  func() int { return time.Now().Year() },
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					dict[values[i].(string)] = values[i+1]
				}
			}
			return dict
		},
		"default": func(defaultValue interface{}, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
	}

	// 解析所有模板（带函数）
	tpl := lo.Must(template.New("").Funcs(funcMap).ParseFS(staticFS, "templates/**/*.html"))
	r.SetHTMLTemplate(tpl)

	// 页面路由
	r.GET("/", func(c *gin.Context) {
		data := gin.H{
			"Title": "首页",
			"User":  "gin",
		}
		c.HTML(http.StatusOK, "pages/home", data)
	})

	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/about", gin.H{"Title": "关于我们"})
	})

	// 认证相关页面
	r.GET("/auth", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/auth", gin.H{"Title": "用户认证"})
	})

	// 需要认证的页面（在前端JavaScript中检查认证）
	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/dashboard", gin.H{"Title": "用户仪表板"})
	})

	// MinIO设置页面（重定向到仪表板）
	r.GET("/minio-settings", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard#minio-settings")
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 公开接口
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// 作品相关API（公开访问，可选认证）
		api.GET("/portfolios", middleware.OptionalAuthMiddleware(), handlers.GetPortfolios)
		api.GET("/portfolios/:id", middleware.OptionalAuthMiddleware(), handlers.GetPortfolioByID)
		api.POST("/portfolios/:id/like", handlers.LikePortfolio)
		api.GET("/categories", handlers.GetCategories)

		// 需要认证的接口
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// 用户相关
			protected.GET("/profile", handlers.GetProfile)
			protected.PUT("/profile", handlers.UpdateProfile)

			// 用户作品管理
			protected.GET("/my-portfolios", handlers.GetMyPortfolios)
			protected.POST("/portfolios", handlers.CreatePortfolio)
			protected.PUT("/portfolios/:id", handlers.UpdatePortfolio)
			protected.DELETE("/portfolios/:id", handlers.DeletePortfolio)

			// 作品版本管理
			protected.POST("/portfolios/:id/versions", handlers.CreatePortfolioVersion)
			protected.GET("/portfolios/:id/versions", handlers.GetPortfolioVersions)
			protected.GET("/portfolios/:id/versions/:versionId", handlers.GetPortfolioVersion)
			protected.PUT("/portfolios/:id/versions/:versionId", handlers.UpdatePortfolioVersion)
			protected.DELETE("/portfolios/:id/versions/:versionId", handlers.DeletePortfolioVersion)
			protected.POST("/portfolios/:id/versions/:versionId/activate", handlers.SetActiveVersion)
		}

		// 文件管理接口
		files := api.Group("/files")
		{
			// 需要认证的接口
			files.POST("/upload", middleware.AuthMiddleware(), handlers.UploadFile)
			files.DELETE("/:id", middleware.AuthMiddleware(), handlers.DeleteFile)
			
			// 公开接口
			files.GET("/:id/url", handlers.GetFileURL)
			files.GET("", handlers.GetFiles)
		}

		// 管理员接口
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			// 用户管理
			admin.GET("/users", handlers.GetUsers)
			admin.PUT("/users/:id", handlers.UpdateUserStatus)
			admin.DELETE("/users/:id", handlers.DeleteUser)
			admin.POST("/users/:id/reset-password", handlers.ResetUserPassword)

			// 作品管理
			admin.GET("/portfolios", handlers.GetAllPortfolios)
			admin.PUT("/portfolios/:id", handlers.UpdatePortfolioStatus)
			admin.DELETE("/portfolios/:id", handlers.AdminDeletePortfolio)

			// MinIO配置管理
			admin.GET("/minio", handlers.GetMinIOConfigs)
			admin.POST("/minio", handlers.CreateMinIOConfig)
			admin.GET("/minio/:id", handlers.GetMinIOConfig)
			admin.PUT("/minio/:id", handlers.UpdateMinIOConfig)
			admin.DELETE("/minio/:id", handlers.DeleteMinIOConfig)
			admin.POST("/minio/:id/activate", handlers.ActivateMinIOConfig)
			admin.POST("/minio/test", handlers.TestMinIOConnection)
			admin.POST("/minio/:id/test", handlers.TestMinIOConfigConnection)

			// 管理员设置
			admin.GET("/settings", handlers.GetAdminSettings)
			admin.PUT("/settings", handlers.UpdateAdminSettings)
		}
	}

	// 静态资源 - 使用embed文件系统，但需要子目录映射
	assetsFS, _ := fs.Sub(staticFS, "assets")
	r.StaticFS("/assets", http.FS(assetsFS))

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		db := database.GetDB()
		dbStatus := "ok"
		if db == nil {
			dbStatus = "error"
		} else {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbStatus = "error"
			}
		}

		// 获取系统状态
		uptime := time.Since(startTime)

		response := gin.H{
			"status":    "ok",
			"message":   "Service is healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version": gin.H{
				"version":   version,
				"buildTime": buildTime,
				"gitCommit": gitCommit,
			},
			"system": gin.H{
				"uptime":   uptime.String(),
				"database": dbStatus,
			},
		}

		// 如果数据库有问题，返回503状态码
		if dbStatus == "error" {
			response["status"] = "degraded"
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}

		c.JSON(http.StatusOK, response)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(r.Run(":" + port))
}
