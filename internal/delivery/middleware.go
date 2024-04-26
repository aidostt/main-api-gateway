package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	authorizationHeader = "Authorization"

	userCtx = "userId"
)

func (h *Handler) userIdentity(c *gin.Context) {
	id, _, err := h.parseAuthHeader(c)
	if err != nil {
		switch err.Error() {
		case http.ErrNoCookie.Error():
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
		case "unauthorized access", "token has xxx elements":
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
		default:
			newResponse(c, http.StatusInternalServerError, "failed to parse jwt to id")
		}
		return

	}
	c.Set(userCtx, id)
}
func (h *Handler) isExpired(c *gin.Context) {
	idCtx, exist := c.Get(userCtx)
	if !exist {
		newResponse(c, http.StatusUnauthorized, "unauthorized access")
		return
	}
	idJWT, expired, err := h.parseAuthHeader(c)
	if err != nil {
		switch err.Error() {
		case http.ErrNoCookie.Error():
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
		case "unauthorized access", "token has xxx elements":
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
		default:
			newResponse(c, http.StatusInternalServerError, "failed to parse jwt to id")
		}
		return

	}
	if idJWT != idCtx {
		newResponse(c, http.StatusUnauthorized, "unauthorized access")
		return
	}
	if !expired {
		c.Next()
		return
	}
	h.refresh(c)
}

func (h *Handler) parseAuthHeader(c *gin.Context) (string, bool, error) {
	var expired bool
	token, err := c.Cookie("jwt")
	if err != nil {
		return "", false, err
	}
	id, err := h.TokenManager.Parse(token)
	if err != nil {
		if err.Error() == "token is expired" {
		} else {
			return "", false, err
		}
	}

	return id, expired, nil
}

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		//TODO: solve problem with CORS policy
		c.AbortWithStatus(http.StatusOK)
	}
}
