package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type InventoryHandler struct{ uc ucDomain.InventoryUseCase }

func NewInventoryHandler(uc ucDomain.InventoryUseCase) *InventoryHandler {
	return &InventoryHandler{uc: uc}
}

// POST /api/inventario
func (h *InventoryHandler) Add(c *gin.Context) {
	var input ucDomain.AddToInventoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.uc.Add(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// GET /api/inventario?search=&status=&sort=price_desc
func (h *InventoryHandler) GetAll(c *gin.Context) {
	filter := repository.InventoryFilter{}
	if s := c.Query("search"); s != "" {
		filter.Search = &s
	}
	if st := c.Query("status"); st != "" {
		status := entity.InventoryStatus(st)
		filter.Status = &status
	}
	filter.SortBy = c.DefaultQuery("sort", "created_desc")

	items, err := h.uc.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// PUT /api/inventario/:id
func (h *InventoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input ucDomain.UpdateInventoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.uc.Update(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DELETE /api/inventario/:id
func (h *InventoryHandler) Delete(c *gin.Context) {
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

// GET /api/inventario/revenue
func (h *InventoryHandler) Revenue(c *gin.Context) {
	rev, err := h.uc.GetRevenue(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"revenue": rev})
}
