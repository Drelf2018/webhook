package model

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"sync"
	"text/template"
	"time"
)

var root = template.New("root").Funcs(template.FuncMap{
	"json": func(v any) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	},
})

type Template struct {
	id   uint64
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

func (t *Template) Do(task *Task) (result []byte, err error) {
	url, err := t.String(task.URL)
	if err != nil {
		return
	}

	var body io.Reader
	if task.Body != "" {
		body, err = t.Reader(task.Body)
		if err != nil {
			return
		}
	}

	req, err := http.NewRequest(task.Method, url, body)
	if err != nil {
		return
	}

	header, err := t.Reader(task.Header.String())
	if err != nil {
		return
	}

	var h map[string]string
	err = json.NewDecoder(header).Decode(&h)
	if err != nil {
		return
	}

	for k, v := range h {
		req.Header.Add(k, v)
	}

	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = http.DefaultClient.Do(req)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (t *Template) RunTask(task *Task) RequestLog {
	log := RequestLog{
		BlogID:    t.id,
		TaskID:    task.ID,
		CreatedAt: time.Now(),
	}
	r, err := t.Do(task)
	if err != nil {
		log.Error = err.Error()
		log.FinishedAt = time.Now()
		return log
	}
	err = json.Unmarshal(r, &log.Result)
	if err != nil {
		log.Result = string(r)
		log.Error = err.Error()
	}
	log.FinishedAt = time.Now()
	return log
}

func (t *Template) RunTasks(tasks []*Task) []RequestLog {
	logs := make([]RequestLog, len(tasks))
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for idx := range tasks {
		idx := idx
		go func() {
			logs[idx] = t.RunTask(tasks[idx])
			wg.Done()
		}()
	}
	wg.Wait()
	return logs
}

func NewTemplate(blog *Blog) *Template {
	return &Template{id: blog.ID, data: reflect.ValueOf(blog)}
}
