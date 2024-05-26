package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
}

const (
	UserRole            = "user"
	AdminRole           = "admin"
	RestaurantAdminRole = "restaurantAdmin"
	WaiterRole          = "waiter"
	ActivatedRole       = "activated"
	Plug                = "plug"
)
