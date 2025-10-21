package main

import (
	"fmt"
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api"
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/Emmanuel326/chatserver/internal/ports/sqlite"
	"github.com/Emmanuel326/chatserver/internal/ws" // Keep this import
	"github.com/Emmanuel326/chatserver/pkg/logger" // Teammate's logger package
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap" // Teammate's logging dependency
)

func init() {
	logger.InitGlobalLogger()
	defer logger.Sync()
	logger.Log().Info("Application is Starting...")
}

// ApplicationServices holds all initialized services and components
type ApplicationServices struct {
	Config *config.Config
	DB *sqlx.DB

	// Repositories (Ports) - These are the concrete implementations
	UserRepository domain.UserRepository
	MessageRepository domain.MessageRepository

	// Domain Services (Interfaces) - These are the logic layers
	UserService domain.UserService
	MessageService domain.MessageService

	// Auth Component
	JWTManager *auth.JWTManager

	// Real-Time Component
	ChatHub *ws.Hub
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
	messageRepo := sqlite.NewMessageRepository(db)

	// --- 4. Initialize Core Components and Domain Services ---
	userService := domain.NewUserService(userRepo)
	messageService := domain.NewMessageService(messageRepo)
	jwtManager := auth.NewJWTManager(cfg)

	// --- Initialize WebSocket Hub and start its main Goroutine ---
	// FIX 1: Pass the messageService to the NewHub constructor (YOUR CORE FIX)
	chatHub := ws.NewHub(messageService)
	go chatHub.Run() // CRUCIAL: Starts the concurrent hub manager

	// --- 5. Package Services for Injection ---
	app := &ApplicationServices{
		Config: cfg,
		DB: db,
		UserRepository: userRepo,
		MessageRepository: messageRepo,
		UserService: userService,
		MessageService: messageService,
		JWTManager: jwtManager,
		ChatHub: chatHub, // Injects the Hub
	}

	// --- 6. Setup and Run Gin Router ---
	router := setupRouter(app)

	// Use the new logger setup from the team
	logger.Log().Info(fmt.Sprintf("🚀 Server running on http://localhost:%s", cfg.SERVER_PORT))
	if err := router.Run(":" + cfg.SERVER_PORT); err != nil {
		logger.Log().Fatal("Server failed to start:", zap.Error(err))
	}
}

// setupRouter initializes Gin and registers routes.
func setupRouter(app *ApplicationServices) *gin.Engine {
	// Teammate's change: logger.Log().Info is likely used elsewhere, but for Gin:
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "db_driver": "sqlite3"})
	})

	// Register all API routes from the /api layer
	// Pass the UserService, JWTManager, ChatHub, and MessageService (for routing logic)
	api.RegisterRoutes(router, app.UserService, app.JWTManager, app.ChatHub, app.MessageService)

	return router
}
