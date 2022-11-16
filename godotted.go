package godotted

import (
	"godotted/splitter"
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

type Fields map[string]interface{}

func Get(v interface{}, attribute string) interface{} {
	value, err := getReflectValue(reflect.ValueOf(v), attribute, false)
	if err != nil {
		return err
	}
	return value.Interface()
}

func GetMany(v interface{}, attributes []string) Fields {
	m := make(Fields, len(attributes))

	for _, attr := range attributes {
		if _, ok := m[attr]; !ok {
			m[attr] = Get(v, attr)
		}
	}

	return m
}

func getReflectValue(value reflect.Value, attribute string, toSet bool) (reflect.Value, error) {
	if attribute == "" {
		return value, nil
	}

	var i, maxSetDepth int
	var attr string
	if toSet {
		maxSetDepth = strings.Count(attribute, ".")
	}

	split := splitter.NewSplitter(attribute, ".")
	for split.HasMore() {
		attr, i = split.Next()

		if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			// If a map key has to be set, skip the last attribute and return the map
			if toSet && i == maxSetDepth {
				break
			}

			getMapValue := func(v reflect.Value, key string) (mapValue reflect.Value, invalidKeyValue bool) {
				defer func() {
					if e := recover(); e != nil {
						invalidKeyValue = true
					}
				}()
				return v.MapIndex(reflect.ValueOf(attr)), false
			}

			mapValue, invalidKeyType := getMapValue(value, attr)
			if !mapValue.IsValid() {
				if invalidKeyType {
					return value, ErrMapKeyNotString
				}
				return value, ErrNotFound
			}
			value = mapValue

		case reflect.Struct:
			field, ok := value.Type().FieldByName(attr)
			if !ok {
				return value, ErrNotFound
			}
			if !field.IsExported() {
				return value, ErrUnexported
			}

			value = value.FieldByName(attr)

		case reflect.Slice, reflect.Array:
			sliceIndex, err := strconv.Atoi(attr)
			if err != nil {
				return value, ErrInvalidIndex
			}
			if sliceIndex < 0 || sliceIndex >= value.Len() {
				return value, ErrIndexOutOfRange
			}
			field := value.Index(sliceIndex)
			value = field

		default:
			return value, ErrNotFound
		}
	}

	return value, nil
}

func Set(v interface{}, attribute string, newValue interface{}) error {
	var err error

	value := reflect.ValueOf(v)

	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	value, err = getReflectValue(value, attribute, true)
	if err != nil {
		return err
	}

	var optZero, optDelete bool

	var val reflect.Value
	switch newValue {
	case Zero:
		optZero = true
	case Delete:
		optDelete = true
	default:
		val = reflect.ValueOf(newValue)
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		}
	}

	if value.Kind() == reflect.Map {
		if !optZero && !optDelete {
			mapValueType := value.Type().Elem()
			if mapValueType.Kind() != reflect.Interface && mapValueType != val.Type() {
				return ErrTypesDoNotMatch
			}
		}
		value.SetMapIndex(reflect.ValueOf(lastStringPart(attribute)), val)
	} else {
		if !optZero && !optDelete {
			if !value.CanAddr() {
				return ErrUnaddressable
			}
			if value.Kind() != reflect.Interface && value.Type() != val.Type() {
				return ErrTypesDoNotMatch
			}
		} else {
			val = reflect.Zero(value.Type())
		}
		value.Set(val)
	}
	return nil
}

func lastStringPart(s string) string {
	index := strings.LastIndexByte(s, '.')
	if index == -1 {
		return s
	}
	return s[index+1:]
}
