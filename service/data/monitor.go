package data

import (
	"sync"

	"golang.org/x/exp/slices"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/cmps"
	"github.com/agnivade/levenshtein"
)

var monitors = sync.Map{}

// var monitors = make(map[string]*Monitor)

// 博文检查器
type Monitor struct {
	sync.Mutex
	Score float64
	Posts []*Post
}

// 判断用户是否已经提交
func (m *Monitor) IsSubmitted(uid string) bool {
	return slices.ContainsFunc(m.Posts, func(p *Post) bool { return p.Submitter.Uid == uid })
}

// 解析接收到的博文
func (m *Monitor) Parse(post *Post) {
	m.Lock()
	defer m.Unlock()
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
	asyncio.ForEach(m.Posts, func(p *Post) { p.Submitter.LevelUP() })
	Posts.DB.Create(&m.Posts)
	m.Posts[0].Webhook()
	// 所有版本都存完就可以删除该检查器了
	monitors.Delete(post.Type())
}
