package data

import (
	"github.com/Drelf2018/webhook/utils"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// 博文检查器
type Monitor struct {
	Score float64
	Posts map[*Post]float64
	final *Post
}

// 判断用户是否已经提交
func (m *Monitor) Has(uid string) bool {
	return slices.ContainsFunc(maps.Keys(m.Posts), func(p *Post) bool { return p.Submitter.Uid == uid })
}

// 储存所有分支
func (m *Monitor) SaveAsBranches() {
	loop := utils.EventLoop[any, any, []any]{Results: &[]any{}}
	for p := range m.Posts {
		p.Blogger.ID = m.final.Blogger.ID
		p.Repost = m.final.Repost
		loop.AddTask(p.SaveAsBranche)
	}
	loop.Wait()
}

// 解析接收到的博文
func (m *Monitor) Parse(post *Post) {
	if m.Score >= 1 {
		// 说明已经在处理了
		return
	}
	// 检测提交的文本与检查器中储存的文本的相似性
	maxPercent := utils.Ternary(len(m.Posts) == 0, 1.0, 0.0)
	totPercent := maxPercent
	for p := range m.Posts {
		percent := utils.SimilarText(p.Text, post.Text)
		if percent > maxPercent {
			maxPercent = percent
		}
		totPercent += percent
		m.Posts[p] += percent
	}
	// 更新可信度得分
	// 假如相似度为 100% 那得分只与 level 有关
	// 即一个 level 1 提交即可超过阈值而至少需要五个 level 5 提交才能超过
	m.Score += maxPercent / post.Submitter.Permission
	m.Posts[post] = totPercent

	// 得分未超过阈值返回
	if m.Score < 1 {
		return
	}
	// 找到相似度最高的
	m.final = post
	for p, s := range m.Posts {
		if s > m.Posts[m.final] {
			m.final = p
		}
	}
	// 下述任务都执行完成就可以删除该检查器了
	m.final.Save()
	m.SaveAsBranches()
	m.final.Webhook()
	delete(Monitors, m.final.Platform+m.final.Mid)
}

type monitors map[string]*Monitor

var Monitors = make(monitors)

// 获取检查器
func GetMonitor(id string) *Monitor {
	m, ok := Monitors[id]
	if !ok {
		m = &Monitor{
			Score: 0,
			Posts: make(map[*Post]float64),
		}
		Monitors[id] = m
	}
	return m
}
