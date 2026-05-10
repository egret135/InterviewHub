package question

import (
	"log"
	"os"
	"regexp"
	"strings"

	domain "interview-hub/internal/domain/question"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type parsedQuestion struct {
	title     string
	catIdx    int
	sortOrder int
}

var categoryConfigs = []struct {
	Name        string
	Slug        string
	Description string
	Icon        string
	Pattern     string
}{
	{"Go 语言核心与并发编程", "go-core-concurrency", "Goroutine、Channel、Context 原理与实践", "zap", `一、Go 语言核心与并发编程`},
	{"Go 内存管理与性能分析", "go-memory-performance", "GC 机制、内存逃逸、pprof 性能调优", "cpu", `二、Go 内存管理与性能分析`},
	{"微服务框架与工程化", "microservice-engineering", "go-zero / go-kit、单元测试与基准测试", "layers", `三、微服务框架与工程化`},
	{"关键三方库应用", "third-party-libs", "go-redis、sarama (Kafka)、gorm 实战", "package", `四、关键三方库应用`},
	{"MySQL 数据库与分库分表", "mysql-sharding", "索引优化、慢查询、分库分表、读写分离", "database", `五、MySQL 数据库与分库分表`},
	{"分布式系统与中间件", "distributed-systems", "Nacos、Kafka/RocketMQ、Redis 分布式锁", "globe", `六、分布式系统与中间件`},
	{"领域经验与场景设计", "domain-scenarios", "智能客服、工单系统、消息投递架构设计", "briefcase", `七、领域经验与场景设计`},
}

func ImportFromFile(db *gorm.DB, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	text := string(data)

	catRepo := &categoryRepo{db: db}
	qRepo := &questionRepo{db: db}
	tagRepo := &tagRepo{db: db}
	answerRepo := &answerRepo{db: db}

	sections := strings.Split(text, "\n---\n")
	questionSection := findQuestionListSection(sections)
	if questionSection == "" {
		questionSection = text
	}

	var parsedQuestions []parsedQuestion
	for i, cfg := range categoryConfigs {
		catQuestions := extractCategoryQuestions(questionSection, cfg.Pattern)
		for j, qTitle := range catQuestions {
			parsedQuestions = append(parsedQuestions, parsedQuestion{
				title: qTitle, catIdx: i, sortOrder: j + 1,
			})
		}
	}

	if len(parsedQuestions) == 0 {
		log.Println("[seed] 未解析到任何题目，跳过导入")
		return nil
	}

	// Extract answers by position (answers are numbered 1..N sequentially across all sections)
	answersByPosition := extractAnswersByPosition(text)

	log.Println("[seed] 导入分类...")
	for i, cfg := range categoryConfigs {
		cat := &domain.Category{
			Name: cfg.Name, Slug: cfg.Slug, Description: cfg.Description,
			Icon: cfg.Icon, SortOrder: i + 1,
		}
		catRepo.Upsert(cat)
	}

	allCats, _ := catRepo.ListAll()
	catMap := make(map[string]int)
	for _, c := range allCats {
		catMap[c.Slug] = c.ID
	}

	log.Printf("[seed] 导入 %d 道题目...\n", len(parsedQuestions))
	imported, skipped, withAnswer := 0, 0, 0
	for idx, pq := range parsedQuestions {
		catID := catMap[categoryConfigs[pq.catIdx].Slug]
		if catID == 0 {
			continue
		}

		difficulty := "medium"
		if strings.Contains(pq.title, "设计") || strings.Contains(pq.title, "架构") {
			difficulty = "hard"
		}

		q := &domain.Question{
			CategoryID: catID, Title: pq.title, Difficulty: difficulty, SortOrder: pq.sortOrder,
		}

		result := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "title"}, {Name: "category_id"}},
			DoNothing: true,
		}).Create(q)

		if result.RowsAffected == 0 {
			db.Where("title = ? AND category_id = ?", pq.title, catID).First(q)
			skipped++
		} else {
			imported++
		}

		qRepo.UpdateSearchVector(q.ID, pq.title)

		// Attach answer by position (1-indexed)
		if ans, ok := answersByPosition[idx+1]; ok && ans != "" {
			var existing domain.Answer
			if err := db.Where("question_id = ?", q.ID).First(&existing).Error; err != nil {
				answerRepo.Create(&domain.Answer{QuestionID: q.ID, ContentMD: ans})
				withAnswer++
			}
		}

		tags := extractTags(pq.title)
		for _, tagName := range tags {
			slug := slugify(tagName)
			t := &domain.Tag{Name: tagName, Slug: slug}
			if err := tagRepo.Upsert(t); err != nil {
				return err
			}
			tagRepo.LinkTag(q.ID, t.ID)
		}
	}

	log.Printf("[seed] 新导入 %d 题，跳过 %d 题（已存在），%d 题含答案", imported, skipped, withAnswer)
	return nil
}

func findQuestionListSection(sections []string) string {
	for _, s := range sections {
		if strings.Contains(s, "根据岗位描述") && strings.Contains(s, "一、Go 语言核心与并发编程") {
			return s
		}
	}
	for _, s := range sections {
		count := 0
		for _, cfg := range categoryConfigs {
			if strings.Contains(s, cfg.Pattern) {
				count++
			}
		}
		if count >= 4 {
			return s
		}
	}
	return ""
}

func extractCategoryQuestions(section string, pattern string) []string {
	idx := strings.Index(section, pattern)
	if idx < 0 {
		return nil
	}
	end := len(section)
	for _, cfg := range categoryConfigs {
		if cfg.Pattern == pattern {
			continue
		}
		if nextIdx := strings.Index(section[idx+len(pattern):], cfg.Pattern); nextIdx >= 0 {
			end = idx + len(pattern) + nextIdx
			break
		}
	}
	block := section[idx:end]
	lines := strings.Split(block, "\n")
	var questions []string
	seen := make(map[string]bool)
	skipPatterns := []string{"这些题目覆盖", "每个分类下", "注意：", "我将", "我会", "好的", "每个分类下3-5题"}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, pattern) || strings.HasPrefix(line, "分类") {
			continue
		}
		skip := false
		for _, sp := range skipPatterns {
			if strings.HasPrefix(line, sp) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if len(line) > 15 && !strings.HasPrefix(line, "根据") {
			cleaned := regexp.MustCompile(`^\d+[.、\s]+`).ReplaceAllString(line, "")
			cleaned = strings.TrimSpace(cleaned)
			if len(cleaned) > 15 && !seen[cleaned] {
				questions = append(questions, cleaned)
				seen[cleaned] = true
			}
		}
	}
	return questions
}

// extractAnswersByPosition finds all numbered answers in the text and returns them by position (1-indexed).
// Answers in content.md are numbered sequentially 1..N across all sections.
func extractAnswersByPosition(text string) map[int]string {
	answers := make(map[int]string)

	// Remove AI reasoning sections
	sections := strings.Split(text, "\n---\n")
	var answerText strings.Builder
	for _, s := range sections {
		if !isAIReasoning(s) {
			answerText.WriteString(s)
			answerText.WriteString("\n")
		}
	}

	lines := strings.Split(answerText.String(), "\n")
	re := regexp.MustCompile(`^(\d+)[.、\s]+`)
	var currentNum int
	var currentAnswer strings.Builder
	inAnswer := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inAnswer {
				currentAnswer.WriteString("\n")
			}
			continue
		}

		// Check if this line starts a numbered answer
		if matches := re.FindStringSubmatch(trimmed); matches != nil {
			num := 0
			for _, c := range matches[1] {
				num = num*10 + int(c-'0')
			}
			rest := re.ReplaceAllString(trimmed, "")
			rest = strings.TrimSpace(rest)

			// Only treat as a question header if it looks like a question (long enough)
			if len([]rune(rest)) > 10 {
				// Save previous answer
				if inAnswer && currentNum > 0 && currentAnswer.Len() > 0 {
					answers[currentNum] = strings.TrimSpace(currentAnswer.String())
				}
				currentNum = num
				currentAnswer.Reset()
				inAnswer = true
				continue
			}
		}

		if inAnswer {
			// Stop conditions
			if strings.HasPrefix(trimmed, "---") ||
				strings.HasPrefix(trimmed, "下面是对") ||
				strings.HasPrefix(trimmed, "我们需要") ||
				strings.HasPrefix(trimmed, "我们被要求") ||
				strings.HasPrefix(trimmed, "当然，下面") ||
				strings.HasPrefix(trimmed, "如果需要") {
				if currentNum > 0 && currentAnswer.Len() > 0 {
					answers[currentNum] = strings.TrimSpace(currentAnswer.String())
				}
				inAnswer = false
				currentNum = 0
				currentAnswer.Reset()
				continue
			}
			// Skip standalone code fence markers
			if trimmed == "go" || trimmed == "bash" || trimmed == "text" || trimmed == "java" ||
				trimmed == "复制" || trimmed == "下载" || trimmed == "json" || trimmed == "yaml" || trimmed == "sql" {
				continue
			}
			currentAnswer.WriteString(line)
			currentAnswer.WriteString("\n")
		}
	}

	// Save last answer
	if inAnswer && currentNum > 0 && currentAnswer.Len() > 0 {
		answers[currentNum] = strings.TrimSpace(currentAnswer.String())
	}

	return answers
}

func isAIReasoning(section string) bool {
	markers := []string{
		"我们需要理解用户请求", "我们需要解析", "我需要生成", "我将输出",
		"好的，开始编写回答", "注意避免", "我们被要求",
	}
	count := 0
	for _, m := range markers {
		if strings.Contains(section, m) {
			count++
		}
	}
	return count >= 2
}

func extractTags(title string) []string {
	tagMap := map[string][]string{
		"goroutine": {"goroutine", "并发"}, "GMP": {"goroutine", "GMP", "调度模型"},
		"map": {"map", "并发安全"}, "race": {"race-detector", "数据竞争"},
		"泄漏": {"goroutine泄漏", "性能排查"}, "context": {"context", "最佳实践"},
		"GC": {"GC", "垃圾回收"}, "逃逸": {"内存逃逸", "编译器优化"},
		"pprof": {"pprof", "性能分析"}, "sync.Pool": {"sync.Pool", "GC优化"},
		"go-zero": {"go-zero", "微服务框架"}, "mock": {"单元测试", "mock"},
		"基准测试": {"基准测试", "性能测试"}, "go-redis": {"go-redis", "Redis"},
		"分布式锁": {"分布式锁", "Redis", "Redlock"}, "sarama": {"sarama", "Kafka"},
		"gorm": {"gorm", "ORM"}, "慢查询": {"MySQL", "慢查询", "索引优化"},
		"覆盖索引": {"覆盖索引", "MySQL优化"},
		"分库分表": {"分库分表", "ShardingSphere", "MyCat"},
		"读写分离": {"读写分离", "主从延迟"},
		"Nacos": {"Nacos", "服务注册", "配置中心"},
		"RocketMQ": {"RocketMQ", "Kafka", "消息队列"},
		"缓存击穿": {"Redis", "缓存击穿", "缓存雪崩"},
		"Spring Boot": {"Java", "Spring Boot"}, "JVM": {"Java", "JVM", "性能分析"},
		"智能客服": {"智能客服", "系统设计", "IM"},
		"工单": {"工单系统", "状态机", "多角色"},
		"数据一致性": {"数据一致性", "分布式事务"},
		"channel": {"channel", "并发"}, "闭包": {"闭包", "内存逃逸"},
		"系统调用": {"系统调用", "阻塞"}, "调度": {"GMP", "调度模型"},
		"写屏障": {"写屏障", "三色标记"}, "GOGC": {"GOGC", "GC调优"},
		"GOMEMLIMIT": {"GOMEMLIMIT", "内存限制"},
		"Redlock": {"Redlock", "分布式锁"}, "Kafka": {"Kafka", "消息队列", "sarama"},
	}
	var tags []string
	for keyword, tagList := range tagMap {
		if strings.Contains(strings.ToLower(title), strings.ToLower(keyword)) {
			tags = append(tags, tagList...)
		}
	}
	seen := make(map[string]bool)
	var unique []string
	for _, t := range tags {
		if !seen[t] {
			seen[t] = true
			unique = append(unique, t)
		}
	}
	return unique
}

func slugify(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9一-鿿]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
