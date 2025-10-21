package main

import (
	"fmt"
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api"
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/Emmanuel326/chatserver/internal/ports/sqlite"
	"github.com/Emmanuel326/chatserver/internal/ws" 
	"github.com/Emmanuel326/chatserver/pkg/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func init(){
  logger.InitGlobalLogger()
  defer logger.Sync()
  logger.Log().Info("Application is Starting...")
}
// ApplicationServices holds all initialized services and components
// This central struct facilitates Dependency Injection across the application.
type ApplicationServices struct {
	Config *config.Config
	DB *sqlx.DB
	
	// Domain Services (Interfaces)
	UserRepository domain.UserRepository
	UserService    domain.UserService
	
	// Auth Component
	JWTManager     *auth.JWTManager

	// Real-Time Component
	ChatHub        *ws.Hub // New: The core WebSocket hub
	
	// TODO: MessageRepository domain.MessageRepository
}

func main() {
	// --- 1. Load Configuration ---
	cfg := config.Load()

	// --- 2. Initialize Database Connection and Migration ---
	db := sqlite.InitDB(cfg)
	defer db.Close()
	sqlite.Migrate(db)

	// --- 3. Initialize Repositories (Ports) ---
	userRepo := sqlite.NewUserRepository(db)
	
	// --- 4. Initialize Core Components and Domain Services ---
	userService := domain.NewUserService(userRepo)
	jwtManager := auth.NewJWTManager(cfg)

	// --- Initialize WebSocket Hub and start its main Goroutine ---
	chatHub := ws.NewHub()
	go chatHub.Run() // CRUCIAL: Starts the concurrent hub manager

	// --- 5. Package Services for Injection ---
	app := &ApplicationServices{
		Config: cfg,
		DB: db,
		UserRepository: userRepo,
		UserService: userService,
		JWTManager: jwtManager,
		ChatHub: chatHub, // Inject the Hub
	}

	// --- 6. Setup and Run Gin Router ---
	router := setupRouter(app)

  logger.Log().Info(fmt.Sprintf("🚀 Server running on http://localhost:%s", cfg.SERVER_PORT))
	if err := router.Run(":" + cfg.SERVER_PORT); err != nil {
		logger.Log().Fatal("Server failed to start: %v", zap.Error(err))
	}
}

// setupRouter initializes Gin and registers routes.
func setupRouter(app *ApplicationServices) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "db_driver": "sqlite3"})
	})

	// Register all API routes from the /api layer
	// Pass the UserService, JWTManager, AND the ChatHub
	api.RegisterRoutes(router, app.UserService, app.JWTManager, app.ChatHub)

	return router
}

