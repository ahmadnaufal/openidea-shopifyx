package validation

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func Validate(data any) error {
	err := validate.Struct(data)
	if err != nil {
		strError := ""

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, vErr := range validationErrors {
				strError += fmt.Sprintf("%s;", vErr.Error())
			}
		} else {
			strError = err.Error()
		}

		return errors.New(strError)
	}

	return nil
}
