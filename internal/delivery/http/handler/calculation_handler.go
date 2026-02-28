package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type CalculationHandler struct {
	uc ucDomain.CalculationUseCase
}

func NewCalculationHandler(uc ucDomain.CalculationUseCase) *CalculationHandler {
	return &CalculationHandler{uc: uc}
}

// POST /api/calculadora — guarda un cálculo
func (h *CalculationHandler) Save(c *gin.Context) {
	var input ucDomain.SaveCalculationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	calc, err := h.uc.Save(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, calc)
}

// GET /api/calculadora — lista cálculos guardados
func (h *CalculationHandler) GetAll(c *gin.Context) {
	calcs, err := h.uc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if calcs == nil {
		calcs = []*entity.PriceCalculation{}
	}
	c.JSON(http.StatusOK, calcs)
}

// DELETE /api/calculadora/:id
func (h *CalculationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
