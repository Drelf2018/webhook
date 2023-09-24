package db

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
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if typ.Kind() == reflect.Slice {
			typ = typ.Elem()
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
					v := reflect.New(typ.Field(i).Type).Interface()
					for _, s := range ParseStruct(v) {
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
