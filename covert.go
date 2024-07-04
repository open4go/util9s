package util9s

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// GetFieldString 获取结构体字段及值的拼接值
func GetFieldString(sendParamEntity interface{}) string {
	m := reflect.TypeOf(sendParamEntity)
	v := reflect.ValueOf(sendParamEntity)

	// Check if the type is a pointer and get the element type
	if m.Kind() == reflect.Ptr {
		m = m.Elem()
		v = v.Elem()
	}

	// If the type is not a struct, return an empty string
	if m.Kind() != reflect.Struct {
		return ""
	}

	var tagName string
	numField := m.NumField()
	w := make([]string, 0, numField)

	for i := 0; i < numField; i++ {
		field := m.Field(i)
		fieldValue := v.Field(i)

		// Check if the field is valid and can be accessed
		if !fieldValue.IsValid() || !fieldValue.CanInterface() {
			continue
		}

		fieldName := field.Name
		tags := strings.Split(string(field.Tag), "\"")
		if len(tags) > 1 {
			tagName = tags[1]
		} else {
			tagName = fieldName
		}
		if tagName == "xml" {
			continue
		}

		// Only add non-empty field values
		if fieldValue.Interface() != "" {
			if strings.Contains(tagName, "omitempty") {
				tagName = strings.Split(tagName, ",")[0]
			}
			s := fmt.Sprintf("%s=%v", tagName, fieldValue.Interface())
			w = append(w, s)
		}
	}

	if len(w) == 0 {
		return ""
	}
	sort.Strings(w)
	return strings.Join(w, "&")
}
