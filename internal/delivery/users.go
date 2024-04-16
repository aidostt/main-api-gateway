package delivery

import (
	"github.com/aidostt/protos/gen/go/reservista/authentication"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("/sign-up", h.userSignUp)
		users.POST("/sign-in", h.userSignIn)
		authenticated := users.Group("/", h.userIdentity)
		authenticated.Use(h.isExpired)
		{
			authenticated.GET("/healthcheck", h.healthcheck)
			authenticated.POST("/sign-out", h.signOut)
		}
	}
}

func (h *Handler) userSignUp(c *gin.Context) {
	var inp userSignUpInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	conn, err := h.dialog.NewConnection(h.dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	auth := proto_auth.NewAuthClient(conn)

	tokens, err := auth.SignUp(c.Request.Context(), &proto_auth.RegisterRequest{
		Name:     inp.Name,
		Surname:  inp.Surname,
		Phone:    inp.Phone,
		Email:    inp.Email,
		Password: inp.Password,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.AlreadyExists:
			newResponse(c, http.StatusBadRequest, "user already exists")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}

	c.SetCookie("jwt", tokens.Jwt, time.Now().Add(h.accessTokenTTL).Second(), "/", "", false, true)
	c.SetCookie("RT", tokens.Rt, time.Now().Add(h.refreshTokenTTL).Second(), "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  tokens.Jwt,
		RefreshToken: tokens.Rt,
	})
	c.Status(http.StatusCreated)
}

func (h *Handler) userSignIn(c *gin.Context) {
	var inp signInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}
	conn, err := h.dialog.NewConnection(h.dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	auth := proto_auth.NewAuthClient(conn)
	tokens, err := auth.SignIn(c.Request.Context(), &proto_auth.SignInRequest{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "wrong credentials")
		case codes.NotFound:
			newResponse(c, http.StatusBadRequest, "user not found")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}

	c.SetCookie("jwt", tokens.Jwt, time.Now().Add(h.accessTokenTTL).Second(), "/", "", false, true)
	c.SetCookie("RT", tokens.Rt, time.Now().Add(h.refreshTokenTTL).Second(), "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  tokens.Jwt,
		RefreshToken: tokens.Rt,
	})
}

func (h *Handler) signOut(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.SetCookie("RT", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, healthResponse{
		Status: "success",
	})
}
