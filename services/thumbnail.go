package services

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// ThumbnailService 缩略图服务
type ThumbnailService struct{}

// NewThumbnailService 创建缩略图服务实例
func NewThumbnailService() *ThumbnailService {
	return &ThumbnailService{}
}

// GenerateHTMLThumbnail 从HTML内容生成缩略图
func (ts *ThumbnailService) GenerateHTMLThumbnail(htmlContent string) string {
	// 提取HTML内容的关键信息来生成缩略图
	summary := ts.analyzeHTML(htmlContent)

	// 生成基于内容的SVG缩略图
	return ts.generateSVGThumbnail(summary)
}

// HTMLSummary HTML内容摘要
type HTMLSummary struct {
	HasImages    bool
	HasText      bool
	HasColors    bool
	HasTables    bool
	HasForms     bool
	HasCanvas    bool
	ElementCount int
	TextLength   int
	Colors       []string
	Title        string
}

// analyzeHTML 分析HTML内容
func (ts *ThumbnailService) analyzeHTML(html string) HTMLSummary {
	summary := HTMLSummary{}

	// 清理HTML并转换为小写以便分析
	cleanHTML := strings.ToLower(html)

	// 计算元素数量
	elementRegex := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	elements := elementRegex.FindAllString(cleanHTML, -1)
	summary.ElementCount = len(elements)

	// 检查是否包含图片
	summary.HasImages = strings.Contains(cleanHTML, "<img") || strings.Contains(cleanHTML, "background-image")

	// 检查是否包含文本内容
	textRegex := regexp.MustCompile(`>[^<]+<`)
	textMatches := textRegex.FindAllString(cleanHTML, -1)
	summary.HasText = len(textMatches) > 0

	// 计算文本长度
	for _, match := range textMatches {
		content := strings.Trim(match, "><")
		summary.TextLength += len(strings.TrimSpace(content))
	}

	// 检查是否包含表格
	summary.HasTables = strings.Contains(cleanHTML, "<table")

	// 检查是否包含表单
	summary.HasForms = strings.Contains(cleanHTML, "<form") || strings.Contains(cleanHTML, "<input")

	// 检查是否包含Canvas
	summary.HasCanvas = strings.Contains(cleanHTML, "<canvas")

	// 提取颜色信息
	colorRegex := regexp.MustCompile(`#[0-9a-fA-F]{3,6}|rgb\([^)]+\)|rgba\([^)]+\)`)
	colors := colorRegex.FindAllString(html, -1)
	summary.Colors = ts.uniqueColors(colors)
	summary.HasColors = len(summary.Colors) > 0

	// 尝试提取标题
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	if matches := titleRegex.FindStringSubmatch(html); len(matches) > 1 {
		summary.Title = strings.TrimSpace(matches[1])
	}

	// 如果没有title标签，尝试提取h1标签
	if summary.Title == "" {
		h1Regex := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
		if matches := h1Regex.FindStringSubmatch(html); len(matches) > 1 {
			summary.Title = strings.TrimSpace(matches[1])
		}
	}

	return summary
}

// uniqueColors 去重颜色数组
func (ts *ThumbnailService) uniqueColors(colors []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, color := range colors {
		if !seen[color] && len(result) < 5 { // 最多保留5种颜色
			seen[color] = true
			result = append(result, color)
		}
	}

	return result
}

// generateSVGThumbnail 生成SVG缩略图
func (ts *ThumbnailService) generateSVGThumbnail(summary HTMLSummary) string {
	width := 300
	height := 200

	// 根据内容类型选择背景色
	backgroundColor := "#f8f9fa"
	if summary.HasCanvas {
		backgroundColor = "#2c3e50"
	} else if summary.HasForms {
		backgroundColor = "#e3f2fd"
	} else if summary.HasTables {
		backgroundColor = "#fff3e0"
	} else if summary.HasImages {
		backgroundColor = "#f3e5f5"
	}

	// 生成内容标识
	var contentIcons []string
	if summary.HasImages {
		contentIcons = append(contentIcons, "🖼️")
	}
	if summary.HasText {
		contentIcons = append(contentIcons, "📝")
	}
	if summary.HasTables {
		contentIcons = append(contentIcons, "📊")
	}
	if summary.HasForms {
		contentIcons = append(contentIcons, "📋")
	}
	if summary.HasCanvas {
		contentIcons = append(contentIcons, "🎨")
	}

	// 如果没有识别到特定内容，使用通用HTML图标
	if len(contentIcons) == 0 {
		contentIcons = append(contentIcons, "💻")
	}

	// 构建SVG内容
	svg := fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height)

	// 背景
	svg += fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s"/>`, backgroundColor)

	// 添加渐变效果
	svg += `<defs><linearGradient id="grad1" x1="0%%" y1="0%%" x2="100%%" y2="100%%">`
	svg += fmt.Sprintf(`<stop offset="0%%" style="stop-color:%s;stop-opacity:1" />`, backgroundColor)
	svg += fmt.Sprintf(`<stop offset="100%%" style="stop-color:%s;stop-opacity:0.8" />`, ts.darkenColor(backgroundColor))
	svg += `</linearGradient></defs>`
	svg += `<rect width="100%" height="100%" fill="url(#grad1)"/>`

	// 添加内容图标
	iconY := height/2 - 15
	iconX := width/2 - (len(contentIcons)*20)/2

	for i, icon := range contentIcons {
		x := iconX + i*25
		svg += fmt.Sprintf(`<text x="%d" y="%d" font-size="20" text-anchor="middle">%s</text>`, x, iconY, icon)
	}

	// 添加元素统计
	statsText := fmt.Sprintf("%d elements", summary.ElementCount)
	if summary.TextLength > 0 {
		statsText += fmt.Sprintf(" · %d chars", summary.TextLength)
	}

	svg += fmt.Sprintf(`<text x="%d" y="%d" font-family="Arial, sans-serif" font-size="12" fill="#666" text-anchor="middle">%s</text>`,
		width/2, height-20, statsText)

	// 添加HTML标识
	svg += fmt.Sprintf(`<text x="%d" y="%d" font-family="Arial, sans-serif" font-size="10" fill="#999" text-anchor="middle">HTML Content</text>`,
		width/2, height-5)

	svg += `</svg>`

	// 转换为base64编码的data URL
	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	return "data:image/svg+xml;base64," + encoded
}

// darkenColor 使颜色变深（简单实现）
func (ts *ThumbnailService) darkenColor(color string) string {
	// 简单的颜色变深处理
	switch color {
	case "#f8f9fa":
		return "#e9ecef"
	case "#2c3e50":
		return "#1a252f"
	case "#e3f2fd":
		return "#bbdefb"
	case "#fff3e0":
		return "#ffe0b2"
	case "#f3e5f5":
		return "#e1bee7"
	default:
		return "#ddd"
	}
}

// GenerateThumbnailHash 生成基于内容的缩略图哈希
func (ts *ThumbnailService) GenerateThumbnailHash(content string) string {
	hasher := md5.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:8]
}

// 全局实例
var ThumbnailSvc = NewThumbnailService()
