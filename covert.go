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
	var tagName string
	numField := m.NumField()
	w := make([]string, numField)
	numFieldCount := 0
	for i := 0; i < numField; i++ {
		fieldName := m.Field(i).Name
		tags := strings.Split(string(m.Field(i).Tag), "\"")
		if len(tags) > 1 {
			tagName = tags[1]
		} else {
			tagName = m.Field(i).Name
		}
		if tagName == "xml" {
			continue
		}
		fieldValue := v.FieldByName(fieldName).Interface()

		if fieldValue != "" {
			if strings.Contains(tagName, "omitempty") {
				tagName = strings.Split(tagName, ",")[0]
			}
			s := fmt.Sprintf("%s=%v", tagName, fieldValue)
			w[numFieldCount] = s
			numFieldCount++
		}
	}
	if numFieldCount == 0 {
		return ""
	}
	w = w[:numFieldCount]
	sort.Strings(w)
	return strings.Join(w, "&")
}
