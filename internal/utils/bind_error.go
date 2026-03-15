package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
)

var bindFieldLabels = map[string]string{
	"Account":        "account",
	"Username":       "username",
	"Password":       "password",
	"LoginType":      "login type",
	"RegisterType":   "register type",
	"RefreshToken":   "refresh token",
	"Nickname":       "nickname",
	"Email":          "email",
	"Phone":          "phone",
	"Status":         "status",
	"RoleIDs":        "roles",
	"Name":           "name",
	"Code":           "code",
	"Remark":         "remark",
	"MenuIDs":        "menus",
	"ParentID":       "parent menu",
	"Title":          "title",
	"Path":           "path",
	"Component":      "component",
	"Icon":           "icon",
	"Type":           "type",
	"Permission":     "permission",
	"Sort":           "sort",
	"Hidden":         "hidden",
	"Company":        "company",
	"Message":        "message",
	"Source":         "source",
	"AppTitle":       "app title",
	"ConsoleName":    "console name",
	"ProductTagline": "product tagline",
	"LogoMarkURL":    "logo mark url",
	"LoginHeroURL":   "login hero url",
	"Primary":        "primary color",
	"PrimaryDark":    "primary dark color",
	"ShellStart":     "shell start color",
	"ShellEnd":       "shell end color",
	"HeroStart":      "hero start color",
	"HeroEnd":        "hero end color",
}

// BindErrorMessage converts bind errors to readable user-facing messages.
func BindErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, io.EOF) {
		return "request body cannot be empty"
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "request body must be valid JSON"
	}

	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		return "request field type is invalid"
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
			return fmt.Sprintf("%s is required", label)
		case "email":
			return fmt.Sprintf("%s format is invalid", label)
		case "oneof":
			return fmt.Sprintf("%s value is invalid", label)
		case "max":
			return fmt.Sprintf("%s must be at most %s characters", label, fieldErr.Param())
		case "min":
			return fmt.Sprintf("%s must be at least %s characters", label, fieldErr.Param())
		case "gte":
			return fmt.Sprintf("%s must be greater than or equal to %s", label, fieldErr.Param())
		case "lte":
			return fmt.Sprintf("%s must be less than or equal to %s", label, fieldErr.Param())
		default:
			return fmt.Sprintf("%s format is invalid", label)
		}
	}

	return "request parameters are invalid"
}
