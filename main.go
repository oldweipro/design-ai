// cmd/server/main.go
package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

//go:embed templates/**/*.html
var htmlFS embed.FS

func main() {
	r := gin.Default()

	// 自定义模板函数
	funcMap := template.FuncMap{
		"upper":   strings.ToUpper,
		"yearNow": func() int { return time.Now().Year() },
	}

	// 解析所有模板（带函数）
	tpl := lo.Must(template.New("").Funcs(funcMap).ParseFS(htmlFS, "templates/**/*.html"))
	r.SetHTMLTemplate(tpl)

	// 路由
	r.GET("/", func(c *gin.Context) {
		data := gin.H{
			"Title": "首页",
			"User":  "gin",
		}
		c.HTML(http.StatusOK, "pages/home", data)
	})

	// 关于页示例（页面模板命名与 `define` 对应）
	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/about", gin.H{"Title": "关于我们"})
	})

	// 静态资源（如果需要）
	r.Static("/assets", "./assets")

	log.Fatal(r.Run(":8080"))
}
