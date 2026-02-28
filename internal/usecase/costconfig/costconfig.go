package costconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type costConfigUseCase struct {
	repo repository.CostConfigRepository
}

func New(repo repository.CostConfigRepository) ucDomain.CostConfigUseCase {
	return &costConfigUseCase{repo: repo}
}

func (c *costConfigUseCase) Get(ctx context.Context) (*entity.CostConfig, error) {
	cfg, err := c.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting cost config: %w", err)
	}
	return cfg, nil
}

func (c *costConfigUseCase) Update(ctx context.Context, input ucDomain.UpdateCostConfigInput) (*entity.CostConfig, error) {
	cfg := &entity.CostConfig{
		ID:                  uuid.New(),
		FilamentPricePerKg:  input.FilamentPricePerKg,
		ElectricityKWhPrice: input.ElectricityKWhPrice,
		PrinterWattage:      input.PrinterWattage,
		PrinterPrice:        input.PrinterPrice,
		AmortizableHours:    input.AmortizableHours,
		SparePartsTotalCost: input.SparePartsTotalCost,
		SparePartsLifeHours: input.SparePartsLifeHours,
		MarginPercent:       input.MarginPercent,
		UpdatedAt:           time.Now(),
	}
	if err := c.repo.Upsert(ctx, cfg); err != nil {
		return nil, fmt.Errorf("upserting cost config: %w", err)
	}
	return cfg, nil
}
