package delivery

import (
	proto_mailer "github.com/aidostt/protos/gen/go/reservista/mailer"
	proto_reservation "github.com/aidostt/protos/gen/go/reservista/reservation"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reservista.kz/internal/domain"
)

func (h *Handler) reservation(api *gin.RouterGroup) {
	reservations := api.Group("/reservations", h.userIdentity)
	{
		reservations.GET("all/restaurant/:id", h.getAllReservationsByRestaurantId)

		activated := reservations.Group("/", h.isActivated())
		{
			activated.POST("/make", h.makeReservation)
			activated.GET("/view/:id", h.getReservation)
			activated.PATCH("/update", h.updateReservation)
			activated.DELETE("/cancel/:id", h.deleteReservationById)
			activated.GET("all/user", h.getAllReservationsByUserId)
			activated.GET("/view/restaurant/:id", h.getRestaurantByReservationId)
			activated.GET("/view/table/:id", h.getTableByReservationId)
			activated.GET("/confirm/:id", h.isPermitted([]string{domain.AdminRole, domain.RestaurantAdminRole, domain.WaiterRole}), h.confirmReservation)
		}
	}
}

func (h *Handler) makeReservation(c *gin.Context) {
	var input reservationInput
	userID, exists := c.Get(idCtx)
	if !exists {
		newResponse(c, http.StatusUnauthorized, "missing id in context")
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	resp, err := client.MakeReservation(c.Request.Context(), &proto_reservation.ReservationSQLRequest{
		UserID:          userID.(string),
		TableID:         input.TableID,
		ReservationTime: input.ReservationTime,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}
	// Sending email to user
	conn, err = h.Dialog.NewConnection(h.Dialog.Addresses.Notifications)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
	}
	mailerClient := proto_mailer.NewMailerClient(conn)
	_, err = mailerClient.SendQR(c.Request.Context(), &proto_mailer.QRInput{
		UserID:        userID.(string),
		ReservationID: resp.GetId(),
		QRUrlBase:     "http://" + h.HttpAddress + "/api/reservations/confirm/",
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusCreated, "unknown error when calling sending notification: "+err.Error())
			return
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusCreated, "failed to send reservation message: "+err.Error())
		default:
			newResponse(c, http.StatusCreated, "unknown error when sending notification: "+err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) getReservation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	reservation, err := client.GetReservation(c.Request.Context(), &proto_reservation.IDRequest{Id: id})
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
	c.JSON(http.StatusOK, reservation)
}

func (h *Handler) updateReservation(c *gin.Context) {
	var input reservationUpdateInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	statusResponse, err := client.UpdateReservation(c.Request.Context(), &proto_reservation.UpdateReservationRequest{
		ReservationID:   input.ReservationID,
		TableID:         input.TableID,
		ReservationTime: input.ReservationTime,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		}
		return
	}
	if !statusResponse.GetStatus() {
		newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": statusResponse.Status})
}

func (h *Handler) deleteReservationById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	reservation, err := client.DeleteReservationById(c.Request.Context(), &proto_reservation.IDRequest{Id: id})
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
	c.JSON(http.StatusOK, reservation)
}

func (h *Handler) confirmReservation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	ok, err := client.ConfirmReservation(c.Request.Context(), &proto_reservation.IDRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.NotFound:
			newResponse(c, http.StatusBadRequest, err.Error())
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, err.Error())
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, ok)
}

func (h *Handler) getAllReservationsByUserId(c *gin.Context) {
	userID, ok := c.Get(idCtx)
	if !ok {
		newResponse(c, http.StatusBadRequest, "unauthorized access")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	reservations, err := client.GetAllReservationByUserId(c.Request.Context(), &proto_reservation.IDRequest{Id: userID.(string)})
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
	c.JSON(http.StatusOK, reservations)
}

func (h *Handler) getAllReservationsByRestaurantId(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	reservations, err := client.GetAllReservationByRestaurantId(c.Request.Context(), &proto_reservation.IDRequest{Id: id})
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
	c.JSON(http.StatusOK, reservations)
}

func (h *Handler) getRestaurantByReservationId(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	restaurant, err := client.GetRestaurantByReservationId(c.Request.Context(), &proto_reservation.IDRequest{Id: id})
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
	c.JSON(http.StatusOK, restaurantInput{
		Id:      restaurant.GetId(),
		Name:    restaurant.GetName(),
		Address: restaurant.GetAddress(),
		Contact: restaurant.GetContact(),
	})
}

func (h *Handler) getTableByReservationId(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_reservation.NewReservationClient(conn)

	table, err := client.GetTableByReservationId(c.Request.Context(), &proto_reservation.IDRequest{Id: id})
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
	c.JSON(http.StatusOK, table)
}
