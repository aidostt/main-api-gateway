package delivery

import (
	proto_qr "github.com/aidostt/protos/gen/go/reservista/qr"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reservista.kz/internal/domain"
)

func (h *Handler) qr(api *gin.RouterGroup) {
	qr := api.Group("/qr", h.userIdentity)
	{
		qr.POST("/generate", h.generateQR)
		qr.GET("/scan/:reservationID", h.isPermitted([]string{domain.AdminRole, domain.WaiterRole, domain.RestaurantAdminRole}), h.scanQR)

	}
}

func (h *Handler) generateQR(c *gin.Context) {
	var inp qrInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.QRs)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	qr := proto_qr.NewQRClient(conn)
	resp, err := qr.Generate(c.Request.Context(), &proto_qr.GenerateRequest{
		Content: "http://" + h.HttpAddress + "/api/reservations/confirm/" + inp.ReservationID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling generate qr:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "invalid argument")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling generate qr:"+err.Error())
		}
		return
	}
	c.Header("Content-Type", "image/png")
	c.Writer.Write(resp.GetQR())
}

func (h *Handler) scanQR(c *gin.Context) {
	reservationID := c.Param("reservationID")
	if reservationID == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.QRs)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	qr := proto_qr.NewQRClient(conn)
	userID, ok := c.Get(idCtx)
	if !ok {
		newResponse(c, http.StatusBadRequest, "unauthorized access")
		return
	}
	resp, err := qr.Scan(c.Request.Context(), &proto_qr.ScanRequest{
		UserID:        userID.(string),
		ReservationID: reservationID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Error was not a gRPC status error
			newResponse(c, http.StatusInternalServerError, "unknown error when calling scan qr:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "invalid argument")
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality:"+err.Error())
		case codes.Unauthenticated:
			newResponse(c, http.StatusInternalServerError, "unauthorized access")
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling scan qr:"+err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}
