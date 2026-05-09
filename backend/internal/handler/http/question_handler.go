package http

import (
	"net/http"
	"strconv"

	"interview-hub/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QuestionHandler struct {
	svc *service.QuestionService
}

func NewQuestionHandler(svc *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: svc}
}

func (h *QuestionHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
}

func (h *QuestionHandler) GetCategory(c *gin.Context) {
	cat, err := h.svc.GetCategory(c.Param("slug"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "分类不存在"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cat})
}

func (h *QuestionHandler) ListByCategory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, cat, err := h.svc.ListByCategory(c.Param("slug"), page, size)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "分类不存在"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{"category": cat, "questions": result.Questions},
		"meta": gin.H{"page": page, "size": size, "total": result.Total},
	})
}

func (h *QuestionHandler) GetQuestion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "BAD_REQUEST", "message": "无效的题目ID"}})
		return
	}
	q, err := h.svc.GetQuestion(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "题目不存在"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": q})
}

func (h *QuestionHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, err := h.svc.Search(c.Query("q"), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result.Questions,
		"meta": gin.H{"page": page, "size": size, "total": result.Total},
	})
}

func (h *QuestionHandler) ListTags(c *gin.Context) {
	tags, err := h.svc.ListTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *QuestionHandler) ListByTag(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, tag, err := h.svc.ListByTag(c.Param("slug"), page, size)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "标签不存在"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{"tag": tag, "questions": result.Questions},
		"meta": gin.H{"page": page, "size": size, "total": result.Total},
	})
}
