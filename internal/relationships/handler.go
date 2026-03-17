package relationships

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

func (h *Handler) SendRequest(c *gin.Context) {
	userID := c.GetString("user_id")

	var req SendRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.SendRequest(userID, req.ReceiverEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Relationship request sent"})
}

func (h *Handler) AcceptRequest(c *gin.Context) {
	userID := c.GetString("user_id")
	requestID := c.Param("id")

	err := h.service.AcceptRequest(userID, requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Relationship request accepted"})
}

func (h *Handler) GetDashboard(c *gin.Context) {
	userID := c.GetString("user_id")

	rel, err := h.service.GetMyRelationship(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rel)
}

func (h *Handler) UpdateWizard(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateWizardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateWizard(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wizard updated successfully"})
}
func (h *Handler) GetRequests(c *gin.Context) {
	userID := c.GetString("user_id")

	requests, err := h.service.GetPendingRequests(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}
