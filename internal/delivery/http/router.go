package http

import (
	"github.com/gin-gonic/gin"
	"github.com/plagora/backend/internal/delivery/http/handler"
	"github.com/plagora/backend/internal/delivery/http/middleware"
)

type Router struct {
	authHandler       *handler.AuthHandler
	saleHandler       *handler.SaleHandler
	clientHandler     *handler.ClientHandler
	costConfigHandler *handler.CostConfigHandler
	calcHandler       *handler.CalculationHandler
	inventoryHandler  *handler.InventoryHandler
	jwtSecret         string
}

func NewRouter(
	authH *handler.AuthHandler,
	saleH *handler.SaleHandler,
	clientH *handler.ClientHandler,
	costH *handler.CostConfigHandler,
	calcH *handler.CalculationHandler,
	invH *handler.InventoryHandler,
	jwtSecret string,
) *Router {
	return &Router{
		authHandler:       authH,
		saleHandler:       saleH,
		clientHandler:     clientH,
		costConfigHandler: costH,
		calcHandler:       calcH,
		inventoryHandler:  invH,
		jwtSecret:         jwtSecret,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.CORS())

	api := engine.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	{
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.Refresh)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWT(r.jwtSecret))
	{
		// Dashboard
		protected.GET("/dashboard/stats", r.saleHandler.DashboardStats)

		// Cost calculator (preview)
		protected.POST("/ventas/calcular", r.saleHandler.Calculate)

		// Sales CRUD
		ventas := protected.Group("/ventas")
		{
			ventas.GET("", r.saleHandler.GetAll)
			ventas.POST("", r.saleHandler.Create)
			ventas.GET("/:id", r.saleHandler.GetByID)
			ventas.PUT("/:id", r.saleHandler.Update)
			ventas.DELETE("/:id", r.saleHandler.Delete)
		}

		// Clients CRUD
		clientes := protected.Group("/clientes")
		{
			clientes.GET("", r.clientHandler.GetAll)
			clientes.POST("", r.clientHandler.Create)
			clientes.GET("/:id", r.clientHandler.GetByID)
			clientes.PUT("/:id", r.clientHandler.Update)
			clientes.DELETE("/:id", r.clientHandler.Delete)
		}

		// Cost configuration
		config := protected.Group("/config")
		{
			config.GET("/costos", r.costConfigHandler.Get)
			config.PUT("/costos", r.costConfigHandler.Update)
		}

		// Price calculator (saved)
		calc := protected.Group("/calculadora")
		{
			calc.GET("", r.calcHandler.GetAll)
			calc.POST("", r.calcHandler.Save)
			calc.DELETE("/:id", r.calcHandler.Delete)
		}

		// Inventory
		inv := protected.Group("/inventario")
		{
			inv.GET("", r.inventoryHandler.GetAll)
			inv.POST("", r.inventoryHandler.Add)
			inv.PUT("/:id", r.inventoryHandler.Update)
			inv.DELETE("/:id", r.inventoryHandler.Delete)
			inv.GET("/revenue", r.inventoryHandler.Revenue)
		}
	}
}
