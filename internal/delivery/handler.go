package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/pkg/dialog"
	manager "reservista.kz/pkg/manager"
	"time"
)

type Handler struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Dialog          *dialog.Dialog
	Environment     string
	TokenManager    manager.TokenManager
	HttpAddress     string
}

func NewHandler(handler Handler) *Handler {
	return &Handler{
		Dialog:          handler.Dialog,
		AccessTokenTTL:  handler.AccessTokenTTL,
		RefreshTokenTTL: handler.RefreshTokenTTL,
		Environment:     handler.Environment,
		TokenManager:    handler.TokenManager,
		HttpAddress:     handler.HttpAddress,
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

		h.auth(api)
		h.qr(api)
		h.user(api)
		h.restaurant(api)
		h.reservation(api)
		h.table(api)
	}

	return router
}
