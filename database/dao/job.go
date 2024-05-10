package dao

import "github.com/Drelf2018/webhook/database/model"

// 获取指定序号任务
func GetJobsByID(uid string, id []string) (jobs []model.Job) {
	userDB.Find(&jobs, "user_uid = ? and id IN ?", uid, id)
	return
}
