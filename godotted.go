package godotted

import (
	"reflect"
	"strconv"
	"strings"
)

type Resp int

const (
	NotFound Resp = iota
	InvalidIndex
	IndexOutOfRange
	MapKeyNotString
	Unexported
)

func GetAttribute(v interface{}, attribute string) interface{} {
	return getAttribute(v, strings.Split(attribute, "."))
}

func GetAttributes(v interface{}, attributes []string) map[string]interface{} {
	m := make(map[string]interface{}, len(attributes))

	for _, attr := range attributes {
		if _, ok := m[attr]; !ok {
			m[attr] = getAttribute(v, strings.Split(attr, "."))
		}
	}

	return m
}

func getAttribute(v interface{}, attribute []string) interface{} {
	for i, attr := range attribute {
		if i == len(attribute) {
			return v
		}

		value := reflect.ValueOf(v)
		vType := reflect.TypeOf(v)

		if value.Kind() == reflect.Pointer {
			value = value.Elem()
			vType = vType.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			mapValue := value.MapIndex(reflect.ValueOf(attr))
			if !mapValue.IsValid() {
				return NotFound
			}
			v = mapValue.Interface()
			continue

		case reflect.Struct:
			field, ok := vType.FieldByName(attr)
			if !ok {
				return NotFound
			}
			if !field.IsExported() {
				return Unexported
			}

			v = value.FieldByName(attr).Interface()

		case reflect.Slice, reflect.Array:
			sliceIndex, err := strconv.Atoi(attr)
			if err != nil {
				return InvalidIndex
			}
			if sliceIndex < 0 || sliceIndex >= value.Len() {
				return IndexOutOfRange
			}
			field := value.Index(sliceIndex)
			v = field.Interface()

		default:
			return NotFound
		}
	}

	return v
}
