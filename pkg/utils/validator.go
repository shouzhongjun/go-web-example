package utils

import (
	"github.com/go-playground/validator/v10"
)

var (
	globalValidator *validator.Validate
)

// InitGlobalValidator 初始化全局验证器
func InitGlobalValidator() error {
	globalValidator = validator.New()
	return nil
}

// GetValidator 获取全局验证器实例
func GetValidator() *validator.Validate {
	return globalValidator
}
