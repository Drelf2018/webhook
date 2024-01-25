package model

import (
	"sync"

	"golang.org/x/exp/slices"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/cmps"
	"github.com/agnivade/levenshtein"
)

var monitors = sync.Map{}

type OrderedPost struct {
	*Post
	// 编辑距离
	Distance int `gorm:"-" cmps:"1"`
}

func (OrderedPost) TableName() string {
	return "posts"
}

// 博文检查器
type Monitor struct {
	sync.Mutex
	Score float64
	Posts []*OrderedPost
}

// 判断用户是否已经提交
func (m *Monitor) IsSubmitted(uid string) bool {
	return slices.ContainsFunc(m.Posts, func(op *OrderedPost) bool { return op.Submitter.Uid == uid })
}

// 解析接收到的博文
func (m *Monitor) Parse(post *OrderedPost) {
	m.Lock()
	defer m.Unlock()
	// 说明已经在处理了
	if m.Score >= 1 {
		return
	}
	if post.Submitter.IsTrusted() {
		m.Score += 1
	} else {
		m.Score += 0.2 * float64(post.Submitter.Level()-1)
	}
	// 检测提交的文本与检查器中储存的文本的相似性
	asyncio.ForEach(m.Posts, func(op *OrderedPost) {
		dis := levenshtein.ComputeDistance(op.Text, post.Text)
		op.Distance += dis
		post.Distance += dis
	})
	m.Posts = append(m.Posts, post)
	// 得分未超过阈值返回
	if m.Score < 1 {
		return
	}
	// 找到相似度最高的
	cmps.Sort(m.Posts)
	// level up
	// asyncio.ForEach(m.Posts, func(p *Post) { p.Submitter.LevelUP() })
	postDB.Create(m.Posts)
	m.Posts[0].Webhook()
	// 所有版本都存完就可以删除该检查器了
	monitors.Delete(post.Key())
}
