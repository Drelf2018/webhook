package utils

import (
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"sync"

	"github.com/Drelf2018/initial"
)

type Node struct {
	Target *Node
	depth  int
}

func (n *Node) Depth() int {
	if n.depth == 0 {
		n.depth = n.Target.Depth() + 1
	}
	return n.depth
}

var (
	ErrInvalidName = errors.New("webhook/utils: invalid join tag value")
	ErrNotPointer  = errors.New("webhook/utils: value must be a pointer to struct")
)

var edgesCache sync.Map

func LoadEdges(in reflect.Type) [][2]int {
	ptr := initial.ValuePtr(in)
	if v, ok := edgesCache.Load(ptr); ok {
		return v.([][2]int)
	}
	actual, _ := edgesCache.LoadOrStore(ptr, ParseEdges(in))
	return actual.([][2]int)
}

func ParseEdges(elem reflect.Type) (edges [][2]int) {
	names := make(map[string]int)

	for idx := 0; idx < elem.NumField(); idx++ {
		field := elem.Field(idx)
		names[field.Name] = idx

		if !field.IsExported() {
			continue
		}

		if field.Type.Kind() != reflect.String {
			continue
		}

		name := field.Tag.Get("join")

		if name == "" {
			edges = append(edges, [2]int{idx, -1})
		} else {
			i, ok := names[name]
			if !ok {
				f, exists := elem.FieldByName(name)
				if !exists {
					panic(ErrInvalidName)
				}
				i = f.Index[0]
				names[name] = i
			}
			edges = append(edges, [2]int{idx, i})
		}

	}

	nodes := make([]Node, elem.NumField())
	for _, data := range edges {
		if data[1] == -1 {
			nodes[data[0]].depth = 1
		} else {
			nodes[data[0]].Target = &nodes[data[1]]
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		less := nodes[edges[i][0]].Depth() - nodes[edges[j][0]].Depth()
		if less == 0 {
			less = edges[i][1] - edges[j][1]
		}
		if less == 0 {
			less = edges[i][0] - edges[j][0]
		}
		return less < 0
	})
	return
}

func Join(in any) error {
	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Pointer {
		return ErrNotPointer
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrNotPointer
	}

	values := make([]string, v.NumField())

	for _, edge := range LoadEdges(v.Type()) {
		index := edge[0]
		target := edge[1]
		if target == -1 {
			values[index] = v.Field(index).String()
		} else {
			field := v.Field(index)
			values[index] = filepath.Join(values[target], field.String())
			field.SetString(values[index])
		}
	}

	return nil
}

func NewJoin[T any](in T) (*T, error) {
	return &in, Join(&in)
}
