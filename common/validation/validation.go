package validation

import (
	"errors"
	"fmt"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	"github.com/go-playground/validator/v10"
)

func ValidateRequestData(dataFromRequest any) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(dataFromRequest)
	if err != nil {
		if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
			return commonerrors.Validation("invalid request data")
		}

		if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			for _, e := range validateErrs {
				switch e.Tag() {
				case "required":
					return commonerrors.Validation(
						fmt.Sprintf("the %s field is required", e.Field()),
					)
				case "min":
					return commonerrors.Validation(
						fmt.Sprintf("the %s field must be at least %s characters long", e.Field(), e.Param()),
					)
				case "max":
					return commonerrors.Validation(
						fmt.Sprintf("the %s field must be at most %s characters long", e.Field(), e.Param()),
					)
				case "email":
					return commonerrors.Validation(
						fmt.Sprintf("the %s field must be a valid email address", e.Field()),
					)
				case "eqfield":
					return commonerrors.Validation(
						fmt.Sprintf("the %s field must be the same as %s field", e.Field(), e.Param()),
					)
				}
			}
			return err
		}

		return err
	}
	return nil
}
