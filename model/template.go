package model

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"text/template"
)

var root = template.New("root").Funcs(template.FuncMap{
	"jsonBytes": json.Marshal,
	"json": func(v any) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	},
})

type Template struct {
	data reflect.Value
}

func (t *Template) Reader(text string) (io.Reader, error) {
	tmpl, err := root.New("").Parse(text)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	return buf, tmpl.Execute(buf, t.data)
}

func (t *Template) String(text string) (string, error) {
	r, err := t.Reader(text)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(r)
	return string(b), err
}

func (t *Template) Any(i any) (err error) {
	switch i := i.(type) {
	case map[string]any:
		for k, v := range i {
			s, ok := v.(string)
			if ok {
				i[k], err = t.String(s)
			} else {
				err = t.Any(v)
			}
			if err != nil {
				return
			}
		}
	case []any:
		for idx, v := range i {
			s, ok := v.(string)
			if ok {
				i[idx], err = t.String(s)
			} else {
				err = t.Any(v)
			}
			if err != nil {
				return
			}
		}
	}
	return nil
}

func NewTemplate(data any) *Template {
	return &Template{data: reflect.ValueOf(data)}
}
