package delivery

type userSignUpInput struct {
	Name     string `json:"name" binding:"required,max=64"`
	Surname  string `json:"surname" binding:"required,max=64"`
	Phone    string `json:"phone" binding:"required,max=64"`
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type restaurantInput struct {
	Id      string   `json:"id"`
	Name    string   `json:"restaurant_name" binding:"required,min=8,max=64"`
	Address string   `json:"restaurant_address" binding:"required,min=8,max=64"`
	Contact string   `json:"restaurant_contact" binding:"required,max=64"`
	Photos  []string `json:"restaurant_photos" binding:"max=64"`
}

type reservationInput struct {
	TableID         string `json:"table_id"`
	ReservationTime string `json:"reservation_time"`
}

type idInput struct {
	ID string `json:"id"`
}

type reservationUpdateInput struct {
	ReservationID   string `json:"reservation_id"`
	TableID         string `json:"table_id"`
	ReservationTime string `json:"reservation_time"`
}

type tableInput struct {
	Id            string `json:"id"`
	NumberOfSeats int32  `json:"number_of_seats"`
	IsReserved    bool   `json:"is_reserved"`
	TableNumber   int32  `json:"table_number"`
	RestaurantID  string `json:"restaurant_id"`
}

type signInInput struct {
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type StatusResponse struct {
	Status bool `json:"status" binding:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type qrInput struct {
	ReservationID string `json:"reservation_id"`
}

type getUserInput struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

type codeInput struct {
	Code string `json:"code"`
}
