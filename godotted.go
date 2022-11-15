package godotted

import (
	"reflect"
	"strconv"
	"strings"
)

type Fields map[string]interface{}

func Get(v interface{}, attribute string) interface{} {
	value, err := getReflectValue(reflect.ValueOf(v), strings.Split(attribute, "."), false)
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

func getReflectValue(value reflect.Value, attributes []string, toSet bool) (reflect.Value, error) {
	if len(attributes) == 1 && attributes[0] == "" {
		return value, nil
	}

	for i, attr := range attributes {
		if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			if toSet && i == len(attributes)-1 {
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
			continue

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

	attributes := strings.Split(attribute, ".")
	value, err = getReflectValue(value, attributes, true)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(newValue)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if value.Kind() == reflect.Map {
		mapValueType := value.Type().Elem()
		if mapValueType.Kind() != reflect.Interface && mapValueType != val.Type() {
			return ErrTypesDoNotMatch
		}
		value.SetMapIndex(reflect.ValueOf(attributes[len(attributes)-1]), val)
	} else {
		if !value.CanAddr() {
			return ErrUnaddressable
		}
		if value.Kind() != reflect.Interface && value.Type() != val.Type() {
			return ErrTypesDoNotMatch
		}
		value.Set(val)
	}
	return nil
}
