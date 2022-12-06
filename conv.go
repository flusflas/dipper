package dipper

// This file contains convenience functions using a default Dipper instance
// prepared for dot notation.

var defaultDipper = New(Options{Separator: "."})

// Fields defines an attribute-value map type, containing the requested
// attributes as the map keys and their resolved values as the map values.
// It implements convenience methods to handle returned errors.
type Fields map[string]interface{}

// Get uses a default Dipper instance to return the value of the given obj
// attribute. The attribute uses dot notation to allow accessing nested fields,
// slice elements or map keys. Field names and key maps are case-sensitive.
// All the struct fields accessed must be exported.
// If an error occurs, it will be returned as the attribute value, so it should
// be handled. All the returned errors are fieldError.
//
// Example:
//
//	v := Get(myObj, "SomeStructField.1.some_key_map")
//	if err := Error(v); err != nil {
//	    return err
//	}
func Get(obj interface{}, attribute string) interface{} {
	return defaultDipper.Get(obj, attribute)
}

// GetMany returns a map with the values of the given obj attributes.
// It works as Get(), but it takes a slice of attributes to return their
// corresponding values. The returned map will have the same length as the
// attributes slice, with the attributes as keys.
//
// Example:
//
//	v := GetMany(myObj, []string{"Name", "Age", "Skills.skydiving})
//	if err := v.FirstError(); err != nil {
//	    return err
//	}
func GetMany(obj interface{}, attributes []string) Fields {
	return defaultDipper.GetMany(obj, attributes)
}

// Set uses a default Dipper instance to set the value of the given obj
// attribute to the new provided value.
// The attribute uses dot notation to allow accessing nested fields, slice
// elements or map keys. Field names and key maps are case-sensitive.
// All the struct fields accessed must be exported.
// ErrUnaddressable will be returned if obj is not addressable.
// It returns nil if the value was successfully set, otherwise it will return
// a fieldError.
//
// Example:
//
//	v := Set(&myObj, "SomeStructField.1.some_key_map", 123)
//	if err != nil {
//	    return err
//	}
func Set(obj interface{}, attribute string, new interface{}) error {
	return defaultDipper.Set(obj, attribute, new)
}
