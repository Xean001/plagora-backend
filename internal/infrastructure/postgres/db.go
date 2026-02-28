package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates and validates a new pgxpool connection to PostgreSQL.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	return pool, nil
}

// Migrate runs the SQL schema creation idempotently.
func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS clients (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		phone TEXT NOT NULL DEFAULT '',
		email TEXT NOT NULL DEFAULT '',
		notes TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Configuración de costos:
	-- Los costos por hora se DERIVAN de estos valores al calcular.
	-- depreciation_per_hour = printer_price / amortizable_hours
	-- spare_parts_per_hour  = spare_parts_total_cost / spare_parts_life_hours
	-- electricity_per_hour  = (printer_wattage / 1000) * electricity_kwh_price
	CREATE TABLE IF NOT EXISTS cost_config (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		filament_price_per_kg  NUMERIC(12,4) NOT NULL DEFAULT 0,
		electricity_kwh_price  NUMERIC(12,4) NOT NULL DEFAULT 0,
		printer_wattage        NUMERIC(10,2) NOT NULL DEFAULT 0,
		printer_price          NUMERIC(14,2) NOT NULL DEFAULT 0,
		amortizable_hours      NUMERIC(10,2) NOT NULL DEFAULT 1,
		spare_parts_total_cost NUMERIC(14,2) NOT NULL DEFAULT 0,
		spare_parts_life_hours NUMERIC(10,2) NOT NULL DEFAULT 1,
		margin_percent         NUMERIC(6,2)  NOT NULL DEFAULT 30,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sales (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
		user_id UUID NOT NULL REFERENCES users(id),
		description TEXT NOT NULL,
		material TEXT NOT NULL DEFAULT '',
		color TEXT NOT NULL DEFAULT '',
		filament_grams NUMERIC(8,2) NOT NULL DEFAULT 0,
		print_time_minutes INT NOT NULL,
		filament_cost     NUMERIC(12,4) NOT NULL DEFAULT 0,
		electricity_cost  NUMERIC(12,4) NOT NULL DEFAULT 0,
		depreciation_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
		spare_parts_cost  NUMERIC(12,4) NOT NULL DEFAULT 0,
		total_production_cost NUMERIC(12,2) NOT NULL,
		profit_margin_percent NUMERIC(6,2)  NOT NULL DEFAULT 0,
		suggested_price   NUMERIC(12,2) NOT NULL,
		final_price       NUMERIC(12,2) NOT NULL,
		profit            NUMERIC(12,2) NOT NULL,
		status TEXT NOT NULL DEFAULT 'pendiente',
		payment_method TEXT NOT NULL DEFAULT '',
		paid BOOLEAN NOT NULL DEFAULT false,
		notes TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status);
	CREATE INDEX IF NOT EXISTS idx_sales_created_at ON sales(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_sales_client_id ON sales(client_id);

	-- ─── Idempotent migrations for existing databases ─────────────────────────

	-- cost_config: remove old derived columns, add new source columns
	ALTER TABLE cost_config DROP COLUMN IF EXISTS failure_rate_percent;
	ALTER TABLE cost_config DROP COLUMN IF EXISTS depreciation_per_hour;
	ALTER TABLE cost_config DROP COLUMN IF EXISTS spare_parts_price_per_hour;
	ALTER TABLE cost_config ADD COLUMN IF NOT EXISTS printer_price          NUMERIC(14,2) NOT NULL DEFAULT 0;
	ALTER TABLE cost_config ADD COLUMN IF NOT EXISTS amortizable_hours      NUMERIC(10,2) NOT NULL DEFAULT 1;
	ALTER TABLE cost_config ADD COLUMN IF NOT EXISTS spare_parts_total_cost NUMERIC(14,2) NOT NULL DEFAULT 0;
	ALTER TABLE cost_config ADD COLUMN IF NOT EXISTS spare_parts_life_hours NUMERIC(10,2) NOT NULL DEFAULT 1;
	ALTER TABLE cost_config ADD COLUMN IF NOT EXISTS margin_percent         NUMERIC(6,2)  NOT NULL DEFAULT 30;

	-- sales: remove failure_cost, ensure spare_parts_cost exists
	ALTER TABLE sales DROP COLUMN IF EXISTS failure_cost;
	ALTER TABLE sales ADD COLUMN IF NOT EXISTS spare_parts_cost NUMERIC(12,4) NOT NULL DEFAULT 0;

	-- price_calculations: guarda cálculos de precio para piezas 3D
	-- permite comparar costo_produccion vs precio_vendido para ver ganancia real
	CREATE TABLE IF NOT EXISTS price_calculations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		piece_name          TEXT NOT NULL DEFAULT '',
		print_hours         NUMERIC(6,2)  NOT NULL DEFAULT 0,
		print_minutes_extra INT           NOT NULL DEFAULT 0,
		filament_grams      NUMERIC(8,2)  NOT NULL DEFAULT 0,
		supplies_cost       NUMERIC(12,2) NOT NULL DEFAULT 0,
		multiplier          NUMERIC(6,2)  NOT NULL DEFAULT 1,
		-- resultados calculados
		material_cost       NUMERIC(12,2) NOT NULL DEFAULT 0,
		electricity_cost    NUMERIC(12,2) NOT NULL DEFAULT 0,
		machine_wear        NUMERIC(12,2) NOT NULL DEFAULT 0,
		subtotal_production NUMERIC(12,2) NOT NULL DEFAULT 0,
		suggested_price     NUMERIC(12,2) NOT NULL DEFAULT 0,
		total_with_supplies NUMERIC(12,2) NOT NULL DEFAULT 0,
		-- precio real de venta (para calcular ganancia real)
		sale_price          NUMERIC(12,2) NOT NULL DEFAULT 0,
		profit              NUMERIC(12,2) NOT NULL DEFAULT 0,
		notes               TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_price_calc_created_at ON price_calculations(created_at DESC);

	-- inventory: tabla de piezas con precio calculado y precio de venta real
	-- es el "inventario" de piezas que el usuario ha calculado/producido
	CREATE TABLE IF NOT EXISTS inventory (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		piece_name          TEXT NOT NULL DEFAULT '',
		-- costos de producción (del calculador)
		production_cost     NUMERIC(12,2) NOT NULL DEFAULT 0,  -- subtotal_produccion
		suggested_price     NUMERIC(12,2) NOT NULL DEFAULT 0,
		supplies_cost       NUMERIC(12,2) NOT NULL DEFAULT 0,
		-- precio que cobra el usuario (editable)
		sale_price          NUMERIC(12,2) NOT NULL DEFAULT 0,
		margin_percent      NUMERIC(6,2)  NOT NULL DEFAULT 0,  -- (sale_price/production_cost - 1)*100
		profit              NUMERIC(12,2) NOT NULL DEFAULT 0,  -- sale_price - production_cost
		-- estado
		status              TEXT NOT NULL DEFAULT 'por_vender', -- por_vender|vendido|descartado
		notes               TEXT NOT NULL DEFAULT '',
		sold_at             TIMESTAMP WITH TIME ZONE,
		created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_inventory_status     ON inventory(status);
	CREATE INDEX IF NOT EXISTS idx_inventory_created_at ON inventory(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_inventory_sale_price ON inventory(sale_price DESC);
	`

	_, err := db.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
