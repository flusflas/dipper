package godotted

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

// IsFieldError returns true when the given error is a FieldError.
func IsFieldError(err interface{}) bool {
	_, ok := err.(FieldError)
	return ok
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
