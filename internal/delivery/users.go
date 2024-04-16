package delivery

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/internal/domain"
	"reservista.kz/internal/service"
	"time"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("/sign-up", h.userSignUp)
		users.POST("/sign-in", h.userSignIn)
		users.POST("/auth/refresh", h.userRefresh)
		authenticated := users.Group("/", h.userIdentity)
		authenticated.Use(h.isExpired)
		{
			authenticated.GET("/healthcheck", h.healthcheck)
			users.POST("/sign-out", h.logout)

		}
	}
}

func (h *Handler) userSignUp(c *gin.Context) {
	var inp userSignUpInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	tokens, err := h.services.Users.SignUp(c.Request.Context(), service.UserSignUpInput{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.SetCookie("jwt", tokens.AccessToken, time.Now().Second()+ATCookieTTL, "/", "", false, true)
	c.SetCookie("RT", tokens.RefreshToken, time.Now().Second()+RTCookieTTL, "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
	c.Status(http.StatusCreated)
}

func (h *Handler) userSignIn(c *gin.Context) {
	var inp signInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	tokens, err := h.services.Users.SignIn(c.Request.Context(), service.UserSignInInput{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}

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

func (h *Handler) userRefresh(c *gin.Context) {
	token, err := c.Cookie("RT")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	tokens, err := h.services.Session.RefreshTokens(c.Request.Context(), token)
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

func (h *Handler) logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.SetCookie("RT", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, healthResponse{
		Status: "success",
	})
}
