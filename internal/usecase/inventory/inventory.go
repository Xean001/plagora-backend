package inventory

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type inventoryUC struct {
	repo repository.InventoryRepository
}

func New(repo repository.InventoryRepository) ucDomain.InventoryUseCase {
	return &inventoryUC{repo: repo}
}

func (uc *inventoryUC) Add(ctx context.Context, in ucDomain.AddToInventoryInput) (*entity.InventoryItem, error) {
	margin := 0.0
	if in.ProductionCost > 0 {
		margin = r2((in.SalePrice/in.ProductionCost - 1) * 100)
	}
	item := &entity.InventoryItem{
		PieceName:      in.PieceName,
		ProductionCost: in.ProductionCost,
		SuggestedPrice: in.SuggestedPrice,
		SuppliesCost:   in.SuppliesCost,
		SalePrice:      in.SalePrice,
		MarginPercent:  margin,
		Profit:         r2(in.SalePrice - in.ProductionCost),
		Status:         entity.InventoryPorVender,
		Notes:          in.Notes,
	}
	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *inventoryUC) GetAll(ctx context.Context, filter repository.InventoryFilter) ([]*entity.InventoryItem, error) {
	items, err := uc.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []*entity.InventoryItem{}, nil
	}
	return items, nil
}

func (uc *inventoryUC) Update(ctx context.Context, id uuid.UUID, in ucDomain.UpdateInventoryInput) (*entity.InventoryItem, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item.SalePrice = in.SalePrice
	item.Status = in.Status
	item.Notes = in.Notes

	if item.ProductionCost > 0 {
		item.MarginPercent = r2((in.SalePrice/item.ProductionCost - 1) * 100)
	}
	item.Profit = r2(in.SalePrice - item.ProductionCost)

	if in.Status == entity.InventoryVendido && item.SoldAt == nil {
		now := time.Now()
		item.SoldAt = &now
	}

	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *inventoryUC) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *inventoryUC) GetRevenue(ctx context.Context) (float64, error) {
	return uc.repo.GetRevenue(ctx)
}

func r2(v float64) float64 { return math.Round(v*100) / 100 }
