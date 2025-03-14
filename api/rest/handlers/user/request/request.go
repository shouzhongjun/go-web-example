package request

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新个人资料请求参数
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=32"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// UpdatePasswordRequest 更新密码请求参数
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// CreateUserRequest 创建用户请求参数
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

// UpdateUserRequest 更新用户请求参数
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"omitempty,oneof=user admin"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}
