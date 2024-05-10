package dao

import (
	"errors"

	"github.com/Drelf2018/webhook/database/model"
	uuid "github.com/satori/go.uuid"
)

func RemoveJobs(u *model.User, jobs []string) error {
	return userDB.Delete(&model.Job{UserUid: u.Uid}, jobs).Error
}

func UpdateUser(u *model.User) error {
	return userDB.Updates(u).Error
}

// 根据 auth 查询
func QueryUserByAuth(auth string) *model.User {
	if auth == "" {
		return nil
	}
	u := &model.User{Auth: auth}
	if userDB.PreloadOK(u) {
		return u
	}
	return nil
}

// 根据 uid 查询
func QueryUserByUID(uid string) *model.User {
	if uid == "" {
		return nil
	}
	u := &model.User{Uid: uid}
	if userDB.FirstOK(u) {
		return u
	}
	return nil
}

func GetUsers(conds ...any) (users []model.User, err error) {
	err = userDB.Clone().Preloads(&users, conds...).Error()
	return
}

var ErrNoOwner = errors.New("owner permissions cannot be revoked")

// 修改权限
func UpdatePermission(uid string, p model.Permission) error {
	// user := QueryUserByUID(uid)

	// if user.Permission == model.Owner {
	// 	if p != model.Owner {
	// 		return ErrNoOwner
	// 	}
	// 	return nil
	// }

	// if config.Global.Owner == uid && !p.Has(Owner) {
	// 	return ErrNoOwner
	// }
	return userDB.Updates(&model.User{Uid: uid, Permission: p}).Error
}

// 构造函数
func NewUser(uid string, p model.Permission) (u *model.User) {
	u = &model.User{
		Uid:        uid,
		Auth:       uuid.NewV4().String(),
		Permission: p,
	}
	userDB.Create(u)
	return
}
