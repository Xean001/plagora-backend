package entity

import (
	"time"

	"github.com/google/uuid"
)

// User is the single admin user of the system.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

// Client represents a customer of Plagora.
type Client struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// CostConfig holds all parameters needed to run the cost calculator.
// Per-hour costs are DERIVED from these values at calculation time.
type CostConfig struct {
	ID uuid.UUID `json:"id"`

	// Precio KG: precio por kg de filamento en moneda local
	FilamentPricePerKg float64 `json:"filament_price_per_kg"`

	// Precio kWh: precio del kilowatt-hora en moneda local
	ElectricityKWhPrice float64 `json:"electricity_kwh_price"`

	// Consumo por hora: watts que consume la impresora (se convierte a kWh internamente)
	PrinterWattage float64 `json:"printer_wattage"`

	// Desgaste máquina: precio de compra de la impresora
	PrinterPrice float64 `json:"printer_price"`

	// Desgaste máquina: horas de vida útil estimada para amortizar la impresora
	AmortizableHours float64 `json:"amortizable_hours"`

	// Precio repuestos: costo total estimado de repuestos durante la vida útil
	SparePartsTotalCost float64 `json:"spare_parts_total_cost"`

	// Precio repuestos: horas de vida útil para distribuir el costo de repuestos
	SparePartsLifeHours float64 `json:"spare_parts_life_hours"`

	// Margen: porcentaje de ganancia a aplicar sobre el costo total (ej: 30)
	MarginPercent float64 `json:"margin_percent"`

	UpdatedAt time.Time `json:"updated_at"`
}

// ExplanationStep documents one step of the cost calculation math.
type ExplanationStep struct {
	Step    int     `json:"step"`
	Label   string  `json:"label"`
	Formula string  `json:"formula"`
	Result  float64 `json:"result"`
}

// CostBreakdown is the full result of a cost calculation (not persisted directly).
type CostBreakdown struct {
	// Job inputs
	FilamentGrams    float64 `json:"filament_grams"`
	PrintTimeMinutes int     `json:"print_time_minutes"`

	// Per-hour costs (derived from CostConfig)
	ElectricityCostPerHour float64 `json:"electricity_cost_per_hour"`
	DepreciationPerHour    float64 `json:"depreciation_per_hour"`
	SparePartsPerHour      float64 `json:"spare_parts_per_hour"`
	BaseCostPerHour        float64 `json:"base_cost_per_hour"`
	MinHourlyRate          float64 `json:"min_hourly_rate"` // base_cost_per_hour con margen

	// Job totals
	FilamentCost        float64 `json:"filament_cost"`
	ElectricityCost     float64 `json:"electricity_cost"`
	DepreciationCost    float64 `json:"depreciation_cost"`
	SparePartsCost      float64 `json:"spare_parts_cost"`
	TotalProductionCost float64 `json:"total_production_cost"`

	// Final pricing
	MarginPercent  float64 `json:"margin_percent"`
	SuggestedPrice float64 `json:"suggested_price"`

	// Step-by-step math explanation
	Explanation []ExplanationStep `json:"explanation"`
}

// SaleStatus represents the lifecycle state of a sale.
type SaleStatus string

const (
	StatusPending    SaleStatus = "pendiente"
	StatusInProgress SaleStatus = "en_proceso"
	StatusCompleted  SaleStatus = "completado"
	StatusCancelled  SaleStatus = "cancelado"
)

// Sale represents a 3D print job sold to a client, including full cost breakdown.
type Sale struct {
	ID uuid.UUID `json:"id"`

	// Relationships
	ClientID *uuid.UUID `json:"client_id"`
	Client   *Client    `json:"client,omitempty"`
	UserID   uuid.UUID  `json:"user_id"`

	// Print job description
	Description string `json:"description"`
	Material    string `json:"material"`
	Color       string `json:"color"`

	// Cost calculator inputs
	FilamentGrams    float64 `json:"filament_grams"`
	PrintTimeMinutes int     `json:"print_time_minutes"`

	// Cost calculator outputs (stored for historical accuracy)
	FilamentCost        float64 `json:"filament_cost"`
	ElectricityCost     float64 `json:"electricity_cost"`
	DepreciationCost    float64 `json:"depreciation_cost"`
	SparePartsCost      float64 `json:"spare_parts_cost"`
	TotalProductionCost float64 `json:"total_production_cost"`
	ProfitMarginPercent float64 `json:"profit_margin_percent"`
	SuggestedPrice      float64 `json:"suggested_price"`

	// Final sale values
	FinalPrice    float64    `json:"final_price"`
	Profit        float64    `json:"profit"`
	Status        SaleStatus `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	Paid          bool       `json:"paid"`
	Notes         string     `json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DashboardStats aggregates business performance metrics.
type DashboardStats struct {
	TotalSales          int     `json:"total_sales"`
	TotalRevenue        float64 `json:"total_revenue"`
	TotalProductionCost float64 `json:"total_production_cost"`
	TotalProfit         float64 `json:"total_profit"`
	PendingSales        int     `json:"pending_sales"`
	InProgressSales     int     `json:"in_progress_sales"`
	UnpaidSales         int     `json:"unpaid_sales"`
	CurrentMonthRevenue float64 `json:"current_month_revenue"`
	CurrentMonthProfit  float64 `json:"current_month_profit"`
}

// PriceCalculation represents a saved price calculation for a 3D printed piece.
// Stores both the calculated (suggested) price and the actual sale price,
// so the user can compare production cost vs what they actually charged.
type PriceCalculation struct {
	ID uuid.UUID `json:"id"`

	// Inputs
	PieceName         string  `json:"piece_name"`
	PrintHours        float64 `json:"print_hours"`
	PrintMinutesExtra int     `json:"print_minutes_extra"`
	FilamentGrams     float64 `json:"filament_grams"`
	SuppliesCost      float64 `json:"supplies_cost"`
	Multiplier        float64 `json:"multiplier"`

	// Calculated results
	MaterialCost       float64 `json:"material_cost"`
	ElectricityCost    float64 `json:"electricity_cost"`
	MachineWear        float64 `json:"machine_wear"`
	SubtotalProduction float64 `json:"subtotal_production"`
	SuggestedPrice     float64 `json:"suggested_price"`
	TotalWithSupplies  float64 `json:"total_with_supplies"`

	// Actual sale (filled by user after selling)
	SalePrice float64 `json:"sale_price"`
	Profit    float64 `json:"profit"` // sale_price - subtotal_production

	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// InventoryStatus represents the lifecycle of an inventory piece.
type InventoryStatus string

const (
	InventoryPorVender  InventoryStatus = "por_vender"
	InventoryVendido    InventoryStatus = "vendido"
	InventoryDescartado InventoryStatus = "descartado"
)

// InventoryItem represents a 3D-printed piece stored in inventory.
// Created from the calculator; tracks production cost vs. actual sale price.
type InventoryItem struct {
	ID uuid.UUID `json:"id"`

	PieceName string `json:"piece_name"`

	// From calculator
	ProductionCost float64 `json:"production_cost"` // subtotal sin insumos
	SuggestedPrice float64 `json:"suggested_price"`
	SuppliesCost   float64 `json:"supplies_cost"`

	// Set by user (editable)
	SalePrice     float64 `json:"sale_price"`
	MarginPercent float64 `json:"margin_percent"` // (sale_price/production_cost -1)*100
	Profit        float64 `json:"profit"`         // sale_price - production_cost

	Status InventoryStatus `json:"status"`
	Notes  string          `json:"notes"`

	SoldAt    *time.Time `json:"sold_at"`
	CreatedAt time.Time  `json:"created_at"`
}
