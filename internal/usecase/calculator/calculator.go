package calculator

import (
	"context"
	"fmt"

	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type calculatorUseCase struct {
	costRepo repository.CostConfigRepository
}

func New(costRepo repository.CostConfigRepository) ucDomain.CalculatorUseCase {
	return &calculatorUseCase{costRepo: costRepo}
}

func (c *calculatorUseCase) Calculate(ctx context.Context, input ucDomain.CalculateInput) (*entity.CostBreakdown, error) {
	cfg, err := c.costRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting cost config: %w", err)
	}

	return compute(cfg, input.FilamentGrams, input.PrintTimeMinutes), nil
}

// compute performs the full cost calculation with step-by-step explanation.
// It is also used by the sale use case via ComputeBreakdown.
func compute(cfg *entity.CostConfig, filamentGrams float64, printTimeMinutes int) *entity.CostBreakdown {
	hours := float64(printTimeMinutes) / 60.0

	// ─── STEP 1: Costo de electricidad por hora ───────────────────────────────
	// Convertimos W → kWh: (Watts / 1000) × precio_kWh
	electricityPerHour := (cfg.PrinterWattage / 1000.0) * cfg.ElectricityKWhPrice

	// ─── STEP 2: Depreciación de la impresora por hora ───────────────────────
	// precio_impresora / horas_amortizables
	depreciationPerHour := 0.0
	if cfg.AmortizableHours > 0 {
		depreciationPerHour = cfg.PrinterPrice / cfg.AmortizableHours
	}

	// ─── STEP 3: Costo de repuestos por hora ─────────────────────────────────
	// costo_total_repuestos / horas_vida_util_repuestos
	sparePartsPerHour := 0.0
	if cfg.SparePartsLifeHours > 0 {
		sparePartsPerHour = cfg.SparePartsTotalCost / cfg.SparePartsLifeHours
	}

	// ─── STEP 4: Costo base por hora (sin filamento, sin margen) ─────────────
	baseCostPerHour := electricityPerHour + depreciationPerHour + sparePartsPerHour

	// ─── STEP 5: Costo mínimo por hora con margen ────────────────────────────
	minHourlyRate := baseCostPerHour * (1 + cfg.MarginPercent/100.0)

	// ─── STEP 6: Costo de filamento para este trabajo ─────────────────────────
	filamentCost := (filamentGrams / 1000.0) * cfg.FilamentPricePerKg

	// ─── STEP 7: Costos de hora × tiempo del trabajo ─────────────────────────
	electricityCost := electricityPerHour * hours
	depreciationCost := depreciationPerHour * hours
	sparePartsCost := sparePartsPerHour * hours

	// ─── STEP 8: Costo total de producción ───────────────────────────────────
	totalProductionCost := filamentCost + electricityCost + depreciationCost + sparePartsCost

	// ─── STEP 9: Precio sugerido con margen ──────────────────────────────────
	suggestedPrice := totalProductionCost * (1 + cfg.MarginPercent/100.0)

	explanation := []entity.ExplanationStep{
		{
			Step:    1,
			Label:   "Costo electricidad/hora",
			Formula: fmt.Sprintf("(%.0f W ÷ 1000) × $%.4f/kWh", cfg.PrinterWattage, cfg.ElectricityKWhPrice),
			Result:  round2(electricityPerHour),
		},
		{
			Step:    2,
			Label:   "Depreciación impresora/hora",
			Formula: fmt.Sprintf("$%.2f ÷ %.0f horas de vida útil", cfg.PrinterPrice, cfg.AmortizableHours),
			Result:  round2(depreciationPerHour),
		},
		{
			Step:    3,
			Label:   "Repuestos/hora",
			Formula: fmt.Sprintf("$%.2f en repuestos ÷ %.0f horas de vida útil", cfg.SparePartsTotalCost, cfg.SparePartsLifeHours),
			Result:  round2(sparePartsPerHour),
		},
		{
			Step:    4,
			Label:   "Costo base por hora (sin filamento, sin margen)",
			Formula: "electricidad/h + depreciación/h + repuestos/h",
			Result:  round2(baseCostPerHour),
		},
		{
			Step:    5,
			Label:   fmt.Sprintf("Mínimo a cobrar por hora (con %.0f%% margen)", cfg.MarginPercent),
			Formula: fmt.Sprintf("$%.4f × (1 + %.0f%% / 100)", baseCostPerHour, cfg.MarginPercent),
			Result:  round2(minHourlyRate),
		},
		{
			Step:    6,
			Label:   "Costo filamento para este trabajo",
			Formula: fmt.Sprintf("(%.0f g ÷ 1000) × $%.2f/kg", filamentGrams, cfg.FilamentPricePerKg),
			Result:  round2(filamentCost),
		},
		{
			Step:    7,
			Label:   "Costo total de producción",
			Formula: fmt.Sprintf("filamento + (%.4f/h × %.2fh)", baseCostPerHour, hours),
			Result:  round2(totalProductionCost),
		},
		{
			Step:    8,
			Label:   fmt.Sprintf("Precio mínimo a cobrar por este trabajo (%.0f%% margen)", cfg.MarginPercent),
			Formula: fmt.Sprintf("$%.2f × (1 + %.0f%% / 100)", totalProductionCost, cfg.MarginPercent),
			Result:  round2(suggestedPrice),
		},
	}

	return &entity.CostBreakdown{
		FilamentGrams:    filamentGrams,
		PrintTimeMinutes: printTimeMinutes,

		ElectricityCostPerHour: round2(electricityPerHour),
		DepreciationPerHour:    round2(depreciationPerHour),
		SparePartsPerHour:      round2(sparePartsPerHour),
		BaseCostPerHour:        round2(baseCostPerHour),
		MinHourlyRate:          round2(minHourlyRate),

		FilamentCost:        round2(filamentCost),
		ElectricityCost:     round2(electricityCost),
		DepreciationCost:    round2(depreciationCost),
		SparePartsCost:      round2(sparePartsCost),
		TotalProductionCost: round2(totalProductionCost),

		MarginPercent:  cfg.MarginPercent,
		SuggestedPrice: round2(suggestedPrice),
		Explanation:    explanation,
	}
}

// round2 rounds a float to 2 decimal places.
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// ComputeBreakdown is exported so the sale use case can call the same math.
func ComputeBreakdown(cfg *entity.CostConfig, filamentGrams float64, printTimeMinutes int) *entity.CostBreakdown {
	return compute(cfg, filamentGrams, printTimeMinutes)
}
