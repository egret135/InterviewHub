package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"interview-hub/internal/infra/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Question struct {
	ID         uint `gorm:"primaryKey"`
	Title      string
	CategoryID uint
}

type Answer struct {
	ID         uint `gorm:"primaryKey"`
	QuestionID uint `gorm:"uniqueIndex"`
	ContentMD  string
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

var (
	configPath    = flag.String("config", "configs/config.yaml", "配置文件路径")
	apiKeyFlag    = flag.String("key", "", "DeepSeek API Key（覆盖配置文件）")
	modelFlag     = flag.String("model", "", "模型名称（覆盖配置文件）")
	delayFlag     = flag.Duration("delay", 0, "单 worker 调用间隔（覆盖配置文件）")
	retriesFlag   = flag.Int("retries", 0, "最大重试次数（覆盖配置文件）")
	concurrency   = flag.Int("concurrency", 100, "并发数")
	allFlag       = flag.Bool("all", false, "回答所有题目（包括已有答案的）")
	dryRun        = flag.Bool("dry-run", false, "仅打印 prompt，不调用 API")
	promptFile    = flag.String("prompt", "", "自定义 prompt 模板文件")
)

type task struct {
	q    Question
	idx  int
	total int
}

type result struct {
	qID     uint
	content string
	err     error
}

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	apiKey := cfg.DeepSeek.APIKey
	if *apiKeyFlag != "" {
		apiKey = *apiKeyFlag
	}
	if apiKey == "" {
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
	}
	if apiKey == "" {
		log.Fatal("请配置 DeepSeek API Key")
	}

	model := cfg.DeepSeek.Model
	if model == "" {
		model = "deepseek-chat"
	}
	if *modelFlag != "" {
		model = *modelFlag
	}

	delay := *delayFlag
	if delay == 0 {
		d, err := time.ParseDuration(cfg.DeepSeek.Delay)
		if err == nil {
			delay = d
		}
	}
	if delay == 0 {
		delay = 100 * time.Millisecond
	}

	maxRetries := *retriesFlag
	if maxRetries == 0 {
		maxRetries = cfg.DeepSeek.Retries
	}
	if maxRetries == 0 {
		maxRetries = 3
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	)

	log.Printf("数据库: %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	log.Printf("模型: %s, 并发: %d, 重试: %d", model, *concurrency, maxRetries)

	promptTemplate := defaultPrompt
	if *promptFile != "" {
		data, err := os.ReadFile(*promptFile)
		if err != nil {
			log.Fatalf("读取 prompt 文件失败: %v", err)
		}
		promptTemplate = string(data)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 查询题目
	var questions []Question
	if *allFlag {
		db.Raw("SELECT id, title, category_id FROM questions ORDER BY id").Scan(&questions)
	} else {
		db.Raw(`
			SELECT q.id, q.title, q.category_id
			FROM questions q
			LEFT JOIN answers a ON a.question_id = q.id
			WHERE a.id IS NULL
			ORDER BY q.id
		`).Scan(&questions)
	}

	if len(questions) == 0 {
		log.Println("所有题目都已有答案，无需处理（使用 --all 强制重新生成）")
		return
	}

	log.Printf("待处理: %d 道题目\n", len(questions))

	if *dryRun {
		for _, q := range questions {
			fmt.Printf("--- Q%d ---\n%s\n--- END ---\n", q.ID, buildPrompt(promptTemplate, q.Title))
		}
		return
	}

	// ---- 并发 worker 池 ----
	tasks := make(chan task, len(questions))
	results := make(chan result, len(questions))

	var success, fail atomic.Int64
	startTime := time.Now()

	// 启动 workers
	var wg sync.WaitGroup
	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go worker(w+1, &wg, apiKey, model, promptTemplate, maxRetries, delay, tasks, results, &success, &fail)
	}

	// 分发任务
	go func() {
		for i, q := range questions {
			tasks <- task{q: q, idx: i + 1, total: len(questions)}
		}
		close(tasks)
	}()

	// 收集结果并写入数据库
	go func() {
		wg.Wait()
		close(results)
	}()

	var logMu sync.Mutex
	for r := range results {
		if r.err != nil {
			logMu.Lock()
			log.Printf("[FAIL] Q%d: %v", r.qID, r.err)
			logMu.Unlock()
			continue
		}

		// Upsert answer
		a := &Answer{QuestionID: r.qID, ContentMD: r.content}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "question_id"}},
			UpdateAll: true,
		}).Create(a).Error; err != nil {
			logMu.Lock()
			log.Printf("[DB ERR] Q%d: %v", r.qID, err)
			logMu.Unlock()
			fail.Add(1)
			success.Add(-1)
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("完成: 成功 %d, 失败 %d, 耗时 %v", success.Load(), fail.Load(), elapsed.Round(time.Millisecond))
}

func worker(
	id int, wg *sync.WaitGroup,
	apiKey, model, promptTemplate string,
	maxRetries int, delay time.Duration,
	tasks <-chan task, results chan<- result,
	success, fail *atomic.Int64,
) {
	defer wg.Done()
	client := &http.Client{Timeout: 180 * time.Second}

	for t := range tasks {
		log.Printf("[W%d] [%d/%d] Q%d: %s", id, t.idx, t.total, t.q.ID, truncate(t.q.Title, 60))

		answer, err := askDeepSeek(client, apiKey, model, promptTemplate, t.q.Title, maxRetries)
		if err != nil {
			results <- result{qID: t.q.ID, err: err}
			fail.Add(1)
			continue
		}

		answer = cleanAnswer(answer)
		results <- result{qID: t.q.ID, content: answer}
		success.Add(1)

		time.Sleep(delay)
	}
}

func askDeepSeek(client *http.Client, apiKey, model, promptTemplate, question string, maxRetries int) (string, error) {
	prompt := buildPrompt(promptTemplate, question)

	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 429 {
			lastErr = fmt.Errorf("rate limited")
			time.Sleep(10 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("API %d: %s", resp.StatusCode, truncate(string(respBody), 200))
			continue
		}

		var chatResp ChatResponse
		if err := json.Unmarshal(respBody, &chatResp); err != nil {
			lastErr = err
			continue
		}

		if len(chatResp.Choices) == 0 {
			lastErr = fmt.Errorf("empty response")
			continue
		}

		return chatResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("retries exhausted: %w", lastErr)
}

func buildPrompt(template, question string) string {
	if strings.Contains(template, "{{.Question}}") {
		return strings.ReplaceAll(template, "{{.Question}}", question)
	}
	return fmt.Sprintf("%s\n\n问题：%s", template, question)
}

func cleanAnswer(text string) string {
	lines := strings.Split(text, "\n")
	var cleaned []string
	skip := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if skip && (strings.HasPrefix(trimmed, "##") || strings.HasPrefix(trimmed, "**") ||
			len(trimmed) > 0 && !strings.HasPrefix(trimmed, "问题") && !strings.HasPrefix(trimmed, "好的")) {
			skip = false
		}
		if !skip {
			cleaned = append(cleaned, line)
		}
	}
	if len(cleaned) == 0 {
		return text
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `answer-bot — 使用 DeepSeek AI 批量回答面试题

用法:
  answer-bot [选项]

配置优先级: 命令行参数 > 环境变量 > configs/config.yaml

选项:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
config.yaml 对应字段:
  deepseek:
    api_key: "sk-xxxx"     # DeepSeek API Key
    model: deepseek-chat   # 模型名称
    delay: 2s              # 单 worker 调用间隔
    retries: 3             # 失败重试次数

并发控制:
  --concurrency 100  同时并发的 API 请求数（默认 100）
  --all              重新生成所有题目的答案（包括已有答案的）

示例:
  # 补全缺失答案
  answer-bot

  # 重新生成所有答案（高并发）
  answer-bot --all

  # 自定义并发数
  answer-bot --all --concurrency 50
`)
	}
}

var defaultPrompt = `你是一位资深的 Go 后端面试官，也是技术专家。请用中文详细回答以下 Go 后端面试题。

要求：
1. 回答要专业、准确、全面
2. 如果涉及概念，先给出定义再展开细节
3. 如果有代码示例，请使用 Go 语言并用三个反引号包裹（标记为 go）
4. 使用 ## 小标题分段，提高可读性
5. 内容长度适中，既不过于简略也不过于冗长
6. 关键术语用 **粗体** 标注`
