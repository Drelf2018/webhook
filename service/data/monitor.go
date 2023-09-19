package data

import (
	"slices"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/cmps"
	"github.com/agnivade/levenshtein"
)

// 博文检查器
type Monitor struct {
	Score float64
	Posts []*Post
}

// 判断用户是否已经提交
func (m *Monitor) IsSubmitted(uid string) bool {
	return slices.ContainsFunc(m.Posts, func(p *Post) bool { return p.Submitter.Uid == uid })
}

// 解析接收到的博文
func (m *Monitor) Parse(post *Post) {
	// 说明已经在处理了
	if m.Score >= 1 {
		return
	}
	m.Score += 1 / post.Submitter.Permission
	// 检测提交的文本与检查器中储存的文本的相似性
	asyncio.ForEach(m.Posts, func(p *Post) {
		dis := levenshtein.ComputeDistance(p.Text, post.Text)
		p.Distance += dis
		post.Distance += dis
	})
	m.Posts = append(m.Posts, post)
	// 得分未超过阈值返回
	if m.Score < 1 {
		return
	}
	// 找到相似度最高的
	cmps.Slice(m.Posts)
	SavePost(m.Posts[0])
	SavePosts(m.Posts[1:]...)
	// 所有版本都存完就可以删除该检查器了
	delete(monitors, post.Type())
}

var monitors = make(map[string]*Monitor)

// 获取检查器
func GetMonitor(typ string) *Monitor {
	m, ok := monitors[typ]
	if !ok {
		m = &Monitor{Posts: make([]*Post, 0)}
		monitors[typ] = m
	}
	return m
}
