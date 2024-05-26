package dipper

// fieldError is an error indicating a wrong operation getting or setting a
// value using the dipper package.
type fieldError string

func (e fieldError) Error() string {
	return string(e)
}

// Field errors returned when an attribute cannot be accessed or set.
const (
	// ErrNotFound is the error returned when an attribute is not found.
	// Depending on the type of the accessed attribute, it can mean that the
	// attribute does not exist as a struct field or as a map key.
	ErrNotFound = fieldError("dipper: field not found")
	// ErrInvalidIndex is the error returned when an attribute references a
	// slice/array element, but the given index is not a number.
	ErrInvalidIndex = fieldError("dipper: invalid index")
	// ErrIndexOutOfRange is the error returned when an attribute references a
	// slice/array element, but the given index is less than 0 or greater than
	// the size of the slice/array.
	ErrIndexOutOfRange = fieldError("dipper: index out of range")
	// ErrMapKeyNotString is the error returned when an attribute references a
	// map whose keys are not of string type.
	ErrMapKeyNotString = fieldError("dipper: map key is not of string type")
	// ErrUnexported is the error returned when an attribute references an
	// unexported struct field.
	ErrUnexported = fieldError("dipper: field is unexported")
	// ErrUnaddressable is the error returned from a set operation when an
	// attribute references a value that is not addressable.
	ErrUnaddressable = fieldError("dipper: field is unaddressable")
	// ErrTypesDoNotMatch is the error returned from a set operation when an
	// attribute references a value that has a different type than the new value.
	ErrTypesDoNotMatch = fieldError("dipper: value type does not match field type")
	// ErrInvalidFilterExpression is the error returned when the format of the
	// given search expression is invalid.
	ErrInvalidFilterExpression = fieldError("dipper: invalid search expression")
	// ErrFilterNotFound is the error returned when a search expression does not
	// match any entries.
	ErrFilterNotFound = fieldError("dipper: no matches for filter expression")
	// ErrInvalidFilterValue is the error returned when a search expression has an
	// invalid value.
	ErrInvalidFilterValue = fieldError("dipper: invalid value for filter expression")
)

// IsFieldError returns true when the given value is a fieldError.
func IsFieldError(v interface{}) bool {
	_, ok := v.(fieldError)
	return ok
}

// Error casts the given value to fieldError if possible, otherwise returns nil.
func Error(v interface{}) error {
	if err, ok := v.(fieldError); ok {
		return err
	}
	return nil
}

// HasErrors returns true if this Fields map has any fieldError.
func (f Fields) HasErrors() bool {
	return f.FirstError() != nil
}

// FirstError returns the first fieldError found in this Fields map.
func (f Fields) FirstError() error {
	for _, v := range f {
		if IsFieldError(v) {
			return v.(fieldError)
		}
	}
	return nil
}
