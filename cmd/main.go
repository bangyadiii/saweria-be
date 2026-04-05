package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"saweria-be/internal/alertqueue"
	"saweria-be/internal/auth"
	"saweria-be/internal/donation"
	"saweria-be/internal/overlay"
	"saweria-be/internal/payment"
	"saweria-be/internal/user"
	"saweria-be/internal/wallet"
	wshub "saweria-be/internal/websocket"
	"saweria-be/internal/widgets"
	"saweria-be/pkg/config"
	"saweria-be/pkg/database"
	"saweria-be/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// ── repositories ─────────────────────────────────────────────────────────
	authRepo := auth.NewRepository(db)
	userRepo := user.NewRepository(db)
	overlayRepo := overlay.NewRepository(db)
	donationRepo := donation.NewRepository(db)
	walletRepo  := wallet.NewRepository(db)
	alertqueueRepo := alertqueue.NewRepository(db)
	widgetRepo  := widgets.NewRepository(db)
	// ── WebSocket hub ─────────────────────────────────────────────────────────
	hub := wshub.NewHub()

	// ── alert queue manager ───────────────────────────────────────────────────
	alertManager := alertqueue.NewManager(context.Background(), alertqueueRepo, donationRepo, overlayRepo, hub)
	if err := alertManager.RecoverPending(context.Background()); err != nil {
		log.Printf("alert queue recovery warning: %v", err)
	}

	// ── payment client ────────────────────────────────────────────────────────
	mtClient := payment.NewMidtransClient(cfg.MidtransServerKey, cfg.MidtransClientKey, cfg.MidtransEnvironment)

	// ── services ──────────────────────────────────────────────────────────────
	authSvc    := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiryHours, cfg.JWTRefreshExpiryDays)
	userSvc := user.NewService(userRepo)
	overlaySvc := overlay.NewService(overlayRepo)
	donationSvc := donation.NewService(donationRepo, userRepo, overlayRepo, mtClient)
	paymentSvc := payment.NewService(donationRepo, walletRepo, alertManager, cfg.PlatformFeePercent)
	walletSvc := wallet.NewService(walletRepo)
	widgetHandler := widgets.NewHandler(widgetRepo, overlayRepo, userRepo)
	// ── WS overlay lookup adapter ─────────────────────────────────────────────
	wsLookup := &overlayLookup{repo: overlayRepo}

	// ── handlers ──────────────────────────────────────────────────────────────
	authHandler := auth.NewHandler(authSvc, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.BaseURL)
	userHandler := user.NewHandler(userSvc)
	overlayHandler := overlay.NewHandler(overlaySvc, hub)
	donationHandler := donation.NewHandler(donationSvc, cfg.PlatformFeePercent)
	paymentHandler := payment.NewHandler(paymentSvc, cfg.MidtransServerKey)
	walletHandler := wallet.NewHandler(walletSvc)
	wsHandler := wshub.NewHandler(hub, wsLookup)

	// ── router ────────────────────────────────────────────────────────────────
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	jwtMiddleware := middleware.JWTAuth(cfg.JWTSecret)

	// Auth
	a := r.Group("/auth")
	{
		a.POST("/register", authHandler.Register)
		a.POST("/login", authHandler.Login)
		a.POST("/refresh", authHandler.Refresh)
		a.GET("/google", authHandler.GoogleLogin)
		a.GET("/google/callback", authHandler.GoogleCallback)
	}

	// Users
	u := r.Group("/users")
	{
		u.GET("/me", jwtMiddleware, userHandler.GetMe)
		u.PUT("/me", jwtMiddleware, userHandler.UpdateMe)
		u.GET("/:username", userHandler.GetPublicProfile)
	}

	// Overlay (all protected)
	o := r.Group("/overlay", jwtMiddleware)
	{
		o.GET("/settings", overlayHandler.GetSettings)
		o.PUT("/alert", overlayHandler.UpdateAlertRules)
		o.PUT("/template", overlayHandler.UpdateTemplate)
		o.PUT("/filter", overlayHandler.UpdateFilterKata)
		o.PUT("/sound", overlayHandler.UpdateSound)
		o.POST("/stream-key/reset", overlayHandler.ResetStreamKey)
		o.PUT("/mediashare/rules", overlayHandler.UpdateMediashareRules)
		o.PUT("/mediashare/template", overlayHandler.UpdateMediashareTemplate)
		o.PUT("/qr", overlayHandler.UpdateQRSettings)
		o.PUT("/milestone", overlayHandler.UpdateMilestoneSettings)
		o.POST("/test-alert", overlayHandler.TestAlert)
		o.POST("/test-mediashare", overlayHandler.TestMediashare)
		o.POST("/control", overlayHandler.Control)
	}

	// Donations
	r.POST("/donate/:username", donationHandler.Submit)
	d := r.Group("/donations", jwtMiddleware)
	{
		d.GET("", donationHandler.GetHistory)
		d.GET("/:id", donationHandler.GetDetail)
	}

	// Payment webhook (no JWT — Midtrans signature verified in handler)
	r.POST("/payment/webhook", paymentHandler.Webhook)

	// Wallet
	w := r.Group("/wallet", jwtMiddleware)
	{
		w.GET("/balance", walletHandler.GetBalance)
		w.POST("/cashout", walletHandler.RequestCashout)
		w.GET("/cashout/history", walletHandler.GetCashoutHistory)
	}

	// Public widget data endpoints (authenticated by streamKey)
	wg := r.Group("/widgets")
	{
		wg.GET("/info", widgetHandler.Info)
		wg.GET("/leaderboard", widgetHandler.Leaderboard)
	}

	// Static uploads (sounds, etc.)
	r.Static("/uploads", "./uploads")

	// WebSocket
	r.GET("/ws", wsHandler.Connect)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// overlayLookup adapts overlay.Repository to websocket.OverlayLookup
type overlayLookup struct {
	repo overlay.Repository
}

func (o *overlayLookup) StreamKeyExists(ctx context.Context, hash string) (bool, error) {
	settings, err := o.repo.FindByStreamKeyHash(ctx, hash)
	if err != nil {
		return false, nil // treat as not found
	}
	return settings != nil, nil
}
