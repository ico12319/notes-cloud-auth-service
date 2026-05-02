package api_facade

import (
	"github.com/go-playground/validator/v10"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/auth"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/access_token"
	refresh_tokens "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/refresh_token"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/users"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/encoder"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/jwt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
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

	structValidator := validator.New()
	uuidService := uuid.NewService()
	timeService := time.NewService()
	passwordService := password.NewService()
	randomService := random.NewService()
	stringEncoder := encoder.NewService()
	jwtGenerator := jwt.NewGenerator(jwt2.SigningMethodHS512)
	transact := database.NewSqlDb(db)

	r := mux.NewRouter()
	r.Use(middleware.JSONContentType)

	healthHandler := probes.NewHealthHandler(db)
	healthHandler.RegisterRoutes(r)

	userConverter := users.NewConverter()
	userRepo := users.NewRepository(userConverter)
	userService := users.NewService(userRepo, passwordService, timeService, uuidService)
	userHandler := users.NewHandler(transact, structValidator, userService)
	userHandler.RegisterRoutes(r)

	refreshTokenConverter := refresh_tokens.NewConverter()
	refreshTokenRepository := refresh_tokens.NewRepository(refreshTokenConverter)
	refreshTokenService := refresh_tokens.NewService(refreshTokenRepository, randomService, stringEncoder, uuidService, timeService,
		cfg.RefreshToken.Secret)
	accessTokenService := access_token.NewService(jwtGenerator, timeService, cfg.AccessToken)

	authService := auth.NewService(userService, accessTokenService, refreshTokenService, passwordService)
	authHandler := auth.NewHandler(authService, transact, refreshTokenService)

	authHandler.RegisterRoutes(r)

	log.Println("Server started on http://localhost:8081")

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
