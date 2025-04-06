package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	cusErr "github.com/tuannm99/podzone/pkg/errors"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]{8,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	slugRegex     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	validate := validator.New()

	validate.RegisterValidation("email", validateEmail)
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("uuid", validateUUID)
	validate.RegisterValidation("slug", validateSlug)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: validate,
	}
}

func (v *Validator) Validate(s any) error {
	if err := v.validate.Struct(s); err != nil {
		validationErrors := v.processErrors(err)
		return validationErrors.ToAppError()
	}
	return nil
}

func (v *Validator) ValidateVar(field any, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		validationErrors := v.processErrors(err)
		return validationErrors.ToAppError()
	}
	return nil
}

func (v *Validator) processErrors(err error) *cusErr.ValidationError {
	validationErrors := cusErr.NewValidationErrors()

	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, e := range validationErrs {
				field := e.Field()
				validationErrors.Add(field, v.getErrorMessage(e))
			}
		} else {
			validationErrors.Add("general", err.Error())
		}
	}

	return validationErrors
}

func (v *Validator) getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		if e.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at least %s characters long", e.Param())
		}
		return fmt.Sprintf("Must be at least %s", e.Param())
	case "max":
		if e.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must not exceed %s characters", e.Param())
		}
		return fmt.Sprintf("Must not exceed %s", e.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", e.Param())
	case "eq":
		return fmt.Sprintf("Must be equal to %s", e.Param())
	case "ne":
		return fmt.Sprintf("Must not be equal to %s", e.Param())
	case "lt":
		return fmt.Sprintf("Must be less than %s", e.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", e.Param())
	case "gt":
		return fmt.Sprintf("Must be greater than %s", e.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", e.Param())
	case "eqfield":
		return fmt.Sprintf("Must be equal to %s", e.Param())
	case "phone":
		return "Invalid phone number format"
	case "password":
		return "Password must be at least 8 characters"
	case "username":
		return "Username can only contain alphanumeric characters, underscores, hyphens, and dots"
	case "uuid":
		return "Invalid UUID format"
	case "slug":
		return "Invalid slug format (must be lowercase letters, numbers, and hyphens)"
	default:
		return e.Error()
	}
}

func validateEmail(fl validator.FieldLevel) bool {
	return emailRegex.MatchString(fl.Field().String())
}

func validatePhone(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

func validatePassword(fl validator.FieldLevel) bool {
	return passwordRegex.MatchString(fl.Field().String())
}

func validateUsername(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}

func validateUUID(fl validator.FieldLevel) bool {
	return uuidRegex.MatchString(fl.Field().String())
}

func validateSlug(fl validator.FieldLevel) bool {
	return slugRegex.MatchString(fl.Field().String())
}

func IsEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func IsPassword(password string) bool {
	return passwordRegex.MatchString(password)
}

func IsUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func IsUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

func IsSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

func ValidateJSON(data []byte, s any) error {
	if err := json.Unmarshal(data, s); err != nil {
		return cusErr.NewValidation("invalid JSON format", err)
	}
	return nil
}
