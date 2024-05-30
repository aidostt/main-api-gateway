package delivery

import (
	proto_restaurant "github.com/aidostt/protos/gen/go/reservista/restaurant"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reservista.kz/internal/domain"
)

//TODO: create some middleware, that will collect data of requested restaurant and use this information in endpoints

func (h *Handler) restaurant(api *gin.RouterGroup) {
	restaurants := api.Group("/restaurants")
	{
		restaurants.GET("/view/:id", h.getRestaurant)
		restaurants.GET("/all", h.getAllRestaurants)

		//admin, restaurant authorities
		restaurants.Use(h.userIdentity)
		restaurants.Use(h.isActivated())
		restaurants.Use(h.isPermitted([]string{domain.AdminRole, domain.RestaurantAdminRole}))
		restaurants.POST("/add", h.addRestaurant)
		restaurants.DELETE("/delete/:id", h.deleteRestaurantById)
		restaurants.PATCH("/update/:id", h.updateRestById)
		restaurants.POST("/photos/upload/:id", h.uploadRestaurantPhotos)
		restaurants.DELETE("/photos/delete/:id", h.deleteRestaurantPhoto)

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
	client := proto_restaurant.NewRestaurantClient(conn)

	restaurant, err := client.GetRestaurant(c.Request.Context(), &proto_restaurant.IDRequest{Id: id})
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
		Photos:  restaurant.GetImageUrls(),
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
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}
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
		Id:      id,
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
	client := proto_restaurant.NewRestaurantClient(conn)

	statusResponse, err := client.DeleteRestaurantById(c.Request.Context(), &proto_restaurant.IDRequest{
		Id: id,
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

func (h *Handler) uploadRestaurantPhotos(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		newResponse(c, http.StatusBadRequest, "failed to parse multipart form")
		return
	}
	files := form.File["photos"]
	if len(files) == 0 {
		newResponse(c, http.StatusBadRequest, "no files uploaded")
		return
	}

	var urls []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			newResponse(c, http.StatusInternalServerError, "failed to open file")
			return
		}
		defer file.Close()

		url, err := h.S3Client.UploadFile(c.Request.Context(), file, fileHeader)
		if err != nil {
			newResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		urls = append(urls, url)
	}

	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_restaurant.NewRestaurantClient(conn)

	statusResponse, err := client.UploadPhotos(c.Request.Context(), &proto_restaurant.UploadPhotoRequest{
		RestaurantID: id,
		Urls:         urls,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			newResponse(c, http.StatusInternalServerError, "unknown error when calling upload photos: "+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "invalid argument: "+err.Error())
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality: "+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling upload photos: "+err.Error())
		}
		return
	}
	if !statusResponse.GetStatus() {
		newResponse(c, http.StatusInternalServerError, "failed to upload photos")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "urls": urls})
}
func (h *Handler) deleteRestaurantPhoto(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "missing ID in the URL")
		return
	}

	var input struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input")
		return
	}

	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_restaurant.NewRestaurantClient(conn)

	statusResponse, err := client.DeletePhoto(c.Request.Context(), &proto_restaurant.DeletePhotoRequest{
		RestaurantID: id,
		Url:          input.URL,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			newResponse(c, http.StatusInternalServerError, "unknown error when calling delete photo: "+err.Error())
			return
		}
		switch st.Code() {
		case codes.InvalidArgument:
			newResponse(c, http.StatusBadRequest, "invalid argument: "+err.Error())
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, "microservice failed to execute functionality: "+err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, "unknown error when calling delete photo: "+err.Error())
		}
		return
	}
	if !statusResponse.GetStatus() {
		newResponse(c, http.StatusInternalServerError, "failed to delete photo")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
