package api

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.infratographer.com/permissions-api/internal/query"
	"go.infratographer.com/x/echojwtx"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/urnx"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("go.infratographer.com/permissions-api/internal/api")

// Router provides a router for the API
type Router struct {
	authMW echo.MiddlewareFunc
	engine query.Engine
	logger *zap.SugaredLogger
}

// NewRouter returns a new api router
func NewRouter(authCfg echojwtx.AuthConfig, engine query.Engine, l *zap.SugaredLogger) (*Router, error) {
	// Ensure auth is skipped for default endpoints.
	authCfg.JWTConfig.Skipper = echox.SkipDefaultEndpoints

	auth, err := echojwtx.NewAuth(context.Background(), authCfg)
	if err != nil {
		return nil, err
	}

	out := &Router{
		authMW: auth.Middleware(),
		engine: engine,
		logger: l.Named("api"),
	}

	return out, nil
}

// Routes will add the routes for this API version to a router group
func (r *Router) Routes(rg *echo.Group) {
	v1 := rg.Group("api/v1")
	{
		v1.Use(r.authMW)

		v1.POST("/resources/:urn/roles", r.roleCreate)
		v1.GET("/resources/:urn/roles", r.rolesList)
		v1.POST("/resources/:urn/relationships", r.relationshipsCreate)
		v1.GET("/resources/:urn/relationships", r.relationshipsList)
		v1.POST("/roles/:role_id/assignments", r.assignmentCreate)
		v1.GET("/roles/:role_id/assignments", r.assignmentsList)

		// /allow is the permissions check endpoint
		v1.GET("/allow", r.checkAction)
	}
}

func currentSubject(c echo.Context) (*urnx.URN, error) {
	subject := echojwtx.Actor(c)

	return urnx.Parse(subject)
}
