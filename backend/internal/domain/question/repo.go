package question

// CategoryRepository 分类仓储接口（依赖倒置，由 infra 层实现）
type CategoryRepository interface {
	ListAll() ([]Category, error)
	GetBySlug(slug string) (*Category, error)
	Upsert(c *Category) error
	CountQuestions(categoryID int) int64
}

// QuestionRepository 题目仓储接口
type QuestionRepository interface {
	ListByCategory(categoryID int, page, size int) ([]Question, int64, error)
	GetByID(id int) (*Question, error)
	Search(query string, page, size int) ([]Question, int64, error)
	ListByTag(tagSlug string, page, size int) ([]Question, int64, error)
	Create(q *Question) error
	UpdateSearchVector(qID int, title string) error
}

// TagRepository 标签仓储接口
type TagRepository interface {
	ListAll() ([]Tag, error)
	GetBySlug(slug string) (*Tag, error)
	Upsert(t *Tag) error
	LinkTag(questionID, tagID int) error
}

// AnswerRepository 答案仓储接口
type AnswerRepository interface {
	Create(a *Answer) error
}
