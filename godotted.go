package godotted

import (
	"reflect"
	"strconv"
	"strings"
)

type Fields map[string]interface{}

func GetAttribute(v interface{}, attribute string) interface{} {
	return getAttribute(v, strings.Split(attribute, "."))
}

func GetAttributes(v interface{}, attributes []string) Fields {
	m := make(Fields, len(attributes))

	for _, attr := range attributes {
		if _, ok := m[attr]; !ok {
			m[attr] = getAttribute(v, strings.Split(attr, "."))
		}
	}

	return m
}

func getAttribute(v interface{}, attribute []string) interface{} {
	for _, attr := range attribute {
		value := reflect.ValueOf(v)
		vType := reflect.TypeOf(v)

		if value.Kind() == reflect.Pointer {
			value = value.Elem()
			vType = vType.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
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
					return ErrMapKeyNotString
				}
				return ErrNotFound
			}
			v = mapValue.Interface()
			continue

		case reflect.Struct:
			field, ok := vType.FieldByName(attr)
			if !ok {
				return ErrNotFound
			}
			if !field.IsExported() {
				return ErrUnexported
			}

			v = value.FieldByName(attr).Interface()

		case reflect.Slice, reflect.Array:
			sliceIndex, err := strconv.Atoi(attr)
			if err != nil {
				return ErrInvalidIndex
			}
			if sliceIndex < 0 || sliceIndex >= value.Len() {
				return ErrIndexOutOfRange
			}
			field := value.Index(sliceIndex)
			v = field.Interface()

		default:
			return ErrNotFound
		}
	}

	return v
}
