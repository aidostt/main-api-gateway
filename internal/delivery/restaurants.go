package delivery

import (
	"fmt"
	proto_restaurant "github.com/aidostt/protos/gen/go/reservista/restaurant"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reservista.kz/internal/domain"
	"strconv"
	"strings"
)

func (h *Handler) restaurant(api *gin.RouterGroup) {
	restaurants := api.Group("/restaurants")
	{
		restaurants.GET("/view/:id", h.getRestaurant)
		restaurants.GET("/all", h.searchRestaurants)
		restaurants.GET("/suggestions", h.getSuggestions)
		//admin, restaurant authorities
		authenticated := restaurants.Group("/", h.userIdentity, h.isActivated(), h.isPermitted([]string{domain.AdminRole, domain.RestaurantAdminRole}))
		{
			authenticated.POST("/add", h.addRestaurant)
			authenticated.DELETE("/delete/:id", h.deleteRestaurantById)
			authenticated.PATCH("/update/:id", h.updateRestById)
			authenticated.POST("/photos/upload/:id", h.uploadRestaurantPhotos)
			authenticated.DELETE("/photos/delete/:id", h.deleteRestaurantPhoto)
		}
	}
}
func (h *Handler) searchRestaurants(c *gin.Context) {
	query := c.Query("q")
	page, err := strconv.Atoi(c.DefaultQuery("page", h.PageDefault))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", h.LimitDefault))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset := (page - 1) * limit
	searchQuery := fmt.Sprintf("%%%s%%", strings.ToLower(query))

	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_restaurant.NewRestaurantClient(conn)

	restaurants, err := client.SearchRestaurants(c.Request.Context(), &proto_restaurant.SearchRequest{
		Query:  searchQuery,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			newResponse(c, http.StatusInternalServerError, "unknown error when calling search: "+err.Error())
			return
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, restaurants)
}

func (h *Handler) getSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
		return
	}

	searchQuery := fmt.Sprintf("%%%s%%", strings.ToLower(query))

	conn, err := h.Dialog.NewConnection(h.Dialog.Addresses.Reservations)
	defer conn.Close()
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "something went wrong...")
		return
	}
	client := proto_restaurant.NewRestaurantClient(conn)

	suggestions, err := client.GetRestaurantSuggestions(c.Request.Context(), &proto_restaurant.SuggestionRequest{Query: searchQuery})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			newResponse(c, http.StatusInternalServerError, "unknown error when calling suggestions:"+err.Error())
			return
		}
		switch st.Code() {
		case codes.Internal:
			newResponse(c, http.StatusInternalServerError, err.Error())
		default:
			newResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, suggestions.Restaurants)
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
	c.JSON(http.StatusOK, restaurant)
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
