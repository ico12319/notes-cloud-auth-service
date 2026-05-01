package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/probes"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/users"
	"log"
	"net/http"
)

type output struct {
	Text string `json:"text"`
}

func main() {
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

	r := mux.NewRouter()
	r.Use(middleware.JSONContentType)

	healthHandler := probes.NewHealthHandler(db)
	healthHandler.RegisterRoutes(r)

	userRepo := users.NewRepository()
	userConverter := users.NewConverter()
	userService := users.NewService(userRepo, userConverter)
	usersHandler := users.NewHandler(userService, database.NewSqlDb(db))
	usersHandler.RegisterRoutes(r)

	r.HandleFunc("/authService/api/v1/hello-world", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(&output{
			Text: "Hello world",
		}); err != nil {
			http.Error(w, "failed to encode JSON response", http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodGet)

	log.Println("Server started on http://localhost:8081")

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
