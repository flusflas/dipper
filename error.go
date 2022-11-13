package godotted

import "reflect"

type FieldError string

func (e FieldError) Error() string {
	return string(e)
}

const (
	ErrNotFound        FieldError = "field not found"
	ErrInvalidIndex    FieldError = "invalid index"
	ErrIndexOutOfRange FieldError = "index out of range"
	ErrMapKeyNotString FieldError = "map key is not of string type"
	ErrUnexported      FieldError = "field is unexported"
)

// pre-calculated reflect.Value of errors to increase performance
var (
	errNotFoundValue        = reflect.ValueOf(ErrNotFound)
	errInvalidIndexValue    = reflect.ValueOf(ErrInvalidIndex)
	errIndexOutOfRangeValue = reflect.ValueOf(ErrIndexOutOfRange)
	errMapKeyNotStringValue = reflect.ValueOf(ErrMapKeyNotString)
	errUnexportedValue      = reflect.ValueOf(ErrUnexported)
)

// IsFieldError returns true when the given error is a FieldError.
func IsFieldError(err interface{}) bool {
	_, ok := err.(FieldError)
	return ok
}

// HasErrors returns true if this Fields map has any FieldError.
func (f Fields) HasErrors() bool {
	return f.FirstError() != nil
}

// FirstError returns the first FieldError found in this Fields map.
func (f Fields) FirstError() error {
	for _, v := range f {
		if IsFieldError(v) {
			return v.(FieldError)
		}
	}
	return nil
}
