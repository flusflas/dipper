package godotted

// FieldError is an error indicating a wrong operation getting or setting a
// value using the godotted package.
type FieldError string

func (e FieldError) Error() string {
	return string(e)
}

const (
	// ErrNotFound is the error returned when an attribute is not found.
	// Depending on the type of the accessed attribute, it can mean that the
	// attribute does not exist as a struct field or as a map key.
	ErrNotFound FieldError = "field not found"
	// ErrInvalidIndex is the error returned when an attribute references a
	// slice/array element, but the given index is not a number.
	ErrInvalidIndex FieldError = "invalid index"
	// ErrIndexOutOfRange is the error returned when an attribute references a
	// slice/array element, but the given index is less than 0 or greater than
	// the size of the slice/array.
	ErrIndexOutOfRange FieldError = "index out of range"
	// ErrMapKeyNotString is the error returned when an attribute references a
	// map whose keys are not of string type.
	ErrMapKeyNotString FieldError = "map key is not of string type"
	// ErrUnexported is the error returned when an attribute references an
	// unexported struct field.
	ErrUnexported FieldError = "field is unexported"
	// ErrUnaddressable is the error returned from a set operation when an
	// attribute references a value that is not addressable.
	ErrUnaddressable FieldError = "field is unaddressable"
	// ErrTypesDoNotMatch is the error returned from a set operation when an
	// attribute references a value that has a different type than the new value.
	ErrTypesDoNotMatch FieldError = "value type does not match field type"
)

// IsFieldError returns true when the given value is a FieldError.
func IsFieldError(v interface{}) bool {
	_, ok := v.(FieldError)
	return ok
}

// Error casts the given value to FieldError if possible, otherwise returns nil.
func Error(v interface{}) error {
	if err, ok := v.(FieldError); ok {
		return err
	}
	return nil
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
