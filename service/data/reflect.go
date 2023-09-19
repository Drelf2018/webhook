package data

import (
	"fmt"
	"reflect"
	"unsafe"
)

// 反射加速 参考: https://www.cnblogs.com/cheyunhua/p/16642488.html
var unsafeCache = make(map[uintptr][]string)

type variable struct {
	typ unsafe.Pointer
	val unsafe.Pointer
}

func ParseStruct(in any) []string {
	inf := (*variable)(unsafe.Pointer(&in))
	preloads, ok := unsafeCache[uintptr(inf.typ)]
	if !ok {
		typ := reflect.TypeOf(in)
		val := reflect.ValueOf(in)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			val = val.Elem()
		}
		if typ.Kind() != reflect.Struct {
			panic("You should pass in a struct or a pointer to a struct.")
		}
		preloads = make([]string, 0, 16)
		for i, l := 0, typ.NumField(); i < l; i++ {
			preload, exists := typ.Field(i).Tag.Lookup("preload")
			if exists {
				name := typ.Field(i).Name
				if preload == "" {
					for _, s := range ParseStruct(val.Field(i).Interface()) {
						preloads = append(preloads, fmt.Sprintf("%v.%v", name, s))
					}
				} else {
					preloads = append(preloads, name)
				}
			}
		}
		unsafeCache[uintptr(inf.typ)] = preloads
	}
	return preloads
}
