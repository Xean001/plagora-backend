package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type CostConfigHandler struct {
	uc ucDomain.CostConfigUseCase
}

func NewCostConfigHandler(uc ucDomain.CostConfigUseCase) *CostConfigHandler {
	return &CostConfigHandler{uc: uc}
}

// GET /api/config/costos
func (h *CostConfigHandler) Get(c *gin.Context) {
	cfg, err := h.uc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// PUT /api/config/costos
func (h *CostConfigHandler) Update(c *gin.Context) {
	var input ucDomain.UpdateCostConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.uc.Update(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}
