package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status: "ok",
	})
}
