# InterviewHub 环境搭建指南

## 环境要求

- Go 1.21+
- Node.js 20+
- PostgreSQL 14+
- npm 或 pnpm

## 一、数据库配置

### 1.1 安装 PostgreSQL

macOS:
```bash
brew install postgresql@16
brew services start postgresql@16
```

### 1.2 创建数据库

```bash
psql -U postgres
```

```sql
CREATE DATABASE interview_hub;
\q
```

### 1.3 配置连接信息

后端通过 `backend/config.yaml` 读取数据库配置：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: 你的密码
  dbname: interview_hub
  sslmode: disable
```

也可以使用环境变量（优先级高于配置文件）：

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=你的密码
export DB_NAME=interview_hub
```

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_HOST` | localhost | 数据库地址 |
| `DB_PORT` | 5432 | 数据库端口 |
| `DB_USER` | postgres | 数据库用户 |
| `DB_PASSWORD` | postgres | 数据库密码 |
| `DB_NAME` | interview_hub | 数据库名 |
| `DB_SSLMODE` | disable | SSL 模式 |

## 二、启动方式

### 2.1 安装后端依赖

```bash
cd backend
go mod tidy
```

### 2.2 安装前端依赖

```bash
cd frontend
npm install
```

### 2.3 一键启动（推荐）

```bash
./start.sh          # 启动前后端
./start.sh stop     # 停止
./start.sh restart  # 重启
./start.sh status   # 查看状态
./start.sh logs     # 查看日志
```

也可以单独管理：

```bash
./start.sh backend start    # 仅启动后端
./start.sh frontend start   # 仅启动前端
```

### 2.4 手动启动后端

```bash
cd backend
go run cmd/http-server/main.go
```

服务启动时会自动：
1. 连接 PostgreSQL 数据库
2. 执行 `migrations/001_init.sql` 建表（categories、questions、answers、tags、question_tags）
3. 读取 `resource-01/content.md` 解析题目数据
4. 幂等导入：已存在的分类/题目/答案通过 `slug` 去重，不会重复插入
5. 监听 `:17001` 端口

```
[seed] 导入分类...
[seed] 导入 32 道题目...
[seed] 数据导入完成
InterviewHub 启动于 http://localhost:17001
```

### 2.5 启动前端

```bash
cd frontend
npm run dev
```

开发服务器默认运行在 `http://localhost:17000`，API 请求通过 Vite proxy 自动转发到后端 17001 端口。

### 2.6 生产部署

构建前端静态文件：

```bash
cd frontend
npm run build    # 输出到 dist/
```

将 `dist/` 目录部署到 Nginx 或 CDN，后端直接运行编译好的二进制：

```bash
cd backend
go build -o interview-hub cmd/http-server/main.go
./interview-hub --config config.yaml
```

## 三、验证

### 3.1 检查后端健康状态

```bash
curl http://localhost:17001/api/health
# { "status": "ok" }
```

### 3.2 检查分类接口

```bash
curl http://localhost:17001/api/v1/categories | jq '.data | length'
# 8
```

### 3.3 检查题目接口

```bash
curl http://localhost:17001/api/v1/categories/go-core-concurrency/questions | jq '.data.questions | length'
# 5
```

### 3.4 检查单题详情

```bash
curl http://localhost:17001/api/v1/questions/1 | jq '.data | {title, answer: .answer.content_md[:50]}'
```

### 3.5 检查全文搜索

```bash
curl "http://localhost:17001/api/v1/questions/search?q=goroutine" | jq '.meta.total'
```

### 3.6 前端页面检查

打开浏览器访问以下页面：

| 页面 | URL | 预期 |
|------|-----|------|
| 首页 | `http://localhost:17000/` | 8 个分类卡片 + 搜索栏 |
| 分类详情 | `http://localhost:17000/category/go-core-concurrency` | 5 道题目，点击展开答案 |
| 搜索 | `http://localhost:17000/search?q=goroutine` | 相关题目列表，答案默认展开 |

## 四、重新导入数据

如果需要清空并重新导入题目数据：

```bash
psql -U postgres -d interview_hub
```

```sql
TRUNCATE question_tags, answers, questions, tags, categories CASCADE;
\q
```

重启后端即可自动重新导入。

## 五、项目结构速览

```
InterviewHub/
├── backend/
│   ├── api/model/model.go                  # 公共 API 模型
│   ├── cmd/http-server/main.go            # 服务入口
│   ├── configs/config.yaml                # 配置文件
│   ├── internal/
│   │   ├── domain/question/               # 领域层：数据模型 + 仓储接口
│   │   │   ├── model.go
│   │   │   └── repo.go
│   │   ├── service/question_svc.go        # 应用服务层
│   │   ├── handler/http/                  # 协议适配层：HTTP 处理器 + 路由
│   │   │   ├── question_handler.go
│   │   │   └── router.go
│   │   └── infra/                         # 基础设施层
│   │       ├── config/config.go           # 配置加载（viper）
│   │       ├── mysql/client.go            # 数据库连接（gorm）
│   │       └── persistence/question/      # 仓储实现 + 数据导入
│   │           ├── repo_impl.go
│   │           └── seed.go
│   ├── migrations/001_init.sql            # DDL
│   └── pkg/errno/errno.go                 # 错误码
├── frontend/                               # React SPA
│   ├── src/
│   │   ├── components/                    # UI 组件
│   │   ├── pages/                         # HomePage, CategoryPage, SearchPage
│   │   ├── hooks/                         # 自定义 hooks
│   │   ├── api/index.ts                   # API 封装
│   │   └── types/index.ts                 # TypeScript 类型
│   └── vite.config.ts                     # 含 API 代理
├── resource-01/content.md                 # 原始题目数据
└── docs/
    ├── design.md                          # 设计方案
    ├── setup.md                           # 本文件
    └── 后端项目目录结构定义.md               # 目录结构规范
```
