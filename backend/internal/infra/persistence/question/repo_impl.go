package question

import (
	"fmt"

	domain "interview-hub/internal/domain/question"

	"gorm.io/gorm"
)

// --- CategoryRepo ---

type categoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) ListAll() ([]domain.Category, error) {
	var cats []domain.Category
	err := r.db.Order("sort_order").Find(&cats).Error
	return cats, err
}

func (r *categoryRepo) GetBySlug(slug string) (*domain.Category, error) {
	var c domain.Category
	err := r.db.Where("slug = ?", slug).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) Upsert(c *domain.Category) error {
	return r.db.Where("slug = ?", c.Slug).Assign(c).FirstOrCreate(c).Error
}

func (r *categoryRepo) CountQuestions(categoryID int) int64 {
	var count int64
	r.db.Model(&domain.Question{}).Where("category_id = ?", categoryID).Count(&count)
	return count
}

// --- QuestionRepo ---

type questionRepo struct{ db *gorm.DB }

func NewQuestionRepo(db *gorm.DB) domain.QuestionRepository {
	return &questionRepo{db: db}
}

func (r *questionRepo) ListByCategory(categoryID int, page, size int) ([]domain.Question, int64, error) {
	var questions []domain.Question
	var total int64

	query := r.db.Model(&domain.Question{}).Where("category_id = ?", categoryID)
	query.Count(&total)

	err := query.
		Preload("Category").
		Preload("Tags").
		Order("sort_order").
		Offset((page - 1) * size).
		Limit(size).
		Find(&questions).Error

	return questions, total, err
}

func (r *questionRepo) GetByID(id int) (*domain.Question, error) {
	var q domain.Question
	err := r.db.
		Preload("Category").
		Preload("Answer").
		Preload("Tags").
		First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *questionRepo) Search(query string, page, size int) ([]domain.Question, int64, error) {
	var questions []domain.Question
	var total int64

	baseQuery := r.db.Model(&domain.Question{}).
		Joins("JOIN categories ON categories.id = questions.category_id").
		Where(
			"questions.search_vector @@ plainto_tsquery('simple', ?) OR questions.title ILIKE ?",
			query, "%"+query+"%",
		)

	baseQuery.Count(&total)

	err := baseQuery.
		Preload("Category").
		Preload("Tags").
		Order(fmt.Sprintf("ts_rank(questions.search_vector, plainto_tsquery('simple', '%s')) DESC", query)).
		Offset((page - 1) * size).
		Limit(size).
		Find(&questions).Error

	return questions, total, err
}

func (r *questionRepo) ListByTag(tagSlug string, page, size int) ([]domain.Question, int64, error) {
	var questions []domain.Question
	var total int64

	baseQuery := r.db.Model(&domain.Question{}).
		Joins("JOIN question_tags ON question_tags.question_id = questions.id").
		Joins("JOIN tags ON tags.id = question_tags.tag_id").
		Where("tags.slug = ?", tagSlug)

	baseQuery.Count(&total)

	err := baseQuery.
		Preload("Category").
		Preload("Tags").
		Order("questions.sort_order").
		Offset((page - 1) * size).
		Limit(size).
		Find(&questions).Error

	return questions, total, err
}

func (r *questionRepo) Create(q *domain.Question) error {
	return r.db.Create(q).Error
}

func (r *questionRepo) UpdateSearchVector(qID int, title string) error {
	return r.db.Exec(
		"UPDATE questions SET search_vector = to_tsvector('simple', ?) WHERE id = ?",
		title, qID,
	).Error
}

// --- TagRepo ---

type tagRepo struct{ db *gorm.DB }

func NewTagRepo(db *gorm.DB) domain.TagRepository {
	return &tagRepo{db: db}
}

func (r *tagRepo) ListAll() ([]domain.Tag, error) {
	var tags []domain.Tag
	err := r.db.Order("name").Find(&tags).Error
	return tags, err
}

func (r *tagRepo) GetBySlug(slug string) (*domain.Tag, error) {
	var t domain.Tag
	err := r.db.Where("slug = ?", slug).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tagRepo) Upsert(t *domain.Tag) error {
	return r.db.Where("slug = ?", t.Slug).Assign(t).FirstOrCreate(t).Error
}

func (r *tagRepo) LinkTag(questionID, tagID int) error {
	return r.db.Exec(
		"INSERT INTO question_tags (question_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		questionID, tagID,
	).Error
}

// --- AnswerRepo ---

type answerRepo struct{ db *gorm.DB }

func NewAnswerRepo(db *gorm.DB) domain.AnswerRepository {
	return &answerRepo{db: db}
}

func (r *answerRepo) Create(a *domain.Answer) error {
	return r.db.Create(a).Error
}
