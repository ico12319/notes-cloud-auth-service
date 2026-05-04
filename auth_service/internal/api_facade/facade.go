package api_facade

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/auth"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/access_token"
	refresh_tokens "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/refresh_token"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/token_bundle"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/users"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/encoder"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/oidc"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/password"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/probes"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/random"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/time"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/uuid"
	"log"
	"net/http"
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

	oidcProvider, err := oidc.NewProvider(context.Background(), cfg.GoogleOIDC)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	log.Println("Google OIDC provider initialized")

	oidcCookieService, err := oidc.NewCookieService(cfg.GoogleOIDC.CookieSecret)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC cookie service: %v", err)
	}

	structValidator := validator.New()
	uuidService := uuid.NewService()
	timeService := time.NewService()
	passwordService := password.NewService()
	randomService := random.NewService()
	stringEncoder := encoder.NewService()
	transact := database.NewSqlDb(db)

	r := mux.NewRouter()
	r.Use(middleware.JSONContentType)

	healthHandler := probes.NewHealthHandler(db)

	userConverter := users.NewConverter()
	userRepo := users.NewRepository(userConverter)
	userService := users.NewService(userRepo, passwordService, timeService, uuidService)
	userHandler := users.NewHandler(transact, structValidator, userService, userConverter)

	refreshTokenConverter := refresh_tokens.NewConverter()
	refreshTokenRepository := refresh_tokens.NewRepository(refreshTokenConverter)
	refreshTokenService := refresh_tokens.NewService(refreshTokenRepository, randomService, stringEncoder, uuidService, timeService,
		cfg.RefreshToken.Secret)
	accessTokenService := access_token.NewService(timeService, cfg.AccessToken, jwt.SigningMethodHS256)

	tokenBundleService := token_bundle.NewService(accessTokenService, refreshTokenService, userService, timeService)
	authService := auth.NewService(userService, tokenBundleService, passwordService)
	authHandler := auth.NewHandler(authService, transact, refreshTokenService, tokenBundleService)

	oidcHandler := oidc.NewHandler(oidcProvider, oidcCookieService, randomService, stringEncoder)

	// Public routes
	r.HandleFunc("/authService/api/v1/healthz", healthHandler.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/authService/api/v1/readyz", healthHandler.Readyz).Methods(http.MethodGet)
	r.HandleFunc("/authService/api/v1/register", userHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/logout", authHandler.Logout).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/refresh", authHandler.Refresh).Methods(http.MethodPost)
	r.HandleFunc("/authService/api/v1/auth/google/start", oidcHandler.Start).Methods(http.MethodGet)

	// Protected routes
	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware(accessTokenService))
	protected.HandleFunc("/authService/api/v1/me", userHandler.Me).Methods(http.MethodGet)

	log.Println("Server started on http://localhost:8081")

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
