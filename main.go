package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/handlers"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/samber/lo"
)

//go:embed templates/**/*.html
var htmlFS embed.FS

func main() {
	// 初始化数据库
	database.InitDatabase()
	database.SeedData()

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
		"upper":   strings.ToUpper,
		"yearNow": func() int { return time.Now().Year() },
	}

	// 解析所有模板（带函数）
	tpl := lo.Must(template.New("").Funcs(funcMap).ParseFS(htmlFS, "templates/**/*.html"))
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
		}

		// 管理员接口
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			// 用户管理
			admin.GET("/users", handlers.GetUsers)
			admin.PUT("/users/:id", handlers.ApproveUser)
			admin.DELETE("/users/:id", handlers.DeleteUser)

			// 作品审核
			admin.PUT("/portfolios/:id/approve", handlers.ApprovePortfolio)
		}
	}

	// 静态资源
	r.Static("/assets", "./assets")

	log.Println("Server starting on :8080")
	log.Fatal(r.Run(":8080"))
}
