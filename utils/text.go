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

func SplitLines(s string) (r []string) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			r = append(r, line)
		}
	}
	return
}
