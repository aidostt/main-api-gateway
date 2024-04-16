package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/internal/service"
	auth "reservista.kz/pkg/manager"
)

type Handler struct {
	services     *service.Services
	tokenManager auth.TokenManager
}

func NewHandler(services *service.Services, tokenManager auth.TokenManager) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
	}
}

func (h *Handler) initAPI(router *gin.Engine) {
	handler := NewHandler(h.services, h.tokenManager)
	api := router.Group("/api")
	{
		handler.initUsersRoutes(api)
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

	h.initAPI(router)

	return router
}
