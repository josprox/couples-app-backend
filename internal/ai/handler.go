package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

type PromptRequest struct {
	Message string `json:"message" binding:"required"`
}

func (h *Handler) PromptAI(c *gin.Context) {
	userID := c.GetString("user_id")

	var req PromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This is commonly launched asynchronously so it doesn't block the caller.
	// We'll execute it synchronously here to return immediate error if needed,
	// or you can do a `go h.service.ProcessAIPrompt` and return 202 Accepted.
	
	err := h.service.ProcessAIPrompt(userID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI is generating a response"})
}
