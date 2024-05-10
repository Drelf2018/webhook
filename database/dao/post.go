package dao

import (
	"github.com/Drelf2018/webhook/database/model"
)

// 回调博文
func Webhook(p *model.Post) {
	rows, _ := userDB.Model(&model.Job{}).Rows()
	defer rows.Close()
	p.ScanAndSend(rows)
}

// 判断博文是否存在
func ExistsPost(p *model.Post) bool {
	return postDB.FirstOK(new(model.Post), "platform = ? AND mid = ?", p.Platform, p.Mid)
}

func SavePost(p *model.Post) error {
	return postDB.Create(p).Error
}

// 获取起始与结束时间范围内所有博文
func GetPosts(begin, end string) (posts []model.Post) {
	postDB.Preloads(&posts, "time BETWEEN ? AND ?", begin, end)
	return
}
