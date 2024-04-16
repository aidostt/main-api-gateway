package delivery

import (
	"errors"
	"github.com/aidostt/protos/gen/go/reservista/authentication"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

func (h *Handler) refresh(c *gin.Context) {
	jwt, err := c.Cookie("jwt")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	rt, err := c.Cookie("RT")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	conn, err := h.dialog.NewConnection(h.dialog.Addresses.Users)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	auth := proto_auth.NewAuthClient(conn)

	tokens, err := auth.Refresh(c.Request.Context(), &proto_auth.TokenRequest{
		Jwt: jwt,
		Rt:  rt,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.Unauthenticated:
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
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
