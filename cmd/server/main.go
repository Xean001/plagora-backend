package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/plagora/backend/config"
	deliveryHTTP "github.com/plagora/backend/internal/delivery/http"
	"github.com/plagora/backend/internal/delivery/http/handler"
	"github.com/plagora/backend/internal/infrastructure/postgres"
	ucAuth "github.com/plagora/backend/internal/usecase/auth"
	ucCalculation "github.com/plagora/backend/internal/usecase/calculation"
	ucCalculator "github.com/plagora/backend/internal/usecase/calculator"
	ucClient "github.com/plagora/backend/internal/usecase/client"
	ucCostConfig "github.com/plagora/backend/internal/usecase/costconfig"
	ucInventory "github.com/plagora/backend/internal/usecase/inventory"
	ucSale "github.com/plagora/backend/internal/usecase/sale"
)

func main() {
	// ─── Config ──────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// ─── Database ────────────────────────────────────────────────────────────
	ctx := context.Background()
	db, err := postgres.Connect(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	if err := postgres.Migrate(ctx, db); err != nil {
		log.Fatalf("running migrations: %v", err)
	}
	log.Println("✅ Migrations applied")

	// ─── Repositories ────────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepository(db)
	clientRepo := postgres.NewClientRepository(db)
	saleRepo := postgres.NewSaleRepository(db)
	costConfigRepo := postgres.NewCostConfigRepository(db)
	calcRepo := postgres.NewCalculationRepo(db)
	invRepo := postgres.NewInventoryRepo(db)

	// ─── Use Cases ───────────────────────────────────────────────────────────
	authUC := ucAuth.New(userRepo, cfg.JWT)
	calcUC := ucCalculator.New(costConfigRepo)
	costConfigUC := ucCostConfig.New(costConfigRepo)
	clientUC := ucClient.New(clientRepo)
	saleUC := ucSale.New(saleRepo, costConfigRepo)
	calculationUC := ucCalculation.New(calcRepo)
	inventoryUC := ucInventory.New(invRepo)

	// ─── Seed admin user ─────────────────────────────────────────────────────
	if err := authUC.SeedAdminIfNeeded(ctx, cfg.Admin.Email, cfg.Admin.Password); err != nil {
		log.Fatalf("seeding admin: %v", err)
	}
	log.Println("✅ Admin user ready")

	// ─── HTTP Handlers ───────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)
	saleHandler := handler.NewSaleHandler(saleUC, calcUC)
	clientHandler := handler.NewClientHandler(clientUC)
	costConfigHandler := handler.NewCostConfigHandler(costConfigUC)
	calculationHandler := handler.NewCalculationHandler(calculationUC)
	inventoryHandler := handler.NewInventoryHandler(inventoryUC)

	// ─── Router ──────────────────────────────────────────────────────────────
	gin.SetMode(cfg.Server.GinMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	router := deliveryHTTP.NewRouter(authHandler, saleHandler, clientHandler, costConfigHandler, calculationHandler, inventoryHandler, cfg.JWT.Secret)
	router.Setup(engine)

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "plagora-backend"})
	})

	// ─── Server ──────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Plagora backend running on http://localhost:%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("👋 Server stopped")
}
