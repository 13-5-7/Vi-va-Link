package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bus-logistics/backend/config"
	"github.com/bus-logistics/backend/db"
	"github.com/bus-logistics/backend/handler"
	custommiddleware "github.com/bus-logistics/backend/middleware"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func main() {
	log.Printf("mainロジックに入りました。")
	// 1. 設定の読み込み
    cfg := config.Load()

	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	inviteRepo := repository.NewInviteRepository(pool)
	authService := service.NewAuthService(userRepo, inviteRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	resetRepo := repository.NewPasswordResetRepository(pool)
	resetService := service.NewPasswordResetService(userRepo, resetRepo)
	resetHandler := handler.NewPasswordResetHandler(resetService)

	adminService := service.NewAdminService(inviteRepo)
	adminHandler := handler.NewAdminHandler(adminService)

	companyRepo := repository.NewCompanyRepository(pool)
	companyHandler := handler.NewCompanyHandler(companyRepo, userRepo, pool)

	scheduleRepo := repository.NewScheduleRepository(pool)
	bookingRepo := repository.NewBookingRepository(pool)
	scheduleService := service.NewScheduleService(scheduleRepo, bookingRepo)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)

	bookingService := service.NewBookingService(pool, bookingRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)

	trackingRepo := repository.NewTrackingRepository(pool)
	trackingService := service.NewTrackingService(pool, bookingRepo, scheduleRepo, trackingRepo)
	trackingHandler := handler.NewTrackingHandler(trackingService)

	routingHandler := handler.NewRoutingHandler(cfg.OSRMBaseURL)

    e := echo.New()
	e.HTTPErrorHandler = custommiddleware.CustomErrorHandler
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(custommiddleware.CORS())
	e.Use(custommiddleware.SecurityHeaders())

	// 認証エンドポイント用レートリミット: 10リクエスト/分（バースト5）
	authRateLimiter := custommiddleware.NewRateLimiterStore(rate.Every(6*time.Second), 5) // 10req/min

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

    api := e.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register, custommiddleware.RateLimit(authRateLimiter))
	auth.POST("/login", authHandler.Login, custommiddleware.RateLimit(authRateLimiter))

    // パスワードリセット（レートリミット適用）
	auth.POST("/password-reset/request", resetHandler.RequestReset, custommiddleware.RateLimit(authRateLimiter))
	auth.POST("/password-reset/confirm", resetHandler.ResetPassword, custommiddleware.RateLimit(authRateLimiter))

	jwtMW := custommiddleware.JWTAuth(cfg.JWTSecret)
	operatorMW := custommiddleware.RequireRole("bus_operator")
	shipperMW := custommiddleware.RequireRole("shipper")

	schedules := api.Group("/schedules", jwtMW)
	schedules.GET("", scheduleHandler.List, operatorMW)
	schedules.POST("", scheduleHandler.Create, operatorMW)
	schedules.GET("/search", scheduleHandler.Search, shipperMW)
	schedules.GET("/:id", scheduleHandler.GetByID)
	schedules.PATCH("/:id/status", scheduleHandler.UpdateStatus, operatorMW)
	schedules.DELETE("/:id", scheduleHandler.Delete, operatorMW)
	schedules.POST("/:id/cancel", scheduleHandler.Cancel, operatorMW)

	bookings := api.Group("/bookings", jwtMW)
	bookings.GET("", bookingHandler.List, shipperMW)
	bookings.POST("", bookingHandler.Create, shipperMW)
	bookings.GET("/:id", bookingHandler.GetByID)
	bookings.DELETE("/:id", bookingHandler.Cancel, shipperMW)
	bookings.PATCH("/:id/status", trackingHandler.UpdateStatus, operatorMW)

	// tracking (認証不要)
	api.GET("/tracking/:tracking_number", trackingHandler.GetByTrackingNumber)

	api.GET("/routing", routingHandler.GetRoute)

	// companies
	api.GET("/companies", companyHandler.List)
	companies := api.Group("/companies", jwtMW)
	companies.GET("/me", companyHandler.GetMyCompany, operatorMW)
	companies.PATCH("/me/storage", companyHandler.UpdateStorage, operatorMW)

	// 管理者API（ADMIN_KEY ヘッダーによる簡易認証）
	adminKey := cfg.AdminKey
	admin := api.Group("/admin", custommiddleware.AdminKeyAuth(adminKey))
	admin.POST("/invite-codes", adminHandler.IssueInviteCode)
	admin.GET("/invite-codes", adminHandler.ListInviteCodes)

	log.Printf("Starting server on :%s", cfg.Port)
    e.Logger.Fatal(e.Start(":" + cfg.Port))
}
