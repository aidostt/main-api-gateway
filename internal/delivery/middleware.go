package delivery

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/internal/domain"
	"reservista.kz/internal/service"
	"time"
)

const (
	authorizationHeader = "Authorization"

	userCtx     = "userId"
	ATCookieTTL = 900
	RTCookieTTL = 43200
)

func (h *Handler) userIdentity(c *gin.Context) {
	id, _, err := h.parseAuthHeader(c)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, "unauthorized access")
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
	var (
		tokens service.TokenPair
		err    error
	)
	idJWT, expired, err := h.parseAuthHeader(c)
	if idJWT != idCtx {
		newResponse(c, http.StatusUnauthorized, "unauthorized access")
		return
	}
	if !expired {
		c.Next()
		return
	}
	tokens.RefreshToken, err = c.Cookie("RT")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
			return
		}
		newResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	storedRefreshToken, err := h.services.Session.GetToken(c.Request.Context(), tokens.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if storedRefreshToken != tokens.RefreshToken {
		newResponse(c, http.StatusUnauthorized, "unauthorized access")
		return
	}
	id, err := h.tokenManager.HexToObjectID(idJWT)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	tokens, err = h.services.Session.CreateSession(c.Request.Context(), id)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.SetCookie("jwt", tokens.AccessToken, time.Now().Second()+ATCookieTTL, "/", "", false, true)
	c.SetCookie("RT", tokens.RefreshToken, time.Now().Second()+RTCookieTTL, "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *Handler) parseAuthHeader(c *gin.Context) (string, bool, error) {
	var expired bool
	token, err := c.Cookie("jwt")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", false, errors.New("unauthorized access")
		}
		return "", false, err
	}
	id, err := h.tokenManager.Parse(token)

	if err != nil {
		switch err.Error() {
		//TODO: fix the error message
		case "unauthorized access", "token has xxx elements":
			return "", false, err
		case "token is expired":
			expired = true
		default:
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
		c.AbortWithStatus(http.StatusOK)
	}
}
