package godotted

import (
	"reflect"
	"strconv"
	"strings"
)

func GetAttributes(v interface{}, attributes []string) map[string]interface{} {
	return getAttributes(v, attributes, make(map[string]interface{}), "")
}

func getAttributes(v interface{}, attributes []string, m map[string]interface{}, root string) map[string]interface{} {
	if len(attributes) == 0 {
		return m
	}

	appendDottedField := func(root, fieldName string) string {
		if len(root) == 0 {
			return fieldName
		} else {
			return root + "." + fieldName
		}
	}

	checkField := func(fullName string, fieldValue reflect.Value) {
		partial, exact, exactIndex := containsString(attributes, fullName)
		if exact {
			m[fullName] = fieldValue.Interface()
			// Remove from attributes (setting it empty is more efficient than actually removing it)
			attributes[exactIndex] = ""
		}
		if !partial {
			return
		}

		switch fieldValue.Kind() {
		case reflect.Map, reflect.Struct, reflect.Interface, reflect.Array, reflect.Slice:
			getAttributes(fieldValue.Interface(), attributes, m, fullName)
		default:
		}
	}

	value := reflect.ValueOf(v)
	vType := reflect.TypeOf(v)

	switch value.Kind() {
	case reflect.Map:
		// Get only the map keys given by attributes for early pruning
		keysWanted := make(map[string]struct{})
		for _, attr := range attributes {
			if strings.HasPrefix(attr, root) && attr != root {
				b := attr[len(root):]
				b = strings.TrimPrefix(b, ".")
				b = strings.Split(b, "[")[0]
				b = strings.Split(b, ".")[0]
				keysWanted[b] = struct{}{}
			}
		}

		iter := value.MapRange()
		for iter.Next() {
			if len(keysWanted) == 0 {
				break
			}
			key, ok := iter.Key().Interface().(string) // Support for map[string] only
			if !ok {
				continue
			}
			if _, ok = keysWanted[key]; !ok {
				continue
			}

			delete(keysWanted, key)
			checkField(appendDottedField(root, key), iter.Value())
		}

	case reflect.Struct:
		for i := 0; i < vType.NumField(); i++ {
			field := vType.Field(i)
			if !field.IsExported() {
				continue
			}

			checkField(appendDottedField(root, field.Name), value.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			fullName := root + "[" + strconv.Itoa(i) + "]"
			checkField(fullName, value.Index(i))
		}
	}

	return m
}

func containsString(slice []string, s string) (partial bool, exact bool, exactIndex int) {
	for i, item := range slice {
		if item == s {
			exact = true
			exactIndex = i
		} else if strings.HasPrefix(item, s) {
			partial = true
		}
	}
	return
}
