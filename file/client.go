package file

import "net/http"

type Client interface {
	Hosts() []string
	Get(url string, options ...string) (*http.Response, error)
}

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.37"

type WithUserAgent []string

func (c WithUserAgent) Hosts() []string { return c }

func (WithUserAgent) Get(url string, options ...string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}

var _ Client = (*WithUserAgent)(nil)
