package delivery

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/about"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/datasync"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/health"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/middleware"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/swagger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// BuildInfoOperator interface for accessing build information.
type BuildInfoOperator about.BuildInfoOperator

// RouteRegistry manages registration of all HTTP routes and their handlers.
// Coordinates authentication, business logic services, and route grouping.
type RouteRegistry struct {
	// authService handles user authentication operations.
	authService auth.Service
	// authJWTService provides JWT authentication middleware.
	authJWTService middleware.AuthWithJWTService
	// buildInfoOperator provides application build information.
	buildInfoOperator BuildInfoOperator
	// bankcardService handles bank card operations.
	bankcardService bankcard.Service
	// credentialService handles credential operations.
	credentialService credential.Service
	// noteService handles note operations.
	noteService note.Service
	// datasyncService handles data synchronization operations.
	datasyncService datasync.Service
	// filedataService handles file data operations.
	filedataService filedata.Service
}

// NewRouteRegistry creates a new RouteRegistry with all required service dependencies.
func NewRouteRegistry(
	authService auth.Service,
	authJWTService middleware.AuthWithJWTService,
	buildInfoOperator BuildInfoOperator,
	bankcardService bankcard.Service,
	credentialService credential.Service,
	noteService note.Service,
	datasyncService datasync.Service,
	filedataService filedata.Service,
) *RouteRegistry {
	return &RouteRegistry{
		authService:       authService,
		authJWTService:    authJWTService,
		buildInfoOperator: buildInfoOperator,
		bankcardService:   bankcardService,
		credentialService: credentialService,
		noteService:       noteService,
		datasyncService:   datasyncService,
		filedataService:   filedataService,
	}
}

// RegisterRoutes configures all application routes on the provided Gin engine.
// Sets up base routes (health, auth, swagger, about) and protected item routes.
func (rr *RouteRegistry) RegisterRoutes(router *gin.Engine) {
	baseGroup := rr.makeBaseGroup(router)
	rr.registerBaseRoutes(baseGroup)
	rr.registerItemsRoutes(baseGroup)
}

// makeBaseGroup creates the base API route group with "/api" prefix.
func (rr *RouteRegistry) makeBaseGroup(router *gin.Engine) *gin.RouterGroup {
	return router.Group("/api")
}

// registerBaseRoutes registers public routes that don't require authentication.
func (rr *RouteRegistry) registerBaseRoutes(group *gin.RouterGroup) {
	health.RegisterRoutes(group, health.NewHandler())
	auth.RegisterRoutes(group, auth.NewHandler(rr.authService))
	swagger.RegisterRoutes(group, ginSwagger.WrapHandler(swaggerFiles.Handler))
	about.RegisterRoutes(group, about.NewHandler(rr.buildInfoOperator))
}

// registerItemsRoutes registers protected routes that require JWT authentication.
// All item endpoints are under "/api/items" with JWT middleware protection.
func (rr *RouteRegistry) registerItemsRoutes(group *gin.RouterGroup) {
	itemsGroup := group.Group("items", middleware.AuthWithJWT(rr.authJWTService))
	bankcard.RegisterRoutes(itemsGroup, bankcard.NewHandler(rr.bankcardService))
	credential.RegisterRoutes(itemsGroup, credential.NewHandler(rr.credentialService))
	note.RegisterRoutes(itemsGroup, note.NewHandler(rr.noteService))
	datasync.RegisterRoutes(itemsGroup, datasync.NewHandler(rr.datasyncService))
	filedata.RegisterRoutes(itemsGroup, filedata.NewHandler(rr.filedataService))
}
