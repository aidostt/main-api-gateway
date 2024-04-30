package delivery

import (
	proto_user "github.com/aidostt/protos/gen/go/reservista/user"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reservista.kz/internal/domain"
)

func (h *Handler) user(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.GET("/get-by-id", h.getByID)
		users.GET("/get-by-email", h.getByEmail)
		users.DELETE("/delete", h.deleteUser)
		users.PATCH("/update", h.updateUser)
		authenticated := users.Group("/", h.userIdentity)
		authenticated.Use(h.isExpired)
		{
		}
	}
}

func (h *Handler) isAdmin(c *gin.Context) {
	//TODO: implement
}

func (h *Handler) updateUser(c *gin.Context) {
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
	client := proto_user.NewUserClient(conn)
	userID, ok := c.Get(userCtx)
	if !ok {
		newResponse(c, http.StatusBadRequest, "unauthorized access")
		return
	}
	statusResponse, err := client.Update(c.Request.Context(), &proto_user.UpdateRequest{
		Id:       userID.(string),
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
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "invalid argument")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}
	if !statusResponse.Status {
		newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) deleteUser(c *gin.Context) {
	var inp getUserInput
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
	client := proto_user.NewUserClient(conn)
	userID, ok := c.Get(userCtx)
	if !ok {
		newResponse(c, http.StatusBadRequest, "unauthorized access")
		return
	}
	statusResponse, err := client.Delete(c.Request.Context(), &proto_user.GetRequest{
		UserId: userID.(string),
		Email:  inp.Email,
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
			newResponse(c, http.StatusBadRequest, "invalid argument")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}
	if !statusResponse.Status {
		newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) getByID(c *gin.Context) {
	var inp getUserInput
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
	client := proto_user.NewUserClient(conn)
	user, err := client.GetByID(c.Request.Context(), &proto_user.GetRequest{
		UserId: inp.Id,
		Email:  domain.Plug,
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

	c.JSON(http.StatusOK, userSignUpInput{
		Name:    user.GetName(),
		Surname: user.GetSurname(),
		Phone:   user.GetPhone(),
		Email:   user.GetEmail(),
	})
}

func (h *Handler) getByEmail(c *gin.Context) {
	var inp getUserInput
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
	client := proto_user.NewUserClient(conn)
	user, err := client.GetByEmail(c.Request.Context(), &proto_user.GetRequest{
		UserId: domain.Plug,
		Email:  inp.Email,
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

	c.JSON(http.StatusOK, userSignUpInput{
		Name:    user.GetName(),
		Surname: user.GetSurname(),
		Phone:   user.GetPhone(),
		Email:   user.GetEmail(),
	})
}
