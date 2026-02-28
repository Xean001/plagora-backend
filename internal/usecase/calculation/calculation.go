package calculation

import (
	"context"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type calculationUseCase struct {
	repo repository.CalculationRepository
}

func New(repo repository.CalculationRepository) ucDomain.CalculationUseCase {
	return &calculationUseCase{repo: repo}
}

func (uc *calculationUseCase) Save(ctx context.Context, in ucDomain.SaveCalculationInput) (*entity.PriceCalculation, error) {
	profit := in.SalePrice - in.SubtotalProduction
	calc := &entity.PriceCalculation{
		PieceName:          in.PieceName,
		PrintHours:         in.PrintHours,
		PrintMinutesExtra:  in.PrintMinutesExtra,
		FilamentGrams:      in.FilamentGrams,
		SuppliesCost:       in.SuppliesCost,
		Multiplier:         in.Multiplier,
		MaterialCost:       in.MaterialCost,
		ElectricityCost:    in.ElectricityCost,
		MachineWear:        in.MachineWear,
		SubtotalProduction: in.SubtotalProduction,
		SuggestedPrice:     in.SuggestedPrice,
		TotalWithSupplies:  in.TotalWithSupplies,
		SalePrice:          in.SalePrice,
		Profit:             r2(profit),
		Notes:              in.Notes,
	}
	if err := uc.repo.Create(ctx, calc); err != nil {
		return nil, err
	}
	return calc, nil
}

func (uc *calculationUseCase) GetAll(ctx context.Context) ([]*entity.PriceCalculation, error) {
	return uc.repo.FindAll(ctx)
}

func (uc *calculationUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

func r2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
