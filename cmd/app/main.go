package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"subscription-service/internal/config"
	"subscription-service/internal/handlers"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"
)

// @title Subscription Service API
// @version 1.0
// @description API сервис для управления подписками пользователей
// @host localhost:8080
// @BasePath /
func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Настраиваем логгер
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting subscription service...")

	// Подключаемся к БД
	db, err := sqlx.Connect("postgres", cfg.DBConnectionString())
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	logger.Info("Connected to database")

	// Инициализируем репозиторий, сервис и хендлер
	subscriptionRepo := repository.NewSubscriptionRepository(db, logger)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, logger)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, logger)

	// Настраиваем роутер
	router := gin.Default()

	// Swagger документация
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API роуты
	api := router.Group("/api")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", subscriptionHandler.CreateSubscription)
			subscriptions.GET("", subscriptionHandler.ListSubscriptions)
			subscriptions.GET("/summary", subscriptionHandler.GetTotalPrice)
			subscriptions.GET("/:id", subscriptionHandler.GetSubscription)
			subscriptions.PUT("/:id", subscriptionHandler.UpdateSubscription)
			subscriptions.DELETE("/:id", subscriptionHandler.DeleteSubscription)
		}
	}

	// Запускаем сервер
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.WithField("port", cfg.Port).Info("Server starting...")
	if err := router.Run(addr); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
