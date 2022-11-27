// Package godotted implements functions to access (get or set) values of a
// generic type using dot notation.
package godotted

import (
	"reflect"
	"strconv"
	"strings"
)

type setOption int

const (
	// Zero is used as the new value in Set() to set the attribute to its zero
	// value (e.g. "" for string, nil for interface{}, etc.).
	Zero setOption = 0
	// Delete is used as the new value in Set() to delete a map key. If the
	// field is not a map value, the value will be zeroed (see Zero).
	Delete setOption = 1
)

// Fields defines an attribute-value map type, containing the requested
// attributes as the map keys and their resolved values as the map values.
// It implements convenience methods to handle returned errors.
type Fields map[string]interface{}

// Get returns the value of the given obj attribute. The attribute uses
// dot-notation to allow accessing nested fields, slice elements or map keys.
// Field names and key maps are case-sensitive.
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
	value, _, err := getReflectValue(reflect.ValueOf(obj), attribute, ".", false)
	if err != nil {
		return err
	}
	return value.Interface()
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
	m := make(Fields, len(attributes))

	for _, attr := range attributes {
		if _, ok := m[attr]; !ok {
			m[attr] = Get(obj, attr)
		}
	}

	return m
}

// Set sets the value of the given obj attribute to the specified new value.
// The attribute uses dot-notation to allow accessing nested fields, slice
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
	var err error

	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	var lastField string
	value, lastField, err = getReflectValue(value, attribute, ".", true)
	if err != nil {
		return err
	}

	var optZero, optDelete bool

	var newValue reflect.Value
	switch new {
	case Zero:
		optZero = true
	case Delete:
		optDelete = true
	default:
		newValue = reflect.ValueOf(new)
		if newValue.Kind() == reflect.Ptr {
			newValue = newValue.Elem()
		}
	}

	if value.Kind() == reflect.Map {
		if !optZero && !optDelete {
			mapValueType := value.Type().Elem()
			if mapValueType.Kind() != reflect.Interface && mapValueType != newValue.Type() {
				return ErrTypesDoNotMatch
			}
		}

		// Initialize map if needed
		if value.IsNil() {
			keyType := value.Type().Key()
			valueType := value.Type().Elem()
			mapType := reflect.MapOf(keyType, valueType)
			value.Set(reflect.MakeMapWithSize(mapType, 0))
		}

		value.SetMapIndex(reflect.ValueOf(lastField), newValue)
	} else {
		if !optZero && !optDelete {
			if !value.CanAddr() {
				return ErrUnaddressable
			}
			if value.Kind() != reflect.Interface && value.Type() != newValue.Type() {
				return ErrTypesDoNotMatch
			}
		} else {
			newValue = reflect.Zero(value.Type())
		}
		value.Set(newValue)
	}
	return nil
}

// getReflectValue gets the reflect.Value of the given value attribute.
// It splits the attribute into the field names, map keys and slice indexes
// and uses reflection to get the final value.
// toSet indicates that the function must return a value that will be set to
// another value, which is used in the special case of maps (maps elements are
// not addressable).
// It also returns the name of the accessed field.
func getReflectValue(value reflect.Value, attribute string, sep string, toSet bool) (_ reflect.Value, fieldName string, _ error) {
	if attribute == "" {
		return value, "", nil
	}

	var i, maxSetDepth int
	if toSet {
		maxSetDepth = strings.Count(attribute, sep)
	}

	splitter := newAttributeSplitter(attribute, sep)
	for splitter.HasMore() {
		fieldName, i = splitter.Next()

		if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			// Check that the map accept string keys
			keyKind := value.Type().Key().Kind()
			if keyKind != reflect.String && keyKind != reflect.Interface {
				return value, "", ErrMapKeyNotString
			}

			// If a map key has to be set, skip the last attribute and return the map
			if toSet && i == maxSetDepth {
				return value, fieldName, nil
			}

			mapValue := value.MapIndex(reflect.ValueOf(fieldName))

			// If the key is not found, it could be because is a dotted key,
			// so try expanding the search with more fields
			if !mapValue.IsValid() {
				splitterMap := newAttributeSplitter(splitter.remain, sep)
				for splitterMap.HasMore() {
					mapKey, mapIndex := splitterMap.Next()
					fieldName += "." + mapKey
					mapValue = value.MapIndex(reflect.ValueOf(fieldName))
					if mapValue.IsValid() {
						// Re-adjust values and splitter
						maxSetDepth -= mapIndex + 1
						if toSet && i == maxSetDepth {
							return value, fieldName, nil
						}
						splitter.remain = splitterMap.remain
						splitter.hasMore = splitterMap.hasMore
						break
					}
				}
			}

			if !mapValue.IsValid() {
				return value, "", ErrNotFound
			}

			value = mapValue

		case reflect.Struct:
			field, ok := value.Type().FieldByName(fieldName)
			if !ok {
				return value, "", ErrNotFound
			}
			// Check if field is unexported (method IsExported() was introduced in Go 1.17)
			if field.PkgPath != "" {
				return value, "", ErrUnexported
			}

			value = value.FieldByName(fieldName)

		case reflect.Slice, reflect.Array:
			sliceIndex, err := strconv.Atoi(fieldName)
			if err != nil {
				return value, "", ErrInvalidIndex
			}
			if sliceIndex < 0 || sliceIndex >= value.Len() {
				return value, "", ErrIndexOutOfRange
			}
			field := value.Index(sliceIndex)
			value = field

		default:
			return value, "", ErrNotFound
		}
	}

	return value, fieldName, nil
}
