package user

import "errors"

// 定义错误
var (
	ErrDBNotConnected = errors.New("数据库未连接")
)
