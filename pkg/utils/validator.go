package utils

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// InitGlobalValidator 注册全局验证器，支持自定义校验
func InitGlobalValidator() error {
	var err error
	once.Do(func() {
		validate = validator.New()

		// 注册自定义验证规则
		rules := map[string]validator.Func{
			"custom_rule":   customRule,
			"chinese_phone": validateChinesePhone,
			"chinese_id":    validateChineseID,
			"safe_password": validatePassword,
		}

		for tag, fn := range rules {
			if err = validate.RegisterValidation(tag, fn); err != nil {
				err = fmt.Errorf("注册验证规则 %s 失败: %w", tag, err)
				return
			}
		}
	})

	return err
}

// GetValidator 获取全局验证器
func GetValidator() *validator.Validate {
	if validate == nil {
		if err := InitGlobalValidator(); err != nil {
			panic(fmt.Sprintf("初始化验证器失败: %v", err))
		}
	}
	return validate
}

// ValidateStruct 验证结构体
func ValidateStruct(s interface{}) error {
	return GetValidator().Struct(s)
}

// 自定义验证规则：字符串必须以 "G" 开头
func customRule(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 0 && fl.Field().String()[0] == 'G'
}

// validateChinesePhone 验证中国手机号
func validateChinesePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// validateChineseID 验证中国身份证号
func validateChineseID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	pattern := `^[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`
	matched, _ := regexp.MatchString(pattern, id)
	return matched
}

// validatePassword 验证密码强度
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	// 至少8位，包含大小写字母和数字
	pattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$`
	matched, _ := regexp.MatchString(pattern, password)
	return matched
}
