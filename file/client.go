package file

import "net/http"

type Client interface {
	Hosts() []string
	Get(url string) (*http.Response, error)
}

type DefaultClient struct{}

func (DefaultClient) Hosts() []string { return nil }

func (DefaultClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}

var cli Client = DefaultClient{}

type WeiboClient struct{}

func (WeiboClient) Hosts() []string {
	return []string{"sinaimg.cn"}
}

func (WeiboClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://weibo.com/")
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}

var _ Client = WeiboClient{}
