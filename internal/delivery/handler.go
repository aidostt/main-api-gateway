package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/pkg/dialog"
	manager "reservista.kz/pkg/manager"
	"time"
)

type Handler struct {
	CookieTTL    time.Duration
	Dialog       *dialog.Dialog
	Environment  string
	TokenManager manager.TokenManager
	HttpAddress  string
}

func NewHandler(handler Handler) *Handler {
	return &Handler{
		Dialog:       handler.Dialog,
		CookieTTL:    handler.CookieTTL,
		Environment:  handler.Environment,
		TokenManager: handler.TokenManager,
		HttpAddress:  handler.HttpAddress,
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
		h.restaurant(api)
		h.table(api)
		api.Use(h.userIdentity)
		h.qr(api)
		h.user(api)
		h.reservation(api)
	}

	return router
}
