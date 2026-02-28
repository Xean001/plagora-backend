package sale

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
	"github.com/plagora/backend/internal/usecase/calculator"
)

type saleUseCase struct {
	saleRepo repository.SaleRepository
	costRepo repository.CostConfigRepository
}

func New(saleRepo repository.SaleRepository, costRepo repository.CostConfigRepository) ucDomain.SaleUseCase {
	return &saleUseCase{saleRepo: saleRepo, costRepo: costRepo}
}

func (s *saleUseCase) GetAll(ctx context.Context, filter repository.SaleFilter) ([]*entity.Sale, error) {
	return s.saleRepo.FindAll(ctx, filter)
}

func (s *saleUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Sale, error) {
	sale, err := s.saleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sale %s not found: %w", id, err)
	}
	return sale, nil
}

func (s *saleUseCase) Create(ctx context.Context, userID uuid.UUID, input ucDomain.CreateSaleInput) (*entity.Sale, error) {
	cfg, err := s.costRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting cost config: %w", err)
	}

	breakdown := calculator.ComputeBreakdown(cfg, input.FilamentGrams, input.PrintTimeMinutes)

	status := input.Status
	if status == "" {
		status = entity.StatusPending
	}

	now := time.Now()
	sale := &entity.Sale{
		ID:                  uuid.New(),
		ClientID:            input.ClientID,
		UserID:              userID,
		Description:         input.Description,
		Material:            input.Material,
		Color:               input.Color,
		FilamentGrams:       breakdown.FilamentGrams,
		PrintTimeMinutes:    breakdown.PrintTimeMinutes,
		FilamentCost:        breakdown.FilamentCost,
		ElectricityCost:     breakdown.ElectricityCost,
		DepreciationCost:    breakdown.DepreciationCost,
		SparePartsCost:      breakdown.SparePartsCost,
		TotalProductionCost: breakdown.TotalProductionCost,
		ProfitMarginPercent: breakdown.MarginPercent,
		SuggestedPrice:      breakdown.SuggestedPrice,
		FinalPrice:          input.FinalPrice,
		Profit:              round2(input.FinalPrice - breakdown.TotalProductionCost),
		Status:              status,
		PaymentMethod:       input.PaymentMethod,
		Paid:                input.Paid,
		Notes:               input.Notes,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.saleRepo.Create(ctx, sale); err != nil {
		return nil, fmt.Errorf("creating sale: %w", err)
	}
	return sale, nil
}

func (s *saleUseCase) Update(ctx context.Context, id uuid.UUID, input ucDomain.UpdateSaleInput) (*entity.Sale, error) {
	sale, err := s.saleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sale %s not found: %w", id, err)
	}

	if input.ClientID != nil {
		sale.ClientID = input.ClientID
	}
	if input.Description != "" {
		sale.Description = input.Description
	}
	if input.Material != "" {
		sale.Material = input.Material
	}
	if input.Color != "" {
		sale.Color = input.Color
	}
	if input.FinalPrice != nil {
		sale.FinalPrice = *input.FinalPrice
		sale.Profit = round2(*input.FinalPrice - sale.TotalProductionCost)
	}
	if input.Status != "" {
		sale.Status = input.Status
	}
	if input.PaymentMethod != "" {
		sale.PaymentMethod = input.PaymentMethod
	}
	if input.Paid != nil {
		sale.Paid = *input.Paid
	}
	if input.Notes != "" {
		sale.Notes = input.Notes
	}
	sale.UpdatedAt = time.Now()

	if err := s.saleRepo.Update(ctx, sale); err != nil {
		return nil, fmt.Errorf("updating sale: %w", err)
	}
	return sale, nil
}

func (s *saleUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return s.saleRepo.Delete(ctx, id)
}

func (s *saleUseCase) GetDashboardStats(ctx context.Context) (*entity.DashboardStats, error) {
	return s.saleRepo.GetDashboardStats(ctx)
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
