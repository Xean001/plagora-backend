package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/plagora/backend/internal/delivery/http/middleware"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type SaleHandler struct {
	saleUC ucDomain.SaleUseCase
	calcUC ucDomain.CalculatorUseCase
}

func NewSaleHandler(saleUC ucDomain.SaleUseCase, calcUC ucDomain.CalculatorUseCase) *SaleHandler {
	return &SaleHandler{saleUC: saleUC, calcUC: calcUC}
}

// POST /api/ventas/calcular — preview cost without saving
func (h *SaleHandler) Calculate(c *gin.Context) {
	var input ucDomain.CalculateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	breakdown, err := h.calcUC.Calculate(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, breakdown)
}

// GET /api/ventas
func (h *SaleHandler) GetAll(c *gin.Context) {
	filter := repository.SaleFilter{}

	if s := c.Query("status"); s != "" {
		st := entity.SaleStatus(s)
		filter.Status = &st
	}
	if p := c.Query("paid"); p == "true" {
		b := true
		filter.Paid = &b
	} else if p == "false" {
		b := false
		filter.Paid = &b
	}
	if cid := c.Query("client_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err == nil {
			filter.ClientID = &id
		}
	}
	if from := c.Query("from"); from != "" {
		filter.FromDate = &from
	}
	if to := c.Query("to"); to != "" {
		filter.ToDate = &to
	}

	sales, err := h.saleUC.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sales == nil {
		sales = []*entity.Sale{}
	}
	c.JSON(http.StatusOK, sales)
}

// GET /api/ventas/:id
func (h *SaleHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	sale, err := h.saleUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sale)
}

// POST /api/ventas
func (h *SaleHandler) Create(c *gin.Context) {
	var input ucDomain.CreateSaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	sale, err := h.saleUC.Create(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sale)
}

// PUT /api/ventas/:id
func (h *SaleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input ucDomain.UpdateSaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sale, err := h.saleUC.Update(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sale)
}

// DELETE /api/ventas/:id
func (h *SaleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.saleUC.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /api/dashboard/stats
func (h *SaleHandler) DashboardStats(c *gin.Context) {
	stats, err := h.saleUC.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
