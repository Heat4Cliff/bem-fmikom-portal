package router

import (
	"database/sql"
	"log/slog"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/handler"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
	"github.com/gin-gonic/gin"
)

func New(logger *slog.Logger) *gin.Engine { return newEngine(logger, nil) }

func NewWithDependencies(logger *slog.Logger, cfg config.Config, db *sql.DB) (*gin.Engine, error) {
	authService, err := service.NewAuthService(db, cfg, logger)
	if err != nil {
		return nil, err
	}
	engine := newEngine(logger, cfg.FrontendOrigins)
	auth := handler.NewAuth(authService, cfg, logger)
	content := handler.NewContent(service.NewContentService(db))
	content.Register(engine, authService)
	organization := handler.NewOrganization(service.NewOrganizationService(db))
	organization.Register(engine, authService)
	aspirations := handler.NewAspirations(service.NewAspirationService(db))
	aspirations.Register(engine, authService, cfg.FrontendOrigins, middleware.NewAspirationLimiter(cfg.AspirationLimitCount, cfg.AspirationLimitWindow))
	registrationStorage, err := storage.NewLocalPrivateDocuments(cfg.RegistrationStoragePath)
	if err != nil {
		return nil, err
	}
	registrations := handler.NewRegistrations(service.NewRegistrationService(db, registrationStorage, cfg.RegistrationImageMaxUploadBytes), registrationStorage, cfg.RegistrationImageMaxUploadBytes)
	registrations.Register(engine, authService, cfg.FrontendOrigins, middleware.NewRegistrationLimiter(cfg.RegistrationLimitCount, cfg.RegistrationLimitWindow))
	documentStorage, err := storage.NewLocalPrivateDocuments(cfg.DocumentStoragePath)
	if err != nil {
		return nil, err
	}
	documents := service.NewDocumentService(db, documentStorage, cfg.DocumentMaxUploadBytes)
	handler.NewDocuments(documents, documentStorage, cfg.DocumentMaxUploadBytes).Register(engine, authService)
	mediaStorage, err := storage.NewLocalPublicMediaStorage(cfg.MediaStoragePath)
	if err != nil {
		return nil, err
	}
	media := service.NewMediaService(db, mediaStorage, cfg.MediaMaxUploadBytes)
	handler.NewMedia(media, mediaStorage, cfg.MediaMaxUploadBytes).Register(engine, authService)
	systemManagement := service.NewSystemManagementService(db, cfg.BcryptCost)
	handler.NewSystemManagementWithDashboard(systemManagement, service.NewDashboardService(db)).Register(engine, authService)
	limiter := middleware.NewLoginLimiter(cfg.LoginLimitCount, cfg.LoginLimitWindow)
	refreshLimiter := middleware.NewAuthRefreshLimiter(cfg.RefreshLimitCount, cfg.RefreshLimitWindow)
	group := engine.Group("/api/v1/auth")
	group.Use(middleware.RequireTrustedOrigin(cfg.FrontendOrigins))
	group.POST("/login", limiter.Middleware(), auth.Login)
	group.POST("/refresh", refreshLimiter.Middleware(), auth.Refresh)
	group.POST("/logout", auth.Logout)
	group.GET("/me", middleware.RequireAuth(authService), auth.Me)
	return engine, nil
}

func newEngine(logger *slog.Logger, origins []string) *gin.Engine {
	if logger == nil {
		logger = slog.Default()
	}
	engine := gin.New()
	_ = engine.SetTrustedProxies(nil)
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.RequestLogger(logger))
	engine.Use(middleware.CORS(origins))
	engine.Use(middleware.SecurityHeaders())
	engine.Use(middleware.RequestGuards())
	engine.GET("/healthz", handler.Healthz)
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, handler.FailureCode("NOT_FOUND", "Resource not found", nil))
	})
	return engine
}
