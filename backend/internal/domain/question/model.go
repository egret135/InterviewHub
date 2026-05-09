package question

import "time"

type Category struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Slug        string    `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Description string    `json:"description"`
	Icon        string    `json:"icon" gorm:"size:50"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Category) TableName() string { return "categories" }

type Question struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	CategoryID int       `json:"category_id" gorm:"not null"`
	Title      string    `json:"title" gorm:"size:500;not null"`
	Difficulty  string    `json:"difficulty" gorm:"size:20;default:medium"`
	SortOrder  int       `json:"sort_order" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Answer   *Answer   `json:"answer,omitempty" gorm:"foreignKey:QuestionID"`
	Tags     []Tag     `json:"tags,omitempty" gorm:"many2many:question_tags;"`
}

func (Question) TableName() string { return "questions" }

type Answer struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	QuestionID int       `json:"question_id" gorm:"uniqueIndex;not null"`
	ContentMD  string    `json:"content_md" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Answer) TableName() string { return "answers" }

type Tag struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:50;uniqueIndex;not null"`
	Slug string `json:"slug" gorm:"size:50;uniqueIndex;not null"`
}

func (Tag) TableName() string { return "tags" }

// CategoryWithCount is used for API responses
type CategoryWithCount struct {
	Category
	QuestionCount int64 `json:"question_count"`
}
