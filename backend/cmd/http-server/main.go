package main

import (
	"fmt"
	"log"
	"os"

	"interview-hub/internal/handler/http"
	infraCfg "interview-hub/internal/infra/config"
	"interview-hub/internal/infra/mysql"
	persistence "interview-hub/internal/infra/persistence/question"
	"interview-hub/internal/service"
)

func main() {
	// 1. 加载配置
	cfg, err := infraCfg.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 初始化数据库
	db, err := mysql.NewDB(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 3. 执行迁移
	migration, err := os.ReadFile("migrations/001_init.sql")
	if err == nil {
		if err := mysql.RunMigrations(db, string(migration)); err != nil {
			log.Printf("[warning] 迁移执行失败: %v", err)
		}
	}

	// 4. 初始化仓储层（infra 实现 domain 接口）
	catRepo := persistence.NewCategoryRepo(db)
	qRepo := persistence.NewQuestionRepo(db)
	tagRepo := persistence.NewTagRepo(db)
	answerRepo := persistence.NewAnswerRepo(db)

	// 5. 初始化服务层
	svc := service.New(catRepo, qRepo, tagRepo, answerRepo)

	// 6. 数据导入（幂等）
	seedFile := cfg.Seed.ContentFile
	if seedFile == "" {
		seedFile = "../resource-01/content.md"
	}
	if _, err := os.Stat(seedFile); err == nil {
		log.Println("[seed] 开始导入题目数据...")
		if err := persistence.ImportFromFile(db, seedFile); err != nil {
			log.Printf("[warning] 数据导入失败: %v", err)
		}
	} else {
		log.Printf("[seed] 未找到数据文件 %s，跳过导入", seedFile)
	}

	// 7. 初始化路由并启动
	r := http.SetupRouter(svc)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("InterviewHub 启动于 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
