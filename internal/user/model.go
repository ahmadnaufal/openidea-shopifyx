package user

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=5,max=15"`
	Name     string `json:"name" validate:"required,min=5,max=50"`
	Password string `json:"password" validate:"required,min=5,max=15"`
}

type AuthenticateRequest struct {
	Username string `json:"username" validate:"required,min=5,max=15"`
	Password string `json:"password" validate:"required,min=5,max=15"`
}

type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	Name     string `db:"name"`
	Password string `db:"password"`
}
