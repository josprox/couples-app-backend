package main

import (
	"log"

	"github.com/gin-gonic/gin"
	
	"github.com/jlhal/parejas/config"
	"github.com/jlhal/parejas/internal/ai"
	"github.com/jlhal/parejas/internal/auth"
	"github.com/jlhal/parejas/internal/chat"
	"github.com/jlhal/parejas/internal/location"
	"github.com/jlhal/parejas/internal/relationships"
	"github.com/jlhal/parejas/internal/users"
	"github.com/jlhal/parejas/middleware"
	"github.com/jlhal/parejas/pkg/database"
	"github.com/jlhal/parejas/pkg/notifications"
	"github.com/jlhal/parejas/pkg/redis"
	wsPkg "github.com/jlhal/parejas/pkg/websocket"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()

	// Initialize Database and Redis
	database.ConnectDB(cfg)
	database.MigrateDB()
	redis.ConnectRedis(cfg)

	// Initialize WebSocket Hub
	wsPkg.GlobalHub = wsPkg.NewHub()
	go wsPkg.GlobalHub.Run()

	// Initialize Firebase if credentials exist
	if cfg.FirebaseCredentialsPath != "" {
		if err := notifications.InitFirebase(cfg.FirebaseCredentialsPath); err != nil {
			log.Printf("Warning: Firebase initialization failed: %v", err)
		} else {
			log.Println("Firebase Admin SDK initialized successfully")
		}
	}

	// Setup Gin Router
	r := gin.Default()

	// Apply CORS Middleware
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
			"message": "Couples API is running",
		})
	})

	// Initialize domains
	authRepo := auth.NewRepository(database.DB)
	authService := auth.NewService(authRepo, cfg)
	authHandler := auth.NewHandler(authService)

	usersRepo := users.NewRepository(database.DB)
	usersService := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersService)

	relRepo := relationships.NewRepository(database.DB)
	relService := relationships.NewService(relRepo, usersRepo)
	relHandler := relationships.NewHandler(relService)

	chatRepo := chat.NewRepository(database.DB)
	
	locRepo := location.NewRepository(database.DB)
	locService := location.NewService(locRepo)
	locHandler := location.NewHandler(locService, relService)

	chatService := chat.NewService(chatRepo, relService)
	chatHandler := chat.NewHandler(chatService, locService)

	aiService := ai.NewService(cfg, chatService, relService)
	aiHandler := ai.NewHandler(aiService)

	authMiddleware := middleware.AuthMiddleware(cfg)

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		usersGroup := v1.Group("/users")
		usersGroup.Use(authMiddleware)
		{
			usersGroup.GET("/me", usersHandler.GetMe)
			usersGroup.POST("/fcm-token", usersHandler.UpdateFCMToken)
		}

		relGroup := v1.Group("/relationships")
		relGroup.Use(authMiddleware)
		{
			relGroup.GET("/requests", relHandler.GetRequests)
			relGroup.POST("/requests", relHandler.SendRequest)
			relGroup.PUT("/requests/:id/accept", relHandler.AcceptRequest)
			relGroup.GET("/me", relHandler.GetDashboard)
			relGroup.PATCH("/me/wizard", relHandler.UpdateWizard)
		}

		chatGroup := v1.Group("/chat")
		chatGroup.Use(authMiddleware)
		{
			chatGroup.GET("/messages", chatHandler.GetMessages)
			chatGroup.GET("/ws", chatHandler.ServeWS)
		}

		aiGroup := v1.Group("/ai")
		aiGroup.Use(authMiddleware)
		{
			aiGroup.POST("/prompt", aiHandler.PromptAI)
		}

		locGroup := v1.Group("/location")
		locGroup.Use(authMiddleware)
		{
			locGroup.GET("/partner", locHandler.GetPartnerLocation)
		}
	}

	log.Printf("Starting server on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
