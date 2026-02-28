package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
)

type calculationRepo struct{ db *pgxpool.Pool }

func NewCalculationRepo(db *pgxpool.Pool) repository.CalculationRepository {
	return &calculationRepo{db: db}
}

func (r *calculationRepo) Create(ctx context.Context, c *entity.PriceCalculation) error {
	c.ID = uuid.New()
	c.CreatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		INSERT INTO price_calculations
			(id, piece_name, print_hours, print_minutes_extra, filament_grams,
			 supplies_cost, multiplier,
			 material_cost, electricity_cost, machine_wear,
			 subtotal_production, suggested_price, total_with_supplies,
			 sale_price, profit, notes, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		c.ID, c.PieceName, c.PrintHours, c.PrintMinutesExtra, c.FilamentGrams,
		c.SuppliesCost, c.Multiplier,
		c.MaterialCost, c.ElectricityCost, c.MachineWear,
		c.SubtotalProduction, c.SuggestedPrice, c.TotalWithSupplies,
		c.SalePrice, c.Profit, c.Notes, c.CreatedAt,
	)
	return err
}

func (r *calculationRepo) FindAll(ctx context.Context) ([]*entity.PriceCalculation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, piece_name, print_hours, print_minutes_extra, filament_grams,
		       supplies_cost, multiplier,
		       material_cost, electricity_cost, machine_wear,
		       subtotal_production, suggested_price, total_with_supplies,
		       sale_price, profit, notes, created_at
		FROM price_calculations
		ORDER BY created_at DESC
		LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("querying calculations: %w", err)
	}
	defer rows.Close()

	var result []*entity.PriceCalculation
	for rows.Next() {
		c := &entity.PriceCalculation{}
		if err := rows.Scan(
			&c.ID, &c.PieceName, &c.PrintHours, &c.PrintMinutesExtra, &c.FilamentGrams,
			&c.SuppliesCost, &c.Multiplier,
			&c.MaterialCost, &c.ElectricityCost, &c.MachineWear,
			&c.SubtotalProduction, &c.SuggestedPrice, &c.TotalWithSupplies,
			&c.SalePrice, &c.Profit, &c.Notes, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *calculationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM price_calculations WHERE id = $1`, id)
	return err
}
