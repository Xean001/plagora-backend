package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
)

type costConfigRepo struct {
	db *pgxpool.Pool
}

func NewCostConfigRepository(db *pgxpool.Pool) repository.CostConfigRepository {
	return &costConfigRepo{db: db}
}

func (r *costConfigRepo) Get(ctx context.Context) (*entity.CostConfig, error) {
	cfg := &entity.CostConfig{}
	err := r.db.QueryRow(ctx, `
		SELECT id, filament_price_per_kg, electricity_kwh_price, printer_wattage,
		       printer_price, amortizable_hours,
		       spare_parts_total_cost, spare_parts_life_hours,
		       margin_percent, updated_at
		FROM cost_config LIMIT 1`,
	).Scan(
		&cfg.ID, &cfg.FilamentPricePerKg, &cfg.ElectricityKWhPrice, &cfg.PrinterWattage,
		&cfg.PrinterPrice, &cfg.AmortizableHours,
		&cfg.SparePartsTotalCost, &cfg.SparePartsLifeHours,
		&cfg.MarginPercent, &cfg.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("configuración de costos no inicializada — llama PUT /api/config/costos primero")
	}
	return cfg, err
}

func (r *costConfigRepo) Upsert(ctx context.Context, cfg *entity.CostConfig) error {
	_, err := r.db.Exec(ctx, `DELETE FROM cost_config`)
	if err != nil {
		return err
	}
	cfg.ID = uuid.New()
	cfg.UpdatedAt = time.Now()
	_, err = r.db.Exec(ctx, `
		INSERT INTO cost_config (
			id, filament_price_per_kg, electricity_kwh_price, printer_wattage,
			printer_price, amortizable_hours,
			spare_parts_total_cost, spare_parts_life_hours,
			margin_percent, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		cfg.ID, cfg.FilamentPricePerKg, cfg.ElectricityKWhPrice, cfg.PrinterWattage,
		cfg.PrinterPrice, cfg.AmortizableHours,
		cfg.SparePartsTotalCost, cfg.SparePartsLifeHours,
		cfg.MarginPercent, cfg.UpdatedAt,
	)
	return err
}
