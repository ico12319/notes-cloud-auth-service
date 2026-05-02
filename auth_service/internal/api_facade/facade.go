package api_facade

import (
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/users"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/password"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/probes"
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

	r := mux.NewRouter()
	r.Use(middleware.JSONContentType)

	healthHandler := probes.NewHealthHandler(db)
	healthHandler.RegisterRoutes(r)

	userConverter := users.NewConverter()
	userRepo := users.NewRepository(userConverter)
	userService := users.NewService(userRepo, passwordService, timeService, uuidService)
	userHandler := users.NewHandler(database.NewSqlDb(db), structValidator, userService)
	userHandler.RegisterRoutes(r)

	log.Println("Server started on http://localhost:8081")

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
