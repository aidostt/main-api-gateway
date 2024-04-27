package delivery

import (
	proto_restaurant "github.com/aidostt/protos/gen/go/reservista/restaurant"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func (h *Handler) restaurant(api *gin.RouterGroup) {
	restaurants := api.Group("/restaurants")
	{
		restaurants.GET("/view", h.getRestaurant)
		restaurants.GET("/all", h.getAllRestaurants)
		restaurants.POST("/add", h.addRestaurant)
		restaurants.DELETE("/delete", h.deleteRestaurantById)
		restaurants.PATCH("/update", h.updateRestById)
	}
}

func (h *Handler) getAllRestaurants(c *gin.Context) {
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_restaurant.NewRestaurantClient(conn)

	restaurants, err := client.GetAllRestaurants(c.Request.Context(), &proto_restaurant.Empty{})
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
	c.JSON(http.StatusOK, restaurants.Restaurants)
}

func (h *Handler) getRestaurant(c *gin.Context) {
	var input idInput
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}
	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_restaurant.NewRestaurantClient(conn)

	restaurant, err := client.GetRestaurant(c.Request.Context(), &proto_restaurant.IDRequest{Id: input.Id})
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

func (h *Handler) addRestaurant(c *gin.Context) {
	var input restaurantInput
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
	client := proto_restaurant.NewRestaurantClient(conn)

	statusResponse, err := client.AddRestaurant(c.Request.Context(), &proto_restaurant.RestaurantObject{
		Name:    input.Name,
		Address: input.Address,
		Contact: input.Contact,
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
	if !statusResponse.GetStatus() {
		newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": statusResponse.Status})
}

func (h *Handler) updateRestById(c *gin.Context) {
	var input restaurantInput
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
	client := proto_restaurant.NewRestaurantClient(conn)

	statusResponse, err := client.UpdateRestById(c.Request.Context(), &proto_restaurant.RestaurantObject{
		Id:      input.Id,
		Name:    input.Name,
		Address: input.Address,
		Contact: input.Contact,
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
	if !statusResponse.GetStatus() {
		newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": statusResponse.Status})
}

func (h *Handler) deleteRestaurantById(c *gin.Context) {
	var input idInput
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
	client := proto_restaurant.NewRestaurantClient(conn)

	statusResponse, err := client.DeleteRestaurantById(c.Request.Context(), &proto_restaurant.IDRequest{
		Id: input.Id,
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
	if !statusResponse.GetStatus() {
		newResponse(c, http.StatusInternalServerError, "unknown error when calling sign up:"+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": statusResponse.Status})
}
