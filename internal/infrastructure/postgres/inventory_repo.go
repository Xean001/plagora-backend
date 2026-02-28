package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
)

type inventoryRepo struct{ db *pgxpool.Pool }

func NewInventoryRepo(db *pgxpool.Pool) repository.InventoryRepository {
	return &inventoryRepo{db: db}
}

const inventorySelect = `
	SELECT id, piece_name,
	       production_cost, suggested_price, supplies_cost,
	       sale_price, margin_percent, profit,
	       status, notes, sold_at, created_at
	FROM inventory`

func scanInventory(row interface{ Scan(...any) error }) (*entity.InventoryItem, error) {
	it := &entity.InventoryItem{}
	return it, row.Scan(
		&it.ID, &it.PieceName,
		&it.ProductionCost, &it.SuggestedPrice, &it.SuppliesCost,
		&it.SalePrice, &it.MarginPercent, &it.Profit,
		&it.Status, &it.Notes, &it.SoldAt, &it.CreatedAt,
	)
}

func (r *inventoryRepo) Create(ctx context.Context, it *entity.InventoryItem) error {
	it.ID = uuid.New()
	it.CreatedAt = time.Now()
	if it.Status == "" {
		it.Status = entity.InventoryPorVender
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO inventory
			(id, piece_name, production_cost, suggested_price, supplies_cost,
			 sale_price, margin_percent, profit, status, notes, sold_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		it.ID, it.PieceName, it.ProductionCost, it.SuggestedPrice, it.SuppliesCost,
		it.SalePrice, it.MarginPercent, it.Profit, it.Status, it.Notes, it.SoldAt, it.CreatedAt,
	)
	return err
}

func (r *inventoryRepo) FindAll(ctx context.Context, f repository.InventoryFilter) ([]*entity.InventoryItem, error) {
	q := inventorySelect
	var args []interface{}
	var conds []string
	n := 1

	if f.Search != nil && *f.Search != "" {
		conds = append(conds, fmt.Sprintf("piece_name ILIKE $%d", n))
		args = append(args, "%"+*f.Search+"%")
		n++
	}
	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", n))
		args = append(args, *f.Status)
		n++
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	switch f.SortBy {
	case "price_asc":
		q += " ORDER BY sale_price ASC"
	case "price_desc":
		q += " ORDER BY sale_price DESC"
	default:
		q += " ORDER BY created_at DESC"
	}
	q += " LIMIT 200"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("querying inventory: %w", err)
	}
	defer rows.Close()

	var result []*entity.InventoryItem
	for rows.Next() {
		it, err := scanInventory(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, it)
	}
	return result, rows.Err()
}

func (r *inventoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.InventoryItem, error) {
	row := r.db.QueryRow(ctx, inventorySelect+" WHERE id = $1", id)
	return scanInventory(row)
}

func (r *inventoryRepo) Update(ctx context.Context, it *entity.InventoryItem) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inventory SET
			piece_name = $1, production_cost = $2, suggested_price = $3, supplies_cost = $4,
			sale_price = $5, margin_percent = $6, profit = $7,
			status = $8, notes = $9, sold_at = $10
		WHERE id = $11`,
		it.PieceName, it.ProductionCost, it.SuggestedPrice, it.SuppliesCost,
		it.SalePrice, it.MarginPercent, it.Profit,
		it.Status, it.Notes, it.SoldAt, it.ID,
	)
	return err
}

func (r *inventoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM inventory WHERE id = $1`, id)
	return err
}

func (r *inventoryRepo) GetRevenue(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(sale_price),0) FROM inventory WHERE status = 'vendido'`,
	).Scan(&total)
	return total, err
}
