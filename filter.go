package dipper

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var filterRegex = regexp.MustCompile(`(?m)^([\w-]*)==?(.*)$`)

// filterSlice takes a slice value and applies on it the given filter
// expression. It returns the first value matching the filter or an empty
// reflect.Value if no match was found.
func filterSlice(value reflect.Value, fieldName string) (reflect.Value, error) {
	if !strings.Contains(fieldName, "=") {
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
		item := value.Index(i)
		itemSafe := getElemSafe(item)

		switch itemSafe.Kind() {
		case reflect.Map:
			for _, mapKey := range itemSafe.MapKeys() {
				if mapKey.String() != filterKey {
					continue
				}

				if compareValues(itemSafe.MapIndex(mapKey)) {
					return item, nil
				}
			}
		case reflect.Struct:
			for i := 0; i < itemSafe.NumField(); i++ {
				if itemSafe.Type().Field(i).Name != filterKey {
					continue
				}

				field := itemSafe.Field(i)
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
