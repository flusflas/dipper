package dipper

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// setOption is a type used for special assignments in a set operation.
type setOption int

const (
	// Zero is used as the new value in Set() to set the attribute to its zero
	// value (e.g. "" for string, nil for interface{}, etc.).
	Zero setOption = 0
	// Delete is used as the new value in Set() to delete a map key. If the
	// field is not a map value, the value will be zeroed (see Zero).
	Delete setOption = 1
)

var filterRegex = regexp.MustCompile(`(?m)\[([\w-]*)==?(.*)]`)

// Options defines the configuration of a Dipper instance.
type Options struct {
	Separator string
}

// Dipper allows to access deeply-nested object attributes to get or set their
// values. Attributes are specified by a string with its fields separated by
// some delimiter (e.g. “Books.3.Author" or "Books->3->Author", with "." and
// "->" as delimiters, respectively).
type Dipper struct {
	separator string
}

// New returns a new Dipper instance.
func New(opts Options) *Dipper {
	return &Dipper{separator: opts.Separator}
}

// Get returns the value of the given obj attribute. The attribute uses some
// delimiter-notation to allow accessing nested fields, slice elements or map
// keys. Field names and key maps are case-sensitive.
// All the struct fields accessed must be exported.
// If an error occurs, it will be returned as the attribute value, so it should
// be handled. All the returned errors are fieldError.
//
// Example:
//
//	 // Using "." as the Dipper separator
//		v := my_dipper.Get(myObj, "SomeStructField.1.some_key_map")
//		if err := Error(v); err != nil {
//		    return err
//		}
func (d *Dipper) Get(obj interface{}, attribute string) interface{} {
	value, _, err := getReflectValue(reflect.ValueOf(obj), attribute, d.separator, false)
	if err != nil {
		return err
	}
	return value.Interface()
}

// GetMany returns a map with the values of the given obj attributes.
// It works as Dipper.Get(), but it takes a slice of attributes to return their
// corresponding values. The returned map will have the same length as the
// attributes slice, with the attributes as keys.
//
// Example:
//
//	 // Using "." as the Dipper separator
//		v := my_dipper.GetMany(myObj, []string{"Name", "Age", "Skills.skydiving})
//		if err := v.FirstError(); err != nil {
//		    return err
//		}
func (d *Dipper) GetMany(obj interface{}, attributes []string) Fields {
	m := make(Fields, len(attributes))

	for _, attr := range attributes {
		if _, ok := m[attr]; !ok {
			m[attr] = d.Get(obj, attr)
		}
	}

	return m
}

// Set sets the value of the given obj attribute to the new provided value.
// The attribute uses some delimiter-notation to allow accessing nested fields,
// slice elements or map keys. Field names and key maps are case-sensitive.
// All the struct fields accessed must be exported.
// ErrUnaddressable will be returned if obj is not addressable.
// It returns nil if the value was successfully set, otherwise it will return
// a fieldError.
//
// Example:
//
//	 // Using "." as the Dipper separator
//		v := my_dipper.Set(&myObj, "SomeStructField.1.some_key_map", 123)
//		if err != nil {
//		    return err
//		}
func (d *Dipper) Set(obj interface{}, attribute string, new interface{}) error {
	var err error

	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	var lastField string
	value, lastField, err = getReflectValue(value, attribute, d.separator, true)
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

	if len(sep) == 0 {
		sep = "."
	}

	splitter := newAttributeSplitter(attribute, sep)

	var i, maxSetDepth int
	if toSet {
		maxSetDepth = splitter.CountRemaining()
	}

	for splitter.HasMore() {
		fieldName, i = splitter.Next()
		value = getElemSafe(value)

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
			// Try to apply the filter to the slice elements
			foundValue, err := filterSlice(value, fieldName)
			if err != nil {
				return value, "", err
			}
			if foundValue.IsValid() {
				value = foundValue
				break
			}

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

// filterSlice takes a slice value and applies on it the given filter
// expression. It returns the first value matching the filter or an empty
// reflect.Value if no match was found.
func filterSlice(value reflect.Value, fieldName string) (reflect.Value, error) {
	if !strings.HasPrefix(fieldName, "[") || !strings.HasSuffix(fieldName, "]") || !strings.Contains(fieldName, "=") {
		return reflect.Value{}, nil
	}

	// Parse filter expression
	match := filterRegex.FindStringSubmatch(fieldName)
	if match == nil {
		return reflect.Value{}, ErrInvalidFilterExpression
	}

	// This function converts the filter value string to the proper type
	parseFilterValue := func(v string) (interface{}, error) {
		if strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'") {
			return v[1 : len(v)-1], nil
		}

		if v == "true" || v == "false" {
			return v == "true", nil
		}

		if v == "null" {
			return nil, nil
		}

		parsed, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return parsed, nil
		}

		return nil, ErrInvalidFilterValue
	}

	filterKey := match[1]
	filterValue, err := parseFilterValue(match[2])
	if err != nil {
		return reflect.Value{}, err
	}

	// This function returns the numeric value of the given reflect.Value in
	// float64 or an error if the value is not numerical.
	toFloat64 := func(v reflect.Value) (float64, error) {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(v.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(v.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return v.Float(), nil
		default:
			return 0, fmt.Errorf("unsupported kind: %s", v.Kind())
		}
	}

	// This function compares the value with filterValue
	compareValues := func(v reflect.Value) bool {
		v = getElemSafe(v)

		floatValue, err := toFloat64(v)
		if err == nil {
			return reflect.DeepEqual(floatValue, filterValue)
		}

		return reflect.DeepEqual(v.Interface(), filterValue)
	}

	// Iterates over the value elements and returns the first matching value
	for i := 0; i < value.Len(); i++ {
		item := getElemSafe(value.Index(i))

		switch item.Kind() {
		case reflect.Map:
			for _, mapKey := range item.MapKeys() {
				if mapKey.String() != filterKey {
					continue
				}

				if compareValues(item.MapIndex(mapKey)) {
					return item, nil
				}
			}
		case reflect.Struct:
			for i := 0; i < item.NumField(); i++ {
				if item.Type().Field(i).Name != filterKey {
					continue
				}

				field := item.Field(i)
				if compareValues(field) {
					return item, nil
				}
			}
		default:
			if filterKey == "" && compareValues(item) {
				return item, nil
			}
		}
	}

	return reflect.Value{}, ErrFilterNotFound
}

// getElemSafe returns the underlying value of an interface/pointer reflect.Value.
func getElemSafe(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Func, reflect.Chan, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return v
		}
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}
