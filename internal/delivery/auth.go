package delivery

import (
	"context"
	proto_auth "github.com/aidostt/protos/gen/go/reservista/authentication"
	proto_mailer "github.com/aidostt/protos/gen/go/reservista/mailer"
	proto_user "github.com/aidostt/protos/gen/go/reservista/user"
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
		authenticated := users.Use(h.userIdentity)
		{
			authenticated.POST("/activate", h.userActivation)
			authenticated.GET("/new-activation-code", h.sendNewVerificationCode)
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
	h.setCookies(c, tokenResponse{
		AccessToken:  resp.Tokens.Jwt,
		RefreshToken: resp.Tokens.Rt,
	})
	err = h.sendVerificationCodeMail(c.Request.Context(), inp.Email, resp.GetActivationToken())
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusCreated, "unknown error when calling sending notification: "+err.Error())
			return
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusCreated, "failed to send welcome message: "+err.Error())
		default:
			newResponse(c, http.StatusCreated, "unknown error when sending notification: "+err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, StatusResponse{Status: true})
}

func (h *Handler) userActivation(c *gin.Context) {
	id, exists := c.Get(idCtx)
	if !exists {
		newResponse(c, http.StatusUnauthorized, "missing id in context")
	}
	var code codeInput
	if err := c.BindJSON(&code); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	dbCode, _, err := h.verificationCode(c, id.(string))
	if err != nil {
		return
	}
	if code.Code != dbCode {
		newResponse(c, http.StatusBadRequest, "renew your activation code")
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
		UserId: id.(string),
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
			newResponse(c, http.StatusBadRequest, "invalid argument: "+err.Error())
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}
	if user.Activated {
		newResponse(c, http.StatusOK, "already activated")
		return
	}
	statusResponse, err := userClient.Activate(c.Request.Context(), &proto_user.ActivateRequest{
		UserID:   id.(string),
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
			newResponse(c, http.StatusBadRequest, err.Error())
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
	conn, err = h.Dialog.NewConnection(h.Dialog.Addresses.Notifications)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	mailerClient := proto_mailer.NewMailerClient(conn)
	_, err = mailerClient.SendWelcome(c.Request.Context(), &proto_mailer.ContentInput{
		Email:   user.GetEmail(),
		Content: "localhost:3000",
	})
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}

	h.refresh(c)
}

func (h *Handler) sendNewVerificationCode(c *gin.Context) {
	id, exists := c.Get(idCtx)
	if !exists {
		newResponse(c, http.StatusUnauthorized, "missing id in context")
	}
	code, email, err := h.verificationCode(c, id.(string))
	if err != nil {
		return
	}
	err = h.sendVerificationCodeMail(c.Request.Context(), email, code)
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
	c.JSON(http.StatusOK, nil)
}

func (h *Handler) sendVerificationCodeMail(ctx context.Context, email string, code string) error {
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Notifications)
	defer conn.Close()
	if err != nil {
		return err
	}
	mailerClient := proto_mailer.NewMailerClient(conn)
	_, err = mailerClient.SendAuthCode(ctx, &proto_mailer.ContentInput{
		Email:   email,
		Content: code,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) verificationCode(c *gin.Context, id string) (string, string, error) {
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return "", "", err
	}
	userClient := proto_user.NewUserClient(conn)
	codeResponse, err := userClient.VerificationCode(c.Request.Context(), &proto_user.GetRequest{UserId: id})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sending notification: "+err.Error())
			return "", "", err
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, err.Error())
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "not found in db: "+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when getting verification code: "+err.Error())
		}
		return "", "", err
	}
	return codeResponse.GetCode(), codeResponse.GetEmail(), err
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
