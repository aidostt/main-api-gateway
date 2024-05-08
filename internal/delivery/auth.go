package delivery

import (
	"github.com/aidostt/protos/gen/go/reservista/authentication"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func (h *Handler) auth(api *gin.RouterGroup) {
	users := api.Group("/auth")
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
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_auth.NewAuthClient(conn)

	tokens, err := client.SignUp(c.Request.Context(), &proto_auth.RegisterRequest{
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

	h.setCookies(c, tokenResponse{
		AccessToken:  tokens.Jwt,
		RefreshToken: tokens.Rt,
	})
}

func (h *Handler) userSignIn(c *gin.Context) {
	var inp signInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_auth.NewAuthClient(conn)
	tokens, err := client.SignIn(c.Request.Context(), &proto_auth.SignInRequest{
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

	h.setCookies(c, tokenResponse{
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

func (h *Handler) setCookies(c *gin.Context, tokens tokenResponse) {
	c.SetCookie("jwt", tokens.AccessToken, int(h.AccessTokenTTL.Seconds()), "/", "", false, true)
	c.SetCookie("RT", tokens.RefreshToken, int(h.RefreshTokenTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, tokens)
}
