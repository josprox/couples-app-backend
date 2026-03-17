package location

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jlhal/parejas/internal/relationships"
)

type Handler struct {
	service    Service
	relService relationships.Service
}

func NewHandler(service Service, relService relationships.Service) *Handler {
	return &Handler{service, relService}
}

func (h *Handler) GetPartnerLocation(c *gin.Context) {
	userID := c.GetString("user_id")
	
	rel, err := h.relService.GetMyRelationship(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "relationship not found"})
		return
	}

	partnerID := rel.User1ID
	if partnerID == userID {
		partnerID = rel.User2ID
	}

	loc, err := h.service.(*service).repo.GetLastLocation(partnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no location found for partner"})
		return
	}

	c.JSON(http.StatusOK, loc)
}
