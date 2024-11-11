package model

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/lib/pq"
)

const MagicKey string = "__magic_key__"

type Header http.Header

func (h Header) Wrap() error {
	if len(h) == 0 {
		h[MagicKey] = []string{}
	} else {
		b, err := json.Marshal(http.Header(h))
		if err != nil {
			return err
		}
		for key := range h {
			delete(h, key)
		}
		h[MagicKey] = []string{string(b)}
	}
	return nil
}

func (h Header) Unwrap() error {
	v, ok := h[MagicKey]
	if !ok || len(v) == 0 {
		return nil
	}
	delete(h, MagicKey)
	return json.Unmarshal([]byte(v[0]), &h)
}

func (Header) GormDataType() string {
	return "TEXT"
}

func (h *Header) Scan(src any) error {
	if src == nil {
		*h = make(Header)
		return nil
	}
	switch src := src.(type) {
	case []byte:
		*h = Header{MagicKey: []string{string(src)}}
	case string:
		*h = Header{MagicKey: []string{src}}
	default:
		return fmt.Errorf("model: failed to unmarshal Header value: %v", src)
	}
	return nil
}

func (h Header) Value() (driver.Value, error) {
	if h == nil {
		return "{}", nil
	}
	if _, ok := h[MagicKey]; ok {
		err := h.Unwrap()
		if err != nil {
			return nil, err
		}
	}
	b, err := json.Marshal(http.Header(h))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (h Header) MarshalJSON() ([]byte, error) {
	if h == nil {
		return nil, nil
	}
	if len(h) == 0 {
		return []byte("{}"), nil
	}
	v, ok := h[MagicKey]
	if !ok {
		return json.Marshal(http.Header(h))
	}
	if len(v) == 0 {
		return []byte("{}"), nil
	}
	return []byte(v[0]), nil
}

func (h Header) ToString() string {
	if h == nil {
		return "{}"
	}
	v, ok := h[MagicKey]
	if !ok {
		if len(h) != 0 {
			b, err := json.Marshal(http.Header(h))
			if err == nil {
				return string(b)
			}
		}
		return "{}"
	}
	if len(v) == 0 {
		return "{}"
	}
	return v[0]
}

const (
	doNotUnmarshal string = "DoNotUnmarshal"
	DoNotUnmarshal        = "--" + doNotUnmarshal
)

type Api struct {
	Method    string         `json:"method"`
	URL       string         `json:"url"`
	Body      string         `json:"body"`
	Header    Header         `json:"header"`
	Parameter pq.StringArray `json:"parameter" gorm:"type:text[]"`

	doNotUnmarshal bool `gorm:"-"`
}

func (api *Api) Parse() error {
	if api.Parameter == nil {
		return nil
	}
	set := flag.NewFlagSet("api", flag.ContinueOnError)
	set.BoolVar(&api.doNotUnmarshal, doNotUnmarshal, false, doNotUnmarshal)
	return set.Parse(api.Parameter)
}

func (api Api) Do(tmpl *Template) ([]byte, error) {
	return api.DoWithContext(context.Background(), tmpl)
}

func (api Api) DoWithContext(ctx context.Context, tmpl *Template) (result []byte, err error) {
	err = api.Parse()
	if err != nil {
		return
	}

	api.URL, err = tmpl.String(api.URL)
	if err != nil {
		return
	}

	var body io.Reader

	if api.Body == "" {
		body = nil
	} else if api.doNotUnmarshal {
		body, err = tmpl.Reader(api.Body)
	} else {
		var i any
		err = json.Unmarshal([]byte(api.Body), &i)
		if err != nil {
			return
		}

		s, ok := i.(string)
		if ok {
			i, err = tmpl.String(s)
		} else {
			err = tmpl.Any(i)
		}
		if err != nil {
			return
		}

		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(i)
		body = buf
	}
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, api.Method, api.URL, body)
	if err != nil {
		return
	}

	header, err := tmpl.Reader(api.Header.ToString())
	if err != nil {
		return
	}

	err = json.NewDecoder(header).Decode(&req.Header)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
