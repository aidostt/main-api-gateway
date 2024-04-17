package delivery

type userSignUpInput struct {
	Name     string `json:"name" binding:"required,max=64"`
	Surname  string `json:"surname" binding:"required,max=64"`
	Phone    string `json:"phone" binding:"required,max=64"`
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type reservationRegisterInput struct {
	Name    string `json:"restaurant_name" binding:"required,min=8,max=64"`
	Address string `json:"restaurant_address" binding:"required,min=8,max=64"`
	TableID int32  `json:"table_id" binding:"required,max=64"`
	Time    string `json:"reservation_time" binding:"required,min=8,max=64"`
}

type signInInput struct {
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type refreshInput struct {
	Token string `json:"token" binding:"required"`
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

type scanResponse struct {
	user        userSignUpInput
	reservation reservationRegisterInput
}
