package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/pkg/dialog"
	manager "reservista.kz/pkg/manager"
	"reservista.kz/pkg/s3client"
	"time"
)

type Handler struct {
	CookieTTL    time.Duration
	Dialog       *dialog.Dialog
	S3Client     *s3client.S3Client
	Environment  string
	TokenManager manager.TokenManager
	HttpAddress  string
	PageDefault  string
	LimitDefault string
}

func NewHandler(handler Handler) *Handler {
	return &Handler{
		Dialog:       handler.Dialog,
		CookieTTL:    handler.CookieTTL,
		Environment:  handler.Environment,
		TokenManager: handler.TokenManager,
		HttpAddress:  handler.HttpAddress,
		S3Client:     handler.S3Client,
		PageDefault:  handler.PageDefault,
		LimitDefault: handler.LimitDefault,
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
		h.qr(api)
		h.user(api)
		h.reservation(api)
	}

	return router
}
