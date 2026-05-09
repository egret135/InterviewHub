package service

import (
	domain "interview-hub/internal/domain/question"
)

type QuestionService struct {
	catRepo    domain.CategoryRepository
	qRepo      domain.QuestionRepository
	tagRepo    domain.TagRepository
	answerRepo domain.AnswerRepository
}

func New(catRepo domain.CategoryRepository, qRepo domain.QuestionRepository,
	tagRepo domain.TagRepository, answerRepo domain.AnswerRepository) *QuestionService {
	return &QuestionService{
		catRepo: catRepo, qRepo: qRepo, tagRepo: tagRepo, answerRepo: answerRepo,
	}
}

type QuestionListResult struct {
	Questions []domain.Question `json:"questions"`
	Total     int64             `json:"total"`
}

func (s *QuestionService) ListCategories() ([]domain.CategoryWithCount, error) {
	cats, err := s.catRepo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]domain.CategoryWithCount, len(cats))
	for i, c := range cats {
		result[i] = domain.CategoryWithCount{
			Category:      c,
			QuestionCount: s.catRepo.CountQuestions(c.ID),
		}
	}
	return result, nil
}

func (s *QuestionService) GetCategory(slug string) (*domain.Category, error) {
	return s.catRepo.GetBySlug(slug)
}

func (s *QuestionService) ListByCategory(slug string, page, size int) (*QuestionListResult, *domain.Category, error) {
	cat, err := s.catRepo.GetBySlug(slug)
	if err != nil {
		return nil, nil, err
	}
	questions, total, err := s.qRepo.ListByCategory(cat.ID, page, size)
	if err != nil {
		return nil, nil, err
	}
	return &QuestionListResult{Questions: questions, Total: total}, cat, nil
}

func (s *QuestionService) GetQuestion(id int) (*domain.Question, error) {
	return s.qRepo.GetByID(id)
}

func (s *QuestionService) Search(query string, page, size int) (*QuestionListResult, error) {
	if query == "" {
		return &QuestionListResult{}, nil
	}
	questions, total, err := s.qRepo.Search(query, page, size)
	if err != nil {
		return nil, err
	}
	return &QuestionListResult{Questions: questions, Total: total}, nil
}

func (s *QuestionService) ListByTag(slug string, page, size int) (*QuestionListResult, *domain.Tag, error) {
	tag, err := s.tagRepo.GetBySlug(slug)
	if err != nil {
		return nil, nil, err
	}
	questions, total, err := s.qRepo.ListByTag(slug, page, size)
	if err != nil {
		return nil, nil, err
	}
	return &QuestionListResult{Questions: questions, Total: total}, tag, nil
}

func (s *QuestionService) ListTags() ([]domain.Tag, error) {
	return s.tagRepo.ListAll()
}
