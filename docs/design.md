# InterviewHub 面试题库 — 设计方案

## 一、概述

基于 `resource-01/content.md` 中的 DeepSeek 生成的 Go 后端面试题（8 大类、30+ 题目，含详细答案），构建一个简洁优雅的面试题库网站。

- **前端**：React + TypeScript + Tailwind CSS + react-markdown + shiki（代码高亮）
- **后端**：Go + gin + gorm
- **数据库**：PostgreSQL

---

## 二、数据模型

```sql
-- 题目分类
CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,      -- 如 "Go 语言核心与并发编程"
    slug        VARCHAR(100) NOT NULL UNIQUE, -- URL 友好标识
    description TEXT,
    icon        VARCHAR(50),                 -- 图标标识（emoji 或 icon key）
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 题目
CREATE TABLE questions (
    id          SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id),
    title       VARCHAR(500) NOT NULL,       -- 题目原文
    difficulty   VARCHAR(20) NOT NULL DEFAULT 'medium',  -- easy/medium/hard
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 答案（1:1 关联题目，Markdown 格式）
CREATE TABLE answers (
    id           SERIAL PRIMARY KEY,
    question_id  INT NOT NULL UNIQUE REFERENCES questions(id),
    content_md   TEXT NOT NULL,              -- Markdown 正文（含代码块）
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 标签
CREATE TABLE tags (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE
);

-- 题目标签关联（多对多）
CREATE TABLE question_tags (
    question_id INT NOT NULL REFERENCES questions(id),
    tag_id      INT NOT NULL REFERENCES tags(id),
    PRIMARY KEY (question_id, tag_id)
);
```

### 8 个分类

| # | 分类 | slug | 题数 |
|---|------|------|------|
| 1 | Go 语言核心与并发编程 | go-core-concurrency | 5 |
| 2 | Go 内存管理与性能分析 | go-memory-performance | 4 |
| 3 | 微服务框架与工程化 | microservice-engineering | 4 |
| 4 | 关键三方库应用 | third-party-libs | 4 |
| 5 | MySQL 数据库与分库分表 | mysql-sharding | 4 |
| 6 | 分布式系统与中间件 | distributed-systems | 4 |
| 7 | Java 基础与跨语言理解 | java-basics | 3 |
| 8 | 领域经验与场景设计 | domain-scenarios | 4 |

### ER 关系

```
categories  ──1:N──>  questions  ──1:1──>  answers
                         │
                    M:N  │
                    question_tags
                         │
                         └────  tags
```

---

## 三、后端 API

### 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/categories` | 所有分类（含题目计数） |
| `GET` | `/api/v1/categories/:slug/questions` | 按分类获取题目（支持 `?page=&size=` 分页） |
| `GET` | `/api/v1/questions/:id` | 单题详情（含答案） |
| `GET` | `/api/v1/questions/search?q=xxx` | 全文搜索 |
| `GET` | `/api/v1/tags` | 所有标签 |
| `GET` | `/api/v1/tags/:slug/questions` | 按标签获取题目 |

### 技术选型

| 组件 | 选择 |
|------|------|
| HTTP 框架 | `gin` |
| ORM | `gorm` (v2) |
| 配置 | `viper`，`config.yaml` |
| 数据库驱动 | `pgx`（通过 gorm 间接使用） |
| 数据初始化 | 启动时从 `resource-01/content.md` 解析导入（幂等，slug 去重） |
| 分层 | `handler → service → repository` |

### 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go            # 入口
├── internal/
│   ├── config/
│   │   └── config.go          # 配置加载
│   ├── model/
│   │   ├── category.go
│   │   ├── question.go
│   │   ├── answer.go
│   │   └── tag.go
│   ├── repository/
│   │   ├── category_repo.go
│   │   ├── question_repo.go
│   │   └── tag_repo.go
│   ├── service/
│   │   ├── category_svc.go
│   │   ├── question_svc.go
│   │   └── tag_svc.go
│   ├── handler/
│   │   ├── category_handler.go
│   │   ├── question_handler.go
│   │   └── tag_handler.go
│   ├── router/
│   │   └── router.go          # 路由注册
│   └── seed/
│       └── seed.go            # content.md 解析与导入
├── migrations/
│   └── 001_init.sql           # 建表 DDL
├── go.mod
└── config.yaml
```

### PostgreSQL 全文搜索

利用 `tsvector` / `tsquery` 实现中文搜索：

```sql
-- questions 表增加 tsvector 列
ALTER TABLE questions ADD COLUMN search_vector tsvector;

-- 创建 GIN 索引
CREATE INDEX idx_questions_search ON questions USING GIN(search_vector);

-- 搜索时用 ts_rank 排序
SELECT q.* FROM questions q
WHERE q.search_vector @@ plainto_tsquery('simple', 'goroutine 泄漏')
ORDER BY ts_rank(q.search_vector, plainto_tsquery('simple', 'goroutine 泄漏')) DESC;
```

---

## 四、前端设计

### 技术栈

| 组件 | 选择 |
|------|------|
| 框架 | React 18 + TypeScript |
| 构建 | Vite |
| 样式 | Tailwind CSS |
| Markdown 渲染 | react-markdown + remark-gfm |
| 代码高亮 | shiki（构建时生成，零运行时体积） |
| HTTP 客户端 | fetch / 自定义 hook |

### 项目结构

```
frontend/
├── src/
│   ├── components/
│   │   ├── Layout.tsx           # 全局布局（顶栏 + 内容区）
│   │   ├── CategoryCard.tsx     # 分类卡片
│   │   ├── QuestionCard.tsx     # 题目卡片（可展开答案）
│   │   ├── SearchBar.tsx        # 搜索栏
│   │   ├── CodeBlock.tsx        # 代码块（shiki 渲染）
│   │   └── DifficultyBadge.tsx  # 难度标签
│   ├── pages/
│   │   ├── HomePage.tsx         # 首页（分类网格）
│   │   ├── CategoryPage.tsx     # 分类详情（题目列表）
│   │   └── SearchPage.tsx       # 搜索结果
│   ├── api/
│   │   └── index.ts             # API 请求封装
│   ├── hooks/
│   │   ├── useCategories.ts
│   │   ├── useQuestions.ts
│   │   └── useSearch.ts
│   ├── types/
│   │   └── index.ts             # TypeScript 类型定义
│   ├── App.tsx
│   └── main.tsx
├── index.html
├── tailwind.config.ts
├── tsconfig.json
├── vite.config.ts
└── package.json
```

### 视觉设计

**配色方案 — 暗色 + 青色 accent**

| 用途 | 色值 | 说明 |
|------|------|------|
| 背景 | `#0b0f19` | 深蓝黑主背景 |
| 面板 | `#111827` + 半透明 | 卡片、侧边栏 |
| 边框 | `rgba(0,229,255,0.10)` | 微青调边框 |
| 主色 | `#00e5ff` | 链接、选中态、accent |
| 文字主 | `#e2e8f0` | 正文 |
| 文字次 | `#94a3b8` | 描述、元信息 |
| 代码背景 | `#0d1117` | 代码块 |
| 难度-简单 | `#22c55e` | 绿色 |
| 难度-中等 | `#f59e0b` | 琥珀色 |
| 难度-困难 | `#ef4444` | 红色 |

**字体**
- 正文：`Inter`（系统无衬线回退）
- 代码：`JetBrains Mono` / `Fira Code`（等宽）

**卡片效果**
- `border-radius: 12px`
- 半透明背景 + `backdrop-filter: blur(8px)`
- `border: 1px solid rgba(0,229,255,0.08)`
- hover 时 `transform: translateY(-2px)` + 边框发亮

### 页面设计

#### 首页

```
┌──────────────────────────────────────────────────────────┐
│  ██  InterviewHub                                       │
│     简洁优雅的面试题库                                      │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  🔍  搜索题目、分类或标签...                          │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  ⚡           │  │  🔬           │  │  🏗            │   │
│  │  Go 核心与    │  │  内存管理与   │  │  微服务框架   │   │
│  │  并发编程     │  │  性能分析     │  │  与工程化     │   │
│  │     5 题      │  │     4 题      │  │     4 题      │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  📦           │  │  🗄            │  │  🌐           │   │
│  │  关键三方库   │  │  MySQL 与     │  │  分布式系统   │   │
│  │  应用         │  │  分库分表     │  │  与中间件     │   │
│  │     4 题      │  │     4 题      │  │     4 题      │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  ☕           │  │  💼           │                    │
│  │  Java 基础与  │  │  领域经验与   │                    │
│  │  跨语言理解   │  │  场景设计     │                    │
│  │     3 题      │  │     4 题      │                    │
│  └──────────────┘  └──────────────┘                     │
└──────────────────────────────────────────────────────────┘
```

#### 分类详情页

```
┌──────────────────────────────────────────────────────────┐
│  ← 返回首页             Go 语言核心与并发编程 · 5 题       │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Q1  goroutine 的调度模型（GMP），系统调用时 M 和    │  │
│  │      P 会如何变化？                          高级   │  │
│  │                                                     │  │
│  │  [展开答案]                                          │  │
│  │  ┌─────────────────────────────────────────────┐    │  │
│  │  │ GMP 模型中，G 代表 goroutine，M 代表 OS 线程， │   │  │
│  │  │ P 代表调度器上下文...                         │   │  │
│  │  │                                              │   │  │
│  │  │ ```go                                         │   │  │
│  │  │ // 示例代码...                                 │   │  │
│  │  │ ```                                           │   │  │
│  │  └─────────────────────────────────────────────┘    │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Q2  map 并发读写的安全问题及解决方案         高级   │  │
│  │  [展开答案]                                          │  │
│  └────────────────────────────────────────────────────┘  │
│  ...                                                     │
└──────────────────────────────────────────────────────────┘
```

#### 搜索页

- 输入关键词后，实时过滤题目和分类
- 搜索结果高亮匹配文字
- 按相关度 + 分类分组展示

### 路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | HomePage | 分类网格 + 搜索入口 |
| `/category/:slug` | CategoryPage | 分类下的题目列表 |
| `/search?q=xxx` | SearchPage | 搜索结果 |

---

## 五、数据初始化策略

`backend/internal/seed/seed.go` 负责在服务启动时：

1. 读取 `resource-01/content.md`
2. 按 `---` 分隔线切分内容块
3. 识别分类标题（如 "一、Go 语言核心与并发编程"）
4. 提取问题文本（如 "1. 请详细说明 goroutine 的调度模型..."）
5. 匹配对应的答案段落
6. 通过 slug 判断是否已导入，已存在则跳过（幂等）

---

## 六、API 响应格式

统一响应结构：

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "size": 10,
    "total": 30
  }
}
```

错误响应：

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "分类不存在"
  }
}
```

### 示例响应

**GET `/api/v1/categories`**

```json
{
  "data": [
    {
      "id": 1,
      "name": "Go 语言核心与并发编程",
      "slug": "go-core-concurrency",
      "description": "Goroutine、Channel、Context 原理与实践",
      "icon": "zap",
      "question_count": 5
    }
  ]
}
```

**GET `/api/v1/questions/1`**

```json
{
  "data": {
    "id": 1,
    "category": { "id": 1, "name": "Go 语言核心与并发编程", "slug": "go-core-concurrency" },
    "title": "请详细说明 goroutine 的调度模型（GMP），当发生系统调用时，M 和 P 会如何变化？",
    "difficulty": "hard",
    "answer": {
      "content_md": "## GMP 模型\n\nG（Goroutine）代表一个 goroutine...\n\n```go\n// 示例\n```"
    },
    "tags": [
      { "id": 1, "name": "goroutine", "slug": "goroutine" },
      { "id": 2, "name": "GMP", "slug": "gmp" }
    ]
  }
}
```
