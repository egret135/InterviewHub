package http

import (
	"interview-hub/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(svc *service.QuestionService) *gin.Engine {
	h := NewQuestionHandler(svc)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/categories", h.ListCategories)
		api.GET("/categories/:slug", h.GetCategory)
		api.GET("/categories/:slug/questions", h.ListByCategory)

		api.GET("/questions/search", h.Search)
		api.GET("/questions/:id", h.GetQuestion)

		api.GET("/tags", h.ListTags)
		api.GET("/tags/:slug/questions", h.ListByTag)
	}

	return r
}
