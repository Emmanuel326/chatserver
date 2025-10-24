package main

import (
	"context" // <-- FIX: ADDED MISSING CONTEXT IMPORT
	"fmt"
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api"
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/Emmanuel326/chatserver/internal/ports/sqlite"
	"github.com/Emmanuel326/chatserver/internal/ws"
	"github.com/Emmanuel326/chatserver/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
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
	GroupRepository domain.GroupRepository

	// Domain Services (Interfaces) - These are the logic layers
	UserService domain.UserService
	MessageService domain.MessageService
	GroupService domain.GroupService

	// Auth Component
	JWTManager *auth.JWTManager

	// Real-Time Component
	ChatHub *ws.Hub
}


// createDefaultUsers checks if Ava (ID 1) and Mike (ID 2) exist and creates them if not.
func createDefaultUsers(ctx context.Context, userService domain.UserService) {
	// 1. Create Ava (ID 1)
	if _, err := userService.GetUserByUsername(ctx, "ava"); err != nil {
		if _, err := userService.CreateUser(ctx, "ava", "password"); err == nil {
			logger.Log().Info("Created default user: ava")
		} else {
			logger.Log().Error("Failed to create default user: ava", zap.Error(err))
		}
	}
	
	// 2. Create Mike (ID 2)
	if _, err := userService.GetUserByUsername(ctx, "mike"); err != nil {
		if _, err := userService.CreateUser(ctx, "mike", "password"); err == nil {
			logger.Log().Info("Created default user: mike")
		} else {
			logger.Log().Error("Failed to create default user: mike", zap.Error(err))
		}
	}
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
	groupRepo := sqlite.NewGroupRepository(db)

	// --- 4. Initialize Core Components and Domain Services ---
	jwtManager := auth.NewJWTManager(cfg)
	userService := domain.NewUserService(userRepo)
	groupService := domain.NewGroupService(groupRepo, userRepo)
	// FIX: The user creation logic will only work after the UserService interface is updated below.
	createDefaultUsers(context.Background(), userService) 
	
	// HANDLE CIRCULAR DEPENDENCY & HUB INITIALIZATION:
	
	// 1. Initialize the Hub with nil services (it just needs to exist for MessageService).
	chatHub := ws.NewHub(nil, nil)
	go chatHub.Run()

    // 2. Initialize MessageService, passing all four required dependencies, including the chatHub.
	messageService := domain.NewMessageService(messageRepo, userRepo, groupRepo, chatHub)

    // 3. Inject the created services back into the Hub.
	chatHub.MessageService = messageService
	chatHub.GroupService = groupService

	// --- 5. Package Services for Injection ---
	app := &ApplicationServices{
		Config: cfg,
		DB: db,
		UserRepository: userRepo,
		MessageRepository: messageRepo,
		GroupRepository: groupRepo,
		UserService: userService,
		MessageService: messageService,
		GroupService: groupService,
		JWTManager: jwtManager,
		ChatHub: chatHub,
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
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "db_driver": "sqlite3"})
	})

	// Register all API routes from the /api layer
	api.RegisterRoutes(
		router,
		app.UserService,
		app.JWTManager,
		app.ChatHub,
		app.MessageService,
		app.GroupService,
	)

	return router
}
