package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStructRequest(data interface{}) error {
	err := validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				return fmt.Errorf("Field '%s' is required", err.Field())
			case "min":
				return fmt.Errorf("Field '%s' must be at least %s characters long", err.Field(), err.Param())
			case "max":
				return fmt.Errorf("Field '%s' must be at most %s characters long", err.Field(), err.Param())
			case "gte":
				return fmt.Errorf("Field '%s' must be greater than or equal to %s", err.Field(), err.Param())
			case "lte":
				return fmt.Errorf("Field '%s' must be less than or equal to %s", err.Field(), err.Param())
			case "gt":
				return fmt.Errorf("Field '%s' must be greater than %s", err.Field(), err.Param())
			case "lt":
				return fmt.Errorf("Field '%s' must be less than %s", err.Field(), err.Param())
			}
		}
	}
	return nil
}
