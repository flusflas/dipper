package godotted

import (
	"reflect"
	"strconv"
	"strings"
)

type Fields map[string]interface{}

func Get(v interface{}, attribute string) interface{} {
	return getReflectValue(v, strings.Split(attribute, ".")).Interface()
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

func getReflectValue(v interface{}, attributes []string) reflect.Value {
	value := reflect.ValueOf(v)

	for _, attr := range attributes {

		if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			value = value.Elem()
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
					return errMapKeyNotStringValue
				}
				return errNotFoundValue
			}
			value = mapValue
			continue

		case reflect.Struct:
			field, ok := value.Type().FieldByName(attr)
			if !ok {
				return errNotFoundValue
			}
			if !field.IsExported() {
				return errUnexportedValue
			}

			value = value.FieldByName(attr)

		case reflect.Slice, reflect.Array:
			sliceIndex, err := strconv.Atoi(attr)
			if err != nil {
				return errInvalidIndexValue
			}
			if sliceIndex < 0 || sliceIndex >= value.Len() {
				return errIndexOutOfRangeValue
			}
			field := value.Index(sliceIndex)
			value = field

		default:
			return errNotFoundValue
		}
	}

	return value
}
