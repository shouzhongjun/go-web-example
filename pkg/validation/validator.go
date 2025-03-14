package utils

import (
	"fmt"
	"regexp"
	"sync"
)

// Validator 定义验证器接口
type Validator interface {
	Validate(interface{}) error
	ValidateField(string, string) bool
}

// CustomValidator 自定义验证器实现
type CustomValidator struct {
	rules map[string]func(string) bool
}

var (
	validate *CustomValidator
	once     sync.Once
)

// InitGlobalValidator 注册全局验证器，支持自定义校验
func InitGlobalValidator() error {
	var err error
	once.Do(func() {
		validate = &CustomValidator{
			rules: map[string]func(string) bool{
				"custom_rule":   customRule,
				"chinese_phone": validateChinesePhone,
				"chinese_id":    validateChineseID,
				"safe_password": validatePassword,
			},
		}
	})

	return err
}

// GetValidator 获取全局验证器
func GetValidator() *CustomValidator {
	if validate == nil {
		if err := InitGlobalValidator(); err != nil {
			panic(fmt.Sprintf("初始化验证器失败: %v", err))
		}
	}
	return validate
}

// Validate 验证结构体
func (v *CustomValidator) Validate(s interface{}) error {
	// 这里可以根据需要实现结构体验证逻辑
	return nil
}

// ValidateField 验证单个字段
func (v *CustomValidator) ValidateField(value string, rule string) bool {
	if fn, ok := v.rules[rule]; ok {
		return fn(value)
	}
	return false
}

// 自定义验证规则：字符串必须以 "G" 开头
func customRule(value string) bool {
	return len(value) > 0 && value[0] == 'G'
}

// validateChinesePhone 验证中国手机号
func validateChinesePhone(value string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// validateChineseID 验证中国身份证号
func validateChineseID(value string) bool {
	pattern := `^[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// validatePassword 验证密码强度
func validatePassword(value string) bool {
	// 至少8位，包含大小写字母和数字
	pattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}
