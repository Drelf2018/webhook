package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
)

const (
	ErrInvalidOp   = `invalid patch op: "%s"`
	ErrInvalidPath = `invalid patch path: "%s"`
)

type PatchBody struct {
	Op    string `json:"op"` // [replace, add, remove, move, copy, test]
	Path  string `json:"path"`
	Value string `json:"value,omitempty"`
	From  string `json:"from,omitempty"`
}

func createError(errs []group.Response) error {
	buf := bytes.NewBufferString("webhook/api: [")
	for i, r := range errs {
		if i != 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(fmt.Sprintf("%s (%d)", r.Error, r.Code))
	}
	buf.WriteByte(']')
	return errors.New(buf.String())
}

// 修改用户信息
func PatchUserUID(ctx *gin.Context) (any, error) {
	var body []PatchBody
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		return 1, err
	}

	user := &model.User{UID: ctx.Param("uid")}
	tx := UserDB.Limit(1).Find(user)
	if tx.Error != nil {
		return 2, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 3, ErrUserNotExist
	}

	me := &model.User{UID: GetUID(ctx)}
	err = UserDB.First(me).Error
	if err != nil {
		return 4, err
	}

	var errs []group.Response
	for i, patch := range body {
		err = nil
		switch patch.Path {
		default:
			err = fmt.Errorf(ErrInvalidPath, patch.Path)
		case "/ban":
			err = PatchUserBan(ctx, me, user, patch)
			if user.Ban.After(time.Now()) {
				tokenIssuedAt.Delete(user.UID)
			}
		case "/role":
			err = PatchUserRole(ctx, me, user, patch)
		case "/name":
			err = PatchUserName(ctx, me, user, patch)
		case "/nickname":
			err = PatchUserNickname(ctx, me, user, patch)
		}
		if err != nil {
			errs = append(errs, group.Response{Code: i, Error: err.Error()})
		}
	}

	if len(errs) != 0 {
		return 5, createError(errs)
	}

	err = UserDB.Save(user).Error
	if err != nil {
		return 6, err
	}
	return Success, nil
}

func PatchUserBan(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if !me.Role.IsAdmin() {
		return ErrPermDenied
	}
	if me.Role <= user.Role {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		return user.Ban.UnmarshalJSON([]byte(patch.Value))
	case "add":
		i, err := strconv.Atoi(patch.Value)
		if err != nil {
			return err
		}
		user.Ban = user.Ban.Add(time.Duration(i))
		return nil
	case "remove":
		user.Ban = time.Time{}
		return nil
	default:
		return fmt.Errorf(ErrInvalidOp, patch.Op)
	}
}

func PatchUserRole(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if !me.Role.IsAdmin() {
		return ErrPermDenied
	}
	if me.Role <= user.Role {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		i, err := strconv.Atoi(patch.Value)
		if err != nil {
			return err
		}
		if i <= 0 || i >= int(me.Role) {
			return ErrPermDenied
		}
		user.Role = model.Role(i)
		return nil
	default:
		return fmt.Errorf(ErrInvalidOp, patch.Op)
	}
}

func PatchUserName(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if !me.Role.IsAdmin() {
		return ErrPermDenied
	}
	if me.Role < user.Role {
		return ErrPermDenied
	}
	if me.Role == user.Role && me.UID != user.UID {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		user.Name = patch.Value
		return nil
	default:
		return fmt.Errorf(ErrInvalidOp, patch.Op)
	}
}

func PatchUserNickname(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if me.UID != user.UID && (!me.Role.IsAdmin() || me.Role <= user.Role) {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		user.Nickname = patch.Value
		return nil
	default:
		return fmt.Errorf(ErrInvalidOp, patch.Op)
	}
}

// 修改任务
func PatchTaskID(ctx *gin.Context) (any, error) {
	var body []PatchBody
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		return 1, err
	}

	task := &model.Task{}
	tx := UserDB.Limit(1).Find(task, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx))
	if tx.Error != nil {
		return 2, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 3, ErrTaskNotExist
	}

	var errs []group.Response
	for i, patch := range body {
		err = nil
		switch patch.Path {
		default:
			err = fmt.Errorf(ErrInvalidPath, patch.Path)
		case "/public":
			if patch.Op == "add" {
				task.Public = true
			} else {
				err = fmt.Errorf(ErrInvalidOp, patch.Op)
			}
		case "/enable":
			task.Enable, err = strconv.ParseBool(patch.Value)
		case "/name":
			task.Name = patch.Value
		case "/icon":
			task.Icon = patch.Value
		case "/method":
			task.Method = patch.Value
		case "/url":
			task.URL = patch.Value
		case "/body":
			task.Body = patch.Value
		case "/header":
			var header model.Header
			err = json.Unmarshal([]byte(patch.Value), &header)
			task.Header = header
		case "/README":
		case "/readme":
			task.README = patch.Value
		case "/filters":
			var filters []model.Filter
			err = json.Unmarshal([]byte(patch.Value), &filters)
			if err != nil {
				break
			}
			if len(filters) == 0 {
				err = ErrFilterNotExist
				break
			}
			task.Filters = DeduplicateFilters(filters)
		}
		if err != nil {
			errs = append(errs, group.Response{Code: i, Error: err.Error()})
		}
	}

	if len(errs) != 0 {
		return 4, createError(errs)
	}

	if len(task.Filters) != 0 {
		err = UserDB.Delete(&model.Filter{}, "task_id = ?", task.ID).Error
		if err != nil {
			return 5, err
		}
	}
	err = UserDB.Save(task).Error
	if err != nil {
		return 6, err
	}
	return Success, nil
}
