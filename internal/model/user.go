package model

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Login    string `json:"login"`
	Role     string `json:"role"`
	Password string `json:"-"`
}
