package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/pkg/dialog"
	manager "reservista.kz/pkg/manager"
	"time"
)

type Dependencies struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Dialog          *dialog.Dialog
	Environment     string
	TokenManager    manager.TokenManager
}

type Handler struct {
	dialog          *dialog.Dialog
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	environment     string
	tokenManager    manager.TokenManager
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		dialog:          deps.Dialog,
		accessTokenTTL:  deps.AccessTokenTTL,
		refreshTokenTTL: deps.RefreshTokenTTL,
		environment:     deps.Environment,
		tokenManager:    deps.TokenManager,
	}
}

func (h *Handler) Init() *gin.Engine {
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		corsMiddleware,
	)

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	api := router.Group("/api")
	{
		h.initUsersRoutes(api)
	}

	return router
}
