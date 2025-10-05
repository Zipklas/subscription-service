package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Zipklas/subscription-service/internal/config"
	"github.com/Zipklas/subscription-service/internal/handler"
	"github.com/Zipklas/subscription-service/internal/logger"
	"github.com/Zipklas/subscription-service/internal/repository"
	"github.com/Zipklas/subscription-service/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	// Swagger
	_ "github.com/Zipklas/subscription-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Subscription Service API
// @version 1.0
// @description REST API для управления подписками пользователей
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	log := logger.New(cfg.LogLevel)
	log.Info(context.Background(), "Starting subscription service",
		"port", cfg.AppPort,
		"log_level", cfg.LogLevel.String(),
	)

	// Подключаемся к базе данных
	db, err := initDatabase(cfg, log)
	if err != nil {
		log.Error(context.Background(), "Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info(context.Background(), "Connected to database successfully")

	// Инициализируем слои приложения
	subscriptionRepo := repository.NewSubscriptionRepository(db, log)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, log)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, log)

	// Настраиваем роутер
	router := setupRouter(subscriptionHandler, log)

	// Запускаем сервер
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info(context.Background(), "Server starting",
		"address", "http://localhost:"+cfg.AppPort,
	)
	log.Info(context.Background(), "Swagger documentation available",
		"url", "http://localhost:"+cfg.AppPort+"/swagger/index.html",
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error(context.Background(), "Failed to start server", "error", err)
		os.Exit(1)
	}
}

// initDatabase инициализирует подключение к базе данных
func initDatabase(cfg *config.Config, log *logger.Logger) (*sql.DB, error) {
	connStr := cfg.GetDBConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Debug(context.Background(), "Database connection pool configured")
	return db, nil
}

// setupRouter настраивает маршруты приложения
// @Summary Health check
// @Description Проверка работоспособности сервиса
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "status"
// @Router /health [get]
func setupRouter(subscriptionHandler *handler.SubscriptionHandler, log *logger.Logger) *gin.Engine {
	// Устанавливаем режим Gin
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Middleware
	router.Use(ginLoggerMiddleware(log)) // Кастомный логгер
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", healthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api/v1")
	{
		// Subscription CRUDL routes
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", subscriptionHandler.CreateSubscription)
			subscriptions.GET("", subscriptionHandler.ListSubscriptions)
			subscriptions.GET("/:id", subscriptionHandler.GetSubscription)
			subscriptions.PUT("/:id", subscriptionHandler.UpdateSubscription)
			subscriptions.DELETE("/:id", subscriptionHandler.DeleteSubscription)

			// Summary route
			subscriptions.GET("/summary", subscriptionHandler.CalculateTotalCost)
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		log.Warn(context.Background(), "Endpoint not found",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)
		c.JSON(404, gin.H{
			"error":         "endpoint not found",
			"message":       "use /api/v1/subscriptions for subscriptions API",
			"documentation": "/swagger/index.html",
		})
	})

	return router
}

// healthCheck возвращает статус сервиса
// @Summary Health check
// @Description Проверка работоспособности сервиса
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "status"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	// Московский часовой пояс
	location, _ := time.LoadLocation("Europe/Moscow")
	currentTime := time.Now().In(location)

	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": currentTime.Format("2006-01-02 15:04:05"),
		"timezone":  "Europe/Moscow",
		"service":   "subscription-service",
		"version":   "1.0.0",
	})
}

func ginLoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Логируем после обработки
		duration := time.Since(start)

		log.Info(c.Request.Context(), "HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
