package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
)

// --- Auth ---

type LoginInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthUseCase interface {
	Login(ctx context.Context, input LoginInput) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	SeedAdminIfNeeded(ctx context.Context, email, password string) error
}

// --- Cost Config ---

// UpdateCostConfigInput holds all editable parameters for the cost calculator.
type UpdateCostConfigInput struct {
	// Precio KG: precio por kg de filamento
	FilamentPricePerKg float64 `json:"filament_price_per_kg" binding:"gte=0"`

	// Precio kWh: precio del kilowatt-hora
	ElectricityKWhPrice float64 `json:"electricity_kwh_price" binding:"gte=0"`

	// Consumo por hora: watts de la impresora
	PrinterWattage float64 `json:"printer_wattage" binding:"gte=0"`

	// Desgaste máquina: precio de compra de la impresora
	PrinterPrice float64 `json:"printer_price" binding:"gte=0"`

	// Desgaste máquina: horas de vida útil para amortizar
	AmortizableHours float64 `json:"amortizable_hours" binding:"gte=0"`

	// Precio repuestos: costo total de repuestos en la vida útil
	SparePartsTotalCost float64 `json:"spare_parts_total_cost" binding:"gte=0"`

	// Precio repuestos: horas de vida útil para distribuir el costo
	SparePartsLifeHours float64 `json:"spare_parts_life_hours" binding:"gte=0"`

	// Margen: porcentaje de ganancia (ej: 30 para 30%)
	MarginPercent float64 `json:"margin_percent" binding:"gte=0,lte=1000"`
}

type CostConfigUseCase interface {
	Get(ctx context.Context) (*entity.CostConfig, error)
	Update(ctx context.Context, input UpdateCostConfigInput) (*entity.CostConfig, error)
}

// --- Calculator ---

type CalculateInput struct {
	FilamentGrams    float64 `json:"filament_grams" binding:"gte=0"`
	PrintTimeMinutes int     `json:"print_time_minutes" binding:"gt=0"`
}

type CalculatorUseCase interface {
	Calculate(ctx context.Context, input CalculateInput) (*entity.CostBreakdown, error)
}

// --- Saved Calculations ---

type SaveCalculationInput struct {
	PieceName         string  `json:"piece_name"`
	PrintHours        float64 `json:"print_hours" binding:"gte=0"`
	PrintMinutesExtra int     `json:"print_minutes_extra" binding:"gte=0"`
	FilamentGrams     float64 `json:"filament_grams" binding:"gte=0"`
	SuppliesCost      float64 `json:"supplies_cost" binding:"gte=0"`
	Multiplier        float64 `json:"multiplier" binding:"gte=0"`
	// Results (sent from frontend after calculating)
	MaterialCost       float64 `json:"material_cost"`
	ElectricityCost    float64 `json:"electricity_cost"`
	MachineWear        float64 `json:"machine_wear"`
	SubtotalProduction float64 `json:"subtotal_production"`
	SuggestedPrice     float64 `json:"suggested_price"`
	TotalWithSupplies  float64 `json:"total_with_supplies"`
	// Optional: filled when user knows the sale price
	SalePrice float64 `json:"sale_price"`
	Notes     string  `json:"notes"`
}

type CalculationUseCase interface {
	Save(ctx context.Context, input SaveCalculationInput) (*entity.PriceCalculation, error)
	GetAll(ctx context.Context) ([]*entity.PriceCalculation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// --- Inventory ---

type AddToInventoryInput struct {
	PieceName      string  `json:"piece_name"`
	ProductionCost float64 `json:"production_cost"`
	SuggestedPrice float64 `json:"suggested_price"`
	SuppliesCost   float64 `json:"supplies_cost"`
	SalePrice      float64 `json:"sale_price" binding:"gte=0"`
	Notes          string  `json:"notes"`
}

type UpdateInventoryInput struct {
	SalePrice float64                `json:"sale_price"`
	Status    entity.InventoryStatus `json:"status"`
	Notes     string                 `json:"notes"`
}

type InventoryUseCase interface {
	Add(ctx context.Context, input AddToInventoryInput) (*entity.InventoryItem, error)
	GetAll(ctx context.Context, filter repository.InventoryFilter) ([]*entity.InventoryItem, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInventoryInput) (*entity.InventoryItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetRevenue(ctx context.Context) (float64, error)
}

// --- Clients ---

type CreateClientInput struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Notes string `json:"notes"`
}

type UpdateClientInput struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Notes string `json:"notes"`
}

type ClientUseCase interface {
	GetAll(ctx context.Context) ([]*entity.Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Client, error)
	Create(ctx context.Context, input CreateClientInput) (*entity.Client, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateClientInput) (*entity.Client, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// --- Sales ---

type CreateSaleInput struct {
	ClientID         *uuid.UUID        `json:"client_id"`
	Description      string            `json:"description" binding:"required"`
	Material         string            `json:"material"`
	Color            string            `json:"color"`
	FilamentGrams    float64           `json:"filament_grams" binding:"gte=0"`
	PrintTimeMinutes int               `json:"print_time_minutes" binding:"required,gt=0"`
	FinalPrice       float64           `json:"final_price" binding:"required,gt=0"`
	Status           entity.SaleStatus `json:"status"`
	PaymentMethod    string            `json:"payment_method"`
	Paid             bool              `json:"paid"`
	Notes            string            `json:"notes"`
}

type UpdateSaleInput struct {
	ClientID      *uuid.UUID        `json:"client_id"`
	Description   string            `json:"description"`
	Material      string            `json:"material"`
	Color         string            `json:"color"`
	FinalPrice    *float64          `json:"final_price"`
	Status        entity.SaleStatus `json:"status"`
	PaymentMethod string            `json:"payment_method"`
	Paid          *bool             `json:"paid"`
	Notes         string            `json:"notes"`
}

type SaleUseCase interface {
	GetAll(ctx context.Context, filter repository.SaleFilter) ([]*entity.Sale, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Sale, error)
	Create(ctx context.Context, userID uuid.UUID, input CreateSaleInput) (*entity.Sale, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateSaleInput) (*entity.Sale, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetDashboardStats(ctx context.Context) (*entity.DashboardStats, error)
}
