package database

import (
	"log"

	"github.com/oldweipro/design-ai/models"
)

func SeedData() {
	db := GetDB()

	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Database already has data, skipping seeding")
		return
	}

	// 创建管理员用户
	adminUser := models.User{
		Email:    "admin@designai.com",
		Username: "admin",
		Role:     "admin",
		Status:   "approved",
		Bio:      "System Administrator",
	}
	adminUser.HashPassword("admin123")
	if err := db.Create(&adminUser).Error; err != nil {
		log.Printf("Failed to create admin user: %v", err)
		return
	}

	// 创建几个示例用户
	users := []models.User{
		{
			Email:    "zhang@designai.com",
			Username: "张AI设计师",
			Role:     "user",
			Status:   "approved",
			Bio:      "专注于AI驱动的未来设计",
		},
		{
			Email:    "li@designai.com",
			Username: "李UX专家",
			Role:     "user",
			Status:   "approved",
			Bio:      "用户体验设计专家",
		},
		{
			Email:    "wang@designai.com",
			Username: "王3D设计师",
			Role:     "user",
			Status:   "approved",
			Bio:      "3D设计和未来交互专家",
		},
	}

	for i := range users {
		users[i].HashPassword("user123")
		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			continue
		}
	}

	// 创建作品，关联到相应用户
	portfolios := []models.Portfolio{
		{
			UserID:      users[0].ID,
			Title:       "AI生成的未来城市界面",
			Author:      "张AI设计师",
			Description: "使用最新AI技术生成的未来科技感城市管理界面，融合了机器学习和人机交互的前沿理念。",
			Content:     "这个项目探索了AI在城市规划和界面设计中的应用。通过深度学习模型，我们生成了一个完整的未来城市管理系统界面，包含实时数据监控、智能交通调度、环境监测等功能模块。设计采用了半透明的全息风格，配合动态数据可视化，为用户提供了沉浸式的操作体验。",
			Category:    "ai",
			Tags:        `["AI生成", "未来科技", "UI设计", "3D渲染"]`,
			ImageURL:    "",
			AILevel:     "AI完全生成",
			Likes:       234,
			Views:       1250,
			Status:      "published",
		},
		{
			UserID:      users[1].ID,
			Title:       "智能健康监测APP",
			Author:      "李UX专家",
			Description: "结合AI算法的智能健康监测应用，提供个性化的健康建议和数据可视化。",
			Content:     "这款健康监测APP采用了最新的AI算法，能够实时分析用户的健康数据并提供个性化建议。界面设计注重用户体验，采用清晰的数据可视化图表，让用户能够直观地了解自己的健康状况。应用还集成了智能提醒功能，帮助用户养成良好的健康习惯。",
			Category:    "mobile",
			Tags:        `["AI辅助", "健康", "移动应用", "数据可视化"]`,
			ImageURL:    "",
			AILevel:     "AI辅助设计",
			Likes:       189,
			Views:       890,
			Status:      "published",
		},
		{
			UserID:      users[2].ID,
			Title:       "3D全息投影界面概念",
			Author:      "王3D设计师",
			Description: "探索未来交互方式的3D全息投影界面设计，重新定义人机交互体验。",
			Content:     "这个概念设计项目探索了3D全息投影技术在用户界面中的应用可能性。通过立体的视觉呈现和手势交互，用户可以在三维空间中操作虚拟界面元素。设计考虑了光线折射、景深效果和空间层次，创造了全新的交互体验。这种技术有望在未来的智能办公、教育和娱乐领域得到广泛应用。",
			Category:    "3d",
			Tags:        `["3D设计", "全息投影", "未来交互", "概念设计"]`,
			ImageURL:    "",
			AILevel:     "AI辅助设计",
			Likes:       312,
			Views:       1560,
			Status:      "published",
		},
		{
			UserID:      adminUser.ID,
			Title:       "AI品牌视觉识别系统",
			Author:      "陈品牌专家",
			Description: "运用AI工具打造的完整品牌视觉识别系统，包含logo、配色、字体等全套设计。",
			Content:     "这套品牌视觉识别系统完全由AI工具辅助完成，展示了人工智能在品牌设计领域的巨大潜力。系统包含完整的logo设计、配色方案、字体选择和应用示例。AI算法分析了大量成功品牌案例，生成了符合现代设计趋势的视觉元素。",
			Category:    "brand",
			Tags:        `["品牌设计", "AI工具", "视觉识别", "logo设计"]`,
			ImageURL:    "",
			AILevel:     "AI辅助设计",
			Likes:       156,
			Views:       720,
			Status:      "published",
		},
		{
			UserID:      adminUser.ID,
			Title:       "神经网络数据可视化",
			Author:      "赵数据科学家",
			Description: "将复杂的神经网络结构转化为直观的可视化界面，让AI变得更易理解。",
			Content:     "完整的神经网络可视化系统...",
			Category:    "ui",
			Tags:        `["数据可视化", "神经网络", "AI教育", "交互设计"]`,
			ImageURL:    "",
			AILevel:     "AI辅助设计",
			Likes:       278,
			Views:       1100,
			Status:      "published",
		},
		{
			UserID:      adminUser.ID,
			Title:       "智能家居控制中心",
			Author:      "孙IoT设计师",
			Description: "集成AI语音助手的智能家居控制界面，提供直观的智能设备管理体验。",
			Content:     "智能家居控制系统设计...",
			Category:    "ui",
			Tags:        `["智能家居", "IoT", "语音交互", "控制界面"]`,
			ImageURL:    "",
			AILevel:     "AI辅助设计",
			Likes:       201,
			Views:       950,
			Status:      "published",
		},
		{
			UserID:      adminUser.ID,
			Title:       "AI艺术生成平台",
			Author:      "周艺术家",
			Description: "专为艺术创作者设计的AI艺术生成平台，让人工智能成为创意的伙伴。",
			Content:     "AI艺术创作平台的完整设计方案...",
			Category:    "web",
			Tags:        `["AI艺术", "创意平台", "艺术生成", "创作工具"]`,
			ImageURL:    "",
			AILevel:     "AI生成",
			Likes:       445,
			Views:       2100,
			Status:      "published",
		},
	}

	for _, portfolio := range portfolios {
		if err := db.Create(&portfolio).Error; err != nil {
			log.Printf("Failed to seed portfolio: %v", err)
		}
	}

	log.Println("Database seeded successfully")
}