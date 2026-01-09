package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"headless_form/internal/adapter/api"
	"headless_form/internal/adapter/email"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/adapter/storage/sqlite"
	"headless_form/internal/adapter/webhook"
	"headless_form/internal/core/domain"
	"headless_form/internal/core/service"
	"headless_form/web"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (optional, won't fail if missing)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 1. Environment Config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production-please!"
		log.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET in production!")
	}

	isDev := os.Getenv("ENV") != "production"
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	// 2. Storage
	dataDir := os.Getenv("DATA_DIR")
	dbPath := "data.db"
	if dataDir != "" {
		dbPath = filepath.Join(dataDir, "data.db")
	}

	store, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	// 3. Email Configuration
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if smtpPort == 0 {
		smtpPort = 587
	}

	emailConfig := email.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     smtpPort,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		FromName: os.Getenv("SMTP_FROM_NAME"),
		UseTLS:   os.Getenv("SMTP_TLS") == "true",
		Enabled:  os.Getenv("SMTP_HOST") != "",
	}

	if emailConfig.FromName == "" {
		emailConfig.FromName = "HeadlessForms"
	}

	emailService := email.NewService(emailConfig)

	if emailConfig.Enabled {
		log.Printf("📧 Email notifications enabled (SMTP: %s:%d)", emailConfig.Host, emailConfig.Port)
	} else {
		log.Println("📧 Email notifications disabled (no SMTP_HOST configured)")
	}

	// 4. Services
	formService := service.NewFormService(store)
	submService := service.NewSubmissionService(store)
	statsService := service.NewStatsService(store)
	authService := service.NewAuthService(store, service.AuthConfig{
		JWTSecret:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	})

	// 5. Webhook service
	webhookService := webhook.NewService()
	log.Println("🔗 Webhook service initialized")

	// 6. Notification callback (email + webhook)
	submService.SetNotificationCallback(func(form *domain.Form, submission *domain.Submission, data map[string]interface{}) {
		// Send email notification
		if len(form.NotifyEmails) > 0 {
			emailData := email.SubmissionData{
				FormName:     form.Name,
				FormID:       form.PublicID,
				SubmissionID: submission.ID,
				SubmittedAt:  submission.CreatedAt,
				Fields:       data,
				DashboardURL: fmt.Sprintf("%s/forms/%s", baseURL, form.PublicID),
			}

			if err := emailService.SendSubmissionNotification(form.NotifyEmails, emailData); err != nil {
				log.Printf("Failed to send email notification: %v", err)
			}
		}

		// Trigger webhook
		webhookService.TriggerSubmission(form, submission, data)
	})

	// 6. Auth Handler
	authHandler := api.NewAuthHandler(authService, emailService, baseURL)

	// 7. API Router
	router := api.NewRouter(formService, submService, statsService)
	mux := http.NewServeMux()

	// Auth routes (public with rate limiting)
	mux.Handle("POST /api/v1/auth/register",
		middleware.AuthLimiter.Middleware()(http.HandlerFunc(authHandler.HandleRegister)))
	mux.Handle("POST /api/v1/auth/login",
		middleware.AuthLimiter.Middleware()(http.HandlerFunc(authHandler.HandleLogin)))
	mux.HandleFunc("GET /api/v1/auth/setup", authHandler.HandleSetupRequired)

	// Password reset routes (public with rate limiting)
	mux.Handle("POST /api/v1/auth/forgot-password",
		middleware.AuthLimiter.Middleware()(http.HandlerFunc(authHandler.HandleForgotPassword)))
	mux.Handle("POST /api/v1/auth/reset-password",
		middleware.AuthLimiter.Middleware()(http.HandlerFunc(authHandler.HandleResetPassword)))

	// Protected auth routes
	authMiddleware := middleware.AuthMiddleware(authService)
	mux.Handle("GET /api/v1/auth/me",
		authMiddleware(http.HandlerFunc(authHandler.HandleMe)))

	// User management routes (admin only, protected by JWT)
	mux.Handle("GET /api/v1/users",
		authMiddleware(http.HandlerFunc(authHandler.HandleListUsers)))
	mux.Handle("POST /api/v1/users",
		authMiddleware(http.HandlerFunc(authHandler.HandleCreateUser)))
	mux.Handle("DELETE /api/v1/users/{user_id}",
		authMiddleware(http.HandlerFunc(authHandler.HandleDeleteUser)))

	// Profile management routes (self-service, protected by JWT)
	mux.Handle("PUT /api/v1/auth/profile",
		authMiddleware(http.HandlerFunc(authHandler.HandleUpdateProfile)))
	mux.Handle("PUT /api/v1/auth/password",
		authMiddleware(http.HandlerFunc(authHandler.HandleUpdatePassword)))

	// User update route (admin only)
	mux.Handle("PUT /api/v1/users/{user_id}",
		authMiddleware(http.HandlerFunc(authHandler.HandleUpdateUser)))

	// Settings routes (super_admin only, protected by JWT)
	settingsHandler := api.NewSettingsHandler(store)
	mux.Handle("GET /api/v1/settings",
		authMiddleware(http.HandlerFunc(settingsHandler.HandleGetSettings)))
	mux.Handle("PUT /api/v1/settings",
		authMiddleware(http.HandlerFunc(settingsHandler.HandleUpdateSettings)))
	mux.Handle("POST /api/v1/settings/test-smtp",
		authMiddleware(http.HandlerFunc(settingsHandler.HandleTestSMTP)))

	// Register public routes (with optional auth for private form submissions)
	optionalAuth := middleware.OptionalAuthMiddleware(authService)
	router.RegisterPublicRoutes(mux, optionalAuth)

	// Register protected routes (JWT required for dashboard management)
	router.RegisterProtectedRoutes(mux, authMiddleware)

	log.Println("🔒 Dashboard routes protected with JWT authentication")

	// 8. Static Files (SvelteKit build)
	webBuild, err := fs.Sub(web.StaticFiles, "build")
	if err != nil {
		log.Fatalf("Failed to load embedded web assets: %v", err)
	}
	fileServer := http.FileServer(http.FS(webBuild))

	// Serve static files for all non-API routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if _, err := fs.Stat(webBuild, path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	// 9. Apply middleware chain
	corsConfig := middleware.SecurityConfig{
		IsDevelopment: isDev,
	}

	handler := middleware.SecurityHeaders()(
		middleware.CORSMiddleware(corsConfig)(
			middleware.LoggingMiddleware(mux)))

	// 10. Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("╔════════════════════════════════════════════╗")
	log.Printf("║   Headless Form Manager v1.0.0             ║")
	log.Printf("║   Server running on port %s               ║", port)
	log.Printf("║   http://localhost:%s                     ║", port)
	log.Printf("╚════════════════════════════════════════════╝")

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
