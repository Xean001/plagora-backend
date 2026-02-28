package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
)

type saleRepo struct {
	db *pgxpool.Pool
}

func NewSaleRepository(db *pgxpool.Pool) repository.SaleRepository {
	return &saleRepo{db: db}
}

func (r *saleRepo) FindAll(ctx context.Context, filter repository.SaleFilter) ([]*entity.Sale, error) {
	query := `SELECT
		s.id, s.client_id, s.user_id, s.description, s.material, s.color,
		s.filament_grams, s.print_time_minutes,
		s.filament_cost, s.electricity_cost, s.depreciation_cost, s.spare_parts_cost,
		s.total_production_cost, s.profit_margin_percent, s.suggested_price,
		s.final_price, s.profit, s.status, s.payment_method, s.paid, s.notes,
		s.created_at, s.updated_at,
		c.name, c.phone
	FROM sales s
	LEFT JOIN clients c ON s.client_id = c.id`

	var args []interface{}
	var conditions []string
	argCount := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("s.status = $%d", argCount))
		args = append(args, *filter.Status)
		argCount++
	}
	if filter.Paid != nil {
		conditions = append(conditions, fmt.Sprintf("s.paid = $%d", argCount))
		args = append(args, *filter.Paid)
		argCount++
	}
	if filter.ClientID != nil {
		conditions = append(conditions, fmt.Sprintf("s.client_id = $%d", argCount))
		args = append(args, *filter.ClientID)
		argCount++
	}
	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("s.created_at >= $%d", argCount))
		args = append(args, *filter.FromDate)
		argCount++
	}
	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("s.created_at <= $%d", argCount))
		args = append(args, *filter.ToDate)
		argCount++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY s.created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []*entity.Sale
	for rows.Next() {
		s := &entity.Sale{}
		var clientName, clientPhone *string
		if err := rows.Scan(
			&s.ID, &s.ClientID, &s.UserID, &s.Description, &s.Material, &s.Color,
			&s.FilamentGrams, &s.PrintTimeMinutes,
			&s.FilamentCost, &s.ElectricityCost, &s.DepreciationCost, &s.SparePartsCost,
			&s.TotalProductionCost, &s.ProfitMarginPercent, &s.SuggestedPrice,
			&s.FinalPrice, &s.Profit, &s.Status, &s.PaymentMethod, &s.Paid, &s.Notes,
			&s.CreatedAt, &s.UpdatedAt,
			&clientName, &clientPhone,
		); err != nil {
			return nil, err
		}
		if clientName != nil {
			s.Client = &entity.Client{Name: *clientName, Phone: *clientPhone}
		}
		sales = append(sales, s)
	}
	return sales, nil
}

func (r *saleRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Sale, error) {
	s := &entity.Sale{}
	var clientName, clientPhone *string
	err := r.db.QueryRow(ctx, `SELECT
		s.id, s.client_id, s.user_id, s.description, s.material, s.color,
		s.filament_grams, s.print_time_minutes,
		s.filament_cost, s.electricity_cost, s.depreciation_cost, s.spare_parts_cost,
		s.total_production_cost, s.profit_margin_percent, s.suggested_price,
		s.final_price, s.profit, s.status, s.payment_method, s.paid, s.notes,
		s.created_at, s.updated_at,
		c.name, c.phone
	FROM sales s
	LEFT JOIN clients c ON s.client_id = c.id
	WHERE s.id = $1`, id,
	).Scan(
		&s.ID, &s.ClientID, &s.UserID, &s.Description, &s.Material, &s.Color,
		&s.FilamentGrams, &s.PrintTimeMinutes,
		&s.FilamentCost, &s.ElectricityCost, &s.DepreciationCost, &s.SparePartsCost,
		&s.TotalProductionCost, &s.ProfitMarginPercent, &s.SuggestedPrice,
		&s.FinalPrice, &s.Profit, &s.Status, &s.PaymentMethod, &s.Paid, &s.Notes,
		&s.CreatedAt, &s.UpdatedAt,
		&clientName, &clientPhone,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("sale not found")
	}
	if clientName != nil {
		s.Client = &entity.Client{Name: *clientName, Phone: *clientPhone}
	}
	return s, err
}

func (r *saleRepo) Create(ctx context.Context, s *entity.Sale) error {
	_, err := r.db.Exec(ctx, `INSERT INTO sales (
		id, client_id, user_id, description, material, color,
		filament_grams, print_time_minutes,
		filament_cost, electricity_cost, depreciation_cost, spare_parts_cost,
		total_production_cost, profit_margin_percent, suggested_price,
		final_price, profit, status, payment_method, paid, notes,
		created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23
	)`,
		s.ID, s.ClientID, s.UserID, s.Description, s.Material, s.Color,
		s.FilamentGrams, s.PrintTimeMinutes,
		s.FilamentCost, s.ElectricityCost, s.DepreciationCost, s.SparePartsCost,
		s.TotalProductionCost, s.ProfitMarginPercent, s.SuggestedPrice,
		s.FinalPrice, s.Profit, s.Status, s.PaymentMethod, s.Paid, s.Notes,
		s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *saleRepo) Update(ctx context.Context, s *entity.Sale) error {
	_, err := r.db.Exec(ctx, `UPDATE sales SET
		client_id=$1, description=$2, material=$3, color=$4,
		final_price=$5, profit=$6, status=$7,
		payment_method=$8, paid=$9, notes=$10, updated_at=$11
	WHERE id=$12`,
		s.ClientID, s.Description, s.Material, s.Color,
		s.FinalPrice, s.Profit, s.Status,
		s.PaymentMethod, s.Paid, s.Notes, s.UpdatedAt,
		s.ID,
	)
	return err
}

func (r *saleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sales WHERE id=$1`, id)
	return err
}

func (r *saleRepo) GetDashboardStats(ctx context.Context) (*entity.DashboardStats, error) {
	stats := &entity.DashboardStats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status != 'cancelado'),
			COALESCE(SUM(final_price) FILTER (WHERE status != 'cancelado'), 0),
			COALESCE(SUM(total_production_cost) FILTER (WHERE status != 'cancelado'), 0),
			COALESCE(SUM(profit) FILTER (WHERE status != 'cancelado'), 0),
			COUNT(*) FILTER (WHERE status = 'pendiente'),
			COUNT(*) FILTER (WHERE status = 'en_proceso'),
			COUNT(*) FILTER (WHERE paid = false AND status != 'cancelado'),
			COALESCE(SUM(final_price) FILTER (WHERE date_trunc('month', created_at) = date_trunc('month', NOW()) AND status != 'cancelado'), 0),
			COALESCE(SUM(profit) FILTER (WHERE date_trunc('month', created_at) = date_trunc('month', NOW()) AND status != 'cancelado'), 0)
		FROM sales
	`).Scan(
		&stats.TotalSales, &stats.TotalRevenue, &stats.TotalProductionCost, &stats.TotalProfit,
		&stats.PendingSales, &stats.InProgressSales, &stats.UnpaidSales,
		&stats.CurrentMonthRevenue, &stats.CurrentMonthProfit,
	)
	return stats, err
}
