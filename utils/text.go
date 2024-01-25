package utils

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 干净文本
func Clean(original string) string {
	html := strings.NewReader("<div id=\"post\">" + original + "</div>")
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return original
	}
	doc.Find("#post .url-icon").Each(
		func(i int, s *goquery.Selection) {
			alt, ok := s.Find("img").Attr("alt")
			if ok {
				s.SetText(alt)
			}
		},
	)
	return doc.Text()
}

// 换行分割字符串
func Cut(b []byte) (r []string) {
	for _, s := range strings.Split(string(b), "\n") {
		if s != "" {
			r = append(r, s)
		}
	}
	return
}
