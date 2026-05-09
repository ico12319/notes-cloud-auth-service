package api_facade

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/clients"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/email_verification_tokens"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/oauth"
	proxy2 "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/proxy"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/auth"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/access_token"
	refresh_tokens "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/refresh_token"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/token_bundle"
	user_identities "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/user_identities"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/users"
	email_verification "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/email-verification"
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

	r := mux.NewRouter()
	r.Use(middleware.JSONContentType)

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
	accessTokenService := access_token.NewService(timeService, cfg.AccessToken, jwt.SigningMethodHS256)

	tokenBundleService := token_bundle.NewService(accessTokenService, refreshTokenService, userService, timeService)
	authService := auth.NewService(userService, tokenBundleService, passwordService)
	authHandler := auth.NewHandler(authService, transact, refreshTokenService, tokenBundleService)

	googleUserAuthInfoExtractor := oidc.NewOIDCUserAuthInfoExtractor(googleOIDCProvider)
	gitLabUserAuthInfoExtractor := oidc.NewOIDCUserAuthInfoExtractor(gitLabOIDCProvider)

	googleOIDCHandler := oidc.NewHandler(googleOIDCProvider, oidcCookieService, randomService, googleUserAuthInfoExtractor, userService, tokenBundleService, transact)
	gitLabOIDCHandler := oidc.NewHandler(gitLabOIDCProvider, oidcCookieService, randomService, gitLabUserAuthInfoExtractor, userService, tokenBundleService, transact)

	todoClient := clients.NewTodoClient(cfg.Services.TodoServiceURL)
	notesClient := clients.NewNotesClient(cfg.Services.NotesServiceURL)
	sharingClient := clients.NewSharingClient(cfg.Services.SharingServiceURL)
	reminderClient := clients.NewReminderNotificationClient(cfg.Services.ReminderServiceURL)
	proxy := proxy2.NewHandler(todoClient, notesClient, sharingClient, reminderClient)
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

	// Protected routes
	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware(accessTokenService))
	protected.HandleFunc("/authService/api/v1/me", userHandler.Me).Methods(http.MethodGet)

	// Todo tasks
	protected.HandleFunc("/authService/api/v1/todos", proxy.GetTodos).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/todos", proxy.CreateTodo).Methods(http.MethodPost)
	protected.HandleFunc("/authService/api/v1/todos/{todo_id}", proxy.Todo).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/todos/{todo_id}", proxy.UpdateTodo).Methods(http.MethodPut)
	protected.HandleFunc("/authService/api/v1/todos/{todo_id}", proxy.DeleteTodo).Methods(http.MethodDelete)

	// Todo lists
	protected.HandleFunc("/authService/api/v1/todo-lists", proxy.GetTodoLists).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/todo-lists", proxy.CreateTodoList).Methods(http.MethodPost)
	protected.HandleFunc("/authService/api/v1/todo-lists/{list_id}", proxy.GetTodoList).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/todo-lists/{list_id}", proxy.UpdateTodoList).Methods(http.MethodPut)
	protected.HandleFunc("/authService/api/v1/todo-lists/{list_id}", proxy.DeleteTodoList).Methods(http.MethodDelete)

	// Notes
	protected.HandleFunc("/authService/api/v1/notes", proxy.GetNotes).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/notes", proxy.CreateNote).Methods(http.MethodPost)
	protected.HandleFunc("/authService/api/v1/notes/{note_id}", proxy.GetNote).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/notes/{note_id}", proxy.UpdateNote).Methods(http.MethodPut)
	protected.HandleFunc("/authService/api/v1/notes/{note_id}", proxy.DeleteNote).Methods(http.MethodDelete)

	// Sharing (protected - create share link)
	protected.HandleFunc("/authService/api/v1/notes/{note_id}/share-links", proxy.CreateShareLink).Methods(http.MethodPost)

	// Sharing (public - open share link)
	r.HandleFunc("/authService/api/v1/share-links/{token}", proxy.OpenShareLink).Methods(http.MethodGet)

	// Reminders
	protected.HandleFunc("/authService/api/v1/reminders", proxy.GetReminders).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/reminders", proxy.CreateReminder).Methods(http.MethodPost)
	protected.HandleFunc("/authService/api/v1/reminders", proxy.UpdateReminder).Methods(http.MethodPut)
	protected.HandleFunc("/authService/api/v1/reminders/{reminder_id}", proxy.GetReminder).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/reminders/{reminder_id}", proxy.DeleteReminder).Methods(http.MethodDelete)

	// Notifications
	protected.HandleFunc("/authService/api/v1/notifications", proxy.GetNotifications).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/notifications", proxy.DeleteAllNotifications).Methods(http.MethodDelete)
	protected.HandleFunc("/authService/api/v1/notifications/unread-count", proxy.GetUnreadNotificationCount).Methods(http.MethodGet)
	protected.HandleFunc("/authService/api/v1/notifications/read-all", proxy.MarkAllNotificationsAsRead).Methods(http.MethodPost)
	protected.HandleFunc("/authService/api/v1/notifications/{notification_id}/read", proxy.MarkNotificationAsRead).Methods(http.MethodPost)

	log.Println("Server started on http://localhost:8081")

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
