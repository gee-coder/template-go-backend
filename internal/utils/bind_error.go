package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
)

var bindFieldLabels = map[string]string{
	"Username":     "用户名",
	"Password":     "密码",
	"RefreshToken": "刷新令牌",
	"Nickname":     "昵称",
	"Email":        "邮箱",
	"Phone":        "手机号",
	"Status":       "状态",
	"RoleIDs":      "角色",
	"Name":         "名称",
	"Code":         "编码",
	"Remark":       "备注",
	"MenuIDs":      "菜单",
	"ParentID":     "上级菜单",
	"Title":        "标题",
	"Path":         "路径",
	"Component":    "组件路径",
	"Icon":         "图标",
	"Type":         "类型",
	"Permission":   "权限标识",
	"Sort":         "排序",
	"Hidden":       "隐藏状态",
	"Company":      "公司名称",
	"Message":      "咨询内容",
	"Source":       "来源",
}

// BindErrorMessage converts request bind errors to Chinese user-facing messages.
func BindErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, io.EOF) {
		return "请求参数不能为空"
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "请求体格式错误，请检查 JSON 内容"
	}

	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		return "请求参数类型不正确"
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) && len(validationErrs) > 0 {
		fieldErr := validationErrs[0]
		label := bindFieldLabels[fieldErr.Field()]
		if label == "" {
			label = fieldErr.Field()
		}

		switch fieldErr.Tag() {
		case "required":
			return fmt.Sprintf("%s不能为空", label)
		case "email":
			return fmt.Sprintf("%s格式不正确", label)
		case "oneof":
			return fmt.Sprintf("%s取值不合法", label)
		case "max":
			return fmt.Sprintf("%s长度不能超过%s", label, fieldErr.Param())
		case "min":
			return fmt.Sprintf("%s长度不能少于%s", label, fieldErr.Param())
		case "gte":
			return fmt.Sprintf("%s不能小于%s", label, fieldErr.Param())
		case "lte":
			return fmt.Sprintf("%s不能大于%s", label, fieldErr.Param())
		default:
			return fmt.Sprintf("%s格式不正确", label)
		}
	}

	return "请求参数不合法"
}
