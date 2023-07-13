package tots

import (
	"fmt"
	"reflect"
	"strings"
)

func indentLines(str string, i int) string {
	lines := strings.Split(str, "\n")
	for n := range lines {
		lines[n] = strings.Repeat("\t", i) + lines[n]
	}
	return strings.Join(lines, "\n")
}

func UnionTsType[T any](_items ...T) string {
	items := make([]string, 0, len(_items))
	for _, item := range _items {
		switch reflect.TypeOf(item).Kind() {
		case reflect.String:
			items = append(items, fmt.Sprintf(`"%s"`, fmt.Sprint(item)))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32,
			reflect.Float64:
			items = append(items, fmt.Sprintf(`%s`, fmt.Sprint(item)))
		default:
			panic("current don't support this type")
		}
	}
	return strings.Join(items, "|")
}
