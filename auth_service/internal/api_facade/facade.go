package api_facade

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/auth"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/email_verification_tokens"
	refresh_tokens "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/refresh_token"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/token_bundle"
	user_identities "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/user_identities"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/users"
	email_verification "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/email-verification"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/encoder"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/oauth"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/oidc"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/password"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/probes"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/random"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/time"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	time2 "time"
)

type apiFacade struct{}

func NewApiFacade() *apiFacade {
	return &apiFacade{}
}

func (*apiFacade) Start() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	ctx := context.Background()
	googleOIDCProvider, err := oidc.NewOIDCProvider(ctx, cfg.OIDCProviders["google"])
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	log.Println("Google OIDC provider initialized")

	gitLabOIDCProvider, err := oidc.NewOIDCProvider(ctx, cfg.OIDCProviders["gitlab"])
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	log.Println("GitLab OIDC provider initialized")

	oidcCookieService, err := oauth.NewCookieService(cfg.Cookie.Secret, false)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC cookie service: %v", err)
	}

	structValidator := validator.New()
	uuidService := uuid.NewService()
	timeService := time.NewService()
	passwordService := password.NewService()
	stringEncoder := encoder.NewService()
	randomService := random.NewService(stringEncoder)
	transact := database.NewSqlDb(db)

	var wg sync.WaitGroup
	r := mux.NewRouter()
	r.Use(middleware.JSONContentType, middleware.HTTPRequestTracker(&wg))

	healthHandler := probes.NewHealthHandler(db)

	userConverter := users.NewConverter()
	userRepo := users.NewRepository(userConverter)

	identityConverter := user_identities.NewConverter()
	identityRepo := user_identities.NewRepository(identityConverter)
	identityService := user_identities.NewService(identityRepo, uuidService, timeService)

	emailVerificationTokenServiceConverter := email_verification_tokens.NewConverter()
	emailVerificationTokenServiceRepository := email_verification_tokens.NewRepository(emailVerificationTokenServiceConverter)
	emailVerificationTokenServiceService := email_verification_tokens.NewService(emailVerificationTokenServiceRepository, uuidService, timeService)

	verificationEmailContentGenerator := email_verification.NewEmailContentGenerator()
	emailSenderService := email_verification.NewService(cfg.Resend.APIKey, cfg.Resend.FromEmail, verificationEmailContentGenerator)

	userService := users.NewService(userRepo, passwordService, timeService, uuidService, identityService, emailVerificationTokenServiceService)
	userHandler := users.NewHandler(transact, structValidator, userService, userConverter, passwordService, emailSenderService, randomService, emailVerificationTokenServiceService)

	refreshTokenConverter := refresh_tokens.NewConverter()
	refreshTokenRepository := refresh_tokens.NewRepository(refreshTokenConverter)
	refreshTokenService := refresh_tokens.NewService(refreshTokenRepository, randomService, stringEncoder, uuidService, timeService,
		cfg.RefreshToken.Secret)

	accessTokenConfig, err := accesstoken.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config for acess token: %s", err.Error())
	}

	accessTokenService := accesstoken.NewService(timeService, *accessTokenConfig, jwt.SigningMethodHS256)

	tokenBundleService := token_bundle.NewService(accessTokenService, refreshTokenService, userService, timeService)
	authService := auth.NewService(userService, tokenBundleService, passwordService)
	authHandler := auth.NewHandler(authService, transact, refreshTokenService, tokenBundleService)

	googleUserAuthInfoExtractor := oidc.NewOIDCUserAuthInfoExtractor(googleOIDCProvider)
	gitLabUserAuthInfoExtractor := oidc.NewOIDCUserAuthInfoExtractor(gitLabOIDCProvider)

	oauthSessionBuilder := oauth.NewOauthSessionBuilder(randomService)
	googleOIDCHandler := oidc.NewHandler(googleOIDCProvider, oidcCookieService, oauthSessionBuilder, googleUserAuthInfoExtractor, userService, tokenBundleService, transact, cfg.FrontendURL)
	gitLabOIDCHandler := oidc.NewHandler(gitLabOIDCProvider, oidcCookieService, oauthSessionBuilder, gitLabUserAuthInfoExtractor, userService, tokenBundleService, transact, cfg.FrontendURL)

	// Public routes
	r.HandleFunc("/authService/api/v1/healthz", healthHandler.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/authService/api/v1/readyz", healthHandler.Readyz).Methods(http.MethodGet)

	r.HandleFunc("/authService/api/v1/auth/google/start", googleOIDCHandler.Start).Methods(http.MethodGet)
	r.HandleFunc("/authService/api/v1/auth/google/callback", googleOIDCHandler.Callback).Methods(http.MethodGet)

	r.HandleFunc("/authService/api/v1/auth/gitlab/start", gitLabOIDCHandler.Start).Methods(http.MethodGet)
	r.HandleFunc("/authService/api/v1/auth/gitlab/callback", gitLabOIDCHandler.Callback).Methods(http.MethodGet)

	r.HandleFunc("/authService/api/v1/register", userHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/logout", authHandler.Logout).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/refresh", authHandler.Refresh).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/email/verify", userHandler.Verify).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/email/resend-verification", userHandler.Resend).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/users/{user_id}", userHandler.User).Methods(http.MethodGet)

	log.Println("Server started on http://localhost:8081")

	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	const shutdownTimeout = 25 * time2.Second
	serverErrorChan := make(chan error, 1)
	
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Println(fmt.Sprintf("error different from ServerClosed occurred %s",
				err.Error()))

			serverErrorChan <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		log.Println("shutting down... (Ctrl+C again to force)")
		stop()

	case err := <-serverErrorChan:
		log.Fatalf("server failed: %v", err)
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutDownErrChan := make(chan error, 1)
	// Shutdown is a blocking operation so it is spawned in a goroutine
	go func() {
		shutDownErrChan <- srv.Shutdown(shutCtx)
	}()

	// waitDone is a cancellation channel and we are spawning one more goroutine because Wait is a blocking operation
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		if err := <-shutDownErrChan; err != nil {
			log.Printf("server shutdown error: %v", err)
		}
		log.Println("all requests done, clean exit")
	case <-shutCtx.Done():
		log.Printf("shutdown timeout after %s, forcing exit\n", shutdownTimeout)
	}

}
