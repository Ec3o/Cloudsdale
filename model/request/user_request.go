package request

type UserRegisterRequest struct {
	Username string `binding:"required" json:"username"`
	Email    string `binding:"required" json:"email"`
	Password string `binding:"required" json:"password"`
}

type UserLoginRequest struct {
	Username string `binding:"required" json:"username"`
	Password string `binding:"required" json:"password"`
}

type UserLogoutRequest struct {
	Username string `binding:"required" json:"username"`
}

type UserUpdateRequest struct {
	Id       string `binding:"required" json:"id"`
	Username string `binding:"required,max=20,min=3" json:"username"`
	Password string `binding:"required,min=6" json:"password"`
	Email    string `binding:"email" json:"email"`
}