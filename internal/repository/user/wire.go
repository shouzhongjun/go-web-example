package user

import (
	"github.com/google/wire"
)

var UserRepoSet = wire.NewSet(
	NewUserRepository, // 提供 UserRepository
)

// InitializeUserRepository 用来构造 user repository
func InitializeUserRepository() RepositoryUser {
	wire.Build(UserRepoSet)
	return nil // 这个返回值不会被执行，Wire 会生成相应代码
}
