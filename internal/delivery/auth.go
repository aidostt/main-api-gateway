package delivery

import (
	proto_auth "github.com/aidostt/protos/gen/go/reservista/authentication"
	proto_mailer "github.com/aidostt/protos/gen/go/reservista/mailer"
	proto_user "github.com/aidostt/protos/gen/go/reservista/user"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reservista.kz/internal/domain"
	"time"
)

func (h *Handler) auth(api *gin.RouterGroup) {
	users := api.Group("/auth")
	{
		users.POST("/sign-up", h.userSignUp)
		users.POST("/activate/:token", h.userActivation)
		users.POST("/sign-in", h.userSignIn)
		authenticated := users.Group("/", h.userIdentity)
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
	authClient := proto_auth.NewAuthClient(conn)
	resp, err := authClient.SignUp(c.Request.Context(), &proto_auth.SignUpRequest{
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
	conn.Close()
	h.setCookies(c, tokenResponse{
		AccessToken:  resp.Tokens.Jwt,
		RefreshToken: resp.Tokens.Rt,
	})
	go func() {
		conn, err = h.Dialog.NewConnection(h.Dialog.Addresses.Notifications)
		defer conn.Close()
		if err != nil {
			newResponse(c, http.StatusInternalServerError, "couldn't open connection with notification service")
			return
		}
		mailerClient := proto_mailer.NewMailerClient(conn)
		_, err = mailerClient.SendWelcome(c.Request.Context(), &proto_mailer.ContentInput{
			Email:   inp.Email,
			Content: "http://" + h.HttpAddress + "/api/auth/activate/" + resp.GetActivationToken(),
		})
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				// Error was not a gRPC status error
				newResponse(c, http.StatusInternalServerError, "unknown error when calling sending notification: "+err.Error())
				return
			}
			switch st.Code() {
			case codes.Internal:
				newResponse(c, http.StatusInternalServerError, "failed to send welcome message: "+err.Error())
			default:
				newResponse(c, http.StatusInternalServerError, "unknown error when sending notification: "+err.Error())
			}
			return
		}
	}()
	c.JSON(http.StatusCreated, nil)
}

//TODO: create new endpoint that will create new activation token and send it to user

func (h *Handler) userActivation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		newResponse(c, http.StatusBadRequest, "missing token in the URL")
		return
	}
	id, expiry, err := h.TokenManager.ParseActivationToken(token)
	if err != nil {
		newResponse(c, http.StatusBadRequest, "failed while parsing token: "+err.Error())
		return
	}
	if expiry.Before(time.Now()) {
		newResponse(c, http.StatusNotFound, "link is expired")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	userClient := proto_user.NewUserClient(conn)
	user, err := userClient.GetByID(c.Request.Context(), &proto_user.GetRequest{
		UserId: id,
		Email:  domain.Plug,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			newResponse(c, http.StatusInternalServerError, "unknown error when calling user activate:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "non-existent id")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when user activate:"+err.Error())
		}
		return
	}
	if user.GetActivated() {
		newResponse(c, http.StatusOK, "already activated")
		return
	}
	statusResponse, err := userClient.Activate(c.Request.Context(), &proto_user.ActivateRequest{
		UserID:   id,
		Activate: true,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			newResponse(c, http.StatusInternalServerError, "unknown error when activate user:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "invalid argument")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when activate user:"+err.Error())
		}
		return
	}
	if !statusResponse.Status {
		newResponse(c, http.StatusInternalServerError, "unknown error when activate user:"+err.Error())
		return
	}
	c.JSON(http.StatusOK, "user activated successfully!")
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
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign in:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.Unauthenticated:
			newResponse(c, http.StatusUnauthorized, "user is not verified")
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "wrong credentials")
		case codes.NotFound:
			newResponse(c, http.StatusBadRequest, "user not found")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign in:"+err.Error())
		}
		return
	}

	h.setCookies(c, tokenResponse{
		AccessToken:  tokens.Jwt,
		RefreshToken: tokens.Rt,
	})
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) signOut(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.SetCookie("RT", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, healthResponse{
		Status: "success",
	})
}

func (h *Handler) setCookies(c *gin.Context, tokens tokenResponse) {
	c.SetCookie("jwt", tokens.AccessToken, int(h.CookieTTL.Seconds()), "/", "", false, true)
	c.SetCookie("RT", tokens.RefreshToken, int(h.CookieTTL.Seconds()), "/", "", false, true)
}
