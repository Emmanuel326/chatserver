package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api" 
	"github.com/Emmanuel326/chatserver/internal/auth" // <--- NEW IMPORT
	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/Emmanuel326/chatserver/internal/ports/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// ApplicationServices holds all initialized services and components
type ApplicationServices struct {
	Config *config.Config
	DB *sqlx.DB
	
	// Domain Services (Interfaces)
	UserRepository domain.UserRepository
	UserService    domain.UserService    
	
	// Auth Component
	JWTManager     *auth.JWTManager // <--- NEW: JWT Manager

	// TODO: MessageRepository domain.MessageRepository
	// TODO: ChatHub *ws.Hub 
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
	jwtManager := auth.NewJWTManager(cfg) // <--- Initialize JWT Manager

	// --- 5. Package Services for Injection ---
	app := &ApplicationServices{
		Config: cfg,
		DB: db,
		UserRepository: userRepo,
		UserService: userService, 
		JWTManager: jwtManager, // <--- Inject into the central struct
	}

	// --- 6. Setup and Run Gin Router ---
	router := setupRouter(app)

	// Start the server
	fmt.Printf("🚀 Server running on http://localhost:%s\n", cfg.SERVER_PORT)
	if err := router.Run(":" + cfg.SERVER_PORT); err != nil {
		log.Fatalf("Server failed to start: %v", err)
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
	// Pass the UserService and the JWTManager
	api.RegisterRoutes(router, app.UserService, app.JWTManager) // <--- Pass JWTManager

	return router
}
