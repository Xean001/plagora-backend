package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	ExistsAny(ctx context.Context) (bool, error)
}

// ClientRepository defines persistence operations for clients.
type ClientRepository interface {
	FindAll(ctx context.Context) ([]*entity.Client, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Client, error)
	Create(ctx context.Context, client *entity.Client) error
	Update(ctx context.Context, client *entity.Client) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SaleFilter holds optional filters for listing sales.
type SaleFilter struct {
	Status   *entity.SaleStatus
	Paid     *bool
	ClientID *uuid.UUID
	FromDate *string // YYYY-MM-DD
	ToDate   *string // YYYY-MM-DD
}

// SaleRepository defines persistence operations for sales.
type SaleRepository interface {
	FindAll(ctx context.Context, filter SaleFilter) ([]*entity.Sale, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Sale, error)
	Create(ctx context.Context, sale *entity.Sale) error
	Update(ctx context.Context, sale *entity.Sale) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetDashboardStats(ctx context.Context) (*entity.DashboardStats, error)
}

// CostConfigRepository defines persistence operations for cost configuration.
type CostConfigRepository interface {
	Get(ctx context.Context) (*entity.CostConfig, error)
	Upsert(ctx context.Context, cfg *entity.CostConfig) error
}

// CalculationRepository defines persistence operations for price calculations.
type CalculationRepository interface {
	Create(ctx context.Context, calc *entity.PriceCalculation) error
	FindAll(ctx context.Context) ([]*entity.PriceCalculation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// InventoryFilter holds optional filters for listing inventory items.
type InventoryFilter struct {
	Search *string
	Status *entity.InventoryStatus
	SortBy string // "price_desc" | "price_asc" | "created_desc"
}

// InventoryRepository defines persistence operations for inventory items.
type InventoryRepository interface {
	Create(ctx context.Context, item *entity.InventoryItem) error
	FindAll(ctx context.Context, filter InventoryFilter) ([]*entity.InventoryItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.InventoryItem, error)
	Update(ctx context.Context, item *entity.InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetRevenue(ctx context.Context) (float64, error) // sum of sale_price where status=vendido
}
